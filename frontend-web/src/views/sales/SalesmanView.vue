<script setup lang="ts">
import { ref, onMounted, computed, watch } from 'vue'
import apiClient from '@/api/axios'
import { toast } from 'vue-sonner'
import { useMasterDataStore } from '@/stores/masterData'
import {
  Search,
  Plus,
  MoreVertical,
  Edit2,
  Trash2,
  Building2,
  Calendar,
  History,
  Eye,
  ArrowRight,
  Users,
  MapPin,
  TrendingUp,
  X,
  Mail,
  Save,
  RefreshCw,
  UploadCloud,
  FileJson,
  CheckCircle2,
  AlertTriangle,
} from 'lucide-vue-next'
import { Button } from '@/components/ui/button'

const searchQuery = ref('')
const selectedPT = ref('')
const isModalOpen = ref(false)
const isLoading = ref(true)
const masterDataStore = useMasterDataStore()

const syncingSalesmanId = ref('')
const isSyncingAll = ref(false)
const isImportModalOpen = ref(false)
const backupJsonText = ref('')
const isImporting = ref(false)
const importResult = ref<any>(null)

const currentPage = ref(1)
const itemsPerPage = ref(10)
const totalPages = ref(1)
const totalItems = ref(0)

const salesmen = ref<any[]>([])
const allEmployees = ref<any[]>([])
const isEditing = ref(false)
const currentSalesman = ref({
  employeeID: '',
  name: '',
  targetOmzet: 0,
  targetNewStores: 0,
  targetDailyVisits: 0,
  targetDailyNewStores: 0,
  workingTerritory: '',
  receivableLimitPercent: 0,
  maxReceivableLimit: 0,
  commission: 0,
  month: new Date().getMonth() + 1,
  year: new Date().getFullYear(),
  companyID: '',
})

const showHistoryModal = ref(false)
const visitHistory = ref<any[]>([])
const selectedSalesman = ref<any>(null)
const isHistoryLoading = ref(false)

const currentMonth = ref(new Date().getMonth() + 1)
const currentYear = ref(new Date().getFullYear())

const months = [
  { value: 1, label: 'Januari' },
  { value: 2, label: 'Februari' },
  { value: 3, label: 'Maret' },
  { value: 4, label: 'April' },
  { value: 5, label: 'Mei' },
  { value: 6, label: 'Juni' },
  { value: 7, label: 'Juli' },
  { value: 8, label: 'Agustus' },
  { value: 9, label: 'September' },
  { value: 10, label: 'Oktober' },
  { value: 11, label: 'November' },
  { value: 12, label: 'Desember' },
]

const years = Array.from({ length: 5 }, (_, i) => new Date().getFullYear() - 2 + i)

const filteredSalesmen = computed(() => {
  return salesmen.value
})

const companies = computed(() => masterDataStore.companies)

watch([selectedPT, () => masterDataStore.selectedBranchId, () => masterDataStore.selectedSalesType], () => {
  currentPage.value = 1
  fetchSalesmen()
})

// Dynamic calculation for receivable limit & percentage
watch(() => currentSalesman.value.receivableLimitPercent, (newPercent) => {
  if (currentSalesman.value.targetOmzet > 0) {
    currentSalesman.value.maxReceivableLimit = Math.round(currentSalesman.value.targetOmzet * (newPercent / 100))
  }
})

watch(() => currentSalesman.value.targetOmzet, (newTarget) => {
  if (currentSalesman.value.receivableLimitPercent > 0) {
    currentSalesman.value.maxReceivableLimit = Math.round(newTarget * (currentSalesman.value.receivableLimitPercent / 100))
  }
})

const handleMaxReceivableLimitInput = (e: any) => {
  const val = Number(e.target.value)
  if (currentSalesman.value.targetOmzet > 0) {
    currentSalesman.value.receivableLimitPercent = Number(((val / currentSalesman.value.targetOmzet) * 100).toFixed(2))
  }
}

const handleSearch = () => {
  currentPage.value = 1
  fetchSalesmen()
}

const nextPage = () => {
  if (currentPage.value < totalPages.value) {
    currentPage.value++
    fetchSalesmen()
  }
}

const prevPage = () => {
  if (currentPage.value > 1) {
    currentPage.value--
    fetchSalesmen()
  }
}

const fetchSalesmen = async () => {
  isLoading.value = true
  const month = currentMonth.value
  const year = currentYear.value

  // Calculate previous month for carry-over targets
  let prevMonth = month - 1
  let prevYear = year
  if (prevMonth === 0) {
    prevMonth = 12
    prevYear = year - 1
  }

  try {
    const branchId = masterDataStore.selectedBranchId || ''
    const branchFilter = branchId ? `&branch_id=${branchId}` : ''

    // 1. Fetch Performance Data (Current & Previous)
    const [perfRes, prevPerfRes, kasbonRes] = await Promise.all([
      apiClient.get(`/admin/sales/reports/performance?month=${month}&year=${year}${branchFilter}`),
      apiClient.get(
        `/admin/sales/reports/performance?month=${prevMonth}&year=${prevYear}${branchFilter}`,
      ),
      apiClient.get(`/admin/sales/kasbon`),
    ])

    const perfData = perfRes.data?.data || []
    const prevPerfData = prevPerfRes.data?.data || []
    const kasbonData = kasbonRes.data?.data || []

    const companyId = selectedPT.value || masterDataStore.selectedBranchCompanyId || ''
    const companyFilter = companyId ? `&company_id=${companyId}` : ''

    // 2. Fetch Paginated Salesmen Employees
    const empRes = await apiClient.get(
      `/employees?is_sales=true&page=${currentPage.value}&limit=${itemsPerPage.value}&search=${searchQuery.value}${companyFilter}${branchFilter}&sales_type=${masterDataStore.selectedSalesType}`,
    )
    if (empRes.data?.data) {
      allEmployees.value = empRes.data.data
      totalItems.value = empRes.data.meta?.total || 0
      totalPages.value = Math.ceil(totalItems.value / itemsPerPage.value) || 1
    } else {
      allEmployees.value = []
      totalItems.value = 0
      totalPages.value = 1
    }

    const salesmenEmployees = allEmployees.value

    // 4. Merge Performance with Salesmen List
    salesmen.value = salesmenEmployees.map((emp) => {
      const kpi = perfData.find((p: any) => p.EmployeeID === emp.id)
      const prevKpi = prevPerfData.find((p: any) => p.EmployeeID === emp.id)

      return {
        id: kpi?.ID || emp.id,
        kpiID: kpi?.ID || null,
        hasCurrentKpi: !!kpi,
        employeeID: emp.id,
        name: `${emp.first_name} ${emp.last_name || ''}`.trim(),
        email: emp.user?.email || '-',
        phone: emp.phone_number || '-',
        area: emp.branch?.name || 'Kantor Pusat',
        // Carry over target and territory if current kpi is missing
        target: kpi?.TargetOmzet || prevKpi?.TargetOmzet || 0,
        targetNewStores: kpi?.TargetNewStores || prevKpi?.TargetNewStores || 0,
        targetDailyVisits: kpi?.target_daily_visits || prevKpi?.target_daily_visits || 0,
        targetDailyNewStores: kpi?.target_daily_new_stores || prevKpi?.target_daily_new_stores || 0,
        workingTerritory:
          kpi?.working_territory || prevKpi?.working_territory || emp.working_territory || '-',
        receivableLimitPercent: kpi?.receivable_limit_percent || 0,
        maxReceivableLimit: kpi?.max_receivable_limit || 0,
        commission: kpi?.commission || 0,
        branch: emp.branch?.name || '-',
        company: emp.company?.name || '-',
        companyID: emp.company_id || '',
        // Current month achievements (always reset if kpi is missing)
        current: kpi?.AchievedOmzet || 0,
        currentLama: kpi?.AchievedOmzetLama || 0,
        currentBaru: kpi?.AchievedOmzetBaru || 0,
        totalVisits: kpi?.TotalVisits || 0,
        totalKasbon: kasbonData.find((k: any) => k.employee_id === emp.id)?.total_kasbon || 0,
        totalOverdueInvoices:
          kasbonData.find((k: any) => k.employee_id === emp.id)?.total_invoices || 0,
      }
    })
  } catch (error) {
    console.error('Failed to fetch salesmen:', error)
  } finally {
    isLoading.value = false
  }
}

const fetchEmployees = async () => {
  try {
    const res = await apiClient.get('/employees?limit=all')
    if (res.data?.data) {
      allEmployees.value = res.data.data
    }
  } catch (error) {
    console.error('Failed to fetch employees:', error)
  }
}

const handleEditTarget = (s: any) => {
  isEditing.value = true
  const rawData = salesmen.value.find((item) => item.id === s.id)
  if (rawData) {
    currentSalesman.value = {
      employeeID: rawData.employeeID,
      name: s.name,
      targetOmzet: s.target,
      targetNewStores: rawData.targetNewStores || 0,
      targetDailyVisits: rawData.targetDailyVisits || 0,
      targetDailyNewStores: rawData.targetDailyNewStores || 0,
      workingTerritory: rawData.workingTerritory || '',
      receivableLimitPercent: rawData.receivableLimitPercent || 0,
      maxReceivableLimit: rawData.maxReceivableLimit || 0,
      commission: rawData.commission || 0,
      month: currentMonth.value,
      year: currentYear.value,
      companyID: rawData.companyID || '',
    }
    isModalOpen.value = true
  } else {
    toast.error('Data salesman tidak ditemukan')
  }
}

const handleSaveTarget = async () => {
  try {
    // Update PT Karyawan
    if (currentSalesman.value.employeeID && currentSalesman.value.companyID) {
      const emp = allEmployees.value.find((e: any) => e.id === currentSalesman.value.employeeID)
      if (emp && emp.company_id !== currentSalesman.value.companyID) {
        const payload = {
          id: emp.id,
          name: emp.first_name || '',
          email: emp.user?.email || '',
          employeeIDNumber: emp.employee_id_number || '',
          departmentId: emp.department_id || null,
          jobPositionId: emp.job_position_id || null,
          branchId: emp.branch_id || null,
          companyId: currentSalesman.value.companyID || null,
          employmentStatus: emp.employment_status || 'Active',
          joinDate: emp.join_date ? new Date(emp.join_date).toISOString().split('T')[0] : '',
          bankName: emp.bank_name || '',
          bankAccountNumber: emp.bank_account_number || '',
          accountHolderName: emp.account_holder_name || '',
          salary: emp.salary || 0,
          phoneNumber: emp.phone_number || '',
          birthPlace: emp.birth_place || '',
          birthDate: emp.birth_date ? new Date(emp.birth_date).toISOString().split('T')[0] : '',
          gender: emp.gender || 'MALE',
          maritalStatus: emp.marital_status || 'SINGLE',
          bpjsKesehatanNumber: emp.bpjs_kesehatan_number || '',
          bpjsKetenagakerjaanNumber: emp.bpjs_ketenagakerjaan_number || '',
          identityNumber: emp.identity_number || '',
          npwpNumber: emp.npwp_number || '',
          lastName: emp.last_name || '',
          addressKTP: emp.address_ktp || '',
          addressResidential: emp.address_residential || '',
          emergencyContact: emp.emergency_contact || '',
          ptkpStatus: emp.ptkp_status || 'TK/0',
          managerId: emp.manager_id || null,
          defaultShiftId: emp.default_shift_id || null,
        }
        await apiClient.put(`/employees/${emp.id}`, payload)
      }
    }

    await apiClient.post('/admin/sales/targets', {
      employee_id: currentSalesman.value.employeeID || null,
      month: currentSalesman.value.month,
      year: currentSalesman.value.year,
      target_omzet: currentSalesman.value.targetOmzet,
      target_new_stores: currentSalesman.value.targetNewStores,
      target_daily_visits: currentSalesman.value.targetDailyVisits,
      target_daily_new_stores: currentSalesman.value.targetDailyNewStores,
      working_territory: currentSalesman.value.workingTerritory,
      receivable_limit_percent: currentSalesman.value.receivableLimitPercent,
      max_receivable_limit: currentSalesman.value.maxReceivableLimit,
      commission: currentSalesman.value.commission,
    })
    toast.success('Target & Limit Piutang berhasil diperbarui!')
    isModalOpen.value = false
    fetchSalesmen()
  } catch (error) {
    console.error('Failed to save target:', error)
    toast.error('Gagal menyimpan target')
  }
}

const handleDeleteTarget = async (s: any) => {
  if (!s.kpiID) {
    toast.error('Data target tidak ditemukan untuk bulan ini')
    return
  }

  if (
    !confirm(
      `Hapus data target & performa ${s.name} untuk bulan ${getMonthLabel(currentMonth.value)}?`,
    )
  )
    return

  try {
    await apiClient.delete(`/admin/sales/targets/${s.kpiID}`)
    toast.success('Target berhasil dihapus!')
    fetchSalesmen()
  } catch (error: any) {
    toast.error('Gagal menghapus target: ' + (error.response?.data?.error || error.message))
  }
}

const openVisitHistory = async (s: any) => {
  selectedSalesman.value = s
  showHistoryModal.value = true
  isHistoryLoading.value = true
  try {
    const res = await apiClient.get(`/admin/sales/visits?employee_id=${s.employeeID}&limit=50`)
    if (res.data?.data) {
      visitHistory.value = res.data.data
    }
  } catch (error) {
    console.error('Failed to fetch visit history:', error)
    toast.error('Gagal mengambil riwayat kunjungan')
  } finally {
    isHistoryLoading.value = false
  }
}

const dailyVisitKPI = computed(() => {
  if (visitHistory.value.length === 0) return 0
  const isCurrentMonth =
    currentMonth.value === new Date().getMonth() + 1 &&
    currentYear.value === new Date().getFullYear()
  const days = isCurrentMonth
    ? new Date().getDate()
    : new Date(currentYear.value, currentMonth.value, 0).getDate()
  return (visitHistory.value.length / days).toFixed(1)
})

const openAddSalesman = () => {
  isEditing.value = false
  currentSalesman.value = {
    employeeID: '',
    name: '',
    targetOmzet: 0,
    targetNewStores: 0,
    targetDailyVisits: 0,
    targetDailyNewStores: 0,
    workingTerritory: '',
    receivableLimitPercent: 0,
    maxReceivableLimit: 0,
    commission: 0,
    month: currentMonth.value,
    year: currentYear.value,
    companyID: '',
  }
  isModalOpen.value = true
}

onMounted(() => {
  masterDataStore.fetchBranches()
  masterDataStore.fetchCompanies()
  fetchSalesmen()
  fetchEmployees()
})

const triggerRemoteSync = async (salesman: any) => {
  syncingSalesmanId.value = salesman.employeeID
  try {
    const res = await apiClient.post(`/admin/sales/remote-sync/${salesman.employeeID}`)
    toast.success(res.data?.message || `Sinyal tarik data dikirim ke ${salesman.name}`)
  } catch (err: any) {
    toast.error(err.response?.data?.error || 'Gagal mengirim sinyal remote sync')
  } finally {
    syncingSalesmanId.value = ''
  }
}

const triggerRemoteSyncAll = async () => {
  isSyncingAll.value = true
  try {
    const res = await apiClient.post('/admin/sales/remote-sync-all')
    toast.success(res.data?.message || 'Sinyal tarik data dikirim ke seluruh sales aktif!')
  } catch (err: any) {
    toast.error(err.response?.data?.error || 'Gagal mengirim broadcast remote sync')
  } finally {
    isSyncingAll.value = false
  }
}

const handleImportBackup = async () => {
  if (!backupJsonText.value.trim()) {
    toast.error('Silakan tempelkan data cadangan JSON terlebih dahulu')
    return
  }

  let parsed: any
  try {
    parsed = JSON.parse(backupJsonText.value.trim())
  } catch (e) {
    toast.error('Format JSON tidak valid. Pastikan data cadangan disalin secara utuh.')
    return
  }

  isImporting.value = true
  importResult.value = null
  try {
    const payload = parsed.transactions
      ? parsed
      : { transactions: Array.isArray(parsed) ? parsed : [parsed] }
    const res = await apiClient.post('/admin/sales/import-offline-backup', payload)
    importResult.value = res.data?.data
    toast.success(res.data?.message || 'Impor data cadangan offline berhasil diproses!')
    fetchSalesmen()
  } catch (err: any) {
    toast.error(err.response?.data?.error || 'Gagal mengimpor data cadangan offline')
  } finally {
    isImporting.value = false
  }
}

const formatCurrency = (val: number) => {
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    maximumFractionDigits: 0,
  }).format(val || 0)
}

const getMonthLabel = (m: number) => {
  const found = months.find((item) => item.value === m)
  return found ? found.label : ''
}
</script>

<template>
  <div class="space-y-8 pb-10">
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-6">
      <div>
        <h1 class="text-3xl font-extrabold text-slate-900 tracking-tight flex items-center flex-wrap gap-2">
          <span>Data Salesman</span>
          <span v-if="masterDataStore.selectedBranchName" class="text-primary text-2xl font-bold">
            — {{ masterDataStore.selectedBranchName }}
          </span>
        </h1>
        <p class="text-[15px] font-medium text-slate-500 mt-1.5 flex items-center gap-2">
          <Users class="w-4 h-4 text-primary" />
          Manajemen tim lapangan dan monitoring target omzet mereka.
        </p>
      </div>
      <div class="flex flex-wrap items-center gap-3">
        <Button
          @click="triggerRemoteSyncAll"
          :disabled="isSyncingAll"
          variant="outline"
          class="border-primary text-primary hover:bg-primary/10 px-5 py-6 rounded-2xl shadow-sm flex items-center gap-2 font-bold"
        >
          <RefreshCw class="w-5 h-5" :class="{ 'animate-spin': isSyncingAll }" />
          <span>{{ isSyncingAll ? 'Mengirim Sinyal...' : 'Tarik Data Semua Sales' }}</span>
        </Button>
        <Button
          @click="isImportModalOpen = true"
          variant="outline"
          class="border-slate-300 text-slate-700 hover:bg-slate-100 px-5 py-6 rounded-2xl shadow-sm flex items-center gap-2 font-bold"
        >
          <UploadCloud class="w-5 h-5 text-indigo-600" />
          <span>Impor Cadangan Offline</span>
        </Button>
        <Button
          @click="openAddSalesman"
          class="bg-primary hover:bg-primary/90 text-white px-6 py-6 rounded-2xl shadow-lg shadow-primary/20 flex items-center gap-2 font-bold"
        >
          <Plus class="w-5 h-5" /> Tambah Salesman
        </Button>
      </div>
    </div>

    <!-- Filters & Search -->
    <div
      class="bg-white p-4 rounded-3xl border border-slate-200 shadow-sm flex flex-col md:flex-row gap-4 items-center"
    >
      <div class="relative flex-1 w-full max-w-[200px]">
        <select
          v-model="selectedPT"
          class="w-full bg-slate-50 border-none rounded-2xl px-5 py-3.5 font-bold text-slate-600 text-sm focus:ring-2 focus:ring-primary/20 appearance-none"
        >
          <option value="">Semua Perusahaan</option>
          <option v-for="pt in companies" :key="pt.id" :value="pt.id">{{ pt.name }}</option>
        </select>
      </div>
      <div class="relative flex-1 w-full">
        <Search class="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-slate-400" />
        <input
          v-model="searchQuery"
          @keyup.enter="handleSearch"
          type="text"
          placeholder="Cari salesman... (Tekan Enter)"
          class="w-full pl-12 pr-4 py-3.5 bg-slate-50 border-none rounded-2xl focus:ring-2 focus:ring-primary/20 transition-all font-medium text-slate-900"
        />
      </div>
      <div class="flex gap-2 w-full md:w-auto">
        <select
          v-model="currentMonth"
          @change="fetchSalesmen"
          class="bg-slate-50 border-none rounded-2xl px-6 py-3.5 font-bold text-slate-600 text-sm focus:ring-2 focus:ring-primary/20"
        >
          <option v-for="m in months" :key="m.value" :value="m.value">{{ m.label }}</option>
        </select>
        <select
          v-model="currentYear"
          @change="fetchSalesmen"
          class="bg-slate-50 border-none rounded-2xl px-6 py-3.5 font-bold text-slate-600 text-sm focus:ring-2 focus:ring-primary/20"
        >
          <option v-for="y in years" :key="y" :value="y">{{ y }}</option>
        </select>
      </div>
    </div>

    <!-- Table -->
    <div class="bg-white rounded-3xl border border-slate-200 shadow-sm overflow-x-auto">
      <table class="w-full text-left border-collapse min-w-[1000px]">
        <thead>
          <tr
            class="bg-slate-50/50 uppercase text-[10px] font-extrabold text-slate-400 tracking-[0.1em]"
          >
            <th class="px-8 py-5">Identitas</th>
            <th class="px-8 py-5">Perusahaan (PT)</th>
            <th class="px-8 py-5">Kantor Cabang</th>
            <th class="px-8 py-5">Wilayah Kerja</th>
            <th class="px-8 py-5">Target Harian</th>
            <th class="px-8 py-5">Performa Omzet</th>
            <th class="px-8 py-5">Total Kasbon (Overdue)</th>
            <th class="px-8 py-5">Limit Piutang</th>
            <th class="px-8 py-5">Total Kunjungan</th>
            <th class="px-8 py-5">Total Realisasi</th>
            <th class="px-8 py-5 text-center">Aksi</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-slate-100">
          <tr
            v-for="s in filteredSalesmen"
            :key="s.id"
            class="group hover:bg-slate-50/70 transition-all"
          >
            <td class="px-8 py-6">
              <div>
                <p class="font-extrabold text-slate-900 group-hover:text-primary transition-colors">
                  {{ s.name }}
                </p>
                <p class="text-[12px] text-slate-400 font-bold flex items-center gap-1.5 mt-0.5">
                  <Mail class="w-3 h-3" /> {{ s.email }}
                </p>
              </div>
            </td>
            <td class="px-8 py-6">
              <div class="flex items-center gap-2 text-sm font-bold text-slate-600">
                <Building2 class="w-4 h-4 text-slate-400" />
                {{ s.company }}
              </div>
            </td>
            <td class="px-8 py-6">
              <div class="flex items-center gap-2 text-sm font-bold text-slate-600">
                <Building2 class="w-4 h-4 text-slate-400" />
                {{ s.branch }}
              </div>
            </td>
            <td class="px-8 py-6">
              <div class="flex items-center gap-2 text-sm font-bold text-blue-600">
                <MapPin class="w-4 h-4 text-blue-400" />
                {{ s.workingTerritory }}
              </div>
            </td>
            <td class="px-8 py-6">
              <div class="space-y-1.5">
                <p class="text-xs font-bold text-slate-600">
                  Kunjungan:
                  <span class="text-slate-900 font-black">{{ s.targetDailyVisits }}</span>
                </p>
                <p class="text-xs font-bold text-slate-600">
                  Toko Baru:
                  <span class="text-slate-900 font-black">{{ s.targetDailyNewStores }}</span>
                </p>
              </div>
            </td>
            <td class="px-8 py-6">
              <div class="space-y-1.5">
                <p class="text-[10px] font-black text-slate-400 uppercase tracking-widest">
                  Target: <span class="text-slate-900">{{ formatCurrency(s.target) }}</span>
                </p>
                <p v-if="s.commission > 0" class="text-[10px] font-black text-amber-500 uppercase tracking-widest mt-0.5">
                  Komisi: <span class="text-amber-600">{{ formatCurrency(s.commission) }}</span>
                </p>
                <div class="flex flex-col gap-0.5">
                  <p class="text-xs font-bold text-slate-600">
                    Lama: {{ formatCurrency(s.currentLama) }}
                  </p>
                  <p class="text-xs font-bold text-blue-600">
                    Baru: {{ formatCurrency(s.currentBaru) }}
                  </p>
                </div>
              </div>
            </td>
            <td class="px-8 py-6">
              <div v-if="s.totalKasbon > 0">
                <p class="text-sm font-black text-red-600">{{ formatCurrency(s.totalKasbon) }}</p>
                <p class="text-[10px] font-bold text-red-400 uppercase tracking-tighter mt-1">
                  {{ s.totalOverdueInvoices }} Tagihan
                </p>
              </div>
              <div v-else>
                <span
                  class="px-3 py-1 bg-emerald-50 text-emerald-600 rounded-lg text-xs font-bold border border-emerald-100"
                  >Bersih</span
                >
              </div>
            </td>
            <td class="px-8 py-6">
              <div v-if="s.maxReceivableLimit > 0">
                <p class="text-sm font-black text-indigo-600">{{ formatCurrency(s.maxReceivableLimit) }}</p>
                <p class="text-[10px] font-bold text-slate-400 mt-1">
                  {{ s.receivableLimitPercent }}% dari Target
                </p>
              </div>
              <div v-else>
                <span
                  class="px-3 py-1 bg-slate-50 text-slate-400 rounded-lg text-xs font-bold border border-slate-100"
                  >Tanpa Limit</span
                >
              </div>
            </td>
            <td class="px-8 py-6">
              <button
                @click="openVisitHistory(s)"
                class="flex items-center gap-2 px-4 py-2 bg-emerald-50 text-emerald-700 rounded-xl border border-emerald-100 hover:bg-emerald-100 transition-all group/visit"
              >
                <span class="text-sm font-black">{{ s.totalVisits }}</span>
                <span class="text-[10px] font-bold uppercase tracking-tighter opacity-70"
                  >Kunjungan</span
                >
                <Eye
                  class="w-3.5 h-3.5 opacity-0 group-hover/visit:opacity-100 transition-all translate-x-[-4px] group-hover/visit:translate-x-0"
                />
              </button>
            </td>
            <td class="px-8 py-6">
              <div class="space-y-2">
                <div class="flex items-center justify-between">
                  <span
                    class="text-[13px] font-black"
                    :class="s.current >= s.target ? 'text-emerald-600' : 'text-primary'"
                  >
                    {{ formatCurrency(s.current) }}
                  </span>
                  <span class="text-[10px] font-black text-slate-400">
                    {{ s.target > 0 ? Math.round((s.current / s.target) * 100) : 0 }}%
                  </span>
                </div>
                <div class="h-2 w-full bg-slate-100 rounded-full overflow-hidden">
                  <div
                    class="h-full transition-all duration-1000 rounded-full"
                    :class="s.current >= s.target ? 'bg-emerald-500' : 'bg-primary'"
                    :style="`width: ${s.target > 0 ? Math.min((s.current / s.target) * 100, 100) : 0}%`"
                  ></div>
                </div>
              </div>
            </td>
            <td class="px-8 py-6 text-center">
              <div class="flex items-center justify-center gap-2">
                <Button
                  @click="triggerRemoteSync(s)"
                  :disabled="syncingSalesmanId === s.employeeID"
                  variant="outline"
                  size="sm"
                  title="Tarik Data Offline (Kirim Sinyal Sync ke HP Sales)"
                  class="h-8 px-3 rounded-lg border-indigo-100 text-indigo-600 hover:bg-indigo-600 hover:text-white transition-all font-bold gap-1.5"
                >
                  <RefreshCw class="w-3.5 h-3.5" :class="{ 'animate-spin': syncingSalesmanId === s.employeeID }" />
                </Button>
                <Button
                  @click="handleEditTarget(s)"
                  variant="outline"
                  size="sm"
                  class="h-8 px-3 rounded-lg border-slate-200 text-slate-600 hover:bg-primary hover:text-white transition-all font-bold gap-1.5"
                >
                  <Edit2 class="w-3.5 h-3.5" />
                </Button>
                <Button
                  @click="handleDeleteTarget(s)"
                  variant="outline"
                  size="sm"
                  class="h-8 px-3 rounded-lg border-red-100 text-red-500 hover:bg-red-500 hover:text-white transition-all font-bold gap-1.5"
                >
                  <Trash2 class="w-3.5 h-3.5" />
                </Button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Pagination Controls -->
    <div
      v-if="totalPages > 1"
      class="p-6 border border-slate-200 mt-6 rounded-3xl flex flex-col sm:flex-row items-center justify-between bg-white shadow-sm gap-4"
    >
      <p class="text-xs font-bold text-slate-400">
        Menampilkan <span class="text-slate-700">{{ (currentPage - 1) * itemsPerPage + 1 }}</span> -
        <span class="text-slate-700">{{ Math.min(currentPage * itemsPerPage, totalItems) }}</span>
        dari <span class="text-slate-700">{{ totalItems }}</span> total data
      </p>
      <div class="flex items-center gap-2">
        <Button
          @click="prevPage"
          :disabled="currentPage === 1"
          class="bg-white border-2 border-slate-200 text-slate-600 hover:bg-slate-100 hover:border-slate-300 rounded-xl px-4 py-2 text-xs font-black disabled:opacity-50 transition-all"
        >
          SEBELUMNYA
        </Button>
        <div
          class="flex items-center justify-center min-w-[2.5rem] px-2 h-10 rounded-xl bg-primary/10 text-primary font-black text-sm"
        >
          {{ currentPage }} / {{ totalPages }}
        </div>
        <Button
          @click="nextPage"
          :disabled="currentPage === totalPages"
          class="bg-white border-2 border-slate-200 text-slate-600 hover:bg-slate-100 hover:border-slate-300 rounded-xl px-4 py-2 text-xs font-black disabled:opacity-50 transition-all"
        >
          SELANJUTNYA
        </Button>
      </div>
    </div>

    <!-- Modal CRUD Target -->
    <div v-if="isModalOpen" class="fixed inset-0 z-50 flex items-center justify-center p-4">
      <div
        class="absolute inset-0 bg-slate-900/60 backdrop-blur-sm"
        @click="isModalOpen = false"
      ></div>
      <div
        class="bg-white rounded-[40px] w-full max-w-lg relative z-10 shadow-2xl flex flex-col max-h-[90vh] overflow-hidden animate-in fade-in zoom-in duration-300"
      >
        <div
          class="p-8 border-b border-slate-100 flex items-center justify-between bg-slate-50/50 shrink-0"
        >
          <div>
            <h3 class="text-2xl font-black text-slate-900 tracking-tight">
              {{ isEditing ? 'Edit Target' : 'Tambah Salesman' }}
            </h3>
            <p class="text-sm font-bold text-slate-400 uppercase tracking-widest mt-1">
              Manajemen Performa
            </p>
          </div>
          <button
            @click="isModalOpen = false"
            class="p-3 hover:bg-white rounded-2xl transition-all text-slate-400 hover:text-slate-900 shadow-sm"
          >
            <Edit2 class="w-6 h-6 rotate-45" />
          </button>
        </div>

        <div class="p-8 space-y-6 overflow-y-auto">
          <div v-if="!isEditing" class="space-y-2">
            <label class="text-[11px] font-black text-slate-400 uppercase tracking-widest"
              >Pilih Karyawan</label
            >
            <select
              v-model="currentSalesman.employeeID"
              class="w-full bg-slate-50 border-none rounded-2xl px-5 py-4 font-bold text-slate-900 focus:ring-2 focus:ring-primary/20 transition-all appearance-none"
            >
              <option value="" disabled>Pilih Karyawan</option>
              <option v-for="e in allEmployees" :key="e.id" :value="e.id">
                {{ e.first_name }} {{ e.last_name }}
              </option>
            </select>
          </div>
          <div v-else class="space-y-2">
            <label class="text-[11px] font-black text-slate-400 uppercase tracking-widest"
              >Salesman</label
            >
            <div
              class="w-full bg-slate-100/50 rounded-2xl px-5 py-4 font-bold text-slate-900 border border-slate-100"
            >
              {{ currentSalesman.name }}
            </div>
          </div>

          <div class="space-y-2">
            <label class="text-[11px] font-black text-slate-400 uppercase tracking-widest"
              >Perusahaan (PT)</label
            >
            <select
              v-model="currentSalesman.companyID"
              class="w-full bg-slate-50 border-none rounded-2xl px-5 py-4 font-bold text-slate-900 focus:ring-2 focus:ring-primary/20 transition-all appearance-none"
            >
              <option value="" disabled>Pilih Perusahaan</option>
              <option v-for="pt in companies" :key="pt.id" :value="pt.id">
                {{ pt.name }}
              </option>
            </select>
          </div>

          <div class="space-y-2">
            <label class="text-[11px] font-black text-slate-400 uppercase tracking-widest"
              >Wilayah Kerja (Area)</label
            >
            <input
              v-model="currentSalesman.workingTerritory"
              type="text"
              placeholder="Contoh: Kabupaten Boyolali"
              class="w-full bg-slate-50 border-none rounded-2xl px-5 py-4 font-bold text-slate-900 focus:ring-2 focus:ring-primary/20 transition-all"
            />
          </div>

          <div>
            <label
              class="text-[11px] font-black text-slate-400 uppercase tracking-widest ml-2 mb-2 block"
              >Target Omzet (Rp)</label
            >
            <input
              v-model.number="currentSalesman.targetOmzet"
              type="number"
              placeholder="0"
              class="w-full bg-slate-50 border-none rounded-2xl px-5 py-4 font-bold text-slate-900 focus:ring-2 focus:ring-primary/20 transition-all"
            />
          </div>

          <div class="grid grid-cols-2 gap-4">
            <div>
              <label
                class="text-[11px] font-black text-slate-400 uppercase tracking-widest ml-2 mb-2 block"
                >Target Harian: Kunjungan</label
              >
              <input
                v-model.number="currentSalesman.targetDailyVisits"
                type="number"
                placeholder="0"
                class="w-full bg-slate-50 border-none rounded-2xl px-5 py-4 font-bold text-slate-900 focus:ring-2 focus:ring-primary/20 transition-all"
              />
            </div>
            <div>
              <label
                class="text-[11px] font-black text-slate-400 uppercase tracking-widest ml-2 mb-2 block"
                >Target Harian: Toko Baru</label
              >
              <input
                v-model.number="currentSalesman.targetDailyNewStores"
                type="number"
                placeholder="0"
                class="w-full bg-slate-50 border-none rounded-2xl px-5 py-4 font-bold text-slate-900 focus:ring-2 focus:ring-primary/20 transition-all"
              />
            </div>
          </div>

          <div>
            <label
              class="text-[11px] font-black text-slate-400 uppercase tracking-widest ml-2 mb-2 block"
              >Total Toko Baru Bulanan</label
            >
            <input
              v-model.number="currentSalesman.targetNewStores"
              type="number"
              placeholder="0"
              class="w-full bg-slate-50 border-none rounded-2xl px-5 py-4 font-bold text-slate-900 focus:ring-2 focus:ring-primary/20 transition-all"
            />
          </div>

          <div class="grid grid-cols-2 gap-4">
            <div>
              <label
                class="text-[11px] font-black text-slate-400 uppercase tracking-widest ml-2 mb-2 block"
                >Limit Piutang (% Omzet)</label
              >
              <input
                v-model.number="currentSalesman.receivableLimitPercent"
                type="number"
                placeholder="0"
                min="0"
                max="100"
                step="0.1"
                class="w-full bg-slate-50 border-none rounded-2xl px-5 py-4 font-bold text-slate-900 focus:ring-2 focus:ring-primary/20 transition-all"
              />
            </div>
            <div>
              <label
                class="text-[11px] font-black text-slate-400 uppercase tracking-widest ml-2 mb-2 block"
                >Limit Piutang (Rp Nominal)</label
              >
              <input
                v-model.number="currentSalesman.maxReceivableLimit"
                @input="handleMaxReceivableLimitInput"
                type="number"
                placeholder="0"
                min="0"
                class="w-full bg-slate-50 border-none rounded-2xl px-5 py-4 font-bold text-slate-900 focus:ring-2 focus:ring-primary/20 transition-all"
              />
            </div>
          </div>

          <div>
            <label
              class="text-[11px] font-black text-slate-400 uppercase tracking-widest ml-2 mb-2 block"
              >Komisi Capai Target (Rp)</label
            >
            <input
              v-model.number="currentSalesman.commission"
              type="number"
              placeholder="0"
              class="w-full bg-slate-50 border-none rounded-2xl px-5 py-4 font-bold text-slate-900 focus:ring-2 focus:ring-primary/20 transition-all"
            />
          </div>

          <div class="grid grid-cols-2 gap-4">
            <div class="space-y-2">
              <label class="text-[11px] font-black text-slate-400 uppercase tracking-widest"
                >Bulan</label
              >
              <select
                v-model="currentSalesman.month"
                class="w-full bg-slate-50 border-none rounded-2xl px-5 py-4 font-bold text-slate-900 focus:ring-2 focus:ring-primary/20 transition-all appearance-none"
              >
                <option v-for="m in 12" :key="m" :value="m">
                  {{ new Date(2024, m - 1).toLocaleString('id-ID', { month: 'long' }) }}
                </option>
              </select>
            </div>
            <div class="space-y-2">
              <label class="text-[11px] font-black text-slate-400 uppercase tracking-widest"
                >Tahun</label
              >
              <select
                v-model="currentSalesman.year"
                class="w-full bg-slate-50 border-none rounded-2xl px-5 py-4 font-bold text-slate-900 focus:ring-2 focus:ring-primary/20 transition-all appearance-none"
              >
                <option :value="new Date().getFullYear()">{{ new Date().getFullYear() }}</option>
                <option :value="new Date().getFullYear() + 1">
                  {{ new Date().getFullYear() + 1 }}
                </option>
              </select>
            </div>
          </div>
        </div>

        <div class="p-8 bg-slate-50/50 border-t border-slate-100 flex gap-4 shrink-0">
          <button
            @click="isModalOpen = false"
            class="flex-1 py-4 text-sm font-black text-slate-500 hover:text-slate-900 transition-all"
          >
            BATAL
          </button>
          <button
            @click="handleSaveTarget"
            class="flex-[2] bg-primary text-white rounded-2xl py-4 text-sm font-black shadow-xl shadow-primary/20 hover:scale-[1.02] active:scale-[0.98] transition-all flex items-center justify-center gap-2"
          >
            <TrendingUp class="w-5 h-5" /> SIMPAN TARGET
          </button>
        </div>
      </div>
    </div>
    <!-- Visit History Modal -->
    <div v-if="showHistoryModal" class="fixed inset-0 z-[60] flex items-center justify-center p-4">
      <div
        class="absolute inset-0 bg-slate-900/60 backdrop-blur-md"
        @click="showHistoryModal = false"
      ></div>
      <div
        class="bg-white rounded-[40px] w-full max-w-2xl relative z-10 shadow-2xl flex flex-col max-h-[85vh] overflow-hidden animate-in fade-in slide-in-from-bottom-4 duration-300"
      >
        <!-- Header -->
        <div
          class="p-8 border-b border-slate-100 flex items-center justify-between bg-slate-50/50 shrink-0"
        >
          <div class="flex items-center gap-5">
            <div
              class="w-14 h-14 rounded-2xl bg-emerald-500 flex items-center justify-center text-white shadow-lg shadow-emerald-200"
            >
              <History class="w-7 h-7" />
            </div>
            <div>
              <h3 class="text-2xl font-black text-slate-900 tracking-tight">Riwayat Kunjungan</h3>
              <p class="text-sm font-bold text-slate-400 mt-0.5">
                {{ selectedSalesman?.name }} • {{ getMonthLabel(currentMonth) }} {{ currentYear }}
              </p>
            </div>
          </div>
          <button
            @click="showHistoryModal = false"
            class="p-3 hover:bg-white rounded-2xl transition-all text-slate-400 hover:text-slate-900 shadow-sm"
          >
            <X class="w-6 h-6" />
          </button>
        </div>

        <!-- Content -->
        <div class="flex-1 overflow-y-auto p-8 custom-scrollbar">
          <div v-if="isHistoryLoading" class="flex flex-col items-center justify-center py-20">
            <div
              class="w-12 h-12 border-4 border-emerald-100 border-t-emerald-500 rounded-full animate-spin"
            ></div>
            <p class="text-sm font-bold text-slate-400 mt-4">Mengambil data kunjungan...</p>
          </div>

          <div
            v-else-if="visitHistory.length === 0"
            class="flex flex-col items-center justify-center py-20 text-center"
          >
            <div class="w-20 h-20 bg-slate-50 rounded-full flex items-center justify-center mb-4">
              <MapPin class="w-10 h-10 text-slate-200" />
            </div>
            <h4 class="text-lg font-black text-slate-900">Belum ada kunjungan</h4>
            <p class="text-sm font-medium text-slate-400 mt-1 max-w-[240px]">
              Salesman ini belum melakukan kunjungan tercatat di periode ini.
            </p>
          </div>

          <div v-else class="space-y-6">
            <div
              v-for="visit in visitHistory"
              :key="visit.id"
              class="group/item flex gap-6 relative"
            >
              <!-- Timeline Line -->
              <div
                class="absolute left-[27px] top-14 bottom-[-24px] w-0.5 bg-slate-100 group-last/item:hidden"
              ></div>

              <!-- Time/Indicator -->
              <div class="shrink-0 flex flex-col items-center">
                <div
                  class="w-14 h-14 rounded-2xl bg-white border-2 border-slate-100 flex flex-col items-center justify-center group-hover/item:border-emerald-500 transition-colors z-10 shadow-sm"
                >
                  <span class="text-[10px] font-black text-slate-400 uppercase leading-none">{{
                    new Date(visit.check_in_time).toLocaleDateString('id-ID', { day: '2-digit' })
                  }}</span>
                  <span class="text-[10px] font-black text-slate-900 uppercase leading-none mt-1">{{
                    new Date(visit.check_in_time).toLocaleDateString('id-ID', { month: 'short' })
                  }}</span>
                </div>
              </div>

              <!-- Card -->
              <div
                class="flex-1 bg-slate-50/50 hover:bg-white rounded-[24px] p-5 border border-slate-100 hover:border-emerald-200 transition-all hover:shadow-xl hover:shadow-emerald-500/5 group-hover/item:translate-x-2"
              >
                <div class="flex justify-between items-start mb-4">
                  <div class="flex items-center gap-2">
                    <div class="w-2 h-2 rounded-full bg-emerald-500"></div>
                    <span class="text-xs font-black text-slate-900">{{
                      new Date(visit.check_in_time).toLocaleTimeString([], {
                        hour: '2-digit',
                        minute: '2-digit',
                      })
                    }}</span>
                    <span class="text-slate-300 mx-1">—</span>
                    <span class="text-xs font-bold text-slate-400">{{
                      visit.check_out_time
                        ? new Date(visit.check_out_time).toLocaleTimeString([], {
                            hour: '2-digit',
                            minute: '2-digit',
                          })
                        : 'Masih Berkunjung'
                    }}</span>
                  </div>
                  <div
                    class="flex items-center gap-1.5 px-2.5 py-1 bg-white border border-slate-100 rounded-lg shadow-sm"
                  >
                    <MapPin class="w-3 h-3 text-emerald-500" />
                    <span class="text-[10px] font-black text-slate-600 uppercase tracking-tighter"
                      >Lokasi Terverifikasi</span
                    >
                  </div>
                </div>

                <div v-if="visit.transactions && visit.transactions.length > 0">
                  <div
                    v-for="trx in visit.transactions"
                    :key="trx.id"
                    class="mb-3 last:mb-0 bg-white p-3 rounded-xl border border-slate-100"
                  >
                    <div
                      class="flex items-center justify-between mb-2 pb-2 border-b border-slate-50"
                    >
                      <div class="flex items-center gap-3">
                        <div
                          class="w-8 h-8 rounded-lg bg-emerald-50 flex items-center justify-center text-emerald-600"
                        >
                          <Building2 class="w-4 h-4" />
                        </div>
                        <div>
                          <p class="text-[11px] font-black text-slate-900">
                            {{ trx.store?.name || 'Toko Unknown' }}
                          </p>
                          <p class="text-[9px] font-bold text-slate-400">Transaksi Berhasil</p>
                        </div>
                      </div>
                      <p class="text-xs font-black text-emerald-600">
                        {{ formatCurrency(trx.total_amount) }}
                      </p>
                    </div>
                    <div v-if="trx.items && trx.items.length > 0" class="pl-11 space-y-1.5 mt-1">
                      <div
                        v-for="item in trx.items"
                        :key="item.id"
                        class="flex justify-between items-center text-[10px]"
                      >
                        <span class="font-bold text-slate-600 truncate mr-2 flex-1">{{
                          item.product?.name || 'Produk'
                        }}</span>
                        <span class="font-bold text-slate-400 whitespace-nowrap"
                          >{{ item.quantity }} × {{ formatCurrency(item.unit_price) }}</span
                        >
                      </div>
                    </div>
                    <div v-else class="pl-11 text-[9px] font-bold text-slate-300 italic">
                      Tidak ada rincian produk
                    </div>
                  </div>
                </div>
                <div
                  v-else-if="visit.notes"
                  class="bg-white p-3 rounded-xl border border-slate-100 border-dashed"
                >
                  <p class="text-[11px] font-bold text-slate-500 italic">"{{ visit.notes }}"</p>
                </div>
                <div v-else class="flex items-center gap-2 text-slate-300">
                  <div class="h-[1px] flex-1 bg-slate-100"></div>
                  <span class="text-[9px] font-black uppercase tracking-widest"
                    >Tidak ada catatan</span
                  >
                  <div class="h-[1px] flex-1 bg-slate-100"></div>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- Footer -->
        <div
          class="p-8 bg-slate-50/50 border-t border-slate-100 flex justify-between items-center shrink-0"
        >
          <div class="flex items-center gap-3">
            <p class="text-xs font-bold text-slate-400">Total Kunjungan:</p>
            <span
              class="px-3 py-1 bg-white border border-slate-200 rounded-lg text-sm font-black text-slate-900 shadow-sm"
              >{{ visitHistory.length }}</span
            >
          </div>
          <div class="flex items-center gap-3 ml-4 border-l border-slate-200 pl-4">
            <p class="text-xs font-bold text-slate-400">KPI Harian:</p>
            <div class="flex items-center gap-1.5">
              <span
                class="px-3 py-1 bg-indigo-50 border border-indigo-100 rounded-lg text-sm font-black text-indigo-700 shadow-sm"
                >{{ dailyVisitKPI }}</span
              >
              <span class="text-[10px] font-bold text-slate-400 uppercase tracking-tight"
                >Kunjungan/Hari</span
              >
            </div>
          </div>
          <div class="flex-1"></div>
          <Button
            @click="showHistoryModal = false"
            class="bg-slate-900 text-white rounded-2xl px-8 py-3.5 font-bold shadow-lg shadow-slate-200 hover:scale-[1.02] active:scale-[0.98] transition-all"
          >
            TUTUP
          </Button>
        </div>
      </div>
    </div>

    <!-- Modal Impor Cadangan Offline -->
    <div
      v-if="isImportModalOpen"
      class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-950/60 backdrop-blur-sm animate-in fade-in duration-200"
    >
      <div
        class="bg-white rounded-[2.5rem] w-full max-w-2xl overflow-hidden shadow-2xl border border-slate-100 flex flex-col max-h-[90vh] animate-in zoom-in-95 duration-200"
      >
        <!-- Modal Header -->
        <div class="p-8 pb-6 bg-slate-50/50 border-b border-slate-100 flex justify-between items-start shrink-0">
          <div class="flex items-center gap-4">
            <div class="w-12 h-12 rounded-2xl bg-indigo-50 border border-indigo-100 flex items-center justify-center text-indigo-600 shadow-sm">
              <UploadCloud class="w-6 h-6" />
            </div>
            <div>
              <h3 class="text-xl font-extrabold text-slate-900 tracking-tight">Impor Cadangan Offline</h3>
              <p class="text-xs font-semibold text-slate-400 mt-1">
                Masukkan teks JSON cadangan transaksi dari menu aplikasi sales
              </p>
            </div>
          </div>
          <button
            @click="isImportModalOpen = false"
            class="w-10 h-10 rounded-xl bg-white border border-slate-200 flex items-center justify-center text-slate-400 hover:text-slate-600 hover:border-slate-300 transition-all shadow-sm"
          >
            <X class="w-5 h-5" />
          </button>
        </div>

        <!-- Modal Body -->
        <div class="p-8 overflow-y-auto space-y-6 custom-scrollbar flex-1">
          <div class="bg-blue-50 border border-blue-100 rounded-2xl p-4 flex gap-3 items-start">
            <FileJson class="w-5 h-5 text-blue-600 shrink-0 mt-0.5" />
            <div class="text-xs text-blue-700 leading-relaxed font-medium">
              Fitur darurat ini digunakan bila sales mengalami kendala jaringan atau sinkronisasi. Minta sales membuka menu drawer di aplikasi mobile &rarr; <strong>Cadangan Data Offline</strong> &rarr; Salin Teks Cadangan dan tempel di bawah.
            </div>
          </div>

          <div class="space-y-2">
            <label class="text-xs font-bold text-slate-600 uppercase tracking-wider">
              Data Cadangan JSON Transaksi
            </label>
            <textarea
              v-model="backupJsonText"
              rows="8"
              placeholder='Tempelkan format JSON di sini, contoh: {"export_time":"...", "transactions":[...]}'
              class="w-full bg-slate-50 border border-slate-200 rounded-2xl p-4 font-mono text-xs text-slate-800 focus:ring-2 focus:ring-primary/20 focus:outline-none transition-all placeholder:text-slate-400"
            ></textarea>
          </div>

          <!-- Hasil Impor -->
          <div v-if="importResult" class="p-5 bg-slate-50 rounded-2xl border border-slate-200 space-y-3">
            <h4 class="text-xs font-extrabold text-slate-700 uppercase tracking-wider">Hasil Pemrosesan:</h4>
            <div class="grid grid-cols-3 gap-3 text-center">
              <div class="bg-emerald-50 border border-emerald-200 rounded-xl p-3">
                <p class="text-[11px] font-bold text-emerald-600">Berhasil</p>
                <p class="text-xl font-black text-emerald-700 mt-1">{{ importResult.success_count }}</p>
              </div>
              <div class="bg-amber-50 border border-amber-200 rounded-xl p-3">
                <p class="text-[11px] font-bold text-amber-600">Duplikat/Sudah Ada</p>
                <p class="text-xl font-black text-amber-700 mt-1">{{ importResult.duplicate_count }}</p>
              </div>
              <div class="bg-rose-50 border border-rose-200 rounded-xl p-3">
                <p class="text-[11px] font-bold text-rose-600">Gagal</p>
                <p class="text-xl font-black text-rose-700 mt-1">{{ importResult.failed_count }}</p>
              </div>
            </div>
            <div v-if="importResult.errors && importResult.errors.length > 0" class="mt-3 space-y-1">
              <p class="text-xs font-bold text-rose-600">Daftar Error:</p>
              <ul class="text-xs text-rose-600 list-disc list-inside bg-rose-50/50 p-2 rounded-lg font-mono">
                <li v-for="(err, idx) in importResult.errors" :key="idx">{{ err }}</li>
              </ul>
            </div>
          </div>
        </div>

        <!-- Modal Footer -->
        <div class="p-8 bg-slate-50/50 border-t border-slate-100 flex justify-end items-center gap-3 shrink-0">
          <Button
            @click="isImportModalOpen = false"
            variant="outline"
            class="rounded-2xl px-6 py-3.5 font-bold border-slate-200 text-slate-600"
          >
            Batal
          </Button>
          <Button
            @click="handleImportBackup"
            :disabled="isImporting || !backupJsonText.trim()"
            class="bg-indigo-600 hover:bg-indigo-700 text-white rounded-2xl px-8 py-3.5 font-bold shadow-lg shadow-indigo-200 flex items-center gap-2"
          >
            <UploadCloud class="w-4 h-4" :class="{ 'animate-bounce': isImporting }" />
            <span>{{ isImporting ? 'Memproses...' : 'Proses Impor' }}</span>
          </Button>
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
