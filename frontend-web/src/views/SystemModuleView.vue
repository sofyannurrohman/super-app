<script setup lang="ts">
import { ref, onMounted, h } from 'vue'
import { Plus, Settings, Trash2 } from 'lucide-vue-next'
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

const modules = ref<any[]>([])
const menus = ref<any[]>([])
const isLoading = ref(true)

const isModuleModalOpen = ref(false)
const isMenuModalOpen = ref(false)
const isSubmitting = ref(false)

const moduleForm = ref({ name: '', description: '' })
const menuForm = ref({ name: '', path: '', module_id: '' })

const fetchModules = async () => {
  try {
    const res = await apiClient.get('/app-modules')
    modules.value = res.data.data || []
  } catch (error) {
    console.error('Failed to fetch modules', error)
  }
}

const fetchMenus = async () => {
  try {
    const res = await apiClient.get('/app-menus')
    menus.value = res.data.data || []
  } catch (error) {
    console.error('Failed to fetch menus', error)
  }
}

const loadData = async () => {
  isLoading.value = true
  await Promise.all([fetchModules(), fetchMenus()])
  isLoading.value = false
}

onMounted(() => {
  loadData()
})

const saveModule = async () => {
  if (!moduleForm.value.name) {
    toast.error('Nama modul wajib diisi')
    return
  }
  isSubmitting.value = true
  try {
    await apiClient.post('/app-modules', moduleForm.value)
    toast.success('Modul berhasil ditambahkan')
    isModuleModalOpen.value = false
    moduleForm.value = { name: '', description: '' }
    loadData()
  } catch (error: any) {
    toast.error(error.response?.data?.message || 'Gagal menyimpan modul')
  } finally {
    isSubmitting.value = false
  }
}

const saveMenu = async () => {
  if (!menuForm.value.name || !menuForm.value.module_id) {
    toast.error('Nama menu dan modul tujuan wajib diisi')
    return
  }
  isSubmitting.value = true
  try {
    await apiClient.post('/app-menus', {
      name: menuForm.value.name,
      path: menuForm.value.path,
      module_id: menuForm.value.module_id,
      level: 1,
    })
    toast.success('Menu berhasil ditambahkan')
    isMenuModalOpen.value = false
    menuForm.value = { name: '', path: '', module_id: '' }
    loadData()
  } catch (error: any) {
    toast.error(error.response?.data?.message || 'Gagal menyimpan menu')
  } finally {
    isSubmitting.value = false
  }
}

const moduleColumns = [
  { accessorKey: 'name', header: 'Nama Modul' },
  { accessorKey: 'description', header: 'Deskripsi' },
  {
    id: 'menus_count',
    header: 'Jumlah Menu',
    cell: ({ row }: any) => {
      const count = menus.value.filter(m => m.module_id === row.original.id).length
      return h('span', {}, count.toString())
    }
  }
]

const menuColumns = [
  { accessorKey: 'name', header: 'Nama Menu' },
  { accessorKey: 'path', header: 'Path URL' },
  {
    id: 'module_name',
    header: 'Modul Induk',
    cell: ({ row }: any) => {
      const mod = modules.value.find(m => m.id === row.original.module_id)
      return h('span', { class: 'text-slate-600 font-medium' }, mod ? mod.name : 'Unknown')
    }
  }
]
</script>

<template>
  <div class="space-y-8">
    <div class="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4">
      <div>
        <h1 class="text-2xl font-bold tracking-tight flex items-center gap-2">
          <Settings class="h-6 w-6 text-primary" />
          Access Control (Modul & Menu)
        </h1>
        <p class="text-muted-foreground mt-1">
          Kelola master data modul aplikasi dan daftar menu untuk dikonfigurasi ke dalam Hak Akses (Role).
        </p>
      </div>
    </div>

    <!-- Modules Section -->
    <div class="space-y-4">
      <div class="flex justify-between items-center">
        <h2 class="text-xl font-bold text-slate-800">Daftar Modul Aplikasi</h2>
        <Button @click="isModuleModalOpen = true" size="sm" class="gap-2">
          <Plus class="h-4 w-4" /> Tambah Modul
        </Button>
      </div>
      <div class="bg-white rounded-xl shadow-sm border border-slate-200 overflow-hidden">
        <DataTable :columns="moduleColumns" :data="modules" :loading="isLoading" searchKey="name" />
      </div>
    </div>

    <hr class="border-slate-200" />

    <!-- Menus Section -->
    <div class="space-y-4">
      <div class="flex justify-between items-center">
        <h2 class="text-xl font-bold text-slate-800">Daftar Menu Tersedia</h2>
        <Button @click="isMenuModalOpen = true" size="sm" variant="secondary" class="gap-2">
          <Plus class="h-4 w-4" /> Tambah Menu
        </Button>
      </div>
      <div class="bg-white rounded-xl shadow-sm border border-slate-200 overflow-hidden">
        <DataTable :columns="menuColumns" :data="menus" :loading="isLoading" searchKey="name" />
      </div>
    </div>

    <!-- Add Module Modal -->
    <Dialog :open="isModuleModalOpen" @update:open="isModuleModalOpen = false">
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Tambah Modul Aplikasi</DialogTitle>
          <DialogDescription>Masukkan nama modul (misal: HRIS, Sales, Finance).</DialogDescription>
        </DialogHeader>
        <div class="grid gap-4 py-4">
          <div class="grid gap-2">
            <label class="text-sm font-medium">Nama Modul</label>
            <Input v-model="moduleForm.name" placeholder="Nama modul..." />
          </div>
          <div class="grid gap-2">
            <label class="text-sm font-medium">Deskripsi</label>
            <Input v-model="moduleForm.description" placeholder="Deskripsi opsional..." />
          </div>
        </div>
        <DialogFooter>
          <Button variant="outline" @click="isModuleModalOpen = false" :disabled="isSubmitting">Batal</Button>
          <Button @click="saveModule" :disabled="isSubmitting">Simpan</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>

    <!-- Add Menu Modal -->
    <Dialog :open="isMenuModalOpen" @update:open="isMenuModalOpen = false">
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Tambah Menu Aplikasi</DialogTitle>
          <DialogDescription>Daftarkan rute menu ke dalam modul tertentu.</DialogDescription>
        </DialogHeader>
        <div class="grid gap-4 py-4">
          <div class="grid gap-2">
            <label class="text-sm font-medium">Modul Induk</label>
            <select v-model="menuForm.module_id" class="flex h-10 w-full items-center justify-between rounded-md border border-slate-200 bg-white px-3 py-2 text-sm ring-offset-white focus:outline-none focus:ring-2 focus:ring-slate-950 focus:ring-offset-2">
              <option value="" disabled>Pilih Modul...</option>
              <option v-for="mod in modules" :key="mod.id" :value="mod.id">{{ mod.name }}</option>
            </select>
          </div>
          <div class="grid gap-2">
            <label class="text-sm font-medium">Nama Menu</label>
            <Input v-model="menuForm.name" placeholder="Misal: Dashboard, Karyawan" />
          </div>
          <div class="grid gap-2">
            <label class="text-sm font-medium">Path / Route URL</label>
            <Input v-model="menuForm.path" placeholder="Misal: /employees" />
          </div>
        </div>
        <DialogFooter>
          <Button variant="outline" @click="isMenuModalOpen = false" :disabled="isSubmitting">Batal</Button>
          <Button @click="saveMenu" :disabled="isSubmitting">Simpan</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  </div>
</template>
