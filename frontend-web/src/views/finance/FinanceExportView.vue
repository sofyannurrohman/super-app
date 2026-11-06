<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { toast } from 'vue-sonner'
import apiClient from '@/api/axios'
import {
  FileSpreadsheet,
  Download,
  Building2,
  Calendar,
  FileText,
  AlertCircle,
} from 'lucide-vue-next'

// State
const isLoading = ref(false)
const companies = ref<any[]>([])
const selectedCompanyId = ref('')
const selectedMonth = ref(new Date().getMonth() + 1)
const selectedYear = ref(new Date().getFullYear())

const months = [
  { value: 1, name: 'Januari' },
  { value: 2, name: 'Februari' },
  { value: 3, name: 'Maret' },
  { value: 4, name: 'April' },
  { value: 5, name: 'Mei' },
  { value: 6, name: 'Juni' },
  { value: 7, name: 'Juli' },
  { value: 8, name: 'Agustus' },
  { value: 9, name: 'September' },
  { value: 10, name: 'Oktober' },
  { value: 11, name: 'November' },
  { value: 12, name: 'Desember' },
]

const years = [2025, 2026, 2027]

const fetchCompanies = async () => {
  try {
    const res = await apiClient.get('/companies')
    if (res.data?.data) {
      companies.value = res.data.data.map((c: any) => ({
        id: c.id,
        name: c.name,
      }))
      if (companies.value.length > 0) {
        selectedCompanyId.value = companies.value[0].id
      }
    }
  } catch (e: any) {
    console.error(e)
    toast.error('Gagal mengambil daftar perusahaan')
  }
}

const handleExport = async () => {
  if (!selectedCompanyId.value) {
    toast.warning('Silakan pilih perusahaan terlebih dahulu')
    return
  }

  isLoading.value = true
  try {
    const res = await apiClient.get('/finance/export/csv', {
      params: {
        company_id: selectedCompanyId.value,
        month: selectedMonth.value,
        year: selectedYear.value,
      },
      responseType: 'blob',
    })

    // Read filename from response headers if present
    let filename = `keuangan_ekspor_${selectedYear.value}_${selectedMonth.value}.csv`
    const disposition = res.headers['content-disposition']
    if (disposition && disposition.indexOf('attachment') !== -1) {
      const filenameRegex = /filename[^;=\n]*=((['"]).*?\2|[^;\n]*)/
      const matches = filenameRegex.exec(disposition)
      if (matches != null && matches[1]) {
        filename = matches[1].replace(/['"]/g, '')
      }
    }

    // Stream download
    const url = window.URL.createObjectURL(new Blob([res.data]))
    const link = document.createElement('a')
    link.href = url
    link.setAttribute('download', filename)
    document.body.appendChild(link)
    link.click()

    // Clean up
    document.body.removeChild(link)
    window.URL.revokeObjectURL(url)

    toast.success('Spreadsheet laporan keuangan berhasil diekspor!')
  } catch (e: any) {
    console.error(e)
    toast.error('Gagal mengekspor laporan keuangan. Pastikan data transaksi sudah lengkap.')
  } finally {
    isLoading.value = false
  }
}

onMounted(() => {
  fetchCompanies()
})
</script>

<template>
  <div class="space-y-8 pb-10">
    <!-- Header -->
    <div>
      <h1 class="text-3xl font-extrabold text-slate-900 tracking-tight">Ekspor Data Finansial</h1>
      <p class="text-[15px] font-medium text-slate-500 mt-1.5 flex items-center gap-2">
        <FileSpreadsheet class="w-4 h-4 text-primary" />
        Tarik laporan keuangan konsolidasi berformat CSV yang terbagi atas Pendapatan, Gaji Payroll,
        dan Operational OPEX.
      </p>
    </div>

    <!-- Export Panel Card -->
    <div
      class="max-w-2xl bg-white rounded-3xl border border-slate-200/60 shadow-[0_4px_25px_rgba(0,0,0,0.03)] p-8 relative overflow-hidden group"
    >
      <!-- Section Decoration -->
      <div
        class="absolute -right-6 -bottom-6 w-32 h-32 bg-primary/5 rounded-full blur-3xl opacity-0 group-hover:opacity-100 transition-opacity duration-500"
      ></div>

      <div class="relative z-10 space-y-6">
        <div class="flex items-center gap-3.5 pb-5 border-b border-slate-100">
          <div class="p-3 bg-primary/10 text-primary rounded-2xl">
            <FileSpreadsheet class="w-6 h-6" />
          </div>
          <div>
            <h3 class="text-base font-extrabold text-slate-900">Form Penarikan Laporan Keuangan</h3>
            <p class="text-xs text-slate-400 font-bold uppercase mt-0.5">
              SISTEM INTEGRASI KONSOLIDASI
            </p>
          </div>
        </div>

        <!-- Form Fields -->
        <div class="space-y-5">
          <!-- Company Dropdown -->
          <div class="space-y-2">
            <label
              class="text-xs font-black text-slate-400 uppercase tracking-wider flex items-center gap-1.5"
            >
              <Building2 class="w-3.5 h-3.5" />
              PILIH PERUSAHAAN (PT)
            </label>
            <select
              v-model="selectedCompanyId"
              class="w-full bg-slate-50 border border-slate-200 rounded-2xl p-4 text-sm font-bold text-slate-700 shadow-sm focus:outline-none focus:ring-2 focus:ring-primary/20 transition-all"
            >
              <option value="" disabled>Pilih Perusahaan...</option>
              <option v-for="c in companies" :key="c.id" :value="c.id">{{ c.name }}</option>
            </select>
          </div>

          <!-- Period Dropdowns -->
          <div class="grid grid-cols-1 sm:grid-cols-2 gap-5">
            <!-- Month selector -->
            <div class="space-y-2">
              <label
                class="text-xs font-black text-slate-400 uppercase tracking-wider flex items-center gap-1.5"
              >
                <Calendar class="w-3.5 h-3.5" />
                BULAN PERIODE
              </label>
              <select
                v-model="selectedMonth"
                class="w-full bg-slate-50 border border-slate-200 rounded-2xl p-4 text-sm font-bold text-slate-700 shadow-sm focus:outline-none focus:ring-2 focus:ring-primary/20 transition-all"
              >
                <option v-for="m in months" :key="m.value" :value="m.value">{{ m.name }}</option>
              </select>
            </div>

            <!-- Year Selector -->
            <div class="space-y-2">
              <label
                class="text-xs font-black text-slate-400 uppercase tracking-wider flex items-center gap-1.5"
              >
                <Calendar class="w-3.5 h-3.5" />
                TAHUN PERIODE
              </label>
              <select
                v-model="selectedYear"
                class="w-full bg-slate-50 border border-slate-200 rounded-2xl p-4 text-sm font-bold text-slate-700 shadow-sm focus:outline-none focus:ring-2 focus:ring-primary/20 transition-all"
              >
                <option v-for="y in years" :key="y" :value="y">{{ y }}</option>
              </select>
            </div>
          </div>
        </div>

        <!-- Information Info Box -->
        <div
          class="bg-slate-50 border border-slate-100 p-4.5 rounded-2xl flex gap-3 text-xs font-semibold text-slate-500 leading-relaxed"
        >
          <AlertCircle class="w-5 h-5 text-primary shrink-0 mt-0.5" />
          <div>
            <p class="font-extrabold text-slate-700">Laporan Keuangan CSV berisi 3 bagian utama:</p>
            <ul class="list-disc pl-4 mt-2 space-y-1 font-medium">
              <li>
                <strong>BAGIAN 1: PENDAPATAN (SALES REVENUE)</strong> — Seluruh penjualan
                lunas/terverifikasi.
              </li>
              <li>
                <strong>BAGIAN 2: BEBAN GAJI KARYAWAN (PAYROLL)</strong> — Seluruh slip gaji
                karyawan yang dibayarkan.
              </li>
              <li>
                <strong>BAGIAN 3: BEBAN OPERASIONAL REIMBURSEMENT</strong> — Pengeluaran operasional
                staf disetujui.
              </li>
            </ul>
          </div>
        </div>

        <!-- Submit Button -->
        <button
          @click="handleExport"
          :disabled="isLoading"
          class="w-full bg-slate-900 text-white hover:bg-slate-800 disabled:opacity-50 py-4.5 rounded-2xl font-black text-sm transition-all flex items-center justify-center gap-2 shadow-lg shadow-slate-900/15 active:scale-[0.98]"
        >
          <span
            v-if="isLoading"
            class="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin"
          ></span>
          <Download v-else class="w-4.5 h-4.5" />
          <span>UNDUH SPREADSHEET KEUANGAN (CSV)</span>
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
button {
  -webkit-tap-highlight-color: transparent;
}
</style>
