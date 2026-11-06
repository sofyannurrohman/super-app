<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import {
  Building2,
  Plus,
  Search,
  Phone,
  Mail,
  MapPin,
  Pencil,
  Trash2,
  FileText,
  DollarSign,
  Percent,
  Calendar,
  X,
  PlusCircle,
  Save,
  Building,
  Briefcase,
  CheckCircle2,
  TrendingUp,
} from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import { toast } from 'vue-sonner'
import apiClient from '@/api/axios'
import { useMasterDataStore } from '@/stores/masterData'

const masterDataStore = useMasterDataStore()

// Tabs
const activeTab = ref<'chains' | 'contracts'>('chains')

// Search & Filter
const searchQuery = ref('')
const selectedType = ref<'ALL' | 'NATIONAL' | 'REGIONAL'>('ALL')

// Real Database State connected to API
const keyAccounts = ref<any[]>([])
const contracts = ref<any[]>([])

// Modal Forms
const showAccountModal = ref(false)
const showContractModal = ref(false)
const isEditing = ref(false)

const accountForm = ref({
  id: '',
  name: '',
  company_name: '',
  type: 'NATIONAL' as 'NATIONAL' | 'REGIONAL',
  logo_color: 'bg-rose-700',
  hq_address: '',
  pic_name: '',
  pic_phone: '',
  pic_email: '',
  total_stores: 0,
  status: 'ACTIVE',
})

const contractForm = ref({
  id: '',
  key_account_id: '',
  start_date: '',
  end_date: '',
  payment_term_days: 30,
  rebate_percent: 0,
  marketing_contribution_percent: 0,
  listing_fee_per_sku: 0,
  listed_skus_count: 0,
  service_level_target: 95.0,
  notes: '',
})

// Fetch Data from Backend
const fetchKeyAccounts = async () => {
  try {
    const res = await apiClient.get('/modern-trade/key-accounts')
    keyAccounts.value = res.data.data || []
  } catch (err: any) {
    toast.error('Gagal mengambil data Key Account: ' + (err.response?.data?.error || err.message))
  }
}

const fetchContracts = async () => {
  try {
    const res = await apiClient.get('/modern-trade/contracts')
    contracts.value = (res.data.data || []).map((c: any) => ({
      ...c,
      key_account_name: c.key_account?.name || 'Ritel Modern',
      start_date: c.start_date ? c.start_date.split('T')[0] : '',
      end_date: c.end_date ? c.end_date.split('T')[0] : '',
    }))
  } catch (err: any) {
    toast.error('Gagal mengambil data kontrak: ' + (err.response?.data?.error || err.message))
  }
}

onMounted(() => {
  fetchKeyAccounts()
  fetchContracts()
})

// Filtered Accounts
const filteredAccounts = computed(() => {
  return keyAccounts.value.filter((acc) => {
    const matchesSearch =
      acc.name.toLowerCase().includes(searchQuery.value.toLowerCase()) ||
      acc.company_name.toLowerCase().includes(searchQuery.value.toLowerCase()) ||
      acc.pic_name.toLowerCase().includes(searchQuery.value.toLowerCase())
    const matchesType = selectedType.value === 'ALL' || acc.type === selectedType.value
    return matchesSearch && matchesType
  })
})

// Stats
const totalAccountsCount = computed(() => keyAccounts.value.length)
const activeContractsCount = computed(() => contracts.value.length)
const totalMTStores = computed(() =>
  keyAccounts.value.reduce((sum, item) => sum + (item.status === 'ACTIVE' ? item.total_stores : 0), 0)
)
const totalEstimatedListingFee = computed(() =>
  contracts.value.reduce((sum, item) => sum + (item.listing_fee_per_sku * item.listed_skus_count), 0)
)

// Actions - Accounts
const openNewAccountModal = () => {
  isEditing.value = false
  accountForm.value = {
    id: '',
    name: '',
    company_name: '',
    type: 'NATIONAL',
    logo_color: 'bg-rose-700',
    hq_address: '',
    pic_name: '',
    pic_phone: '',
    pic_email: '',
    total_stores: 0,
    status: 'ACTIVE',
  }
  showAccountModal.value = true
}

const openEditAccountModal = (account: any) => {
  isEditing.value = true
  accountForm.value = { ...account }
  showAccountModal.value = true
}

const saveAccount = async () => {
  if (!accountForm.value.name || !accountForm.value.company_name) {
    toast.error('Nama Ritel dan Perusahaan wajib diisi!')
    return
  }

  try {
    if (isEditing.value) {
      await apiClient.put(`/modern-trade/key-accounts/${accountForm.value.id}`, accountForm.value)
      toast.success('Key Account berhasil diperbarui!')
    } else {
      const colors = ['bg-purple-600', 'bg-pink-600', 'bg-cyan-600', 'bg-teal-600']
      const randomColor = colors[Math.floor(Math.random() * colors.length)]
      await apiClient.post('/modern-trade/key-accounts', {
        ...accountForm.value,
        logo_color: randomColor || 'bg-rose-700',
      })
      toast.success('Key Account baru berhasil ditambahkan!')
    }
    showAccountModal.value = false
    fetchKeyAccounts()
  } catch (err: any) {
    toast.error('Gagal menyimpan Key Account: ' + (err.response?.data?.error || err.message))
  }
}

const deleteAccount = async (id: string) => {
  if (confirm('Apakah Anda yakin ingin menghapus akun Modern Trade ini?')) {
    try {
      await apiClient.delete(`/modern-trade/key-accounts/${id}`)
      toast.success('Key Account berhasil dihapus!')
      fetchKeyAccounts()
      fetchContracts()
    } catch (err: any) {
      toast.error('Gagal menghapus Key Account: ' + (err.response?.data?.error || err.message))
    }
  }
}

// Actions - Contracts
const openNewContractModal = () => {
  isEditing.value = false
  const firstDate = new Date().toISOString().split('T')[0] || ''
  const nextDate = new Date(new Date().setFullYear(new Date().getFullYear() + 1)).toISOString().split('T')[0] || ''
  contractForm.value = {
    id: '',
    key_account_id: keyAccounts.value[0]?.id || '',
    start_date: firstDate,
    end_date: nextDate,
    payment_term_days: 30,
    rebate_percent: 0,
    marketing_contribution_percent: 0,
    listing_fee_per_sku: 0,
    listed_skus_count: 0,
    service_level_target: 95.0,
    notes: '',
  }
  showContractModal.value = true
}

const openEditContractModal = (contract: any) => {
  isEditing.value = true
  contractForm.value = { ...contract }
  showContractModal.value = true
}

const saveContract = async () => {
  const selectedKA = keyAccounts.value.find((acc) => acc.id === contractForm.value.key_account_id)
  if (!selectedKA) {
    toast.error('Pilih Key Account terlebih dahulu!')
    return
  }

  try {
    if (isEditing.value) {
      await apiClient.put(`/modern-trade/contracts/${contractForm.value.id}`, contractForm.value)
      toast.success('Kontrak Trading Terms diperbarui!')
    } else {
      await apiClient.post('/modern-trade/contracts', contractForm.value)
      toast.success('Kontrak baru berhasil terdaftar!')
    }
    showContractModal.value = false
    fetchKeyAccounts()
    fetchContracts()
  } catch (err: any) {
    toast.error('Gagal menyimpan kontrak: ' + (err.response?.data?.error || err.message))
  }
}

const deleteContract = async (id: string, kaId: string) => {
  if (confirm('Apakah Anda yakin ingin mematalkan/menghapus kontrak ini?')) {
    try {
      await apiClient.delete(`/modern-trade/contracts/${id}`)
      toast.success('Kontrak berhasil dihapus!')
      fetchKeyAccounts()
      fetchContracts()
    } catch (err: any) {
      toast.error('Gagal menghapus kontrak: ' + (err.response?.data?.error || err.message))
    }
  }
}

// Helpers
const formatCurrency = (val: number) => {
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    maximumFractionDigits: 0,
  }).format(val)
}
</script>

<template>
  <div class="space-y-6">
    <!-- Header Section -->
    <div class="flex flex-col md:flex-row md:items-center justify-between gap-4">
      <div>
        <div class="flex items-center gap-2">
          <span class="px-2.5 py-1 text-xs font-bold text-rose-700 bg-rose-50 border border-rose-200/50 rounded-full">
            Modern Trade (MT)
          </span>
        </div>
        <h1 class="text-2xl md:text-3xl font-black text-slate-900 tracking-tight mt-1 flex items-center flex-wrap gap-2">
          <span>Key Account Management (KAM)</span>
          <span v-if="masterDataStore.selectedBranchName" class="text-primary text-xl md:text-2xl font-bold">
            — {{ masterDataStore.selectedBranchName }}
          </span>
        </h1>
        <p class="text-sm text-slate-500 mt-1">
          Kelola hubungan ritel modern, ketentuan perdagangan (Trading Terms), dan kontrak utama produk Karsa HR.
        </p>
      </div>

      <div class="flex items-center gap-2.5">
        <Button
          v-if="activeTab === 'chains'"
          @click="openNewAccountModal"
          class="bg-rose-700 hover:bg-rose-800 text-white font-bold rounded-xl shadow-md shadow-rose-100"
        >
          <Plus class="w-4 h-4 mr-2" />
          Key Account Baru
        </Button>
        <Button
          v-else
          @click="openNewContractModal"
          class="bg-rose-700 hover:bg-rose-800 text-white font-bold rounded-xl shadow-md shadow-rose-100"
        >
          <Plus class="w-4 h-4 mr-2" />
          Kontrak Baru
        </Button>
      </div>
    </div>

    <!-- Stats Cards -->
    <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
      <!-- Stat 1 -->
      <div class="bg-white p-5 rounded-2xl border border-slate-100 shadow-sm flex items-center gap-4">
        <div class="p-3 bg-rose-50 text-rose-700 rounded-xl">
          <Building class="w-6 h-6" />
        </div>
        <div>
          <p class="text-xs font-bold text-slate-400 uppercase tracking-wider">Total Key Account</p>
          <p class="text-2xl font-black text-slate-800">{{ totalAccountsCount }} Akun Ritel</p>
        </div>
      </div>
      <!-- Stat 2 -->
      <div class="bg-white p-5 rounded-2xl border border-slate-100 shadow-sm flex items-center gap-4">
        <div class="p-3 bg-blue-50 text-blue-600 rounded-xl">
          <Building2 class="w-6 h-6" />
        </div>
        <div>
          <p class="text-xs font-bold text-slate-400 uppercase tracking-wider">Total Cabang Toko</p>
          <p class="text-2xl font-black text-slate-800">{{ totalMTStores.toLocaleString('id-ID') }} Outlet</p>
        </div>
      </div>
      <!-- Stat 3 -->
      <div class="bg-white p-5 rounded-2xl border border-slate-100 shadow-sm flex items-center gap-4">
        <div class="p-3 bg-emerald-50 text-emerald-600 rounded-xl">
          <FileText class="w-6 h-6" />
        </div>
        <div>
          <p class="text-xs font-bold text-slate-400 uppercase tracking-wider">Kontrak Aktif</p>
          <p class="text-2xl font-black text-slate-800">{{ activeContractsCount }} Trading Terms</p>
        </div>
      </div>
      <!-- Stat 4 -->
      <div class="bg-white p-5 rounded-2xl border border-slate-100 shadow-sm flex items-center gap-4">
        <div class="p-3 bg-amber-50 text-amber-600 rounded-xl">
          <DollarSign class="w-6 h-6" />
        </div>
        <div>
          <p class="text-xs font-bold text-slate-400 uppercase tracking-wider">Estimasi Listing Fee</p>
          <p class="text-xl font-black text-slate-800 truncate" :title="formatCurrency(totalEstimatedListingFee)">
            {{ formatCurrency(totalEstimatedListingFee) }}
          </p>
        </div>
      </div>
    </div>

    <!-- Navigation & Filter Tabs -->
    <div class="bg-white rounded-2xl border border-slate-100 shadow-sm p-4">
      <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <!-- Tab Selector -->
        <div class="flex bg-slate-100 p-1 rounded-xl w-fit">
          <button
            @click="activeTab = 'chains'"
            class="px-4 py-2 text-xs font-extrabold rounded-lg transition-all"
            :class="activeTab === 'chains' ? 'bg-white text-rose-700 shadow-sm' : 'text-slate-500 hover:text-slate-850'"
          >
            <span class="flex items-center gap-1.5">
              <Building class="w-3.5 h-3.5" />
              Key Account Ritel
            </span>
          </button>
          <button
            @click="activeTab = 'contracts'"
            class="px-4 py-2 text-xs font-extrabold rounded-lg transition-all"
            :class="activeTab === 'contracts' ? 'bg-white text-rose-700 shadow-sm' : 'text-slate-500 hover:text-slate-850'"
          >
            <span class="flex items-center gap-1.5">
              <FileText class="w-3.5 h-3.5" />
              Kontrak & Trading Terms
            </span>
          </button>
        </div>

        <!-- Search Bar and Filter (Only on Account Tab) -->
        <div v-if="activeTab === 'chains'" class="flex flex-wrap items-center gap-3">
          <div class="relative w-full sm:w-[240px]">
            <Search class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400" />
            <input
              v-model="searchQuery"
              type="text"
              placeholder="Cari ritel modern..."
              class="w-full pl-9 pr-4 py-2 text-xs border border-slate-200 rounded-xl focus:outline-none focus:ring-1 focus:ring-rose-500"
            />
          </div>
          <select
            v-model="selectedType"
            class="px-3 py-2 text-xs border border-slate-200 rounded-xl focus:outline-none focus:ring-1 focus:ring-rose-500"
          >
            <option value="ALL">Semua Cakupan</option>
            <option value="NATIONAL">Skala Nasional</option>
            <option value="REGIONAL">Skala Regional</option>
          </select>
        </div>
      </div>
    </div>

    <!-- Active Tab: Chains (Ritel Modern) -->
    <div v-if="activeTab === 'chains'" class="space-y-6">
      <div class="bg-white rounded-2xl border border-slate-100 shadow-sm overflow-hidden">
        <div class="overflow-x-auto">
          <table class="w-full border-collapse text-left">
            <thead>
              <tr class="bg-slate-50/70 border-b border-slate-100 text-xs font-black text-slate-400 uppercase tracking-wider">
                <th class="py-4 px-6">Nama Grup & PT</th>
                <th class="py-4 px-6 text-center">Cakupan</th>
                <th class="py-4 px-6">PIC Info</th>
                <th class="py-4 px-6 text-center">Cabang & Kontrak</th>
                <th class="py-4 px-6 text-center">Status</th>
                <th class="py-4 px-6 text-right">Aksi</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-50 text-slate-700 text-sm">
              <tr v-for="account in filteredAccounts" :key="account.id" class="hover:bg-slate-50/30 transition-all">
                <td class="py-4 px-6">
                  <div class="flex items-center gap-3">
                    <div class="h-10 w-10 rounded-xl flex items-center justify-center text-white font-black text-sm shadow-sm" :class="account.logo_color">
                      {{ account.name.substring(0, 2).toUpperCase() }}
                    </div>
                    <div>
                      <h3 class="font-extrabold text-slate-900 leading-tight">{{ account.name }}</h3>
                      <span class="text-xs text-slate-400 font-medium">{{ account.company_name }}</span>
                    </div>
                  </div>
                </td>
                <td class="py-4 px-6 text-center">
                  <span
                    class="px-2 py-0.5 text-[10px] font-bold rounded-full uppercase"
                    :class="account.type === 'NATIONAL' ? 'bg-rose-50 text-rose-700' : 'bg-orange-50 text-orange-600'"
                  >
                    {{ account.type }}
                  </span>
                </td>
                <td class="py-4 px-6">
                  <div class="space-y-1 text-xs text-slate-600">
                    <div class="flex items-center gap-1.5">
                      <Briefcase class="w-3.5 h-3.5 text-slate-400" />
                      <span class="font-bold">{{ account.pic_name }}</span>
                    </div>
                    <div class="flex items-center gap-1.5">
                      <Phone class="w-3.5 h-3.5 text-slate-400" />
                      <span>{{ account.pic_phone }}</span>
                    </div>
                    <div class="flex items-center gap-1.5">
                      <Mail class="w-3.5 h-3.5 text-slate-400" />
                      <span class="truncate">{{ account.pic_email }}</span>
                    </div>
                  </div>
                </td>
                <td class="py-4 px-6">
                  <div class="flex flex-col items-center gap-1">
                    <span class="text-xs font-black text-slate-700">{{ account.total_stores.toLocaleString('id-ID') }} Outlet</span>
                    <span class="text-[10px] font-bold text-slate-400">{{ account.active_contracts }} Kontrak Aktif</span>
                  </div>
                </td>
                <td class="py-4 px-6 text-center">
                  <span
                    class="inline-flex items-center justify-center gap-1 px-2.5 py-1 rounded-full text-[11px] font-extrabold"
                    :class="account.status === 'ACTIVE' ? 'bg-emerald-50 text-emerald-600' : 'bg-slate-100 text-slate-500'"
                  >
                    <span class="w-1.5 h-1.5 rounded-full" :class="account.status === 'ACTIVE' ? 'bg-emerald-500' : 'bg-slate-400'"></span>
                    {{ account.status === 'ACTIVE' ? 'Telah Bekerjasama' : 'Belum Berkontrak' }}
                  </span>
                </td>
                <td class="py-4 px-6 text-right">
                  <div class="flex items-center justify-end gap-1.5">
                    <button
                      @click="openEditAccountModal(account)"
                      class="p-2 text-slate-500 hover:text-rose-700 hover:bg-rose-50 rounded-lg transition-all"
                      title="Edit Akun"
                    >
                      <Pencil class="w-4 h-4" />
                    </button>
                    <button
                      @click="deleteAccount(account.id)"
                      class="p-2 text-slate-500 hover:text-destructive hover:bg-destructive/5 rounded-lg transition-all"
                      title="Hapus"
                    >
                      <Trash2 class="w-4 h-4" />
                    </button>
                  </div>
                </td>
              </tr>
              <tr v-if="filteredAccounts.length === 0">
                <td colspan="6" class="py-8 text-center text-slate-400">
                  Data ritel modern belum tersedia atau tidak ditemukan.
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>

    <!-- Active Tab: Contracts (Trading Terms) -->
    <div v-else class="space-y-6">
      <div class="bg-white rounded-2xl border border-slate-100 shadow-sm overflow-hidden">
        <div class="overflow-x-auto">
          <table class="w-full border-collapse text-left">
            <thead>
              <tr class="bg-slate-50/70 border-b border-slate-100 text-xs font-black text-slate-400 uppercase tracking-wider">
                <th class="py-4 px-6">Key Account</th>
                <th class="py-4 px-6">Masa Berlaku Contract</th>
                <th class="py-4 px-6 text-center">TOP (Tempo)</th>
                <th class="py-4 px-6 text-center">Rebate (%)</th>
                <th class="py-4 px-6 text-center">Marketing (%)</th>
                <th class="py-4 px-6 text-right">Listing Fee per SKU</th>
                <th class="py-4 px-6 text-center">SKU Terdaftar</th>
                <th class="py-4 px-6 text-center">Target SLA</th>
                <th class="py-4 px-6 text-right">Aksi</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-50 text-slate-700 text-sm">
              <tr v-for="contract in contracts" :key="contract.id" class="hover:bg-slate-50/30 transition-all">
                <td class="py-4 px-6">
                  <div class="flex items-center gap-2">
                    <span class="w-2.5 h-2.5 rounded-full bg-rose-500"></span>
                    <div>
                      <span class="font-extrabold text-slate-900">{{ contract.key_account_name }}</span>
                      <span class="block text-[10px] text-slate-400 font-bold uppercase">{{ contract.id }}</span>
                    </div>
                  </div>
                </td>
                <td class="py-4 px-6 font-medium text-xs whitespace-nowrap">
                  <div class="flex items-center gap-1.5 text-slate-600">
                    <Calendar class="w-3.5 h-3.5 text-slate-400" />
                    <span>{{ contract.start_date }} s/d {{ contract.end_date }}</span>
                  </div>
                </td>
                <td class="py-4 px-6 text-center font-bold text-slate-800">
                  {{ contract.payment_term_days }} Hari
                </td>
                <td class="py-4 px-6 text-center font-bold text-rose-700">
                  {{ contract.rebate_percent }}%
                </td>
                <td class="py-4 px-6 text-center font-medium">
                  {{ contract.marketing_contribution_percent }}%
                </td>
                <td class="py-4 px-6 text-right font-black text-slate-800">
                  {{ formatCurrency(contract.listing_fee_per_sku) }}
                </td>
                <td class="py-4 px-6 text-center">
                  <span class="px-2.5 py-0.5 text-xs font-bold text-slate-700 bg-slate-100 rounded-full">
                    {{ contract.listed_skus_count }} SKU
                  </span>
                </td>
                <td class="py-4 px-6 text-center">
                  <span class="px-2 py-0.5 text-xs font-bold text-emerald-600 bg-emerald-50 border border-emerald-100 rounded-full">
                    {{ contract.service_level_target }}%
                  </span>
                </td>
                <td class="py-4 px-6 text-right">
                  <div class="flex items-center justify-end gap-1.5">
                    <button
                      @click="openEditContractModal(contract)"
                      class="p-2 text-slate-500 hover:text-rose-700 hover:bg-rose-50 rounded-lg transition-all"
                      title="Edit"
                    >
                      <Pencil class="w-4 h-4" />
                    </button>
                    <button
                      @click="deleteContract(contract.id, contract.key_account_id)"
                      class="p-2 text-slate-500 hover:text-destructive hover:bg-destructive/5 rounded-lg transition-all"
                      title="Hapus"
                    >
                      <Trash2 class="w-4 h-4" />
                    </button>
                  </div>
                </td>
              </tr>
              <tr v-if="contracts.length === 0">
                <td colspan="9" class="py-8 text-center text-slate-400">
                  Belum ada kontrak kerjasama Modern Trade terdaftar.
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>

    <!-- MODAL: Add/Edit Key Account -->
    <div v-if="showAccountModal" class="fixed inset-0 z-[100] flex items-center justify-center p-4 bg-slate-900/60 backdrop-blur-sm">
      <div class="bg-white rounded-2xl w-full max-w-lg shadow-2xl overflow-hidden border border-slate-100">
        <div class="px-6 py-4 bg-slate-50 border-b border-slate-100 flex items-center justify-between">
          <h3 class="font-extrabold text-slate-800 text-lg">
            {{ isEditing ? 'Edit Key Account' : 'Registrasi Key Account Baru' }}
          </h3>
          <button @click="showAccountModal = false" class="p-1 rounded-lg hover:bg-slate-200 transition-all text-slate-400 hover:text-slate-600">
            <X class="w-5 h-5" />
          </button>
        </div>

        <div class="p-6 space-y-4 max-h-[75vh] overflow-y-auto">
          <!-- Chain Details -->
          <div class="grid grid-cols-2 gap-4">
            <div class="col-span-2">
              <label class="block text-xs font-bold text-slate-500 uppercase tracking-wider mb-1">Nama Grup Ritel / Modern Trade</label>
              <input
                v-model="accountForm.name"
                type="text"
                placeholder="Contoh: Indomaret Group"
                class="w-full px-3 py-2 text-sm border border-slate-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-rose-500/20 focus:border-rose-500"
              />
            </div>
            <div class="col-span-2">
              <label class="block text-xs font-bold text-slate-500 uppercase tracking-wider mb-1">Nama Legal PT Perusahaan</label>
              <input
                v-model="accountForm.company_name"
                type="text"
                placeholder="Contoh: PT Indomarco Prismatama"
                class="w-full px-3 py-2 text-sm border border-slate-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-rose-500/20 focus:border-rose-500"
              />
            </div>
            <div>
              <label class="block text-xs font-bold text-slate-500 uppercase tracking-wider mb-1">Cakupan Distribusi</label>
              <select
                v-model="accountForm.type"
                class="w-full px-3 py-2 text-sm border border-slate-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-rose-500/20 focus:border-rose-500"
              >
                <option value="NATIONAL">Nasional</option>
                <option value="REGIONAL">Regional</option>
              </select>
            </div>
            <div>
              <label class="block text-xs font-bold text-slate-500 uppercase tracking-wider mb-1">Jumlah Cabang Toko</label>
              <input
                v-model.number="accountForm.total_stores"
                type="number"
                placeholder="Jumlah toko"
                class="w-full px-3 py-2 text-sm border border-slate-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-rose-500/20 focus:border-rose-500"
              />
            </div>
          </div>

          <!-- HQ Address -->
          <div>
            <label class="block text-xs font-bold text-slate-500 uppercase tracking-wider mb-1">Alamat Head Office (Kantor Pusat)</label>
            <textarea
              v-model="accountForm.hq_address"
              rows="2"
              placeholder="Alamat kantor pusat..."
              class="w-full px-3 py-2 text-sm border border-slate-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-rose-500/20 focus:border-rose-500"
            ></textarea>
          </div>

          <!-- Contact Person Details -->
          <div class="border-t border-slate-100 pt-4 space-y-3">
            <h4 class="font-extrabold text-xs text-rose-700 uppercase tracking-wider">Contact Person (PIC KAM)</h4>
            <div class="grid grid-cols-2 gap-4">
              <div class="col-span-2">
                <label class="block text-[10px] font-bold text-slate-500 uppercase mb-1">Nama PIC</label>
                <input
                  v-model="accountForm.pic_name"
                  type="text"
                  placeholder="Nama lengkap PIC modern trade"
                  class="w-full px-3 py-2 text-sm border border-slate-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-rose-500/20 focus:border-rose-500"
                />
              </div>
              <div>
                <label class="block text-[10px] font-bold text-slate-500 uppercase mb-1">No. HP / WhatsApp</label>
                <input
                  v-model="accountForm.pic_phone"
                  type="text"
                  placeholder="0812..."
                  class="w-full px-3 py-2 text-sm border border-slate-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-rose-500/20 focus:border-rose-500"
                />
              </div>
              <div>
                <label class="block text-[10px] font-bold text-slate-500 uppercase mb-1">Email Ritel</label>
                <input
                  v-model="accountForm.pic_email"
                  type="email"
                  placeholder="pic@ritel.co.id"
                  class="w-full px-3 py-2 text-sm border border-slate-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-rose-500/20 focus:border-rose-500"
                />
              </div>
            </div>
          </div>
        </div>

        <div class="px-6 py-4 bg-slate-50 border-t border-slate-100 flex items-center justify-end gap-2.5">
          <Button variant="ghost" @click="showAccountModal = false" class="rounded-xl font-bold">
            Batal
          </Button>
          <Button @click="saveAccount" class="bg-rose-700 hover:bg-rose-800 text-white font-bold rounded-xl shadow-md">
            <Save class="w-4 h-4 mr-2" />
            Simpan Akun
          </Button>
        </div>
      </div>
    </div>

    <!-- MODAL: Add/Edit Contract -->
    <div v-if="showContractModal" class="fixed inset-0 z-[100] flex items-center justify-center p-4 bg-slate-900/60 backdrop-blur-sm">
      <div class="bg-white rounded-2xl w-full max-w-lg shadow-2xl overflow-hidden border border-slate-100">
        <div class="px-6 py-4 bg-slate-50 border-b border-slate-100 flex items-center justify-between">
          <h3 class="font-extrabold text-slate-800 text-lg">
            {{ isEditing ? 'Edit Kontrak Trading Terms' : 'Registrasi Kontrak Baru' }}
          </h3>
          <button @click="showContractModal = false" class="p-1 rounded-lg hover:bg-slate-200 transition-all text-slate-400 hover:text-slate-600">
            <X class="w-5 h-5" />
          </button>
        </div>

        <div class="p-6 space-y-4 max-h-[75vh] overflow-y-auto">
          <!-- Select Ritel -->
          <div>
            <label class="block text-xs font-bold text-slate-500 uppercase tracking-wider mb-1">Key Account Modern Trade</label>
            <select
              v-model="contractForm.key_account_id"
              :disabled="isEditing"
              class="w-full px-3 py-2 text-sm border border-slate-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-rose-500/20 focus:border-rose-500 disabled:bg-slate-100"
            >
              <option v-for="acc in keyAccounts" :key="acc.id" :value="acc.id">
                {{ acc.name }} ({{ acc.company_name }})
              </option>
            </select>
          </div>

          <!-- Dates & Terms -->
          <div class="grid grid-cols-2 gap-4">
            <div>
              <label class="block text-xs font-bold text-slate-500 uppercase tracking-wider mb-1">Mulai Kontrak</label>
              <input
                v-model="contractForm.start_date"
                type="date"
                class="w-full px-3 py-2 text-sm border border-slate-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-rose-500/20"
              />
            </div>
            <div>
              <label class="block text-xs font-bold text-slate-500 uppercase tracking-wider mb-1">Akhir Kontrak</label>
              <input
                v-model="contractForm.end_date"
                type="date"
                class="w-full px-3 py-2 text-sm border border-slate-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-rose-500/20"
              />
            </div>
            <div>
              <label class="block text-xs font-bold text-slate-500 uppercase tracking-wider mb-1">TOP (Hari Kerja)</label>
              <input
                v-model.number="contractForm.payment_term_days"
                type="number"
                placeholder="Contoh: 30"
                class="w-full px-3 py-2 text-sm border border-slate-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-rose-500/20"
              />
            </div>
            <div>
              <label class="block text-xs font-bold text-slate-500 uppercase tracking-wider mb-1">Target Service Level (%)</label>
              <input
                v-model.number="contractForm.service_level_target"
                type="number"
                step="0.1"
                placeholder="Contoh: 95.0"
                class="w-full px-3 py-2 text-sm border border-slate-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-rose-500/20"
              />
            </div>
          </div>

          <!-- Rates -->
          <div class="grid grid-cols-2 gap-4 border-t border-slate-100 pt-4">
            <div>
              <label class="block text-xs font-bold text-slate-500 uppercase tracking-wider mb-1">Rebate Tahunan (%)</label>
              <input
                v-model.number="contractForm.rebate_percent"
                type="number"
                step="0.01"
                placeholder="Contoh: 5.5"
                class="w-full px-3 py-2 text-sm border border-slate-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-rose-500/20"
              />
            </div>
            <div>
              <label class="block text-xs font-bold text-slate-500 uppercase tracking-wider mb-1">Marketing Contrib (%)</label>
              <input
                v-model.number="contractForm.marketing_contribution_percent"
                type="number"
                step="0.01"
                placeholder="Contoh: 2.0"
                class="w-full px-3 py-2 text-sm border border-slate-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-rose-500/20"
              />
            </div>
            <div>
              <label class="block text-xs font-bold text-slate-500 uppercase tracking-wider mb-1">Listing Fee per SKU (Rp)</label>
              <input
                v-model.number="contractForm.listing_fee_per_sku"
                type="number"
                placeholder="Contoh: 10000000"
                class="w-full px-3 py-2 text-sm border border-slate-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-rose-500/20"
              />
            </div>
            <div>
              <label class="block text-xs font-bold text-slate-500 uppercase tracking-wider mb-1">Jumlah SKU Terdaftar</label>
              <input
                v-model.number="contractForm.listed_skus_count"
                type="number"
                placeholder="Contoh: 8"
                class="w-full px-3 py-2 text-sm border border-slate-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-rose-500/20"
              />
            </div>
          </div>

          <!-- Notes -->
          <div>
            <label class="block text-xs font-bold text-slate-500 uppercase tracking-wider mb-1">Catatan Tambahan Kontrak</label>
            <textarea
              v-model="contractForm.notes"
              rows="3"
              placeholder="Catatan khusus, syarat retur, atau klaim rebate..."
              class="w-full px-3 py-2 text-sm border border-slate-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-rose-500/20 focus:border-rose-500"
            ></textarea>
          </div>
        </div>

        <div class="px-6 py-4 bg-slate-50 border-t border-slate-100 flex items-center justify-end gap-2.5">
          <Button variant="ghost" @click="showContractModal = false" class="rounded-xl font-bold">
            Batal
          </Button>
          <Button @click="saveContract" class="bg-rose-700 hover:bg-rose-800 text-white font-bold rounded-xl shadow-md">
            <Save class="w-4 h-4 mr-2" />
            Simpan Kontrak
          </Button>
        </div>
      </div>
    </div>
  </div>
</template>
