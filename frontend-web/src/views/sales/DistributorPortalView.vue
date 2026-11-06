<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import apiClient from '@/api/axios'
import { useAuthStore } from '@/stores/auth'
import { useFactoryStore } from '@/stores/factory'
import {
  TrendingUp,
  Store,
  ShoppingCart,
  MapPin,
  UploadCloud,
  FileSpreadsheet,
  Download,
  Plus,
  Search,
  CheckCircle,
  AlertCircle,
  Clock,
  Calendar,
  Trash2,
  X,
  ExternalLink,
  ChevronLeft,
  ChevronRight,
  RefreshCw,
} from 'lucide-vue-next'

const route = useRoute()
const authStore = useAuthStore()
const factoryStore = useFactoryStore()

// State
const activeTab = ref<'sales' | 'import' | 'stores' | 'visits'>('sales')
const loading = ref(false)
const summary = ref({
  month: new Date().getMonth() + 1,
  year: new Date().getFullYear(),
  total_sell_out_amount: 0,
  total_transactions: 0,
  total_assigned_stores: 0,
  total_visits: 0,
  top_products: [] as any[],
})

// Stores under this distributor
const stores = ref<any[]>([])
const storesTotal = ref(0)
const storesSearch = ref('')
const storesPage = ref(1)

// Sales transactions
const sales = ref<any[]>([])
const salesTotal = ref(0)
const salesPage = ref(1)
const salesStartDate = ref('')
const salesEndDate = ref('')

// Visits
const visits = ref<any[]>([])
const visitsTotal = ref(0)
const visitsPage = ref(1)

// Form Modals
const showCreateSaleModal = ref(false)
const showCreateStoreModal = ref(false)
const savingSale = ref(false)
const savingStore = ref(false)

// New Sale Form
const newSaleForm = ref({
  store_id: '',
  transaction_date: new Date().toISOString().split('T')[0],
  payment_method: 'CASH',
  notes: '',
  items: [
    {
      product_id: '',
      quantity: 1,
      price: 0,
      discount_percent: 0,
    },
  ],
})

// New Store Form
const newStoreForm = ref({
  name: '',
  owner_name: '',
  phone_number: '',
  address: '',
  customer_type: 'TOKO',
  province: '',
  regency: '',
  district: '',
  village: '',
})

// Import Excel State
const selectedFile = ref<File | null>(null)
const importing = ref(false)
const importResult = ref<any>(null)
const importError = ref('')

// Currency formatter
const formatRupiah = (val: number) => {
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    maximumFractionDigits: 0,
  }).format(val || 0)
}

// Format Date
const formatDate = (dateStr: string) => {
  if (!dateStr) return '-'
  const d = new Date(dateStr)
  return d.toLocaleDateString('id-ID', {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
  })
}

// Fetch Summary
const fetchSummary = async () => {
  try {
    const res = await apiClient.get('/distributor/summary')
    if (res.data?.status && res.data?.data) {
      summary.value = res.data.data
    }
  } catch (err: any) {
    console.error('Failed to fetch distributor summary:', err)
  }
}

// Fetch Stores
const fetchStores = async () => {
  try {
    loading.value = true
    const res = await apiClient.get('/distributor/stores', {
      params: {
        search: storesSearch.value,
        page: storesPage.value,
        limit: 15,
      },
    })
    if (res.data?.status && res.data?.data) {
      stores.value = res.data.data.stores || []
      storesTotal.value = res.data.data.total || 0
    }
  } catch (err: any) {
    console.error('Failed to fetch distributor stores:', err)
  } finally {
    loading.value = false
  }
}

// Fetch Sales
const fetchSales = async () => {
  try {
    loading.value = true
    const res = await apiClient.get('/distributor/sales', {
      params: {
        start_date: salesStartDate.value || undefined,
        end_date: salesEndDate.value || undefined,
        page: salesPage.value,
        limit: 15,
      },
    })
    if (res.data?.status && res.data?.data) {
      sales.value = res.data.data.sales || []
      salesTotal.value = res.data.data.total || 0
    }
  } catch (err: any) {
    console.error('Failed to fetch distributor sales:', err)
  } finally {
    loading.value = false
  }
}

// Fetch Visits
const fetchVisits = async () => {
  try {
    loading.value = true
    const res = await apiClient.get('/distributor/visits', {
      params: {
        page: visitsPage.value,
        limit: 15,
      },
    })
    if (res.data?.status && res.data?.data) {
      visits.value = res.data.data.visits || []
      visitsTotal.value = res.data.data.total || 0
    }
  } catch (err: any) {
    console.error('Failed to fetch distributor visits:', err)
  } finally {
    loading.value = false
  }
}

// Add Item Row to New Sale
const addSaleItem = () => {
  newSaleForm.value.items.push({
    product_id: '',
    quantity: 1,
    price: 0,
    discount_percent: 0,
  })
}

// Remove Item Row
const removeSaleItem = (idx: number) => {
  if (newSaleForm.value.items.length > 1) {
    newSaleForm.value.items.splice(idx, 1)
  }
}

// When product is selected, auto-fill price from RBP or selling_price
const onProductChange = (idx: number) => {
  const item = newSaleForm.value.items[idx]
  if (!item) return
  const prod = factoryStore.products.find((p: any) => p.id === item.product_id)
  if (prod) {
    item.price = prod.rbp > 0 ? prod.rbp : (prod.selling_price > 0 ? prod.selling_price : prod.cost_price || 0)
  }
}

// Calculate total for sale form
const calculatedTotal = computed(() => {
  return newSaleForm.value.items.reduce((acc, it) => {
    const sub = (it.quantity || 0) * (it.price || 0)
    const disc = sub * ((it.discount_percent || 0) / 100)
    return acc + (sub - disc)
  }, 0)
})

// Submit New Sale
const submitSale = async () => {
  if (!newSaleForm.value.store_id) {
    alert('Pilih toko retail terlebih dahulu')
    return
  }
  for (const it of newSaleForm.value.items) {
    if (!it.product_id) {
      alert('Semua baris produk harus dipilih')
      return
    }
    if (it.quantity <= 0) {
      alert('Jumlah kuantitas harus lebih dari 0')
      return
    }
  }

  try {
    savingSale.value = true
    const res = await apiClient.post('/distributor/sales', newSaleForm.value)
    if (res.data?.status) {
      alert('Penjualan toko berhasil dicatat!')
      showCreateSaleModal.value = false
      newSaleForm.value = {
        store_id: '',
        transaction_date: new Date().toISOString().split('T')[0],
        payment_method: 'CASH',
        notes: '',
        items: [
          {
            product_id: '',
            quantity: 1,
            price: 0,
            discount_percent: 0,
          },
        ],
      }
      fetchSales()
      fetchSummary()
    }
  } catch (err: any) {
    alert(err.response?.data?.message || 'Gagal menyimpan transaksi')
  } finally {
    savingSale.value = false
  }
}

// Submit New Store
const submitStore = async () => {
  if (!newStoreForm.value.name) {
    alert('Nama toko wajib diisi')
    return
  }

  try {
    savingStore.value = true
    const res = await apiClient.post('/distributor/stores', newStoreForm.value)
    if (res.data?.status) {
      alert('Toko binaan berhasil didaftarkan!')
      showCreateStoreModal.value = false
      newStoreForm.value = {
        name: '',
        owner_name: '',
        phone_number: '',
        address: '',
        customer_type: 'TOKO',
        province: '',
        regency: '',
        district: '',
        village: '',
      }
      fetchStores()
      fetchSummary()
    }
  } catch (err: any) {
    alert(err.response?.data?.message || 'Gagal mendaftarkan toko')
  } finally {
    savingStore.value = false
  }
}

// Handle File Selection
const onFileSelected = (event: Event) => {
  const target = event.target as HTMLInputElement
  if (target.files && target.files[0]) {
    selectedFile.value = target.files[0]
    importError.value = ''
    importResult.value = null
  }
}

// Upload Excel File
const uploadExcel = async () => {
  if (!selectedFile.value) {
    alert('Pilih file Excel terlebih dahulu')
    return
  }

  const formData = new FormData()
  formData.append('file', selectedFile.value)

  try {
    importing.value = true
    importError.value = ''
    importResult.value = null

    const res = await apiClient.post('/distributor/sales/import-excel', formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    })

    if (res.data?.status && res.data?.data) {
      importResult.value = res.data.data
      fetchSales()
      fetchSummary()
      fetchStores()
    }
  } catch (err: any) {
    importError.value = err.response?.data?.message || 'Gagal memproses file Excel'
  } finally {
    importing.value = false
  }
}

// Download Excel Template
const downloadTemplate = async () => {
  try {
    const res = await apiClient.get('/distributor/sales/template-excel', {
      responseType: 'blob',
    })
    const url = window.URL.createObjectURL(new Blob([res.data]))
    const link = document.createElement('a')
    link.href = url
    link.setAttribute('download', `Template_Penjualan_Distributor.xlsx`)
    document.body.appendChild(link)
    link.click()
    link.remove()
  } catch (err: any) {
    alert('Gagal mengunduh template Excel')
  }
}

onMounted(async () => {
  if (route.query.tab) {
    activeTab.value = route.query.tab as any
  }
  await factoryStore.fetchProducts()
  fetchSummary()
  fetchStores()
  fetchSales()
  fetchVisits()
})
</script>

<template>
  <div class="space-y-6">
    <!-- Header -->
    <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 bg-white p-6 rounded-2xl shadow-sm border border-slate-100">
      <div>
        <div class="flex items-center gap-2">
          <span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-semibold bg-emerald-50 text-emerald-700 border border-emerald-200">
            Kanal Distributor
          </span>
          <h1 class="text-2xl font-bold text-slate-800">Portal Penjualan Distributor (Sell-Out)</h1>
        </div>
        <p class="text-sm text-slate-500 mt-1">
          Catat transaksi penjualan ke toko retail/warung binaan dan pantau aktivitas kunjungan sales di wilayah Anda.
        </p>
      </div>

      <div class="flex items-center gap-2">
        <button
          @click="showCreateStoreModal = true"
          class="inline-flex items-center gap-1.5 px-4 py-2 text-sm font-medium text-slate-700 bg-white border border-slate-200 rounded-xl hover:bg-slate-50 transition shadow-sm"
        >
          <Store class="w-4 h-4 text-slate-500" />
          + Toko Binaan
        </button>
        <button
          @click="showCreateSaleModal = true"
          class="inline-flex items-center gap-1.5 px-4 py-2 text-sm font-semibold text-white bg-blue-600 rounded-xl hover:bg-blue-700 transition shadow-md shadow-blue-500/20"
        >
          <Plus class="w-4 h-4" />
          Input Penjualan
        </button>
      </div>
    </div>

    <!-- Summary KPI Cards -->
    <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
      <div class="bg-white p-5 rounded-2xl border border-slate-100 shadow-sm flex items-center gap-4">
        <div class="w-12 h-12 rounded-xl bg-blue-50 flex items-center justify-center text-blue-600">
          <TrendingUp class="w-6 h-6" />
        </div>
        <div>
          <div class="text-xs font-medium text-slate-400 uppercase tracking-wider">Omzet Sell-Out Bulan Ini</div>
          <div class="text-xl font-bold text-slate-800 mt-0.5">{{ formatRupiah(summary.total_sell_out_amount) }}</div>
        </div>
      </div>

      <div class="bg-white p-5 rounded-2xl border border-slate-100 shadow-sm flex items-center gap-4">
        <div class="w-12 h-12 rounded-xl bg-emerald-50 flex items-center justify-center text-emerald-600">
          <ShoppingCart class="w-6 h-6" />
        </div>
        <div>
          <div class="text-xs font-medium text-slate-400 uppercase tracking-wider">Total Transaksi</div>
          <div class="text-xl font-bold text-slate-800 mt-0.5">{{ summary.total_transactions }} Nota</div>
        </div>
      </div>

      <div class="bg-white p-5 rounded-2xl border border-slate-100 shadow-sm flex items-center gap-4">
        <div class="w-12 h-12 rounded-xl bg-indigo-50 flex items-center justify-center text-indigo-600">
          <Store class="w-6 h-6" />
        </div>
        <div>
          <div class="text-xs font-medium text-slate-400 uppercase tracking-wider">Toko Binaan Aktif</div>
          <div class="text-xl font-bold text-slate-800 mt-0.5">{{ summary.total_assigned_stores }} Toko</div>
        </div>
      </div>

      <div class="bg-white p-5 rounded-2xl border border-slate-100 shadow-sm flex items-center gap-4">
        <div class="w-12 h-12 rounded-xl bg-purple-50 flex items-center justify-center text-purple-600">
          <MapPin class="w-6 h-6" />
        </div>
        <div>
          <div class="text-xs font-medium text-slate-400 uppercase tracking-wider">Kunjungan Sales</div>
          <div class="text-xl font-bold text-slate-800 mt-0.5">{{ summary.total_visits }} Kali</div>
        </div>
      </div>
    </div>

    <!-- Navigation Tabs -->
    <div class="bg-white rounded-2xl border border-slate-100 shadow-sm overflow-hidden">
      <div class="flex border-b border-slate-100 px-6 pt-2 bg-slate-50/50">
        <button
          @click="activeTab = 'sales'"
          :class="[
            'py-3.5 px-4 text-sm font-semibold border-b-2 transition-all flex items-center gap-2',
            activeTab === 'sales'
              ? 'border-blue-600 text-blue-600'
              : 'border-transparent text-slate-500 hover:text-slate-800',
          ]"
        >
          <ShoppingCart class="w-4 h-4" />
          Riwayat Penjualan (Sell-Out)
        </button>
        <button
          @click="activeTab = 'import'"
          :class="[
            'py-3.5 px-4 text-sm font-semibold border-b-2 transition-all flex items-center gap-2',
            activeTab === 'import'
              ? 'border-blue-600 text-blue-600'
              : 'border-transparent text-slate-500 hover:text-slate-800',
          ]"
        >
          <FileSpreadsheet class="w-4 h-4" />
          Bulk Import Excel
        </button>
        <button
          @click="activeTab = 'stores'"
          :class="[
            'py-3.5 px-4 text-sm font-semibold border-b-2 transition-all flex items-center gap-2',
            activeTab === 'stores'
              ? 'border-blue-600 text-blue-600'
              : 'border-transparent text-slate-500 hover:text-slate-800',
          ]"
        >
          <Store class="w-4 h-4" />
          Toko Binaan
        </button>
        <button
          @click="activeTab = 'visits'"
          :class="[
            'py-3.5 px-4 text-sm font-semibold border-b-2 transition-all flex items-center gap-2',
            activeTab === 'visits'
              ? 'border-blue-600 text-blue-600'
              : 'border-transparent text-slate-500 hover:text-slate-800',
          ]"
        >
          <MapPin class="w-4 h-4" />
          Kunjungan Lapangan
        </button>
      </div>

      <!-- Tab Content: Penjualan -->
      <div v-if="activeTab === 'sales'" class="p-6 space-y-4">
        <div class="flex flex-col sm:flex-row items-center justify-between gap-4 pb-2">
          <div class="flex items-center gap-2 w-full sm:w-auto">
            <input
              type="date"
              v-model="salesStartDate"
              @change="fetchSales"
              class="text-xs bg-slate-50 border border-slate-200 rounded-lg px-3 py-2 text-slate-600 outline-none"
            />
            <span class="text-slate-400 text-xs">s/d</span>
            <input
              type="date"
              v-model="salesEndDate"
              @change="fetchSales"
              class="text-xs bg-slate-50 border border-slate-200 rounded-lg px-3 py-2 text-slate-600 outline-none"
            />
          </div>

          <div class="text-xs text-slate-500">
            Total Transaksi: <span class="font-bold text-slate-800">{{ salesTotal }}</span>
          </div>
        </div>

        <!-- Table -->
        <div class="overflow-x-auto border border-slate-100 rounded-xl">
          <table class="w-full text-left text-sm text-slate-600">
            <thead class="bg-slate-50 text-xs font-semibold text-slate-700 uppercase tracking-wider border-b border-slate-100">
              <tr>
                <th class="py-3 px-4">No. Nota</th>
                <th class="py-3 px-4">Tanggal</th>
                <th class="py-3 px-4">Toko Retail</th>
                <th class="py-3 px-4">Item & Qty</th>
                <th class="py-3 px-4 text-right">Total Transaksi</th>
                <th class="py-3 px-4 text-center">Metode</th>
                <th class="py-3 px-4 text-center">Status</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-100">
              <tr v-if="loading" class="text-center py-8">
                <td colspan="7" class="py-8 text-slate-400">Memuat data penjualan...</td>
              </tr>
              <tr v-else-if="sales.length === 0" class="text-center py-8">
                <td colspan="7" class="py-8 text-slate-400">Belum ada transaksi penjualan yang tercatat.</td>
              </tr>
              <tr v-for="sale in sales" :key="sale.id" class="hover:bg-slate-50/50 transition">
                <td class="py-3.5 px-4 font-mono font-medium text-blue-600 text-xs">
                  {{ sale.receipt_no }}
                </td>
                <td class="py-3.5 px-4 text-xs text-slate-500 whitespace-nowrap">
                  {{ formatDate(sale.transaction_date) }}
                </td>
                <td class="py-3.5 px-4 font-medium text-slate-800">
                  {{ sale.store?.name || 'Umum / Walk-in' }}
                </td>
                <td class="py-3.5 px-4 text-xs">
                  <div v-for="item in sale.items" :key="item.id" class="text-slate-600">
                    • {{ item.product?.name || 'Produk' }} x{{ item.quantity }}
                  </div>
                </td>
                <td class="py-3.5 px-4 font-bold text-slate-800 text-right whitespace-nowrap">
                  {{ formatRupiah(sale.total_amount) }}
                </td>
                <td class="py-3.5 px-4 text-center">
                  <span class="inline-block px-2 py-0.5 text-[11px] font-semibold bg-slate-100 text-slate-700 rounded-md">
                    {{ sale.payment_method || 'CASH' }}
                  </span>
                </td>
                <td class="py-3.5 px-4 text-center">
                  <span class="inline-block px-2.5 py-0.5 text-[11px] font-semibold bg-emerald-50 text-emerald-700 border border-emerald-200 rounded-full">
                    {{ sale.status || 'VERIFIED' }}
                  </span>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <!-- Tab Content: Bulk Import Excel -->
      <div v-if="activeTab === 'import'" class="p-6 max-w-4xl mx-auto space-y-6">
        <div class="bg-blue-50/50 border border-blue-200/60 rounded-2xl p-6 flex flex-col md:flex-row items-center justify-between gap-6">
          <div class="space-y-1">
            <h3 class="font-bold text-slate-800 text-base flex items-center gap-2">
              <Download class="w-5 h-5 text-blue-600" />
              Unduh Template Excel Resmi
            </h3>
            <p class="text-xs text-slate-600 leading-relaxed">
              Gunakan format spreadsheet standar agar sistem dapat memvalidasi SKU produk, nama toko retail binaan, dan harga jual secara otomatis.
            </p>
          </div>
          <button
            @click="downloadTemplate"
            class="whitespace-nowrap px-4 py-2.5 bg-blue-600 hover:bg-blue-700 text-white font-semibold text-xs rounded-xl shadow-md shadow-blue-500/20 transition flex items-center gap-2"
          >
            <Download class="w-4 h-4" />
            Unduh Template .xlsx
          </button>
        </div>

        <!-- Upload Dropzone -->
        <div class="border-2 border-dashed border-slate-200 rounded-2xl p-8 text-center bg-slate-50/30 hover:bg-slate-50/60 transition">
          <UploadCloud class="w-12 h-12 text-slate-400 mx-auto mb-3" />
          <h4 class="text-sm font-bold text-slate-700">Pilih atau Seret File Excel Penjualan</h4>
          <p class="text-xs text-slate-500 mt-1 mb-4">Mendukung format .xlsx dengan ukuran file hingga 10MB</p>

          <input
            type="file"
            id="excelFile"
            accept=".xlsx"
            @change="onFileSelected"
            class="hidden"
          />
          <label
            for="excelFile"
            class="inline-block px-5 py-2.5 bg-white border border-slate-200 text-slate-700 font-semibold text-xs rounded-xl cursor-pointer hover:bg-slate-50 shadow-sm transition"
          >
            {{ selectedFile ? selectedFile.name : 'Pilih Berkas Excel' }}
          </label>

          <div v-if="selectedFile" class="mt-4 flex items-center justify-center gap-3">
            <button
              @click="uploadExcel"
              :disabled="importing"
              class="px-6 py-2.5 bg-emerald-600 hover:bg-emerald-700 text-white font-semibold text-xs rounded-xl shadow-md shadow-emerald-500/20 transition flex items-center gap-2 disabled:opacity-50"
            >
              <RefreshCw v-if="importing" class="w-4 h-4 animate-spin" />
              <CheckCircle v-else class="w-4 h-4" />
              {{ importing ? 'Memproses Data...' : 'Mulai Import Penjualan' }}
            </button>
          </div>
        </div>

        <!-- Import Alert & Results -->
        <div v-if="importError" class="p-4 bg-red-50 border border-red-200 rounded-xl text-xs text-red-700 flex items-start gap-3">
          <AlertCircle class="w-5 h-5 flex-shrink-0 text-red-600" />
          <div>
            <div class="font-bold">Gagal Mengimpor File</div>
            <div>{{ importError }}</div>
          </div>
        </div>

        <div v-if="importResult" class="p-5 bg-white border border-slate-100 shadow-sm rounded-2xl space-y-3">
          <div class="flex items-center justify-between border-b border-slate-100 pb-3">
            <h4 class="font-bold text-slate-800 text-sm">Hasil Pemrosesan File</h4>
            <span class="text-xs font-semibold px-2.5 py-0.5 rounded-full bg-emerald-50 text-emerald-700 border border-emerald-200">
              Selesai
            </span>
          </div>

          <div class="grid grid-cols-3 gap-4 text-center py-2">
            <div class="p-3 bg-slate-50 rounded-xl">
              <div class="text-xs text-slate-400">Total Baris</div>
              <div class="text-lg font-bold text-slate-800">{{ importResult.total_rows }}</div>
            </div>
            <div class="p-3 bg-emerald-50 rounded-xl">
              <div class="text-xs text-emerald-600">Berhasil Masuk</div>
              <div class="text-lg font-bold text-emerald-700">{{ importResult.success_count }}</div>
            </div>
            <div class="p-3 bg-rose-50 rounded-xl">
              <div class="text-xs text-rose-600">Gagal / Dilewati</div>
              <div class="text-lg font-bold text-rose-700">{{ importResult.error_count }}</div>
            </div>
          </div>

          <div v-if="importResult.errors && importResult.errors.length > 0" class="mt-3">
            <div class="text-xs font-semibold text-rose-700 mb-1">Daftar Peringatan / Baris Bermasalah:</div>
            <ul class="text-xs text-rose-600 space-y-1 max-h-40 overflow-y-auto bg-rose-50/50 p-3 rounded-xl border border-rose-100">
              <li v-for="(e, i) in importResult.errors" :key="i">• {{ e }}</li>
            </ul>
          </div>
        </div>
      </div>

      <!-- Tab Content: Toko Binaan -->
      <div v-if="activeTab === 'stores'" class="p-6 space-y-4">
        <div class="flex flex-col sm:flex-row items-center justify-between gap-4">
          <div class="relative w-full sm:w-72">
            <Search class="w-4 h-4 text-slate-400 absolute left-3 top-2.5" />
            <input
              type="text"
              v-model="storesSearch"
              @input="fetchStores"
              placeholder="Cari toko binaan..."
              class="w-full text-xs pl-9 pr-3 py-2 bg-slate-50 border border-slate-200 rounded-xl outline-none"
            />
          </div>

          <div class="text-xs text-slate-500">
            Total Toko Binaan: <span class="font-bold text-slate-800">{{ storesTotal }}</span>
          </div>
        </div>

        <div class="overflow-x-auto border border-slate-100 rounded-xl">
          <table class="w-full text-left text-sm text-slate-600">
            <thead class="bg-slate-50 text-xs font-semibold text-slate-700 uppercase tracking-wider border-b border-slate-100">
              <tr>
                <th class="py-3 px-4">Nama Toko</th>
                <th class="py-3 px-4">Pemilik</th>
                <th class="py-3 px-4">No. Telepon</th>
                <th class="py-3 px-4">Alamat & Wilayah</th>
                <th class="py-3 px-4 text-center">Tipe</th>
                <th class="py-3 px-4 text-center">Status</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-100">
              <tr v-if="stores.length === 0" class="text-center py-8">
                <td colspan="6" class="py-8 text-slate-400">Belum ada toko binaan yang terdaftar.</td>
              </tr>
              <tr v-for="st in stores" :key="st.id" class="hover:bg-slate-50/50 transition">
                <td class="py-3.5 px-4 font-semibold text-slate-800">
                  {{ st.name }}
                </td>
                <td class="py-3.5 px-4 text-xs text-slate-600">
                  {{ st.owner_name || '-' }}
                </td>
                <td class="py-3.5 px-4 text-xs text-slate-600 font-mono">
                  {{ st.phone_number || '-' }}
                </td>
                <td class="py-3.5 px-4 text-xs text-slate-500 max-w-xs truncate">
                  {{ st.address || '-' }} {{ st.district ? `(${st.district})` : '' }}
                </td>
                <td class="py-3.5 px-4 text-center">
                  <span class="inline-block px-2 py-0.5 text-[11px] font-semibold bg-slate-100 text-slate-700 rounded-md">
                    {{ st.customer_type || 'TOKO' }}
                  </span>
                </td>
                <td class="py-3.5 px-4 text-center">
                  <span class="inline-block px-2.5 py-0.5 text-[11px] font-semibold bg-emerald-50 text-emerald-700 border border-emerald-200 rounded-full">
                    Aktif
                  </span>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <!-- Tab Content: Kunjungan -->
      <div v-if="activeTab === 'visits'" class="p-6 space-y-4">
        <div class="overflow-x-auto border border-slate-100 rounded-xl">
          <table class="w-full text-left text-sm text-slate-600">
            <thead class="bg-slate-50 text-xs font-semibold text-slate-700 uppercase tracking-wider border-b border-slate-100">
              <tr>
                <th class="py-3 px-4">Waktu Check-in</th>
                <th class="py-3 px-4">Toko Dikunjungi</th>
                <th class="py-3 px-4">Petugas / Sales</th>
                <th class="py-3 px-4">Catatan Kunjungan</th>
                <th class="py-3 px-4 text-center">Foto Toko</th>
                <th class="py-3 px-4 text-center">Status</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-100">
              <tr v-if="visits.length === 0" class="text-center py-8">
                <td colspan="6" class="py-8 text-slate-400">Belum ada data kunjungan yang tercatat.</td>
              </tr>
              <tr v-for="v in visits" :key="v.id" class="hover:bg-slate-50/50 transition">
                <td class="py-3.5 px-4 text-xs font-mono text-slate-600 whitespace-nowrap">
                  {{ formatDate(v.check_in_time) }} {{ new Date(v.check_in_time).toLocaleTimeString('id-ID', { hour: '2-digit', minute: '2-digit' }) }}
                </td>
                <td class="py-3.5 px-4 font-semibold text-slate-800">
                  {{ v.store?.name || 'Toko' }}
                </td>
                <td class="py-3.5 px-4 text-xs text-slate-600">
                  {{ v.user?.email || v.employee?.first_name || 'Sales Distributor' }}
                </td>
                <td class="py-3.5 px-4 text-xs text-slate-500">
                  {{ v.notes || '-' }}
                </td>
                <td class="py-3.5 px-4 text-center">
                  <a
                    v-if="v.selfie_url"
                    :href="v.selfie_url"
                    target="_blank"
                    class="text-xs text-blue-600 font-semibold hover:underline inline-flex items-center gap-1"
                  >
                    Lihat Foto
                    <ExternalLink class="w-3 h-3" />
                  </a>
                  <span v-else class="text-xs text-slate-400">-</span>
                </td>
                <td class="py-3.5 px-4 text-center">
                  <span class="inline-block px-2.5 py-0.5 text-[11px] font-semibold bg-emerald-50 text-emerald-700 border border-emerald-200 rounded-full">
                    {{ v.status || 'COMPLETED' }}
                  </span>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>

    <!-- Modal Input Penjualan Baru -->
    <div v-if="showCreateSaleModal" class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-900/60 backdrop-blur-sm">
      <div class="bg-white rounded-2xl shadow-xl w-full max-w-3xl overflow-hidden flex flex-col max-h-[90vh]">
        <div class="px-6 py-4 border-b border-slate-100 flex items-center justify-between">
          <h3 class="font-bold text-slate-800 text-base">Input Penjualan Toko Retail (Sell-Out)</h3>
          <button @click="showCreateSaleModal = false" class="text-slate-400 hover:text-slate-600">
            <X class="w-5 h-5" />
          </button>
        </div>

        <div class="p-6 overflow-y-auto space-y-4 flex-1">
          <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <div>
              <label class="block text-xs font-semibold text-slate-700 mb-1">Toko Retail Binaan *</label>
              <select
                v-model="newSaleForm.store_id"
                class="w-full text-xs p-2.5 bg-slate-50 border border-slate-200 rounded-xl outline-none"
              >
                <option value="">-- Pilih Toko Binaan --</option>
                <option v-for="st in stores" :key="st.id" :value="st.id">
                  {{ st.name }} ({{ st.owner_name || st.district || 'Retail' }})
                </option>
              </select>
            </div>

            <div>
              <label class="block text-xs font-semibold text-slate-700 mb-1">Tanggal Transaksi</label>
              <input
                type="date"
                v-model="newSaleForm.transaction_date"
                class="w-full text-xs p-2.5 bg-slate-50 border border-slate-200 rounded-xl outline-none"
              />
            </div>
          </div>

          <!-- Items Table -->
          <div>
            <div class="flex items-center justify-between mb-2">
              <label class="text-xs font-semibold text-slate-700">Item Produk yang Dijual</label>
              <button
                type="button"
                @click="addSaleItem"
                class="text-xs text-blue-600 font-semibold hover:underline flex items-center gap-1"
              >
                <Plus class="w-3.5 h-3.5" />
                Tambah Produk
              </button>
            </div>

            <div class="space-y-2">
              <div
                v-for="(item, idx) in newSaleForm.items"
                :key="idx"
                class="grid grid-cols-12 gap-2 p-3 bg-slate-50/70 border border-slate-100 rounded-xl items-center"
              >
                <div class="col-span-5">
                  <select
                    v-model="item.product_id"
                    @change="onProductChange(idx)"
                    class="w-full text-xs p-2 bg-white border border-slate-200 rounded-lg outline-none"
                  >
                    <option value="">-- Pilih Produk --</option>
                    <option v-for="p in factoryStore.products" :key="p.id" :value="p.id">
                      {{ p.name }} ({{ p.sku || 'Produk' }})
                    </option>
                  </select>
                </div>

                <div class="col-span-2">
                  <input
                    type="number"
                    min="1"
                    v-model.number="item.quantity"
                    placeholder="Qty"
                    class="w-full text-xs p-2 bg-white border border-slate-200 rounded-lg outline-none text-center"
                  />
                </div>

                <div class="col-span-3">
                  <input
                    type="number"
                    v-model.number="item.price"
                    placeholder="Harga Satuan"
                    class="w-full text-xs p-2 bg-white border border-slate-200 rounded-lg outline-none text-right font-mono"
                  />
                </div>

                <div class="col-span-2 flex items-center justify-end gap-1">
                  <button
                    type="button"
                    @click="removeSaleItem(idx)"
                    class="text-rose-500 hover:text-rose-700 p-1"
                    title="Hapus baris"
                  >
                    <Trash2 class="w-4 h-4" />
                  </button>
                </div>
              </div>
            </div>
          </div>

          <!-- Total Calculation -->
          <div class="p-4 bg-slate-50 rounded-xl flex items-center justify-between">
            <span class="text-xs font-semibold text-slate-600">Total Transaksi:</span>
            <span class="text-lg font-bold text-blue-600 font-mono">{{ formatRupiah(calculatedTotal) }}</span>
          </div>

          <div>
            <label class="block text-xs font-semibold text-slate-700 mb-1">Catatan Tambahan (Opsional)</label>
            <input
              type="text"
              v-model="newSaleForm.notes"
              placeholder="Misal: Penjualan tunai langsung di toko"
              class="w-full text-xs p-2.5 bg-slate-50 border border-slate-200 rounded-xl outline-none"
            />
          </div>
        </div>

        <div class="px-6 py-4 border-t border-slate-100 flex items-center justify-end gap-2 bg-slate-50/50">
          <button
            @click="showCreateSaleModal = false"
            class="px-4 py-2 text-xs font-semibold text-slate-600 hover:text-slate-800"
          >
            Batal
          </button>
          <button
            @click="submitSale"
            :disabled="savingSale"
            class="px-5 py-2 bg-blue-600 hover:bg-blue-700 text-white font-semibold text-xs rounded-xl shadow-md shadow-blue-500/20 disabled:opacity-50"
          >
            {{ savingSale ? 'Menyimpan...' : 'Simpan Transaksi' }}
          </button>
        </div>
      </div>
    </div>

    <!-- Modal Tambah Toko Binaan -->
    <div v-if="showCreateStoreModal" class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-900/60 backdrop-blur-sm">
      <div class="bg-white rounded-2xl shadow-xl w-full max-w-lg overflow-hidden flex flex-col">
        <div class="px-6 py-4 border-b border-slate-100 flex items-center justify-between">
          <h3 class="font-bold text-slate-800 text-base">Daftarkan Toko Binaan Baru</h3>
          <button @click="showCreateStoreModal = false" class="text-slate-400 hover:text-slate-600">
            <X class="w-5 h-5" />
          </button>
        </div>

        <div class="p-6 space-y-3">
          <div>
            <label class="block text-xs font-semibold text-slate-700 mb-1">Nama Toko / Warung *</label>
            <input
              type="text"
              v-model="newStoreForm.name"
              placeholder="Contoh: Toko Berkah Jaya"
              class="w-full text-xs p-2.5 bg-slate-50 border border-slate-200 rounded-xl outline-none"
            />
          </div>

          <div class="grid grid-cols-2 gap-3">
            <div>
              <label class="block text-xs font-semibold text-slate-700 mb-1">Nama Pemilik</label>
              <input
                type="text"
                v-model="newStoreForm.owner_name"
                placeholder="Nama pemilik toko"
                class="w-full text-xs p-2.5 bg-slate-50 border border-slate-200 rounded-xl outline-none"
              />
            </div>
            <div>
              <label class="block text-xs font-semibold text-slate-700 mb-1">No. WhatsApp / HP</label>
              <input
                type="text"
                v-model="newStoreForm.phone_number"
                placeholder="08xxxxxxxx"
                class="w-full text-xs p-2.5 bg-slate-50 border border-slate-200 rounded-xl outline-none"
              />
            </div>
          </div>

          <div>
            <label class="block text-xs font-semibold text-slate-700 mb-1">Alamat Lengkap</label>
            <textarea
              v-model="newStoreForm.address"
              rows="2"
              placeholder="Alamat jalan / patokan toko"
              class="w-full text-xs p-2.5 bg-slate-50 border border-slate-200 rounded-xl outline-none"
            ></textarea>
          </div>

          <div class="grid grid-cols-2 gap-3">
            <div>
              <label class="block text-xs font-semibold text-slate-700 mb-1">Kecamatan</label>
              <input
                type="text"
                v-model="newStoreForm.district"
                placeholder="Nama kecamatan"
                class="w-full text-xs p-2.5 bg-slate-50 border border-slate-200 rounded-xl outline-none"
              />
            </div>
            <div>
              <label class="block text-xs font-semibold text-slate-700 mb-1">Kota / Kabupaten</label>
              <input
                type="text"
                v-model="newStoreForm.regency"
                placeholder="Nama kota/kabupaten"
                class="w-full text-xs p-2.5 bg-slate-50 border border-slate-200 rounded-xl outline-none"
              />
            </div>
          </div>
        </div>

        <div class="px-6 py-4 border-t border-slate-100 flex items-center justify-end gap-2 bg-slate-50/50">
          <button
            @click="showCreateStoreModal = false"
            class="px-4 py-2 text-xs font-semibold text-slate-600 hover:text-slate-800"
          >
            Batal
          </button>
          <button
            @click="submitStore"
            :disabled="savingStore"
            class="px-5 py-2 bg-blue-600 hover:bg-blue-700 text-white font-semibold text-xs rounded-xl shadow-md shadow-blue-500/20 disabled:opacity-50"
          >
            {{ savingStore ? 'Mendaftarkan...' : 'Daftarkan Toko' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
