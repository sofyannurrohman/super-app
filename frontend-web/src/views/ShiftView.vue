<script setup lang="ts">
import { ref, h, onMounted } from 'vue'
import { Pencil, Trash2 } from 'lucide-vue-next'
import apiClient from '@/api/axios'
import DataTable from '@/components/DataTable.vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { toast } from 'vue-sonner'

import { useMasterDataStore } from '@/stores/masterData'

const masterData = useMasterDataStore()

const isModalOpen = ref(false)
const newShift = ref({
  id: '',
  name: '',
  startTime: '',
  endTime: '',
  breakStart: '',
  breakEnd: '',
  branchId: '',
  isFlexible: false,
})

const isEditMode = ref(false)
const isSubmitting = ref(false)
const shifts = ref<any[]>([])
const isLoading = ref(true)

const openAddModal = () => {
  isEditMode.value = false
  newShift.value = {
    id: '',
    name: '',
    startTime: '',
    endTime: '',
    breakStart: '',
    breakEnd: '',
    branchId: '',
    isFlexible: false,
  }
  isModalOpen.value = true
}

const formatTime = (timeStr: string, isEndTime = false) => {
  if (!timeStr) return '-'

  // Handle HH:mm:ss format
  if (timeStr.includes(':') && !timeStr.includes('-') && !timeStr.includes('T')) {
    const parts = timeStr.split(':')
    const h = (parts[0] || '00').padStart(2, '0')
    const m = (parts[1] || '00').padStart(2, '0')

    if (isEndTime && h === '00' && m === '00') {
      return '24.00'
    }

    return `${h}.${m}`
  }

  const d = new Date(timeStr)
  if (isNaN(d.getTime())) return '-'
  let h = d.getUTCHours()
  let m = d.getUTCMinutes()

  if (isEndTime && h === 0 && m === 0) {
    return '24.00'
  }

  return `${String(h).padStart(2, '0')}.${String(m).padStart(2, '0')}`
}

const formatTimeForInput = (timeStr: string, isEndTime = false) => {
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
  const h = d.getUTCHours()
  const m = d.getUTCMinutes()
  if (isEndTime && h === 0 && m === 0) return '24.00'
  return `${String(h).padStart(2, '0')}.${String(m).padStart(2, '0')}`
}

const openEditModal = (shift: any) => {
  isEditMode.value = true

  newShift.value = {
    id: shift?.id || '',
    name: shift?.name || '',
    startTime: formatTimeForInput(shift?.start_time || ''),
    endTime: formatTimeForInput(shift?.end_time || '', true),
    breakStart: shift?.break_start ? formatTimeForInput(shift.break_start) : '',
    breakEnd: shift?.break_end ? formatTimeForInput(shift.break_end, true) : '',
    branchId: shift?.branch_id || 'all',
    isFlexible: shift?.is_flexible || false,
  }
  isModalOpen.value = true
}

const closeAddModal = () => {
  isModalOpen.value = false
}

const fetchShifts = async () => {
  isLoading.value = true
  try {
    const response = await apiClient.get('/shifts')
    shifts.value = response.data.data
  } catch (error) {
    console.error('Failed to fetch shifts:', error)
  } finally {
    isLoading.value = false
  }
}

const saveShift = async () => {
  isSubmitting.value = true
  try {
    const formatTimeForBackend = (timeStr: string) => {
      if (!timeStr) return null
      // Convert 08.30 to 08:30:00
      let formatted = timeStr.trim().replace('.', ':')

      // Ensure HH:mm format
      if (formatted.includes(':')) {
        const [hour, minute] = formatted.split(':')
        const hh = (hour || '00').padStart(2, '0')
        const mm = (minute || '00').padEnd(2, '0')

        // Handle 24.00 as 00:00:00
        if (hh === '24' && mm === '00') {
          return '00:00:00'
        }

        return `${hh}:${mm}:00`
      }

      // Handle plain numbers like "8" -> "08:00:00"
      if (/^\d+$/.test(formatted)) {
        return `${formatted.padStart(2, '0')}:00:00`
      }

      return formatted
    }

    const payload = {
      name: newShift.value.name,
      start_time: formatTimeForBackend(newShift.value.startTime),
      end_time: formatTimeForBackend(newShift.value.endTime),
      break_start: formatTimeForBackend(newShift.value.breakStart),
      break_end: formatTimeForBackend(newShift.value.breakEnd),
      branch_id:
        newShift.value.branchId === 'all' || !newShift.value.branchId
          ? null
          : newShift.value.branchId,
      is_flexible: newShift.value.isFlexible,
    }

    if (isEditMode.value) {
      await apiClient.put(`/shifts/${newShift.value.id}`, payload)
      toast.success('Shift berhasil diperbarui!')
    } else {
      await apiClient.post('/shifts', payload)
      toast.success('Shift berhasil ditambahkan!')
    }
    closeAddModal()
    fetchShifts()
  } catch (error: any) {
    toast.error(
      `Gagal ${isEditMode.value ? 'memperbarui' : 'menambahkan'} shift: ` +
        (error.response?.data?.error || error.message),
    )
  } finally {
    isSubmitting.value = false
  }
}

const deleteShift = async (id: string) => {
  if (!confirm('Apakah Anda yakin ingin menghapus shift ini? Operasi ini tidak dapat dibatalkan.'))
    return

  try {
    await apiClient.delete(`/shifts/${id}`)
    toast.success('Shift berhasil dihapus!')
    fetchShifts()
  } catch (error: any) {
    toast.error('Gagal menghapus shift: ' + (error.response?.data?.error || error.message))
  }
}

onMounted(() => {
  fetchShifts()
  masterData.fetchBranches()
})

const columns = [
  {
    accessorKey: 'name',
    header: 'NAMA SHIFT',
    cell: (info: any) =>
      h(
        'span',
        { class: 'font-bold text-gray-900 uppercase tracking-tight' },
        info.getValue() || '-',
      ),
  },
  {
    accessorFn: (row: any) => row.branch?.name || 'Semua Cabang',
    id: 'branch',
    header: 'CABANG',
    cell: (info: any) => h('span', { class: 'text-gray-500' }, info.getValue()),
  },
  {
    accessorFn: (row: any) => `${formatTime(row.start_time)} - ${formatTime(row.end_time, true)}`,
    id: 'waktu',
    header: 'JAM KERJA',
    cell: ({ getValue }: any) => h('span', { class: 'text-gray-600' }, getValue()),
  },
  {
    accessorFn: (row: any) => {
      if (!row.break_start || !row.break_end) return 'Tidak Ada'
      return `${formatTime(row.break_start)} - ${formatTime(row.break_end, true)}`
    },
    id: 'istirahat',
    header: 'ISTIRAHAT',
    cell: ({ getValue }: any) => h('span', { class: 'text-gray-600 text-[13px]' }, getValue()),
  },
  {
    accessorKey: 'is_flexible',
    header: 'FLEXIBLE',
    cell: (info: any) => {
      const isFlex = info.getValue()
      return h(
        'span',
        {
          class: isFlex
            ? 'border border-[#bbf7d0] text-[#16a34a] px-3 py-1 rounded-full text-[11px] font-bold inline-block'
            : 'border border-[#e5e7eb] text-gray-600 px-3 py-1 rounded-full text-[11px] font-bold inline-block',
        },
        isFlex ? 'Ya' : 'Tidak',
      )
    },
  },
  {
    id: 'actions',
    header: 'AKSI',
    cell: ({ row }: any) => {
      const shift = row.original
      return h('div', { class: 'flex items-center gap-2' }, [
        h(
          Button,
          {
            variant: 'ghost',
            size: 'sm',
            class: 'h-8 px-2 text-primary hover:text-primary hover:bg-primary/5',
            onClick: () => openEditModal(shift),
          },
          () => h(Pencil, { class: 'w-4 h-4' }),
        ),
        h(
          Button,
          {
            variant: 'ghost',
            size: 'sm',
            class: 'h-8 px-2 text-red-600 hover:text-red-700 hover:bg-red-50',
            onClick: () => deleteShift(shift.id),
          },
          () => h(Trash2, { class: 'w-4 h-4' }),
        ),
      ])
    },
    enableSorting: false,
  },
]
</script>

<template>
  <div class="space-y-6">
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
      <div>
        <h1 class="text-[22px] font-bold text-gray-900 leading-tight">Master Shift Kerja</h1>
        <p class="text-[14px] text-gray-500 mt-1">
          Kelola data jadwal dan aturan waktu shift kerja.
        </p>
      </div>

      <div class="flex items-center gap-3">
        <Button @click="openAddModal" class="gap-2 shadow-sm rounded-lg" size="sm">
          <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2"
              d="M12 4v16m8-8H4"
            />
          </svg>
          Tambah Shift
        </Button>
      </div>
    </div>

    <DataTable :data="shifts" :columns="columns" :isLoading="isLoading" />

    <!-- Add/Edit Modal -->
    <Dialog v-model:open="isModalOpen">
      <DialogContent
        class="sm:max-w-xl rounded-[2rem] md:rounded-[3rem] p-0 overflow-hidden border-none shadow-2xl max-h-[95vh] flex flex-col"
      >
        <div class="bg-slate-900 p-8 md:p-10 text-white relative shrink-0">
          <DialogHeader>
            <DialogTitle class="text-xl md:text-2xl font-black tracking-tight">{{
              isEditMode ? 'Edit Shift Kerja' : 'Tambah Shift Kerja'
            }}</DialogTitle>
            <DialogDescription
              class="text-slate-400 font-bold mt-2 uppercase text-[10px] md:text-[11px] tracking-widest border-l-4 border-primary pl-4 leading-relaxed"
            >
              Tentukan konfigurasi waktu untuk shift ini.
            </DialogDescription>
          </DialogHeader>
        </div>

        <div class="p-6 md:p-10 bg-white space-y-6 overflow-y-auto custom-scrollbar flex-1">
          <div class="grid gap-2">
            <label class="text-[13px] font-medium text-gray-700">Nama Shift</label>
            <Input v-model="newShift.name" placeholder="e.g. Shift Pagi" />
          </div>

          <div class="grid gap-2">
            <label class="text-[13px] font-medium text-gray-700">Berlaku di Cabang</label>
            <Select v-model="newShift.branchId">
              <SelectTrigger>
                <SelectValue placeholder="Pilih Cabang" />
              </SelectTrigger>
              <SelectContent>
                <SelectGroup>
                  <SelectItem value="all">Semua Cabang (Global)</SelectItem>
                  <SelectItem v-for="b in masterData.branches" :key="b.id" :value="b.id">
                    {{ b.name }}
                  </SelectItem>
                </SelectGroup>
              </SelectContent>
            </Select>
          </div>

          <div class="grid grid-cols-2 gap-4">
            <div class="grid gap-2">
              <div class="flex items-center justify-between">
                <label class="text-[13px] font-medium text-gray-700">Jam Masuk (Start)</label>
                <span class="text-[10px] text-gray-400 font-bold uppercase tracking-wider"
                  >Format: 08.30</span
                >
              </div>
              <Input v-model="newShift.startTime" placeholder="08.30" class="h-10" />
            </div>
            <div class="grid gap-2">
              <div class="flex items-center justify-between">
                <label class="text-[13px] font-medium text-gray-700">Jam Keluar (End)</label>
                <span class="text-[10px] text-gray-400 font-bold uppercase tracking-wider"
                  >Format: 24.00</span
                >
              </div>
              <Input v-model="newShift.endTime" placeholder="17.00 atau 24.00" class="h-10" />
            </div>
          </div>

          <div class="h-px bg-gray-100 my-2"></div>
          <h3 class="text-[13px] font-bold text-gray-900 uppercase tracking-wider">
            Jam Istirahat (Opsional)
          </h3>

          <div class="grid grid-cols-2 gap-4">
            <div class="grid gap-2">
              <div class="flex items-center justify-between">
                <label class="text-[13px] font-medium text-gray-700">Mulai Istirahat</label>
                <span class="text-[10px] text-gray-400 font-bold uppercase tracking-wider"
                  >Format: 12.00</span
                >
              </div>
              <Input v-model="newShift.breakStart" placeholder="12.00" class="h-10" />
            </div>
            <div class="grid gap-2">
              <div class="flex items-center justify-between">
                <label class="text-[13px] font-medium text-gray-700">Selesai Istirahat</label>
                <span class="text-[10px] text-gray-400 font-bold uppercase tracking-wider"
                  >Format: 13.00</span
                >
              </div>
              <Input v-model="newShift.breakEnd" placeholder="13.00" class="h-10" />
            </div>
          </div>

          <div class="flex items-center space-x-2 mt-4">
            <input
              type="checkbox"
              id="isFlexible"
              v-model="newShift.isFlexible"
              class="w-4 h-4 rounded border-gray-300 text-primary focus:ring-primary"
            />
            <label for="isFlexible" class="text-sm font-medium leading-none">
              Shift Fleksibel (Boleh absen kapan saja selagi 8 jam)
            </label>
          </div>
        </div>

        <DialogFooter class="p-6 md:p-10 pt-0 bg-white grid grid-cols-2 gap-4 shrink-0">
          <Button
            variant="ghost"
            @click="closeAddModal"
            :disabled="isSubmitting"
            class="rounded-2xl h-12 md:h-14 font-black text-slate-400 border border-slate-100 hover:bg-slate-50 transition-all"
            >BATAL</Button
          >
          <Button
            @click="saveShift"
            :disabled="isSubmitting"
            class="bg-primary hover:bg-primary/90 text-white rounded-2xl h-12 md:h-14 font-black shadow-xl shadow-primary/20 transform active:scale-95 transition-all"
          >
            {{ isSubmitting ? 'MENYIMPAN...' : 'SIMPAN DATA' }}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  </div>
</template>
