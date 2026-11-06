<script setup lang="ts">
import { ref, h, onMounted } from 'vue'
import { Pencil, Trash2 } from 'lucide-vue-next'
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
import { toast } from 'vue-sonner'

const isModalOpen = ref(false)
const newPosition = ref({
  id: '',
  title: '',
  level: 1,
})

const isEditMode = ref(false)

const isSubmitting = ref(false)
const positions = ref<any[]>([])
const isLoading = ref(true)

const openAddModal = () => {
  isEditMode.value = false
  newPosition.value = {
    id: '',
    title: '',
    level: 1,
  }
  isModalOpen.value = true
}

const openEditModal = (position: any) => {
  isEditMode.value = true
  newPosition.value = {
    id: position.id,
    title: position.title || '',
    level: position.level || 1,
  }
  isModalOpen.value = true
}

const closeAddModal = () => {
  isModalOpen.value = false
}

const fetchPositions = async () => {
  isLoading.value = true
  try {
    const response = await apiClient.get('/job-positions')
    positions.value = response.data.data
  } catch (error) {
    console.error('Failed to fetch job positions:', error)
  } finally {
    isLoading.value = false
  }
}

const savePosition = async () => {
  isSubmitting.value = true
  try {
    if (isEditMode.value) {
      await apiClient.put(`/job-positions/${newPosition.value.id}`, newPosition.value)
      toast.success('Data Jabatan berhasil diperbarui!')
    } else {
      await apiClient.post('/job-positions', newPosition.value)
      toast.success('Jabatan berhasil ditambahkan!')
    }
    closeAddModal()
    fetchPositions() // Refresh data
  } catch (error: any) {
    toast.error(
      `Gagal ${isEditMode.value ? 'memperbarui' : 'menambahkan'} jabatan: ` +
        (error.response?.data?.error || error.message),
    )
  } finally {
    isSubmitting.value = false
  }
}

const deletePosition = async (id: string) => {
  if (
    !confirm('Apakah Anda yakin ingin menghapus jabatan ini? Operasi ini tidak dapat dibatalkan.')
  )
    return

  try {
    await apiClient.delete(`/job-positions/${id}`)
    toast.success('Jabatan berhasil dihapus!')
    fetchPositions()
  } catch (error: any) {
    toast.error('Gagal menghapus jabatan: ' + (error.response?.data?.error || error.message))
  }
}

onMounted(() => {
  fetchPositions()
})

const columns = [
  {
    accessorKey: 'title',
    header: 'NAMA JABATAN',
    cell: (info: any) => h('span', { class: 'font-bold text-gray-900' }, info.getValue() || '-'),
  },
  {
    accessorKey: 'level',
    header: 'LEVEL',
    cell: (info: any) => h('span', { class: 'text-gray-500 font-medium' }, info.getValue()),
  },
  {
    accessorFn: (row: any) =>
      row.created_at ? new Date(row.created_at).toLocaleDateString() : '-',
    id: 'createdAt',
    header: 'TANGGAL DIBUAT',
    cell: (info: any) => h('span', { class: 'text-gray-500 text-[13px]' }, info.getValue()),
  },
  {
    id: 'actions',
    header: 'AKSI',
    cell: ({ row }: any) => {
      const position = row.original
      return h('div', { class: 'flex items-center gap-2' }, [
        h(
          Button,
          {
            variant: 'ghost',
            size: 'sm',
            class: 'h-8 px-2 text-primary hover:text-primary hover:bg-primary/5',
            onClick: () => openEditModal(position),
          },
          () => h(Pencil, { class: 'w-4 h-4' }),
        ),
        h(
          Button,
          {
            variant: 'ghost',
            size: 'sm',
            class: 'h-8 px-2 text-red-600 hover:text-red-700 hover:bg-red-50',
            onClick: () => deletePosition(position.id),
          },
          () => h(Trash2, { class: 'w-4 h-4' }),
        ),
      ])
    },
  },
]
</script>

<template>
  <div class="space-y-6">
    <!-- Header Page -->
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
      <div>
        <h1 class="text-[22px] font-bold text-gray-900 leading-tight">Master Jabatan</h1>
        <p class="text-[14px] text-gray-500 mt-1">Kelola data posisi/jabatan organisasi Anda.</p>
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
          Tambah Jabatan
        </Button>
      </div>
    </div>

    <!-- DataTable Instance -->
    <DataTable :data="positions" :columns="columns" :isLoading="isLoading" />

    <!-- Add/Edit Modal -->
    <Dialog v-model:open="isModalOpen">
      <DialogContent
        class="sm:max-w-xl rounded-[2rem] md:rounded-[3rem] p-0 overflow-hidden border-none shadow-2xl max-h-[95vh] flex flex-col"
      >
        <div class="bg-slate-900 p-8 md:p-10 text-white relative shrink-0">
          <DialogHeader>
            <DialogTitle class="text-xl md:text-2xl font-black">{{
              isEditMode ? 'Edit Jabatan' : 'Tambah Jabatan'
            }}</DialogTitle>
            <DialogDescription
              class="text-slate-400 font-bold mt-2 uppercase text-[10px] md:text-[11px] tracking-widest border-l-4 border-primary pl-4"
            >
              Lengkapi data posisi/jabatan di organisasi Anda.
            </DialogDescription>
          </DialogHeader>
        </div>

        <div class="p-6 md:p-10 bg-white space-y-6 overflow-y-auto custom-scrollbar flex-1">
          <div class="grid gap-2">
            <label class="text-[13px] font-medium text-gray-700">Nama Jabatan</label>
            <Input v-model="newPosition.title" placeholder="e.g. Software Engineer" />
          </div>
          <div class="grid gap-2">
            <label class="text-[13px] font-medium text-gray-700">Level</label>
            <Input v-model.number="newPosition.level" type="number" placeholder="e.g. 1" min="1" />
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
            @click="savePosition"
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
