<script setup lang="ts">
import { ref, h, onMounted, computed } from 'vue'
import {
  Pencil,
  Trash2,
  Building2,
  MapPin,
  Network,
  Plus,
  Globe,
  Building,
  ArrowRight,
  Settings2,
  CheckCircle2,
  Info,
} from 'lucide-vue-next'
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
import MapPicker from '@/components/MapPicker.vue'
import { toast } from 'vue-sonner'

// Main Tabs State
const activeTab = ref('companies') // 'companies', 'branches', 'departments'

// --- COMMON STATE ---
const isLoading = ref(true)
const isSubmitting = ref(false)

// --- COMPANY STATE ---
const companies = ref<any[]>([])
const isCompanyModalOpen = ref(false)
const companyForm = ref({
  id: '',
  name: '',
  taxNumber: '',
  address: '',
  logoUrl: '',
  branchIds: [] as string[],
})
const isCompanyEdit = ref(false)

// --- BRANCH STATE ---
const branches = ref<any[]>([])
const isBranchModalOpen = ref(false)
const branchForm = ref({
  id: '',
  companyId: '',
  name: '',
  address: '',
  timezone: 'Asia/Jakarta',
  latitude: 0,
  longitude: 0,
  radiusMeter: 100,
})
const isBranchEdit = ref(false)
const timezones = ['Asia/Jakarta', 'Asia/Makassar', 'Asia/Jayapura']

// --- DEPARTMENT STATE ---
const departments = ref<any[]>([])
const isDepartmentModalOpen = ref(false)
const departmentForm = ref({
  id: '',
  name: '',
  parentId: 'none',
})
const isDepartmentEdit = ref(false)

// --- FETCHING DATA ---
const fetchCompanies = async () => {
  try {
    const response = await apiClient.get('/companies')
    companies.value = response.data.data
  } catch (error) {
    console.error(error)
  }
}

const fetchBranches = async () => {
  try {
    const response = await apiClient.get('/branches')
    branches.value = response.data.data
  } catch (error) {
    console.error(error)
  }
}

const fetchDepartments = async () => {
  try {
    const response = await apiClient.get('/departments')
    departments.value = response.data.data
  } catch (error) {
    console.error(error)
  }
}

const fetchAll = async () => {
  isLoading.value = true
  await Promise.all([fetchCompanies(), fetchBranches(), fetchDepartments()])
  isLoading.value = false
}

onMounted(fetchAll)

// --- COMPANY ACTIONS ---
const openCompanyModal = async (company?: any) => {
  isCompanyEdit.value = !!company
  companyForm.value = company
    ? {
        id: company.id,
        name: company.name || '',
        taxNumber: company.tax_number || '',
        address: company.address || '',
        logoUrl: company.logo_url || '',
        branchIds: [],
      }
    : { id: '', name: '', taxNumber: '', address: '', logoUrl: '', branchIds: [] }
  if (company) {
    try {
      const response = await apiClient.get(`/companies/${company.id}/branches`)
      companyForm.value.branchIds = response.data.data || []
    } catch (error) {
      toast.error('Gagal memuat relasi cabang perusahaan.')
    }
  }
  isCompanyModalOpen.value = true
}

const saveCompany = async () => {
  isSubmitting.value = true
  try {
    if (isCompanyEdit.value) {
      await apiClient.put(`/companies/${companyForm.value.id}`, companyForm.value)
      toast.success('Perusahaan diperbarui.')
    } else {
      await apiClient.post('/companies', companyForm.value)
      toast.success('Perusahaan ditambahkan.')
    }
    isCompanyModalOpen.value = false
    fetchCompanies()
  } catch (e) {
    toast.error('Gagal menyimpan.')
  } finally {
    isSubmitting.value = false
  }
}

const deleteCompany = async (id: string) => {
  if (!confirm('Hapus perusahaan ini?')) return
  try {
    await apiClient.delete(`/companies/${id}`)
    toast.success('Dihapus.')
    fetchCompanies()
  } catch (e) {
    toast.error('Gagal menghapus.')
  }
}

// --- BRANCH ACTIONS ---
const openBranchModal = (branch?: any) => {
  isBranchEdit.value = !!branch
  branchForm.value = branch
    ? {
        id: branch.id,
        companyId: branch.company_id || '',
        name: branch.name || '',
        address: branch.address || '',
        timezone: branch.timezone || 'Asia/Jakarta',
        latitude: branch.latitude || 0,
        longitude: branch.longitude || 0,
        radiusMeter: branch.radius_meter || 100,
      }
    : {
        id: '',
        companyId: '',
        name: '',
        address: '',
        timezone: 'Asia/Jakarta',
        latitude: 0,
        longitude: 0,
        radiusMeter: 100,
      }
  isBranchModalOpen.value = true
}

const saveBranch = async () => {
  isSubmitting.value = true
  try {
    const payload = { ...branchForm.value, companyId: branchForm.value.companyId || null }
    if (isBranchEdit.value) {
      await apiClient.put(`/branches/${branchForm.value.id}`, payload)
      toast.success('Cabang diperbarui.')
    } else {
      await apiClient.post('/branches', payload)
      toast.success('Cabang ditambahkan.')
    }
    isBranchModalOpen.value = false
    fetchBranches()
  } catch (e) {
    toast.error('Gagal menyimpan.')
  } finally {
    isSubmitting.value = false
  }
}

const deleteBranch = async (id: string) => {
  if (!confirm('Hapus cabang ini?')) return
  try {
    await apiClient.delete(`/branches/${id}`)
    toast.success('Dihapus.')
    fetchBranches()
  } catch (e) {
    toast.error('Gagal menghapus.')
  }
}

// --- DEPARTMENT ACTIONS ---
const openDepartmentModal = (dept?: any) => {
  isDepartmentEdit.value = !!dept
  departmentForm.value = dept
    ? {
        id: dept.id,
        name: dept.name || '',
        parentId: dept.parent_id || 'none',
      }
    : { id: '', name: '', parentId: 'none' }
  isDepartmentModalOpen.value = true
}

const validParentOptions = computed(() => {
  if (!isDepartmentEdit.value) return departments.value
  return departments.value.filter((d: any) => d.id !== departmentForm.value.id)
})

const saveDepartment = async () => {
  isSubmitting.value = true
  try {
    const payload = {
      name: departmentForm.value.name,
      parentId: departmentForm.value.parentId === 'none' ? null : departmentForm.value.parentId,
    }
    if (isDepartmentEdit.value) {
      await apiClient.put(`/departments/${departmentForm.value.id}`, payload)
      toast.success('Departemen diperbarui.')
    } else {
      await apiClient.post('/departments', payload)
      toast.success('Departemen ditambahkan.')
    }
    isDepartmentModalOpen.value = false
    fetchDepartments()
  } catch (e) {
    toast.error('Gagal menyimpan.')
  } finally {
    isSubmitting.value = false
  }
}

const deleteDepartment = async (id: string) => {
  if (!confirm('Hapus departemen ini?')) return
  try {
    await apiClient.delete(`/departments/${id}`)
    toast.success('Dihapus.')
    fetchDepartments()
  } catch (e) {
    toast.error('Gagal menghapus.')
  }
}

// --- COLUMNS ---
const companyColumns = [
  {
    accessorKey: 'name',
    header: 'PERUSAHAAN',
    cell: (i: any) => h('span', { class: 'font-black text-slate-900' }, i.getValue() || '-'),
  },
  {
    accessorKey: 'tax_number',
    header: 'NPWP',
    cell: (i: any) => h('span', { class: 'text-slate-500 font-bold' }, i.getValue() || '-'),
  },
  {
    accessorKey: 'address',
    header: 'ALAMAT',
    cell: (i: any) =>
      h(
        'span',
        { class: 'text-slate-400 text-xs truncate max-w-[250px] inline-block' },
        i.getValue() || '-',
      ),
  },
  {
    id: 'actions',
    header: 'AKSI',
    cell: ({ row }: any) =>
      h('div', { class: 'flex gap-1' }, [
        h(
          Button,
          {
            variant: 'ghost',
            size: 'sm',
            class: 'h-8 px-2 text-primary hover:bg-primary/5',
            onClick: () => openCompanyModal(row.original),
          },
          () => h(Pencil, { class: 'w-4 h-4' }),
        ),
        h(
          Button,
          {
            variant: 'ghost',
            size: 'sm',
            class: 'h-8 px-2 text-rose-600 hover:bg-rose-50',
            onClick: () => deleteCompany(row.original.id),
          },
          () => h(Trash2, { class: 'w-4 h-4' }),
        ),
      ]),
  },
]

const branchColumns = [
  {
    accessorKey: 'name',
    header: 'CABANG',
    cell: (i: any) => h('span', { class: 'font-black text-slate-900' }, i.getValue() || '-'),
  },
  {
    id: 'company',
    header: 'PERUSAHAAN',
    cell: ({ row }: any) =>
      h('span', { class: 'text-slate-500 font-bold' }, row.original.company?.name || '-'),
  },
  {
    accessorKey: 'timezone',
    header: 'ZONA WAKTU',
    cell: (i: any) =>
      h(
        'span',
        { class: 'text-[10px] font-black uppercase bg-slate-100 px-2 py-1 rounded-lg' },
        i.getValue() || '-',
      ),
  },
  {
    id: 'actions',
    header: 'AKSI',
    cell: ({ row }: any) =>
      h('div', { class: 'flex gap-1' }, [
        h(
          Button,
          {
            variant: 'ghost',
            size: 'sm',
            class: 'h-8 px-2 text-primary hover:bg-primary/5',
            onClick: () => openBranchModal(row.original),
          },
          () => h(Pencil, { class: 'w-4 h-4' }),
        ),
        h(
          Button,
          {
            variant: 'ghost',
            size: 'sm',
            class: 'h-8 px-2 text-rose-600 hover:bg-rose-50',
            onClick: () => deleteBranch(row.original.id),
          },
          () => h(Trash2, { class: 'w-4 h-4' }),
        ),
      ]),
  },
]

const departmentColumns = [
  {
    accessorKey: 'name',
    header: 'DEPARTEMEN',
    cell: (i: any) => h('span', { class: 'font-black text-slate-900' }, i.getValue() || '-'),
  },
  {
    id: 'parent',
    header: 'INDUK',
    cell: ({ row }: any) => {
      const p = row.original.parent_department
      return h(
        'span',
        {
          class: p
            ? 'text-primary font-bold bg-primary/5 px-3 py-1 rounded-full text-xs'
            : 'text-slate-300 italic text-xs',
        },
        p?.name || 'Top Level',
      )
    },
  },
  {
    id: 'actions',
    header: 'AKSI',
    cell: ({ row }: any) =>
      h('div', { class: 'flex gap-1' }, [
        h(
          Button,
          {
            variant: 'ghost',
            size: 'sm',
            class: 'h-8 px-2 text-primary hover:bg-primary/5',
            onClick: () => openDepartmentModal(row.original),
          },
          () => h(Pencil, { class: 'w-4 h-4' }),
        ),
        h(
          Button,
          {
            variant: 'ghost',
            size: 'sm',
            class: 'h-8 px-2 text-rose-600 hover:bg-rose-50',
            onClick: () => deleteDepartment(row.original.id),
          },
          () => h(Trash2, { class: 'w-4 h-4' }),
        ),
      ]),
  },
]
</script>

<template>
  <div class="space-y-10 pb-20 animate-in fade-in duration-1000">
    <!-- Premium Navigation Tabs -->
    <div
      class="flex gap-4 p-2 bg-slate-100/50 rounded-[2rem] w-fit border border-slate-200/60 shadow-sm sticky top-0 z-30 backdrop-blur-md"
    >
      <button
        @click="activeTab = 'companies'"
        :class="[
          'px-8 py-4 rounded-[1.5rem] font-black text-xs uppercase tracking-[0.2em] transition-all flex items-center gap-3',
          activeTab === 'companies'
            ? 'bg-white text-primary shadow-md'
            : 'text-slate-400 hover:text-slate-600',
        ]"
      >
        <Building2 class="w-4 h-4" /> Perusahaan
      </button>
      <button
        @click="activeTab = 'branches'"
        :class="[
          'px-8 py-4 rounded-[1.5rem] font-black text-xs uppercase tracking-[0.2em] transition-all flex items-center gap-3',
          activeTab === 'branches'
            ? 'bg-white text-primary shadow-md'
            : 'text-slate-400 hover:text-slate-600',
        ]"
      >
        <MapPin class="w-4 h-4" /> Cabang
      </button>
      <button
        @click="activeTab = 'departments'"
        :class="[
          'px-8 py-4 rounded-[1.5rem] font-black text-xs uppercase tracking-[0.2em] transition-all flex items-center gap-3',
          activeTab === 'departments'
            ? 'bg-white text-primary shadow-md'
            : 'text-slate-400 hover:text-slate-600',
        ]"
      >
        <Network class="w-4 h-4" /> Departemen
      </button>
    </div>

    <!-- Dynamic Content Header -->
    <div
      class="flex flex-col lg:flex-row lg:items-end justify-between gap-8 border-b border-slate-100 pb-10"
    >
      <div class="space-y-4">
        <div class="flex items-center gap-4">
          <div
            class="w-14 h-14 bg-primary rounded-2xl flex items-center justify-center text-white shadow-2xl shadow-primary/20"
          >
            <Building2 v-if="activeTab === 'companies'" class="w-7 h-7" />
            <MapPin v-else-if="activeTab === 'branches'" class="w-7 h-7" />
            <Network v-else class="w-7 h-7" />
          </div>
          <div>
            <h1 class="text-4xl font-black text-slate-900 tracking-tighter">
              Struktur <span class="text-primary">Organisasi</span>
            </h1>
            <p class="text-[11px] font-black text-slate-400 uppercase tracking-[0.3em] mt-1">
              Master Data
              {{
                activeTab === 'companies'
                  ? 'Entitas Bisnis'
                  : activeTab === 'branches'
                    ? 'Lokasi Operasional'
                    : 'Hirarki Departemen'
              }}
            </p>
          </div>
        </div>
      </div>

      <Button
        @click="
          activeTab === 'companies'
            ? openCompanyModal()
            : activeTab === 'branches'
              ? openBranchModal()
              : openDepartmentModal()
        "
        class="h-16 px-10 rounded-[2rem] font-black gap-3 shadow-2xl shadow-primary/20 hover:scale-105 active:scale-95 transition-all"
      >
        <Plus class="w-5 h-5" />
        <span>TAMBAH DATA BARU</span>
      </Button>
    </div>

    <!-- Data Tables Area -->
    <div v-if="activeTab === 'companies'" class="animate-in slide-in-from-bottom-4 duration-700">
      <div
        class="bg-white rounded-[3rem] border border-slate-100 shadow-2xl shadow-slate-100/30 overflow-hidden"
      >
        <DataTable :data="companies" :columns="companyColumns" :isLoading="isLoading" />
      </div>
    </div>

    <div
      v-else-if="activeTab === 'branches'"
      class="animate-in slide-in-from-bottom-4 duration-700"
    >
      <div
        class="bg-white rounded-[3rem] border border-slate-100 shadow-2xl shadow-slate-100/30 overflow-hidden"
      >
        <DataTable :data="branches" :columns="branchColumns" :isLoading="isLoading" />
      </div>
    </div>

    <div v-else class="animate-in slide-in-from-bottom-4 duration-700">
      <div
        class="bg-white rounded-[3rem] border border-slate-100 shadow-2xl shadow-slate-100/30 overflow-hidden"
      >
        <DataTable :data="departments" :columns="departmentColumns" :isLoading="isLoading" />
      </div>
    </div>

    <!-- MODALS -->

    <!-- Company Modal -->
    <Dialog v-model:open="isCompanyModalOpen">
      <DialogContent class="sm:max-w-xl rounded-[3rem] p-0 overflow-hidden border-none shadow-3xl">
        <div class="bg-slate-900 p-10 text-white relative">
          <DialogHeader>
            <DialogTitle class="text-2xl font-black">{{
              isCompanyEdit ? 'Edit Data Perusahaan' : 'Tambah Data Perusahaan'
            }}</DialogTitle>
          </DialogHeader>
          <Globe class="absolute -right-10 -bottom-10 w-48 h-48 opacity-10" />
        </div>
        <div class="p-10 bg-white space-y-6">
          <div class="space-y-2">
            <label class="text-[11px] font-black text-slate-400 uppercase tracking-widest pl-1"
              >Nama Perusahaan</label
            >
            <Input
              v-model="companyForm.name"
              placeholder="Contoh: PT Maju Bersama"
              class="h-14 rounded-2xl bg-slate-50 border-none font-bold px-6"
            />
          </div>
          <div class="space-y-2">
            <label class="text-[11px] font-black text-slate-400 uppercase tracking-widest pl-1">Cabang Operasional</label>
            <div class="grid grid-cols-1 sm:grid-cols-2 gap-2 rounded-2xl bg-slate-50 p-4 max-h-48 overflow-y-auto">
              <label v-for="branch in branches" :key="branch.id" class="flex items-center gap-2 text-sm font-semibold text-slate-700">
                <input v-model="companyForm.branchIds" type="checkbox" :value="branch.id" class="rounded border-slate-300 text-primary" />
                <span>{{ branch.name }}</span>
              </label>
              <p v-if="branches.length === 0" class="text-xs text-slate-500">Belum ada data cabang.</p>
            </div>
            <p class="text-xs text-slate-500">Pilih cabang tempat PT ini beroperasi dan boleh mengelola transaksi keuangan.</p>
          </div>
          <div class="space-y-2">
            <label class="text-[11px] font-black text-slate-400 uppercase tracking-widest pl-1"
              >Nomor NPWP</label
            >
            <Input
              v-model="companyForm.taxNumber"
              placeholder="XX.XXX.XXX.X-XXX.XXX"
              class="h-14 rounded-2xl bg-slate-50 border-none font-bold px-6"
            />
          </div>
          <div class="space-y-2">
            <label class="text-[11px] font-black text-slate-400 uppercase tracking-widest pl-1"
              >Alamat Utama</label
            >
            <textarea
              v-model="companyForm.address"
              rows="3"
              class="w-full rounded-2xl bg-slate-50 border-none font-bold px-6 py-4 focus:ring-2 focus:ring-primary/20 outline-none"
            ></textarea>
          </div>
        </div>
        <DialogFooter class="p-10 pt-0 bg-white grid grid-cols-2 gap-4">
          <Button
            variant="ghost"
            @click="isCompanyModalOpen = false"
            class="rounded-2xl h-14 font-black text-slate-400"
            >BATAL</Button
          >
          <Button
            @click="saveCompany"
            :disabled="isSubmitting"
            class="rounded-2xl h-14 font-black shadow-xl shadow-primary/20"
          >
            {{ isSubmitting ? 'PROSES...' : 'SIMPAN DATA' }}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>

    <!-- Branch Modal -->
    <Dialog v-model:open="isBranchModalOpen">
      <DialogContent
        class="sm:max-w-2xl rounded-[3rem] p-0 overflow-hidden border-none shadow-3xl flex flex-col max-h-[95vh]"
      >
        <div class="bg-slate-900 p-10 text-white relative shrink-0">
          <DialogHeader>
            <DialogTitle class="text-2xl font-black">{{
              isBranchEdit ? 'Edit Cabang' : 'Tambah Cabang'
            }}</DialogTitle>
            <DialogDescription
              class="text-slate-400 font-bold uppercase text-[10px] tracking-widest mt-2 border-l-2 border-primary pl-4"
              >Master Lokasi Kantor</DialogDescription
            >
          </DialogHeader>
          <MapPin class="absolute -right-10 -bottom-10 w-48 h-48 opacity-10" />
        </div>
        <div class="p-10 bg-white space-y-6 overflow-y-auto custom-scrollbar flex-1">
          <div class="grid grid-cols-2 gap-6">
            <div class="space-y-2">
              <label class="text-[11px] font-black text-slate-400 uppercase tracking-widest pl-1"
                >Nama Cabang</label
              >
              <Input
                v-model="branchForm.name"
                placeholder="e.g. Kantor Pusat"
                class="h-14 rounded-2xl bg-slate-50 border-none font-bold px-6"
              />
            </div>
            <div class="space-y-2">
              <label class="text-[11px] font-black text-slate-400 uppercase tracking-widest pl-1"
                >Perusahaan</label
              >
              <Select v-model="branchForm.companyId">
                <SelectTrigger class="h-14 rounded-2xl bg-slate-50 border-none font-bold px-6">
                  <SelectValue placeholder="Pilih..." />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem v-for="c in companies" :key="c.id" :value="c.id">{{
                    c.name
                  }}</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>
          <div class="grid grid-cols-2 gap-6">
            <div class="space-y-2">
              <label class="text-[11px] font-black text-slate-400 uppercase tracking-widest pl-1"
                >Zona Waktu</label
              >
              <Select v-model="branchForm.timezone">
                <SelectTrigger class="h-14 rounded-2xl bg-slate-50 border-none font-bold px-6">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem v-for="tz in timezones" :key="tz" :value="tz">{{ tz }}</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div class="space-y-2">
              <label class="text-[11px] font-black text-slate-400 uppercase tracking-widest pl-1"
                >Radius Absen (m)</label
              >
              <Input
                type="number"
                v-model.number="branchForm.radiusMeter"
                class="h-14 rounded-2xl bg-slate-50 border-none font-bold px-6"
              />
            </div>
          </div>

          <div class="space-y-4">
            <label class="text-[11px] font-black text-slate-400 uppercase tracking-widest pl-1"
              >Penentuan Geofence</label
            >
            <div class="rounded-3xl border-2 border-slate-50 overflow-hidden shadow-inner">
              <MapPicker
                :latitude="branchForm.latitude"
                :longitude="branchForm.longitude"
                :radius="branchForm.radiusMeter"
                height="250px"
                @update:location="
                  (lat, lng) => {
                    branchForm.latitude = lat
                    branchForm.longitude = lng
                  }
                "
                @update:address="
                  (addr) => {
                    branchForm.address = addr
                  }
                "
              />
            </div>
          </div>

          <div class="space-y-2">
            <label class="text-[11px] font-black text-slate-400 uppercase tracking-widest pl-1"
              >Alamat Cabang</label
            >
            <textarea
              v-model="branchForm.address"
              rows="2"
              class="w-full rounded-2xl bg-slate-50 border-none font-bold px-6 py-4 focus:ring-2 focus:ring-primary/20 outline-none"
            ></textarea>
          </div>
        </div>
        <DialogFooter class="p-10 pt-0 bg-white grid grid-cols-2 gap-4 shrink-0">
          <Button
            variant="ghost"
            @click="isBranchModalOpen = false"
            class="rounded-2xl h-14 font-black text-slate-400"
            >BATAL</Button
          >
          <Button
            @click="saveBranch"
            :disabled="isSubmitting"
            class="rounded-2xl h-14 font-black shadow-xl shadow-primary/20"
          >
            {{ isSubmitting ? 'PROSES...' : 'SIMPAN CABANG' }}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>

    <!-- Department Modal -->
    <Dialog v-model:open="isDepartmentModalOpen">
      <DialogContent class="sm:max-w-xl rounded-[3rem] p-0 overflow-hidden border-none shadow-3xl">
        <div class="bg-slate-900 p-10 text-white relative">
          <DialogHeader>
            <DialogTitle class="text-2xl font-black">{{
              isDepartmentEdit ? 'Edit Departemen' : 'Tambah Departemen'
            }}</DialogTitle>
            <DialogDescription
              class="text-slate-400 font-bold uppercase text-[10px] tracking-widest mt-2 border-l-2 border-primary pl-4"
              >Master Hirarki Organisasi</DialogDescription
            >
          </DialogHeader>
          <Network class="absolute -right-10 -bottom-10 w-48 h-48 opacity-10" />
        </div>
        <div class="p-10 bg-white space-y-6">
          <div class="space-y-2">
            <label class="text-[11px] font-black text-slate-400 uppercase tracking-widest pl-1"
              >Nama Departemen</label
            >
            <Input
              v-model="departmentForm.name"
              placeholder="e.g. Sales & Marketing"
              class="h-14 rounded-2xl bg-slate-50 border-none font-bold px-6"
            />
          </div>
          <div class="space-y-2">
            <label class="text-[11px] font-black text-slate-400 uppercase tracking-widest pl-1"
              >Departemen Induk</label
            >
            <Select v-model="departmentForm.parentId">
              <SelectTrigger class="h-14 rounded-2xl bg-slate-50 border-none font-bold px-6">
                <SelectValue placeholder="Pilih Induk (Atau Top Level)" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="none">-- Top Level (Utama) --</SelectItem>
                <SelectItem v-for="d in validParentOptions" :key="d.id" :value="d.id">{{
                  d.name
                }}</SelectItem>
              </SelectContent>
            </Select>
          </div>
        </div>
        <DialogFooter class="p-10 pt-0 bg-white grid grid-cols-2 gap-4">
          <Button
            variant="ghost"
            @click="isDepartmentModalOpen = false"
            class="rounded-2xl h-14 font-black text-slate-400"
            >BATAL</Button
          >
          <Button
            @click="saveDepartment"
            :disabled="isSubmitting"
            class="rounded-2xl h-14 font-black shadow-xl shadow-primary/20"
          >
            {{ isSubmitting ? 'PROSES...' : 'SIMPAN DATA' }}
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
.no-scrollbar::-webkit-scrollbar {
  display: none;
}
.no-scrollbar {
  -ms-overflow-style: none;
  scrollbar-width: none;
}
</style>
