<script setup lang="ts">
import { ref, h, onMounted } from 'vue'
import { Pencil, Trash2, Shield, Plus } from 'lucide-vue-next'
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

const roles = ref<any[]>([])
const menus = ref<any[]>([])
const modules = ref<any[]>([])
const isLoading = ref(true)

const isModalOpen = ref(false)
const isEditMode = ref(false)
const isSubmitting = ref(false)

const form = ref({
  id: '',
  name: '',
  description: '',
  menu_ids: [] as string[],
})

const fetchMenus = async () => {
  try {
    const res = await apiClient.get('/app-menus')
    menus.value = res.data.data || []
  } catch (error) {
    console.error('Failed to fetch menus', error)
  }
}

const fetchModules = async () => {
  try {
    const res = await apiClient.get('/app-modules')
    modules.value = res.data.data || []
  } catch (error) {
    console.error('Failed to fetch modules', error)
  }
}

const fetchRoles = async () => {
  isLoading.value = true
  try {
    const res = await apiClient.get('/access-roles')
    roles.value = res.data.data || []
  } catch (error: any) {
    toast.error(error.response?.data?.message || 'Gagal memuat data hak akses')
  } finally {
    isLoading.value = false
  }
}

onMounted(async () => {
  await fetchModules()
  await fetchMenus()
  await fetchRoles()
})

const openAddModal = () => {
  isEditMode.value = false
  form.value = { id: '', name: '', description: '', menu_ids: [] }
  isModalOpen.value = true
}

const openEditModal = (role: any) => {
  isEditMode.value = true
  form.value = {
    id: role.id,
    name: role.name,
    description: role.description,
    menu_ids: role.menus?.map((m: any) => m.id) || [],
  }
  isModalOpen.value = true
}

const closeDialog = () => {
  isModalOpen.value = false
}

const saveRole = async () => {
  if (!form.value.name) {
    toast.error('Nama hak akses wajib diisi')
    return
  }

  isSubmitting.value = true
  try {
    const payload = {
      name: form.value.name,
      description: form.value.description,
      menu_ids: form.value.menu_ids,
    }

    if (isEditMode.value) {
      await apiClient.put(`/access-roles/${form.value.id}`, payload)
      toast.success('Hak akses berhasil diperbarui')
    } else {
      await apiClient.post('/access-roles', payload)
      toast.success('Hak akses berhasil ditambahkan')
    }
    closeDialog()
    fetchRoles()
  } catch (error: any) {
    toast.error(error.response?.data?.message || 'Gagal menyimpan hak akses')
  } finally {
    isSubmitting.value = false
  }
}

const deleteRole = async (id: string) => {
  if (!confirm('Apakah Anda yakin ingin menghapus hak akses ini?')) return
  try {
    await apiClient.delete(`/access-roles/${id}`)
    toast.success('Hak akses berhasil dihapus')
    fetchRoles()
  } catch (error: any) {
    toast.error(error.response?.data?.message || 'Gagal menghapus hak akses')
  }
}

const toggleMenu = (menuId: string) => {
  const index = form.value.menu_ids.indexOf(menuId)
  if (index === -1) {
    form.value.menu_ids.push(menuId)
  } else {
    form.value.menu_ids.splice(index, 1)
  }
}

const columns = [
  {
    accessorKey: 'name',
    header: 'Nama Hak Akses',
    cell: ({ row }: any) => h('div', { class: 'font-medium' }, row.original.name),
  },
  {
    accessorKey: 'description',
    header: 'Deskripsi',
  },
  {
    accessorKey: 'menus',
    header: 'Jumlah Akses Menu',
    cell: ({ row }: any) => h('div', {}, `${row.original.menus?.length || 0} Menu`),
  },
  {
    id: 'actions',
    header: 'Aksi',
    cell: ({ row }: any) => {
      return h('div', { class: 'flex gap-2' }, [
        h(
          Button,
          {
            variant: 'outline',
            size: 'sm',
            onClick: () => openEditModal(row.original),
          },
          () => h(Pencil, { class: 'h-4 w-4 text-blue-600' }),
        ),
        h(
          Button,
          {
            variant: 'outline',
            size: 'sm',
            onClick: () => deleteRole(row.original.id),
          },
          () => h(Trash2, { class: 'h-4 w-4 text-red-600' }),
        ),
      ])
    },
  },
]
</script>

<template>
  <div class="space-y-6">
    <div class="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4">
      <div>
        <h1 class="text-2xl font-bold tracking-tight flex items-center gap-2">
          <Shield class="h-6 w-6 text-primary" />
          Manajemen Hak Akses
        </h1>
        <p class="text-muted-foreground mt-1">
          Kelola role pengguna, modul, dan level menu yang dapat diakses.
        </p>
      </div>
      <Button @click="openAddModal" class="gap-2">
        <Plus class="h-4 w-4" />
        Tambah Role
      </Button>
    </div>

    <!-- Data Table -->
    <div class="bg-white rounded-xl shadow-sm border border-slate-200 overflow-hidden">
      <DataTable :columns="columns" :data="roles" :loading="isLoading" searchKey="name" />
    </div>

    <!-- Modal Form -->
    <Dialog :open="isModalOpen" @update:open="closeDialog">
      <DialogContent class="sm:max-w-[700px] max-h-[90vh] flex flex-col">
        <DialogHeader>
          <DialogTitle>{{ isEditMode ? 'Edit Hak Akses' : 'Tambah Hak Akses' }}</DialogTitle>
          <DialogDescription>
            Atur nama role dan pilih modul serta menu yang diizinkan untuk role ini.
          </DialogDescription>
        </DialogHeader>

        <div class="space-y-4 py-4 overflow-y-auto flex-1 pr-2">
          <div class="grid gap-2">
            <label class="text-sm font-medium">Nama Hak Akses</label>
            <Input v-model="form.name" placeholder="Misal: HR Admin, Sales Manager" />
          </div>

          <div class="grid gap-2">
            <label class="text-sm font-medium">Deskripsi</label>
            <Input v-model="form.description" placeholder="Deskripsi hak akses..." />
          </div>

          <div class="mt-4">
            <h3 class="text-sm font-semibold mb-2">Konfigurasi Akses Menu & Modul</h3>
            <div class="space-y-4">
              <div v-if="modules.length === 0" class="text-sm text-gray-500 italic">
                Belum ada modul yang terdaftar. Harap tambahkan modul terlebih dahulu melalui seeder atau API.
              </div>
              <div
                v-for="module in modules"
                :key="module.id"
                class="border rounded-md p-4 bg-slate-50"
              >
                <h4 class="font-medium text-slate-800 mb-2">{{ module.name }}</h4>
                <div class="grid grid-cols-1 sm:grid-cols-2 gap-2">
                  <label
                    v-for="menu in menus.filter((m) => m.module_id === module.id)"
                    :key="menu.id"
                    class="flex items-center space-x-2 bg-white p-2 rounded border cursor-pointer hover:bg-slate-50 transition-colors"
                  >
                    <input
                      type="checkbox"
                      :checked="form.menu_ids.includes(menu.id)"
                      @change="toggleMenu(menu.id)"
                      class="rounded border-gray-300 text-primary focus:ring-primary"
                    />
                    <span class="text-sm">
                      {{ menu.name }}
                      <span class="text-xs text-gray-400 block">{{ menu.path }}</span>
                    </span>
                  </label>
                  <div v-if="menus.filter((m) => m.module_id === module.id).length === 0" class="text-xs text-gray-400 col-span-2">
                    Tidak ada menu di modul ini
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>

        <DialogFooter class="pt-4 border-t mt-auto">
          <Button variant="outline" @click="closeDialog" :disabled="isSubmitting">Batal</Button>
          <Button @click="saveRole" :disabled="isSubmitting">
            {{ isSubmitting ? 'Menyimpan...' : 'Simpan' }}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  </div>
</template>
