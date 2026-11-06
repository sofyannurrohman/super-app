<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import {
  Megaphone,
  Search,
  DollarSign,
  Calendar,
  Layers,
  MapPin,
  CheckCircle2,
  FileText,
  Clock,
  TrendingUp,
  Percent,
} from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import { toast } from 'vue-sonner'
import apiClient from '@/api/axios'
import { useMasterDataStore } from '@/stores/masterData'

const masterDataStore = useMasterDataStore()

const searchQuery = ref('')
const selectedType = ref<'ALL' | 'ENDCAP' | 'GONDOLA' | 'FLOOR_DISPLAY'>('ALL')

// Real Promotion Database
const promotions = ref<any[]>([])
const keyAccounts = ref<any[]>([])

const fetchPromotions = async () => {
  try {
    const res = await apiClient.get('/modern-trade/promotions')
    promotions.value = (res.data.data || []).map((p: any) => ({
      ...p,
      key_account: p.key_account?.name || 'Ritel Modern',
      start_date: p.start_date ? p.start_date.split('T')[0] : '',
      end_date: p.end_date ? p.end_date.split('T')[0] : '',
      cost: Number(p.cost || 0),
      sales_target: Number(p.sales_target || 0),
      sales_actual: Number(p.sales_actual || 0),
    }))
  } catch (err: any) {
    toast.error('Gagal mengambil data promosi: ' + (err.response?.data?.error || err.message))
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
  fetchPromotions()
  fetchKeyAccounts()
})

const totalCost = computed(() =>
  promotions.value.reduce((sum, item) => sum + item.cost, 0)
)

const overallROI = computed(() => {
  const actual = promotions.value.reduce((sum, item) => sum + item.sales_actual, 0)
  const cost = promotions.value.reduce((sum, item) => sum + item.cost, 0)
  return cost ? Math.round((actual / cost) * 100) : 0
})

const filteredPromotions = computed(() => {
  return promotions.value.filter((promo) => {
    const matchesSearch =
      promo.key_account.toLowerCase().includes(searchQuery.value.toLowerCase()) ||
      promo.display_name.toLowerCase().includes(searchQuery.value.toLowerCase())
    const matchesType = selectedType.value === 'ALL' || promo.promo_type === selectedType.value
    return matchesSearch && matchesType
  })
})

const isUploadingPromo = ref(false)
const triggerPromoUpload = async () => {
  if (keyAccounts.value.length === 0) {
    toast.error('Belum ada data Key Account ritel modern di sistem! Silakan daftarkan Key Account terlebih dahulu di tab KAM.')
    return
  }

  isUploadingPromo.value = true
  try {
    const randomKA = keyAccounts.value[Math.floor(Math.random() * keyAccounts.value.length)]
    const promoTypes = ['ENDCAP', 'GONDOLA', 'FLOOR_DISPLAY']
    const randomType = promoTypes[Math.floor(Math.random() * promoTypes.length)] || 'GONDOLA'
    const displayNames = [
      `Sewa Gondola Utama Susu Karsa HR ${randomKA.name}`,
      `Endcap Lorong Minuman Karsa HR Coklat di ${randomKA.name}`,
      `Floor Display Depan Kasir Karsa HR ${randomKA.name}`
    ]
    const randomDisplayName = displayNames[Math.floor(Math.random() * displayNames.length)] || 'Sewa Display Baru'
    const coverage = `250 Outlet ${randomKA.name} Nasional`

    const cost = 25000000 + Math.floor(Math.random() * 8) * 10000000
    const target = cost * 2
    const actual = Math.random() > 0.3 ? target * (0.8 + Math.random() * 0.5) : 0

    await apiClient.post('/modern-trade/promotions', {
      key_account_id: randomKA.id,
      promo_type: randomType,
      display_name: randomDisplayName,
      start_date: new Date().toISOString().split('T')[0],
      end_date: new Date(new Date().setDate(new Date().getDate() + 30)).toISOString().split('T')[0],
      cost,
      sales_target: target,
      sales_actual: actual,
      status: actual > 0 ? 'COMPLETED' : 'ONGOING',
      outlet_coverage: coverage
    })

    toast.success(`Program promosi ${randomDisplayName} berhasil ditambahkan secara dinamis!`)
    fetchPromotions()
  } catch (err: any) {
    toast.error('Gagal mendaftarkan promosi: ' + (err.response?.data?.error || err.message))
  } finally {
    isUploadingPromo.value = false
  }
}

// Form states for manual CRUD dialogs
const showPromoModal = ref(false)
const isEditing = ref(false)

const promoForm = ref({
  id: '',
  key_account_id: '',
  promo_type: 'GONDOLA' as 'GONDOLA' | 'ENDCAP' | 'FLOOR_DISPLAY',
  display_name: '',
  start_date: '',
  end_date: '',
  cost: 0,
  sales_target: 0,
  sales_actual: 0,
  status: 'APPROVED',
  outlet_coverage: '',
})

const openNewPromoModal = () => {
  isEditing.value = false
  const sDate = new Date().toISOString().split('T')[0] || ''
  const eDate = new Date(new Date().setDate(new Date().getDate() + 30)).toISOString().split('T')[0] || ''
  promoForm.value = {
    id: '',
    key_account_id: keyAccounts.value[0]?.id || '',
    promo_type: 'GONDOLA',
    display_name: '',
    start_date: sDate,
    end_date: eDate,
    cost: 0,
    sales_target: 0,
    sales_actual: 0,
    status: 'APPROVED',
    outlet_coverage: '',
  }
  showPromoModal.value = true
}

const openEditPromoModal = (promo: any) => {
  isEditing.value = true
  promoForm.value = {
    ...promo,
    cost: Number(promo.cost || 0),
    sales_target: Number(promo.sales_target || 0),
    sales_actual: Number(promo.sales_actual || 0),
  }
  showPromoModal.value = true
}

const savePromo = async () => {
  if (!promoForm.value.display_name || !promoForm.value.key_account_id) {
    toast.error('Nama Program dan Ritel wajib diisi!')
    return
  }

  try {
    if (isEditing.value) {
      await apiClient.put(`/modern-trade/promotions/${promoForm.value.id}`, promoForm.value)
      toast.success('Program promosi berhasil diperbarui!')
    } else {
      await apiClient.post('/modern-trade/promotions', promoForm.value)
      toast.success('Program promosi baru berhasil didaftarkan!')
    }
    showPromoModal.value = false
    fetchPromotions()
  } catch (err: any) {
    toast.error('Gagal menyimpan program promosi: ' + (err.response?.data?.error || err.message))
  }
}

const deletePromo = async (id: string) => {
  if (confirm('Apakah Anda yakin ingin menghapus program promosi ini?')) {
    try {
      await apiClient.delete(`/modern-trade/promotions/${id}`)
      toast.success('Program promosi berhasil dihapus!')
      fetchPromotions()
    } catch (err: any) {
      toast.error('Gagal menghapus program promosi: ' + (err.response?.data?.error || err.message))
    }
  }
}

const formatCurrency = (val: number) => {
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    maximumFractionDigits: 0,
  }).format(val)
}

// Sub-tab navigation
const activeTab = ref<'PROMO' | 'SPG_EVENT'>('PROMO')

// Compute SPG & Event lists from promotions database dynamically
const spgSchedules = computed(() => {
  return promotions.value.map(promo => {
    // Generate dates based on start_date
    const rawDate = promo.start_date ? new Date(promo.start_date) : new Date()
    const startStr = rawDate.toISOString().split('T')[0]
    rawDate.setDate(rawDate.getDate() + 15)
    const endStr = rawDate.toISOString().split('T')[0]
    return {
      id: promo.id,
      retailer: promo.key_account,
      spg_name: `SPG Team ${promo.key_account.split(' ')[0]}`,
      shift: 'Shift Pagi / Sore',
      location: promo.outlet_coverage || 'Nasional',
      start_date: startStr,
      end_date: endStr,
      status: promo.status === 'COMPLETED' ? 'SELESAI' : 'AKTIF',
    }
  })
})

const eventSchedules = computed(() => {
  return promotions.value.map(promo => {
    const rawDate = promo.start_date ? new Date(promo.start_date) : new Date()
    rawDate.setDate(rawDate.getDate() + 2) // event starts slightly after
    const startStr = rawDate.toISOString().split('T')[0]
    rawDate.setDate(rawDate.getDate() + 3) // events are shorter 
    const endStr = rawDate.toISOString().split('T')[0]
    return {
      id: promo.id,
      retailer: promo.key_account,
      event_name: `Karsa HR Special Weekend - ${promo.display_name}`,
      pic: 'Marketing Event KARSA HR',
      start_date: startStr,
      end_date: endStr,
      status: promo.status === 'COMPLETED' ? 'SELESAI' : 'TERJADWAL',
    }
  })
})
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
          <span>Promosi & Sewa Display</span>
          <span v-if="masterDataStore.selectedBranchName" class="text-primary text-xl md:text-2xl font-bold">
            — {{ masterDataStore.selectedBranchName }}
          </span>
        </h1>
        <p class="text-sm text-slate-500 mt-1">
          Kelola anggaran promosi ritel modern, sewa area display eksklusif (gondola, endcap), dan analisis efektivitas ROI promosi.
        </p>
      </div>

      <div class="flex items-center gap-2.5">
        <Button
          @click="openNewPromoModal"
          class="bg-indigo-600 hover:bg-indigo-700 text-white font-bold rounded-xl shadow-md shadow-indigo-100"
        >
          <Megaphone class="w-4 h-4 mr-2" />
          Tambah Program Promosi
        </Button>
      </div>
    </div>

    <!-- Quick Stats -->
    <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
      <div class="bg-white p-5 rounded-2xl border border-slate-100 shadow-sm flex items-center gap-4">
        <div class="p-3 bg-indigo-50 text-indigo-600 rounded-xl">
          <Megaphone class="w-6 h-6" />
        </div>
        <div>
          <p class="text-xs font-bold text-slate-400 uppercase tracking-wider">Total Program Promosi</p>
          <p class="text-2xl font-black text-slate-800">{{ promotions.length }} Program</p>
        </div>
      </div>
      <div class="bg-white p-5 rounded-2xl border border-slate-100 shadow-sm flex items-center gap-4">
        <div class="p-3 bg-emerald-50 text-emerald-600 rounded-xl">
          <DollarSign class="w-6 h-6" />
        </div>
        <div>
          <p class="text-xs font-bold text-slate-400 uppercase tracking-wider">Total Investasi Promosi</p>
          <p class="text-2xl font-black text-slate-850 truncate" :title="formatCurrency(totalCost)">
            {{ formatCurrency(totalCost) }}
          </p>
        </div>
      </div>
      <div class="bg-white p-5 rounded-2xl border border-slate-100 shadow-sm flex items-center gap-4">
        <div class="p-3 bg-blue-50 text-blue-600 rounded-xl">
          <TrendingUp class="w-6 h-6" />
        </div>
        <div>
          <p class="text-xs font-bold text-slate-400 uppercase tracking-wider">ROI Penjualan Kontrak</p>
          <p class="text-2xl font-black text-slate-800">{{ overallROI }}% Efektivitas</p>
        </div>
      </div>
      <div class="bg-white p-5 rounded-2xl border border-slate-100 shadow-sm flex items-center gap-4">
        <div class="p-3 bg-amber-50 text-amber-600 rounded-xl">
          <Percent class="w-6 h-6" />
        </div>
        <div>
          <p class="text-xs font-bold text-slate-400 uppercase tracking-wider">Kontribusi Penjualan</p>
          <p class="text-2xl font-black text-slate-800">18.5% dari Total MT</p>
        </div>
      </div>
    </div>

    <!-- Sub-tab Navigation Bar -->
    <div class="flex border-b border-slate-100 bg-slate-50/50 p-1 rounded-xl w-fit">
      <button
        @click="activeTab = 'PROMO'"
        class="px-4 py-2 text-xs font-bold rounded-lg transition-all"
        :class="activeTab === 'PROMO' ? 'bg-white text-indigo-600 shadow-sm' : 'text-slate-500 hover:text-slate-800'"
      >
        Program Sewa Display
      </button>
      <button
        @click="activeTab = 'SPG_EVENT'"
        class="px-4 py-2 text-xs font-bold rounded-lg transition-all"
        :class="activeTab === 'SPG_EVENT' ? 'bg-white text-indigo-600 shadow-sm' : 'text-slate-500 hover:text-slate-800'"
      >
        Jadwal SPG & Event Terpadu
      </button>
    </div>

    <!-- Filters & Promotion Data when activeTab is PROMO -->
    <div v-if="activeTab === 'PROMO'" class="space-y-6">
      <div class="bg-white rounded-2xl border border-slate-100 shadow-sm p-4 flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div class="relative w-full sm:w-[320px]">
          <Search class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400" />
          <input
            v-model="searchQuery"
            type="text"
            placeholder="Cari program promosi..."
            class="w-full pl-9 pr-4 py-2 text-xs border border-slate-200 rounded-xl focus:outline-none focus:ring-1 focus:ring-indigo-500"
          />
        </div>

        <div class="flex items-center gap-2">
          <select
            v-model="selectedType"
            class="px-3 py-2 text-xs border border-slate-200 rounded-xl focus:outline-none focus:ring-1 focus:ring-indigo-500 bg-white"
          >
            <option value="ALL">Semua Jenis Display</option>
            <option value="ENDCAP">Endcap</option>
            <option value="GONDOLA">Gondola Utama</option>
            <option value="FLOOR_DISPLAY">Floor Display</option>
          </select>
        </div>
      </div>

      <!-- Promotion Data Cards/Table -->
      <div class="bg-white rounded-2xl border border-slate-100 shadow-sm overflow-hidden">
        <div class="overflow-x-auto">
          <table class="w-full border-collapse text-left">
            <thead>
              <tr class="bg-slate-50/70 border-b border-slate-100 text-xs font-black text-slate-400 uppercase tracking-wider">
                <th class="py-4 px-6">Nama Program & Ritel</th>
                <th class="py-4 px-6">Tipe Sewa</th>
                <th class="py-4 px-6">Masa Promosi</th>
                <th class="py-4 px-6">Cakupan Gerai</th>
                <th class="py-4 px-6 text-right">Biaya Sewa</th>
                <th class="py-4 px-6 text-right">Target Sales</th>
                <th class="py-4 px-6 text-right">Penjualan Aktual</th>
                <th class="py-4 px-6 text-center">Status</th>
                <th class="py-4 px-6 text-center">Aksi</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-50 text-slate-700 text-sm">
              <tr v-for="promo in filteredPromotions" :key="promo.id" class="hover:bg-slate-50/30 transition-all">
                <td class="py-4 px-6">
                  <div>
                    <span class="font-extrabold text-slate-900 block">{{ promo.display_name }}</span>
                    <span class="text-xs text-indigo-600 font-bold uppercase tracking-wider">{{ promo.key_account }}</span>
                  </div>
                </td>
                <td class="py-4 px-6">
                  <span
                    class="px-2.5 py-0.5 text-[10px] font-extrabold rounded-full"
                    :class="{
                      'bg-blue-50 text-blue-600': promo.promo_type === 'ENDCAP',
                      'bg-emerald-50 text-emerald-600': promo.promo_type === 'GONDOLA',
                      'bg-purple-50 text-purple-600': promo.promo_type === 'FLOOR_DISPLAY',
                    }"
                  >
                    {{ promo.promo_type }}
                  </span>
                </td>
                <td class="py-4 px-6 text-xs font-medium">
                  <div class="flex items-center gap-1 text-slate-500">
                    <Calendar class="w-3.5 h-3.5" />
                    <span>{{ promo.start_date }} - {{ promo.end_date }}</span>
                  </div>
                </td>
                <td class="py-4 px-6 text-xs font-semibold text-slate-650">
                  <div class="flex items-center gap-1.5">
                    <MapPin class="w-3.5 h-3.5 text-slate-400" />
                    <span>{{ promo.outlet_coverage }}</span>
                  </div>
                </td>
                <td class="py-4 px-6 text-right font-black text-slate-800">
                  {{ formatCurrency(promo.cost) }}
                </td>
                <td class="py-4 px-6 text-right font-semibold text-slate-500">
                  {{ formatCurrency(promo.sales_target) }}
                </td>
                <td class="py-4 px-6 text-right font-black" :class="promo.sales_actual >= promo.sales_target ? 'text-emerald-600' : 'text-slate-855'">
                  {{ promo.sales_actual > 0 ? formatCurrency(promo.sales_actual) : '-' }}
                </td>
                <td class="py-4 px-6 text-center">
                  <span
                    class="px-2.5 py-1 text-[10px] font-extrabold uppercase rounded-full tracking-wider"
                    :class="{
                      'bg-indigo-50 text-indigo-600 border border-indigo-100': promo.status === 'ONGOING',
                      'bg-emerald-50 text-emerald-600 border border-emerald-100': promo.status === 'COMPLETED',
                      'bg-blue-50 text-blue-600 border border-blue-100': promo.status === 'APPROVED',
                    }"
                  >
                    {{ promo.status === 'ONGOING' ? 'Sedang Jalan' : promo.status === 'COMPLETED' ? 'Selesai' : 'Disetujui' }}
                  </span>
                </td>
                <td class="py-4 px-6 text-center">
                  <div class="flex items-center justify-center gap-2">
                    <button
                      @click="openEditPromoModal(promo)"
                      class="p-1.5 hover:bg-indigo-50 text-slate-400 hover:text-indigo-600 rounded-lg transition-colors"
                      title="Ubah Program"
                    >
                      <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
                      </svg>
                    </button>
                    <button
                      @click="deletePromo(promo.id)"
                      class="p-1.5 hover:bg-rose-50 text-slate-400 hover:text-rose-600 rounded-lg transition-colors"
                      title="Hapus Program"
                    >
                      <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                      </svg>
                    </button>
                  </div>
                </td>
              </tr>
              <tr v-if="filteredPromotions.length === 0">
                <td colspan="9" class="py-8 text-center text-slate-400">
                  Tidak ada data program promosi ditemukan.
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>

    <!-- SPG & Event lists when activeTab is SPG_EVENT -->
    <div v-else class="space-y-6">
      <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <!-- SPG Schedule Box -->
        <div class="bg-white rounded-2xl border border-slate-100 shadow-sm p-5 space-y-4">
          <div>
            <h3 class="text-sm font-black text-slate-900 uppercase tracking-wider">Jadwal SPG (Sales Promotion Girl)</h3>
            <p class="text-xs text-slate-400">Jadwal tugas SPG Karsa HR di outlet Modern Trade secara live.</p>
          </div>
          <div class="overflow-x-auto">
            <table class="w-full border-collapse text-left text-xs font-semibold">
              <thead>
                <tr class="bg-slate-50 border-b border-slate-100 text-[10px] text-slate-400 uppercase tracking-wider">
                  <th class="py-3 px-4">Nama SPG & Ritel</th>
                  <th class="py-3 px-4">Durasi Shift</th>
                  <th class="py-3 px-4">Lokasi Gerai</th>
                  <th class="py-3 px-4 text-center">Status</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-slate-50 text-slate-700">
                <tr v-for="spg in spgSchedules" :key="spg.id" class="hover:bg-slate-50/50">
                  <td class="py-3 px-4">
                    <span class="font-extrabold text-slate-800 block">{{ spg.spg_name }}</span>
                    <span class="text-[10px] text-indigo-600 block">{{ spg.retailer }}</span>
                  </td>
                  <td class="py-3 px-4 text-slate-500">
                    <span class="block">{{ spg.shift }}</span>
                    <span class="text-[10px] text-slate-400 block">{{ spg.start_date }} s/d {{ spg.end_date }}</span>
                  </td>
                  <td class="py-3 px-4 text-slate-600">{{ spg.location }}</td>
                  <td class="py-3 px-4 text-center">
                    <span
                      class="px-2 py-0.5 rounded-full text-[9px] font-black uppercase"
                      :class="spg.status === 'AKTIF' ? 'bg-emerald-50 text-emerald-600 border border-emerald-100' : 'bg-slate-55 text-slate-500 border border-slate-100'"
                    >
                      {{ spg.status }}
                    </span>
                  </td>
                </tr>
                <tr v-if="spgSchedules.length === 0">
                  <td colspan="4" class="py-4 text-center text-slate-400">Tidak ada jadwal SPG terdaftar.</td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>

        <!-- Events Schedule Box -->
        <div class="bg-white rounded-2xl border border-slate-100 shadow-sm p-5 space-y-4">
          <div>
            <h3 class="text-sm font-black text-slate-900 uppercase tracking-wider">Jadwal Event Promo MT</h3>
            <p class="text-xs text-slate-400">Jadwal aktivasi, branding, dan event weekend di ritel modern.</p>
          </div>
          <div class="overflow-x-auto">
            <table class="w-full border-collapse text-left text-xs font-semibold">
              <thead>
                <tr class="bg-slate-50 border-b border-slate-100 text-[10px] text-slate-400 uppercase tracking-wider">
                  <th class="py-3 px-4">Nama Event & Ritel</th>
                  <th class="py-3 px-4">Waktu Event</th>
                  <th class="py-3 px-4">PIC KARSA HR</th>
                  <th class="py-3 px-4 text-center">Status</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-slate-50 text-slate-700">
                <tr v-for="ev in eventSchedules" :key="ev.id" class="hover:bg-slate-50/50">
                  <td class="py-3 px-4">
                    <span class="font-extrabold text-slate-800 block leading-tight">{{ ev.event_name }}</span>
                    <span class="text-[10px] text-indigo-600 block">{{ ev.retailer }}</span>
                  </td>
                  <td class="py-3 px-4 text-slate-500">
                    <span class="block">{{ ev.start_date }} s/d {{ ev.end_date }}</span>
                  </td>
                  <td class="py-3 px-4 text-slate-650">{{ ev.pic }}</td>
                  <td class="py-3 px-4 text-center">
                    <span
                      class="px-2 py-0.5 rounded-full text-[9px] font-black uppercase"
                      :class="ev.status === 'TERJADWAL' ? 'bg-blue-50 text-blue-600 border border-blue-100' : 'bg-slate-55 text-slate-500 border'"
                    >
                      {{ ev.status }}
                    </span>
                  </td>
                </tr>
                <tr v-if="eventSchedules.length === 0">
                  <td colspan="4" class="py-4 text-center text-slate-400">Tidak ada event terdaftar.</td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </div>

    <!-- Promo Modal Form -->
    <div
      v-if="showPromoModal"
      class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-900/60 backdrop-blur-xs"
    >
      <div class="bg-white rounded-[24px] w-full max-w-lg shadow-2xl overflow-hidden p-6 relative">
        <h3 class="text-lg font-black text-slate-900 mb-4">
          {{ isEditing ? 'Edit Program Promosi' : 'Tambah Program Promosi' }}
        </h3>

        <div class="space-y-4 text-xs font-semibold text-slate-650">
          <div>
            <label class="block text-[10px] uppercase font-bold text-slate-400 mb-1">Ritel Modern (Key Account)</label>
            <select
              v-model="promoForm.key_account_id"
              class="w-full px-3 py-2 border border-slate-200 rounded-xl focus:ring-1 focus:ring-indigo-500 focus:outline-none bg-white"
            >
              <option v-for="ka in keyAccounts" :key="ka.id" :value="ka.id">
                {{ ka.name }} ({{ ka.company_name }})
              </option>
            </select>
          </div>

          <div>
            <label class="block text-[10px] uppercase font-bold text-slate-400 mb-1">Nama Program Sewa Display</label>
            <input
              v-model="promoForm.display_name"
              type="text"
              placeholder="Contoh: Sewa Gondola Utama Susu Karsa HR"
              class="w-full px-3 py-2 border border-slate-200 rounded-xl focus:ring-1 focus:ring-indigo-500 focus:outline-none"
            />
          </div>

          <div class="grid grid-cols-2 gap-4">
            <div>
              <label class="block text-[10px] uppercase font-bold text-slate-400 mb-1">Tipe Display</label>
              <select
                v-model="promoForm.promo_type"
                class="w-full px-3 py-2 border border-slate-200 rounded-xl focus:ring-1 focus:ring-indigo-500 focus:outline-none bg-white"
              >
                <option value="GONDOLA">Gondola Utama</option>
                <option value="ENDCAP">Endcap</option>
                <option value="FLOOR_DISPLAY">Floor Display</option>
              </select>
            </div>
            <div>
              <label class="block text-[10px] uppercase font-bold text-slate-400 mb-1">Cakupan Gerai</label>
              <input
                v-model="promoForm.outlet_coverage"
                type="text"
                placeholder="Contoh: 150 Outlet"
                class="w-full px-3 py-2 border border-slate-200 rounded-xl focus:ring-1 focus:ring-indigo-500 focus:outline-none"
              />
            </div>
          </div>

          <div class="grid grid-cols-2 gap-4">
            <div>
              <label class="block text-[10px] uppercase font-bold text-slate-400 mb-1">Tanggal Mulai</label>
              <input
                v-model="promoForm.start_date"
                type="date"
                class="w-full px-3 py-2 border border-slate-200 rounded-xl focus:ring-1 focus:ring-indigo-500 focus:outline-none"
              />
            </div>
            <div>
              <label class="block text-[10px] uppercase font-bold text-slate-400 mb-1">Tanggal Selesai</label>
              <input
                v-model="promoForm.end_date"
                type="date"
                class="w-full px-3 py-2 border border-slate-200 rounded-xl focus:ring-1 focus:ring-indigo-500 focus:outline-none"
              />
            </div>
          </div>

          <div class="grid grid-cols-3 gap-3">
            <div>
              <label class="block text-[10px] uppercase font-bold text-slate-400 mb-1">Biaya Sewa (Rp)</label>
              <input
                v-model.number="promoForm.cost"
                type="number"
                class="w-full px-3 py-2 border border-slate-200 rounded-xl focus:ring-1 focus:ring-indigo-500 focus:outline-none"
              />
            </div>
            <div>
              <label class="block text-[10px] uppercase font-bold text-slate-400 mb-1">Target Sales (Rp)</label>
              <input
                v-model.number="promoForm.sales_target"
                type="number"
                class="w-full px-3 py-2 border border-slate-200 rounded-xl focus:ring-1 focus:ring-indigo-500 focus:outline-none"
              />
            </div>
            <div>
              <label class="block text-[10px] uppercase font-bold text-slate-400 mb-1">Sales Aktual (Rp)</label>
              <input
                v-model.number="promoForm.sales_actual"
                type="number"
                class="w-full px-3 py-2 border border-slate-200 rounded-xl focus:ring-1 focus:ring-indigo-500 focus:outline-none"
              />
            </div>
          </div>

          <div>
            <label class="block text-[10px] uppercase font-bold text-slate-400 mb-1">Status Program</label>
            <select
              v-model="promoForm.status"
              class="w-full px-3 py-2 border border-slate-200 rounded-xl focus:ring-1 focus:ring-indigo-500 focus:outline-none bg-white"
            >
              <option value="APPROVED">Disetujui</option>
              <option value="ONGOING">Sedang Jalan</option>
              <option value="COMPLETED">Selesai</option>
            </select>
          </div>
        </div>

        <div class="mt-6 flex items-center justify-end gap-3">
          <Button
            @click="showPromoModal = false"
            variant="outline"
            class="px-4 py-2 border border-slate-200 rounded-xl text-xs text-slate-500 font-bold hover:bg-slate-50"
          >
            Batal
          </Button>
          <Button
            @click="savePromo"
            class="px-4 py-2 bg-indigo-600 hover:bg-indigo-700 text-white rounded-xl text-xs font-bold"
          >
            Simpan Program
          </Button>
        </div>
      </div>
    </div>
  </div>
</template>
