<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <div>
        <h2 class="text-2xl font-bold tracking-tight text-gray-900">Karyawan Harian</h2>
        <p class="text-sm text-gray-500">Kelola data karyawan harian dan PIN absen Kiosk.</p>
      </div>
      <div class="flex gap-2">
        <Button variant="outline" @click="showImportModal = true">
          <Upload class="w-4 h-4 mr-2" />
          Import Excel
        </Button>
        <Button @click="showAddModal = true">
          <Plus class="w-4 h-4 mr-2" />
          Tambah Harian
        </Button>
      </div>
    </div>

    <!-- Table Section -->
    <div class="bg-white rounded-lg border shadow-sm">
      <div class="p-4 border-b flex justify-between items-center">
        <div class="relative w-72">
          <Search class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" />
          <Input v-model="searchQuery" placeholder="Cari nama atau PIN..." class="pl-9" />
        </div>
        <Button variant="outline" @click="fetchEmployees">
          <RefreshCcw class="w-4 h-4 mr-2" />
          Refresh
        </Button>
      </div>

      <div class="overflow-x-auto">
        <table class="w-full text-sm text-left">
          <thead class="text-xs text-gray-700 uppercase bg-gray-50 border-b">
            <tr>
              <th class="px-6 py-3 font-semibold">Nama Lengkap</th>
              <th class="px-6 py-3 font-semibold">Cabang</th>
              <th class="px-6 py-3 font-semibold text-center">PIN Kiosk</th>
              <th class="px-6 py-3 font-semibold text-center">Status</th>
              <th class="px-6 py-3 font-semibold text-right">Aksi</th>
            </tr>
          </thead>
          <tbody>
            <tr v-if="loading" class="border-b">
              <td colspan="5" class="px-6 py-8 text-center text-gray-500">
                Memuat data...
              </td>
            </tr>
            <tr v-else-if="filteredEmployees.length === 0" class="border-b">
              <td colspan="5" class="px-6 py-8 text-center text-gray-500">
                Tidak ada data karyawan harian.
              </td>
            </tr>
            <tr v-else v-for="emp in filteredEmployees" :key="emp.id" class="border-b hover:bg-gray-50">
              <td class="px-6 py-4 font-medium">{{ emp.first_name }} {{ emp.last_name }}</td>
              <td class="px-6 py-4 text-gray-600">{{ getBranchName(emp.branch_id) }}</td>
              <td class="px-6 py-4 text-center">
                <span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-blue-100 text-blue-800">
                  {{ emp.pin || '-' }}
                </span>
              </td>
              <td class="px-6 py-4 text-center">
                <span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium"
                      :class="emp.status === 'Aktif' ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800'">
                  {{ emp.status || 'Aktif' }}
                </span>
              </td>
              <td class="px-6 py-4 text-right">
                <Button variant="ghost" size="sm" class="text-gray-500 hover:text-red-600" @click="confirmDelete(emp.id)">
                  <Trash2 class="w-4 h-4" />
                </Button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Add Modal -->
    <Dialog :open="showAddModal" @update:open="showAddModal = $event">
      <DialogContent class="sm:max-w-[425px]">
        <DialogHeader>
          <DialogTitle>Tambah Karyawan Harian</DialogTitle>
          <DialogDescription>
            Sistem akan otomatis men-generate PIN 3 digit setelah disimpan.
          </DialogDescription>
        </DialogHeader>
        <div class="grid gap-4 py-4">
          <div class="space-y-2">
            <Label for="name">Nama Lengkap</Label>
            <Input id="name" v-model="newEmployee.first_name" placeholder="Contoh: Budi Santoso" />
          </div>
          <div class="space-y-2">
            <Label for="branch">Cabang</Label>
            <Select v-model="newEmployee.branch_id">
              <SelectTrigger>
                <SelectValue placeholder="Pilih Cabang" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem v-for="b in branches" :key="b.id" :value="b.id">
                  {{ b.name }}
                </SelectItem>
              </SelectContent>
            </Select>
          </div>
        </div>
        <DialogFooter>
          <Button variant="outline" @click="showAddModal = false">Batal</Button>
          <Button @click="saveEmployee" :disabled="saving">
            <Loader2 v-if="saving" class="w-4 h-4 mr-2 animate-spin" />
            Simpan & Generate PIN
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>

    <!-- Import Modal -->
    <Dialog :open="showImportModal" @update:open="showImportModal = $event">
      <DialogContent class="sm:max-w-[600px]">
        <DialogHeader>
          <DialogTitle>Import Data Karyawan Harian</DialogTitle>
          <DialogDescription>
            Unggah file Excel (.xlsx) untuk menambahkan banyak karyawan sekaligus.
          </DialogDescription>
        </DialogHeader>
        <div class="py-4 space-y-4">
          <div class="bg-blue-50 p-4 rounded-lg text-sm text-blue-800">
            <strong>Aturan Format Excel:</strong>
            <ul class="list-disc ml-5 mt-2 space-y-1">
              <li>Baris pertama (Header) akan diabaikan.</li>
              <li><strong>Kolom A:</strong> Nama Lengkap Karyawan</li>
              <li><strong>Kolom B:</strong> ID Cabang (Copy/Paste dari tabel di bawah)</li>
            </ul>
          </div>
          
          <div class="border rounded-md max-h-96 overflow-y-auto">
            <table class="w-full text-sm text-left">
              <thead class="text-xs text-gray-700 bg-gray-50 sticky top-0">
                <tr>
                  <th class="px-4 py-2 font-semibold">Nama Cabang</th>
                  <th class="px-4 py-2 font-semibold">ID Cabang (Copy ini)</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="b in branches" :key="b.id" class="border-b">
                  <td class="px-4 py-2 font-medium">{{ b.name }}</td>
                  <td class="px-4 py-2 text-xs font-mono bg-gray-50 select-all cursor-pointer" @click="copyId(b.id)">{{ b.id }}</td>
                </tr>
              </tbody>
            </table>
          </div>

          <div class="space-y-2 pt-2">
            <Label for="excel-file">Pilih File Excel (.xlsx)</Label>
            <Input id="excel-file" type="file" accept=".xlsx" @change="onFileChange" class="cursor-pointer" />
          </div>
        </div>
        <DialogFooter>
          <Button variant="outline" @click="showImportModal = false">Batal</Button>
          <Button @click="uploadExcel" :disabled="!selectedFile || uploading">
            <Loader2 v-if="uploading" class="w-4 h-4 mr-2 animate-spin" />
            Upload & Import
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { Plus, Search, RefreshCcw, Trash2, Loader2, Upload } from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { toast } from 'vue-sonner'
import api from '@/api/axios'

const employees = ref<any[]>([])
const branches = ref<any[]>([])
const loading = ref(true)
const saving = ref(false)
const showAddModal = ref(false)
const showImportModal = ref(false)
const searchQuery = ref('')
const selectedFile = ref<File | null>(null)
const uploading = ref(false)

const newEmployee = ref({
  first_name: '',
  branch_id: '',
  employment_type: 'Harian'
})

const filteredEmployees = computed(() => {
  if (!searchQuery.value) return employees.value
  const q = searchQuery.value.toLowerCase()
  return employees.value.filter(e => 
    (e.first_name?.toLowerCase().includes(q) || '') ||
    (e.last_name?.toLowerCase().includes(q) || '') ||
    (e.pin?.includes(q) || '')
  )
})

const fetchBranches = async () => {
  try {
    const res = await api.get('/branches')
    branches.value = res.data.data || []
  } catch (error) {
    console.error('Error fetching branches:', error)
  }
}

const getBranchName = (id: string) => {
  const branch = branches.value.find(b => b.id === id)
  return branch ? branch.name : '-'
}

const fetchEmployees = async () => {
  loading.value = true
  try {
    const res = await api.get('/employees?employment_type=harian&limit=1000')
    employees.value = res.data.data || []
  } catch (error) {
    toast.error('Gagal mengambil data karyawan harian')
  } finally {
    loading.value = false
  }
}

const saveEmployee = async () => {
  if (!newEmployee.value.first_name || !newEmployee.value.branch_id) {
    toast.error('Nama dan Cabang wajib diisi')
    return
  }

  saving.value = true
  try {
    const uniqueId = Math.random().toString(36).substring(2, 8)
    const payload = {
      name: newEmployee.value.first_name,
      email: `harian_${uniqueId}@demo.local`,
      branchId: newEmployee.value.branch_id,
      employmentStatus: 'Harian'
    }
    
    await api.post('/employees', payload)
    toast.success('Karyawan harian ditambahkan & PIN terbuat')
    showAddModal.value = false
    newEmployee.value.first_name = ''
    fetchEmployees()
  } catch (error) {
    toast.error('Gagal menyimpan karyawan')
  } finally {
    saving.value = false
  }
}

const confirmDelete = async (id: string) => {
  if (confirm('Yakin ingin menghapus karyawan harian ini?')) {
    try {
      await api.delete(`/employees/${id}`)
      toast.success('Data karyawan berhasil dihapus')
      fetchEmployees()
    } catch (error) {
      toast.error('Gagal menghapus data')
    }
  }
}

const onFileChange = (e: Event) => {
  const target = e.target as HTMLInputElement
  if (target.files && target.files.length > 0) {
    selectedFile.value = target.files[0] || null
  } else {
    selectedFile.value = null
  }
}

const copyId = (id: string) => {
  navigator.clipboard.writeText(id)
  toast.success('ID Tersalin!')
}

const uploadExcel = async () => {
  if (!selectedFile.value) return
  
  uploading.value = true
  try {
    const formData = new FormData()
    formData.append('file', selectedFile.value)
    
    await api.post('/employees/import-daily', formData, {
      headers: {
        'Content-Type': 'multipart/form-data'
      }
    })
    toast.success('Berhasil mengimpor data karyawan harian')
    showImportModal.value = false
    selectedFile.value = null
    fetchEmployees()
  } catch (error: any) {
    const msg = error.response?.data?.message || 'Gagal mengimpor file Excel'
    toast.error(msg)
  } finally {
    uploading.value = false
  }
}


onMounted(() => {
  fetchBranches()
  fetchEmployees()
})
</script>
