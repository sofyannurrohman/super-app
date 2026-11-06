<template>
  <div class="p-6 lg:p-8 space-y-8 bg-[#f8fafc] min-h-screen">
    <!-- Header -->
    <div class="flex flex-col md:flex-row justify-between items-start md:items-center gap-6 bg-white p-6 lg:p-8 rounded-[2rem] shadow-sm border border-slate-100">
      <div class="flex items-center gap-5">
        <div class="w-14 h-14 rounded-[1.5rem] bg-gradient-to-br from-rose-700 to-rose-900 flex items-center justify-center text-white shadow-xl shadow-rose-200">
          <Sparkles class="h-7 w-7" />
        </div>
        <div>
          <h1 class="text-2xl font-black tracking-tight text-slate-900">Inventaris Tambahan</h1>
          <p class="text-sm text-slate-500 font-medium mt-0.5">Kelola stok inventaris non-produk (gelas, mangkuk, piring, dll)</p>
        </div>
      </div>
      <button
        @click="openCreateModal"
        class="flex items-center gap-2 bg-gradient-to-r from-rose-700 to-rose-900 text-white px-6 py-3 rounded-2xl font-bold text-sm shadow-lg shadow-rose-200 hover:shadow-xl hover:shadow-rose-300 hover:-translate-y-0.5 transition-all duration-200"
      >
        <Plus class="h-4 w-4" />
        Tambah Item
      </button>
    </div>

    <!-- Search -->
    <div class="flex flex-col sm:flex-row gap-3">
      <div class="relative flex-1">
        <Search class="absolute left-4 top-1/2 -translate-y-1/2 h-4 w-4 text-slate-400" />
        <input
          v-model="searchQuery"
          type="text"
          placeholder="Cari nama item..."
          class="w-full pl-11 pr-4 py-3.5 bg-white border border-slate-200 rounded-2xl text-sm font-medium text-slate-700 placeholder-slate-400 outline-none focus:ring-2 focus:ring-rose-700/20 focus:border-rose-400 transition-all"
        />
      </div>
    </div>

    <!-- Table -->
    <div class="bg-white rounded-[2rem] border border-slate-100 shadow-sm overflow-hidden">
      <div v-if="isLoading" class="flex items-center justify-center py-20">
        <div class="w-8 h-8 border-4 border-rose-700 border-t-transparent rounded-full animate-spin"></div>
      </div>

      <div v-else-if="filteredItems.length === 0" class="flex flex-col items-center justify-center py-20 text-slate-400">
        <Sparkles class="h-16 w-16 mb-4 opacity-20" />
        <p class="font-bold text-lg">Belum ada inventaris</p>
        <p class="text-sm">Tambahkan item baru untuk memulai</p>
      </div>

      <table v-else class="w-full">
        <thead>
          <tr class="border-b border-slate-100 bg-slate-50/50">
            <th class="px-6 py-4 text-left text-[10px] font-black uppercase tracking-widest text-slate-400">Nama Item</th>
            <th class="px-6 py-4 text-left text-[10px] font-black uppercase tracking-widest text-slate-400">Stok & Unit</th>
            <th class="px-6 py-4 text-left text-[10px] font-black uppercase tracking-widest text-slate-400">Deskripsi</th>
            <th class="px-6 py-4 text-right text-[10px] font-black uppercase tracking-widest text-slate-400">Aksi</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-slate-50">
          <tr
            v-for="item in filteredItems"
            :key="item.id"
            class="hover:bg-slate-50/50 transition-colors group"
          >
            <td class="px-6 py-4">
              <p class="font-bold text-sm text-slate-900">{{ item.name }}</p>
            </td>
            <td class="px-6 py-4">
              <span class="inline-flex items-center gap-1.5 px-3 py-1 bg-rose-50 text-rose-800 rounded-full font-bold text-xs border border-rose-100">
                {{ item.quantity }} {{ item.unit }}
              </span>
            </td>
            <td class="px-6 py-4">
              <p class="text-xs text-slate-500 truncate max-w-[200px]">{{ item.description || '-' }}</p>
            </td>
            <td class="px-6 py-4 text-right">
              <div class="flex items-center justify-end gap-2 opacity-0 group-hover:opacity-100 transition-opacity">
                <button
                  @click="openAdjustModal(item)"
                  class="p-2 text-slate-400 hover:text-emerald-600 hover:bg-emerald-50 rounded-xl transition-all"
                  title="Sesuaikan Stok"
                >
                  <ArrowUpDown class="h-4 w-4" />
                </button>
                <button
                  @click="openEditModal(item)"
                  class="p-2 text-slate-400 hover:text-rose-700 hover:bg-rose-50 rounded-xl transition-all"
                  title="Edit Info"
                >
                  <Pencil class="h-4 w-4" />
                </button>
                <button
                  @click="handleDelete(item)"
                  class="p-2 text-slate-400 hover:text-rose-600 hover:bg-rose-50 rounded-xl transition-all"
                  title="Hapus"
                >
                  <Trash2 class="h-4 w-4" />
                </button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Create/Edit Modal -->
    <Teleport to="body">
      <Transition name="modal">
        <div v-if="showModal" class="fixed inset-0 z-50 flex items-center justify-center p-4">
          <div class="absolute inset-0 bg-black/40 backdrop-blur-sm" @click="closeModal" />
          <div class="relative bg-white rounded-[2rem] shadow-2xl w-full max-w-md overflow-hidden">
            <div class="px-8 pt-8 pb-6 border-b border-slate-100">
              <div class="flex items-center justify-between">
                <div>
                  <h2 class="text-xl font-black text-slate-900">{{ isEditing ? 'Edit Item' : 'Tambah Item' }}</h2>
                </div>
                <button @click="closeModal" class="p-2 text-slate-400 hover:text-slate-700 hover:bg-slate-100 rounded-xl transition-all">
                  <X class="h-5 w-5" />
                </button>
              </div>
            </div>
            
            <form @submit.prevent="submitForm" class="p-8 space-y-5">
              <div>
                <label class="block text-xs font-black uppercase tracking-widest text-slate-500 mb-2">Nama Item</label>
                <input
                  v-model="formData.name"
                  type="text"
                  required
                  placeholder="Contoh: Gelas Kaca, Mangkuk Sup"
                  class="w-full px-4 py-3 bg-slate-50 border border-slate-200 rounded-xl text-sm font-medium focus:ring-2 focus:ring-rose-700/20 focus:border-rose-700 focus:bg-white transition-all outline-none"
                />
              </div>

              <div v-if="!masterDataStore.selectedBranchId || masterDataStore.selectedBranchId === 'ALL_BRANCHES'">
                <label class="block text-xs font-black uppercase tracking-widest text-slate-500 mb-2">Cabang</label>
                <select
                  v-model="formData.branch_id"
                  required
                  class="w-full px-4 py-3 bg-slate-50 border border-slate-200 rounded-xl text-sm font-medium focus:ring-2 focus:ring-rose-700/20 focus:border-rose-700 focus:bg-white transition-all outline-none appearance-none"
                >
                  <option value="ALL_BRANCHES" disabled>Pilih Cabang</option>
                  <option v-for="branch in masterDataStore.branches" :key="branch.id" :value="branch.id">
                    {{ branch.name }}
                  </option>
                </select>
              </div>

              <div class="grid grid-cols-2 gap-4">
                <div v-if="!isEditing">
                  <label class="block text-xs font-black uppercase tracking-widest text-slate-500 mb-2">Stok Awal</label>
                  <input
                    v-model.number="formData.quantity"
                    type="number"
                    min="0"
                    required
                    class="w-full px-4 py-3 bg-slate-50 border border-slate-200 rounded-xl text-sm font-medium focus:ring-2 focus:ring-rose-700/20 focus:border-rose-700 focus:bg-white transition-all outline-none"
                  />
                </div>
                <div :class="{ 'col-span-2': isEditing }">
                  <label class="block text-xs font-black uppercase tracking-widest text-slate-500 mb-2">Satuan Unit</label>
                  <input
                    v-model="formData.unit"
                    type="text"
                    required
                    placeholder="PCS, LUSIN, dll"
                    class="w-full px-4 py-3 bg-slate-50 border border-slate-200 rounded-xl text-sm font-medium focus:ring-2 focus:ring-rose-700/20 focus:border-rose-700 focus:bg-white transition-all outline-none uppercase"
                  />
                </div>
              </div>

              <div>
                <label class="block text-xs font-black uppercase tracking-widest text-slate-500 mb-2">Deskripsi</label>
                <textarea
                  v-model="formData.description"
                  rows="2"
                  placeholder="Opsional..."
                  class="w-full px-4 py-3 bg-slate-50 border border-slate-200 rounded-xl text-sm font-medium focus:ring-2 focus:ring-rose-700/20 focus:border-rose-700 focus:bg-white transition-all outline-none resize-none"
                ></textarea>
              </div>

              <div class="pt-2 flex gap-3">
                <button
                  type="button"
                  @click="closeModal"
                  class="flex-1 px-4 py-3 text-sm font-bold text-slate-700 bg-slate-100 rounded-xl hover:bg-slate-200 transition-colors"
                >
                  Batal
                </button>
                <button
                  type="submit"
                  :disabled="isSubmitting"
                  class="flex-1 px-4 py-3 text-sm font-bold text-white bg-rose-700 rounded-xl hover:bg-rose-800 transition-colors disabled:opacity-50"
                >
                  {{ isSubmitting ? 'Menyimpan...' : 'Simpan' }}
                </button>
              </div>
            </form>
          </div>
        </div>
      </Transition>
    </Teleport>

    <!-- Adjust Modal -->
    <Teleport to="body">
      <Transition name="modal">
        <div v-if="showAdjustModal" class="fixed inset-0 z-50 flex items-center justify-center p-4">
          <div class="absolute inset-0 bg-black/40 backdrop-blur-sm" @click="closeAdjustModal" />
          <div class="relative bg-white rounded-[2rem] shadow-2xl w-full max-w-sm overflow-hidden">
            <div class="px-8 pt-8 pb-6 border-b border-slate-100">
              <h2 class="text-xl font-black text-slate-900">Sesuaikan Stok</h2>
              <p class="text-sm text-slate-500">{{ adjustItemData?.name }}</p>
            </div>
            
            <form @submit.prevent="submitAdjust" class="p-8 space-y-5">
              <div>
                <label class="block text-xs font-black uppercase tracking-widest text-slate-500 mb-2">Stok Baru</label>
                <input
                  v-model.number="adjustForm.quantity"
                  type="number"
                  min="0"
                  required
                  class="w-full px-4 py-3 bg-slate-50 border border-slate-200 rounded-xl text-sm font-medium focus:ring-2 focus:ring-rose-700/20 focus:border-rose-700 focus:bg-white transition-all outline-none"
                />
              </div>

              <div>
                <label class="block text-xs font-black uppercase tracking-widest text-slate-500 mb-2">Catatan Perubahan</label>
                <input
                  v-model="adjustForm.notes"
                  type="text"
                  required
                  placeholder="Alasan penyesuaian..."
                  class="w-full px-4 py-3 bg-slate-50 border border-slate-200 rounded-xl text-sm font-medium focus:ring-2 focus:ring-rose-700/20 focus:border-rose-700 focus:bg-white transition-all outline-none"
                />
              </div>

              <div class="pt-2 flex gap-3">
                <button
                  type="button"
                  @click="closeAdjustModal"
                  class="flex-1 px-4 py-3 text-sm font-bold text-slate-700 bg-slate-100 rounded-xl hover:bg-slate-200 transition-colors"
                >
                  Batal
                </button>
                <button
                  type="submit"
                  :disabled="isSubmitting"
                  class="flex-1 px-4 py-3 text-sm font-bold text-white bg-rose-700 rounded-xl hover:bg-rose-800 transition-colors disabled:opacity-50"
                >
                  {{ isSubmitting ? 'Menyimpan...' : 'Simpan' }}
                </button>
              </div>
            </form>
          </div>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { Plus, Search, Pencil, Trash2, X, Sparkles, ArrowUpDown } from 'lucide-vue-next'
import { useMasterDataStore } from '@/stores/masterData'
import { toast } from 'vue-sonner'
import api from '@/api/axios'

interface BonusInventory {
  id: string
  branch_id: string
  name: string
  quantity: number
  unit: string
  description: string
  created_at: string
}

const masterDataStore = useMasterDataStore()
const items = ref<BonusInventory[]>([])
const isLoading = ref(true)
const searchQuery = ref('')

// Modal States
const showModal = ref(false)
const isEditing = ref(false)
const isSubmitting = ref(false)

const showAdjustModal = ref(false)
const adjustItemData = ref<BonusInventory | null>(null)
const adjustForm = ref({
  quantity: 0,
  notes: ''
})

const formData = ref({
  id: '',
  branch_id: '',
  name: '',
  quantity: 0,
  unit: 'PCS',
  description: ''
})

const filteredItems = computed(() => {
  let filtered = items.value
  
  if (searchQuery.value) {
    const q = searchQuery.value.toLowerCase()
    filtered = filtered.filter(i => 
      i.name.toLowerCase().includes(q) || 
      (i.description && i.description.toLowerCase().includes(q))
    )
  }
  
  return filtered
})

const fetchItems = async () => {
  if (!masterDataStore.selectedBranchId) {
    items.value = []
    isLoading.value = false
    return
  }
  
  isLoading.value = true
  try {
    const response = await api.get(`/bonus-inventory/branch/${masterDataStore.selectedBranchId}`)
    items.value = response.data || []
  } catch (error: any) {
    toast.error(error.response?.data?.error || 'Gagal memuat data inventaris')
  } finally {
    isLoading.value = false
  }
}

watch(() => masterDataStore.selectedBranchId, () => {
  fetchItems()
})

onMounted(() => {
  fetchItems()
})

const openCreateModal = () => {
  isEditing.value = false
  formData.value = {
    id: '',
    branch_id: masterDataStore.selectedBranchId || '',
    name: '',
    quantity: 0,
    unit: 'PCS',
    description: ''
  }
  showModal.value = true
}

const openEditModal = (item: BonusInventory) => {
  isEditing.value = true
  formData.value = {
    id: item.id,
    branch_id: item.branch_id,
    name: item.name,
    quantity: item.quantity,
    unit: item.unit,
    description: item.description || ''
  }
  showModal.value = true
}

const closeModal = () => {
  showModal.value = false
}

const submitForm = async () => {
  if (!formData.value.branch_id || formData.value.branch_id === 'ALL_BRANCHES') {
    toast.error('Silakan pilih cabang spesifik terlebih dahulu')
    return
  }

  isSubmitting.value = true
  try {
    if (isEditing.value) {
      await api.put(`/bonus-inventory/${formData.value.id}`, formData.value)
      toast.success('Data inventaris berhasil diperbarui')
    } else {
      const { id, ...postData } = formData.value
      await api.post('/bonus-inventory', postData)
      toast.success('Inventaris berhasil ditambahkan')
    }
    closeModal()
    fetchItems()
  } catch (error: any) {
    toast.error(error.response?.data?.error || 'Gagal menyimpan data')
  } finally {
    isSubmitting.value = false
  }
}

const handleDelete = async (item: BonusInventory) => {
  if (!confirm(`Apakah Anda yakin ingin menghapus ${item.name}?`)) return
  
  try {
    await api.delete(`/bonus-inventory/${item.id}`)
    toast.success('Inventaris berhasil dihapus')
    fetchItems()
  } catch (error: any) {
    toast.error(error.response?.data?.error || 'Gagal menghapus inventaris')
  }
}

const openAdjustModal = (item: BonusInventory) => {
  adjustItemData.value = item
  adjustForm.value = {
    quantity: item.quantity,
    notes: ''
  }
  showAdjustModal.value = true
}

const closeAdjustModal = () => {
  showAdjustModal.value = false
  adjustItemData.value = null
}

const submitAdjust = async () => {
  if (!adjustItemData.value) return
  
  isSubmitting.value = true
  try {
    await api.post(`/bonus-inventory/${adjustItemData.value.id}/adjust`, adjustForm.value)
    toast.success('Stok berhasil disesuaikan')
    closeAdjustModal()
    fetchItems()
  } catch (error: any) {
    toast.error(error.response?.data?.error || 'Gagal menyesuaikan stok')
  } finally {
    isSubmitting.value = false
  }
}
</script>

<style scoped>
.modal-enter-active,
.modal-leave-active {
  transition: all 0.3s ease;
}

.modal-enter-from,
.modal-leave-to {
  opacity: 0;
  transform: scale(0.95);
}
</style>
