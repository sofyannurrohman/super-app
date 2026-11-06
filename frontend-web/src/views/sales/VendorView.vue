<script setup lang="ts">
import { ref, onMounted } from 'vue'
import apiClient from '@/api/axios'
import { toast } from 'vue-sonner'
import {
  Building2,
  RefreshCw,
  Plus,
  Pencil,
  Trash2,
  Save,
  X,
  Mail,
  MapPin,
  Phone,
  Tag,
  Search,
  CheckCircle2,
  XCircle,
} from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import { useMasterDataStore } from '@/stores/masterData'

const masterDataStore = useMasterDataStore()

const isLoading = ref(true)
const vendors = ref<any[]>([])
const searchQuery = ref('')
const showModal = ref(false)
const isEditing = ref(false)
const isSubmitting = ref(false)

const currentVendor = ref({
  id: '',
  name: '',
  email: '',
  phone: '',
  address: '',
  price_per_meter: 0,
  is_active: true,
})

const fetchVendors = async () => {
  isLoading.value = true
  try {
    const query = searchQuery.value ? `?search=${encodeURIComponent(searchQuery.value)}` : ''
    const res = await apiClient.get(`/branding/vendors${query}`)
    if (res.data?.data) {
      vendors.value = res.data.data
    } else {
      vendors.value = []
    }
  } catch (error: any) {
    console.error(error)
    toast.error('Gagal mengambil data vendor percetakan')
  } finally {
    isLoading.value = false
  }
}

let searchTimeout: any = null
const onSearch = () => {
  if (searchTimeout) clearTimeout(searchTimeout)
  searchTimeout = setTimeout(() => {
    fetchVendors()
  }, 400)
}

onMounted(() => {
  fetchVendors()
})

const openAddModal = () => {
  isEditing.value = false
  currentVendor.value = {
    id: '',
    name: '',
    email: '',
    phone: '',
    address: '',
    price_per_meter: 0,
    is_active: true,
  }
  showModal.value = true
}

const openEditModal = (vendor: any) => {
  isEditing.value = true
  currentVendor.value = {
    id: vendor.id,
    name: vendor.name,
    email: vendor.email || '',
    phone: vendor.phone || '',
    address: vendor.address || '',
    price_per_meter: vendor.price_per_meter || 0,
    is_active: vendor.is_active !== undefined ? vendor.is_active : true,
  }
  showModal.value = true
}

const saveVendor = async () => {
  if (!currentVendor.value.name.trim()) {
    toast.error('Nama vendor percetakan wajib diisi')
    return
  }

  isSubmitting.value = true
  try {
    if (isEditing.value) {
      await apiClient.put(`/branding/vendors/${currentVendor.value.id}`, currentVendor.value)
      toast.success('Data vendor berhasil diperbarui')
    } else {
      await apiClient.post('/branding/vendors', currentVendor.value)
      toast.success('Vendor baru berhasil ditambahkan')
    }
    showModal.value = false
    fetchVendors()
  } catch (error: any) {
    console.error(error)
    toast.error(error.response?.data?.error || 'Gagal menyimpan vendor')
  } finally {
    isSubmitting.value = false
  }
}

const deleteVendor = async (id: string, name: string) => {
  if (!confirm(`Apakah Anda yakin ingin menghapus vendor "${name}"?`)) return

  try {
    await apiClient.delete(`/branding/vendors/${id}`)
    toast.success('Vendor berhasil dihapus')
    fetchVendors()
  } catch (error: any) {
    console.error(error)
    toast.error(error.response?.data?.error || 'Gagal menghapus vendor')
  }
}

const formatCurrency = (amount: number) => {
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    maximumFractionDigits: 0,
  }).format(amount)
}
</script>

<template>
  <div class="space-y-6 pb-12">
    <!-- Header -->
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
      <div>
        <h1 class="text-3xl font-extrabold text-slate-900 tracking-tight flex items-center flex-wrap gap-2">
          <Building2 class="w-8 h-8 text-primary" />
          <span>Kelola Vendor Percetakan</span>
          <span v-if="masterDataStore.selectedBranchName" class="text-primary text-2xl font-bold">
            — {{ masterDataStore.selectedBranchName }}
          </span>
        </h1>
        <p class="text-[15px] font-medium text-slate-500 mt-1">
          Master data vendor percetakan spanduk untuk alokasi konsisten di menu Purchasing.
        </p>
      </div>
      <div class="flex items-center gap-3">
        <Button
          @click="fetchVendors"
          variant="outline"
          class="flex items-center gap-2 rounded-xl bg-white shadow-sm font-bold text-slate-600 hover:bg-slate-50"
        >
          <RefreshCw :class="['w-4 h-4', { 'animate-spin': isLoading }]" />
          Refresh
        </Button>
        <Button
          @click="openAddModal"
          class="flex items-center gap-2 rounded-xl bg-primary text-white shadow-md shadow-primary/20 font-bold"
        >
          <Plus class="w-4 h-4" />
          Tambah Vendor
        </Button>
      </div>
    </div>

    <!-- Main Card -->
    <div class="bg-white border border-slate-200 rounded-3xl shadow-sm overflow-hidden">
      <!-- Toolbar -->
      <div class="p-5 border-b border-slate-200 bg-slate-50/50 flex flex-col sm:flex-row gap-4 justify-between items-center">
        <div class="relative w-full sm:w-96">
          <Search class="absolute left-3.5 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400" />
          <input
            type="text"
            v-model="searchQuery"
            @input="onSearch"
            placeholder="Cari nama vendor, kontak, atau alamat..."
            class="w-full pl-10 pr-4 py-2.5 bg-white border border-slate-200 rounded-xl text-sm font-medium focus:outline-none focus:ring-2 focus:ring-primary/20 focus:border-primary transition placeholder:text-slate-400 shadow-sm"
          />
        </div>
        <div class="text-xs font-bold text-slate-500">
          Total Vendor: <span class="text-slate-800 text-sm font-extrabold">{{ vendors.length }}</span>
        </div>
      </div>

      <!-- Table -->
      <div class="overflow-x-auto">
        <table class="w-full text-left border-collapse min-w-[700px]">
          <thead>
            <tr class="bg-slate-50 border-b border-slate-200">
              <th class="py-4 px-6 text-xs font-black text-slate-500 uppercase tracking-wider">Nama Vendor</th>
              <th class="py-4 px-6 text-xs font-black text-slate-500 uppercase tracking-wider">Kontak & Email</th>
              <th class="py-4 px-6 text-xs font-black text-slate-500 uppercase tracking-wider">Alamat Workshop</th>
              <th class="py-4 px-6 text-xs font-black text-slate-500 uppercase tracking-wider">Tarif Dasar / m²</th>
              <th class="py-4 px-6 text-xs font-black text-slate-500 uppercase tracking-wider text-center">Status</th>
              <th class="py-4 px-6 text-xs font-black text-slate-500 uppercase tracking-wider text-center">Aksi</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-100">
            <tr v-if="isLoading && vendors.length === 0">
              <td colspan="6" class="py-16 text-center">
                <RefreshCw class="w-8 h-8 text-primary animate-spin mx-auto mb-3" />
                <p class="text-sm font-medium text-slate-500">Memuat data vendor...</p>
              </td>
            </tr>
            <tr v-else-if="vendors.length === 0">
              <td colspan="6" class="py-16 text-center">
                <Building2 class="w-12 h-12 text-slate-300 mx-auto mb-3" />
                <p class="text-base font-bold text-slate-700">Belum ada data vendor percetakan</p>
                <p class="text-xs text-slate-400 mt-1">Klik tombol "Tambah Vendor" untuk membuat vendor baru.</p>
              </td>
            </tr>
            <tr
              v-for="vendor in vendors"
              :key="vendor.id"
              class="hover:bg-slate-50/60 transition-colors group"
            >
              <!-- Nama Vendor -->
              <td class="py-4 px-6 align-middle">
                <div class="flex items-center gap-3">
                  <div class="w-10 h-10 rounded-xl bg-primary/10 text-primary flex items-center justify-center font-black text-sm">
                    {{ vendor.name.substring(0, 2).toUpperCase() }}
                  </div>
                  <div>
                    <div class="font-extrabold text-slate-900 text-sm">{{ vendor.name }}</div>
                    <div class="text-[11px] text-slate-400 font-medium">ID: {{ vendor.id.substring(0, 8) }}...</div>
                  </div>
                </div>
              </td>

              <!-- Kontak -->
              <td class="py-4 px-6 align-middle">
                <div class="space-y-1">
                  <div v-if="vendor.phone" class="flex items-center gap-1.5 text-xs font-semibold text-slate-700">
                    <Phone class="w-3.5 h-3.5 text-emerald-500" />
                    <span>{{ vendor.phone }}</span>
                  </div>
                  <div v-if="vendor.email" class="flex items-center gap-1.5 text-xs font-medium text-slate-500">
                    <Mail class="w-3.5 h-3.5 text-slate-400" />
                    <span>{{ vendor.email }}</span>
                  </div>
                  <span v-if="!vendor.phone && !vendor.email" class="text-xs text-slate-400 italic">-</span>
                </div>
              </td>

              <!-- Alamat -->
              <td class="py-4 px-6 align-middle max-w-[240px]">
                <div class="flex items-start gap-1.5 text-xs text-slate-600 font-medium line-clamp-2" :title="vendor.address">
                  <MapPin v-if="vendor.address" class="w-3.5 h-3.5 text-rose-400 shrink-0 mt-0.5" />
                  <span>{{ vendor.address || '-' }}</span>
                </div>
              </td>

              <!-- Tarif Dasar -->
              <td class="py-4 px-6 align-middle">
                <div v-if="vendor.price_per_meter > 0" class="inline-flex items-center gap-1 px-2.5 py-1 rounded-lg bg-indigo-50 text-indigo-700 text-xs font-bold border border-indigo-100">
                  <Tag class="w-3 h-3" />
                  {{ formatCurrency(vendor.price_per_meter) }} / m²
                </div>
                <span v-else class="text-xs text-slate-400 italic">Fleksibel</span>
              </td>

              <!-- Status -->
              <td class="py-4 px-6 align-middle text-center">
                <span
                  class="inline-flex items-center gap-1 px-2.5 py-1 rounded-full text-xs font-black"
                  :class="vendor.is_active ? 'bg-emerald-50 text-emerald-600 border border-emerald-100' : 'bg-slate-100 text-slate-400 border border-slate-200'"
                >
                  <CheckCircle2 v-if="vendor.is_active" class="w-3 h-3" />
                  <XCircle v-else class="w-3 h-3" />
                  {{ vendor.is_active ? 'Aktif' : 'Nonaktif' }}
                </span>
              </td>

              <!-- Aksi -->
              <td class="py-4 px-6 align-middle text-center">
                <div class="flex items-center justify-center gap-2">
                  <button
                    @click="openEditModal(vendor)"
                    class="p-2 text-blue-600 hover:bg-blue-50 rounded-lg transition"
                    title="Edit Vendor"
                  >
                    <Pencil class="w-4 h-4" />
                  </button>
                  <button
                    @click="deleteVendor(vendor.id, vendor.name)"
                    class="p-2 text-rose-600 hover:bg-rose-50 rounded-lg transition"
                    title="Hapus Vendor"
                  >
                    <Trash2 class="w-4 h-4" />
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Modal Form Tambah / Edit Vendor -->
    <div v-if="showModal" class="fixed inset-0 z-50 flex items-center justify-center p-4">
      <div class="absolute inset-0 bg-slate-900/60 backdrop-blur-sm" @click="!isSubmitting && (showModal = false)"></div>
      <div class="bg-white rounded-3xl w-full max-w-lg relative z-10 shadow-2xl p-7 animate-in fade-in zoom-in duration-200">
        <div class="flex justify-between items-center pb-4 border-b border-slate-100 mb-6">
          <h3 class="text-xl font-black text-slate-900 flex items-center gap-2.5">
            <Building2 class="w-6 h-6 text-primary" />
            {{ isEditing ? 'Edit Vendor Percetakan' : 'Tambah Vendor Baru' }}
          </h3>
          <button
            v-if="!isSubmitting"
            @click="showModal = false"
            class="p-2 hover:bg-slate-100 rounded-xl text-slate-400 transition"
          >
            <X class="w-5 h-5" />
          </button>
        </div>

        <form @submit.prevent="saveVendor" class="space-y-4">
          <!-- Nama Vendor -->
          <div>
            <label class="block text-xs font-bold text-slate-700 uppercase tracking-wider mb-1.5">
              Nama Vendor Percetakan <span class="text-rose-500">*</span>
            </label>
            <input
              type="text"
              v-model="currentVendor.name"
              placeholder="Contoh: CV Sentosa Digital Print"
              required
              class="w-full px-4 py-2.5 bg-slate-50 border border-slate-200 rounded-xl text-sm font-medium text-slate-800 focus:bg-white focus:outline-none focus:border-primary transition"
            />
          </div>

          <!-- Telepon & Email -->
          <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <div>
              <label class="block text-xs font-bold text-slate-700 uppercase tracking-wider mb-1.5">
                No. Telepon / WhatsApp
              </label>
              <input
                type="text"
                v-model="currentVendor.phone"
                placeholder="08123456789"
                class="w-full px-4 py-2.5 bg-slate-50 border border-slate-200 rounded-xl text-sm font-medium text-slate-800 focus:bg-white focus:outline-none focus:border-primary transition"
              />
            </div>
            <div>
              <label class="block text-xs font-bold text-slate-700 uppercase tracking-wider mb-1.5">
                Email
              </label>
              <input
                type="email"
                v-model="currentVendor.email"
                placeholder="vendor@print.com"
                class="w-full px-4 py-2.5 bg-slate-50 border border-slate-200 rounded-xl text-sm font-medium text-slate-800 focus:bg-white focus:outline-none focus:border-primary transition"
              />
            </div>
          </div>

          <!-- Tarif Dasar per meter -->
          <div>
            <label class="block text-xs font-bold text-slate-700 uppercase tracking-wider mb-1.5">
              Tarif Dasar per m² (Opsional)
            </label>
            <div class="relative">
              <span class="absolute left-4 top-1/2 -translate-y-1/2 text-xs font-bold text-slate-400">Rp</span>
              <input
                type="number"
                v-model="currentVendor.price_per_meter"
                placeholder="0"
                min="0"
                step="500"
                class="w-full pl-10 pr-4 py-2.5 bg-slate-50 border border-slate-200 rounded-xl text-sm font-medium text-slate-800 focus:bg-white focus:outline-none focus:border-primary transition"
              />
            </div>
            <p class="text-[11px] text-slate-400 mt-1">Digunakan untuk estimasi otomatis biaya cetak banner.</p>
          </div>

          <!-- Alamat -->
          <div>
            <label class="block text-xs font-bold text-slate-700 uppercase tracking-wider mb-1.5">
              Alamat Workshop / Kantor
            </label>
            <textarea
              v-model="currentVendor.address"
              rows="2"
              placeholder="Alamat lengkap vendor percetakan..."
              class="w-full px-4 py-2.5 bg-slate-50 border border-slate-200 rounded-xl text-sm font-medium text-slate-800 focus:bg-white focus:outline-none focus:border-primary transition resize-none"
            ></textarea>
          </div>

          <!-- Status Aktif -->
          <div class="flex items-center gap-3 pt-2">
            <input
              type="checkbox"
              id="vendor_active"
              v-model="currentVendor.is_active"
              class="w-4 h-4 text-primary rounded border-slate-300 focus:ring-primary"
            />
            <label for="vendor_active" class="text-sm font-bold text-slate-700 select-none cursor-pointer">
              Vendor Aktif (Dapat dipilih saat purchasing)
            </label>
          </div>

          <!-- Buttons -->
          <div class="pt-4 flex gap-3">
            <Button
              type="button"
              variant="outline"
              @click="showModal = false"
              :disabled="isSubmitting"
              class="w-1/2 rounded-xl"
            >
              Batal
            </Button>
            <Button
              type="submit"
              :disabled="isSubmitting"
              class="w-1/2 rounded-xl bg-primary text-white font-bold flex items-center justify-center gap-2 shadow-md shadow-primary/20"
            >
              <RefreshCw v-if="isSubmitting" class="w-4 h-4 animate-spin" />
              <Save v-else class="w-4 h-4" />
              Simpan
            </Button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>
