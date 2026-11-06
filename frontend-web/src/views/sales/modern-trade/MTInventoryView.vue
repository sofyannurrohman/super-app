<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import {
  Boxes,
  Search,
  Plus,
  ArrowDownLeft,
  ArrowUpRight,
  RefreshCw,
  AlertTriangle,
  History,
  Pencil,
  Trash2,
  Package,
  Layers,
  Sparkles,
  Calendar,
  X,
  UploadCloud,
  CheckCircle2,
  ArrowUpDown,
  Gift,
  Wrench,
  ShoppingBag,
  Info,
  Filter,
} from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import { toast } from 'vue-sonner'
import apiClient from '@/api/axios'
import { useMasterDataStore } from '@/stores/masterData'

const masterDataStore = useMasterDataStore()

interface MTInventoryItem {
  id: string
  branch_id?: string
  company_id?: string
  name: string
  category: 'PROMO_GIFT' | 'DEMO_TOOL' | 'PACKAGING' | 'GENERAL'
  quantity: number
  unit: string
  min_stock_alert: number
  photo_url?: string
  description?: string
  created_at: string
  updated_at: string
  branch?: {
    id: string
    name: string
  }
}

interface MTInventoryLog {
  id: string
  mt_inventory_id: string
  type: 'IN' | 'OUT' | 'ADJUSTMENT' | 'RETURN'
  quantity: number
  quantity_before: number
  quantity_after: number
  reference_no: string
  recipient_type: string
  recipient_name: string
  notes: string
  created_by: string
  created_at: string
  mt_inventory?: MTInventoryItem
}

// State
const activeTab = ref<'STOCKS' | 'LOGS'>('STOCKS')
const inventories = ref<MTInventoryItem[]>([])
const inventoryLogs = ref<MTInventoryLog[]>([])
const isLoading = ref(false)
const isLoadingLogs = ref(false)

// Filter & Search
const searchQuery = ref('')
const selectedCategory = ref<'ALL' | 'PROMO_GIFT' | 'DEMO_TOOL' | 'PACKAGING' | 'GENERAL'>('ALL')
const selectedStockStatus = ref<'ALL' | 'LOW' | 'AVAILABLE' | 'OUT_OF_STOCK'>('ALL')
const selectedLogType = ref<'ALL' | 'IN' | 'OUT' | 'ADJUSTMENT' | 'RETURN'>('ALL')

// Modal State: Create / Edit Master Item
const showItemModal = ref(false)
const isEditing = ref(false)
const isSubmitting = ref(false)
const itemForm = ref({
  id: '',
  name: '',
  category: 'PROMO_GIFT' as 'PROMO_GIFT' | 'DEMO_TOOL' | 'PACKAGING' | 'GENERAL',
  quantity: 0,
  unit: 'PCS',
  min_stock_alert: 10,
  branch_id: '',
  description: '',
  photo_url: '',
})

// Modal State: Mutate Stock (IN / OUT / ADJUSTMENT / RETURN)
const showMutateModal = ref(false)
const selectedItemForMutate = ref<MTInventoryItem | null>(null)
const mutateForm = ref({
  type: 'OUT' as 'IN' | 'OUT' | 'ADJUSTMENT' | 'RETURN',
  quantity: 1,
  reference_no: '',
  recipient_type: 'SPG',
  recipient_name: '',
  notes: '',
})

// Modal State: Single Item Log History
const showItemHistoryModal = ref(false)
const selectedItemHistory = ref<MTInventoryItem | null>(null)
const selectedItemLogs = ref<MTInventoryLog[]>([])
const isLoadingItemLogs = ref(false)

// Image Preview
const previewImageUrl = ref('')

// Helpers
const getCategoryLabel = (category: string) => {
  switch (category) {
    case 'PROMO_GIFT':
      return 'Gimmick / Gift Promo'
    case 'DEMO_TOOL':
      return 'Perlengkapan Demo / SPG'
    case 'PACKAGING':
      return 'Wadah / Packaging MT'
    default:
      return 'Umum / Lainnya'
  }
}

const getCategoryBadgeClass = (category: string) => {
  switch (category) {
    case 'PROMO_GIFT':
      return 'bg-pink-50 text-pink-700 border-pink-200'
    case 'DEMO_TOOL':
      return 'bg-amber-50 text-amber-700 border-amber-200'
    case 'PACKAGING':
      return 'bg-blue-50 text-blue-700 border-blue-200'
    default:
      return 'bg-slate-50 text-slate-700 border-slate-200'
  }
}

const getLogTypeBadgeClass = (type: string) => {
  switch (type) {
    case 'IN':
      return 'bg-emerald-50 text-emerald-700 border-emerald-200'
    case 'OUT':
      return 'bg-rose-50 text-rose-700 border-rose-200'
    case 'RETURN':
      return 'bg-indigo-50 text-indigo-700 border-indigo-200'
    case 'ADJUSTMENT':
      return 'bg-amber-50 text-amber-700 border-amber-200'
    default:
      return 'bg-slate-50 text-slate-700 border-slate-200'
  }
}

const getLogTypeLabel = (type: string) => {
  switch (type) {
    case 'IN':
      return 'Penerimaan (IN)'
    case 'OUT':
      return 'Pengeluaran (OUT)'
    case 'RETURN':
      return 'Pengembalian (RETURN)'
    case 'ADJUSTMENT':
      return 'Penyesuaian (OPNAME)'
    default:
      return type
  }
}

// Metrics
const totalItems = computed(() => inventories.value.length)
const totalQuantity = computed(() => inventories.value.reduce((acc, item) => acc + item.quantity, 0))
const lowStockCount = computed(() =>
  inventories.value.filter((item) => item.quantity > 0 && item.quantity <= item.min_stock_alert).length
)
const outOfStockCount = computed(() => inventories.value.filter((item) => item.quantity === 0).length)

// Filtered List
const filteredInventories = computed(() => {
  return inventories.value.filter((item) => {
    const matchesSearch =
      item.name.toLowerCase().includes(searchQuery.value.toLowerCase()) ||
      (item.description && item.description.toLowerCase().includes(searchQuery.value.toLowerCase()))

    const matchesCategory =
      selectedCategory.value === 'ALL' || item.category === selectedCategory.value

    let matchesStock = true
    if (selectedStockStatus.value === 'LOW') {
      matchesStock = item.quantity > 0 && item.quantity <= item.min_stock_alert
    } else if (selectedStockStatus.value === 'OUT_OF_STOCK') {
      matchesStock = item.quantity === 0
    } else if (selectedStockStatus.value === 'AVAILABLE') {
      matchesStock = item.quantity > item.min_stock_alert
    }

    return matchesSearch && matchesCategory && matchesStock
  })
})

// Filtered Logs
const filteredLogs = computed(() => {
  return inventoryLogs.value.filter((log) => {
    const matchesType = selectedLogType.value === 'ALL' || log.type === selectedLogType.value
    const matchesSearch =
      (log.mt_inventory?.name && log.mt_inventory.name.toLowerCase().includes(searchQuery.value.toLowerCase())) ||
      (log.reference_no && log.reference_no.toLowerCase().includes(searchQuery.value.toLowerCase())) ||
      (log.recipient_name && log.recipient_name.toLowerCase().includes(searchQuery.value.toLowerCase())) ||
      (log.notes && log.notes.toLowerCase().includes(searchQuery.value.toLowerCase()))

    return matchesType && matchesSearch
  })
})

// Fetch Data
const fetchInventories = async () => {
  isLoading.value = true
  try {
    const res = await apiClient.get('/modern-trade/inventories')
    inventories.value = res.data.data || []
  } catch (err: any) {
    toast.error('Gagal memuat data inventaris MT: ' + (err.response?.data?.error || err.message))
  } finally {
    isLoading.value = false
  }
}

const fetchLogs = async () => {
  isLoadingLogs.value = true
  try {
    const res = await apiClient.get('/modern-trade/inventory-logs')
    inventoryLogs.value = res.data.data || []
  } catch (err: any) {
    toast.error('Gagal memuat riwayat mutasi: ' + (err.response?.data?.error || err.message))
  } finally {
    isLoadingLogs.value = false
  }
}

onMounted(() => {
  fetchInventories()
  fetchLogs()
})

// Actions: Item Modal
const openCreateModal = () => {
  isEditing.value = false
  itemForm.value = {
    id: '',
    name: '',
    category: 'PROMO_GIFT',
    quantity: 0,
    unit: 'PCS',
    min_stock_alert: 10,
    branch_id: masterDataStore.selectedBranchId !== 'ALL_BRANCHES' ? masterDataStore.selectedBranchId : '',
    description: '',
    photo_url: '',
  }
  showItemModal.value = true
}

const openEditModal = (item: MTInventoryItem) => {
  isEditing.value = true
  itemForm.value = {
    id: item.id,
    name: item.name,
    category: item.category,
    quantity: item.quantity,
    unit: item.unit,
    min_stock_alert: item.min_stock_alert,
    branch_id: item.branch_id || '',
    description: item.description || '',
    photo_url: item.photo_url || '',
  }
  showItemModal.value = true
}

const submitItemForm = async () => {
  if (!itemForm.value.name.trim()) {
    toast.error('Nama barang wajib diisi')
    return
  }

  isSubmitting.value = true
  try {
    if (isEditing.value) {
      await apiClient.put(`/modern-trade/inventories/${itemForm.value.id}`, itemForm.value)
      toast.success('Item inventaris MT berhasil diperbarui')
    } else {
      await apiClient.post('/modern-trade/inventories', itemForm.value)
      toast.success('Item inventaris MT baru berhasil ditambahkan')
    }
    showItemModal.value = false
    fetchInventories()
    fetchLogs()
  } catch (err: any) {
    toast.error('Gagal menyimpan item: ' + (err.response?.data?.error || err.message))
  } finally {
    isSubmitting.value = false
  }
}

const handleDeleteItem = async (item: MTInventoryItem) => {
  if (!confirm(`Apakah Anda yakin ingin menghapus "${item.name}"? Riwayat mutasi item ini juga akan dihapus.`)) {
    return
  }

  try {
    await apiClient.delete(`/modern-trade/inventories/${item.id}`)
    toast.success(`"${item.name}" berhasil dihapus`)
    fetchInventories()
    fetchLogs()
  } catch (err: any) {
    toast.error('Gagal menghapus item: ' + (err.response?.data?.error || err.message))
  }
}

// Actions: Mutate Modal
const openMutateModal = (item: MTInventoryItem, defaultType: 'IN' | 'OUT' | 'ADJUSTMENT' | 'RETURN' = 'OUT') => {
  selectedItemForMutate.value = item
  mutateForm.value = {
    type: defaultType,
    quantity: defaultType === 'ADJUSTMENT' ? item.quantity : 1,
    reference_no: '',
    recipient_type: defaultType === 'OUT' ? 'SPG' : 'GUDANG',
    recipient_name: '',
    notes: '',
  }
  showMutateModal.value = true
}

const submitMutateForm = async () => {
  if (!selectedItemForMutate.value) return

  if (mutateForm.value.type !== 'ADJUSTMENT' && mutateForm.value.quantity <= 0) {
    toast.error('Jumlah mutasi harus lebih besar dari 0')
    return
  }

  if (
    mutateForm.value.type === 'OUT' &&
    mutateForm.value.quantity > selectedItemForMutate.value.quantity
  ) {
    toast.error(`Stok tidak mencukupi. Stok saat ini: ${selectedItemForMutate.value.quantity}`)
    return
  }

  isSubmitting.value = true
  try {
    await apiClient.post(`/modern-trade/inventories/${selectedItemForMutate.value.id}/mutate`, mutateForm.value)
    toast.success('Mutasi stok berhasil dicatat')
    showMutateModal.value = false
    fetchInventories()
    fetchLogs()
  } catch (err: any) {
    toast.error('Gagal melakukan mutasi stok: ' + (err.response?.data?.error || err.message))
  } finally {
    isSubmitting.value = false
  }
}

// Actions: Item History Modal
const openItemHistory = async (item: MTInventoryItem) => {
  selectedItemHistory.value = item
  showItemHistoryModal.value = true
  isLoadingItemLogs.value = true
  try {
    const res = await apiClient.get(`/modern-trade/inventories/${item.id}/logs`)
    selectedItemLogs.value = res.data.data || []
  } catch (err: any) {
    toast.error('Gagal mengambil kartu stok item: ' + (err.response?.data?.error || err.message))
  } finally {
    isLoadingItemLogs.value = false
  }
}

// Photo Upload
const isUploadingPhoto = ref(false)
const onPhotoUpload = async (e: Event) => {
  const target = e.target as HTMLInputElement
  const files = target?.files
  if (!files || files.length === 0) return

  const file = files[0]
  if (!file) return

  if (file.size > 5 * 1024 * 1024) {
    toast.error('Ukuran file maksimal 5MB')
    return
  }

  const formData = new FormData()
  formData.append('file', file)

  isUploadingPhoto.value = true
  try {
    const res = await apiClient.post('/modern-trade/inventories/upload-photo', formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
    })
    itemForm.value.photo_url = res.data.url
    toast.success('Foto barang berhasil diunggah')
  } catch (err: any) {
    toast.error('Gagal mengunggah foto: ' + (err.response?.data?.error || err.message))
  } finally {
    isUploadingPhoto.value = false
  }
}
</script>

<template>
  <div class="p-6 lg:p-8 space-y-8 bg-[#f8fafc] min-h-screen">
    <!-- Header -->
    <div
      class="flex flex-col md:flex-row justify-between items-start md:items-center gap-6 bg-white p-6 lg:p-8 rounded-[2rem] shadow-sm border border-slate-100"
    >
      <div class="flex items-center gap-5">
        <div
          class="w-14 h-14 rounded-[1.5rem] bg-gradient-to-br from-rose-600 via-rose-700 to-rose-900 flex items-center justify-center text-white shadow-xl shadow-rose-200"
        >
          <Boxes class="h-7 w-7" />
        </div>
        <div>
          <div class="flex items-center gap-3">
            <h1 class="text-2xl font-black tracking-tight text-slate-900">Stok & Inventaris MT</h1>
            <span class="px-2.5 py-0.5 rounded-full text-[11px] font-bold bg-rose-50 text-rose-700 border border-rose-100">
              Modern Trade Special
            </span>
          </div>
          <p class="text-sm text-slate-500 font-medium mt-0.5">
            Manajemen stok toples, solet, gimmick promosi, wadah, & perlengkapan SPG Modern Trade
          </p>
        </div>
      </div>

      <div class="flex items-center gap-3 w-full md:w-auto">
        <button
          @click="fetchInventories(); fetchLogs();"
          class="p-3 bg-slate-50 hover:bg-slate-100 text-slate-600 rounded-2xl transition-all border border-slate-200"
          title="Refresh Data"
        >
          <RefreshCw class="h-4 w-4" :class="{ 'animate-spin': isLoading || isLoadingLogs }" />
        </button>
        <button
          @click="openCreateModal"
          class="flex-1 md:flex-none flex items-center justify-center gap-2 bg-gradient-to-r from-rose-700 to-rose-900 text-white px-6 py-3 rounded-2xl font-bold text-sm shadow-lg shadow-rose-200 hover:shadow-xl hover:shadow-rose-300 hover:-translate-y-0.5 transition-all duration-200"
        >
          <Plus class="h-4 w-4" />
          Tambah Barang MT
        </button>
      </div>
    </div>

    <!-- Metrics Cards -->
    <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-5">
      <div class="bg-white p-6 rounded-[2rem] border border-slate-100 shadow-sm flex items-center gap-4">
        <div class="w-12 h-12 rounded-2xl bg-indigo-50 text-indigo-600 flex items-center justify-center">
          <Layers class="h-6 w-6" />
        </div>
        <div>
          <p class="text-xs font-bold uppercase tracking-wider text-slate-400">Total Jenis Item</p>
          <p class="text-2xl font-black text-slate-900 mt-0.5">{{ totalItems }} <span class="text-xs font-medium text-slate-400">SKU MT</span></p>
        </div>
      </div>

      <div class="bg-white p-6 rounded-[2rem] border border-slate-100 shadow-sm flex items-center gap-4">
        <div class="w-12 h-12 rounded-2xl bg-emerald-50 text-emerald-600 flex items-center justify-center">
          <Package class="h-6 w-6" />
        </div>
        <div>
          <p class="text-xs font-bold uppercase tracking-wider text-slate-400">Total Stok Fisik</p>
          <p class="text-2xl font-black text-slate-900 mt-0.5">{{ totalQuantity.toLocaleString('id-ID') }} <span class="text-xs font-medium text-slate-400">Pcs/Unit</span></p>
        </div>
      </div>

      <div class="bg-white p-6 rounded-[2rem] border border-slate-100 shadow-sm flex items-center gap-4">
        <div class="w-12 h-12 rounded-2xl bg-amber-50 text-amber-600 flex items-center justify-center">
          <AlertTriangle class="h-6 w-6" />
        </div>
        <div>
          <p class="text-xs font-bold uppercase tracking-wider text-slate-400">Stok Menipis</p>
          <p class="text-2xl font-black text-amber-600 mt-0.5">{{ lowStockCount }} <span class="text-xs font-medium text-slate-400">Item</span></p>
        </div>
      </div>

      <div class="bg-white p-6 rounded-[2rem] border border-slate-100 shadow-sm flex items-center gap-4">
        <div class="w-12 h-12 rounded-2xl bg-rose-50 text-rose-600 flex items-center justify-center">
          <Sparkles class="h-6 w-6" />
        </div>
        <div>
          <p class="text-xs font-bold uppercase tracking-wider text-slate-400">Stok Habis</p>
          <p class="text-2xl font-black text-rose-600 mt-0.5">{{ outOfStockCount }} <span class="text-xs font-medium text-slate-400">Item</span></p>
        </div>
      </div>
    </div>

    <!-- Navigation Tabs -->
    <div class="flex border-b border-slate-200">
      <button
        @click="activeTab = 'STOCKS'"
        :class="[
          'pb-4 px-6 text-sm font-bold transition-all relative',
          activeTab === 'STOCKS'
            ? 'text-rose-700 border-b-2 border-rose-700'
            : 'text-slate-400 hover:text-slate-700'
        ]"
      >
        <div class="flex items-center gap-2">
          <Boxes class="h-4 w-4" />
          Katalog & Stok Barang MT
          <span class="px-2 py-0.5 rounded-full text-xs bg-rose-50 text-rose-700 font-black">
            {{ filteredInventories.length }}
          </span>
        </div>
      </button>

      <button
        @click="activeTab = 'LOGS'"
        :class="[
          'pb-4 px-6 text-sm font-bold transition-all relative',
          activeTab === 'LOGS'
            ? 'text-rose-700 border-b-2 border-rose-700'
            : 'text-slate-400 hover:text-slate-700'
        ]"
      >
        <div class="flex items-center gap-2">
          <History class="h-4 w-4" />
          Riwayat Mutasi & Audit Stok
          <span class="px-2 py-0.5 rounded-full text-xs bg-slate-100 text-slate-600 font-black">
            {{ inventoryLogs.length }}
          </span>
        </div>
      </button>
    </div>

    <!-- TAB 1: STOCKS VIEW -->
    <div v-if="activeTab === 'STOCKS'" class="space-y-6">
      <!-- Search & Filters -->
      <div class="flex flex-col md:flex-row gap-4 justify-between items-stretch md:items-center">
        <div class="relative flex-1">
          <Search class="absolute left-4 top-1/2 -translate-y-1/2 h-4 w-4 text-slate-400" />
          <input
            v-model="searchQuery"
            type="text"
            placeholder="Cari toples, solet, gimmick promo MT..."
            class="w-full pl-11 pr-4 py-3 bg-white border border-slate-200 rounded-2xl text-sm font-medium text-slate-700 placeholder-slate-400 outline-none focus:ring-2 focus:ring-rose-700/20 focus:border-rose-400 transition-all shadow-sm"
          />
        </div>

        <div class="flex flex-wrap items-center gap-3">
          <!-- Filter Kategori -->
          <div class="flex items-center gap-2 bg-white px-3 py-2 border border-slate-200 rounded-2xl shadow-sm">
            <Filter class="h-4 w-4 text-slate-400" />
            <select
              v-model="selectedCategory"
              class="text-xs font-bold text-slate-700 bg-transparent outline-none cursor-pointer"
            >
              <option value="ALL">Semua Kategori</option>
              <option value="PROMO_GIFT">Gimmick / Gift Promo</option>
              <option value="DEMO_TOOL">Perlengkapan Demo / SPG</option>
              <option value="PACKAGING">Wadah / Packaging MT</option>
              <option value="GENERAL">Umum / Lainnya</option>
            </select>
          </div>

          <!-- Filter Kondisi Stok -->
          <div class="flex items-center gap-2 bg-white px-3 py-2 border border-slate-200 rounded-2xl shadow-sm">
            <select
              v-model="selectedStockStatus"
              class="text-xs font-bold text-slate-700 bg-transparent outline-none cursor-pointer"
            >
              <option value="ALL">Semua Status Stok</option>
              <option value="AVAILABLE">Stok Aman</option>
              <option value="LOW">Stok Menipis (Warning)</option>
              <option value="OUT_OF_STOCK">Stok Habis</option>
            </select>
          </div>
        </div>
      </div>

      <!-- Table / Grid -->
      <div class="bg-white rounded-[2rem] border border-slate-100 shadow-sm overflow-hidden">
        <div v-if="isLoading" class="flex items-center justify-center py-24">
          <div class="w-8 h-8 border-4 border-rose-700 border-t-transparent rounded-full animate-spin"></div>
        </div>

        <div v-else-if="filteredInventories.length === 0" class="flex flex-col items-center justify-center py-20 text-slate-400">
          <Boxes class="h-16 w-16 mb-4 opacity-20 text-rose-800" />
          <p class="font-bold text-lg text-slate-700">Tidak ada data inventaris MT</p>
          <p class="text-sm text-slate-400">Tambahkan barang seperti toples atau solet menggunakan tombol Tambah Barang MT</p>
        </div>

        <div v-else class="overflow-x-auto">
          <table class="w-full text-left">
            <thead>
              <tr class="border-b border-slate-100 bg-slate-50/50">
                <th class="px-6 py-4 text-[11px] font-black uppercase tracking-widest text-slate-400">Barang & Kategori</th>
                <th class="px-6 py-4 text-[11px] font-black uppercase tracking-widest text-slate-400">Stok Gudang</th>
                <th class="px-6 py-4 text-[11px] font-black uppercase tracking-widest text-slate-400">Batas Min. (Alert)</th>
                <th class="px-6 py-4 text-[11px] font-black uppercase tracking-widest text-slate-400">Cabang / Lokasi</th>
                <th class="px-6 py-4 text-[11px] font-black uppercase tracking-widest text-slate-400">Keterangan</th>
                <th class="px-6 py-4 text-right text-[11px] font-black uppercase tracking-widest text-slate-400">Aksi Cepat</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-100">
              <tr
                v-for="item in filteredInventories"
                :key="item.id"
                class="hover:bg-slate-50/60 transition-colors group"
              >
                <!-- Nama & Foto -->
                <td class="px-6 py-4">
                  <div class="flex items-center gap-3">
                    <div
                      class="w-12 h-12 rounded-xl bg-slate-100 border border-slate-200 overflow-hidden flex items-center justify-center shrink-0 cursor-pointer"
                      @click="item.photo_url ? previewImageUrl = item.photo_url : null"
                    >
                      <img
                        v-if="item.photo_url"
                        :src="item.photo_url"
                        alt="Photo"
                        class="w-full h-full object-cover hover:scale-110 transition-transform"
                      />
                      <Boxes v-else class="h-6 w-6 text-slate-400" />
                    </div>
                    <div>
                      <p class="font-bold text-sm text-slate-900 leading-snug">{{ item.name }}</p>
                      <div class="flex items-center gap-2 mt-1">
                        <span
                          class="px-2 py-0.5 rounded-md text-[10px] font-extrabold border"
                          :class="getCategoryBadgeClass(item.category)"
                        >
                          {{ getCategoryLabel(item.category) }}
                        </span>
                      </div>
                    </div>
                  </div>
                </td>

                <!-- Stok -->
                <td class="px-6 py-4">
                  <div class="flex items-center gap-2">
                    <span
                      class="inline-flex items-center px-3 py-1 rounded-full text-xs font-black border"
                      :class="{
                        'bg-emerald-50 text-emerald-700 border-emerald-200': item.quantity > item.min_stock_alert,
                        'bg-amber-50 text-amber-700 border-amber-200': item.quantity > 0 && item.quantity <= item.min_stock_alert,
                        'bg-rose-50 text-rose-700 border-rose-200': item.quantity === 0
                      }"
                    >
                      {{ item.quantity.toLocaleString('id-ID') }} {{ item.unit }}
                    </span>
                    <span
                      v-if="item.quantity === 0"
                      class="text-[10px] font-bold text-rose-600 bg-rose-50 px-1.5 py-0.5 rounded border border-rose-100"
                    >
                      Habis
                    </span>
                    <span
                      v-else-if="item.quantity <= item.min_stock_alert"
                      class="text-[10px] font-bold text-amber-600 bg-amber-50 px-1.5 py-0.5 rounded border border-amber-100"
                    >
                      Menipis
                    </span>
                  </div>
                </td>

                <!-- Min Alert -->
                <td class="px-6 py-4">
                  <p class="text-xs font-bold text-slate-600">{{ item.min_stock_alert }} {{ item.unit }}</p>
                </td>

                <!-- Cabang -->
                <td class="px-6 py-4">
                  <p class="text-xs font-medium text-slate-700">{{ item.branch?.name || 'Semua Cabang / Gudang Pusat' }}</p>
                </td>

                <!-- Deskripsi -->
                <td class="px-6 py-4">
                  <p class="text-xs text-slate-500 max-w-[200px] truncate">{{ item.description || '-' }}</p>
                </td>

                <!-- Aksi -->
                <td class="px-6 py-4 text-right">
                  <div class="flex items-center justify-end gap-1.5">
                    <!-- Mutasi Masuk (IN) -->
                    <button
                      @click="openMutateModal(item, 'IN')"
                      class="p-2 text-emerald-600 hover:bg-emerald-50 rounded-xl transition-all font-bold text-xs flex items-center gap-1 border border-emerald-100"
                      title="Barang Masuk (Penerimaan)"
                    >
                      <ArrowDownLeft class="h-3.5 w-3.5" />
                      IN
                    </button>

                    <!-- Mutasi Keluar (OUT) -->
                    <button
                      @click="openMutateModal(item, 'OUT')"
                      class="p-2 text-rose-700 hover:bg-rose-50 rounded-xl transition-all font-bold text-xs flex items-center gap-1 border border-rose-100"
                      title="Alokasi / Pengeluaran (SPG/Outlet)"
                    >
                      <ArrowUpRight class="h-3.5 w-3.5" />
                      OUT
                    </button>

                    <!-- Mutasi Koreksi / Opname -->
                    <button
                      @click="openMutateModal(item, 'ADJUSTMENT')"
                      class="p-2 text-amber-600 hover:bg-amber-50 rounded-xl transition-all"
                      title="Penyesuaian Stok / Opname"
                    >
                      <ArrowUpDown class="h-3.5 w-3.5" />
                    </button>

                    <!-- Riwayat Kartu Stok -->
                    <button
                      @click="openItemHistory(item)"
                      class="p-2 text-indigo-600 hover:bg-indigo-50 rounded-xl transition-all"
                      title="Kartu Stok & Riwayat"
                    >
                      <History class="h-3.5 w-3.5" />
                    </button>

                    <!-- Edit -->
                    <button
                      @click="openEditModal(item)"
                      class="p-2 text-slate-400 hover:text-slate-700 hover:bg-slate-100 rounded-xl transition-all"
                      title="Edit Item"
                    >
                      <Pencil class="h-3.5 w-3.5" />
                    </button>

                    <!-- Hapus -->
                    <button
                      @click="handleDeleteItem(item)"
                      class="p-2 text-slate-400 hover:text-rose-600 hover:bg-rose-50 rounded-xl transition-all"
                      title="Hapus Item"
                    >
                      <Trash2 class="h-3.5 w-3.5" />
                    </button>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>

    <!-- TAB 2: AUDIT & MUTATION LOGS VIEW -->
    <div v-if="activeTab === 'LOGS'" class="space-y-6">
      <div class="flex flex-col md:flex-row gap-4 justify-between items-stretch md:items-center">
        <div class="relative flex-1">
          <Search class="absolute left-4 top-1/2 -translate-y-1/2 h-4 w-4 text-slate-400" />
          <input
            v-model="searchQuery"
            type="text"
            placeholder="Cari referensi, nama barang, penerima, atau keterangan..."
            class="w-full pl-11 pr-4 py-3 bg-white border border-slate-200 rounded-2xl text-sm font-medium text-slate-700 placeholder-slate-400 outline-none focus:ring-2 focus:ring-rose-700/20 focus:border-rose-400 transition-all shadow-sm"
          />
        </div>

        <div class="flex items-center gap-3">
          <div class="flex items-center gap-2 bg-white px-3 py-2 border border-slate-200 rounded-2xl shadow-sm">
            <Filter class="h-4 w-4 text-slate-400" />
            <select
              v-model="selectedLogType"
              class="text-xs font-bold text-slate-700 bg-transparent outline-none cursor-pointer"
            >
              <option value="ALL">Semua Tipe Mutasi</option>
              <option value="IN">Penerimaan (IN)</option>
              <option value="OUT">Pengeluaran (OUT)</option>
              <option value="RETURN">Pengembalian (RETURN)</option>
              <option value="ADJUSTMENT">Penyesuaian (OPNAME)</option>
            </select>
          </div>
        </div>
      </div>

      <!-- Table Logs -->
      <div class="bg-white rounded-[2rem] border border-slate-100 shadow-sm overflow-hidden">
        <div v-if="isLoadingLogs" class="flex items-center justify-center py-24">
          <div class="w-8 h-8 border-4 border-rose-700 border-t-transparent rounded-full animate-spin"></div>
        </div>

        <div v-else-if="filteredLogs.length === 0" class="flex flex-col items-center justify-center py-20 text-slate-400">
          <History class="h-16 w-16 mb-4 opacity-20 text-rose-800" />
          <p class="font-bold text-lg text-slate-700">Belum ada riwayat mutasi stok</p>
          <p class="text-sm text-slate-400">Setiap pergerakan stok barang MT akan tercatat otomatis di sini</p>
        </div>

        <div v-else class="overflow-x-auto">
          <table class="w-full text-left">
            <thead>
              <tr class="border-b border-slate-100 bg-slate-50/50">
                <th class="px-6 py-4 text-[11px] font-black uppercase tracking-widest text-slate-400">Waktu & PIC</th>
                <th class="px-6 py-4 text-[11px] font-black uppercase tracking-widest text-slate-400">Barang</th>
                <th class="px-6 py-4 text-[11px] font-black uppercase tracking-widest text-slate-400">Tipe Mutasi</th>
                <th class="px-6 py-4 text-[11px] font-black uppercase tracking-widest text-slate-400">Perubahan Qty</th>
                <th class="px-6 py-4 text-[11px] font-black uppercase tracking-widest text-slate-400">Sisa Akhir</th>
                <th class="px-6 py-4 text-[11px] font-black uppercase tracking-widest text-slate-400">Peruntukan / Penerima</th>
                <th class="px-6 py-4 text-[11px] font-black uppercase tracking-widest text-slate-400">Keterangan & Ref</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-100">
              <tr v-for="log in filteredLogs" :key="log.id" class="hover:bg-slate-50/60 transition-colors">
                <td class="px-6 py-4">
                  <p class="font-bold text-xs text-slate-900">{{ new Date(log.created_at).toLocaleString('id-ID') }}</p>
                  <p class="text-[11px] text-slate-400 font-medium">Oleh: {{ log.created_by || 'Admin' }}</p>
                </td>

                <td class="px-6 py-4">
                  <p class="font-bold text-sm text-slate-800">{{ log.mt_inventory?.name || 'Item MT' }}</p>
                  <span class="text-[10px] font-semibold text-slate-400">
                    {{ log.mt_inventory?.unit || 'PCS' }}
                  </span>
                </td>

                <td class="px-6 py-4">
                  <span
                    class="px-2.5 py-1 rounded-full text-[10px] font-black border"
                    :class="getLogTypeBadgeClass(log.type)"
                  >
                    {{ getLogTypeLabel(log.type) }}
                  </span>
                </td>

                <td class="px-6 py-4">
                  <div class="flex items-center gap-1 font-bold text-xs">
                    <span v-if="log.type === 'IN' || log.type === 'RETURN'" class="text-emerald-600 font-black">
                      +{{ log.quantity }}
                    </span>
                    <span v-else-if="log.type === 'OUT'" class="text-rose-600 font-black">
                      -{{ log.quantity }}
                    </span>
                    <span v-else class="text-amber-600 font-black">
                      {{ log.quantity_before }} &rarr; {{ log.quantity_after }}
                    </span>
                    <span class="text-[11px] font-medium text-slate-400">{{ log.mt_inventory?.unit || 'PCS' }}</span>
                  </div>
                </td>

                <td class="px-6 py-4">
                  <span class="px-2.5 py-0.5 rounded-md text-xs font-black bg-slate-100 text-slate-700">
                    {{ log.quantity_after }}
                  </span>
                </td>

                <td class="px-6 py-4">
                  <p class="text-xs font-bold text-slate-800">{{ log.recipient_name || '-' }}</p>
                  <span v-if="log.recipient_type" class="text-[10px] font-bold text-slate-400">
                    ({{ log.recipient_type }})
                  </span>
                </td>

                <td class="px-6 py-4">
                  <p class="text-xs font-medium text-slate-600">{{ log.notes || '-' }}</p>
                  <p v-if="log.reference_no" class="text-[10px] font-mono text-slate-400">Ref: {{ log.reference_no }}</p>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>

    <!-- MODAL 1: TAMBAH / EDIT MASTER ITEM -->
    <Teleport to="body">
      <div v-if="showItemModal" class="fixed inset-0 z-50 flex items-center justify-center p-4">
        <div class="absolute inset-0 bg-black/40 backdrop-blur-sm" @click="showItemModal = false" />
        <div class="relative bg-white rounded-[2rem] shadow-2xl w-full max-w-lg overflow-hidden flex flex-col max-h-[90vh]">
          <div class="px-8 pt-8 pb-5 border-b border-slate-100 flex items-center justify-between">
            <div class="flex items-center gap-3">
              <div class="w-10 h-10 rounded-xl bg-rose-50 text-rose-700 flex items-center justify-center font-bold">
                <Boxes class="h-5 w-5" />
              </div>
              <h2 class="text-xl font-black text-slate-900">
                {{ isEditing ? 'Edit Barang MT' : 'Tambah Barang MT' }}
              </h2>
            </div>
            <button @click="showItemModal = false" class="p-2 text-slate-400 hover:text-slate-700 hover:bg-slate-100 rounded-xl transition-all">
              <X class="h-5 w-5" />
            </button>
          </div>

          <form @submit.prevent="submitItemForm" class="p-8 space-y-4 overflow-y-auto flex-1">
            <!-- Nama Item -->
            <div>
              <label class="block text-xs font-black uppercase tracking-widest text-slate-500 mb-1.5">Nama Barang *</label>
              <input
                v-model="itemForm.name"
                type="text"
                required
                placeholder="Contoh: Toples Kaca Eksklusif Karsa HR, Solet Silikon Merah"
                class="w-full px-4 py-3 bg-slate-50 border border-slate-200 rounded-xl text-sm font-medium focus:ring-2 focus:ring-rose-700/20 focus:border-rose-700 focus:bg-white transition-all outline-none"
              />
            </div>

            <!-- Kategori & Satuan -->
            <div class="grid grid-cols-2 gap-4">
              <div>
                <label class="block text-xs font-black uppercase tracking-widest text-slate-500 mb-1.5">Kategori</label>
                <select
                  v-model="itemForm.category"
                  class="w-full px-4 py-3 bg-slate-50 border border-slate-200 rounded-xl text-sm font-bold text-slate-700 focus:ring-2 focus:ring-rose-700/20 focus:border-rose-700 focus:bg-white transition-all outline-none"
                >
                  <option value="PROMO_GIFT">Gimmick / Gift Promo</option>
                  <option value="DEMO_TOOL">Perlengkapan Demo / SPG</option>
                  <option value="PACKAGING">Wadah / Packaging MT</option>
                  <option value="GENERAL">Umum / Lainnya</option>
                </select>
              </div>

              <div>
                <label class="block text-xs font-black uppercase tracking-widest text-slate-500 mb-1.5">Satuan Unit *</label>
                <input
                  v-model="itemForm.unit"
                  type="text"
                  required
                  placeholder="PCS, SET, DUS"
                  class="w-full px-4 py-3 bg-slate-50 border border-slate-200 rounded-xl text-sm font-bold uppercase focus:ring-2 focus:ring-rose-700/20 focus:border-rose-700 focus:bg-white transition-all outline-none"
                />
              </div>
            </div>

            <!-- Stok Awal & Min Alert -->
            <div class="grid grid-cols-2 gap-4">
              <div v-if="!isEditing">
                <label class="block text-xs font-black uppercase tracking-widest text-slate-500 mb-1.5">Stok Awal</label>
                <input
                  v-model.number="itemForm.quantity"
                  type="number"
                  min="0"
                  required
                  class="w-full px-4 py-3 bg-slate-50 border border-slate-200 rounded-xl text-sm font-bold focus:ring-2 focus:ring-rose-700/20 focus:border-rose-700 focus:bg-white transition-all outline-none"
                />
              </div>

              <div :class="{ 'col-span-2': isEditing }">
                <label class="block text-xs font-black uppercase tracking-widest text-slate-500 mb-1.5">Batas Minimum (Alert)</label>
                <input
                  v-model.number="itemForm.min_stock_alert"
                  type="number"
                  min="1"
                  required
                  placeholder="10"
                  class="w-full px-4 py-3 bg-slate-50 border border-slate-200 rounded-xl text-sm font-bold focus:ring-2 focus:ring-rose-700/20 focus:border-rose-700 focus:bg-white transition-all outline-none"
                />
              </div>
            </div>

            <!-- Cabang (Opsional) -->
            <div v-if="!masterDataStore.selectedBranchId || masterDataStore.selectedBranchId === 'ALL_BRANCHES'">
              <label class="block text-xs font-black uppercase tracking-widest text-slate-500 mb-1.5">Alokasi Cabang (Opsional)</label>
              <select
                v-model="itemForm.branch_id"
                class="w-full px-4 py-3 bg-slate-50 border border-slate-200 rounded-xl text-sm font-medium focus:ring-2 focus:ring-rose-700/20 focus:border-rose-700 focus:bg-white transition-all outline-none"
              >
                <option value="">Semua Cabang / Gudang Pusat</option>
                <option v-for="b in masterDataStore.branches" :key="b.id" :value="b.id">
                  {{ b.name }}
                </option>
              </select>
            </div>

            <!-- Foto Barang -->
            <div>
              <label class="block text-xs font-black uppercase tracking-widest text-slate-500 mb-1.5">Foto Barang</label>
              <div class="flex items-center gap-3">
                <div
                  v-if="itemForm.photo_url"
                  class="w-14 h-14 rounded-xl border border-slate-200 overflow-hidden shrink-0"
                >
                  <img :src="itemForm.photo_url" class="w-full h-full object-cover" />
                </div>
                <label
                  class="flex-1 flex items-center justify-center gap-2 px-4 py-3 border-2 border-dashed border-slate-200 hover:border-rose-700 rounded-xl cursor-pointer transition-colors bg-slate-50 text-slate-600 text-xs font-bold"
                >
                  <UploadCloud class="h-4 w-4 text-rose-700" />
                  <span>{{ isUploadingPhoto ? 'Mengunggah...' : 'Unggah Foto Barang' }}</span>
                  <input type="file" accept="image/*" class="hidden" @change="onPhotoUpload" :disabled="isUploadingPhoto" />
                </label>
              </div>
            </div>

            <!-- Deskripsi -->
            <div>
              <label class="block text-xs font-black uppercase tracking-widest text-slate-500 mb-1.5">Deskripsi / Catatan Tambahan</label>
              <textarea
                v-model="itemForm.description"
                rows="2"
                placeholder="Contoh: Digunakan untuk promo bundling toples di Superindo & Hypermart..."
                class="w-full px-4 py-3 bg-slate-50 border border-slate-200 rounded-xl text-sm font-medium focus:ring-2 focus:ring-rose-700/20 focus:border-rose-700 focus:bg-white transition-all outline-none resize-none"
              ></textarea>
            </div>

            <div class="pt-4 flex gap-3">
              <button
                type="button"
                @click="showItemModal = false"
                class="flex-1 px-4 py-3 text-sm font-bold text-slate-700 bg-slate-100 rounded-xl hover:bg-slate-200 transition-colors"
              >
                Batal
              </button>
              <button
                type="submit"
                :disabled="isSubmitting"
                class="flex-1 px-4 py-3 text-sm font-bold text-white bg-rose-700 rounded-xl hover:bg-rose-800 transition-colors disabled:opacity-50"
              >
                {{ isSubmitting ? 'Menyimpan...' : 'Simpan Barang' }}
              </button>
            </div>
          </form>
        </div>
      </div>
    </Teleport>

    <!-- MODAL 2: FORM MUTASI STOK (IN / OUT / ADJUSTMENT) -->
    <Teleport to="body">
      <div v-if="showMutateModal && selectedItemForMutate" class="fixed inset-0 z-50 flex items-center justify-center p-4">
        <div class="absolute inset-0 bg-black/40 backdrop-blur-sm" @click="showMutateModal = false" />
        <div class="relative bg-white rounded-[2rem] shadow-2xl w-full max-w-md overflow-hidden flex flex-col max-h-[90vh]">
          <div class="px-8 pt-8 pb-5 border-b border-slate-100 flex items-center justify-between">
            <div>
              <div class="flex items-center gap-2">
                <span
                  class="px-2 py-0.5 rounded text-[10px] font-black border"
                  :class="getLogTypeBadgeClass(mutateForm.type)"
                >
                  {{ getLogTypeLabel(mutateForm.type) }}
                </span>
                <h2 class="text-lg font-black text-slate-900">Mutasi Stok MT</h2>
              </div>
              <p class="text-xs text-slate-500 font-bold mt-1">
                {{ selectedItemForMutate.name }} (Stok Saat Ini: {{ selectedItemForMutate.quantity }} {{ selectedItemForMutate.unit }})
              </p>
            </div>
            <button @click="showMutateModal = false" class="p-2 text-slate-400 hover:text-slate-700 hover:bg-slate-100 rounded-xl transition-all">
              <X class="h-5 w-5" />
            </button>
          </div>

          <form @submit.prevent="submitMutateForm" class="p-8 space-y-4 overflow-y-auto flex-1">
            <!-- Pilihan Tipe Mutasi -->
            <div>
              <label class="block text-xs font-black uppercase tracking-widest text-slate-500 mb-1.5">Tipe Mutasi</label>
              <div class="grid grid-cols-3 gap-2">
                <button
                  type="button"
                  @click="mutateForm.type = 'OUT'; mutateForm.recipient_type = 'SPG'"
                  :class="[
                    'py-2 px-3 text-xs font-black rounded-xl border transition-all text-center',
                    mutateForm.type === 'OUT'
                      ? 'bg-rose-50 text-rose-700 border-rose-300 ring-2 ring-rose-700/20'
                      : 'bg-slate-50 text-slate-600 border-slate-200'
                  ]"
                >
                  Keluar (OUT)
                </button>
                <button
                  type="button"
                  @click="mutateForm.type = 'IN'; mutateForm.recipient_type = 'GUDANG'"
                  :class="[
                    'py-2 px-3 text-xs font-black rounded-xl border transition-all text-center',
                    mutateForm.type === 'IN'
                      ? 'bg-emerald-50 text-emerald-700 border-emerald-300 ring-2 ring-emerald-700/20'
                      : 'bg-slate-50 text-slate-600 border-slate-200'
                  ]"
                >
                  Masuk (IN)
                </button>
                <button
                  type="button"
                  @click="mutateForm.type = 'ADJUSTMENT'; mutateForm.quantity = selectedItemForMutate.quantity"
                  :class="[
                    'py-2 px-3 text-xs font-black rounded-xl border transition-all text-center',
                    mutateForm.type === 'ADJUSTMENT'
                      ? 'bg-amber-50 text-amber-700 border-amber-300 ring-2 ring-amber-700/20'
                      : 'bg-slate-50 text-slate-600 border-slate-200'
                  ]"
                >
                  Koreksi/Opname
                </button>
              </div>
            </div>

            <!-- Jumlah / Quantity -->
            <div>
              <label class="block text-xs font-black uppercase tracking-widest text-slate-500 mb-1.5">
                {{ mutateForm.type === 'ADJUSTMENT' ? 'Stok Fisik Sebenarnya (Target Qty)' : 'Jumlah Mutasi (' + selectedItemForMutate.unit + ') *' }}
              </label>
              <input
                v-model.number="mutateForm.quantity"
                type="number"
                min="0"
                required
                class="w-full px-4 py-3 bg-slate-50 border border-slate-200 rounded-xl text-base font-black focus:ring-2 focus:ring-rose-700/20 focus:border-rose-700 focus:bg-white transition-all outline-none"
              />
              <p v-if="mutateForm.type === 'OUT'" class="text-[11px] text-slate-400 mt-1">
                Maksimal pengeluaran: <strong class="text-slate-700">{{ selectedItemForMutate.quantity }} {{ selectedItemForMutate.unit }}</strong>
              </p>
              <p v-if="mutateForm.type === 'ADJUSTMENT'" class="text-[11px] text-amber-600 font-bold mt-1">
                Selisih: {{ mutateForm.quantity - selectedItemForMutate.quantity }} {{ selectedItemForMutate.unit }}
              </p>
            </div>

            <!-- Peruntukan / Penerima (Jika OUT atau IN) -->
            <div v-if="mutateForm.type === 'OUT' || mutateForm.type === 'RETURN'">
              <label class="block text-xs font-black uppercase tracking-widest text-slate-500 mb-1.5">Peruntukan / Penerima</label>
              <div class="grid grid-cols-2 gap-3 mb-2">
                <select
                  v-model="mutateForm.recipient_type"
                  class="w-full px-3 py-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-bold text-slate-700 focus:ring-2 focus:ring-rose-700/20 focus:border-rose-700 focus:bg-white transition-all outline-none"
                >
                  <option value="SPG">SPG MT</option>
                  <option value="OUTLET_MT">Outlet / Supermarket MT</option>
                  <option value="EVENT">Event / Pameran MT</option>
                  <option value="BUNDLING_PO">Bundling PO Retailer</option>
                  <option value="LAINNYA">Lainnya</option>
                </select>

                <input
                  v-model="mutateForm.recipient_name"
                  type="text"
                  placeholder="Nama SPG / Nama Toko"
                  class="w-full px-3 py-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-medium focus:ring-2 focus:ring-rose-700/20 focus:border-rose-700 focus:bg-white transition-all outline-none"
                />
              </div>
            </div>

            <!-- No Referensi -->
            <div>
              <label class="block text-xs font-black uppercase tracking-widest text-slate-500 mb-1.5">Nomor Referensi (Surat Jalan / PO / Nota)</label>
              <input
                v-model="mutateForm.reference_no"
                type="text"
                placeholder="Contoh: SJ-MT-2026/09/001, PO-INDOMARCO-09"
                class="w-full px-4 py-3 bg-slate-50 border border-slate-200 rounded-xl text-sm font-medium focus:ring-2 focus:ring-rose-700/20 focus:border-rose-700 focus:bg-white transition-all outline-none"
              />
            </div>

            <!-- Keterangan / Alasan -->
            <div>
              <label class="block text-xs font-black uppercase tracking-widest text-slate-500 mb-1.5">Keterangan / Alasan</label>
              <textarea
                v-model="mutateForm.notes"
                rows="2"
                placeholder="Contoh: Alokasi toples promo beli 2 gratis 1 di Superindo Cabang Dago..."
                class="w-full px-4 py-3 bg-slate-50 border border-slate-200 rounded-xl text-sm font-medium focus:ring-2 focus:ring-rose-700/20 focus:border-rose-700 focus:bg-white transition-all outline-none resize-none"
              ></textarea>
            </div>

            <div class="pt-3 flex gap-3">
              <button
                type="button"
                @click="showMutateModal = false"
                class="flex-1 px-4 py-3 text-sm font-bold text-slate-700 bg-slate-100 rounded-xl hover:bg-slate-200 transition-colors"
              >
                Batal
              </button>
              <button
                type="submit"
                :disabled="isSubmitting"
                class="flex-1 px-4 py-3 text-sm font-bold text-white bg-rose-700 rounded-xl hover:bg-rose-800 transition-colors disabled:opacity-50"
              >
                {{ isSubmitting ? 'Memproses...' : 'Simpan Mutasi' }}
              </button>
            </div>
          </form>
        </div>
      </div>
    </Teleport>

    <!-- MODAL 3: KARTU STOK ITEM -->
    <Teleport to="body">
      <div v-if="showItemHistoryModal && selectedItemHistory" class="fixed inset-0 z-50 flex items-center justify-center p-4">
        <div class="absolute inset-0 bg-black/40 backdrop-blur-sm" @click="showItemHistoryModal = false" />
        <div class="relative bg-white rounded-[2rem] shadow-2xl w-full max-w-2xl overflow-hidden flex flex-col max-h-[85vh]">
          <div class="px-8 pt-8 pb-5 border-b border-slate-100 flex items-center justify-between">
            <div class="flex items-center gap-3">
              <div class="w-10 h-10 rounded-xl bg-indigo-50 text-indigo-700 flex items-center justify-center font-bold">
                <History class="h-5 w-5" />
              </div>
              <div>
                <h2 class="text-xl font-black text-slate-900">Kartu Riwayat Stok</h2>
                <p class="text-xs text-slate-500 font-bold">{{ selectedItemHistory.name }} (Stok: {{ selectedItemHistory.quantity }} {{ selectedItemHistory.unit }})</p>
              </div>
            </div>
            <button @click="showItemHistoryModal = false" class="p-2 text-slate-400 hover:text-slate-700 hover:bg-slate-100 rounded-xl transition-all">
              <X class="h-5 w-5" />
            </button>
          </div>

          <div class="p-8 overflow-y-auto flex-1">
            <div v-if="isLoadingItemLogs" class="flex items-center justify-center py-16">
              <div class="w-8 h-8 border-4 border-rose-700 border-t-transparent rounded-full animate-spin"></div>
            </div>

            <div v-else-if="selectedItemLogs.length === 0" class="text-center py-12 text-slate-400">
              <History class="h-12 w-12 mx-auto mb-3 opacity-20" />
              <p class="font-bold text-slate-600">Belum ada riwayat mutasi untuk barang ini</p>
            </div>

            <div v-else class="space-y-4">
              <div
                v-for="log in selectedItemLogs"
                :key="log.id"
                class="p-4 bg-slate-50 rounded-2xl border border-slate-100 flex items-start justify-between gap-4"
              >
                <div>
                  <div class="flex items-center gap-2 mb-1">
                    <span
                      class="px-2 py-0.5 rounded text-[10px] font-black border"
                      :class="getLogTypeBadgeClass(log.type)"
                    >
                      {{ getLogTypeLabel(log.type) }}
                    </span>
                    <span class="text-xs font-bold text-slate-500">
                      {{ new Date(log.created_at).toLocaleString('id-ID') }}
                    </span>
                  </div>

                  <p v-if="log.recipient_name" class="text-xs font-bold text-slate-800 mt-1">
                    Penerima: {{ log.recipient_name }} <span class="text-slate-400">({{ log.recipient_type }})</span>
                  </p>
                  <p v-if="log.notes" class="text-xs text-slate-600 mt-0.5">{{ log.notes }}</p>
                  <p v-if="log.reference_no" class="text-[10px] font-mono text-slate-400 mt-0.5">Ref: {{ log.reference_no }}</p>
                </div>

                <div class="text-right">
                  <p class="text-sm font-black" :class="{
                    'text-emerald-600': log.type === 'IN' || log.type === 'RETURN',
                    'text-rose-600': log.type === 'OUT',
                    'text-amber-600': log.type === 'ADJUSTMENT'
                  }">
                    <span v-if="log.type === 'IN' || log.type === 'RETURN'">+</span>
                    <span v-else-if="log.type === 'OUT'">-</span>
                    {{ log.quantity }} {{ selectedItemHistory.unit }}
                  </p>
                  <p class="text-[11px] font-bold text-slate-400 mt-1">
                    Sisa: {{ log.quantity_after }}
                  </p>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </Teleport>

    <!-- Photo Preview Modal -->
    <Teleport to="body">
      <div v-if="previewImageUrl" class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/80" @click="previewImageUrl = ''">
        <div class="relative max-w-xl max-h-[85vh] bg-white p-2 rounded-2xl overflow-hidden" @click.stop>
          <img :src="previewImageUrl" class="w-full h-auto max-h-[75vh] object-contain rounded-xl" />
          <div class="pt-3 text-right">
            <button @click="previewImageUrl = ''" class="px-4 py-2 bg-slate-100 text-slate-700 font-bold text-xs rounded-xl hover:bg-slate-200">
              Tutup
            </button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>
