<script setup lang="ts">
import { ref, h, onMounted, watch } from 'vue'
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
import { useAuthStore } from '@/stores/auth'

const authStore = useAuthStore()
const isModalOpen = ref(false)

const globalAccessEligibleRoles = [
  'superadmin',
  'hr_admin',
  'manager',
  'auditor',
  'customer_support',
  'admin_operasional',
  'admin_pabrik',
  'supervisor_gudang',
  'cashier_pusat',
  'admin_designer',
  'validator_desain',
  'validator_toko',
  'admin_purchasing',
]

const accessRoles = ref<any[]>([])

const newUser = ref({
  id: '',
  email: '',
  password: '',
  role: 'employee',
  access_role_id: '',
  isActive: true,
  hasGlobalAccess: false,
  canAbsenAnywhere: false,
})

watch(
  () => newUser.value.role,
  (newRole) => {
    if (newRole === 'admin_designer' || newRole === 'validator_desain' || newRole === 'validator_toko' || newRole === 'admin_purchasing') {
      newUser.value.hasGlobalAccess = true
    }
  },
)

const isEditMode = ref(false)
const isSubmitting = ref(false)
const users = ref<any[]>([])
const isLoading = ref(true)

const fetchAccessRoles = async () => {
  try {
    const res = await apiClient.get('/access-roles')
    accessRoles.value = res.data?.data || []
  } catch (err) {
    console.error('Failed to fetch access roles:', err)
  }
}

const openAddModal = () => {
  isEditMode.value = false
  newUser.value = {
    id: '',
    email: '',
    password: '',
    role: 'employee',
    access_role_id: '',
    isActive: true,
    hasGlobalAccess: false,
    canAbsenAnywhere: false,
  }
  isModalOpen.value = true
}

const openEditModal = (user: any) => {
  isEditMode.value = true
  const initialGlobal =
    user.has_global_access !== undefined
      ? user.has_global_access
      : (user.role === 'admin_designer' || user.role === 'validator_desain' || user.role === 'validator_toko' || user.role === 'admin_purchasing')

  newUser.value = {
    id: user.id,
    email: user.email || '',
    password: '', // Blank initially for update
    role: user.role || 'employee',
    access_role_id: user.access_role_id || user.access_role?.id || '',
    isActive: user.is_active,
    hasGlobalAccess: initialGlobal,
    canAbsenAnywhere: user.can_absen_anywhere || false,
  }
  
  // Fetch permissions
  apiClient.get(`/users/${user.id}/permissions`).then(res => {
    const perms = res.data?.data?.permissions || res.data?.permissions || []
    if (perms.includes('view_all_branches')) {
      newUser.value.hasGlobalAccess = true
    }
  }).catch(e => console.error(e))

  isModalOpen.value = true
}

const closeAddModal = () => {
  isModalOpen.value = false
}

const fetchUsers = async () => {
  isLoading.value = true
  try {
    const response = await apiClient.get('/users')
    // Handle both {data: [...]} and {data: {data: [...]}} structures
    const rawData = response.data.data
    users.value = Array.isArray(rawData) ? rawData : rawData?.data || []
  } catch (error: any) {
    console.error('Failed to fetch users:', error)
    toast.error(
      'Gagal mengambil data pengguna: ' + (error.response?.data?.message || error.message),
    )
  } finally {
    isLoading.value = false
  }
}

const saveUser = async () => {
  isSubmitting.value = true
  try {
    const roleIdPayload = newUser.value.access_role_id && newUser.value.access_role_id !== 'none'
      ? newUser.value.access_role_id
      : null

    const cleanEmail = newUser.value.email.trim().toLowerCase()
    const cleanPassword = newUser.value.password ? newUser.value.password.trim() : ''

    if (isEditMode.value) {
      // Don't validate password on edit unless typed
      await apiClient.put(`/users/${newUser.value.id}`, {
        email: cleanEmail,
        password: cleanPassword,
        role: newUser.value.role,
        access_role_id: roleIdPayload,
        isActive: newUser.value.isActive,
        canAbsenAnywhere: newUser.value.canAbsenAnywhere,
      })
      
      // Update permissions
      await apiClient.put(`/users/${newUser.value.id}/permissions`, {
        permissions: newUser.value.hasGlobalAccess ? ['view_all_branches'] : []
      })
      
      toast.success('Pengguna berhasil diperbarui!')
    } else {
      const res = await apiClient.post('/users', {
        ...newUser.value,
        email: cleanEmail,
        password: cleanPassword,
        access_role_id: roleIdPayload,
        canAbsenAnywhere: newUser.value.canAbsenAnywhere
      })
      const createdUserId = res.data?.data?.id || res.data?.id
      if (createdUserId) {
        await apiClient.put(`/users/${createdUserId}/permissions`, {
          permissions: newUser.value.hasGlobalAccess ? ['view_all_branches'] : []
        })
      }
      toast.success('Pengguna berhasil ditambahkan!')
    }
    closeAddModal()
    fetchUsers()
  } catch (error: any) {
    toast.error(
      `Gagal ${isEditMode.value ? 'memperbarui' : 'menambahkan'} pengguna: ` +
        (error.response?.data?.error || error.response?.data?.message || error.message),
    )
  } finally {
    isSubmitting.value = false
  }
}

const deleteUser = async (id: string) => {
  if (
    !confirm('Apakah Anda yakin ingin menghapus pengguna ini? Operasi ini tidak dapat dibatalkan.')
  )
    return

  try {
    await apiClient.delete(`/users/${id}`)
    toast.success('Pengguna berhasil dihapus!')
    fetchUsers()
  } catch (error: any) {
    toast.error(
      'Gagal menghapus pengguna: ' +
        (error.response?.data?.error || error.response?.data?.message || error.message),
    )
  }
}

onMounted(() => {
  fetchUsers()
  fetchAccessRoles()
})

const columns = [
  {
    accessorKey: 'email',
    header: 'EMAIL',
    cell: (info: any) => h('span', { class: 'font-bold text-gray-900' }, info.getValue() || '-'),
  },
  {
    accessorKey: 'role',
    header: 'HAK AKSES',
    cell: ({ getValue }: any) => {
      const val = getValue() as string
      let classes = 'border border-[#e5e7eb] text-gray-600'
      if (val === 'superadmin') classes = 'border border-[#fecaca] text-[#dc2626]'
      else if (val === 'hr_admin') classes = 'border border-[#fef08a] text-[#ca8a04]'
      else if (val === 'employee') classes = 'border border-[#bbf7d0] text-[#16a34a]'
      else if (val === 'admin_operasional') classes = 'border border-[#bfdbfe] text-[#2563eb]' // blue badge
      else if (val === 'supervisor_sales') classes = 'border border-[#fde047] text-[#854d0e]'
      else if (val === 'cashier') classes = 'border border-[#d8b4fe] text-[#7e22ce]'
      else if (val === 'cashier_pusat') classes = 'border border-[#f0abfc] text-[#a21caf]' // fuchsia/purple badge
      else if (val === 'admin_kasir') classes = 'border border-[#c4b5fd] text-[#6d28d9]'
      else if (val === 'admin_pabrik') classes = 'border border-[#fed7aa] text-[#c2410c]' // orange badge
      else if (val === 'supervisor_gudang') classes = 'border border-[#6ee7b7] text-[#047857]' // emerald badge
      else if (val === 'admin_designer') classes = 'border border-[#c7d2fe] text-[#4338ca]' // indigo badge
      else if (val === 'validator_desain') classes = 'border border-[#e9d5ff] text-[#7e22ce]' // purple badge
      else if (val === 'validator_toko') classes = 'border border-[#bae6fd] text-[#0284c7]' // sky/blue badge
      else if (val === 'admin_purchasing') classes = 'border border-[#fed7aa] text-[#ea580c]' // orange badge
      return h(
        'span',
        { class: `${classes} px-3 py-1 rounded-full text-[11px] font-bold inline-block uppercase` },
        val,
      )
    },
  },
  {
    accessorKey: 'access_role',
    header: 'ROLE DINAMIS / MENU',
    cell: ({ row }: any) => {
      const user = row.original
      if (user.access_role?.name) {
        return h(
          'span',
          {
            class:
              'inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-[11px] font-semibold bg-violet-50 text-violet-700 border border-violet-200',
          },
          user.access_role.name,
        )
      }
      return h('span', { class: 'text-gray-400 text-xs italic' }, 'Default Role')
    },
  },
  {
    accessorKey: 'has_global_access',
    header: 'AKSES CABANG',
    cell: ({ row }: any) => {
      const user = row.original
      const isGlobal =
        user.has_global_access ||
        user.role === 'superadmin' ||
        user.role === 'customer_support' ||
        user.role === 'auditor' ||
        user.role === 'admin_designer' ||
        user.role === 'validator_desain' ||
        user.role === 'validator_toko' ||
        user.role === 'admin_purchasing'
      if (isGlobal) {
        return h(
          'span',
          {
            class:
              'inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-[11px] font-bold bg-blue-50 text-blue-700 border border-blue-200',
          },
          [h('span', { class: 'w-1.5 h-1.5 rounded-full bg-blue-500' }), 'Semua Cabang'],
        )
      }
      return h(
        'span',
        {
          class:
            'inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-[11px] font-medium text-slate-500 bg-slate-50 border border-slate-200',
        },
        'Cabang Terbatas',
      )
    },
  },
  {
    accessorKey: 'is_active',
    header: 'STATUS',
    cell: (info: any) => {
      const active = info.getValue()
      return h(
        'span',
        {
          class: active ? 'text-black font-bold' : 'text-red-500 font-bold',
        },
        active ? 'Aktif' : 'Non-Aktif',
      )
    },
  },
  {
    accessorFn: (row: any) =>
      row.created_at ? new Date(row.created_at).toLocaleDateString() : '-',
    id: 'createdAt',
    header: 'DIBUAT PADA',
    cell: (info: any) => h('span', { class: 'text-gray-500 text-[13px]' }, info.getValue()),
  },
  {
    accessorFn: (row: any) =>
      row.updated_at ? new Date(row.updated_at).toLocaleDateString() : '-',
    id: 'updatedAt',
    header: 'DIPERBARUI PADA',
    cell: (info: any) => h('span', { class: 'text-gray-500 text-[13px]' }, info.getValue()),
  },
  {
    id: 'actions',
    header: 'AKSI',
    cell: ({ row }: any) => {
      const user = row.original
      // Disable action buttons for auditor
      if (authStore.user?.role === 'auditor') return h('span', { class: 'text-gray-400 text-xs italic' }, 'Read Only')
      
      return h('div', { class: 'flex items-center gap-2' }, [
        h(
          Button,
          {
            variant: 'ghost',
            size: 'sm',
            class: 'h-8 px-2 text-primary hover:text-primary hover:bg-primary/5',
            onClick: () => openEditModal(user),
          },
          () => h(Pencil, { class: 'w-4 h-4' }),
        ),
        h(
          Button,
          {
            variant: 'ghost',
            size: 'sm',
            class: 'h-8 px-2 text-red-600 hover:text-red-700 hover:bg-red-50',
            onClick: () => deleteUser(user.id),
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
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
      <div>
        <h1 class="text-[22px] font-bold text-gray-900 leading-tight">Manajemen Akun Pengguna</h1>
        <p class="text-[14px] text-gray-500 mt-1">
          Atur kredensial dan hak akses aplikasi (Admin, Karyawan).
        </p>
      </div>

      <div class="flex items-center gap-3">
        <Button v-if="authStore.user?.role !== 'auditor'" @click="openAddModal" class="gap-2 shadow-sm rounded-lg" size="sm">
          <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2"
              d="M12 4v16m8-8H4"
            />
          </svg>
          Tambah Pengguna
        </Button>
      </div>
    </div>

    <DataTable :data="users" :columns="columns" :isLoading="isLoading" />

    <Dialog v-model:open="isModalOpen">
      <DialogContent
        class="sm:max-w-xl rounded-[2rem] md:rounded-[3rem] p-0 overflow-hidden border-none shadow-2xl max-h-[95vh] flex flex-col"
      >
        <div class="bg-slate-900 p-8 md:p-10 text-white relative shrink-0">
          <DialogHeader>
            <DialogTitle class="text-xl md:text-2xl font-black">{{
              isEditMode ? 'Edit Pengguna' : 'Tambah Pengguna'
            }}</DialogTitle>
            <DialogDescription
              class="text-slate-400 font-bold mt-2 uppercase text-[10px] md:text-[11px] tracking-widest border-l-4 border-primary pl-4"
            >
              Kelola hak akses dan kredensial akun sistem.
            </DialogDescription>
          </DialogHeader>
        </div>

        <div class="p-6 md:p-10 bg-white space-y-6 overflow-y-auto custom-scrollbar flex-1">
          <div class="grid gap-2">
            <label class="text-[13px] font-medium text-gray-700">Email Login</label>
            <Input
              v-model="newUser.email"
              type="email"
              autocomplete="off"
              autocapitalize="none"
              autocorrect="off"
              spellcheck="false"
              placeholder="e.g. jdoe@company.com"
            />
          </div>

          <div class="grid gap-2">
            <label class="text-[13px] font-medium text-gray-700">Password</label>
            <Input
              v-model="newUser.password"
              type="password"
              autocomplete="new-password"
              autocapitalize="none"
              autocorrect="off"
              spellcheck="false"
              :placeholder="isEditMode ? 'Kosongkan jika tidak ingin diubah' : 'Minimal 6 karakter'"
            />
          </div>

          <div class="grid gap-2">
            <label class="text-[13px] font-medium text-gray-700">Hak Akses Utama (Base Role)</label>
            <Select v-model="newUser.role">
              <SelectTrigger>
                <SelectValue placeholder="Pilih Role" />
              </SelectTrigger>
              <SelectContent>
                <SelectGroup>
                  <SelectItem value="superadmin">Super Admin</SelectItem>
                  <SelectItem value="hr_admin">HR Admin</SelectItem>
                  <SelectItem value="manager">Manager</SelectItem>
                  <SelectItem value="supervisor_sales">Supervisor Sales</SelectItem>
                  <SelectItem value="cashier">Cashier Cabang</SelectItem>
                  <SelectItem value="cashier_pusat">Cashier Pusat</SelectItem>
                  <SelectItem value="admin_kasir">Admin Kasir</SelectItem>
                  <SelectItem value="admin_nota">Admin Nota</SelectItem>
                  <SelectItem value="admin_gudang">Admin Gudang</SelectItem>
                  <SelectItem value="admin_marketplace">Admin Marketplace</SelectItem>
                  <SelectItem value="admin_designer">Admin Designer</SelectItem>
                  <SelectItem value="validator_desain">Validator Desain</SelectItem>
                  <SelectItem value="validator_toko">Validator Toko</SelectItem>
                  <SelectItem value="admin_purchasing">Admin Purchasing</SelectItem>
                  <SelectItem value="admin_operasional">Admin Operasional</SelectItem>
                  <SelectItem value="admin_pabrik">Admin Pabrik</SelectItem>
                  <SelectItem value="supervisor_gudang">Supervisor Gudang</SelectItem>
                  <SelectItem value="auditor">Auditor</SelectItem>
                  <SelectItem value="customer_support">Customer Support</SelectItem>
                  <SelectItem value="employee">Employee (Karyawan)</SelectItem>
                </SelectGroup>
              </SelectContent>
            </Select>
            <p v-if="newUser.role === 'auditor'" class="text-[11px] text-amber-700 bg-amber-50 p-2.5 rounded-lg border border-amber-200 mt-1">
              💡 <b>Auditor Cabang Booster:</b> Jika akun ini terikat pada cabang dengan nama mengandung <b>"Booster"</b>, sistem otomatis membuka akses operasional penuh (Admin Gudang, Admin Nota, dan Kasir). Pada cabang non-booster, Auditor tetap Read-Only.
            </p>
          </div>

          <div class="grid gap-2">
            <div class="flex items-center justify-between">
              <label class="text-[13px] font-medium text-gray-700">Role Dinamis (Menu Khusus)</label>
              <span class="text-[11px] text-slate-400">Opsional untuk filter menu spesifik</span>
            </div>
            <Select v-model="newUser.access_role_id">
              <SelectTrigger>
                <SelectValue placeholder="Pilih Role Dinamis (Opsional)" />
              </SelectTrigger>
              <SelectContent>
                <SelectGroup>
                  <SelectItem value="none">-- Default (Ikuti Hak Akses Utama) --</SelectItem>
                  <SelectItem v-for="ar in accessRoles" :key="ar.id" :value="ar.id">
                    {{ ar.name }}
                  </SelectItem>
                </SelectGroup>
              </SelectContent>
            </Select>
            <p class="text-[11px] text-slate-500">
              Contoh: Pilih <b>Admin Gudang Bahan Jadi (Stok Only)</b> atau <b>Admin Gudang Bahan Baku & PO</b> untuk membatasi tampilan menu Pabrik sesuai jobdesk.
            </p>
          </div>

          <div class="flex items-center space-x-2 mt-4">
            <input
              type="checkbox"
              id="isActive"
              v-model="newUser.isActive"
              class="w-4 h-4 rounded border-gray-300 text-primary focus:ring-primary"
            />
            <label for="isActive" class="text-sm font-medium leading-none">
              Akun Aktif (Dapat Login)
            </label>
          </div>

          <div
            class="space-y-1.5 mt-2"
            v-if="globalAccessEligibleRoles.includes(newUser.role)"
          >
            <div class="flex items-center space-x-2">
              <input
                type="checkbox"
                id="hasGlobalAccess"
                v-model="newUser.hasGlobalAccess"
                class="w-4 h-4 rounded border-gray-300 text-primary focus:ring-primary"
              />
              <label for="hasGlobalAccess" class="text-sm font-medium leading-none text-blue-600">
                Izinkan Lihat Data Semua Cabang (Hak Akses Global)
              </label>
            </div>
            <p
              v-if="newUser.role === 'admin_designer' || newUser.role === 'validator_desain' || newUser.role === 'validator_toko' || newUser.role === 'admin_purchasing'"
              class="text-[11px] text-slate-500 pl-6"
            >
              Memberikan akses penuh untuk melihat dan memproses data seluruh cabang tanpa batasan lokasi.
            </p>
          </div>

          <div class="flex items-center space-x-2 mt-2">
            <input
              type="checkbox"
              id="canAbsenAnywhere"
              v-model="newUser.canAbsenAnywhere"
              class="w-4 h-4 rounded border-gray-300 text-primary focus:ring-primary"
            />
            <label for="canAbsenAnywhere" class="text-sm font-medium leading-none text-emerald-600">
              Bisa Absen Dari Mana Saja (Bypass Radius Lokasi)
            </label>
          </div>
        </div>

        <DialogFooter class="p-6 md:p-10 pt-0 bg-white grid grid-cols-2 gap-4 shrink-0">
          <Button
            variant="ghost"
            @click="isModalOpen = false"
            :disabled="isSubmitting"
            class="rounded-2xl h-12 md:h-14 font-black text-slate-400 border border-slate-100 hover:bg-slate-50 transition-all"
            >BATAL</Button
          >
          <Button
            @click="saveUser"
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
