<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { toast } from 'vue-sonner'
import axios from '@/api/axios'
import { useAuthStore } from '@/stores/auth'
import {
  CheckCircle2,
  X,
  Download,
  Search,
  Calculator,
  Printer,
  FileText,
  Eye,
  AlertCircle,
  Plus,
  Trash2,
  Building2,
  Calendar,
  Clock,
  ShieldCheck,
  CreditCard,
  Sparkles,
} from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from '@/components/ui/dialog'
import { Badge } from '@/components/ui/badge'

const authStore = useAuthStore()
const companyName = computed(() => authStore.user?.company_name || 'Perusahaan')
const companyAddress = computed(() => authStore.user?.company_address || '')
const requests = ref<any[]>([])
const isLoading = ref(true)
const statusFilter = ref('')
const searchQuery = ref('')

// Auto-Lookup Processing State
const isProcessModalOpen = ref(false)
const isLookupLoading = ref(false)
const selectedRequest = ref<any>(null)
const lookupData = ref<any>(null)

// Processing Form
const processForm = ref({
  basicSalary: 0,
  allowances: [] as Array<{ name: string; amount: number }>,
  deductions: [] as Array<{ name: string; amount: number }>,
  file: null as File | null,
})
const isSubmitting = ref(false)

// Reject Modal State
const isRejectModalOpen = ref(false)
const rejectReason = ref('')
const requestToReject = ref<any>(null)

// Detail / Print Modal State
const isDetailModalOpen = ref(false)
const selectedDetailRequest = ref<any>(null)

const formatCurrency = (val: number | null | undefined): string => {
  if (val === null || val === undefined || isNaN(val)) return 'Rp 0'
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    maximumFractionDigits: 0,
  }).format(val)
}

const getMonthName = (month: number): string => {
  const months = [
    '', 'Januari', 'Februari', 'Maret', 'April', 'Mei', 'Juni',
    'Juli', 'Agustus', 'September', 'Oktober', 'November', 'Desember',
  ]
  return months[month] || `Bulan ${month}`
}

const fetchRequests = async () => {
  isLoading.value = true
  try {
    const url = `/admin/payslip-requests?limit=100${statusFilter.value ? '&status=' + statusFilter.value : ''}`
    const res = await axios.get(url)
    requests.value = res.data?.data || []
  } catch (error: any) {
    toast.error(error.response?.data?.error || 'Gagal memuat request slip gaji')
  } finally {
    isLoading.value = false
  }
}

// Metrics
const stats = computed(() => {
  const total = requests.value.length
  const pending = requests.value.filter((r) => r.status === 'pending').length
  const approved = requests.value.filter((r) => r.status === 'approved').length
  const rejected = requests.value.filter((r) => r.status === 'rejected').length
  return { total, pending, approved, rejected }
})

// Filtered Requests
const filteredRequests = computed(() => {
  return requests.value.filter((r) => {
    const emp = r.user?.employee
    const empName = emp ? `${emp.first_name} ${emp.last_name || ''}`.trim() : r.user?.email || ''
    const nik = emp?.employee_id_number || ''
    const reason = r.reason || ''

    const q = searchQuery.value.toLowerCase().trim()
    if (!q) return true
    return (
      empName.toLowerCase().includes(q) ||
      nik.toLowerCase().includes(q) ||
      reason.toLowerCase().includes(q)
    )
  })
})

// Auto-Lookup Trigger
const openProcessModal = async (req: any) => {
  selectedRequest.value = req
  lookupData.value = null
  processForm.value = {
    basicSalary: 0,
    allowances: [],
    deductions: [],
    file: null,
  }
  isProcessModalOpen.value = true
  isLookupLoading.value = true

  try {
    const res = await axios.get(`/admin/payslip-requests/${req.id}/lookup`)
    const data = res.data?.data
    lookupData.value = data

    // Initialize editable form from lookup data
    processForm.value.basicSalary = data.basic_salary || 0
    processForm.value.allowances = (data.allowances || []).map((item: any) => ({
      name: item.name,
      amount: item.amount,
    }))
    processForm.value.deductions = (data.deductions || []).map((item: any) => ({
      name: item.name,
      amount: item.amount,
    }))
  } catch (error: any) {
    toast.error('Gagal melakukan lookup payroll: ' + (error.response?.data?.error || error.message))
  } finally {
    isLookupLoading.value = false
  }
}

// Computed Calculations in Process Modal
const calculatedTotalEarning = computed(() => {
  const allowanceSum = processForm.value.allowances.reduce((acc, curr) => acc + (Number(curr.amount) || 0), 0)
  return (Number(processForm.value.basicSalary) || 0) + allowanceSum
})

const calculatedTotalDeduction = computed(() => {
  return processForm.value.deductions.reduce((acc, curr) => acc + (Number(curr.amount) || 0), 0)
})

const calculatedTakeHomePay = computed(() => {
  const thp = calculatedTotalEarning.value - calculatedTotalDeduction.value
  return thp > 0 ? thp : 0
})

const addAllowanceItem = () => {
  processForm.value.allowances.push({ name: 'Tunjangan Tambahan', amount: 0 })
}

const removeAllowanceItem = (index: number) => {
  processForm.value.allowances.splice(index, 1)
}

const addDeductionItem = () => {
  processForm.value.deductions.push({ name: 'Potongan Tambahan', amount: 0 })
}

const removeDeductionItem = (index: number) => {
  processForm.value.deductions.splice(index, 1)
}

const handleFileChange = (e: Event) => {
  const target = e.target as HTMLInputElement
  if (target.files && target.files.length > 0) {
    processForm.value.file = target.files[0] as File
  }
}

const submitApprove = async () => {
  if (calculatedTakeHomePay.value <= 0 && !processForm.value.file) {
    toast.error('Take Home Pay bernilai Rp 0 dan tidak ada file yang diunggah')
    return
  }

  isSubmitting.value = true
  try {
    const formData = new FormData()
    formData.append('nominal', calculatedTakeHomePay.value.toString())

    if (processForm.value.file) {
      formData.append('file', processForm.value.file)
    }

    // Build breakdown snapshot JSON
    const payslipDataSnapshot = {
      period_month: selectedRequest.value.period_month,
      period_year: selectedRequest.value.period_year,
      period_label: lookupData.value?.period_label || `${getMonthName(selectedRequest.value.period_month)} ${selectedRequest.value.period_year}`,
      employee: lookupData.value?.employee || {},
      basic_salary: processForm.value.basicSalary,
      allowances: processForm.value.allowances,
      total_allowance: calculatedTotalEarning.value - processForm.value.basicSalary,
      total_earning: calculatedTotalEarning.value,
      attendance: lookupData.value?.attendance || {},
      leaves: lookupData.value?.leaves || {},
      bpjs: lookupData.value?.bpjs || {},
      deductions: processForm.value.deductions,
      total_deduction: calculatedTotalDeduction.value,
      take_home_pay: calculatedTakeHomePay.value,
    }

    formData.append('payslip_data', JSON.stringify(payslipDataSnapshot))

    await axios.put(`/admin/payslip-requests/${selectedRequest.value.id}/approve`, formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    })

    toast.success('Permohonan slip gaji berhasil disetujui & diterbitkan!')
    isProcessModalOpen.value = false
    fetchRequests()
  } catch (error: any) {
    toast.error(error.response?.data?.error || 'Gagal menyetujui request')
  } finally {
    isSubmitting.value = false
  }
}

const openRejectDialog = (req: any) => {
  requestToReject.value = req
  rejectReason.value = ''
  isRejectModalOpen.value = true
}

const submitReject = async () => {
  if (!rejectReason.value.trim()) {
    toast.error('Harap masukkan alasan penolakan')
    return
  }

  try {
    await axios.put(`/admin/payslip-requests/${requestToReject.value.id}/reject`, {
      reason: rejectReason.value,
    })
    toast.success('Permohonan slip gaji berhasil ditolak')
    isRejectModalOpen.value = false
    fetchRequests()
  } catch (error: any) {
    toast.error(error.response?.data?.error || 'Gagal menolak request')
  }
}

const openDetailModal = (req: any) => {
  selectedDetailRequest.value = req
  isDetailModalOpen.value = true
}

const getFullUrl = (url: string | undefined) => {
  if (!url) return ''
  if (url.startsWith('http')) return url
  const apiBaseUrl = import.meta.env.VITE_API_BASE_URL || ''
  try {
    const origin = new URL(apiBaseUrl).origin
    return `${origin}${url.startsWith('/') ? url : '/' + url}`
  } catch (e) {
    return url
  }
}

const parsedPayslipData = (req: any) => {
  if (!req?.payslip_data) return null
  try {
    if (typeof req.payslip_data === 'string') {
      return JSON.parse(req.payslip_data)
    }
    return req.payslip_data
  } catch (e) {
    return null
  }
}

const printPayslip = () => {
  const printArea = document.getElementById('printable-payslip-area')
  if (!printArea) return

  const printWindow = window.open('', '_blank')
  if (!printWindow) return

  printWindow.document.write(`<html><head><title>SLIP GAJI - ${companyName.value}</title>`)
  printWindow.document.write(`<base href="${window.location.origin}/">`)
  const styles = document.querySelectorAll('style')
  styles.forEach((s) => printWindow.document.write(s.outerHTML))
  const links = document.querySelectorAll('link[rel="stylesheet"]')
  links.forEach((l) => printWindow.document.write(l.outerHTML))

  printWindow.document.write('</head><body style="padding: 20px;">')
  printWindow.document.write(printArea.innerHTML)
  printWindow.document.write('</body></html>')
  printWindow.document.close()

  setTimeout(() => {
    printWindow.print()
  }, 400)
}

onMounted(() => {
  fetchRequests()
})
</script>

<template>
  <div class="space-y-6 animate-in fade-in slide-in-from-bottom-4 duration-500">
    <!-- Header & Metric Cards -->
    <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
      <div
        @click="statusFilter = '' ; fetchRequests()"
        class="bg-white p-5 rounded-2xl border border-slate-100 shadow-sm cursor-pointer transition-all hover:shadow"
        :class="statusFilter === '' ? 'ring-2 ring-primary border-primary/20' : ''"
      >
        <div class="flex items-center justify-between text-slate-500">
          <span class="text-xs font-black uppercase tracking-wider">Total Request</span>
          <FileText class="w-5 h-5 text-primary" />
        </div>
        <p class="text-3xl font-black text-slate-900 mt-2">{{ stats.total }}</p>
        <p class="text-[11px] font-bold text-slate-400 mt-1">Semua permohonan</p>
      </div>

      <div
        @click="statusFilter = 'pending' ; fetchRequests()"
        class="bg-white p-5 rounded-2xl border border-slate-100 shadow-sm cursor-pointer transition-all hover:shadow"
        :class="statusFilter === 'pending' ? 'ring-2 ring-amber-500 border-amber-500/20' : ''"
      >
        <div class="flex items-center justify-between text-amber-600">
          <span class="text-xs font-black uppercase tracking-wider">Menunggu Tindakan</span>
          <Clock class="w-5 h-5 text-amber-500 animate-pulse" />
        </div>
        <p class="text-3xl font-black text-amber-600 mt-2">{{ stats.pending }}</p>
        <p class="text-[11px] font-bold text-amber-700/70 mt-1">Butuh proses & lookup</p>
      </div>

      <div
        @click="statusFilter = 'approved' ; fetchRequests()"
        class="bg-white p-5 rounded-2xl border border-slate-100 shadow-sm cursor-pointer transition-all hover:shadow"
        :class="statusFilter === 'approved' ? 'ring-2 ring-emerald-500 border-emerald-500/20' : ''"
      >
        <div class="flex items-center justify-between text-emerald-600">
          <span class="text-xs font-black uppercase tracking-wider">Disetujui / Terbit</span>
          <CheckCircle2 class="w-5 h-5 text-emerald-500" />
        </div>
        <p class="text-3xl font-black text-emerald-600 mt-2">{{ stats.approved }}</p>
        <p class="text-[11px] font-bold text-emerald-700/70 mt-1">Slip gaji aktif</p>
      </div>

      <div
        @click="statusFilter = 'rejected' ; fetchRequests()"
        class="bg-white p-5 rounded-2xl border border-slate-100 shadow-sm cursor-pointer transition-all hover:shadow"
        :class="statusFilter === 'rejected' ? 'ring-2 ring-red-500 border-red-500/20' : ''"
      >
        <div class="flex items-center justify-between text-red-600">
          <span class="text-xs font-black uppercase tracking-wider">Ditolak</span>
          <X class="w-5 h-5 text-red-500" />
        </div>
        <p class="text-3xl font-black text-red-600 mt-2">{{ stats.rejected }}</p>
        <p class="text-[11px] font-bold text-red-700/70 mt-1">Permohonan dibatalkan</p>
      </div>
    </div>

    <!-- Filters & Search -->
    <div class="flex flex-col md:flex-row items-center justify-between gap-4 bg-white p-4 rounded-2xl border border-slate-100 shadow-sm">
      <div class="relative w-full md:w-96">
        <Search class="absolute left-3.5 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400" />
        <Input
          v-model="searchQuery"
          placeholder="Cari nama karyawan, NIK, atau keperluan..."
          class="pl-10 h-11 rounded-xl bg-slate-50 border-none text-sm font-medium"
        />
      </div>

      <div class="flex items-center gap-2 w-full md:w-auto">
        <select
          v-model="statusFilter"
          @change="fetchRequests"
          class="h-11 rounded-xl border border-slate-200 bg-white px-4 text-xs font-bold text-slate-700 shadow-sm focus:outline-none"
        >
          <option value="">Semua Status</option>
          <option value="pending">Menunggu (Pending)</option>
          <option value="approved">Disetujui (Approved)</option>
          <option value="rejected">Ditolak (Rejected)</option>
        </select>
        <Button variant="outline" @click="fetchRequests" class="h-11 rounded-xl font-bold text-xs">
          Refresh
        </Button>
      </div>
    </div>

    <!-- Table of Requests -->
    <div class="rounded-2xl border bg-white shadow-sm overflow-hidden">
      <div v-if="isLoading" class="p-12 text-center text-slate-400 flex flex-col items-center justify-center">
        <div class="w-8 h-8 border-4 border-primary border-t-transparent rounded-full animate-spin mb-4"></div>
        Memuat data permohonan slip gaji...
      </div>

      <div v-else-if="filteredRequests.length === 0" class="p-12 text-center text-slate-400 flex flex-col items-center justify-center">
        <div class="w-16 h-16 bg-slate-100 rounded-full flex items-center justify-center mb-4 text-slate-300">
          <FileText class="w-8 h-8" />
        </div>
        <p class="text-base font-bold text-slate-700">Tidak ada permohonan slip gaji</p>
        <p class="text-xs text-slate-400 mt-1">Belum ada request karyawan yang cocok dengan filter saat ini.</p>
      </div>

      <div v-else class="overflow-x-auto">
        <table class="w-full text-sm text-left">
          <thead class="text-xs text-slate-400 uppercase bg-slate-50/80 border-b font-extrabold tracking-wider">
            <tr>
              <th class="px-5 py-4">Karyawan</th>
              <th class="px-5 py-4">Periode</th>
              <th class="px-5 py-4">Keperluan / Alasan</th>
              <th class="px-5 py-4">Tgl Pengajuan</th>
              <th class="px-5 py-4">Nominal / THP</th>
              <th class="px-5 py-4">Status</th>
              <th class="px-5 py-4 text-center">Aksi HRD</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-100">
            <tr v-for="req in filteredRequests" :key="req.id" class="hover:bg-slate-50/50 transition-colors">
              <td class="px-5 py-4">
                <div class="flex flex-col">
                  <span class="font-extrabold text-slate-900 text-sm">
                    {{ req.user?.employee ? `${req.user.employee.first_name} ${req.user.employee.last_name || ''}`.trim() : (req.user?.email || '-') }}
                  </span>
                  <div class="flex items-center gap-2 mt-0.5">
                    <span class="text-[11px] font-bold text-slate-400">
                      NIK: {{ req.user?.employee?.employee_id_number || '-' }}
                    </span>
                    <span class="text-[11px] font-semibold text-slate-500">
                      • {{ req.user?.employee?.job_position?.title || '-' }}
                    </span>
                  </div>
                </div>
              </td>
              <td class="px-5 py-4">
                <span class="font-bold text-slate-800 text-xs">
                  {{ getMonthName(req.period_month) }} {{ req.period_year }}
                </span>
              </td>
              <td class="px-5 py-4">
                <span class="text-xs text-slate-600 font-medium italic">
                  {{ req.reason || 'Tidak ada catatan keperluan' }}
                </span>
              </td>
              <td class="px-5 py-4 text-xs font-semibold text-slate-500">
                {{ new Date(req.created_at).toLocaleDateString('id-ID', { day: 'numeric', month: 'short', year: 'numeric' }) }}
              </td>
              <td class="px-5 py-4">
                <span v-if="req.nominal" class="font-black text-slate-900 text-xs">
                  {{ formatCurrency(req.nominal) }}
                </span>
                <span v-else class="text-xs text-slate-400 italic">Belum diproses</span>
              </td>
              <td class="px-5 py-4">
                <Badge
                  :class="`text-[10px] font-extrabold uppercase tracking-wider px-2.5 py-0.5 rounded-full ${
                    req.status === 'approved'
                      ? 'bg-emerald-100 text-emerald-800 border border-emerald-200'
                      : req.status === 'rejected'
                        ? 'bg-red-100 text-red-800 border border-red-200'
                        : 'bg-amber-100 text-amber-800 border border-amber-200 animate-pulse'
                  }`"
                >
                  {{ req.status === 'approved' ? 'Disetujui' : req.status === 'rejected' ? 'Ditolak' : 'Menunggu' }}
                </Badge>
              </td>
              <td class="px-5 py-4">
                <div class="flex items-center justify-center gap-2">
                  <template v-if="req.status === 'pending'">
                    <Button
                      size="sm"
                      @click="openProcessModal(req)"
                      class="h-8 rounded-xl font-bold text-xs bg-primary hover:bg-primary/90 text-white gap-1.5 shadow-md shadow-primary/20"
                    >
                      <Calculator class="w-3.5 h-3.5" />
                      Lookup & Proses
                    </Button>
                    <Button
                      size="sm"
                      variant="outline"
                      @click="openRejectDialog(req)"
                      class="h-8 rounded-xl font-bold text-xs text-red-600 border-red-200 hover:bg-red-50 hover:text-red-700"
                    >
                      <X class="w-3.5 h-3.5 mr-1" />
                      Tolak
                    </Button>
                  </template>

                  <template v-else-if="req.status === 'approved'">
                    <Button
                      size="sm"
                      variant="outline"
                      @click="openDetailModal(req)"
                      class="h-8 rounded-xl font-bold text-xs text-slate-700 gap-1.5 hover:bg-slate-50"
                    >
                      <Eye class="w-3.5 h-3.5 text-slate-500" />
                      Rincian
                    </Button>
                    <a
                      v-if="req.file_url"
                      :href="getFullUrl(req.file_url)"
                      target="_blank"
                      rel="noopener noreferrer"
                      class="inline-flex items-center gap-1 h-8 px-2.5 rounded-xl bg-slate-100 hover:bg-slate-200 text-slate-700 text-xs font-bold transition-colors"
                      title="Unduh File Resmi"
                    >
                      <Download class="w-3.5 h-3.5" />
                    </a>
                  </template>

                  <template v-else>
                    <span class="text-[11px] text-slate-400 italic">
                      Ditolak: {{ req.reject_reason || '-' }}
                    </span>
                  </template>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Modal Auto-Lookup & Proses Slip Gaji -->
    <Dialog :open="isProcessModalOpen" @update:open="isProcessModalOpen = $event">
      <DialogContent class="sm:max-w-4xl rounded-[2.5rem] p-0 overflow-hidden border-none shadow-2xl max-h-[95vh] flex flex-col bg-slate-50">
        <!-- Header -->
        <div class="bg-slate-900 p-6 text-white shrink-0">
          <div class="flex items-center justify-between">
            <div class="flex items-center gap-3">
              <div class="w-10 h-10 rounded-xl bg-primary/20 flex items-center justify-center text-primary font-bold">
                <Calculator class="w-5 h-5" />
              </div>
              <div>
                <h2 class="text-xl font-black">Proses E-Slip Gaji dengan Auto-Lookup</h2>
                <p class="text-slate-400 text-xs mt-0.5">
                  Kalkulasi otomatis kehadiran, cuti & izin, dan potongan BPJS Ketenagakerjaan
                </p>
              </div>
            </div>
            <button
              @click="isProcessModalOpen = false"
              class="p-2 rounded-full bg-white/10 hover:bg-white/20 transition-colors"
            >
              <X class="w-5 h-5" />
            </button>
          </div>
        </div>

        <!-- Content Area -->
        <div class="flex-1 overflow-y-auto p-6 space-y-6">
          <div v-if="isLookupLoading" class="py-16 text-center text-slate-400 space-y-3">
            <div class="w-10 h-10 border-4 border-primary border-t-transparent rounded-full animate-spin mx-auto"></div>
            <p class="text-sm font-bold text-slate-700">Menghubungkan ke database presensi, cuti & BPJS...</p>
            <p class="text-xs text-slate-400">Sedang menarik data absensi dan menghitung potongan otomatis.</p>
          </div>

          <div v-else-if="lookupData" class="space-y-6 animate-in fade-in duration-300">
            <!-- Profil & Periode Header Card -->
            <div class="bg-white p-5 rounded-2xl border border-slate-100 shadow-sm flex flex-col md:flex-row md:items-center justify-between gap-4">
              <div>
                <span class="text-[10px] font-black uppercase tracking-wider text-primary bg-primary/10 px-2.5 py-0.5 rounded-full">
                  Periode: {{ lookupData.period_label }}
                </span>
                <h3 class="text-lg font-black text-slate-900 mt-2">{{ lookupData.employee?.name }}</h3>
                <p class="text-xs text-slate-500 font-semibold mt-0.5">
                  NIK: {{ lookupData.employee?.employee_id_number }} • {{ lookupData.employee?.job_position || 'Staff' }} • {{ lookupData.employee?.department || '-' }}
                </p>
              </div>
              <div class="p-3 bg-slate-50 rounded-xl border border-slate-100 text-right space-y-0.5">
                <p class="text-[10px] text-slate-400 font-bold uppercase">Rekening Transfer</p>
                <p class="text-xs font-black text-slate-800">{{ lookupData.employee?.bank_name || 'BCA' }} - {{ lookupData.employee?.bank_account_number || '-' }}</p>
                <p class="text-[10px] text-slate-400 font-bold">PTKP: {{ lookupData.employee?.ptkp_status || 'TK/0' }}</p>
              </div>
            </div>

            <!-- Auto-Lookup Insights Badges -->
            <div class="grid grid-cols-2 md:grid-cols-4 gap-3">
              <div class="bg-white p-3.5 rounded-xl border border-slate-100 shadow-sm">
                <span class="text-[10px] font-bold text-slate-400 uppercase">Presensi Hadir</span>
                <p class="text-lg font-black text-emerald-600 mt-1">{{ lookupData.attendance?.present_days }} Hari</p>
                <p class="text-[10px] text-slate-400">Terlambat: {{ lookupData.attendance?.late_days }} hari</p>
              </div>

              <div class="bg-white p-3.5 rounded-xl border border-slate-100 shadow-sm">
                <span class="text-[10px] font-bold text-slate-400 uppercase">Mangkir / Alpha</span>
                <p class="text-lg font-black text-red-600 mt-1">{{ lookupData.attendance?.absent_days }} Hari</p>
                <p class="text-[10px] text-red-500 font-bold">Potongan: {{ formatCurrency(lookupData.attendance?.absent_deduction) }}</p>
              </div>

              <div class="bg-white p-3.5 rounded-xl border border-slate-100 shadow-sm">
                <span class="text-[10px] font-bold text-slate-400 uppercase">Cuti & Izin</span>
                <p class="text-lg font-black text-blue-600 mt-1">{{ lookupData.leaves?.paid_leave_days }} Hari</p>
                <p class="text-[10px] text-slate-400">Unpaid: {{ lookupData.leaves?.unpaid_leave_days }} hari</p>
              </div>

              <div class="bg-white p-3.5 rounded-xl border border-slate-100 shadow-sm">
                <span class="text-[10px] font-bold text-slate-400 uppercase">BPJS Ketenagakerjaan</span>
                <p class="text-lg font-black text-purple-600 mt-1">{{ formatCurrency((lookupData.bpjs?.jht_employee || 0) + (lookupData.bpjs?.jp_employee || 0)) }}</p>
                <p class="text-[10px] text-slate-400">JHT 2% + JP 1%</p>
              </div>
            </div>

            <!-- Earnings & Deductions Two Columns -->
            <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
              <!-- Earnings Column -->
              <div class="bg-white p-5 rounded-2xl border border-slate-100 shadow-sm space-y-4">
                <div class="flex items-center justify-between border-b pb-2">
                  <h4 class="font-black text-sm text-slate-900 flex items-center gap-1.5">
                    <span class="w-2.5 h-2.5 rounded-full bg-emerald-500"></span>
                    Pendapatan (Earnings)
                  </h4>
                  <Button size="sm" variant="ghost" @click="addAllowanceItem" class="h-7 text-xs font-bold text-primary gap-1">
                    <Plus class="w-3.5 h-3.5" /> Tambah
                  </Button>
                </div>

                <div class="space-y-3">
                  <div class="space-y-1">
                    <label class="text-xs font-bold text-slate-600">Gaji Pokok (Rp)</label>
                    <Input
                      type="number"
                      v-model.number="processForm.basicSalary"
                      class="h-10 rounded-xl bg-slate-50 border-none font-bold text-slate-900"
                    />
                  </div>

                  <div v-for="(allowance, idx) in processForm.allowances" :key="'allowance-' + idx" class="flex items-center gap-2">
                    <Input
                      v-model="allowance.name"
                      placeholder="Nama Tunjangan"
                      class="h-10 rounded-xl bg-slate-50 border-none text-xs font-medium flex-1"
                    />
                    <Input
                      type="number"
                      v-model.number="allowance.amount"
                      placeholder="Nominal"
                      class="h-10 rounded-xl bg-slate-50 border-none text-xs font-bold w-36 text-right"
                    />
                    <button @click="removeAllowanceItem(idx)" class="text-slate-400 hover:text-red-500 p-1">
                      <Trash2 class="w-4 h-4" />
                    </button>
                  </div>
                </div>

                <div class="pt-3 border-t flex justify-between items-center font-black text-sm text-slate-900">
                  <span>Total Pendapatan</span>
                  <span class="text-emerald-600">{{ formatCurrency(calculatedTotalEarning) }}</span>
                </div>
              </div>

              <!-- Deductions Column -->
              <div class="bg-white p-5 rounded-2xl border border-slate-100 shadow-sm space-y-4">
                <div class="flex items-center justify-between border-b pb-2">
                  <h4 class="font-black text-sm text-slate-900 flex items-center gap-1.5">
                    <span class="w-2.5 h-2.5 rounded-full bg-red-500"></span>
                    Potongan (Deductions)
                  </h4>
                  <Button size="sm" variant="ghost" @click="addDeductionItem" class="h-7 text-xs font-bold text-primary gap-1">
                    <Plus class="w-3.5 h-3.5" /> Tambah
                  </Button>
                </div>

                <div class="space-y-3">
                  <div v-for="(deduction, idx) in processForm.deductions" :key="'deduction-' + idx" class="flex items-center gap-2">
                    <Input
                      v-model="deduction.name"
                      placeholder="Nama Potongan"
                      class="h-10 rounded-xl bg-slate-50 border-none text-xs font-medium flex-1"
                    />
                    <Input
                      type="number"
                      v-model.number="deduction.amount"
                      placeholder="Nominal"
                      class="h-10 rounded-xl bg-slate-50 border-none text-xs font-bold w-36 text-right"
                    />
                    <button @click="removeDeductionItem(idx)" class="text-slate-400 hover:text-red-500 p-1">
                      <Trash2 class="w-4 h-4" />
                    </button>
                  </div>
                </div>

                <div class="pt-3 border-t flex justify-between items-center font-black text-sm text-slate-900">
                  <span>Total Potongan</span>
                  <span class="text-red-600">- {{ formatCurrency(calculatedTotalDeduction) }}</span>
                </div>
              </div>
            </div>

            <!-- Take Home Pay Highlight Card -->
            <div class="bg-slate-900 p-6 rounded-2xl text-white flex flex-col md:flex-row items-center justify-between gap-4 shadow-xl">
              <div>
                <p class="text-xs uppercase font-extrabold tracking-widest text-primary">Estimasi Gaji Bersih</p>
                <h3 class="text-3xl font-black mt-1">{{ formatCurrency(calculatedTakeHomePay) }}</h3>
                <p class="text-xs text-slate-400 mt-1">
                  Total Pendapatan ({{ formatCurrency(calculatedTotalEarning) }}) - Total Potongan ({{ formatCurrency(calculatedTotalDeduction) }})
                </p>
              </div>

              <!-- Upload PDF Fisik (Opsional) -->
              <div class="bg-white/10 p-3.5 rounded-xl border border-white/10 text-xs space-y-2 w-full md:w-80">
                <label class="font-bold text-slate-200 block">Lampirkan Berkas PDF Slip Resmi (Opsional)</label>
                <input
                  type="file"
                  accept=".pdf,image/*"
                  @change="handleFileChange"
                  class="text-[11px] text-slate-300 file:mr-2 file:py-1 file:px-2.5 file:rounded-lg file:border-0 file:text-[10px] file:font-black file:bg-primary file:text-white hover:file:bg-primary/90 cursor-pointer"
                />
              </div>
            </div>
          </div>
        </div>

        <!-- Footer -->
        <div class="p-6 bg-white border-t border-slate-100 flex items-center justify-between shrink-0">
          <Button variant="outline" @click="isProcessModalOpen = false" class="rounded-xl">Batal</Button>
          <div class="flex items-center gap-2">
            <Button
              @click="submitApprove"
              :disabled="isSubmitting || isLookupLoading"
              class="rounded-xl font-black bg-primary hover:bg-primary/90 text-white gap-2 shadow-lg shadow-primary/20 px-6"
            >
              {{ isSubmitting ? 'Menerbitkan...' : 'Setujui & Terbitkan E-Slip' }}
            </Button>
          </div>
        </div>
      </DialogContent>
    </Dialog>

    <!-- Modal Reject -->
    <Dialog :open="isRejectModalOpen" @update:open="isRejectModalOpen = $event">
      <DialogContent class="sm:max-w-md rounded-[2rem] p-6">
        <DialogHeader>
          <DialogTitle class="text-lg font-black text-red-600">Tolak Permohonan Slip Gaji</DialogTitle>
          <DialogDescription class="text-xs text-slate-500">
            Berikan alasan penolakan agar karyawan mengetahui kendala pengajuan.
          </DialogDescription>
        </DialogHeader>

        <div class="space-y-3 py-3">
          <label class="text-xs font-bold text-slate-700">Alasan Penolakan</label>
          <textarea
            v-model="rejectReason"
            rows="3"
            placeholder="Contoh: Periode gaji belum tutup buku, harap ajukan ulang tanggal 25..."
            class="w-full rounded-xl border border-slate-200 p-3 text-xs focus:outline-none focus:ring-1 focus:ring-primary"
          ></textarea>
        </div>

        <DialogFooter class="flex gap-2">
          <Button variant="outline" @click="isRejectModalOpen = false" class="rounded-xl">Batal</Button>
          <Button variant="destructive" @click="submitReject" class="rounded-xl font-bold">
            Tolak Permohonan
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>

    <!-- Modal Detail & Cetak Slip Gaji -->
    <Dialog :open="isDetailModalOpen" @update:open="isDetailModalOpen = $event">
      <DialogContent class="sm:max-w-2xl rounded-[2.5rem] p-6 max-h-[90vh] flex flex-col bg-white">
        <div class="flex items-center justify-between border-b pb-4 shrink-0">
          <div>
            <h3 class="text-xl font-black text-slate-900">Rincian Slip Gaji Digital</h3>
            <p class="text-xs text-slate-500 mt-0.5">
              Periode {{ selectedDetailRequest?.period_month }}/{{ selectedDetailRequest?.period_year }}
            </p>
          </div>
          <div class="flex items-center gap-2">
            <Button size="sm" @click="printPayslip" class="rounded-xl font-bold bg-primary hover:bg-primary/90 text-white gap-1.5">
              <Printer class="w-4 h-4" /> Cetak Slip
            </Button>
            <button @click="isDetailModalOpen = false" class="p-2 rounded-full hover:bg-slate-100">
              <X class="w-5 h-5 text-slate-400" />
            </button>
          </div>
        </div>

        <div class="flex-1 overflow-y-auto py-4">
          <div id="printable-payslip-area" class="space-y-6 p-4">
            <!-- Kop Slip -->
            <div class="border-b-2 border-slate-900 pb-4 flex justify-between items-start">
              <div>
                <h2 class="text-lg font-black text-slate-900">{{ companyName }}</h2>
                <p class="text-[11px] text-slate-500">{{ companyAddress }}</p>
                <p class="text-[11px] text-slate-500"></p>
              </div>
              <div class="text-right">
                <span class="text-xs font-black uppercase tracking-wider px-3 py-1 bg-slate-900 text-white rounded-md">
                  SLIP GAJI RESMI
                </span>
                <p class="text-xs font-bold text-slate-700 mt-2">
                  Periode: {{ parsedPayslipData(selectedDetailRequest)?.period_label || `${selectedDetailRequest?.period_month}/${selectedDetailRequest?.period_year}` }}
                </p>
              </div>
            </div>

            <!-- Employee Info Grid -->
            <div class="grid grid-cols-2 gap-4 text-xs bg-slate-50 p-4 rounded-xl border border-slate-100">
              <div>
                <p class="text-slate-400 font-bold">NAMA KARYAWAN</p>
                <p class="font-black text-slate-900 text-sm mt-0.5">
                  {{ parsedPayslipData(selectedDetailRequest)?.employee?.name || selectedDetailRequest?.user?.employee?.first_name || '-' }}
                </p>
              </div>
              <div>
                <p class="text-slate-400 font-bold">NOMOR INDUK KARYAWAN (NIK)</p>
                <p class="font-bold text-slate-800 mt-0.5">
                  {{ parsedPayslipData(selectedDetailRequest)?.employee?.employee_id_number || selectedDetailRequest?.user?.employee?.employee_id_number || '-' }}
                </p>
              </div>
              <div>
                <p class="text-slate-400 font-bold">JABATAN & DEPARTEMEN</p>
                <p class="font-bold text-slate-800 mt-0.5">
                  {{ parsedPayslipData(selectedDetailRequest)?.employee?.job_position || '-' }} ({{ parsedPayslipData(selectedDetailRequest)?.employee?.department || '-' }})
                </p>
              </div>
              <div>
                <p class="text-slate-400 font-bold">REKENING BANK</p>
                <p class="font-bold text-slate-800 mt-0.5">
                  {{ parsedPayslipData(selectedDetailRequest)?.employee?.bank_name || 'BCA' }} - {{ parsedPayslipData(selectedDetailRequest)?.employee?.bank_account_number || '-' }}
                </p>
              </div>
            </div>

            <!-- Itemized Table -->
            <div class="space-y-4 text-xs">
              <div v-if="parsedPayslipData(selectedDetailRequest)">
                <div class="grid grid-cols-2 gap-6">
                  <!-- Earnings -->
                  <div class="space-y-2">
                    <p class="font-black uppercase text-slate-900 border-b pb-1">Pendapatan</p>
                    <div class="flex justify-between py-1">
                      <span>Gaji Pokok</span>
                      <span class="font-bold">{{ formatCurrency(parsedPayslipData(selectedDetailRequest).basic_salary) }}</span>
                    </div>
                    <div
                      v-for="(item, idx) in parsedPayslipData(selectedDetailRequest).allowances"
                      :key="'det-al-' + idx"
                      class="flex justify-between py-1"
                    >
                      <span>{{ item.name }}</span>
                      <span class="font-bold">{{ formatCurrency(item.amount) }}</span>
                    </div>
                    <div class="flex justify-between font-black border-t pt-2 text-emerald-700">
                      <span>Total Pendapatan</span>
                      <span>{{ formatCurrency(parsedPayslipData(selectedDetailRequest).total_earning) }}</span>
                    </div>
                  </div>

                  <!-- Deductions -->
                  <div class="space-y-2">
                    <p class="font-black uppercase text-slate-900 border-b pb-1">Potongan</p>
                    <div
                      v-for="(item, idx) in parsedPayslipData(selectedDetailRequest).deductions"
                      :key="'det-ded-' + idx"
                      class="flex justify-between py-1 text-slate-600"
                    >
                      <span>{{ item.name }}</span>
                      <span class="font-bold text-red-600">- {{ formatCurrency(item.amount) }}</span>
                    </div>
                    <div class="flex justify-between font-black border-t pt-2 text-red-700">
                      <span>Total Potongan</span>
                      <span>- {{ formatCurrency(parsedPayslipData(selectedDetailRequest).total_deduction) }}</span>
                    </div>
                  </div>
                </div>
              </div>

              <!-- Net Pay Banner -->
              <div class="bg-slate-100 p-4 rounded-xl flex justify-between items-center text-sm font-black border border-slate-200 mt-4">
                <span class="uppercase tracking-wider text-slate-700">Gaji Bersih Diterima (Take Home Pay)</span>
                <span class="text-xl text-slate-900">{{ formatCurrency(selectedDetailRequest?.nominal) }}</span>
              </div>
            </div>

            <!-- Signature Footer for Print -->
            <div class="grid grid-cols-2 pt-8 text-center text-xs">
              <div class="space-y-12">
                <p class="font-semibold text-slate-500">Penerima,</p>
                <p class="font-bold text-slate-800">
                  ({{ parsedPayslipData(selectedDetailRequest)?.employee?.name || selectedDetailRequest?.user?.employee?.first_name || 'Karyawan' }})
                </p>
              </div>
              <div class="space-y-12">
                <p class="font-semibold text-slate-500">Manager HRD,</p>
                <p class="font-bold text-slate-800">( Andi Muhammad Santoso )</p>
              </div>
            </div>
          </div>
        </div>

        <DialogFooter class="border-t pt-4">
          <Button variant="outline" @click="isDetailModalOpen = false" class="rounded-xl">Tutup</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  </div>
</template>
