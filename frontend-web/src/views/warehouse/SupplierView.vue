<template>
  <div class="p-6 lg:p-8 space-y-8 bg-[#f8fafc] min-h-screen">
    <!-- Header -->
    <div class="flex flex-col md:flex-row justify-between items-start md:items-center gap-6 bg-white p-6 lg:p-8 rounded-[2rem] shadow-sm border border-slate-100">
      <div class="flex items-center gap-5">
        <div class="w-14 h-14 rounded-[1.5rem] bg-gradient-to-br from-indigo-600 to-violet-600 flex items-center justify-center text-white shadow-xl shadow-indigo-200">
          <Building2 class="h-7 w-7" />
        </div>
        <div>
          <h1 class="text-2xl font-black tracking-tight text-slate-900">Data Supplier</h1>
          <p class="text-sm text-slate-500 font-medium mt-0.5">Kelola data supplier dan vendor pengadaan barang/bahan</p>
        </div>
      </div>
      <div class="flex items-center gap-3 flex-wrap">
        <button
          @click="openImportModal"
          class="flex items-center gap-2 bg-emerald-50 text-emerald-700 border border-emerald-200 px-5 py-3 rounded-2xl font-bold text-sm shadow-sm hover:bg-emerald-100 hover:shadow-md hover:-translate-y-0.5 transition-all duration-200"
        >
          <FileSpreadsheet class="h-4 w-4 text-emerald-600" />
          Import Excel
        </button>
        <button
          @click="openCreateModal"
          class="flex items-center gap-2 bg-gradient-to-r from-indigo-600 to-violet-600 text-white px-6 py-3 rounded-2xl font-bold text-sm shadow-lg shadow-indigo-200 hover:shadow-xl hover:shadow-indigo-300 hover:-translate-y-0.5 transition-all duration-200"
        >
          <Plus class="h-4 w-4" />
          Tambah Supplier
        </button>
      </div>
    </div>

    <!-- Search & Filter -->
    <div class="flex flex-col sm:flex-row gap-3">
      <div class="relative flex-1">
        <Search class="absolute left-4 top-1/2 -translate-y-1/2 h-4 w-4 text-slate-400" />
        <input
          v-model="searchQuery"
          type="text"
          placeholder="Cari nama, kota, produk, atau PIC supplier..."
          class="w-full pl-11 pr-4 py-3.5 bg-white border border-slate-200 rounded-2xl text-sm font-medium text-slate-700 placeholder-slate-400 outline-none focus:ring-2 focus:ring-indigo-500/20 focus:border-indigo-300 transition-all"
        />
      </div>
      <select
        v-model="filterActive"
        class="bg-white border border-slate-200 rounded-2xl px-4 py-3.5 text-sm font-bold text-slate-700 outline-none focus:ring-2 focus:ring-indigo-500/20 focus:border-indigo-300 transition-all"
      >
        <option value="">Semua Status</option>
        <option value="true">Aktif</option>
        <option value="false">Non-Aktif</option>
      </select>
    </div>

    <!-- Stats Cards -->
    <div class="grid grid-cols-2 lg:grid-cols-4 gap-4">
      <div v-for="stat in stats" :key="stat.label" class="bg-white p-5 rounded-2xl border border-slate-100 shadow-sm">
        <p class="text-xs font-black uppercase tracking-widest text-slate-400">{{ stat.label }}</p>
        <p class="text-3xl font-black text-slate-900 mt-1">{{ stat.value }}</p>
      </div>
    </div>

    <!-- Supplier Table -->
    <div class="bg-white rounded-[2rem] border border-slate-100 shadow-sm overflow-hidden">
      <div v-if="isLoading" class="flex items-center justify-center py-20">
        <div class="w-8 h-8 border-4 border-indigo-500 border-t-transparent rounded-full animate-spin"></div>
      </div>

      <div v-else-if="filteredSuppliers.length === 0" class="flex flex-col items-center justify-center py-20 text-slate-400">
        <Building2 class="h-16 w-16 mb-4 opacity-20" />
        <p class="font-bold text-lg">Belum ada supplier</p>
        <p class="text-sm">Tambahkan supplier baru atau import data dari file Excel</p>
      </div>

      <table v-else class="w-full">
        <thead>
          <tr class="border-b border-slate-100 bg-slate-50/50">
            <th class="px-6 py-4 text-left text-[10px] font-black uppercase tracking-widest text-slate-400">Supplier</th>
            <th class="px-6 py-4 text-left text-[10px] font-black uppercase tracking-widest text-slate-400 hidden md:table-cell">Kontak / PIC</th>
            <th class="px-6 py-4 text-left text-[10px] font-black uppercase tracking-widest text-slate-400 hidden lg:table-cell">Kota & Alamat</th>
            <th class="px-6 py-4 text-left text-[10px] font-black uppercase tracking-widest text-slate-400 hidden lg:table-cell">Produk / Bahan</th>
            <th class="px-6 py-4 text-left text-[10px] font-black uppercase tracking-widest text-slate-400">Status</th>
            <th class="px-6 py-4 text-right text-[10px] font-black uppercase tracking-widest text-slate-400">Aksi</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-slate-50">
          <tr
            v-for="supplier in filteredSuppliers"
            :key="supplier.id"
            class="hover:bg-slate-50/50 transition-colors group"
          >
            <td class="px-6 py-4">
              <div class="flex items-center gap-3">
                <div class="w-10 h-10 rounded-2xl bg-gradient-to-br from-indigo-100 to-violet-100 flex items-center justify-center shrink-0">
                  <span class="text-indigo-700 font-black text-sm">{{ supplier.name?.charAt(0) || 'S' }}</span>
                </div>
                <div>
                  <p class="font-bold text-sm text-slate-900">{{ supplier.name }}</p>
                  <p v-if="supplier.bank_name" class="text-xs text-slate-400">{{ supplier.bank_name }} - {{ supplier.bank_account || '' }}</p>
                </div>
              </div>
            </td>
            <td class="px-6 py-4 hidden md:table-cell">
              <p class="text-sm font-medium text-slate-700">{{ supplier.contact_name || '-' }}</p>
              <p class="text-xs text-slate-400">{{ supplier.phone || supplier.email || '-' }}</p>
            </td>
            <td class="px-6 py-4 hidden lg:table-cell">
              <p class="text-sm font-medium text-slate-700">{{ supplier.city || '-' }}</p>
              <p v-if="supplier.address" class="text-xs text-slate-400 max-w-[200px] truncate">{{ supplier.address }}</p>
            </td>
            <td class="px-6 py-4 hidden lg:table-cell">
              <p class="text-xs text-slate-500 max-w-[180px] truncate">{{ supplier.typical_products || '-' }}</p>
            </td>
            <td class="px-6 py-4">
              <span
                :class="supplier.is_active
                  ? 'bg-emerald-50 text-emerald-700 border border-emerald-200'
                  : 'bg-slate-100 text-slate-500 border border-slate-200'"
                class="inline-flex px-3 py-1 rounded-full text-[10px] font-black uppercase tracking-widest"
              >{{ supplier.is_active ? 'Aktif' : 'Non-Aktif' }}</span>
            </td>
            <td class="px-6 py-4 text-right">
              <div class="flex items-center justify-end gap-2 opacity-0 group-hover:opacity-100 transition-opacity">
                <button
                  @click="openEditModal(supplier)"
                  class="p-2 text-slate-400 hover:text-indigo-600 hover:bg-indigo-50 rounded-xl transition-all"
                  title="Edit"
                >
                  <Pencil class="h-4 w-4" />
                </button>
                <button
                  @click="handleDelete(supplier)"
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
          <div class="relative bg-white rounded-[2rem] shadow-2xl w-full max-w-2xl max-h-[90vh] overflow-y-auto">
            <!-- Modal Header -->
            <div class="sticky top-0 bg-white rounded-t-[2rem] px-8 pt-8 pb-6 border-b border-slate-100 z-10">
              <div class="flex items-center justify-between">
                <div>
                  <h2 class="text-xl font-black text-slate-900">{{ isEditing ? 'Edit Supplier' : 'Tambah Supplier Baru' }}</h2>
                  <p class="text-sm text-slate-500 mt-1">{{ isEditing ? 'Perbarui data supplier' : 'Isi data supplier dengan lengkap' }}</p>
                </div>
                <button @click="closeModal" class="p-2 text-slate-400 hover:text-slate-700 hover:bg-slate-100 rounded-xl transition-all">
                  <X class="h-5 w-5" />
                </button>
              </div>
            </div>

            <!-- Modal Body -->
            <div class="px-8 py-6 space-y-6">
              <!-- Section: Info Utama -->
              <div>
                <h3 class="text-xs font-black uppercase tracking-widest text-slate-400 mb-4">Informasi Utama</h3>
                <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
                  <div class="sm:col-span-2 space-y-1.5">
                    <label class="text-xs font-bold text-slate-500">Nama Supplier <span class="text-rose-500">*</span></label>
                    <input v-model="form.name" type="text" placeholder="PT. Supplier Indonesia" class="input-field" />
                  </div>
                  <div class="space-y-1.5">
                    <label class="text-xs font-bold text-slate-500">Nama Kontak / PIC</label>
                    <input v-model="form.contact_name" type="text" placeholder="Budi Santoso" class="input-field" />
                  </div>
                  <div class="space-y-1.5">
                    <label class="text-xs font-bold text-slate-500">Telepon / WhatsApp</label>
                    <input v-model="form.phone" type="tel" placeholder="08xx-xxxx-xxxx" class="input-field" />
                  </div>
                  <div class="space-y-1.5">
                    <label class="text-xs font-bold text-slate-500">Email</label>
                    <input v-model="form.email" type="email" placeholder="supplier@email.com" class="input-field" />
                  </div>
                  <div class="space-y-1.5">
                    <label class="text-xs font-bold text-slate-500">Kota</label>
                    <input v-model="form.city" type="text" placeholder="Jakarta" class="input-field" />
                  </div>
                  <div class="sm:col-span-2 space-y-1.5">
                    <label class="text-xs font-bold text-slate-500">Alamat Lengkap</label>
                    <textarea v-model="form.address" rows="2" placeholder="Jl. Industri No. 1..." class="input-field resize-none"></textarea>
                  </div>
                </div>
              </div>

              <!-- Section: Data Pajak & Bank -->
              <div>
                <h3 class="text-xs font-black uppercase tracking-widest text-slate-400 mb-4">Data Pajak & Bank</h3>
                <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
                  <div class="space-y-1.5">
                    <label class="text-xs font-bold text-slate-500">NPWP</label>
                    <input v-model="form.npwp" type="text" placeholder="xx.xxx.xxx.x-xxx.xxx" class="input-field" />
                  </div>
                  <div class="space-y-1.5">
                    <label class="text-xs font-bold text-slate-500">Nama Bank</label>
                    <input v-model="form.bank_name" type="text" placeholder="BCA, BRI, Mandiri..." class="input-field" />
                  </div>
                  <div class="space-y-1.5">
                    <label class="text-xs font-bold text-slate-500">No. Rekening</label>
                    <input v-model="form.bank_account" type="text" placeholder="xxx-xxx-xxxxx" class="input-field" />
                  </div>
                  <div class="space-y-1.5">
                    <label class="text-xs font-bold text-slate-500">Atas Nama Rekening</label>
                    <input v-model="form.bank_account_name" type="text" placeholder="PT. Supplier Indonesia" class="input-field" />
                  </div>
                </div>
              </div>

              <!-- Section: Produk & Catatan -->
              <div>
                <h3 class="text-xs font-black uppercase tracking-widest text-slate-400 mb-4">Produk & Catatan</h3>
                <div class="space-y-4">
                  <div class="space-y-1.5">
                    <label class="text-xs font-bold text-slate-500">Produk / Bahan Baku yang Disuplai</label>
                    <textarea v-model="form.typical_products" rows="2" placeholder="Kedelai, Botol PET, Garam..." class="input-field resize-none"></textarea>
                  </div>
                  <div class="space-y-1.5">
                    <label class="text-xs font-bold text-slate-500">Catatan</label>
                    <textarea v-model="form.notes" rows="2" placeholder="Catatan tambahan tentang supplier..." class="input-field resize-none"></textarea>
                  </div>
                  <div class="flex items-center gap-3">
                    <input v-model="form.is_active" type="checkbox" id="is_active" class="w-4 h-4 rounded text-indigo-600 focus:ring-indigo-500" />
                    <label for="is_active" class="text-sm font-bold text-slate-700">Supplier Aktif</label>
                  </div>
                </div>
              </div>
            </div>

            <!-- Modal Footer -->
            <div class="sticky bottom-0 bg-white rounded-b-[2rem] px-8 py-6 border-t border-slate-100 flex gap-3">
              <button @click="closeModal" class="flex-1 py-3 border-2 border-slate-200 rounded-2xl text-sm font-bold text-slate-600 hover:bg-slate-50 transition-all">
                Batal
              </button>
              <button
                @click="handleSubmit"
                :disabled="isSubmitting || !form.name"
                class="flex-1 py-3 bg-gradient-to-r from-indigo-600 to-violet-600 text-white rounded-2xl text-sm font-bold shadow-lg shadow-indigo-200 hover:shadow-xl hover:-translate-y-0.5 disabled:opacity-50 disabled:cursor-not-allowed disabled:transform-none transition-all"
              >
                <span v-if="isSubmitting" class="flex items-center justify-center gap-2">
                  <span class="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin"></span>
                  Menyimpan...
                </span>
                <span v-else>{{ isEditing ? 'Simpan Perubahan' : 'Tambah Supplier' }}</span>
              </button>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>

    <!-- Import Excel Modal -->
    <Teleport to="body">
      <Transition name="modal">
        <div v-if="showImportModal" class="fixed inset-0 z-50 flex items-center justify-center p-4">
          <div class="absolute inset-0 bg-black/40 backdrop-blur-sm" @click="closeImportModal" />
          <div class="relative bg-white rounded-[2rem] shadow-2xl w-full max-w-4xl max-h-[90vh] overflow-y-auto flex flex-col">
            <!-- Modal Header -->
            <div class="sticky top-0 bg-white rounded-t-[2rem] px-8 pt-8 pb-6 border-b border-slate-100 z-10">
              <div class="flex items-center justify-between">
                <div class="flex items-center gap-3">
                  <div class="w-10 h-10 rounded-2xl bg-emerald-100 text-emerald-700 flex items-center justify-center shrink-0">
                    <FileSpreadsheet class="w-5 h-5" />
                  </div>
                  <div>
                    <h2 class="text-xl font-black text-slate-900">Import Data Supplier dari Excel</h2>
                    <p class="text-sm text-slate-500 mt-0.5">Unggah file spreadsheet (.xlsx / .xls / .csv) untuk mengimpor data secara massal</p>
                  </div>
                </div>
                <button @click="closeImportModal" class="p-2 text-slate-400 hover:text-slate-700 hover:bg-slate-100 rounded-xl transition-all">
                  <X class="h-5 w-5" />
                </button>
              </div>
            </div>

            <!-- Modal Body -->
            <div class="p-8 space-y-6 flex-1">
              <!-- Download Template Banner -->
              <div class="p-5 rounded-2xl bg-slate-50 border border-slate-200/80 flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
                <div>
                  <p class="font-bold text-sm text-slate-800">Format Template Excel Supplier</p>
                  <p class="text-xs text-slate-500 mt-0.5">Gunakan format kolom yang sesuai agar data terbaca otomatis dan akurat.</p>
                </div>
                <button
                  type="button"
                  @click="downloadTemplate"
                  class="flex items-center gap-2 bg-white text-indigo-600 border border-indigo-200 px-4 py-2.5 rounded-xl font-bold text-xs shadow-sm hover:bg-indigo-50 transition-all shrink-0"
                >
                  <Download class="w-4 h-4" />
                  Download Template (.xlsx)
                </button>
              </div>

              <!-- Upload Area -->
              <div
                @dragover.prevent
                @drop.prevent="handleFileDrop"
                class="border-2 border-dashed rounded-2xl p-8 text-center transition-all cursor-pointer flex flex-col items-center justify-center bg-slate-50/50 hover:bg-indigo-50/30 hover:border-indigo-300 border-slate-300"
                @click="triggerFileInput"
              >
                <input
                  ref="fileInputRef"
                  type="file"
                  accept=".xlsx, .xls, .csv"
                  class="hidden"
                  @change="handleFileChange"
                />
                <div class="w-12 h-12 rounded-2xl bg-indigo-50 text-indigo-600 flex items-center justify-center mb-3">
                  <Upload class="w-6 h-6" />
                </div>
                <p class="font-bold text-sm text-slate-800">
                  {{ selectedFileName || 'Klik untuk pilih file Excel atau tarik file ke sini' }}
                </p>
                <p class="text-xs text-slate-400 mt-1">Mendukung format .xlsx, .xls, atau .csv (Maksimal 5MB)</p>
              </div>

              <!-- Preview Table -->
              <div v-if="parsedData.length > 0" class="space-y-3">
                <div class="flex items-center justify-between">
                  <div class="flex items-center gap-3">
                    <h3 class="text-xs font-black uppercase tracking-widest text-slate-400">Preview Data ({{ parsedData.length }} Baris)</h3>
                    <span
                      v-if="invalidRowCount > 0"
                      class="px-2.5 py-0.5 rounded-full text-[10px] font-bold bg-amber-50 text-amber-700 border border-amber-200 flex items-center gap-1"
                    >
                      <AlertCircle class="w-3 h-3" /> {{ invalidRowCount }} Baris Tanpa Nama (Akan Dilewati)
                    </span>
                    <span
                      v-else
                      class="px-2.5 py-0.5 rounded-full text-[10px] font-bold bg-emerald-50 text-emerald-700 border border-emerald-200 flex items-center gap-1"
                    >
                      <CheckCircle2 class="w-3 h-3" /> Semua Baris Valid
                    </span>
                  </div>
                </div>

                <div class="border border-slate-200 rounded-2xl overflow-hidden max-h-64 overflow-y-auto">
                  <table class="w-full text-left text-xs">
                    <thead class="bg-slate-100/70 border-b border-slate-200 sticky top-0">
                      <tr>
                        <th class="px-4 py-2.5 font-bold text-slate-600">No</th>
                        <th class="px-4 py-2.5 font-bold text-slate-600">Nama Supplier *</th>
                        <th class="px-4 py-2.5 font-bold text-slate-600">Kontak / PIC</th>
                        <th class="px-4 py-2.5 font-bold text-slate-600">Telepon</th>
                        <th class="px-4 py-2.5 font-bold text-slate-600">Kota</th>
                        <th class="px-4 py-2.5 font-bold text-slate-600">Bank & Rekening</th>
                        <th class="px-4 py-2.5 font-bold text-slate-600">Produk / Bahan</th>
                      </tr>
                    </thead>
                    <tbody class="divide-y divide-slate-100 bg-white">
                      <tr
                        v-for="(row, idx) in parsedData"
                        :key="idx"
                        :class="row.name ? 'hover:bg-slate-50' : 'bg-rose-50/50'"
                      >
                        <td class="px-4 py-2 font-semibold text-slate-400">{{ idx + 1 }}</td>
                        <td class="px-4 py-2 font-bold" :class="row.name ? 'text-slate-800' : 'text-rose-500 italic'">
                          {{ row.name || '(Wajib diisi)' }}
                        </td>
                        <td class="px-4 py-2 text-slate-600">{{ row.contact_name || '-' }}</td>
                        <td class="px-4 py-2 text-slate-600">{{ row.phone || '-' }}</td>
                        <td class="px-4 py-2 text-slate-600">{{ row.city || '-' }}</td>
                        <td class="px-4 py-2 text-slate-600">
                          <span v-if="row.bank_name || row.bank_account">{{ row.bank_name }} {{ row.bank_account }}</span>
                          <span v-else>-</span>
                        </td>
                        <td class="px-4 py-2 text-slate-500 truncate max-w-[150px]">{{ row.typical_products || '-' }}</td>
                      </tr>
                    </tbody>
                  </table>
                </div>
              </div>
            </div>

            <!-- Modal Footer -->
            <div class="sticky bottom-0 bg-white rounded-b-[2rem] px-8 py-6 border-t border-slate-100 flex gap-3">
              <button
                type="button"
                @click="closeImportModal"
                class="flex-1 py-3 border-2 border-slate-200 rounded-2xl text-sm font-bold text-slate-600 hover:bg-slate-50 transition-all"
              >
                Batal
              </button>
              <button
                type="button"
                @click="handleImportSubmit"
                :disabled="isImporting || validRowCount === 0"
                class="flex-1 py-3 bg-gradient-to-r from-emerald-600 to-teal-600 text-white rounded-2xl text-sm font-bold shadow-lg shadow-emerald-200 hover:shadow-xl hover:-translate-y-0.5 disabled:opacity-50 disabled:cursor-not-allowed disabled:transform-none transition-all"
              >
                <span v-if="isImporting" class="flex items-center justify-center gap-2">
                  <span class="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin"></span>
                  Mengimpor Data...
                </span>
                <span v-else>Import {{ validRowCount }} Supplier</span>
              </button>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import {
  Building2,
  Plus,
  Search,
  Pencil,
  Trash2,
  X,
  FileSpreadsheet,
  Download,
  Upload,
  CheckCircle2,
  AlertCircle
} from 'lucide-vue-next'
import { toast } from 'vue-sonner'
import * as XLSX from 'xlsx'
import apiClient from '@/api/axios'

const suppliers = ref<any[]>([])
const isLoading = ref(false)
const isSubmitting = ref(false)
const showModal = ref(false)
const isEditing = ref(false)
const editingId = ref<string | null>(null)
const searchQuery = ref('')
const filterActive = ref('')

// Import Excel states
const showImportModal = ref(false)
const isImporting = ref(false)
const selectedFileName = ref('')
const fileInputRef = ref<HTMLInputElement | null>(null)
const parsedData = ref<any[]>([])

const defaultForm = () => ({
  name: '',
  contact_name: '',
  phone: '',
  email: '',
  address: '',
  city: '',
  npwp: '',
  bank_name: '',
  bank_account: '',
  bank_account_name: '',
  typical_products: '',
  notes: '',
  is_active: true,
})

const form = ref(defaultForm())

const filteredSuppliers = computed(() => {
  return suppliers.value.filter(s => {
    const matchSearch = !searchQuery.value ||
      s.name?.toLowerCase().includes(searchQuery.value.toLowerCase()) ||
      s.city?.toLowerCase().includes(searchQuery.value.toLowerCase()) ||
      s.typical_products?.toLowerCase().includes(searchQuery.value.toLowerCase()) ||
      s.contact_name?.toLowerCase().includes(searchQuery.value.toLowerCase())
    const matchActive = filterActive.value === '' ||
      String(s.is_active) === filterActive.value
    return matchSearch && matchActive
  })
})

const stats = computed(() => [
  { label: 'Total Supplier', value: suppliers.value.length },
  { label: 'Aktif', value: suppliers.value.filter(s => s.is_active).length },
  { label: 'Non-Aktif', value: suppliers.value.filter(s => !s.is_active).length },
  { label: 'Hasil Filter', value: filteredSuppliers.value.length },
])

const validRowCount = computed(() => parsedData.value.filter(r => r.name && String(r.name).trim() !== '').length)
const invalidRowCount = computed(() => parsedData.value.length - validRowCount.value)

const fetchSuppliers = async () => {
  isLoading.value = true
  try {
    const res = await apiClient.get('/warehouse/suppliers')
    suppliers.value = res.data?.data || []
  } catch {
    toast.error('Gagal memuat data supplier')
  } finally {
    isLoading.value = false
  }
}

const openCreateModal = () => {
  isEditing.value = false
  editingId.value = null
  form.value = defaultForm()
  showModal.value = true
}

const openEditModal = (supplier: any) => {
  isEditing.value = true
  editingId.value = supplier.id
  form.value = {
    name: supplier.name || '',
    contact_name: supplier.contact_name || '',
    phone: supplier.phone || '',
    email: supplier.email || '',
    address: supplier.address || '',
    city: supplier.city || '',
    npwp: supplier.npwp || '',
    bank_name: supplier.bank_name || '',
    bank_account: supplier.bank_account || '',
    bank_account_name: supplier.bank_account_name || '',
    typical_products: supplier.typical_products || '',
    notes: supplier.notes || '',
    is_active: supplier.is_active !== false,
  }
  showModal.value = true
}

const closeModal = () => {
  showModal.value = false
}

const handleSubmit = async () => {
  if (!form.value.name) return
  isSubmitting.value = true
  try {
    if (isEditing.value && editingId.value) {
      await apiClient.put(`/warehouse/suppliers/${editingId.value}`, form.value)
      toast.success('Supplier berhasil diperbarui!')
    } else {
      await apiClient.post('/warehouse/suppliers', form.value)
      toast.success('Supplier berhasil ditambahkan!')
    }
    closeModal()
    await fetchSuppliers()
  } catch (err: any) {
    toast.error(err.response?.data?.error || 'Gagal menyimpan data supplier')
  } finally {
    isSubmitting.value = false
  }
}

const handleDelete = async (supplier: any) => {
  if (!confirm(`Hapus supplier "${supplier.name}"? Tindakan ini tidak bisa dibatalkan.`)) return
  try {
    await apiClient.delete(`/warehouse/suppliers/${supplier.id}`)
    toast.success('Supplier berhasil dihapus')
    await fetchSuppliers()
  } catch (err: any) {
    toast.error(err.response?.data?.error || 'Gagal menghapus supplier')
  }
}

// ─── Import Excel Methods ───────────────────────────────────────────────────

const openImportModal = () => {
  selectedFileName.value = ''
  parsedData.value = []
  showImportModal.value = true
}

const closeImportModal = () => {
  showImportModal.value = false
  selectedFileName.value = ''
  parsedData.value = []
}

const triggerFileInput = () => {
  fileInputRef.value?.click()
}

const downloadTemplate = () => {
  const headers = [
    'Nama Supplier *',
    'Nama PIC / Kontak',
    'No Telepon / WA',
    'Email',
    'Alamat',
    'Kota',
    'NPWP',
    'Nama Bank',
    'No Rekening',
    'Atas Nama Rekening',
    'Produk / Bahan yang Disuplai',
    'Catatan',
  ]

  const sampleRows = [
    [
      'PT. Sumber Bahan Utama',
      'Bpk. Joko',
      '081234567890',
      'joko@sumberbahan.com',
      'Jl. Industri Raya No. 12',
      'Surabaya',
      '01.234.567.8-012.000',
      'BCA',
      '1234567890',
      'PT Sumber Bahan Utama',
      'Kedelai, Gula Kelapa, Garam',
      'Supplier utama kedelai',
    ],
    [
      'CV. Plastik Jaya Abadi',
      'Ibu Maria',
      '081987654321',
      'sales@plastikjaya.com',
      'Jl. Rungkut Industri III/5',
      'Surabaya',
      '',
      'Mandiri',
      '9876543210',
      'CV Plastik Jaya Abadi',
      'Botol PET 135ml, Jerigen 5L',
      'TOP 30 hari',
    ],
  ]

  const ws = XLSX.utils.aoa_to_sheet([headers, ...sampleRows])

  // Set column widths
  ws['!cols'] = [
    { wch: 25 },
    { wch: 20 },
    { wch: 18 },
    { wch: 22 },
    { wch: 30 },
    { wch: 15 },
    { wch: 20 },
    { wch: 12 },
    { wch: 18 },
    { wch: 22 },
    { wch: 28 },
    { wch: 25 },
  ]

  const wb = XLSX.utils.book_new()
  XLSX.utils.book_append_sheet(wb, ws, 'Template Supplier')
  XLSX.writeFile(wb, 'Template_Import_Supplier_Karsa HR.xlsx')
}

const processFile = async (file: File) => {
  selectedFileName.value = file.name
  try {
    const data = await file.arrayBuffer()
    const workbook = XLSX.read(data, { type: 'array' })
    const sheetName = workbook.SheetNames[0]
    if (!sheetName) {
      toast.warning('Sheet Excel tidak ditemukan.')
      parsedData.value = []
      return
    }
    const worksheet = workbook.Sheets[sheetName]
    if (!worksheet) {
      toast.warning('Worksheet kosong.')
      parsedData.value = []
      return
    }
    const rawRows = XLSX.utils.sheet_to_json<any>(worksheet)

    if (!rawRows || rawRows.length === 0) {
      toast.warning('File Excel kosong atau tidak memiliki baris data.')
      parsedData.value = []
      return
    }

    const mapped = rawRows.map((row) => {
      // Find keys flexibly
      const findVal = (keys: string[]) => {
        for (const k of keys) {
          for (const rawKey of Object.keys(row)) {
            if (rawKey.trim().toLowerCase() === k.trim().toLowerCase()) {
              return row[rawKey] !== undefined && row[rawKey] !== null ? String(row[rawKey]).trim() : ''
            }
          }
        }
        return ''
      }

      return {
        name: findVal(['Nama Supplier *', 'Nama Supplier', 'nama_supplier', 'Supplier', 'Name']),
        contact_name: findVal(['Nama PIC / Kontak', 'Nama PIC', 'Kontak', 'PIC', 'Contact Name', 'contact_name']),
        phone: findVal(['No Telepon / WA', 'Telepon', 'No Telepon', 'No HP', 'Phone', 'phone']),
        email: findVal(['Email', 'email']),
        address: findVal(['Alamat', 'Alamat Lengkap', 'Address', 'address']),
        city: findVal(['Kota', 'City', 'city']),
        npwp: findVal(['NPWP', 'npwp']),
        bank_name: findVal(['Nama Bank', 'Bank', 'bank_name']),
        bank_account: findVal(['No Rekening', 'No. Rekening', 'Nomor Rekening', 'bank_account']),
        bank_account_name: findVal(['Atas Nama Rekening', 'Atas Nama', 'bank_account_name', 'Nama Rekening']),
        typical_products: findVal(['Produk / Bahan yang Disuplai', 'Produk', 'Bahan Baku', 'typical_products']),
        notes: findVal(['Catatan', 'Notes', 'notes', 'Keterangan']),
        is_active: true,
      }
    })

    parsedData.value = mapped
    toast.info(`${mapped.length} baris data berhasil dibaca dari file Excel`)
  } catch (err: any) {
    console.error('Error parsing Excel file:', err)
    toast.error('Gagal membaca file Excel. Pastikan format file valid.')
    parsedData.value = []
  }
}

const handleFileChange = (e: Event) => {
  const target = e.target as HTMLInputElement
  if (target.files && target.files[0]) {
    processFile(target.files[0])
  }
}

const handleFileDrop = (e: DragEvent) => {
  if (e.dataTransfer && e.dataTransfer.files && e.dataTransfer.files[0]) {
    processFile(e.dataTransfer.files[0])
  }
}

const handleImportSubmit = async () => {
  const validRows = parsedData.value.filter(r => r.name && String(r.name).trim() !== '')
  if (validRows.length === 0) {
    toast.error('Tidak ada baris data supplier yang valid untuk diimpor.')
    return
  }

  isImporting.value = true
  try {
    const res = await apiClient.post('/warehouse/suppliers/bulk', validRows)
    const count = res.data?.count || validRows.length
    toast.success(`Berhasil mengimpor ${count} data supplier!`)
    closeImportModal()
    await fetchSuppliers()
  } catch (err: any) {
    toast.error(err.response?.data?.error || 'Gagal mengimpor data supplier')
  } finally {
    isImporting.value = false
  }
}

onMounted(fetchSuppliers)
</script>

<style scoped>
.input-field {
  width: 100%;
  background: #f8fafc;
  border: 2px solid transparent;
  border-radius: 1rem;
  padding: 0.875rem 1.25rem;
  font-size: 0.875rem;
  font-weight: 600;
  color: #0f172a;
  outline: none;
  transition: all 0.2s;
}
.input-field:focus {
  border-color: rgba(99, 102, 241, 0.3);
  background: white;
}

.modal-enter-active,
.modal-leave-active {
  transition: all 0.25s ease;
}
.modal-enter-from,
.modal-leave-to {
  opacity: 0;
}
.modal-enter-from .bg-white,
.modal-leave-to .bg-white {
  transform: scale(0.95) translateY(8px);
}
</style>

