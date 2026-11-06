<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import apiClient from '@/api/axios'
import { toast } from 'vue-sonner'
import { useMasterDataStore } from '@/stores/masterData'
import {
  Palette,
  RefreshCw,
  Plus,
  Pencil,
  Trash2,
  Save,
  X,
  User,
  Mail,
  MapPin,
  CheckCircle2,
  XCircle,
  Phone,
  Search,
} from 'lucide-vue-next'
import { Button } from '@/components/ui/button'

const isLoading = ref(true)
const editors = ref<any[]>([])
const searchQuery = ref('')
const showModal = ref(false)
const isEditing = ref(false)

const masterStore = useMasterDataStore()

const filteredEditors = computed(() => {
  if (!searchQuery.value.trim()) return editors.value
  const q = searchQuery.value.toLowerCase().trim()
  return editors.value.filter(
    (e: any) =>
      e.name?.toLowerCase().includes(q) ||
      e.email?.toLowerCase().includes(q) ||
      e.area?.toLowerCase().includes(q) ||
      e.phone?.toLowerCase().includes(q),
  )
})

const currentEditor = ref({
  id: '',
  name: '',
  email: '',
  phone: '',
  area_code: '',
  is_active: true,
})

const fetchEditors = async () => {
  isLoading.value = true
  try {
    const branchId = masterStore.selectedBrandingBranchId
    const query = branchId && branchId !== 'ALL_BRANCHES' ? `?branch_id=${branchId}` : ''
    const res = await apiClient.get(`/branding/editors${query}`)
    if (res.data?.data) {
      editors.value = res.data.data
    } else {
      editors.value = []
    }
  } catch (error: any) {
    console.error(error)
    toast.error('Gagal mengambil data editor')
  } finally {
    isLoading.value = false
  }
}

watch(
  () => masterStore.selectedBrandingBranchId,
  () => {
    fetchEditors()
  },
)

onMounted(() => {
  fetchEditors()
})

const openAddModal = () => {
  isEditing.value = false
  currentEditor.value = {
    id: '',
    name: '',
    email: '',
    phone: '',
    area_code: '',
    is_active: true,
  }
  showModal.value = true
}

const handleEdit = (editor: any) => {
  isEditing.value = true
  currentEditor.value = {
    id: editor.id,
    name: editor.name,
    email: editor.email || '',
    phone: editor.phone || '',
    area_code: editor.area || '', // MySQL 'area_code' mapped to JSON 'area' in BrandingSalesAgent, wait, in BrandingSalesAgent it's json:"area"
    is_active: editor.is_active,
  }
  showModal.value = true
}

const handleDelete = async (id: string) => {
  if (!confirm('Yakin ingin menghapus editor ini?')) return
  try {
    await apiClient.delete(`/branding/editors/${id}`)
    toast.success('Editor berhasil dihapus')
    fetchEditors()
  } catch (error) {
    toast.error('Gagal menghapus editor')
  }
}

const handleSubmit = async () => {
  try {
    const payload = {
      name: currentEditor.value.name,
      email: currentEditor.value.email,
      phone: currentEditor.value.phone,
      area: currentEditor.value.area_code,
      is_active: currentEditor.value.is_active,
    }

    if (isEditing.value) {
      await apiClient.put(`/branding/editors/${currentEditor.value.id}`, payload)
      toast.success('Data editor diperbarui')
    } else {
      await apiClient.post('/branding/editors', payload)
      toast.success('Editor baru berhasil ditambahkan')
    }
    showModal.value = false
    fetchEditors()
  } catch (error) {
    toast.error('Gagal menyimpan data editor')
  }
}
</script>

<template>
  <div class="space-y-8 pb-10">
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-6">
      <div class="flex items-center gap-4">
        <div
          class="w-14 h-14 rounded-2xl bg-primary text-white flex items-center justify-center shadow-lg shadow-primary/30"
        >
          <Palette class="w-7 h-7" />
        </div>
        <div>
          <h1 class="text-3xl font-extrabold text-slate-900 tracking-tight flex items-center flex-wrap gap-2">
            <span>Kelola Editor</span>
            <span v-if="masterStore.selectedBrandingBranchName" class="text-primary text-2xl font-bold">
              — {{ masterStore.selectedBrandingBranchName }}
            </span>
          </h1>
          <p class="text-[15px] font-medium text-slate-500 mt-1.5 flex items-center gap-2">
            Manajemen data tim desain (Editor) untuk produksi spanduk.
          </p>
        </div>
      </div>
      <div class="flex items-center gap-3">
        <Button
          @click="fetchEditors"
          class="bg-white border-2 border-slate-100 text-slate-700 hover:bg-slate-50 rounded-2xl px-4 py-6 shadow-sm transition-all font-bold flex items-center gap-2"
        >
          <RefreshCw class="w-5 h-5" :class="{ 'animate-spin': isLoading }" /> Refresh
        </Button>
        <Button
          @click="openAddModal"
          class="bg-primary hover:bg-primary/90 text-white rounded-2xl px-6 py-6 font-bold shadow-lg shadow-primary/20 flex items-center gap-2"
        >
          <Plus class="w-5 h-5" /> Tambah Editor
        </Button>
      </div>
    </div>

    <!-- Search Toolbar -->
    <div class="flex flex-col sm:flex-row items-center justify-between gap-4">
      <div class="relative w-full sm:w-80">
        <Search class="w-4 h-4 text-slate-400 absolute left-3.5 top-1/2 -translate-y-1/2" />
        <input
          v-model="searchQuery"
          type="text"
          placeholder="Cari nama, email, atau area..."
          class="w-full pl-10 pr-4 py-2.5 bg-white border border-slate-200 rounded-2xl text-sm font-medium focus:ring-2 focus:ring-primary/20 focus:border-primary outline-none transition-all placeholder:text-slate-400 shadow-xs"
        />
      </div>
      <div class="text-xs font-bold text-slate-500 self-end sm:self-center">
        Total Desainer: <span class="text-slate-800 font-extrabold text-sm">{{ filteredEditors.length }}</span>
      </div>
    </div>

    <!-- Data Table -->
    <div
      class="bg-white rounded-[32px] border border-slate-200 shadow-sm overflow-hidden flex flex-col"
    >
      <div class="overflow-x-auto no-scrollbar">
        <table class="w-full border-collapse text-left">
          <thead class="bg-slate-50/50 border-b border-slate-100">
            <tr>
              <th class="p-6 text-[10px] font-black text-slate-400 uppercase tracking-widest">
                Nama Editor
              </th>
              <th class="p-6 text-[10px] font-black text-slate-400 uppercase tracking-widest">
                Kontak
              </th>
              <th class="p-6 text-[10px] font-black text-slate-400 uppercase tracking-widest">
                Kode Area
              </th>
              <th
                class="p-6 text-[10px] font-black text-slate-400 uppercase tracking-widest text-center"
              >
                Status
              </th>
              <th
                class="p-6 text-[10px] font-black text-slate-400 uppercase tracking-widest text-right"
              >
                Aksi
              </th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-50">
            <tr v-if="isLoading" v-for="i in 3" :key="i" class="animate-pulse">
              <td v-for="j in 5" :key="j" class="p-6">
                <div class="h-4 bg-slate-100 rounded-lg w-full"></div>
              </td>
            </tr>

            <tr v-else-if="filteredEditors.length === 0">
              <td colspan="5" class="py-24 text-center">
                <div class="flex flex-col items-center gap-3">
                  <Palette class="w-12 h-12 text-slate-200" />
                  <p class="font-bold text-slate-400">
                    {{ searchQuery ? 'Tidak ada editor yang cocok dengan pencarian.' : 'Belum ada data editor.' }}
                  </p>
                </div>
              </td>
            </tr>

            <tr
              v-for="editor in filteredEditors"
              :key="editor.id"
              class="group hover:bg-slate-50/50 transition-all"
            >
              <td class="p-6">
                <div class="flex items-center gap-4">
                  <div
                    class="w-10 h-10 rounded-xl bg-indigo-50 border border-indigo-100 text-indigo-500 flex items-center justify-center shrink-0"
                  >
                    <User class="w-5 h-5" />
                  </div>
                  <div>
                    <p class="text-sm font-black text-slate-900 leading-none">{{ editor.name }}</p>
                    <p class="text-[10px] font-bold text-slate-400 mt-1 uppercase tracking-tighter">
                      ID: {{ editor.id }}
                    </p>
                  </div>
                </div>
              </td>
              <td class="p-6">
                <div class="flex flex-col gap-1">
                  <div class="flex items-center gap-2">
                    <Mail class="w-3.5 h-3.5 text-slate-300" />
                    <p class="text-sm font-bold text-slate-700">{{ editor.email || '-' }}</p>
                  </div>
                  <div class="flex items-center gap-2">
                    <Phone class="w-3.5 h-3.5 text-slate-300" />
                    <p class="text-sm font-bold text-slate-700">{{ editor.phone || '-' }}</p>
                  </div>
                </div>
              </td>
              <td class="p-6">
                <div class="flex items-center gap-2">
                  <MapPin class="w-3.5 h-3.5 text-primary" />
                  <p class="text-sm font-bold text-slate-700 uppercase">
                    {{ editor.area || 'SEMUA AREA' }}
                  </p>
                </div>
              </td>
              <td class="p-6">
                <div class="flex justify-center">
                  <span
                    :class="[
                      'text-[10px] font-black px-3 py-1 rounded-full uppercase tracking-wider border flex items-center gap-1',
                      editor.is_active
                        ? 'bg-emerald-50 text-emerald-600 border-emerald-100'
                        : 'bg-red-50 text-red-600 border-red-100',
                    ]"
                  >
                    <CheckCircle2 v-if="editor.is_active" class="w-3 h-3" />
                    <XCircle v-else class="w-3 h-3" />
                    {{ editor.is_active ? 'AKTIF' : 'NON-AKTIF' }}
                  </span>
                </div>
              </td>
              <td class="p-6 text-right">
                <div class="flex justify-end gap-1">
                  <button
                    @click="handleEdit(editor)"
                    class="p-2 hover:bg-slate-100 rounded-xl transition-all text-slate-400 hover:text-primary"
                    title="Edit Editor"
                  >
                    <Pencil class="w-5 h-5" />
                  </button>
                  <button
                    @click="handleDelete(editor.id)"
                    class="p-2 hover:bg-red-50 rounded-xl transition-all text-red-400 hover:text-red-600"
                    title="Hapus Editor"
                  >
                    <Trash2 class="w-5 h-5" />
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Form Modal -->
    <div v-if="showModal" class="fixed inset-0 z-50 flex items-center justify-center p-4">
      <div
        class="absolute inset-0 bg-slate-900/60 backdrop-blur-sm"
        @click="showModal = false"
      ></div>
      <div
        class="bg-white rounded-[32px] w-full max-w-xl relative z-10 shadow-2xl flex flex-col max-h-[90vh] overflow-hidden animate-in fade-in zoom-in duration-300"
      >
        <!-- Header -->
        <div
          class="px-8 py-6 border-b border-slate-100 flex items-center justify-between bg-white shrink-0"
        >
          <div>
            <h3 class="text-2xl font-black text-slate-900 tracking-tight">
              {{ isEditing ? 'Edit Data Editor' : 'Tambah Editor Baru' }}
            </h3>
            <p class="text-[10px] font-black text-slate-400 uppercase tracking-widest mt-0.5">
              Informasi Profil Editor
            </p>
          </div>
          <button
            @click="showModal = false"
            class="p-2 hover:bg-slate-100 rounded-xl transition-all text-slate-400 hover:text-slate-900"
          >
            <X class="w-6 h-6" />
          </button>
        </div>

        <!-- Body -->
        <div class="p-8 overflow-y-auto space-y-6">
          <div class="space-y-2">
            <label class="text-[10px] font-black text-slate-400 uppercase tracking-widest ml-1"
              >Nama Editor</label
            >
            <input
              v-model="currentEditor.name"
              type="text"
              placeholder="Masukkan nama..."
              class="w-full bg-slate-50 border-2 border-transparent rounded-2xl px-5 py-4 font-bold text-slate-900 focus:bg-white focus:border-primary/20 focus:ring-0 transition-all"
            />
          </div>

          <div class="space-y-2">
            <label class="text-[10px] font-black text-slate-400 uppercase tracking-widest ml-1"
              >Email</label
            >
            <input
              v-model="currentEditor.email"
              type="email"
              placeholder="contoh@demo.local"
              class="w-full bg-slate-50 border-2 border-transparent rounded-2xl px-5 py-4 font-bold text-slate-900 focus:bg-white focus:border-primary/20 focus:ring-0 transition-all"
            />
          </div>

          <div class="space-y-2">
            <label class="text-[10px] font-black text-slate-400 uppercase tracking-widest ml-1"
              >Nomor Telepon</label
            >
            <input
              v-model="currentEditor.phone"
              type="text"
              placeholder="0812..."
              class="w-full bg-slate-50 border-2 border-transparent rounded-2xl px-5 py-4 font-bold text-slate-900 focus:bg-white focus:border-primary/20 focus:ring-0 transition-all"
            />
          </div>

          <div class="space-y-2">
            <label class="text-[10px] font-black text-slate-400 uppercase tracking-widest ml-1"
              >Kode Area (Opsional)</label
            >
            <input
              v-model="currentEditor.area_code"
              type="text"
              placeholder="Misal: JBR"
              class="w-full bg-slate-50 border-2 border-transparent rounded-2xl px-5 py-4 font-bold text-slate-900 focus:bg-white focus:border-primary/20 focus:ring-0 transition-all uppercase"
            />
            <p class="text-xs text-slate-500 mt-1 ml-1">
              Jika dikosongkan, editor dapat mengambil desain dari area mana saja.
            </p>
          </div>

          <div
            class="flex items-center gap-3 bg-slate-50 p-4 rounded-2xl border border-slate-100 mt-2"
          >
            <input
              type="checkbox"
              v-model="currentEditor.is_active"
              id="is_active"
              class="w-5 h-5 rounded-lg border-2 border-slate-300 text-primary focus:ring-primary/20 transition-all"
            />
            <label for="is_active" class="text-sm font-bold text-slate-700 cursor-pointer"
              >Status Editor Aktif</label
            >
          </div>
        </div>

        <!-- Footer -->
        <div class="p-8 border-t border-slate-100 flex gap-4 bg-white shrink-0">
          <button
            @click="showModal = false"
            class="flex-1 py-4 text-xs font-black text-slate-400 hover:text-slate-900 transition-all uppercase tracking-widest"
          >
            BATAL
          </button>
          <button
            @click="handleSubmit"
            class="flex-[2] bg-primary text-white rounded-2xl py-4 text-xs font-black shadow-lg shadow-primary/20 hover:scale-[1.02] active:scale-[0.98] transition-all flex items-center justify-center gap-2 uppercase tracking-widest"
          >
            <Save class="w-5 h-5" /> SIMPAN DATA EDITOR
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.custom-scrollbar::-webkit-scrollbar {
  width: 6px;
}
.custom-scrollbar::-webkit-scrollbar-track {
  background: transparent;
}
.custom-scrollbar::-webkit-scrollbar-thumb {
  background: #e2e8f0;
  border-radius: 10px;
}
.custom-scrollbar::-webkit-scrollbar-thumb:hover {
  background: #cbd5e1;
}
</style>
