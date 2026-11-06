<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import {
  Users,
  Search,
  MapPin,
  CheckCircle2,
  AlertCircle,
  Camera,
  Calendar,
  Grid,
  TrendingUp,
  Image,
  ExternalLink,
  ZoomIn,
  ZoomOut,
} from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import { toast } from 'vue-sonner'
import apiClient from '@/api/axios'
import { useMasterDataStore } from '@/stores/masterData'

const masterDataStore = useMasterDataStore()

const searchQuery = ref('')
const selectedStatus = ref<'ALL' | 'COMPLIANT' | 'NON_COMPLIANT' | 'PENDING'>('ALL')

// Real Merchandising Visits & Planogram submissions from API
const mdSubmissions = ref<any[]>([])
const keyAccounts = ref<any[]>([])

// Image Preview State
const previewImageUrl = ref('')
const zoomLevel = ref(1)

const openPreview = (url: string) => {
  previewImageUrl.value = url
  zoomLevel.value = 1
}

const closePreview = () => {
  previewImageUrl.value = ''
}

const zoomIn = () => {
  if (zoomLevel.value < 4) zoomLevel.value += 0.25
}
const zoomOut = () => {
  if (zoomLevel.value > 0.5) zoomLevel.value -= 0.25
}

const fetchMerchandising = async () => {
  try {
    const res = await apiClient.get('/modern-trade/merchandising?all_branches=true')
    mdSubmissions.value = (res.data.data || []).map((sub: any) => {
      let oosList: string[] = []
      if (sub.out_of_stock_skus) {
        try {
          oosList = JSON.parse(sub.out_of_stock_skus)
        } catch {
          // If not valid JSON, treat as comma separated or raw text
          oosList = sub.out_of_stock_skus.split(',').map((s: string) => s.trim()).filter(Boolean)
        }
      }
      const cleanUrl = (url: string) => {
        if (!url) return '';
        let cleaned = url.replace(/([^:]\/)\/+/g, "$1");
        cleaned = cleaned.replace('/v1/uploads', '/api/v1/uploads');
        return cleaned;
      }

      return {
        ...sub,
        store_name: sub.store_name || 'Ritel Modern Store',
        visit_time: sub.visit_time ? sub.visit_time.replace('T', ' ').substring(0, 16) : '',
        out_of_stock_skus: oosList,
        photo_url: cleanUrl(sub.photo_url),
        invoice_photo_url: cleanUrl(sub.invoice_photo_url),
        receipt_photo_url: cleanUrl(sub.receipt_photo_url),
      }
    })
  } catch (err: any) {
    toast.error('Gagal mengambil data Merchandising: ' + (err.response?.data?.error || err.message))
  }
}

const fetchKeyAccounts = async () => {
  try {
    const res = await apiClient.get('/modern-trade/key-accounts')
    keyAccounts.value = res.data.data || []
  } catch (err: any) {
    toast.error('Gagal mengambil data Key Account: ' + (err.response?.data?.error || err.message))
  }
}

onMounted(() => {
  fetchMerchandising()
  fetchKeyAccounts()
})

const filteredSubmissions = computed(() => {
  return mdSubmissions.value.filter((sub) => {
    const matchesSearch =
      sub.merchandiser_name.toLowerCase().includes(searchQuery.value.toLowerCase()) ||
      sub.store_name.toLowerCase().includes(searchQuery.value.toLowerCase())
    const matchesStatus =
      selectedStatus.value === 'ALL' || sub.status === selectedStatus.value
    return matchesSearch && matchesStatus
  })
})

const totalLaporan = computed(() => mdSubmissions.value.length)

const uniqueStores = computed(() => {
  const stores = new Set(mdSubmissions.value.map(s => s.store_name))
  return stores.size
})

const uniqueMDs = computed(() => {
  const mds = new Set(mdSubmissions.value.map(s => s.merchandiser_name))
  return mds.size
})

// Trigger a real photo upload dynamically saved to database
const isUploadingPhoto = ref(false)

const uploadPhoto = async (event: Event, field: 'photo_url' | 'invoice_photo_url' | 'receipt_photo_url') => {
  const target = event.target as HTMLInputElement
  const file = target.files?.[0]
  if (!file) return

  const formData = new FormData()
  formData.append('file', file)

  isUploadingPhoto.value = true
  try {
    const res = await apiClient.post('/modern-trade/merchandising/upload-photo', formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
    })
    const baseUrl = import.meta.env.VITE_API_BASE_URL || ''
    const rawUrl = baseUrl + res.data.url
    mdForm.value[field] = rawUrl.replace(/([^:]\/)\/+/g, "$1")
    toast.success('Foto berhasil diunggah')
  } catch (err: any) {
    toast.error('Gagal mengunggah foto: ' + (err.response?.data?.error || err.message))
  } finally {
    isUploadingPhoto.value = false
    target.value = '' // reset input
  }
}

// Repeat Order Action
const createRepeatOrder = async (sub: any) => {
  if (!confirm(`Buat Repeat Order (Purchase Order) untuk toko ${sub.store_name}?`)) return
  try {
    const poDate = new Date()
    poDate.setDate(poDate.getDate() + 7) // expected delivery 7 days later
    
    const payload = {
      key_account_id: sub.key_account_id,
      po_number: `PO-REP-${Date.now()}`,
      expected_delivery: poDate.toISOString().substring(0, 10),
      total_amount: 0,
      status: 'DRAFT',
      invoice_photo_url: sub.invoice_photo_url,
      receipt_photo_url: sub.receipt_photo_url
    }
    
    await apiClient.post('/modern-trade/purchase-orders', payload)
    toast.success('Repeat Order berhasil dibuat (Draft Purchase Order)! Cek tab Purchase Order untuk detailnya.')
  } catch (err: any) {
    toast.error('Gagal membuat Repeat Order: ' + (err.response?.data?.error || err.message))
  }
}

// Modal CRUD setup
const showMDModal = ref(false)
const isEditing = ref(false)
const mdForm = ref({
  id: '',
  key_account_id: '',
  merchandiser_name: '',
  store_name: '',
  visit_time: '',
  shelf_share_percent: 0,
  planogram_compliant: true,
  out_of_stock_skus: '',
  photo_url: '',
  supermarket_name: '',
  md_name: '',
  stock_update_log: '',
  promo_update_log: '',
  invoice_photo_url: '',
  receipt_photo_url: '',
})

const openNewMDModal = () => {
  isEditing.value = false
  const now = new Date().toISOString().substring(0, 16)
  mdForm.value = {
    id: '',
    key_account_id: keyAccounts.value[0]?.id || '',
    merchandiser_name: '',
    store_name: '',
    visit_time: now,
    shelf_share_percent: 0,
    planogram_compliant: true,
    out_of_stock_skus: '',
    photo_url: 'https://images.unsplash.com/photo-1534482421-64566f976cfa?auto=format&fit=crop&w=600&q=80',
    supermarket_name: '',
    md_name: '',
    stock_update_log: '',
    promo_update_log: '',
    invoice_photo_url: '',
    receipt_photo_url: '',
  }
  showMDModal.value = true
}

const openEditMDModal = (sub: any) => {
  isEditing.value = true
  mdForm.value = {
    ...sub,
    out_of_stock_skus: Array.isArray(sub.out_of_stock_skus) ? sub.out_of_stock_skus.join(', ') : sub.out_of_stock_skus,
    visit_time: sub.visit_time ? sub.visit_time.replace(' ', 'T').substring(0, 16) : '',
  }
  showMDModal.value = true
}

const saveMD = async () => {
  if (!mdForm.value.merchandiser_name || !mdForm.value.store_name) {
    toast.error('Nama Merchandiser dan Nama Toko wajib diisi!')
    return
  }

  try {
    const payload = {
      ...mdForm.value,
      shelf_share_percent: Number(mdForm.value.shelf_share_percent),
    }

    if (isEditing.value) {
      await apiClient.put(`/modern-trade/merchandising/${mdForm.value.id}`, payload)
      toast.success('Laporan Merchandising berhasil diperbarui!')
    } else {
      await apiClient.post('/modern-trade/merchandising', payload)
      toast.success('Laporan Merchandising baru didaftarkan!')
    }
    showMDModal.value = false
    fetchMerchandising()
  } catch (err: any) {
    toast.error('Gagal menyimpan laporan Merchandising: ' + (err.response?.data?.error || err.message))
  }
}

const deleteMD = async (id: string) => {
  if (confirm('Apakah Anda yakin ingin menghapus laporan audit MD ini?')) {
    try {
      await apiClient.delete(`/modern-trade/merchandising/${id}?all_branches=true`)
      toast.success('Laporan audit MD berhasil dihapus!')
      fetchMerchandising()
    } catch (err: any) {
      toast.error('Gagal menghapus laporan audit MD: ' + (err.response?.data?.error || err.message))
    }
  }
}
</script>

<template>
  <div class="space-y-6">
    <!-- Header Section -->
    <div class="flex flex-col md:flex-row md:items-center justify-between gap-4">
      <div>
        <div class="flex items-center gap-2">
          <span class="px-2.5 py-1 text-xs font-bold text-indigo-600 bg-indigo-50 border border-indigo-200/50 rounded-full">
            Modern Trade (MT)
          </span>
        </div>
        <h1 class="text-2xl md:text-3xl font-black text-slate-900 tracking-tight mt-1 flex items-center flex-wrap gap-2">
          <span>Merchandising & Planogram Compliance</span>
          <span v-if="masterDataStore.selectedBranchName" class="text-primary text-xl md:text-2xl font-bold">
            — {{ masterDataStore.selectedBranchName }}
          </span>
        </h1>
        <p class="text-sm text-slate-500 mt-1">
          Pantau penataan produk, kepatuhan planogram rak display, dan ketersediaan stok produk (OOS) real-time di gerai ritel modern.
        </p>
      </div>

      <div class="flex items-center gap-2.5">
        <Button
          @click="openNewMDModal"
          class="bg-red-800 hover:bg-red-600 text-white font-bold rounded-xl shadow-md border border-red-200"
        >
          <Plus class="w-4 h-4 mr-2" v-if="true" />
          Tambah Laporan MD
        </Button>
      </div>
    </div>

    <!-- Quick Stats -->
    <div class="grid grid-cols-1 sm:grid-cols-3 gap-4">
      <div class="bg-white p-5 rounded-2xl border border-slate-100 shadow-sm flex items-center gap-4">
        <div class="p-3 bg-indigo-50 text-indigo-600 rounded-xl">
          <Grid class="w-6 h-6" />
        </div>
        <div>
          <p class="text-xs font-bold text-slate-400 uppercase tracking-wider">Total Laporan MD</p>
          <p class="text-2xl font-black text-slate-800">{{ totalLaporan }} Laporan</p>
        </div>
      </div>
      <div class="bg-white p-5 rounded-2xl border border-slate-100 shadow-sm flex items-center gap-4">
        <div class="p-3 bg-emerald-50 text-emerald-600 rounded-xl">
          <MapPin class="w-6 h-6" />
        </div>
        <div>
          <p class="text-xs font-bold text-slate-400 uppercase tracking-wider">Toko Dikunjungi</p>
          <p class="text-2xl font-black text-slate-800">{{ uniqueStores }} Toko</p>
        </div>
      </div>
      <div class="bg-white p-5 rounded-2xl border border-slate-100 shadow-sm flex items-center gap-4">
        <div class="p-3 bg-blue-50 text-blue-600 rounded-xl">
          <Users class="w-6 h-6" />
        </div>
        <div>
          <p class="text-xs font-bold text-slate-400 uppercase tracking-wider">MD Aktif</p>
          <p class="text-2xl font-black text-slate-800">{{ uniqueMDs }} Personil</p>
        </div>
      </div>
    </div>

    <!-- Filters -->
    <div class="bg-white rounded-2xl border border-slate-100 shadow-sm p-4 flex flex-col sm:flex-row sm:items-center justify-between gap-4">
      <div class="relative w-full sm:w-[320px]">
        <Search class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400" />
        <input
          v-model="searchQuery"
          type="text"
          placeholder="Cari MD atau nama toko..."
          class="w-full pl-9 pr-4 py-2 text-xs border border-slate-200 rounded-xl focus:outline-none focus:ring-1 focus:ring-indigo-500"
        />
      </div>

      <div class="flex items-center gap-2">
        <select
          v-model="selectedStatus"
          class="px-3 py-2 text-xs border border-slate-200 rounded-xl focus:outline-none focus:ring-1 focus:ring-indigo-500 bg-white"
        >
          <option value="ALL">Semua Status Audit</option>
          <option value="PENDING">Menunggu Verifikasi (Pending)</option>
          <option value="COMPLIANT">Patuh Planogram</option>
          <option value="NON_COMPLIANT">Tidak Patuh Planogram</option>
        </select>
      </div>
    </div>

    <!-- Data Table -->
    <div class="bg-white rounded-2xl border border-slate-100 shadow-sm overflow-hidden">
      <div class="overflow-x-auto">
        <table class="w-full border-collapse text-left">
          <thead>
            <tr class="bg-slate-50/70 border-b border-slate-100 text-xs font-black text-slate-400 uppercase tracking-wider">
              <th class="py-4 px-6">Toko & MD</th>
              <th class="py-4 px-6">Waktu Kunjungan</th>
              <th class="py-4 px-6 text-center">Performa & Planogram</th>
              <th class="py-4 px-6">Status Stok (OOS)</th>
              <th class="py-4 px-6 text-center">Aksi</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-50 text-slate-700 text-sm">
            <tr v-for="sub in filteredSubmissions" :key="sub.id" class="hover:bg-slate-50/30 transition-all">
              <td class="py-4 px-6">
                <div>
                  <span class="font-extrabold text-slate-900 block">{{ sub.supermarket_name || sub.store_name }}</span>
                  <span class="text-xs text-indigo-600 font-bold block">{{ sub.md_name || sub.merchandiser_name }}</span>
                </div>
              </td>
              <td class="py-4 px-6 text-xs font-medium text-slate-500">
                <div class="flex flex-col gap-1">
                  <div class="flex items-center gap-1.5">
                    <Calendar class="w-3.5 h-3.5" />
                    <span>{{ sub.visit_time }}</span>
                  </div>
                </div>
              </td>
              <td class="py-4 px-6 text-center">
                <div class="flex flex-col items-center gap-1">
                  <span
                    v-if="sub.status === 'PENDING'"
                    class="px-2.5 py-1 text-[10px] font-extrabold uppercase rounded-full tracking-wider bg-amber-50 text-amber-600 border border-amber-200"
                  >
                    Pending Verifikasi
                  </span>
                  <span
                    v-else
                    class="px-2.5 py-1 text-[10px] font-extrabold uppercase rounded-full tracking-wider"
                    :class="sub.planogram_compliant ? 'bg-emerald-50 text-emerald-600 border border-emerald-100' : 'bg-rose-50 text-rose-600 border border-rose-100'"
                  >
                    {{ sub.planogram_compliant ? 'Planogram OK' : 'Mismatch' }}
                  </span>
                </div>
              </td>
              <td class="py-4 px-6">
                <div class="flex flex-col gap-1.5">
                  <div class="flex flex-wrap gap-1.5" v-if="sub.out_of_stock_skus.length">
                    <span
                      v-for="sku in sub.out_of_stock_skus"
                      :key="sku"
                      class="px-2 py-0.5 text-[10px] font-semibold text-rose-600 bg-rose-50 border border-rose-100 rounded-lg"
                    >
                      {{ sku }}
                    </span>
                  </div>
                  <span v-else class="text-xs text-emerald-600 font-bold flex items-center gap-1">
                    <CheckCircle2 class="w-3.5 h-3.5" />
                    Semua SKU Ready
                  </span>
                  <div class="flex gap-2 mt-1" v-if="sub.photo_url || sub.invoice_photo_url || sub.receipt_photo_url">
                    <button v-if="sub.photo_url" @click="openPreview(sub.photo_url)" class="text-[9px] text-indigo-500 hover:underline flex items-center gap-0.5"><Camera class="w-3 h-3"/> Foto</button>
                    <button v-if="sub.invoice_photo_url" @click="openPreview(sub.invoice_photo_url)" class="text-[9px] text-indigo-500 hover:underline flex items-center gap-0.5"><Image class="w-3 h-3"/> Nota</button>
                    <button v-if="sub.receipt_photo_url" @click="openPreview(sub.receipt_photo_url)" class="text-[9px] text-indigo-500 hover:underline flex items-center gap-0.5"><Image class="w-3 h-3"/> Tanda Terima</button>
                  </div>
                </div>
              </td>
              <td class="py-4 px-6 text-center">
                <div class="flex items-center justify-center gap-2">
                  <button
                    @click="openEditMDModal(sub)"
                    class="p-1.5 hover:bg-indigo-50 text-slate-400 hover:text-indigo-650 rounded-lg transition-all"
                    title="Edit Laporan"
                  >
                    <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
                    </svg>
                  </button>
                  <button
                    @click="deleteMD(sub.id)"
                    class="p-1.5 hover:bg-rose-50 text-slate-400 hover:text-rose-600 rounded-lg transition-all"
                    title="Hapus Laporan"
                  >
                    <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                    </svg>
                  </button>
                  <button
                    @click="createRepeatOrder(sub)"
                    class="p-1.5 hover:bg-emerald-50 text-slate-400 hover:text-emerald-600 rounded-lg transition-all"
                    title="Buat Repeat Order"
                  >
                    <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 3h2l.4 2M7 13h10l4-8H5.4M7 13L5.4 5M7 13l-2.293 2.293c-.63.63-.184 1.707.707 1.707H17m0 0a2 2 0 100 4 2 2 0 000-4zm-8 2a2 2 0 11-4 0 2 2 0 014 0z" />
                    </svg>
                  </button>
                </div>
              </td>
            </tr>
            <tr v-if="filteredSubmissions.length === 0">
              <td colspan="5" class="py-8 text-center text-slate-400">
                Tidak ada data Merchandising ditemukan.
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- MD Laporan Modal Form -->
    <div
      v-if="showMDModal"
      class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-900/60 backdrop-blur-xs overflow-y-auto"
    >
      <div class="bg-white rounded-[24px] w-full max-w-2xl shadow-2xl overflow-hidden p-6 relative my-8">
        <h3 class="text-lg font-black text-slate-900 mb-4">
          {{ isEditing ? 'Edit Laporan Audit MD' : 'Tambah Laporan Audit MD' }}
        </h3>

        <div class="space-y-4 text-xs font-semibold text-slate-650 max-h-[60vh] overflow-y-auto pr-1">
          <div class="grid grid-cols-2 gap-4">
            <div>
              <label class="block text-[10px] uppercase font-bold text-slate-400 mb-1">Ritel Modern (Key Account)</label>
              <select
                v-model="mdForm.key_account_id"
                class="w-full px-3 py-2 border border-slate-200 rounded-xl focus:ring-1 focus:ring-indigo-500 focus:outline-none bg-white"
              >
                <option v-for="ka in keyAccounts" :key="ka.id" :value="ka.id">
                  {{ ka.name }} ({{ ka.company_name }})
                </option>
              </select>
            </div>
            <div>
              <label class="block text-[10px] uppercase font-bold text-slate-400 mb-1">Nama Toko / Outlet</label>
              <input
                v-model="mdForm.store_name"
                type="text"
                placeholder="Contoh: Karsa HR Mart Menteng"
                class="w-full px-3 py-2 border border-slate-200 rounded-xl focus:ring-1 focus:ring-indigo-500 focus:outline-none"
              />
            </div>
          </div>

          <div class="grid grid-cols-2 gap-4">
            <div>
              <label class="block text-[10px] uppercase font-bold text-slate-400 mb-1">Nama Merchandiser (MD)</label>
              <input
                v-model="mdForm.merchandiser_name"
                type="text"
                placeholder="Contoh: Rian Anggara"
                class="w-full px-3 py-2 border border-slate-200 rounded-xl focus:ring-1 focus:ring-indigo-500 focus:outline-none"
              />
            </div>
            <div>
              <label class="block text-[10px] uppercase font-bold text-slate-400 mb-1">Waktu Kunjungan</label>
              <input
                v-model="mdForm.visit_time"
                type="datetime-local"
                class="w-full px-3 py-2 border border-slate-200 rounded-xl focus:ring-1 focus:ring-indigo-500 focus:outline-none"
              />
            </div>
          </div>

          <div class="grid grid-cols-2 gap-4">
            <div>
              <label class="block text-[10px] uppercase font-bold text-slate-400 mb-1">Kepatuhan Planogram</label>
              <select
                v-model="mdForm.planogram_compliant"
                class="w-full px-3 py-2 border border-slate-200 rounded-xl focus:ring-1 focus:ring-indigo-500 focus:outline-none bg-white"
              >
                <option :value="true">Patuh (Compliant)</option>
                <option :value="false">Tidak Patuh (Mismatch)</option>
              </select>
            </div>
            <div>
              <label class="block text-[10px] uppercase font-bold text-slate-400 mb-1">Foto Audit Shelf URL</label>
              <div class="flex gap-2">
                <input
                  type="file"
                  accept="image/*"
                  @change="uploadPhoto($event, 'photo_url')"
                  class="w-full text-xs text-slate-500 file:mr-2 file:py-1 file:px-3 file:rounded-lg file:border-0 file:text-xs file:font-semibold file:bg-indigo-50 file:text-indigo-700 hover:file:bg-indigo-100"
                />
              </div>
              <input
                v-model="mdForm.photo_url"
                type="text"
                placeholder="URL Gambar akan terisi otomatis..."
                class="w-full px-3 py-2 border border-slate-200 rounded-xl focus:ring-1 focus:ring-indigo-500 focus:outline-none mt-2"
                readonly
              />
            </div>
          </div>

          <div class="grid grid-cols-2 gap-4">
            <div>
              <label class="block text-[10px] uppercase font-bold text-slate-400 mb-1">Foto Nota / Invoice URL</label>
              <div class="flex gap-2">
                <input
                  type="file"
                  accept="image/*"
                  @change="uploadPhoto($event, 'invoice_photo_url')"
                  class="w-full text-xs text-slate-500 file:mr-2 file:py-1 file:px-3 file:rounded-lg file:border-0 file:text-xs file:font-semibold file:bg-indigo-50 file:text-indigo-700 hover:file:bg-indigo-100"
                />
              </div>
              <input
                v-model="mdForm.invoice_photo_url"
                type="text"
                placeholder="URL Gambar akan terisi otomatis..."
                class="w-full px-3 py-2 border border-slate-200 rounded-xl focus:ring-1 focus:ring-indigo-500 focus:outline-none mt-2"
                readonly
              />
            </div>
            <div>
              <label class="block text-[10px] uppercase font-bold text-slate-400 mb-1">Foto Tanda Terima MT URL</label>
              <div class="flex gap-2">
                <input
                  type="file"
                  accept="image/*"
                  @change="uploadPhoto($event, 'receipt_photo_url')"
                  class="w-full text-xs text-slate-500 file:mr-2 file:py-1 file:px-3 file:rounded-lg file:border-0 file:text-xs file:font-semibold file:bg-indigo-50 file:text-indigo-700 hover:file:bg-indigo-100"
                />
              </div>
              <input
                v-model="mdForm.receipt_photo_url"
                type="text"
                placeholder="URL Gambar akan terisi otomatis..."
                class="w-full px-3 py-2 border border-slate-200 rounded-xl focus:ring-1 focus:ring-indigo-500 focus:outline-none mt-2"
                readonly
              />
            </div>
          </div>
        </div>

        <div class="mt-6 flex items-center justify-end gap-3">
          <Button
            @click="showMDModal = false"
            variant="outline"
            class="px-4 py-2 border border-slate-200 rounded-xl text-xs text-slate-500 font-bold hover:bg-slate-50"
          >
            Batal
          </Button>
          <Button
            @click="saveMD"
            class="px-4 py-2 bg-indigo-600 hover:bg-indigo-700 text-white rounded-xl text-xs font-bold"
          >
            Simpan Audit MD
          </Button>
        </div>
      </div>
    </div>

    <!-- Image Preview Modal -->
    <div
      v-if="previewImageUrl"
      class="fixed inset-0 z-[100] flex items-center justify-center bg-slate-900/90 backdrop-blur-sm"
      @click.self="closePreview"
    >
      <div class="relative w-full h-full flex flex-col items-center justify-center p-4" @click.self="closePreview">
        <!-- Toolbar -->
        <div class="absolute top-6 right-6 flex items-center gap-4 bg-white/10 backdrop-blur-md px-4 py-2 rounded-full border border-white/20 shadow-2xl z-10">
          <button @click="zoomOut" class="text-white hover:text-indigo-300 transition-colors" title="Zoom Out">
            <ZoomOut class="w-6 h-6" />
          </button>
          <span class="text-white text-xs font-bold w-12 text-center">{{ Math.round(zoomLevel * 100) }}%</span>
          <button @click="zoomIn" class="text-white hover:text-indigo-300 transition-colors" title="Zoom In">
            <ZoomIn class="w-6 h-6" />
          </button>
          <div class="w-px h-4 bg-white/30 mx-2"></div>
          <button @click="closePreview" class="text-white hover:text-rose-400 transition-colors" title="Tutup">
            <X class="w-6 h-6" />
          </button>
        </div>

        <!-- Image Container -->
        <div 
          class="overflow-auto max-w-full max-h-full transition-transform duration-200 ease-out flex items-center justify-center pointer-events-none"
          :style="{ transform: `scale(${zoomLevel})` }"
        >
          <img 
            :src="previewImageUrl" 
            class="max-w-[90vw] max-h-[90vh] object-contain rounded-lg shadow-2xl ring-1 ring-white/20 pointer-events-auto"
            alt="Preview"
          />
        </div>
      </div>
    </div>
  </div>
</template>
