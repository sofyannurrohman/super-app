<script setup lang="ts">
import { ref, onMounted, computed, h } from 'vue'
import { toast } from 'vue-sonner'
import apiClient from '@/api/axios'
import { useAuthStore } from '@/stores/auth'
import DataTable from '@/components/DataTable.vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { AlertTriangle, CheckCircle2, Search, Plus, Trash2, Edit } from 'lucide-vue-next'

const authStore = useAuthStore()

const activeTab = ref<'system' | 'manual'>('system')

// --- DATA ---
const isLoading = ref(false)
const isLoadingManual = ref(false)
const problematicEmployees = ref<any[]>([])
const manualWarnings = ref<any[]>([])
const allEmployees = ref<any[]>([])
const searchQuery = ref('')
const searchManualQuery = ref('')
const employeeSearch = ref('')

const fetchProblematicEmployees = async () => {
  isLoading.value = true
  try {
    const res = await apiClient.get('/hr/problematic-employees')
    problematicEmployees.value = res.data.data
  } catch (e: any) {
    toast.error('Gagal mengambil data sistem: ' + (e.response?.data?.message || e.message))
  } finally {
    isLoading.value = false
  }
}

const fetchManualWarnings = async () => {
  isLoadingManual.value = true
  try {
    const res = await apiClient.get('/hr/warnings')
    manualWarnings.value = res.data.data
  } catch (e: any) {
    toast.error('Gagal mengambil data riwayat peringatan: ' + (e.response?.data?.message || e.message))
  } finally {
    isLoadingManual.value = false
  }
}

const fetchAllEmployees = async () => {
  try {
    const res = await apiClient.get('/employees?limit=all&status=active')
    allEmployees.value = res.data.data
  } catch (e: any) {
    console.error('Failed to fetch employees', e)
  }
}

onMounted(() => {
  fetchProblematicEmployees()
  fetchManualWarnings()
  fetchAllEmployees()
})

const filteredSystemData = computed(() => {
  if (!searchQuery.value) return problematicEmployees.value
  const q = searchQuery.value.toLowerCase()
  return problematicEmployees.value.filter(
    (emp) => emp.name.toLowerCase().includes(q) || emp.employee_id.toLowerCase().includes(q)
  )
})

const filteredManualData = computed(() => {
  if (!searchManualQuery.value) return manualWarnings.value
  const q = searchManualQuery.value.toLowerCase()
  return manualWarnings.value.filter(
    (emp) => emp.name.toLowerCase().includes(q) || emp.employee_id.toLowerCase().includes(q)
  )
})

const filteredAllEmployees = computed(() => {
  if (!employeeSearch.value) return allEmployees.value
  const q = employeeSearch.value.toLowerCase()
  return allEmployees.value.filter(
    (emp) => (emp.first_name + ' ' + emp.last_name).toLowerCase().includes(q) || 
             emp.employee_id_number.toLowerCase().includes(q)
  )
})

// --- COLUMNS ---
const systemColumns = [
  { accessorKey: 'employee_id', header: 'ID' },
  { accessorKey: 'name', header: 'Nama Karyawan' },
  { accessorKey: 'role', header: 'Jabatan' },
  { accessorKey: 'branch', header: 'Cabang' },
  { accessorKey: 'issue', header: 'Keterangan Masalah' },
  {
    accessorKey: 'status',
    header: 'Status Terkini',
    cell: ({ row }: any) => {
      const status = row.original.status
      if (status === 'pending') {
        return h('span', { class: 'inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-semibold bg-yellow-100 text-yellow-800' }, 'Belum Ditangani')
      }
      
      let badgeClass = 'bg-gray-100 text-gray-800'
      let label = status.toUpperCase()

      if (status === 'cs_followup') {
        badgeClass = 'bg-blue-100 text-blue-800'
        label = 'Level 1: Tindak Lanjut CS'
      } else if (status === 'spv_followup') {
        badgeClass = 'bg-orange-100 text-orange-800'
        label = 'Level 2: Tindak Lanjut SPV'
      } else if (status === 'teguran') {
        badgeClass = 'bg-red-100 text-red-800'
        label = 'Teguran'
      } else if (status.startsWith('sp')) {
        badgeClass = 'bg-red-200 text-red-900'
        label = `Level 3: ${status.toUpperCase()}`
      }

      return h('span', { class: `inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-semibold ${badgeClass}` }, label)
    }
  },
  {
    accessorKey: 'handled_by',
    header: 'Ditangani Oleh',
    cell: ({ row }: any) => row.original.handled_by || '-'
  },
  {
    accessorKey: 'actions',
    header: 'Aksi',
    cell: ({ row }: any) => {
      const data = row.original
      const role = authStore.user?.role
      const actionButtons = []

      // Tombol CS (Tampil jika masih pending)
      if ((role === 'superadmin' || role === 'customer_support') && data.status === 'pending') {
        actionButtons.push(
          h(Button, { size: 'sm', class: 'bg-blue-500 mr-2 hover:bg-blue-600', onClick: () => openDirectActionModal(data, 'cs_followup') }, () => 'Tindak Lanjut CS')
        )
      }
      // Tombol SPV (Tampil jika pending atau cs_followup)
      if ((role === 'superadmin' || role === 'supervisor_sales' || role === 'supervisor_sales_mt') && (data.status === 'pending' || data.status === 'cs_followup')) {
        actionButtons.push(
          h(Button, { size: 'sm', class: 'bg-orange-500 mr-2 hover:bg-orange-600', onClick: () => openDirectActionModal(data, 'spv_followup') }, () => 'Tindak Lanjut SPV')
        )
      }
      // Tombol HR (Bisa kapan saja)
      if (role === 'superadmin' || role === 'hr_admin') {
        actionButtons.push(
          h(Button, { size: 'sm', variant: 'destructive', onClick: () => openActionModal(data) }, () => 'Tindak Lanjut HR')
        )
      }

      // Jika sudah tidak ada tombol yang bisa dirender dan status tidak pending
      if (actionButtons.length === 0 && data.status !== 'pending') {
        return h('div', { class: 'flex items-center text-green-600 text-sm font-medium' }, [
          h(CheckCircle2, { class: 'w-4 h-4 mr-1' }), 'Sedang Proses / Selesai'
        ])
      }

      return h('div', { class: 'flex items-center' }, actionButtons)
    }
  }
]

const manualColumns = [
  { accessorKey: 'employee_id', header: 'ID' },
  { accessorKey: 'name', header: 'Nama Karyawan' },
  { accessorKey: 'role', header: 'Jabatan' },
  { 
    accessorKey: 'warning_type', 
    header: 'Jenis Peringatan', 
    cell: ({ row }: any) => {
      const wType = row.original.warning_type
      if (wType === 'cs_followup') return 'Level 1: Tindak Lanjut CS'
      if (wType === 'spv_followup') return 'Level 2: Tindak Lanjut SPV'
      if (wType === 'teguran') return 'Teguran'
      if (wType.startsWith('sp')) return `Level 3: ${wType.toUpperCase()}`
      return wType.toUpperCase()
    } 
  },
  { accessorKey: 'notes', header: 'Catatan' },
  { accessorKey: 'handled_by', header: 'Oleh' },
  {
    accessorKey: 'is_active',
    header: 'Status Aktif',
    cell: ({ row }: any) => {
      const isActive = row.original.is_active
      return h(
        'button',
        {
          class: `relative inline-flex h-5 w-9 shrink-0 cursor-pointer items-center justify-center rounded-full transition-colors duration-200 ease-in-out focus:outline-none focus:ring-2 focus:ring-green-600 focus:ring-offset-2 ${isActive ? 'bg-green-600' : 'bg-gray-200'}`,
          onClick: () => toggleWarningStatus(row.original.id),
        },
        [
          h('span', {
            'aria-hidden': 'true',
            class: `pointer-events-none inline-block h-4 w-4 transform rounded-full bg-white shadow ring-0 transition duration-200 ease-in-out ${isActive ? 'translate-x-4' : 'translate-x-0'}`
          })
        ]
      )
    }
  },
  {
    accessorKey: 'created_at',
    header: 'Tanggal',
    cell: ({ row }: any) => new Date(row.original.created_at).toLocaleDateString('id-ID')
  },
  {
    accessorKey: 'actions',
    header: 'Aksi',
    cell: ({ row }: any) => {
      const data = row.original
      return h('div', { class: 'flex items-center gap-2' }, [
        h(Button, {
          size: 'icon',
          variant: 'outline',
          class: 'h-8 w-8 text-blue-600',
          onClick: () => openEditModal(data)
        }, () => h(Edit, { class: 'w-4 h-4' })),
        h(Button, {
          size: 'icon',
          variant: 'outline',
          class: 'h-8 w-8 text-red-600',
          onClick: () => deleteWarning(data.id)
        }, () => h(Trash2, { class: 'w-4 h-4' }))
      ])
    }
  }
]

// --- MODALS ---
const isModalOpen = ref(false)
const modalMode = ref<'system' | 'manual_add' | 'manual_edit' | 'direct_action'>('system')
const selectedEmployee = ref<any>(null)
const selectedWarningId = ref<string | null>(null)
const actionForm = ref({
  employee_id: '',
  type: 'teguran',
  notes: ''
})
const isSubmitting = ref(false)

const openActionModal = (employee: any) => {
  modalMode.value = 'system'
  selectedEmployee.value = employee
  actionForm.value = { employee_id: employee.employee_id, type: 'teguran', notes: '' }
  isModalOpen.value = true
}

const openDirectActionModal = (employee: any, type: string) => {
  modalMode.value = 'direct_action'
  selectedEmployee.value = employee
  actionForm.value = { employee_id: employee.employee_id, type: type, notes: '' }
  isModalOpen.value = true
}

const openAddManualModal = () => {
  modalMode.value = 'manual_add'
  selectedEmployee.value = null
  selectedWarningId.value = null
  actionForm.value = { employee_id: '', type: 'teguran', notes: '' }
  employeeSearch.value = ''
  isModalOpen.value = true
}

const openEditModal = (warning: any) => {
  modalMode.value = 'manual_edit'
  selectedEmployee.value = warning
  selectedWarningId.value = warning.id
  actionForm.value = { 
    employee_id: warning.employee_id, 
    type: warning.warning_type, 
    notes: warning.notes 
  }
  isModalOpen.value = true
}

const closeActionModal = () => {
  isModalOpen.value = false
  
  setTimeout(() => {
    selectedEmployee.value = null
    selectedWarningId.value = null
    actionForm.value = { employee_id: '', type: 'teguran', notes: '' }
  }, 300)
}

const submitAction = async () => {
  isSubmitting.value = true
  try {
    if (modalMode.value === 'manual_edit') {
      await apiClient.put(`/hr/warnings/${selectedWarningId.value}`, {
        action_type: actionForm.value.type,
        notes: actionForm.value.notes
      })
      toast.success('Peringatan berhasil diperbarui!')
    } else {
      let empIdToSubmit = actionForm.value.employee_id
      if (modalMode.value === 'system' || modalMode.value === 'direct_action') empIdToSubmit = selectedEmployee.value.employee_id
      
      if (!empIdToSubmit) {
        toast.error('Pilih karyawan terlebih dahulu')
        isSubmitting.value = false
        return
      }

      await apiClient.post('/hr/warnings', {
        employee_id: empIdToSubmit,
        action_type: actionForm.value.type,
        notes: actionForm.value.notes
      })
      let successMsg = 'Tindakan'
      if (actionForm.value.type === 'cs_followup') successMsg = 'Tindak Lanjut CS'
      else if (actionForm.value.type === 'spv_followup') successMsg = 'Tindak Lanjut SPV Sales'
      else if (actionForm.value.type === 'sp1') successMsg = 'SP 1'
      else if (actionForm.value.type === 'sp2') successMsg = 'SP 2'
      else if (actionForm.value.type === 'sp3') successMsg = 'SP 3'
      else if (actionForm.value.type === 'teguran') successMsg = 'Teguran'
      toast.success(`${successMsg} berhasil dikirim!`)
    }
    
    await fetchProblematicEmployees()
    await fetchManualWarnings()
    closeActionModal()
  } catch (e: any) {
    toast.error('Gagal menyimpan: ' + (e.response?.data?.message || e.message))
  } finally {
    isSubmitting.value = false
  }
}

const deleteWarning = async (id: string) => {
  if (!confirm('Apakah Anda yakin ingin menghapus data ini?')) return
  try {
    await apiClient.delete(`/hr/warnings/${id}`)
    toast.success('Peringatan berhasil dihapus')
    await fetchProblematicEmployees()
    await fetchManualWarnings()
  } catch (e: any) {
    toast.error('Gagal menghapus: ' + (e.response?.data?.message || e.message))
  }
}

const toggleWarningStatus = async (id: string) => {
  try {
    await apiClient.put(`/hr/warnings/${id}/toggle`)
    toast.success('Status peringatan berhasil diubah')
    await fetchManualWarnings()
  } catch (e: any) {
    toast.error('Gagal merubah status: ' + (e.response?.data?.message || e.message))
  }
}
</script>

<template>
  <div class="space-y-6">
    <!-- Header -->
    <div class="flex flex-col md:flex-row justify-between items-start md:items-center gap-4">
      <div>
        <h1 class="text-2xl font-bold tracking-tight text-slate-900 flex items-center gap-2">
          <AlertTriangle class="h-6 w-6 text-red-500" />
          Manajemen Surat Peringatan
        </h1>
        <p class="text-slate-500 mt-1">
          Pantau daftar karyawan bermasalah dan kelola Surat Peringatan / Teguran.
        </p>
      </div>
      <Button v-if="activeTab === 'manual'" @click="openAddManualModal" class="bg-indigo-600 hover:bg-indigo-700 text-white shadow-sm">
        <Plus class="w-4 h-4 mr-2" />
        Tambah Manual
      </Button>
    </div>

    <!-- Tabs -->
    <div class="border-b border-slate-200">
      <nav class="-mb-px flex space-x-8" aria-label="Tabs">
        <button
          @click="activeTab = 'system'"
          :class="[
            activeTab === 'system'
              ? 'border-indigo-500 text-indigo-600'
              : 'border-transparent text-slate-500 hover:border-slate-300 hover:text-slate-700',
            'whitespace-nowrap border-b-2 py-4 px-1 text-sm font-medium'
          ]"
        >
          Deteksi Sistem Otomatis
        </button>
        <button
          @click="activeTab = 'manual'"
          :class="[
            activeTab === 'manual'
              ? 'border-indigo-500 text-indigo-600'
              : 'border-transparent text-slate-500 hover:border-slate-300 hover:text-slate-700',
            'whitespace-nowrap border-b-2 py-4 px-1 text-sm font-medium'
          ]"
        >
          Riwayat & Peringatan Manual
        </button>
      </nav>
    </div>

    <!-- Tab 1: System Detected -->
    <div v-if="activeTab === 'system'" class="space-y-4 animate-in fade-in duration-300">
      <div class="bg-white p-4 rounded-xl shadow-sm border border-slate-200 flex flex-col md:flex-row justify-between gap-4">
        <div class="relative w-full md:w-96">
          <Search class="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-slate-400" />
          <Input v-model="searchQuery" placeholder="Cari nama atau ID karyawan..." class="pl-9 w-full" />
        </div>
      </div>
      <div class="bg-white rounded-xl shadow-sm border border-slate-200 overflow-hidden">
        <DataTable :columns="systemColumns" :data="filteredSystemData" :loading="isLoading" />
      </div>
    </div>

    <!-- Tab 2: Manual / History -->
    <div v-if="activeTab === 'manual'" class="space-y-4 animate-in fade-in duration-300">
      <div class="bg-white p-4 rounded-xl shadow-sm border border-slate-200 flex flex-col md:flex-row justify-between gap-4">
        <div class="relative w-full md:w-96">
          <Search class="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-slate-400" />
          <Input v-model="searchManualQuery" placeholder="Cari nama peringatan manual..." class="pl-9 w-full" />
        </div>
      </div>
      <div class="bg-white rounded-xl shadow-sm border border-slate-200 overflow-hidden">
        <DataTable :columns="manualColumns" :data="filteredManualData" :loading="isLoadingManual" />
      </div>
    </div>

    <!-- Action Modal -->
    <Dialog :open="isModalOpen" @update:open="isModalOpen = $event">
      <DialogContent class="sm:max-w-[500px]">
        <DialogHeader>
          <DialogTitle class="flex items-center gap-2">
            <AlertTriangle class="w-5 h-5 text-red-500" />
            <span v-if="modalMode === 'system'">Tindak Lanjut HR</span>
            <span v-else-if="modalMode === 'direct_action'">Tindak Lanjut {{ actionForm.type === 'cs_followup' ? 'CS' : 'SPV' }}</span>
            <span v-else-if="modalMode === 'manual_add'">Tambah Peringatan Manual</span>
            <span v-else>Edit Peringatan</span>
          </DialogTitle>
          <DialogDescription>
            <span v-if="modalMode === 'system'">Berikan teguran atau Surat Peringatan (SP 1) kepada <strong>{{ selectedEmployee?.name }}</strong>.</span>
            <span v-else-if="modalMode === 'direct_action'">Berikan catatan/teguran dari {{ actionForm.type === 'cs_followup' ? 'CS' : 'SPV' }} kepada <strong>{{ selectedEmployee?.name }}</strong>.</span>
            <span v-else-if="modalMode === 'manual_add'">Berikan teguran/SP secara manual kepada karyawan pilihan Anda.</span>
            <span v-else>Ubah data peringatan yang telah dibuat.</span>
          </DialogDescription>
        </DialogHeader>

        <div class="space-y-4 py-4">
          <!-- Issue Info (only for system) -->
          <div v-if="modalMode === 'system' || modalMode === 'direct_action'" class="space-y-2">
            <label class="text-sm font-medium">Masalah Terdeteksi</label>
            <div class="p-3 bg-red-50 text-red-700 text-sm rounded-lg border border-red-100">
              {{ selectedEmployee?.issue }}
            </div>
          </div>

          <!-- Employee Selector (only for manual add) -->
          <div v-if="modalMode === 'manual_add'" class="space-y-2">
            <label class="text-sm font-medium">Pilih Karyawan</label>
            <div class="relative">
              <Search class="absolute left-3 top-3 h-4 w-4 text-slate-400" />
              <Input v-model="employeeSearch" placeholder="Cari nama atau ID karyawan..." class="pl-9 w-full mb-2" />
            </div>
            <div class="max-h-48 overflow-y-auto border border-slate-200 rounded-md bg-white">
              <div v-if="filteredAllEmployees.length === 0" class="p-3 text-sm text-slate-500 text-center">
                Karyawan tidak ditemukan
              </div>
              <div v-for="emp in filteredAllEmployees" :key="emp.employee_id_number" 
                   @click="actionForm.employee_id = emp.employee_id_number"
                   :class="['px-3 py-2 text-sm cursor-pointer hover:bg-slate-100 transition-colors', actionForm.employee_id === emp.employee_id_number ? 'bg-indigo-50 text-indigo-700 font-medium' : 'text-slate-700']">
                {{ emp.first_name }} {{ emp.last_name }} ({{ emp.employee_id_number }}) - {{ emp.job_position?.title }}
              </div>
            </div>
          </div>

          <!-- Employee Info (for manual edit) -->
          <div v-if="modalMode === 'manual_edit'" class="space-y-2">
            <label class="text-sm font-medium">Karyawan</label>
            <Input :value="selectedEmployee?.name + ' (' + selectedEmployee?.employee_id + ')'" disabled />
          </div>

          <!-- Action Type -->
          <div v-if="modalMode !== 'direct_action'" class="space-y-2">
            <label class="text-sm font-medium">Jenis Tindakan</label>
            <select v-model="actionForm.type" class="w-full flex h-10 items-center justify-between rounded-md border border-slate-200 bg-white px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-slate-950">
              <option value="teguran">Teguran Lisan / Tertulis</option>
              <option value="sp1">Surat Peringatan 1 (SP 1)</option>
              <option value="sp2">Surat Peringatan 2 (SP 2)</option>
              <option value="sp3">Surat Peringatan 3 (SP 3)</option>
            </select>
          </div>

          <!-- Notes -->
          <div class="space-y-2">
            <label class="text-sm font-medium">Catatan / Pesan Teguran</label>
            <Textarea
              v-model="actionForm.notes"
              placeholder="Masukkan pesan teguran atau alasan..."
              rows="3"
            />
          </div>
        </div>

        <DialogFooter>
          <Button variant="outline" @click="closeActionModal" :disabled="isSubmitting">Batal</Button>
          <Button @click="submitAction" :disabled="isSubmitting" class="bg-red-600 hover:bg-red-700 text-white">
            <span v-if="isSubmitting">Memproses...</span>
            <span v-else>Simpan Tindakan</span>
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  </div>
</template>
