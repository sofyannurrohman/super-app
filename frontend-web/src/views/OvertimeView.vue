<script setup lang="ts">
import { ref, onMounted, h, computed } from 'vue'
import { toast } from 'vue-sonner'
import {
  Pencil,
  Trash2,
  CheckCircle2,
  XCircle,
  Clock,
  Calendar,
  Info,
  MoreVertical,
  ChevronRight,
  Plus,
  Image,
  ExternalLink,
} from 'lucide-vue-next'
import { resolvePhotoUrl } from '@/utils/imageUrl'
import apiClient from '@/api/axios'
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
import { useAuthStore } from '@/stores/auth'

const authStore = useAuthStore()
const isLoading = ref(true)
const allOvertimes = ref<any[]>([])
const myOvertimes = ref<any[]>([])
const isSubmitting = ref(false)

const isAdmin = computed(() => {
  const roles = ['superadmin', 'hr_admin', 'manager']
  return authStore.user && roles.includes(authStore.user.role)
})

// Attachment Viewer State
const isAttachmentViewerOpen = ref(false)
const selectedAttachmentUrl = ref('')
const hasImageError = ref(false)

const openAttachmentViewer = (url: string) => {
  selectedAttachmentUrl.value = resolvePhotoUrl(url)
  hasImageError.value = false
  isAttachmentViewerOpen.value = true
}

// Form State
const isModalOpen = ref(false)
const isEditMode = ref(false)
const requestForm = ref({
  id: '',
  date: new Date().toISOString().split('T')[0],
  start_time: '17:00',
  end_time: '19:00',
  duration_minutes: 120,
  type: 'working_day',
  reason: '',
  attachment_url: '',
})
const attachment = ref<File | null>(null)

// Reject modal
const isRejectModalOpen = ref(false)
const selectedOvertimeId = ref<string>('')
const rejectReason = ref<string>('')

const overtimeTypes = [
  { value: 'working_day', label: 'Hari Kerja' },
  { value: 'holiday', label: 'Hari Libur / Weekend' },
  { value: 'emergency', label: 'Emergency / Call-out' },
]

const getInitials = (name: string) => {
  if (!name) return 'KA'
  const parts = name.trim().split(/\s+/).filter(Boolean)
  if (parts.length >= 2 && parts[0]?.[0] && parts[1]?.[0]) {
    return `${parts[0][0]}${parts[1][0]}`.toUpperCase()
  }
  return name.substring(0, 2).toUpperCase()
}

const extractTimeHHMM = (isoStr: string, separator = '.') => {
  if (!isoStr) return `00${separator}00`
  const str = String(isoStr)
  const match = str.match(/T(\d{2}):(\d{2})/)
  if (match && str.endsWith('Z')) {
    return `${match[1]}${separator}${match[2]}`
  }
  const d = new Date(str)
  if (isNaN(d.getTime())) return `00${separator}00`
  const hh = String(d.getHours()).padStart(2, '0')
  const mm = String(d.getMinutes()).padStart(2, '0')
  return `${hh}${separator}${mm}`
}

const columns = [
  {
    accessorKey: 'name',
    header: 'KARYAWAN',
    cell: ({ row }: any) => {
      const item = row.original
      return h('div', { class: 'flex items-center gap-3' }, [
        h(
          'div',
          {
            class:
              'w-8 h-8 rounded-full bg-primary/10 text-primary flex items-center justify-center font-bold text-[10px] border border-primary/20 shrink-0',
          },
          item.initials || item.name.substring(0, 2).toUpperCase(),
        ),
        h('div', { class: 'flex flex-col' }, [
          h('span', { class: 'font-bold text-gray-900 leading-tight' }, item.name),
          item.position || item.branch
            ? h(
                'span',
                { class: 'text-[11px] text-gray-500 leading-tight mt-0.5' },
                [item.position, item.branch].filter(Boolean).join(' • '),
              )
            : null,
          h(
            'span',
            { class: 'text-[10px] text-primary/80 font-semibold uppercase tracking-wider mt-0.5' },
            item.type_label,
          ),
        ]),
      ])
    },
  },
  {
    accessorKey: 'date',
    header: 'TANGGAL & WAKTU',
    cell: ({ row }: any) => {
      const item = row.original
      return h('div', { class: 'flex flex-col' }, [
        h('div', { class: 'flex items-center gap-1.5 text-gray-900 font-medium' }, [
          h(Calendar, { class: 'w-3.5 h-3.5 text-gray-400' }),
          h('span', {}, item.date_formatted),
        ]),
        h('div', { class: 'flex items-center gap-1.5 text-gray-500 text-[12px]' }, [
          h(Clock, { class: 'w-3.5 h-3.5' }),
          h(
            'span',
            {},
            `${item.start_time_only} - ${item.end_time_only} (${item.duration_formatted})`,
          ),
        ]),
      ])
    },
  },
  {
    accessorKey: 'reason',
    header: 'KETERANGAN',
    cell: (info: any) =>
      h('div', { class: 'max-w-[200px]' }, [
        h(
          'p',
          { class: 'text-[13px] text-gray-600 line-clamp-2 leading-relaxed italic' },
          `"${info.getValue() || '-'}"`,
        ),
      ]),
  },
  {
    accessorKey: 'status',
    header: 'STATUS',
    cell: ({ row }: any) => {
      const val = row.original.status || 'pending'
      let styles = 'bg-amber-50 text-amber-700 border-amber-200'
      let icon = Clock

      if (val === 'approved') {
        styles = 'bg-emerald-50 text-emerald-700 border-emerald-200'
        icon = CheckCircle2
      } else if (val === 'rejected') {
        styles = 'bg-rose-50 text-rose-700 border-rose-200'
        icon = XCircle
      }

      const approver = row.original.approver
      const approverName = approver ? `${approver.first_name || ''} ${approver.last_name || ''}`.trim() : ''
      const approverPosition = approver?.job_position?.title || ''

      return h('div', { class: 'flex flex-col gap-1' }, [
        h(
          'span',
          {
            class: `${styles} border px-2.5 py-0.5 rounded-full text-[11px] font-bold inline-flex items-center gap-1.5 w-fit uppercase tracking-wider`,
          },
          [h(icon, { class: 'w-3 h-3' }), val],
        ),
        val === 'approved' && approverName
          ? h('div', { class: 'flex flex-col text-[10px] text-gray-500 mt-0.5 leading-tight' }, [
              h('span', { class: 'font-semibold text-gray-700' }, `✓ ${approverName}`),
              approverPosition ? h('span', { class: 'text-gray-400 text-[9px]' }, approverPosition) : null,
            ])
          : null,
        row.original.reject_reason
          ? h(
              'span',
              { class: 'text-[10px] text-rose-500 font-medium ml-1 truncate max-w-[120px]' },
              `Ket: ${row.original.reject_reason}`,
            )
          : null,
      ])
    },
  },
  {
    id: 'actions',
    header: 'AKSI',
    cell: ({ row }: any) => {
      const item = row.original
      const isOwner = authStore.user && item.employee_id === authStore.user.employee_id

      return h('div', { class: 'flex items-center gap-2' }, [
        // Admin Actions (Approve/Reject)
        item.status === 'pending' && isAdmin.value
          ? [
              h(
                Button,
                {
                  variant: 'ghost',
                  size: 'sm',
                  class: 'h-8 px-2 text-emerald-600 hover:text-emerald-700 hover:bg-emerald-50',
                  onClick: () => updateStatus(item.id, 'approved'),
                },
                () => h(CheckCircle2, { class: 'w-4 h-4' }),
              ),
              h(
                Button,
                {
                  variant: 'ghost',
                  size: 'sm',
                  class: 'h-8 px-2 text-rose-600 hover:text-rose-700 hover:bg-rose-50',
                  onClick: () => openRejectModal(item.id),
                },
                () => h(XCircle, { class: 'w-4 h-4' }),
              ),
            ]
          : null,

        // View Attachment
        item.attachment_url
          ? h(
              'button',
              {
                type: 'button',
                onClick: () => openAttachmentViewer(item.attachment_url),
                class: 'inline-flex h-8 px-2 items-center justify-center rounded-md text-sm font-medium transition-colors text-indigo-600 hover:text-indigo-700 hover:bg-indigo-50',
                title: 'Lihat Lampiran',
              },
              [h(Image, { class: 'w-4 h-4' })],
            )
          : null,

        // Owner Actions (Edit/Delete)
        item.status === 'pending' && (isOwner || isAdmin.value)
          ? [
              h(
                Button,
                {
                  variant: 'ghost',
                  size: 'sm',
                  class: 'h-8 px-2 text-slate-400 hover:text-slate-900',
                  onClick: () => openEditModal(item),
                },
                () => h(Pencil, { class: 'w-4 h-4' }),
              ),
              h(
                Button,
                {
                  variant: 'ghost',
                  size: 'sm',
                  class: 'h-8 px-2 text-slate-400 hover:text-rose-600',
                  onClick: () => deleteOvertime(item.id),
                },
                () => h(Trash2, { class: 'w-4 h-4' }),
              ),
            ]
          : null,

        // Finalized Label
        item.status !== 'pending'
          ? h('span', { class: 'text-[11px] font-bold text-slate-300 uppercase' }, 'Selesai')
          : null,
      ])
    },
  },
]

const fetchOvertimes = async () => {
  isLoading.value = true
  try {
    const requests = [apiClient.get('/overtimes/me')]
    if (isAdmin.value) {
      requests.push(apiClient.get('/overtimes/'))
    }

    const responses = await Promise.all(requests)

    const transform = (item: any) => {
      const emp = item.employee || item.Employee
      const fullName = emp
        ? `${emp.first_name || ''} ${emp.last_name || ''}`.trim()
        : ''
      const name = fullName || 'Karyawan'
      return {
        ...item,
        name,
        initials: getInitials(name),
        position: emp?.job_position?.title || '',
        branch: emp?.branch?.name || '',
        type_label: overtimeTypes.find((t) => t.value === item.type)?.label || 'Lembur',
        date_formatted: new Date(item.date).toLocaleDateString('id-ID', {
          day: '2-digit',
          month: 'long',
          year: 'numeric',
        }),
        start_time_only: extractTimeHHMM(item.start_time, '.'),
        end_time_only: extractTimeHHMM(item.end_time, '.'),
        duration_formatted: `${Math.floor(item.duration_minutes / 60)}j ${item.duration_minutes % 60}m`,
      }
    }

    const myRes = responses[0]
    if (myRes && myRes.data) {
      myOvertimes.value = (myRes.data.data || []).map(transform)
    }

    if (isAdmin.value && responses.length > 1) {
      const allRes = responses[1]
      if (allRes && allRes.data) {
        allOvertimes.value = (allRes.data.data || []).map(transform)
      }
    }
  } catch (e) {
    console.error('Failed to fetch overtime data', e)
    toast.error('Gagal mengambil data lembur')
  } finally {
    isLoading.value = false
  }
}

const calculateDuration = () => {
  if (requestForm.value.start_time && requestForm.value.end_time) {
    const start = new Date(`2000-01-01T${requestForm.value.start_time}`)
    const end = new Date(`2000-01-01T${requestForm.value.end_time}`)
    let diff = (end.getTime() - start.getTime()) / (1000 * 60)
    if (diff < 0) diff += 24 * 60 // Handle overnight
    requestForm.value.duration_minutes = Math.max(0, diff)
  }
}

const handleFileChange = (e: Event) => {
  const target = e.target as HTMLInputElement
  if (target.files && target.files.length > 0) {
    attachment.value = (target.files[0] as File)
  } else {
    attachment.value = null
  }
}

const openAddModal = () => {
  isEditMode.value = false
  requestForm.value = {
    id: '',
    date: new Date().toISOString().split('T')[0],
    start_time: '17:00',
    end_time: '19:00',
    duration_minutes: 120,
    type: 'working_day',
    reason: '',
    attachment_url: '',
  }
  attachment.value = null
  isModalOpen.value = true
}

const openEditModal = (item: any) => {
  isEditMode.value = true
  requestForm.value = {
    id: item.id,
    date: new Date(item.date).toISOString().split('T')[0],
    start_time: extractTimeHHMM(item.start_time, ':'),
    end_time: extractTimeHHMM(item.end_time, ':'),
    duration_minutes: item.duration_minutes,
    type: item.type || 'working_day',
    reason: item.reason || '',
    attachment_url: item.attachment_url || '',
  }
  attachment.value = null
  isModalOpen.value = true
}

const saveOvertime = async () => {
  if (!requestForm.value.reason.trim()) {
    toast.error('Alasan lembur harus diisi')
    return
  }

  isSubmitting.value = true
  try {
    const formData = new FormData()
    formData.append('date', (requestForm.value.date as string) || '')
    formData.append('start_time', `${requestForm.value.date}T${requestForm.value.start_time}:00`)
    formData.append('end_time', `${requestForm.value.date}T${requestForm.value.end_time}:00`)
    formData.append('duration_minutes', String(requestForm.value.duration_minutes || 0))
    formData.append('type', (requestForm.value.type as string) || '')
    formData.append('reason', (requestForm.value.reason as string) || '')
    
    if (attachment.value) {
      formData.append('attachment', attachment.value)
    }

    if (isEditMode.value) {
      await apiClient.put(`/overtimes/${requestForm.value.id}`, formData, {
        headers: { 'Content-Type': 'multipart/form-data' }
      })
      toast.success('Pengajuan lembur diperbarui')
    } else {
      await apiClient.post('/overtimes/', formData, {
        headers: { 'Content-Type': 'multipart/form-data' }
      })
      toast.success('Pengajuan lembur berhasil dikirim')
    }

    isModalOpen.value = false
    fetchOvertimes()
  } catch (e: any) {
    toast.error('Gagal menyimpan data: ' + (e.response?.data?.message || e.message))
  } finally {
    isSubmitting.value = false
  }
}

const deleteOvertime = async (id: string) => {
  if (!confirm('Hapus pengajuan lembur ini?')) return
  try {
    await apiClient.delete(`/overtimes/${id}`)
    toast.success('Pengajuan dihapus')
    fetchOvertimes()
  } catch (e: any) {
    toast.error('Gagal menghapus: ' + (e.response?.data?.message || e.message))
  }
}

const updateStatus = async (id: string, newStatus: string) => {
  if (newStatus === 'approved' && !confirm('Setujui lembur ini?')) return
  isSubmitting.value = true
  try {
    await apiClient.put(`/overtimes/${id}/status`, { status: newStatus })
    toast.success(`Lembur ${newStatus === 'approved' ? 'disetujui' : 'ditolak'}`)
    fetchOvertimes()
  } catch (e: any) {
    toast.error('Gagal update status: ' + (e.response?.data?.message || e.message))
  } finally {
    isSubmitting.value = false
  }
}

const openRejectModal = (id: string) => {
  selectedOvertimeId.value = id
  rejectReason.value = ''
  isRejectModalOpen.value = true
}

const confirmReject = async () => {
  if (!rejectReason.value.trim()) {
    toast.error('Alasan penolakan harus diisi')
    return
  }
  isSubmitting.value = true
  try {
    await apiClient.put(`/overtimes/${selectedOvertimeId.value}/status`, {
      status: 'rejected',
      reject_reason: rejectReason.value,
    })
    toast.success('Pengajuan ditolak')
    isRejectModalOpen.value = false
    fetchOvertimes()
  } catch (e: any) {
    toast.error('Gagal menolak: ' + (e.response?.data?.message || e.message))
  } finally {
    isSubmitting.value = false
  }
}

onMounted(() => {
  fetchOvertimes()
})
</script>

<template>
  <div class="space-y-6">
    <!-- Header -->
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
      <div>
        <h1 class="text-[24px] font-black text-slate-900 tracking-tight leading-tight italic">
          Manajemen Lembur
        </h1>
        <p class="text-[14px] text-slate-500 mt-1 font-medium italic">
          Kelola penugasan waktu kerja tambahan dan persetujuan.
        </p>
      </div>
      <div class="flex items-center gap-3">
        <Button
          @click="openAddModal"
          class="gap-2 shadow-lg shadow-primary/20 rounded-2xl h-11 px-6 font-bold"
          size="sm"
        >
          <Plus class="w-4 h-4" />
          Ajukan Lembur
        </Button>
      </div>
    </div>

    <!-- Table container -->
    <div
      class="bg-white rounded-[2.5rem] border border-slate-100 shadow-xl shadow-slate-200/50 overflow-hidden"
    >
      <DataTable
        :data="isAdmin ? allOvertimes : myOvertimes"
        :columns="columns"
        :isLoading="isLoading"
      />
    </div>

    <!-- Add/Edit Modal -->
    <Dialog v-model:open="isModalOpen">
      <DialogContent
        class="sm:max-w-xl rounded-[2.5rem] md:rounded-[3.5rem] p-0 overflow-hidden border-none shadow-2xl flex flex-col"
      >
        <div class="bg-slate-900 p-8 md:p-12 text-white relative shrink-0">
          <DialogHeader>
            <DialogTitle class="text-2xl md:text-3xl font-black tracking-tight italic">{{
              isEditMode ? 'Koreksi Data' : 'Buat Pengajuan'
            }}</DialogTitle>
            <DialogDescription
              class="text-slate-400 font-bold mt-2 uppercase text-[10px] md:text-[11px] tracking-[0.2em] border-l-4 border-primary pl-4 leading-relaxed"
            >
              Pastikan jam lembur sesuai dengan aktivitas pekerjaan aktual.
            </DialogDescription>
          </DialogHeader>
        </div>

        <div class="p-8 md:p-12 bg-white space-y-8 overflow-y-auto max-h-[60vh] custom-scrollbar">
          <div class="grid grid-cols-2 gap-6">
            <div class="space-y-2">
              <label class="text-[12px] font-black text-slate-400 uppercase tracking-widest ml-1"
                >Tanggal</label
              >
              <div class="relative">
                <Calendar class="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400" />
                <Input
                  v-model="requestForm.date"
                  type="date"
                  class="pl-11 h-12 rounded-2xl bg-slate-50 border-none font-bold"
                />
              </div>
            </div>
            <div class="space-y-2">
              <label class="text-[12px] font-black text-slate-400 uppercase tracking-widest ml-1"
                >Tipe Hari</label
              >
              <Select v-model="requestForm.type">
                <SelectTrigger
                  class="h-12 rounded-2xl bg-slate-50 border-none font-bold focus:ring-0"
                >
                  <SelectValue placeholder="Pilih Tipe" />
                </SelectTrigger>
                <SelectContent class="rounded-xl border-slate-100 shadow-xl">
                  <SelectItem
                    v-for="t in overtimeTypes"
                    :key="t.value"
                    :value="t.value"
                    class="font-bold py-3"
                    >{{ t.label }}</SelectItem
                  >
                </SelectContent>
              </Select>
            </div>
          </div>

          <div class="grid grid-cols-2 gap-6">
            <div class="space-y-2">
              <label class="text-[12px] font-black text-slate-400 uppercase tracking-widest ml-1"
                >Mulai</label
              >
              <Input
                v-model="requestForm.start_time"
                type="time"
                @input="calculateDuration"
                class="h-12 rounded-2xl bg-slate-50 border-none font-bold"
              />
            </div>
            <div class="space-y-2">
              <label class="text-[12px] font-black text-slate-400 uppercase tracking-widest ml-1"
                >Selesai</label
              >
              <Input
                v-model="requestForm.end_time"
                type="time"
                @input="calculateDuration"
                class="h-12 rounded-2xl bg-slate-50 border-none font-bold"
              />
            </div>
          </div>

          <div
            class="p-6 bg-primary/5 rounded-[2.5rem] border border-primary/10 flex items-center justify-between"
          >
            <div class="flex items-center gap-4">
              <div class="bg-primary/10 p-3 rounded-2xl text-primary shadow-sm shadow-primary/10">
                <Clock class="w-6 h-6" />
              </div>
              <div>
                <p class="text-[11px] font-black text-primary/60 uppercase tracking-[0.2em] mb-1">
                  Durasi Kalkulasi
                </p>
                <p class="text-xl font-black text-primary leading-none">
                  {{ Math.floor(requestForm.duration_minutes / 60) }}j
                  {{ requestForm.duration_minutes % 60 }}m
                </p>
              </div>
            </div>
            <div class="text-right">
              <span class="text-[10px] font-bold text-slate-400 italic">Terhitung Otomatis</span>
            </div>
          </div>

          <div class="space-y-2">
            <label class="text-[12px] font-black text-slate-400 uppercase tracking-widest ml-1"
              >Tugas / Alasan Lembur</label
            >
            <div class="relative group">
              <Info
                class="absolute left-4 top-4 w-4 h-4 text-slate-300 group-focus-within:text-primary transition-colors duration-300"
              />
              <textarea
                v-model="requestForm.reason"
                placeholder="Deskripsikan pekerjaan yang Anda selesaikan..."
                class="w-full min-h-[120px] pl-11 pr-4 py-4 rounded-3xl bg-slate-50 border-none font-bold text-slate-700 placeholder:text-slate-300 focus:ring-4 focus:ring-primary/5 transition-all outline-none resize-none"
              ></textarea>
            </div>
          </div>

          <div class="space-y-2">
            <label class="text-[12px] font-black text-slate-400 uppercase tracking-widest ml-1"
              >Lampiran Bukti (Opsional)</label
            >
            <div class="relative group space-y-3">
              <!-- Existing Image Preview -->
              <div v-if="isEditMode && requestForm.attachment_url" class="p-3 bg-slate-50 border border-slate-100 rounded-2xl flex items-center gap-3">
                <div class="w-10 h-10 rounded-xl bg-indigo-100 text-indigo-600 flex items-center justify-center">
                  <Image class="w-5 h-5" />
                </div>
                <div class="flex-1">
                  <p class="text-xs font-bold text-slate-700">Lampiran Tersimpan</p>
                  <a :href="resolvePhotoUrl(requestForm.attachment_url)" target="_blank" class="text-[11px] text-indigo-600 hover:underline">Lihat Gambar Saat Ini</a>
                </div>
              </div>

              <!-- File Input -->
              <div class="relative">
                <Image class="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400" />
                <Input
                  type="file"
                  accept="image/*"
                  @change="handleFileChange"
                  class="pl-11 h-12 rounded-2xl bg-slate-50 border-none font-bold pt-3 w-full"
                />
              </div>
              <p v-if="isEditMode && requestForm.attachment_url" class="text-[10px] text-slate-400 font-bold ml-1 italic">
                * Pilih file baru jika ingin mengganti lampiran yang sudah ada.
              </p>
            </div>
          </div>
        </div>

        <DialogFooter
          class="p-8 md:p-12 pt-0 bg-white grid grid-cols-2 gap-4 shrink-0 border-t border-slate-50 pt-8"
        >
          <Button
            variant="ghost"
            @click="isModalOpen = false"
            :disabled="isSubmitting"
            class="rounded-2xl h-14 font-black text-slate-400 border border-slate-100 hover:bg-slate-50 transition-all"
            >BATAL</Button
          >
          <Button
            @click="saveOvertime"
            :disabled="isSubmitting"
            class="bg-primary hover:bg-primary/90 text-white rounded-2xl h-14 font-black shadow-xl shadow-primary/20 transform active:scale-95 transition-all uppercase tracking-widest"
          >
            {{ isSubmitting ? 'MENYIMPAN...' : 'SIMPAN DATA' }}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>

    <!-- Reject Reason Modal -->
    <Dialog v-model:open="isRejectModalOpen">
      <DialogContent
        class="sm:max-w-md rounded-[2.5rem] p-0 overflow-hidden border-none shadow-2xl flex flex-col"
      >
        <div class="bg-rose-900 p-8 text-white relative shrink-0">
          <DialogHeader>
            <DialogTitle class="text-xl font-black uppercase tracking-tighter italic"
              >Tolak Pengajuan</DialogTitle
            >
            <DialogDescription
              class="text-rose-300 font-bold mt-2 uppercase text-[10px] tracking-widest border-l-4 border-white/30 pl-4"
            >
              Wajib menyertakan alasan penolakan.
            </DialogDescription>
          </DialogHeader>
        </div>
        <div class="p-8 bg-white space-y-6">
          <div class="space-y-2">
            <label class="text-[12px] font-black text-slate-400 uppercase tracking-widest ml-1"
              >Alasan Penolakan</label
            >
            <textarea
              v-model="rejectReason"
              placeholder="Contoh: Belum ada koordinasi supervisor, tugas bisa diselesaikan di jam reguler..."
              class="w-full min-h-[100px] px-5 py-4 rounded-3xl bg-slate-50 border-none font-bold text-slate-700 outline-none focus:ring-4 focus:ring-rose-500/5 transition-all resize-none"
            ></textarea>
          </div>
        </div>
        <DialogFooter class="p-8 pt-0 bg-white grid grid-cols-2 gap-4">
          <Button
            variant="ghost"
            @click="isRejectModalOpen = false"
            class="rounded-2xl h-14 font-black text-slate-400 border border-slate-100"
            >BATAL</Button
          >
          <Button
            @click="confirmReject"
            :disabled="isSubmitting || !rejectReason.trim()"
            class="bg-rose-600 hover:bg-rose-700 text-white rounded-2xl h-14 font-black shadow-xl shadow-rose-600/20 uppercase tracking-widest"
          >
            KONFIRMASI
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>

    <!-- ─────────── ATTACHMENT VIEWER MODAL ─────────── -->
    <Dialog v-model:open="isAttachmentViewerOpen">
      <DialogContent class="sm:max-w-2xl max-h-[90vh] flex flex-col p-6 rounded-[2rem] overflow-hidden bg-white shadow-2xl border-slate-100">
        <DialogHeader class="pb-3 border-b border-slate-100">
          <div class="flex items-center justify-between">
            <div class="space-y-1">
              <DialogTitle class="flex items-center gap-2 text-lg font-black text-slate-800">
                <Image class="w-5 h-5 text-indigo-600" />
                Lampiran Bukti Lembur
              </DialogTitle>
              <DialogDescription class="text-xs font-medium text-slate-400">
                Foto / bukti penyelesaian aktivitas lembur karyawan.
              </DialogDescription>
            </div>
            <a
              v-if="selectedAttachmentUrl"
              :href="selectedAttachmentUrl"
              target="_blank"
              class="inline-flex items-center gap-1.5 px-3 py-1.5 bg-slate-100 hover:bg-slate-200 text-slate-700 rounded-xl text-xs font-bold transition-all"
              title="Buka file asli di tab baru"
            >
              <ExternalLink class="w-3.5 h-3.5" />
              Tab Baru
            </a>
          </div>
        </DialogHeader>

        <div class="flex-1 my-3 min-h-[300px] max-h-[62vh] bg-slate-950 rounded-2xl overflow-hidden flex items-center justify-center p-4 relative border border-slate-800">
          <img
            v-show="!hasImageError"
            :src="selectedAttachmentUrl"
            class="max-w-full max-h-[58vh] object-contain rounded-lg transition-transform duration-200"
            alt="Bukti Lembur"
            @error="hasImageError = true"
          />
          <div v-if="hasImageError" class="text-center p-6 text-slate-300">
            <Info class="w-12 h-12 text-amber-400 mx-auto mb-3" />
            <p class="font-bold text-sm text-white">Lampiran Tidak Ditemukan (404)</p>
            <p class="text-xs text-slate-400 mt-1 max-w-sm mx-auto">
              File lampiran fisik belum ada atau hilang akibat pembaruan sistem sebelumnya. Silakan unggah ulang bukti melalui tombol <b>Edit</b>.
            </p>
          </div>
        </div>

        <DialogFooter class="pt-2 border-t border-slate-100 flex justify-end">
          <Button
            variant="outline"
            class="rounded-xl font-bold text-xs h-10 px-5 border-slate-200"
            @click="isAttachmentViewerOpen = false"
          >
            Tutup
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  </div>
</template>

<style scoped>
.custom-scrollbar::-webkit-scrollbar {
  width: 6px;
}
.custom-scrollbar::-webkit-scrollbar-track {
  background: #f8fafc;
}
.custom-scrollbar::-webkit-scrollbar-thumb {
  background: #e2e8f0;
  border-radius: 10px;
}
.custom-scrollbar::-webkit-scrollbar-thumb:hover {
  background: #cbd5e1;
}
</style>
