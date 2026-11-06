<script setup lang="ts">
import { ref, h, onMounted, watch, computed } from 'vue'
import { toast } from 'vue-sonner'
import {
  Trash2,
  FileSearch,
  Download,
  Play,
  Calendar,
  Banknote,
  CheckCircle2,
  Search,
  Filter,
  X,
  TrendingUp,
  Wallet,
  ShieldCheck,
  Settings2,
  Heart,
  Percent,
  ShieldAlert,
  Calculator,
  Plus,
  Pencil,
  Timer,
  Briefcase,
  Building2,
  ChevronDown,
  AlertTriangle,
  Loader2,
  Check,
  Sparkles,
  FileSpreadsheet,
} from 'lucide-vue-next'
import axios from '@/api/axios'
import { useMasterDataStore } from '@/stores/masterData'
import { useAuthStore } from '@/stores/auth'
import DataTable from '@/components/DataTable.vue'
import PayslipRequestTab from '@/components/PayslipRequestTab.vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'

// Main Tabs State: only 'history' or 'requests'
const activeMainTab = ref<'history' | 'requests'>('history')
const isSettingsModalOpen = ref(false)

const masterDataStore = useMasterDataStore()
const authStore = useAuthStore()

watch(() => masterDataStore.selectedBranchId, () => {
  if (activeMainTab.value === 'history') {
    fetchPayrollRuns()
  }
})

// --- PAYROLL HISTORY SECTION ---
const payrollRuns = ref<any[]>([])
const isLoading = ref(true)
const isModalOpen = ref(false)
const isSubmitting = ref(false)

// Filter & Search State
const searchQuery = ref('')
const statusFilter = ref<string>('ALL')

const filteredPayrollRuns = computed(() => {
  return payrollRuns.value.filter((run) => {
    // Status filter
    if (statusFilter.value !== 'ALL' && run.status !== statusFilter.value) {
      return false
    }
    // Search query
    if (searchQuery.value.trim()) {
      const q = searchQuery.value.toLowerCase().trim()
      const matchId = (run.id || '').toLowerCase().includes(q)
      const matchBranch = (run.branchName || '').toLowerCase().includes(q)
      const matchPeriod = `${run.periodStart} ${run.periodEnd}`.toLowerCase().includes(q)
      if (!matchId && !matchBranch && !matchPeriod) {
        return false
      }
    }
    return true
  })
})

// KPI Stats
const stats = computed(() => {
  const totalRuns = payrollRuns.value.length
  const draftRuns = payrollRuns.value.filter((r) => r.status === 'DRAFT').length
  const approvedRuns = payrollRuns.value.filter((r) => r.status === 'APPROVED' || r.status === 'PROCESSED').length
  const totalPayout = payrollRuns.value.reduce((acc, r) => acc + (Number(r.totalPayout) || 0), 0)

  return {
    totalRuns,
    draftRuns,
    approvedRuns,
    totalPayout,
  }
})

// Details State
const isDetailModalOpen = ref(false)
const selectedRun = ref<any>(null)
const payslips = ref<any[]>([])
const isDetailLoading = ref(false)
const expandedSlipId = ref<string | null>(null)
const isApproving = ref(false)

// Adjustment State
const isAdjustModalOpen = ref(false)
const adjustingSlip = ref<any>(null)
const adjustItems = ref<Array<{ component_name: string; amount: number; type: string }>>([])
const isAdjustSubmitting = ref(false)

const toggleSlip = (id: string) => {
  expandedSlipId.value = expandedSlipId.value === id ? null : id
}

const newRun = ref({
  branchId: 'ALL_BRANCHES',
  periodStart: '',
  periodEnd: '',
  paymentSchedule: '',
})

const fetchPayrollRuns = async () => {
  isLoading.value = true
  try {
    let url = '/payroll/all?limit=50'
    if (masterDataStore.selectedBranchId && masterDataStore.selectedBranchId !== 'ALL_BRANCHES') {
      url += `&branch_id=${masterDataStore.selectedBranchId}`
    }
    const res = await axios.get(url)
    if (res.data?.data) {
      payrollRuns.value = res.data.data.map((r: any) => ({
        id: r.id || r.ID,
        branchId: r.branch_id || r.BranchID,
        branchName: r.branch?.name || r.Branch?.Name || 'Semua Cabang',
        periodStart:
          r.period_start || r.PeriodStart
            ? new Date(r.period_start || r.PeriodStart).toISOString().split('T')[0]
            : '',
        periodEnd:
          r.period_end || r.PeriodEnd
            ? new Date(r.period_end || r.PeriodEnd).toISOString().split('T')[0]
            : '',
        paymentSchedule:
          r.payment_schedule || r.PaymentSchedule
            ? new Date(r.payment_schedule || r.PaymentSchedule).toISOString().split('T')[0]
            : '',
        status: r.status || r.Status,
        totalPayout: r.total_payout || r.TotalPayout || 0,
      }))
    }
  } catch (error) {
    console.error('Failed to fetch payroll history', error)
  } finally {
    isLoading.value = false
  }
}

const openDetails = async (run: any) => {
  selectedRun.value = run
  isDetailModalOpen.value = true
  isDetailLoading.value = true
  try {
    const res = await axios.get(`/payroll/${run.id}/details`)
    payslips.value = res.data.data || []
  } catch (e) {
    toast.error('Gagal memuat rincian slip gaji.')
  } finally {
    isDetailLoading.value = false
  }
}

const openAddModal = () => {
  newRun.value = {
    branchId: masterDataStore.selectedBranchId || 'ALL_BRANCHES',
    periodStart: '',
    periodEnd: '',
    paymentSchedule: '',
  }
  isModalOpen.value = true
}

const approvePayrollRun = async (runId: string) => {
  if (!confirm('Setujui (Approve) batch penggajian ini? Setelah disetujui, nominal gaji difinalisasi dan siap ditransfer.')) return
  isApproving.value = true
  try {
    const res = await axios.put(`/payroll/${runId}/approve`)
    toast.success('Batch penggajian berhasil disetujui!')
    if (selectedRun.value && selectedRun.value.id === runId) {
      selectedRun.value.status = 'APPROVED'
      if (res.data?.data?.total_payout) {
        selectedRun.value.totalPayout = res.data.data.total_payout
      }
    }
    fetchPayrollRuns()
  } catch (e: any) {
    toast.error('Gagal menyetujui batch: ' + (e.response?.data?.message || e.message))
  } finally {
    isApproving.value = false
  }
}

const openAdjustModal = (slip: any) => {
  adjustingSlip.value = slip
  adjustItems.value = (slip.items || []).map((it: any) => ({
    component_name: it.component_name,
    amount: it.amount,
    type: it.type,
  }))
  // Jika belum ada item
  if (adjustItems.value.length === 0) {
    adjustItems.value.push({ component_name: 'Gaji Pokok', amount: slip.basic_salary || 0, type: 'EARNING' })
  }
  isAdjustModalOpen.value = true
}

const addAdjustItem = () => {
  adjustItems.value.push({
    component_name: 'Penyesuaian Bonus / Koreksi',
    amount: 0,
    type: 'EARNING',
  })
}

const removeAdjustItem = (index: number) => {
  adjustItems.value.splice(index, 1)
}

const saveAdjustment = async () => {
  if (!adjustingSlip.value) return
  isAdjustSubmitting.value = true
  try {
    const res = await axios.put(`/payroll/payslips/${adjustingSlip.value.id}/adjust`, {
      items: adjustItems.value.map((i) => ({
        component_name: i.component_name,
        amount: Number(i.amount) || 0,
        type: i.type,
      })),
    })
    toast.success('Penyesuaian slip gaji berhasil disimpan!')
    isAdjustModalOpen.value = false

    // Refresh detail slips
    if (selectedRun.value) {
      const updatedDetails = await axios.get(`/payroll/${selectedRun.value.id}/details`)
      payslips.value = updatedDetails.data.data || []
      fetchPayrollRuns()
    }
  } catch (e: any) {
    toast.error('Gagal menyimpan penyesuaian: ' + (e.response?.data?.message || e.message))
  } finally {
    isAdjustSubmitting.value = false
  }
}

const exportBankCSV = async (runId: string) => {
  try {
    const res = await axios.get(`/payroll/${runId}/export`, { responseType: 'blob' })
    const url = window.URL.createObjectURL(new Blob([res.data]))
    const link = document.createElement('a')
    link.href = url
    link.setAttribute('download', `payout_bank_${runId.substring(0, 8)}.csv`)
    document.body.appendChild(link)
    link.click()
    if (link.parentNode) link.parentNode.removeChild(link)
    toast.success('File ekspor bank siap diunduh.')
  } catch (e: any) {
    toast.error('Gagal mengunduh CSV.')
  }
}

// --- TRANSFER FILE MODAL ---
interface BankSummaryItem {
  bank_name: string
  employee_count: number
  total_transfer: number
}
interface TransferSummary {
  payroll_run_id: string
  period_start: string
  period_end: string
  payment_schedule: string
  banks: BankSummaryItem[]
  total_employees: number
  total_transfer: number
  missing_bank: number
}

const isTransferModalOpen = ref(false)
const transferSummary = ref<TransferSummary | null>(null)
const isTransferLoading = ref(false)
const transferRun = ref<any>(null)
const downloadingKey = ref<string | null>(null)

const formatIDR = (amount: number) =>
  new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    minimumFractionDigits: 0,
  }).format(amount)

const bankLabel = (name: string) => {
  const upper = name.toUpperCase()
  if (upper === 'BCA') return '🏦 BCA'
  if (upper === 'MANDIRI') return '🏦 Mandiri'
  if (upper === 'BNI') return '🏦 BNI'
  if (upper === 'BRI') return '🏦 BRI'
  return `🏦 ${name}`
}

const openTransferModal = async (run: any) => {
  transferRun.value = run
  isTransferModalOpen.value = true
  isTransferLoading.value = true
  transferSummary.value = null
  try {
    const res = await axios.get(`/payroll/${run.id}/transfer-summary`)
    transferSummary.value = res.data?.data || res.data
  } catch (e: any) {
    toast.error('Gagal memuat ringkasan transfer: ' + (e.response?.data?.message || e.message))
    isTransferModalOpen.value = false
  } finally {
    isTransferLoading.value = false
  }
}

const downloadTransferFile = async (runId: string, bank: string, format: string) => {
  const key = `${bank}-${format}`
  downloadingKey.value = key
  try {
    const token = localStorage.getItem('token')
    const baseURL = `${import.meta.env.VITE_API_BASE_URL}/v1`
    const url = `${baseURL}/payroll/${runId}/transfer-file?bank=${bank}&format=${format}`
    const response = await fetch(url, {
      headers: { Authorization: `Bearer ${token}` },
    })
    if (!response.ok) throw new Error(`HTTP ${response.status}`)
    const blob = await response.blob()
    const contentDisposition = response.headers.get('content-disposition') || ''
    const filenameMatch = contentDisposition.match(/filename[^;=\n]*=(["']?)([^"'\n;]+)\1/)
    const filename =
      filenameMatch?.[2] ||
      `transfer_${bank}_${format}_${runId.substring(0, 8)}.${format === 'GENERIC' ? 'csv' : 'txt'}`
    const objectUrl = URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = objectUrl
    link.download = filename
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    URL.revokeObjectURL(objectUrl)
    toast.success(`File ${bankLabel(bank)} (${format}) berhasil diunduh.`)
  } catch (e: any) {
    toast.error('Gagal mengunduh file transfer: ' + e.message)
  } finally {
    downloadingKey.value = null
  }
}

const downloadTransferExcel = async (runId: string, bank: string, format: string) => {
  const key = `${bank}-${format}-EXCEL`
  downloadingKey.value = key
  try {
    const token = localStorage.getItem('token')
    const baseURL = `${import.meta.env.VITE_API_BASE_URL}/v1`
    const url = `${baseURL}/payroll/${runId}/export-excel?bank=${bank}&format=${format}`
    const response = await fetch(url, {
      headers: { Authorization: `Bearer ${token}` },
    })
    if (!response.ok) throw new Error(`HTTP ${response.status}`)
    const blob = await response.blob()
    const contentDisposition = response.headers.get('content-disposition') || ''
    const filenameMatch = contentDisposition.match(/filename[^;=\n]*=(["']?)([^"'\n;]+)\1/)
    const filename =
      filenameMatch?.[2] ||
      `CMS_Payroll_${bank}_${format}_${runId.substring(0, 8)}.xlsx`
    const objectUrl = URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = objectUrl
    link.download = filename
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    URL.revokeObjectURL(objectUrl)
    toast.success(`File Excel CMS ${bankLabel(bank)} (${format}) berhasil diunduh.`)
  } catch (e: any) {
    toast.error('Gagal mengunduh Excel CMS: ' + e.message)
  } finally {
    downloadingKey.value = null
  }
}

const deletePayrollRun = async (runId: string) => {
  if (!confirm('Hapus siklus ini? Tindakan ini akan menghapus semua slip gaji terkait.')) return
  try {
    await axios.delete(`/payroll/${runId}`)
    toast.success('Siklus gaji dihapus.')
    fetchPayrollRuns()
  } catch (e: any) {
    toast.error('Gagal menghapus.')
  }
}

const generatePayroll = async () => {
  isSubmitting.value = true
  try {
    await axios.post('/payroll/generate', {
      branch_id: newRun.value.branchId === 'ALL_BRANCHES' ? undefined : newRun.value.branchId,
      period_start: newRun.value.periodStart,
      period_end: newRun.value.periodEnd,
      payment_schedule: newRun.value.paymentSchedule,
    })
    toast.success('Batch Payroll Berhasil Diproses (Status: DRAFT)! Menunggu Review HRD.')
    isModalOpen.value = false
    fetchPayrollRuns()
  } catch (e: any) {
    toast.error('Gagal memproses payroll: ' + (e.response?.data?.message || e.message))
  } finally {
    isSubmitting.value = false
  }
}

const columns = [
  {
    accessorKey: 'id',
    header: 'KODE RUN',
    cell: (info: any) =>
      h(
        'span',
        { class: 'font-black text-slate-400 text-[11px] tracking-tight uppercase' },
        '#' + info.getValue().substring(0, 8),
      ),
  },
  {
    accessorKey: 'branchName',
    header: 'CABANG',
    cell: ({ row }: any) =>
      h(
        'span',
        { class: 'px-3 py-1 bg-slate-100 text-slate-700 rounded-full text-[11px] font-black tracking-wide uppercase border border-slate-200/60' },
        row.original.branchName || 'Semua Cabang',
      ),
  },
  {
    accessorKey: 'periodStart',
    id: 'period',
    header: 'PERIODE GAJI',
    cell: ({ row }: any) =>
      h('div', { class: 'flex flex-col' }, [
        h(
          'span',
          { class: 'text-slate-900 font-black text-[13px]' },
          `${row.original.periodStart} - ${row.original.periodEnd}`,
        ),
        h(
          'span',
          { class: 'text-slate-400 text-[11px] font-bold uppercase tracking-widest mt-0.5' },
          'Performance log',
        ),
      ]),
  },
  {
    accessorKey: 'paymentSchedule',
    header: 'TGL PENCAIRAN',
    cell: (info: any) =>
      h('div', { class: 'flex items-center gap-2 text-rose-600 font-black text-[13px]' }, [
        h(Calendar, { class: 'w-3.5 h-3.5' }),
        h('span', {}, info.getValue()),
      ]),
  },
  {
    accessorKey: 'totalPayout',
    header: 'TOTAL BIAYA (PAYOUT)',
    cell: ({ getValue }: any) => {
      const val = getValue() as number
      return h(
        'span',
        { class: 'font-black text-slate-900 text-[14px] tabular-nums' },
        `Rp ${val.toLocaleString('id-ID')}`,
      )
    },
  },
  {
    accessorKey: 'status',
    header: 'STATUS',
    cell: ({ getValue }: any) => {
      const val = getValue() as string
      let classes = 'bg-amber-50 text-amber-700 border-amber-200'
      let label = val
      if (val === 'DRAFT') {
        classes = 'bg-amber-50 text-amber-700 border-amber-200'
        label = 'REVIEW HRD (DRAFT)'
      } else if (val === 'APPROVED') {
        classes = 'bg-blue-50 text-blue-700 border-blue-200'
        label = 'APPROVED HRD'
      } else if (val === 'PROCESSED') {
        classes = 'bg-emerald-50 text-emerald-700 border-emerald-200'
        label = 'SELESAI (PROCESSED)'
      }
      return h(
        'span',
        {
          class: `${classes} px-3 py-1 rounded-full text-[10px] font-black uppercase border tracking-[0.15em]`,
        },
        label,
      )
    },
  },
  {
    id: 'actions',
    header: 'NAVIGASI',
    cell: ({ row }: any) => {
      const run = row.original
      const actions = []

      actions.push(
        h(
          Button,
          {
            onClick: () => openDetails(run),
            variant: 'ghost',
            size: 'sm',
            class:
              'h-9 px-3 text-slate-500 hover:bg-rose-50 hover:text-rose-600 rounded-xl font-bold gap-2 transition-all',
          },
          () => [h(FileSearch, { class: 'w-4 h-4' }), 'Rincian'],
        ),
      )

      if (run.status === 'APPROVED' || run.status === 'PROCESSED') {
        actions.push(
          h(
            Button,
            {
              onClick: () => openTransferModal(run),
              variant: 'outline',
              size: 'sm',
              class:
                'h-9 px-3 border-blue-100 bg-blue-50 text-blue-600 hover:bg-blue-100 hover:text-blue-700 rounded-xl font-bold gap-2',
            },
            () => [h(Building2, { class: 'w-4 h-4' }), 'CMS Bank Transfer'],
          ),
        )
      }

      actions.push(
        h(
          Button,
          {
            onClick: () => deletePayrollRun(run.id),
            variant: 'ghost',
            size: 'sm',
            class: 'h-9 px-2 text-slate-400 hover:bg-rose-50 hover:text-rose-600 rounded-xl',
          },
          () => h(Trash2, { class: 'w-4 h-4' }),
        ),
      )

      return h('div', { class: 'flex items-center gap-2' }, actions)
    },
    enableSorting: false,
  },
]

// --- PAYROLL SETTINGS SECTION ---
const isSettingsLoading = ref(false)
const activeSettingsTab = ref('bpjs')

const config = ref({
  id: '',
  bpjsKesCompanyPercentage: 4.0,
  bpjsKesEmployeePercentage: 1.0,
  bpjsKesMaxWageBase: 12000000,
  jhtCompanyPercentage: 3.7,
  jhtEmployeePercentage: 2.0,
  jpCompanyPercentage: 2.0,
  jpEmployeePercentage: 1.0,
  jpMaxWageBase: 10042300,
  jkkCompanyPercentage: 0.24,
  jkmCompanyPercentage: 0.3,
  ptkpBaseTK0: 54000000,
  absentDeduction: 100000,
  kpiBonusAmount: 500000,
})

const fetchConfig = async () => {
  isSettingsLoading.value = true
  try {
    const res = await axios.get('/payroll/config')
    if (res.data?.data) {
      const data = res.data.data
      config.value = {
        id: data.id || data.ID,
        bpjsKesCompanyPercentage: data.bpjs_kes_company_percentage || data.BPJSKesCompanyPercentage,
        bpjsKesEmployeePercentage:
          data.bpjs_kes_employee_percentage || data.BPJSKesEmployeePercentage,
        bpjsKesMaxWageBase: data.bpjs_kes_max_wage_base || data.BPJSKesMaxWageBase,
        jhtCompanyPercentage: data.jht_company_percentage || data.JHTCompanyPercentage,
        jhtEmployeePercentage: data.jht_employee_percentage || data.JHTEmployeePercentage,
        jpCompanyPercentage: data.jp_company_percentage || data.JPCompanyPercentage,
        jpEmployeePercentage: data.jp_employee_percentage || data.JPEmployeePercentage,
        jpMaxWageBase: data.jp_max_wage_base || data.JPMaxWageBase,
        jkkCompanyPercentage: data.jkk_company_percentage || data.JKKCompanyPercentage,
        jkmCompanyPercentage: data.jkm_company_percentage || data.JKMCompanyPercentage,
        ptkpBaseTK0: data.ptkp_base_tk0 || data.PtkpBaseTK0,
        absentDeduction: data.absent_deduction || data.AbsentDeduction || 100000,
        kpiBonusAmount: data.kpi_bonus_amount || data.KPIBonusAmount || 500000,
      }
    }
  } catch (error) {
    console.error('Failed to fetch payroll config', error)
  } finally {
    isSettingsLoading.value = false
  }
}

const saveConfig = async () => {
  isSettingsLoading.value = true
  try {
    const payload = {
      id: config.value.id || undefined,
      jht_company_percentage: Number(config.value.jhtCompanyPercentage),
      jht_employee_percentage: Number(config.value.jhtEmployeePercentage),
      jp_company_percentage: Number(config.value.jpCompanyPercentage),
      jp_employee_percentage: Number(config.value.jpEmployeePercentage),
      jp_max_wage_base: Number(config.value.jpMaxWageBase),
      jkk_company_percentage: Number(config.value.jkkCompanyPercentage),
      jkm_company_percentage: Number(config.value.jkmCompanyPercentage),
      bpjs_kes_company_percentage: Number(config.value.bpjsKesCompanyPercentage),
      bpjs_kes_employee_percentage: Number(config.value.bpjsKesEmployeePercentage),
      bpjs_kes_max_wage_base: Number(config.value.bpjsKesMaxWageBase),
      ptkp_base_tk0: Number(config.value.ptkpBaseTK0),
      absent_deduction: Number(config.value.absentDeduction),
      kpi_bonus_amount: Number(config.value.kpiBonusAmount),
    }
    await axios.put('/payroll/config', payload)
    toast.success('Konfigurasi Penggajian berhasil diperbarui!')
    fetchConfig()
  } catch (error: any) {
    toast.error('Gagal menyimpan konfigurasi: ' + (error.response?.data?.message || error.message))
  } finally {
    isSettingsLoading.value = false
  }
}

const isCompModalOpen = ref(false)
const newComponent = ref({
  id: '',
  name: '',
  type: 'EARNING',
  isTaxable: true,
})
const isCompEditMode = ref(false)
const isCompSubmitting = ref(false)
const components = ref<any[]>([])
const isCompLoading = ref(true)

const openCompAddModal = () => {
  isCompEditMode.value = false
  newComponent.value = { id: '', name: '', type: 'EARNING', isTaxable: true }
  isCompModalOpen.value = true
}

const openCompEditModal = (component: any) => {
  isCompEditMode.value = true
  newComponent.value = {
    id: component.id,
    name: component.name || '',
    type: component.type || 'EARNING',
    isTaxable: component.is_taxable,
  }
  isCompModalOpen.value = true
}

const fetchComponents = async () => {
  isCompLoading.value = true
  try {
    const response = await axios.get('/payroll-components')
    components.value = response.data.data
  } catch (error) {
    console.error('Failed to fetch payroll components:', error)
  } finally {
    isCompLoading.value = false
  }
}

const saveComponent = async () => {
  isCompSubmitting.value = true
  try {
    if (isCompEditMode.value) {
      await axios.put(`/payroll-components/${newComponent.value.id}`, newComponent.value)
      toast.success('Komponen gaji berhasil diperbarui!')
    } else {
      await axios.post('/payroll-components', newComponent.value)
      toast.success('Komponen gaji berhasil ditambahkan!')
    }
    isCompModalOpen.value = false
    fetchComponents()
  } catch (error: any) {
    toast.error(`Gagal menyimpan komponen: ` + (error.response?.data?.error || error.message))
  } finally {
    isCompSubmitting.value = false
  }
}

const deleteComponent = async (id: string) => {
  if (!confirm('Apakah Anda yakin ingin menghapus komponen gaji ini?')) return
  try {
    await axios.delete(`/payroll-components/${id}`)
    toast.success('Komponen gaji berhasil dihapus!')
    fetchComponents()
  } catch (error: any) {
    toast.error('Gagal menghapus komponen.')
  }
}

const componentColumns = [
  {
    accessorKey: 'name',
    header: 'NAMA KOMPONEN',
    cell: (info: any) =>
      h('span', { class: 'font-black text-slate-800 text-[13px]' }, info.getValue() || '-'),
  },
  {
    accessorKey: 'type',
    header: 'TIPE',
    cell: ({ getValue }: any) => {
      const val = (getValue() as string) || 'Unknown'
      let classes = 'bg-slate-50 text-slate-400 border-slate-100'
      if (val === 'EARNING') classes = 'bg-emerald-50 text-emerald-600 border-emerald-100'
      else if (val === 'DEDUCTION') classes = 'bg-rose-50 text-rose-600 border-rose-100'
      return h(
        'span',
        {
          class: `${classes} px-3 py-1 rounded-full text-[10px] font-black uppercase border tracking-widest`,
        },
        val,
      )
    },
  },
  {
    accessorKey: 'is_taxable',
    header: 'KENA PAJAK',
    cell: (info: any) =>
      h(
        'span',
        { class: 'text-slate-500 font-bold text-[12px]' },
        info.getValue() ? 'Ya' : 'Tidak',
      ),
  },
  {
    id: 'actions',
    header: 'AKSI',
    cell: ({ row }: any) => {
      const component = row.original
      return h('div', { class: 'flex items-center gap-2' }, [
        h(
          Button,
          {
            variant: 'ghost',
            size: 'sm',
            class: 'h-9 px-2 text-rose-600 hover:bg-rose-50 rounded-xl',
            onClick: () => openCompEditModal(component),
          },
          () => h(Pencil, { class: 'w-4 h-4' }),
        ),
        h(
          Button,
          {
            variant: 'ghost',
            size: 'sm',
            class: 'h-9 px-2 text-slate-400 hover:bg-slate-50 rounded-xl',
            onClick: () => deleteComponent(component.id),
          },
          () => h(Trash2, { class: 'w-4 h-4' }),
        ),
      ])
    },
  },
]

onMounted(() => {
  fetchPayrollRuns()
  fetchConfig()
  fetchComponents()
})
</script>

<template>
  <div class="space-y-6 pb-20 animate-in fade-in duration-500">
    <!-- Header Page (Aligned with Reimbursement & other modules) -->
    <div class="flex flex-col md:flex-row md:items-center justify-between gap-4">
      <div>
        <div class="flex items-center gap-2">
          <h1 class="text-2xl font-black text-slate-900 tracking-tight">
            Manajemen Siklus Penggajian
          </h1>
          <span class="px-2.5 py-0.5 bg-rose-50 text-rose-600 text-xs font-black rounded-full border border-rose-200">
            FINANCE & PAYROLL
          </span>
        </div>
        <p class="text-xs md:text-sm font-medium text-slate-500 mt-1">
          Pusat pengelolaan batch penggajian, peninjauan approval HRD, serta ekspor mass transfer CMS perbankan.
        </p>
      </div>

      <!-- Action Buttons -->
      <div class="flex items-center gap-2.5">
        <Button
          variant="outline"
          @click="isSettingsModalOpen = true"
          class="gap-2 h-11 px-4 rounded-2xl text-xs md:text-sm font-bold border-slate-200 text-slate-700 hover:text-rose-600 hover:bg-rose-50/50"
        >
          <Settings2 class="w-4 h-4" />
          Parameter BPJS & Denda
        </Button>
        <Button
          @click="openAddModal"
          class="gap-2 h-11 px-5 rounded-2xl text-xs md:text-sm font-bold bg-rose-600 hover:bg-rose-700 text-white shadow-lg shadow-rose-600/20"
        >
          <Plus class="w-4 h-4" />
          Jalankan Batch Baru
        </Button>
      </div>
    </div>

    <!-- Navigation Tabs Bar -->
    <div class="bg-white p-2 rounded-2xl border border-slate-200/80 shadow-sm flex items-center justify-between gap-4 flex-wrap">
      <div class="flex items-center p-1 bg-slate-100 rounded-xl">
        <button
          @click="activeMainTab = 'history'"
          :class="[
            'px-5 py-2 rounded-lg font-bold text-xs transition-all flex items-center gap-2',
            activeMainTab === 'history'
              ? 'bg-white text-rose-600 shadow-sm font-black'
              : 'text-slate-500 hover:text-slate-800',
          ]"
        >
          <Wallet class="w-3.5 h-3.5" />
          Riwayat Penggajian
        </button>
        <button
          @click="activeMainTab = 'requests'"
          :class="[
            'px-5 py-2 rounded-lg font-bold text-xs transition-all flex items-center gap-2',
            activeMainTab === 'requests'
              ? 'bg-white text-rose-600 shadow-sm font-black'
              : 'text-slate-500 hover:text-slate-800',
          ]"
        >
          <FileSearch class="w-3.5 h-3.5" />
          Request Slip Gaji
        </button>
      </div>

      <!-- Summary Pill of Selected Branch -->
      <div class="px-3 py-1.5 bg-slate-50 border border-slate-200 rounded-xl flex items-center gap-2 text-xs font-bold text-slate-600">
        <Building2 class="w-3.5 h-3.5 text-slate-400" />
        <span>Filter Cabang:</span>
        <span class="text-rose-600 font-black">
          {{ masterDataStore.branches.find(b => b.id === masterDataStore.selectedBranchId)?.name || 'Semua Cabang' }}
        </span>
      </div>
    </div>

    <!-- PAYROLL REQUESTS VIEW -->
    <div
      v-if="activeMainTab === 'requests'"
      class="space-y-6 animate-in slide-in-from-bottom-2 duration-500"
    >
      <PayslipRequestTab />
    </div>

    <!-- PAYROLL HISTORY VIEW -->
    <div
      v-if="activeMainTab === 'history'"
      class="space-y-6 animate-in slide-in-from-bottom-2 duration-500"
    >
      <!-- KPI Statistic Cards -->
      <div class="grid grid-cols-2 lg:grid-cols-4 gap-3 md:gap-4">
        <!-- Total Runs -->
        <div class="p-4 md:p-5 rounded-3xl bg-white border border-slate-200/80 shadow-sm flex flex-col justify-between">
          <div class="flex items-center justify-between">
            <span class="text-[10px] md:text-xs font-black text-slate-400 uppercase tracking-wider">
              Total Batch
            </span>
            <div class="w-8 h-8 rounded-xl bg-slate-100 text-slate-700 flex items-center justify-center">
              <Calendar class="w-4 h-4" />
            </div>
          </div>
          <div class="mt-3">
            <p class="text-lg md:text-2xl font-black text-slate-900 tracking-tight">
              {{ stats.totalRuns }} Batch
            </p>
            <p class="text-[10px] md:text-xs text-slate-500 mt-0.5 font-medium">
              Seluruh riwayat payroll
            </p>
          </div>
        </div>

        <!-- Menunggu Review HRD -->
        <div class="p-4 md:p-5 rounded-3xl bg-amber-50/60 border border-amber-200/80 shadow-sm flex flex-col justify-between">
          <div class="flex items-center justify-between">
            <span class="text-[10px] md:text-xs font-black text-amber-800 uppercase tracking-wider">
              Review HRD (Draft)
            </span>
            <div class="w-8 h-8 rounded-xl bg-amber-100 text-amber-700 flex items-center justify-center">
              <AlertTriangle class="w-4 h-4" />
            </div>
          </div>
          <div class="mt-3">
            <p class="text-lg md:text-2xl font-black text-amber-900 tracking-tight">
              {{ stats.draftRuns }} Batch
            </p>
            <p class="text-[10px] md:text-xs text-amber-700 mt-0.5 font-medium">
              Menunggu adjustment & approval
            </p>
          </div>
        </div>

        <!-- Disetujui / Selesai -->
        <div class="p-4 md:p-5 rounded-3xl bg-emerald-50/60 border border-emerald-200/80 shadow-sm flex flex-col justify-between">
          <div class="flex items-center justify-between">
            <span class="text-[10px] md:text-xs font-black text-emerald-800 uppercase tracking-wider">
              Siap / Selesai Transfer
            </span>
            <div class="w-8 h-8 rounded-xl bg-emerald-100 text-emerald-700 flex items-center justify-center">
              <CheckCircle2 class="w-4 h-4" />
            </div>
          </div>
          <div class="mt-3">
            <p class="text-lg md:text-2xl font-black text-emerald-950 tracking-tight">
              {{ stats.approvedRuns }} Batch
            </p>
            <p class="text-[10px] md:text-xs text-emerald-700 mt-0.5 font-medium">
              Status Approved / Processed
            </p>
          </div>
        </div>

        <!-- Total Payout -->
        <div class="p-4 md:p-5 rounded-3xl bg-slate-900 text-white shadow-sm flex flex-col justify-between">
          <div class="flex items-center justify-between">
            <span class="text-[10px] md:text-xs font-black text-slate-400 uppercase tracking-wider">
              Akumulasi Gaji
            </span>
            <div class="w-8 h-8 rounded-xl bg-white/10 text-rose-400 flex items-center justify-center">
              <Banknote class="w-4 h-4" />
            </div>
          </div>
          <div class="mt-3">
            <p class="text-lg md:text-xl font-black text-white truncate tabular-nums">
              Rp {{ stats.totalPayout.toLocaleString('id-ID') }}
            </p>
            <p class="text-[10px] md:text-xs text-slate-400 mt-0.5 font-medium">
              Total pencairan dari semua batch
            </p>
          </div>
        </div>
      </div>

      <!-- Data Table Card with Integrated Search & Status Filter -->
      <div class="bg-white rounded-3xl border border-slate-200/80 shadow-sm overflow-hidden">
        <DataTable :data="filteredPayrollRuns" :columns="columns" :isLoading="isLoading">
          <template #headerActions>
            <div class="flex flex-wrap items-center gap-3">
              <!-- Status Filter Select -->
              <div class="w-44">
                <Select v-model="statusFilter">
                  <SelectTrigger class="h-10 rounded-xl bg-slate-50 border-slate-200 text-xs font-bold">
                    <SelectValue placeholder="Semua Status" />
                  </SelectTrigger>
                  <SelectContent class="rounded-xl">
                    <SelectItem value="ALL" class="text-xs font-bold">Semua Status</SelectItem>
                    <SelectItem value="DRAFT" class="text-xs font-bold text-amber-600">Review HRD (Draft)</SelectItem>
                    <SelectItem value="APPROVED" class="text-xs font-bold text-blue-600">Approved HRD</SelectItem>
                    <SelectItem value="PROCESSED" class="text-xs font-bold text-emerald-600">Selesai (Processed)</SelectItem>
                  </SelectContent>
                </Select>
              </div>

              <!-- Search Input -->
              <div class="relative w-64 group">
                <Search
                  class="absolute left-3.5 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400 group-focus-within:text-rose-600 transition-colors"
                />
                <Input
                  v-model="searchQuery"
                  placeholder="Cari kode batch, cabang..."
                  class="pl-10 h-10 rounded-xl border-slate-200 bg-slate-50 text-xs font-semibold focus:bg-white focus:ring-1 focus:ring-rose-500"
                />
              </div>
            </div>
          </template>
        </DataTable>
      </div>
    </div>

    <!-- PAYROLL SETTINGS MODAL (Parameter Sistem & BPJS) -->
    <Dialog v-model:open="isSettingsModalOpen">
      <DialogContent
        class="sm:max-w-5xl rounded-3xl p-0 overflow-hidden border border-slate-200 shadow-2xl max-h-[92vh] flex flex-col"
      >
        <div class="space-y-6 overflow-y-auto custom-scrollbar p-6 md:p-8 bg-white flex-1">
          <!-- Settings Header -->
          <div
            class="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b border-slate-100 pb-6"
          >
            <div class="space-y-1">
              <div class="flex items-center gap-2.5">
                <div
                  class="w-9 h-9 bg-rose-50 text-rose-600 rounded-xl flex items-center justify-center border border-rose-100 shrink-0"
                >
                  <Settings2 class="w-4 h-4" />
                </div>
                <h2 class="text-xl md:text-2xl font-black text-slate-900 tracking-tight">
                  Pengaturan <span class="text-rose-600">Gaji & Kebijakan</span>
                </h2>
              </div>
              <p class="text-xs font-medium text-slate-500">
                Konfigurasi parameter tarif BPJS, potongan kehadiran, ambang batas PTKP, dan master komponen.
              </p>
            </div>

            <div class="flex items-center gap-3">
              <button
                v-if="activeSettingsTab !== 'master-components'"
                @click="saveConfig"
                :disabled="isSettingsLoading"
                class="bg-rose-600 text-white hover:bg-rose-700 font-black py-4 px-8 rounded-2xl text-xs shadow-xl shadow-rose-200 transition-all hover:translate-y-[-2px] active:translate-y-[0px] disabled:opacity-50 flex items-center justify-center gap-2 group"
              >
                <span
                  v-if="isSettingsLoading"
                  class="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin"
                ></span>
                <Wallet v-else class="w-4 h-4 group-hover:rotate-12 transition-transform" />
                <span>SIMPAN KONFIGURASI</span>
              </button>
              <button
                v-else
                @click="openCompAddModal"
                class="bg-slate-900 text-white hover:bg-slate-800 font-black py-4 px-8 rounded-2xl text-xs shadow-xl shadow-slate-200 transition-all hover:translate-y-[-2px] active:translate-y-[0px] flex items-center justify-center gap-2 group"
              >
                <Plus class="w-4 h-4 group-hover:scale-125 transition-transform" />
                <span>TAMBAH KOMPONEN</span>
              </button>
            </div>
          </div>

      <div class="grid grid-cols-1 lg:grid-cols-12 gap-10 items-start">
        <!-- Settings Sidebar Navigation -->
        <div class="lg:col-span-3 flex flex-col gap-3 shrink-0 sticky top-10">
          <button
            @click="activeSettingsTab = 'bpjs'"
            :class="[
              'text-left px-6 py-6 rounded-[1.75rem] text-[13px] font-black transition-all flex items-center gap-5 border-2',
              activeSettingsTab === 'bpjs'
                ? 'bg-white border-rose-600 text-rose-600 shadow-2xl shadow-rose-100 scale-[1.05]'
                : 'bg-transparent border-transparent text-slate-400 hover:text-slate-900 hover:bg-slate-50',
            ]"
          >
            <Heart class="w-5 h-5" /> BPJS & ASURANSI
          </button>
          <button
            @click="activeSettingsTab = 'master-components'"
            :class="[
              'text-left px-6 py-6 rounded-[1.75rem] text-[13px] font-black transition-all flex items-center gap-5 border-2',
              activeSettingsTab === 'master-components'
                ? 'bg-white border-rose-600 text-rose-600 shadow-2xl shadow-rose-100 scale-[1.05]'
                : 'bg-transparent border-transparent text-slate-400 hover:text-slate-900 hover:bg-slate-50',
            ]"
          >
            <Calculator class="w-5 h-5" /> KOMPONEN GAJI
          </button>
          <button
            @click="activeSettingsTab = 'deduction'"
            :class="[
              'text-left px-6 py-6 rounded-[1.75rem] text-[13px] font-black transition-all flex items-center gap-5 border-2',
              activeSettingsTab === 'deduction'
                ? 'bg-white border-rose-600 text-rose-600 shadow-2xl shadow-rose-100 scale-[1.05]'
                : 'bg-transparent border-transparent text-slate-400 hover:text-slate-900 hover:bg-slate-50',
            ]"
          >
            <ShieldAlert class="w-5 h-5" /> KEBIJAKAN DENDA
          </button>
          <button
            @click="activeSettingsTab = 'pajak'"
            :class="[
              'text-left px-6 py-6 rounded-[1.75rem] text-[13px] font-black transition-all flex items-center gap-5 border-2',
              activeSettingsTab === 'pajak'
                ? 'bg-white border-rose-600 text-rose-600 shadow-2xl shadow-rose-100 scale-[1.05]'
                : 'bg-transparent border-transparent text-slate-400 hover:text-slate-900 hover:bg-slate-50',
            ]"
          >
            <Percent class="w-5 h-5" /> AMBANG BATAS PAJAK
          </button>
        </div>

        <!-- Settings Dynamic Content -->
        <div
          class="lg:col-span-9 bg-white rounded-[3.5rem] border border-slate-100 shadow-[0_40px_100px_-20px_rgba(0,0,0,0.05)] overflow-hidden min-h-[600px]"
        >
          <!-- BPJS Configuration -->
          <div
            v-if="activeSettingsTab === 'bpjs'"
            class="animate-in slide-in-from-bottom-8 duration-700"
          >
            <div class="p-12 border-b border-slate-50 bg-slate-50/20">
              <h2 class="text-2xl font-black text-slate-900 tracking-tight">
                Perlindungan Sosial (BPJS)
              </h2>
              <p class="text-xs font-black text-rose-500 mt-1 uppercase tracking-[0.2em]">
                Parameter Iuran BPJS Kesehatan & Ketenagakerjaan
              </p>
            </div>

            <div class="p-12 space-y-16">
              <section class="space-y-10">
                <div class="flex items-center justify-between">
                  <div class="flex items-center gap-4">
                    <div
                      class="w-12 h-12 bg-rose-50 rounded-[1.25rem] flex items-center justify-center text-rose-600 border border-rose-100"
                    >
                      <Heart class="w-6 h-6" />
                    </div>
                    <div>
                      <h3 class="font-black text-slate-900 text-xl">BPJS Kesehatan</h3>
                      <p class="text-[11px] text-slate-400 font-bold uppercase tracking-widest">
                        Kontribusi Jaminan Kesehatan Nasional
                      </p>
                    </div>
                  </div>
                </div>

                <div class="grid grid-cols-1 md:grid-cols-3 gap-8">
                  <div class="space-y-3 p-6 bg-slate-50/50 rounded-3xl border border-slate-100/50">
                    <label
                      class="block text-[10px] font-black text-slate-400 uppercase tracking-widest ml-1"
                      >Company Contrib (%)</label
                    >
                    <input
                      type="number"
                      step="0.1"
                      v-model="config.bpjsKesCompanyPercentage"
                      class="w-full bg-white border-2 border-slate-100 rounded-2xl py-5 px-6 focus:outline-none focus:border-rose-500 font-black text-slate-800 transition-all text-xl shadow-sm"
                    />
                  </div>
                  <div class="space-y-3 p-6 bg-slate-50/50 rounded-3xl border border-slate-100/50">
                    <label
                      class="block text-[10px] font-black text-slate-400 uppercase tracking-widest ml-1"
                      >Employee Contrib (%)</label
                    >
                    <input
                      type="number"
                      step="0.1"
                      v-model="config.bpjsKesEmployeePercentage"
                      class="w-full bg-white border-2 border-slate-100 rounded-2xl py-5 px-6 focus:outline-none focus:border-rose-500 font-black text-slate-800 transition-all text-xl shadow-sm"
                    />
                  </div>
                  <div class="space-y-3 p-6 bg-slate-50/50 rounded-3xl border border-slate-100/50">
                    <label
                      class="block text-[10px] font-black text-slate-400 uppercase tracking-widest ml-1"
                      >Max Wage Base (Rp)</label
                    >
                    <input
                      type="number"
                      v-model="config.bpjsKesMaxWageBase"
                      class="w-full bg-white border-2 border-slate-100 rounded-2xl py-5 px-6 focus:outline-none focus:border-rose-500 font-black text-rose-600 transition-all text-xl shadow-sm"
                    />
                  </div>
                </div>
              </section>

              <hr class="border-slate-50" />

              <section class="space-y-10">
                <div class="flex items-center gap-4">
                  <div
                    class="w-12 h-12 bg-slate-900 rounded-[1.25rem] flex items-center justify-center text-white shadow-xl shadow-slate-200"
                  >
                    <Briefcase class="w-6 h-6" />
                  </div>
                  <div>
                    <h3 class="font-black text-slate-900 text-xl">Program Jamsostek (JHT & JP)</h3>
                    <p class="text-[11px] text-slate-400 font-bold uppercase tracking-widest">
                      Jaminan Hari Tua & Pensiun Berkala
                    </p>
                  </div>
                </div>

                <div class="grid grid-cols-1 md:grid-cols-2 gap-10">
                  <div
                    class="p-8 bg-rose-50/30 rounded-[2.5rem] border border-rose-100/50 space-y-8"
                  >
                    <h4 class="font-black text-rose-700 text-sm tracking-[0.1em] uppercase">
                      Jaminan Hari Tua (JHT)
                    </h4>
                    <div class="grid grid-cols-2 gap-4">
                      <div class="space-y-2">
                        <label class="text-[9px] font-black text-rose-400 uppercase tracking-widest"
                          >Company %</label
                        >
                        <input
                          type="number"
                          step="0.1"
                          v-model="config.jhtCompanyPercentage"
                          class="w-full bg-white border border-rose-100 rounded-xl p-4 font-black text-lg focus:ring-2 focus:ring-rose-500 outline-none"
                        />
                      </div>
                      <div class="space-y-2">
                        <label class="text-[9px] font-black text-rose-400 uppercase tracking-widest"
                          >Employee %</label
                        >
                        <input
                          type="number"
                          step="0.1"
                          v-model="config.jhtEmployeePercentage"
                          class="w-full bg-white border border-rose-100 rounded-xl p-4 font-black text-lg focus:ring-2 focus:ring-rose-500 outline-none"
                        />
                      </div>
                    </div>
                  </div>

                  <div
                    class="p-8 bg-slate-50/50 rounded-[2.5rem] border border-slate-100 space-y-8"
                  >
                    <h4 class="font-black text-slate-800 text-sm tracking-[0.1em] uppercase">
                      Jaminan Pensiun (JP)
                    </h4>
                    <div class="grid grid-cols-2 gap-4">
                      <div class="space-y-2">
                        <label
                          class="text-[9px] font-black text-slate-400 uppercase tracking-widest"
                          >Company %</label
                        >
                        <input
                          type="number"
                          step="0.1"
                          v-model="config.jpCompanyPercentage"
                          class="w-full bg-white border border-slate-100 rounded-xl p-4 font-black text-lg focus:ring-2 focus:ring-rose-500 outline-none"
                        />
                      </div>
                      <div class="space-y-2">
                        <label
                          class="text-[9px] font-black text-slate-400 uppercase tracking-widest"
                          >Employee %</label
                        >
                        <input
                          type="number"
                          step="0.1"
                          v-model="config.jpEmployeePercentage"
                          class="w-full bg-white border border-slate-100 rounded-xl p-4 font-black text-lg focus:ring-2 focus:ring-rose-500 outline-none"
                        />
                      </div>
                    </div>
                    <div class="space-y-2 pt-2">
                      <label class="text-[9px] font-black text-slate-400 uppercase tracking-widest"
                        >Max Wage Limit (Rp)</label
                      >
                      <input
                        type="number"
                        v-model="config.jpMaxWageBase"
                        class="w-full bg-white border border-slate-100 rounded-xl p-4 font-black text-rose-600 focus:ring-2 focus:ring-rose-500 outline-none"
                      />
                    </div>
                  </div>
                </div>
              </section>
            </div>
          </div>

          <!-- Master Components Table -->
          <div
            v-if="activeSettingsTab === 'master-components'"
            class="animate-in slide-in-from-bottom-8 duration-700"
          >
            <div class="p-12 border-b border-slate-50 bg-slate-50/20">
              <h2 class="text-2xl font-black text-slate-900 tracking-tight">
                Katalog Komponen Pendapatan
              </h2>
              <p class="text-xs font-black text-rose-500 mt-1 uppercase tracking-[0.2em]">
                Master data tunjangan, bonus, and potongan variabel
              </p>
            </div>

            <div class="p-6">
              <DataTable
                :data="components"
                :columns="componentColumns"
                :isLoading="isCompLoading"
              />
            </div>
          </div>

          <!-- Automated Deductions -->
          <div
            v-if="activeSettingsTab === 'deduction'"
            class="animate-in slide-in-from-bottom-8 duration-700"
          >
            <div class="p-12 border-b border-slate-50 bg-slate-50/20">
              <h2 class="text-2xl font-black text-slate-900 tracking-tight">
                Kebijakan Denda Kehadiran
              </h2>
              <p class="text-xs font-black text-rose-500 mt-1 uppercase tracking-[0.2em]">
                Penalti Otomatis Berdasarkan Log Absensi
              </p>
            </div>

            <div class="p-12 space-y-12">
              <section class="max-w-2xl">
                <div class="flex items-center gap-4 mb-10">
                  <div
                    class="w-12 h-12 bg-rose-600 rounded-[1.25rem] flex items-center justify-center text-white shadow-xl shadow-rose-200"
                  >
                    <ShieldAlert class="w-6 h-6" />
                  </div>
                  <div>
                    <h3 class="font-black text-slate-900 text-xl tracking-tight">
                      Potongan Mangkir (Absent)
                    </h3>
                    <p class="text-[11px] text-slate-400 font-bold uppercase tracking-widest">
                      Denda ketidakhadiran tanpa keterangan
                    </p>
                  </div>
                </div>

                <div
                  class="bg-slate-900 p-12 rounded-[3.5rem] text-white shadow-3xl shadow-slate-200 mb-10 overflow-hidden relative group"
                >
                  <div class="relative z-10">
                    <label
                      class="block text-[10px] font-black text-rose-500 uppercase tracking-[0.3em] mb-6"
                      >Nominal Per Kejadian (IDR)</label
                    >
                    <div class="flex items-center gap-6">
                      <span class="text-3xl font-black text-white/20 tracking-tighter">Rp</span>
                      <input
                        type="number"
                        v-model="config.absentDeduction"
                        class="bg-transparent border-none text-6xl font-black text-white focus:ring-0 w-full p-0 tracking-tighter"
                      />
                    </div>
                    <div class="h-px bg-white/10 my-10"></div>
                    <p class="text-[12px] font-bold text-slate-400 flex items-center gap-3">
                      <Timer class="w-4 h-4 text-rose-500" /> Kalkulasi otomatis saat Payroll
                      diproses
                    </p>
                  </div>
                  <ShieldAlert
                    class="absolute -right-10 -bottom-10 w-48 h-48 opacity-[0.03] group-hover:scale-110 transition-transform duration-700"
                  />
                </div>
              </section>

              <hr class="border-slate-50" />

              <section class="max-w-2xl">
                <div class="flex items-center gap-4 mb-10">
                  <div
                    class="w-12 h-12 bg-emerald-600 rounded-[1.25rem] flex items-center justify-center text-white shadow-xl shadow-emerald-200"
                  >
                    <TrendingUp class="w-6 h-6" />
                  </div>
                  <div>
                    <h3 class="font-black text-slate-900 text-xl tracking-tight">
                      Bonus Performa (KPI)
                    </h3>
                    <p class="text-[11px] text-slate-400 font-bold uppercase tracking-widest">
                      Insentif Produktivitas Berbasis Skor
                    </p>
                  </div>
                </div>

                <div
                  class="bg-emerald-600 p-12 rounded-[3.5rem] text-white shadow-3xl shadow-emerald-100 mb-10 overflow-hidden relative group"
                >
                  <div class="relative z-10">
                    <label
                      class="block text-[10px] font-black text-emerald-100 uppercase tracking-[0.3em] mb-6"
                      >Maksimal Bonus Tahunan/Bulanan</label
                    >
                    <div class="flex items-center gap-6">
                      <span class="text-3xl font-black text-emerald-900/20 tracking-tighter"
                        >Rp</span
                      >
                      <input
                        type="number"
                        v-model="config.kpiBonusAmount"
                        class="bg-transparent border-none text-6xl font-black text-white focus:ring-0 w-full p-0 tracking-tighter"
                      />
                    </div>
                    <div class="h-px bg-white/10 my-10"></div>
                    <p class="text-[12px] font-bold text-emerald-100 flex items-center gap-3">
                      <Calculator class="w-4 h-4" /> Berdasarkan finalisasi skor KPI (0-100)
                    </p>
                  </div>
                  <TrendingUp
                    class="absolute -right-10 -bottom-10 w-48 h-48 opacity-[0.05] group-hover:scale-110 transition-transform duration-700"
                  />
                </div>
              </section>
            </div>
          </div>

          <!-- Tax Configurations -->
          <div
            v-if="activeSettingsTab === 'pajak'"
            class="animate-in slide-in-from-bottom-8 duration-700"
          >
            <div class="p-12 border-b border-slate-50 bg-slate-50/20">
              <h2 class="text-2xl font-black text-slate-900 tracking-tight">Perpajakan (PPh 21)</h2>
              <p class="text-xs font-black text-rose-500 mt-1 uppercase tracking-[0.2em]">
                Ambang Batas Penghasilan Tidak Kena Pajak (PTKP)
              </p>
            </div>

            <div class="p-12 space-y-12">
              <section class="max-w-2xl">
                <div class="flex items-center gap-4 mb-10">
                  <div
                    class="w-12 h-12 bg-amber-500 rounded-[1.25rem] flex items-center justify-center text-white shadow-xl shadow-amber-200"
                  >
                    <Percent class="w-6 h-6" />
                  </div>
                  <div>
                    <h3 class="font-black text-slate-900 text-xl tracking-tight">
                      Standard PTKP TK/0
                    </h3>
                    <p class="text-[11px] text-slate-400 font-bold uppercase tracking-widest">
                      Batas penghasilan tahunan untuk status Single
                    </p>
                  </div>
                </div>

                <div class="p-10 bg-slate-50 rounded-[2.5rem] border-2 border-slate-100 space-y-6">
                  <div class="space-y-4">
                    <label
                      class="block text-[11px] font-black text-slate-400 uppercase tracking-widest ml-1"
                      >Nilai Dasar Tahunan (Rp)</label
                    >
                    <input
                      type="number"
                      v-model="config.ptkpBaseTK0"
                      class="w-full bg-white border-2 border-slate-100 rounded-3xl py-6 px-10 focus:outline-none focus:border-rose-500 font-black text-3xl text-slate-900 transition-all shadow-sm"
                    />
                    <p class="text-[11px] text-slate-400 font-bold ml-1 italic">
                      Default Regulation: Rp 54.000.000 / Tahun
                    </p>
                  </div>
                </div>
              </section>
            </div>
          </div>
        </div>
      </div>
    </div>
  </DialogContent>
</Dialog>

    <!-- MODALS & DIALOGS -->

    <!-- Run Detail Modal (Rincian Slip Gaji) -->
    <Dialog v-model:open="isDetailModalOpen">
      <DialogContent
        class="sm:max-w-5xl rounded-3xl p-0 overflow-hidden border border-slate-200 shadow-2xl max-h-[92vh] flex flex-col"
      >
        <!-- Modal Header -->
        <div class="bg-slate-900 p-6 md:p-8 text-white flex flex-col md:flex-row justify-between items-start md:items-center gap-4 shrink-0">
          <div>
            <div class="flex items-center gap-2">
              <h2 class="text-xl md:text-2xl font-black tracking-tight">
                Rincian <span class="text-rose-500">Slip Gaji</span>
              </h2>
              <span
                v-if="selectedRun?.status === 'DRAFT'"
                class="px-2.5 py-0.5 bg-amber-500/20 text-amber-300 border border-amber-500/30 text-[10px] font-black rounded-full uppercase"
              >
                Menunggu Review HRD
              </span>
              <span
                v-else-if="selectedRun?.status === 'APPROVED'"
                class="px-2.5 py-0.5 bg-blue-500/20 text-blue-300 border border-blue-500/30 text-[10px] font-black rounded-full uppercase"
              >
                Disetujui HRD
              </span>
              <span
                v-else
                class="px-2.5 py-0.5 bg-emerald-500/20 text-emerald-300 border border-emerald-500/30 text-[10px] font-black rounded-full uppercase"
              >
                Selesai (Processed)
              </span>
            </div>
            <p class="text-xs text-slate-400 font-medium mt-1 flex items-center gap-3">
              <span>Periode: <strong class="text-white">{{ selectedRun?.periodStart }}</strong> s/d <strong class="text-white">{{ selectedRun?.periodEnd }}</strong></span>
              <span>•</span>
              <span>Cabang: <strong class="text-white">{{ selectedRun?.branchName || 'Semua Cabang' }}</strong></span>
            </p>
          </div>

          <div class="bg-white/10 px-5 py-3 rounded-2xl border border-white/10 text-left md:text-right shrink-0">
            <p class="text-[10px] font-black text-slate-400 uppercase tracking-widest">
              Total Pengeluaran Batch
            </p>
            <p class="text-xl md:text-2xl font-black text-rose-400 tracking-tight tabular-nums mt-0.5">
              Rp {{ (selectedRun?.totalPayout || 0).toLocaleString('id-ID') }}
            </p>
          </div>
        </div>

        <!-- Modal Body (Payslip list) -->
        <div class="p-6 md:p-8 flex-1 overflow-y-auto custom-scrollbar bg-slate-50/50">
          <div
            v-if="isDetailLoading"
            class="py-20 flex flex-col items-center justify-center gap-3 text-slate-400"
          >
            <Loader2 class="w-8 h-8 text-rose-600 animate-spin" />
            <p class="font-bold text-xs">Memuat rincian slip gaji...</p>
          </div>

          <div v-else class="space-y-4">
            <div
              v-for="slip in payslips"
              :key="slip.id"
              class="bg-white rounded-2xl border border-slate-200/80 shadow-sm overflow-hidden transition-all"
            >
              <!-- Slip Header (Always visible summary) -->
              <div
                @click="toggleSlip(slip.id)"
                class="p-4 md:p-5 flex items-center justify-between cursor-pointer hover:bg-slate-50/70 transition-colors gap-4"
              >
                <div class="flex items-center gap-3.5 min-w-0">
                  <div
                    class="w-11 h-11 bg-rose-50 border border-rose-100 rounded-xl flex items-center justify-center font-black text-rose-600 shrink-0 text-base"
                  >
                    {{ slip.employee?.first_name?.charAt(0) || '?' }}
                  </div>
                  <div class="min-w-0">
                    <p class="font-black text-slate-900 text-sm truncate">
                      {{ slip.employee ? `${slip.employee.first_name} ${slip.employee.last_name || ''}`.trim() : 'Unknown Employee' }}
                    </p>
                    <p class="text-[11px] text-slate-400 font-semibold mt-0.5 truncate">
                      {{ slip.snapshot_job_title || 'Staff' }} • Rekening: {{ slip.employee?.bank_account_number || '-' }} ({{ slip.employee?.bank_name || '-' }})
                    </p>
                  </div>
                </div>

                <div class="flex items-center gap-4 shrink-0">
                  <div class="text-right">
                    <span class="text-[10px] font-black text-slate-400 uppercase tracking-wider block">
                      Take Home Pay
                    </span>
                    <span class="text-base md:text-lg font-black text-slate-900 tabular-nums">
                      Rp {{ (slip.take_home_pay || 0).toLocaleString('id-ID') }}
                    </span>
                  </div>
                  <div
                    :class="[
                      'w-8 h-8 rounded-lg flex items-center justify-center transition-transform duration-300 text-slate-400 bg-slate-100',
                      expandedSlipId === slip.id ? 'rotate-180 bg-rose-50 text-rose-600' : '',
                    ]"
                  >
                    <ChevronDown class="w-4 h-4" />
                  </div>
                </div>
              </div>

              <!-- Expanded Slip Detail Breakdown -->
              <div
                v-if="expandedSlipId === slip.id"
                class="p-5 md:p-6 bg-slate-50/70 border-t border-slate-100 animate-in slide-in-from-top-1 duration-300 space-y-6"
              >
                <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
                  <!-- Earnings Box -->
                  <div class="bg-white p-5 rounded-2xl border border-emerald-100 shadow-sm space-y-3">
                    <div class="flex items-center gap-2 pb-2 border-b border-emerald-50 text-emerald-700">
                      <TrendingUp class="w-4 h-4" />
                      <h4 class="text-xs font-black uppercase tracking-wider">
                        Pendapatan (Earnings)
                      </h4>
                    </div>
                    <div class="space-y-2 text-xs">
                      <div
                        v-for="item in slip.items?.filter((i: any) => i.type === 'EARNING')"
                        :key="item.id"
                        class="flex justify-between items-center py-1"
                      >
                        <span class="font-medium text-slate-600">{{ item.component_name }}</span>
                        <span class="font-bold text-slate-900 tabular-nums">
                          Rp {{ (item.amount || 0).toLocaleString('id-ID') }}
                        </span>
                      </div>
                      <div class="pt-3 mt-2 border-t border-slate-100 flex justify-between items-center font-black">
                        <span class="text-slate-500 uppercase text-[11px]">Total Pendapatan</span>
                        <span class="text-emerald-600 text-sm tabular-nums">
                          Rp {{ (slip.total_earning || 0).toLocaleString('id-ID') }}
                        </span>
                      </div>
                    </div>
                  </div>

                  <!-- Deductions Box -->
                  <div class="bg-white p-5 rounded-2xl border border-rose-100 shadow-sm space-y-3">
                    <div class="flex items-center gap-2 pb-2 border-b border-rose-50 text-rose-600">
                      <ShieldAlert class="w-4 h-4" />
                      <h4 class="text-xs font-black uppercase tracking-wider">
                        Potongan (Deductions)
                      </h4>
                    </div>
                    <div class="space-y-2 text-xs">
                      <div
                        v-for="item in slip.items?.filter((i: any) => i.type === 'DEDUCTION')"
                        :key="item.id"
                        class="flex justify-between items-center py-1"
                      >
                        <span class="font-medium text-slate-600">{{ item.component_name }}</span>
                        <span class="font-bold text-rose-500 tabular-nums">
                          - Rp {{ (item.amount || 0).toLocaleString('id-ID') }}
                        </span>
                      </div>
                      <div class="pt-3 mt-2 border-t border-slate-100 flex justify-between items-center font-black">
                        <span class="text-slate-500 uppercase text-[11px]">Total Potongan</span>
                        <span class="text-rose-600 text-sm tabular-nums">
                          - Rp {{ (slip.total_deduction || 0).toLocaleString('id-ID') }}
                        </span>
                      </div>
                    </div>
                  </div>
                </div>

                <!-- Action Bar inside expanded slip (HRD Adjust) -->
                <div class="flex flex-col sm:flex-row items-center justify-between gap-4 pt-2">
                  <div class="text-xs text-slate-500 flex items-center gap-2">
                    <ShieldCheck class="w-4 h-4 text-emerald-600" />
                    <span>Take Home Pay Final: <strong class="text-slate-900 font-bold">Rp {{ (slip.take_home_pay || 0).toLocaleString('id-ID') }}</strong></span>
                  </div>

                  <Button
                    v-if="selectedRun?.status === 'DRAFT'"
                    @click="openAdjustModal(slip)"
                    variant="outline"
                    size="sm"
                    class="rounded-xl border-rose-200 text-rose-600 hover:bg-rose-50 font-bold gap-2 text-xs"
                  >
                    <Pencil class="w-3.5 h-3.5" />
                    Koreksi / Sesuaikan Komponen
                  </Button>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- Modal Footer -->
        <div class="p-4 md:p-6 bg-white border-t border-slate-200 flex flex-wrap items-center justify-between gap-3 shrink-0">
          <div class="text-xs text-slate-500 font-medium">
            Total <strong class="text-slate-800">{{ payslips.length }} karyawan</strong> dalam batch ini.
          </div>

          <div class="flex items-center gap-3">
            <Button
              variant="outline"
              @click="isDetailModalOpen = false"
              class="rounded-xl h-10 px-5 font-bold text-slate-600 border-slate-200 hover:bg-slate-50 text-xs"
            >
              Tutup
            </Button>
            <Button
              v-if="selectedRun?.status === 'DRAFT'"
              @click="approvePayrollRun(selectedRun.id)"
              :disabled="isApproving"
              class="rounded-xl h-10 px-6 font-bold bg-emerald-600 hover:bg-emerald-700 text-white shadow-md shadow-emerald-600/20 text-xs gap-2"
            >
              <Loader2 v-if="isApproving" class="w-4 h-4 animate-spin" />
              <CheckCircle2 v-else class="w-4 h-4" />
              Setujui Batch Ini (Approve HRD)
            </Button>
          </div>
        </div>
      </DialogContent>
    </Dialog>

    <!-- Create Payroll Batch Dialog -->
    <Dialog v-model:open="isModalOpen">
      <DialogContent
        class="sm:max-w-xl rounded-3xl p-0 overflow-hidden border border-slate-200 shadow-2xl max-h-[92vh] flex flex-col"
      >
        <div class="bg-rose-600 p-6 md:p-8 text-white relative shrink-0">
          <DialogHeader>
            <DialogTitle class="text-xl md:text-2xl font-black tracking-tight text-white">
              Jalankan Batch Penggajian Baru
            </DialogTitle>
            <DialogDescription class="text-rose-100 font-medium mt-1 text-xs">
              Sistem akan mengalkulasi gaji pokok, tunjangan, potongan absensi, dan KPI secara otomatis.
            </DialogDescription>
          </DialogHeader>
        </div>

        <div class="p-6 md:p-8 space-y-5 bg-white overflow-y-auto custom-scrollbar flex-1">
          <div class="space-y-2">
            <label class="text-xs font-black text-slate-700 uppercase tracking-wider">
              Cabang Perusahaan
            </label>
            <Select v-model="newRun.branchId">
              <SelectTrigger class="h-11 rounded-xl bg-slate-50 border-slate-200 text-xs font-bold text-slate-800">
                <SelectValue placeholder="Pilih Cabang" />
              </SelectTrigger>
              <SelectContent class="rounded-xl border-slate-200 max-h-60">
                <SelectGroup>
                  <SelectItem value="ALL_BRANCHES" class="text-xs font-bold">
                    🏢 Semua Cabang
                  </SelectItem>
                  <SelectItem
                    v-for="branch in masterDataStore.branches"
                    :key="branch.id"
                    :value="branch.id"
                    class="text-xs font-bold"
                  >
                    🏢 {{ branch.name }}
                  </SelectItem>
                </SelectGroup>
              </SelectContent>
            </Select>
            <p class="text-[11px] font-medium text-slate-400">
              Pilih 1 cabang tertentu untuk mengisolasi batch penggajian, atau semua cabang sekaligus.
            </p>
          </div>

          <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <div class="space-y-2">
              <label class="text-xs font-black text-slate-700 uppercase tracking-wider">
                Mulai Periode
              </label>
              <div class="relative">
                <Calendar class="absolute left-3.5 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400" />
                <Input
                  v-model="newRun.periodStart"
                  type="date"
                  class="h-11 pl-10 rounded-xl bg-slate-50 border-slate-200 font-bold text-xs"
                />
              </div>
            </div>
            <div class="space-y-2">
              <label class="text-xs font-black text-slate-700 uppercase tracking-wider">
                Akhir Periode
              </label>
              <div class="relative">
                <Calendar class="absolute left-3.5 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400" />
                <Input
                  v-model="newRun.periodEnd"
                  type="date"
                  class="h-11 pl-10 rounded-xl bg-slate-50 border-slate-200 font-bold text-xs"
                />
              </div>
            </div>
          </div>

          <div class="space-y-2">
            <label class="text-xs font-black text-slate-700 uppercase tracking-wider">
              Target Tanggal Pencairan (Transfer)
            </label>
            <div class="relative">
              <Banknote class="absolute left-3.5 top-1/2 -translate-y-1/2 w-4 h-4 text-rose-600" />
              <Input
                v-model="newRun.paymentSchedule"
                type="date"
                class="h-11 pl-10 rounded-xl bg-rose-50/50 border-rose-200 font-black text-rose-700 text-xs"
              />
            </div>
          </div>

          <div class="p-4 bg-slate-50 rounded-2xl border border-slate-200/80 flex items-start gap-3">
            <ShieldCheck class="w-5 h-5 text-rose-600 shrink-0 mt-0.5" />
            <p class="text-xs font-medium text-slate-600 leading-relaxed">
              Batch baru akan berstatus <strong class="text-slate-900 font-black">DRAFT</strong> untuk ditinjau dan dikoreksi oleh HRD terlebih dahulu sebelum finalisasi dan pencairan bank.
            </p>
          </div>
        </div>

        <DialogFooter class="p-4 md:p-6 bg-white border-t border-slate-200 flex items-center justify-end gap-3 shrink-0">
          <Button
            variant="outline"
            @click="isModalOpen = false"
            :disabled="isSubmitting"
            class="rounded-xl h-10 px-5 font-bold text-slate-600 border-slate-200 hover:bg-slate-50 text-xs"
          >
            Batal
          </Button>
          <Button
            @click="generatePayroll"
            :disabled="isSubmitting"
            class="bg-rose-600 hover:bg-rose-700 text-white rounded-xl h-10 px-6 font-bold text-xs shadow-md shadow-rose-600/20 gap-2"
          >
            <Loader2 v-if="isSubmitting" class="w-4 h-4 animate-spin" />
            <Play v-else class="w-3.5 h-3.5 fill-current" />
            <span>Proses Batch Gaji</span>
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>

    <!-- HRD Payslip Adjustment Modal -->
    <Dialog v-model:open="isAdjustModalOpen">
      <DialogContent
        class="sm:max-w-3xl rounded-3xl p-0 overflow-hidden border border-slate-200 shadow-2xl max-h-[90vh] flex flex-col"
      >
        <div class="bg-slate-900 p-6 md:p-8 text-white relative shrink-0">
          <DialogHeader>
            <DialogTitle class="text-xl md:text-2xl font-black tracking-tight text-white">
              Koreksi & Penyesuaian <span class="text-rose-500">Gaji Karyawan</span>
            </DialogTitle>
            <DialogDescription
              class="text-slate-300 font-medium mt-1 text-xs flex flex-wrap items-center gap-2"
            >
              <span>Karyawan: <strong class="text-white">{{ adjustingSlip?.employee_name || adjustingSlip?.employee?.name || '-' }}</strong></span>
              <span>•</span>
              <span>Gaji Pokok: <strong class="text-white">Rp {{ (adjustingSlip?.basic_salary || 0).toLocaleString('id-ID') }}</strong></span>
            </DialogDescription>
          </DialogHeader>
        </div>

        <div class="p-6 md:p-8 bg-white overflow-y-auto custom-scrollbar flex-1 space-y-5">
          <div class="flex items-center justify-between pb-2 border-b border-slate-100">
            <div>
              <p class="font-bold text-slate-800 text-xs uppercase tracking-wider">Daftar Komponen Penggajian</p>
              <p class="text-[11px] text-slate-400 font-medium">Ubah nominal atau tambahkan penyesuaian/koreksi HRD sebelum batch disetujui.</p>
            </div>
            <Button
              @click="addAdjustItem"
              variant="outline"
              size="sm"
              class="rounded-xl border-rose-200 text-rose-600 hover:bg-rose-50 font-bold gap-1 text-xs h-8"
            >
              <Plus class="w-3.5 h-3.5" /> Tambah Baris
            </Button>
          </div>

          <div class="space-y-3">
            <div
              v-for="(item, idx) in adjustItems"
              :key="idx"
              class="flex flex-col sm:flex-row sm:items-center gap-3 p-3.5 bg-slate-50 rounded-2xl border border-slate-200/80"
            >
              <div class="flex-1 w-full sm:w-auto">
                <label class="text-[10px] font-black uppercase tracking-wider text-slate-400 block mb-1">Nama Komponen</label>
                <Input
                  v-model="item.component_name"
                  placeholder="Nama komponen"
                  class="h-10 rounded-xl bg-white border-slate-200 font-bold text-xs"
                />
              </div>

              <div class="w-full sm:w-36">
                <label class="text-[10px] font-black uppercase tracking-wider text-slate-400 block mb-1">Tipe</label>
                <Select v-model="item.type">
                  <SelectTrigger class="h-10 rounded-xl bg-white border-slate-200 font-bold text-xs">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent class="rounded-xl">
                    <SelectItem value="EARNING" class="text-xs font-bold text-emerald-600">+ Pendapatan</SelectItem>
                    <SelectItem value="DEDUCTION" class="text-xs font-bold text-rose-600">- Potongan</SelectItem>
                  </SelectContent>
                </Select>
              </div>

              <div class="w-full sm:w-44">
                <label class="text-[10px] font-black uppercase tracking-wider text-slate-400 block mb-1">Nominal (Rp)</label>
                <Input
                  v-model.number="item.amount"
                  type="number"
                  min="0"
                  class="h-10 rounded-xl bg-white border-slate-200 font-black text-xs tabular-nums"
                />
              </div>

              <div class="sm:pt-5 flex items-center justify-end">
                <button
                  @click="removeAdjustItem(idx)"
                  class="w-10 h-10 rounded-xl bg-rose-50 hover:bg-rose-100 text-rose-600 flex items-center justify-center transition-colors"
                  title="Hapus baris"
                >
                  <Trash2 class="w-4 h-4" />
                </button>
              </div>
            </div>
          </div>

          <!-- Total Calculation preview -->
          <div class="p-4 bg-slate-900 rounded-2xl text-white flex items-center justify-between">
            <div>
              <p class="text-[10px] font-black uppercase tracking-widest text-slate-400">Estimasi THP Bersih</p>
              <p class="text-[11px] font-medium text-slate-400">Total Pendapatan dikurangi Potongan</p>
            </div>
            <p class="text-xl font-black text-rose-400 tabular-nums">
              Rp {{ (
                adjustItems.reduce((acc, it) => it.type === 'EARNING' ? acc + (Number(it.amount) || 0) : acc - (Number(it.amount) || 0), 0)
              ).toLocaleString('id-ID') }}
            </p>
          </div>
        </div>

        <DialogFooter class="p-4 md:p-6 bg-white border-t border-slate-200 flex items-center justify-end gap-3 shrink-0">
          <Button
            variant="outline"
            @click="isAdjustModalOpen = false"
            :disabled="isAdjustSubmitting"
            class="h-10 px-5 rounded-xl font-bold text-slate-600 text-xs border-slate-200 hover:bg-slate-50"
          >
            Batal
          </Button>
          <Button
            @click="saveAdjustment"
            :disabled="isAdjustSubmitting"
            class="h-10 px-6 rounded-xl bg-rose-600 hover:bg-rose-700 text-white font-bold text-xs gap-2 shadow-md shadow-rose-600/20"
          >
            <Loader2 v-if="isAdjustSubmitting" class="w-4 h-4 animate-spin" />
            <Check v-else class="w-4 h-4" />
            Simpan Penyesuaian
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>

    <!-- Component Editor Modal -->
    <Dialog v-model:open="isCompModalOpen">
      <DialogContent
        class="sm:max-w-md rounded-3xl p-0 overflow-hidden border border-slate-200 shadow-2xl"
      >
        <div class="bg-slate-900 p-6 text-white relative">
          <DialogHeader>
            <DialogTitle class="text-xl font-black tracking-tight text-white">{{
              isCompEditMode ? 'Edit Komponen Gaji' : 'Tambah Komponen Baru'
            }}</DialogTitle>
            <DialogDescription class="text-slate-400 font-medium text-xs mt-1">
              Master komponen penghasilan atau potongan payroll
            </DialogDescription>
          </DialogHeader>
        </div>

        <div class="p-6 bg-white space-y-4">
          <div class="space-y-1.5">
            <label class="text-xs font-black text-slate-700 uppercase tracking-wider">
              Nama Komponen Gaji
            </label>
            <Input
              v-model="newComponent.name"
              placeholder="e.g. Tunjangan Transport, Bonus Khusus"
              class="h-10 rounded-xl bg-slate-50 border-slate-200 font-bold text-xs"
            />
          </div>
          <div class="space-y-1.5">
            <label class="text-xs font-black text-slate-700 uppercase tracking-wider">
              Kategori Tipe Gaji
            </label>
            <Select v-model="newComponent.type">
              <SelectTrigger class="h-10 rounded-xl bg-slate-50 border-slate-200 font-bold text-xs">
                <SelectValue placeholder="Pilih Tipe" />
              </SelectTrigger>
              <SelectContent class="rounded-xl border-slate-200">
                <SelectGroup>
                  <SelectItem value="EARNING" class="text-xs font-bold text-emerald-600">
                    Pendapatan (EARNING)
                  </SelectItem>
                  <SelectItem value="DEDUCTION" class="text-xs font-bold text-rose-600">
                    Potongan (DEDUCTION)
                  </SelectItem>
                </SelectGroup>
              </SelectContent>
            </Select>
          </div>
          <div class="flex items-center space-x-3 p-3 bg-slate-50 rounded-xl border border-slate-200">
            <input
              type="checkbox"
              id="taxable"
              v-model="newComponent.isTaxable"
              class="w-4 h-4 rounded border-slate-300 text-rose-600 focus:ring-rose-500"
            />
            <label for="taxable" class="text-xs font-bold text-slate-800 cursor-pointer">
              Termasuk objek pajak (Kena PPh 21)
            </label>
          </div>
        </div>

        <DialogFooter class="p-4 bg-white border-t border-slate-200 flex items-center justify-end gap-2">
          <Button
            variant="outline"
            @click="isCompModalOpen = false"
            class="h-10 px-5 rounded-xl font-bold text-slate-600 border-slate-200 hover:bg-slate-50 text-xs"
          >
            Batal
          </Button>
          <Button
            @click="saveComponent"
            :disabled="isCompSubmitting"
            class="h-10 px-6 rounded-xl bg-rose-600 hover:bg-rose-700 text-white font-bold text-xs shadow-md shadow-rose-600/20"
          >
            {{ isCompSubmitting ? 'Menyimpan...' : isCompEditMode ? 'Perbarui' : 'Tambahkan' }}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>

    <!-- Transfer File Modal (CMS Bank Mass Transfer) -->
    <Teleport to="body">
      <Transition
        enter-active-class="transition-all duration-300 ease-out"
        enter-from-class="opacity-0"
        enter-to-class="opacity-100"
        leave-active-class="transition-all duration-200 ease-in"
        leave-from-class="opacity-100"
        leave-to-class="opacity-0"
      >
        <div
          v-if="isTransferModalOpen"
          class="fixed inset-0 z-50 flex items-center justify-center p-4 sm:p-6"
          @click.self="isTransferModalOpen = false"
        >
          <!-- Backdrop -->
          <div class="absolute inset-0 bg-slate-900/60 backdrop-blur-sm"></div>

          <!-- Modal Card -->
          <Transition
            enter-active-class="transition-all duration-300 ease-out"
            enter-from-class="opacity-0 scale-95 translate-y-4"
            enter-to-class="opacity-100 scale-100 translate-y-0"
            leave-active-class="transition-all duration-200 ease-in"
            leave-from-class="opacity-100 scale-100 translate-y-0"
            leave-to-class="opacity-0 scale-95 translate-y-4"
            appear
          >
            <div
              v-if="isTransferModalOpen"
              class="relative w-full max-w-3xl max-h-[90vh] flex flex-col rounded-3xl overflow-hidden bg-white shadow-2xl border border-slate-200"
            >
              <!-- Modal Header -->
              <div class="bg-gradient-to-r from-blue-700 via-blue-800 to-indigo-900 p-6 md:p-8 text-white shrink-0 relative">
                <button
                  @click="isTransferModalOpen = false"
                  class="absolute top-5 right-5 w-8 h-8 rounded-lg bg-white/10 hover:bg-white/20 flex items-center justify-center text-white/80 hover:text-white transition-colors"
                >
                  <X class="w-4 h-4" />
                </button>
                <div class="flex items-center gap-3.5">
                  <div class="w-11 h-11 bg-white/10 rounded-xl flex items-center justify-center shrink-0">
                    <Building2 class="w-5 h-5 text-white" />
                  </div>
                  <div>
                    <h2 class="text-xl md:text-2xl font-black tracking-tight">Download File Transfer Gaji</h2>
                    <p class="text-blue-200 text-xs font-semibold mt-0.5">
                      Mass Transfer CMS Bank • Batch Payroll Run
                    </p>
                  </div>
                </div>

                <!-- Run Meta Info -->
                <div v-if="transferRun" class="mt-4 flex flex-wrap gap-2 text-xs font-bold text-white/90">
                  <span class="px-3 py-1 bg-white/10 rounded-lg flex items-center gap-1.5">
                    <Calendar class="w-3.5 h-3.5" />
                    {{ transferRun.periodStart }} s/d {{ transferRun.periodEnd }}
                  </span>
                  <span class="px-3 py-1 bg-white/10 rounded-lg flex items-center gap-1.5">
                    <Banknote class="w-3.5 h-3.5" />
                    Pencairan: {{ transferRun.paymentSchedule }}
                  </span>
                  <span class="px-3 py-1 bg-emerald-400/20 text-emerald-300 rounded-lg uppercase tracking-wider text-[10px] font-black">
                    {{ transferRun.status }}
                  </span>
                </div>
              </div>

              <!-- Body Area -->
              <div class="bg-slate-50 flex-1 overflow-y-auto custom-scrollbar p-6 space-y-4">
                <!-- Loading State -->
                <div v-if="isTransferLoading" class="py-20 flex flex-col items-center justify-center gap-3 text-slate-400">
                  <Loader2 class="w-8 h-8 text-blue-600 animate-spin" />
                  <p class="font-bold text-xs">Menyiapkan rincian data transfer bank...</p>
                </div>

                <!-- Loaded Content -->
                <div v-else-if="transferSummary" class="space-y-4">
                  <!-- Warning Missing Bank -->
                  <div
                    v-if="transferSummary.missing_bank > 0"
                    class="flex items-start gap-3 p-4 bg-amber-50 border border-amber-200 rounded-2xl"
                  >
                    <AlertTriangle class="w-4 h-4 text-amber-500 shrink-0 mt-0.5" />
                    <p class="text-amber-800 font-medium text-xs leading-relaxed">
                      <strong class="font-bold">{{ transferSummary.missing_bank }} karyawan</strong> belum melengkapi nomor/nama rekening bank. Karyawan ini tidak dimasukkan ke dalam file mass transfer.
                    </p>
                  </div>

                  <!-- Bank Card List -->
                  <div class="space-y-3">
                    <p class="text-xs font-black uppercase tracking-wider text-slate-400 px-1">
                      Rincian Per Bank Penerima
                    </p>

                    <div
                      v-for="bank in transferSummary.banks"
                      :key="bank.bank_name"
                      class="bg-white rounded-2xl border border-slate-200 p-4 sm:p-5 shadow-sm space-y-3"
                    >
                      <!-- Top row: Bank Info & Total -->
                      <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-2 border-b border-slate-100 pb-3">
                        <div class="flex items-center gap-3">
                          <div class="w-10 h-10 rounded-xl bg-blue-50 text-blue-600 flex items-center justify-center font-bold text-lg shrink-0">
                            🏦
                          </div>
                          <div>
                            <p class="font-black text-slate-900 text-sm">
                              {{ bankLabel(bank.bank_name) }}
                            </p>
                            <p class="text-[11px] text-slate-400 font-medium">
                              {{ bank.employee_count }} karyawan terdaftar
                            </p>
                          </div>
                        </div>

                        <div class="text-left sm:text-right">
                          <p class="text-base font-black text-slate-900 tabular-nums">
                            {{ formatIDR(bank.total_transfer) }}
                          </p>
                          <p class="text-[10px] text-slate-400 font-bold uppercase tracking-wider">
                            Total Transfer
                          </p>
                        </div>
                      </div>

                      <!-- Bottom row: Download Actions -->
                      <div class="flex flex-wrap items-center justify-end gap-2 pt-1">
                        <!-- BCA Actions -->
                        <template v-if="bank.bank_name.toUpperCase() === 'BCA'">
                          <button
                            @click="downloadTransferExcel(transferRun.id, 'BCA', 'BCA')"
                            :disabled="downloadingKey !== null"
                            class="inline-flex items-center gap-1.5 px-3.5 py-2 bg-emerald-600 hover:bg-emerald-700 disabled:opacity-50 text-white rounded-xl text-xs font-bold transition-all shadow-sm shadow-emerald-600/20"
                          >
                            <Loader2 v-if="downloadingKey === 'BCA-BCA-EXCEL'" class="w-3.5 h-3.5 animate-spin" />
                            <FileSpreadsheet v-else class="w-3.5 h-3.5" />
                            Excel KlikBCA CMS (.xlsx)
                          </button>
                          <button
                            @click="downloadTransferFile(transferRun.id, 'BCA', 'BCA')"
                            :disabled="downloadingKey !== null"
                            class="inline-flex items-center gap-1.5 px-3 py-2 bg-slate-100 hover:bg-slate-200 text-slate-700 disabled:opacity-50 rounded-xl text-xs font-bold transition-all"
                          >
                            <Loader2 v-if="downloadingKey === 'BCA-BCA'" class="w-3.5 h-3.5 animate-spin" />
                            <Download v-else class="w-3.5 h-3.5" />
                            CSV BCA
                          </button>
                        </template>

                        <!-- Mandiri Actions -->
                        <template v-else-if="bank.bank_name.toUpperCase() === 'MANDIRI'">
                          <button
                            @click="downloadTransferExcel(transferRun.id, 'MANDIRI', 'MANDIRI')"
                            :disabled="downloadingKey !== null"
                            class="inline-flex items-center gap-1.5 px-3.5 py-2 bg-emerald-600 hover:bg-emerald-700 disabled:opacity-50 text-white rounded-xl text-xs font-bold transition-all shadow-sm shadow-emerald-600/20"
                          >
                            <Loader2 v-if="downloadingKey === 'MANDIRI-MANDIRI-EXCEL'" class="w-3.5 h-3.5 animate-spin" />
                            <FileSpreadsheet v-else class="w-3.5 h-3.5" />
                            Excel Mandiri MCM (.xlsx)
                          </button>
                          <button
                            @click="downloadTransferFile(transferRun.id, 'MANDIRI', 'MANDIRI')"
                            :disabled="downloadingKey !== null"
                            class="inline-flex items-center gap-1.5 px-3 py-2 bg-slate-100 hover:bg-slate-200 text-slate-700 disabled:opacity-50 rounded-xl text-xs font-bold transition-all"
                          >
                            <Loader2 v-if="downloadingKey === 'MANDIRI-MANDIRI'" class="w-3.5 h-3.5 animate-spin" />
                            <Download v-else class="w-3.5 h-3.5" />
                            CSV Mandiri
                          </button>
                        </template>

                        <!-- Other Banks / Generic -->
                        <template v-else>
                          <button
                            @click="downloadTransferExcel(transferRun.id, bank.bank_name, 'GENERIC')"
                            :disabled="downloadingKey !== null"
                            class="inline-flex items-center gap-1.5 px-3.5 py-2 bg-emerald-600 hover:bg-emerald-700 disabled:opacity-50 text-white rounded-xl text-xs font-bold transition-all shadow-sm shadow-emerald-600/20"
                          >
                            <Loader2 v-if="downloadingKey === `${bank.bank_name}-GENERIC-EXCEL`" class="w-3.5 h-3.5 animate-spin" />
                            <FileSpreadsheet v-else class="w-3.5 h-3.5" />
                            Excel CMS (.xlsx)
                          </button>
                          <button
                            @click="downloadTransferFile(transferRun.id, bank.bank_name, 'GENERIC')"
                            :disabled="downloadingKey !== null"
                            class="inline-flex items-center gap-1.5 px-3 py-2 bg-slate-100 hover:bg-slate-200 text-slate-700 disabled:opacity-50 rounded-xl text-xs font-bold transition-all"
                          >
                            <Loader2 v-if="downloadingKey === `${bank.bank_name}-GENERIC`" class="w-3.5 h-3.5 animate-spin" />
                            <Download v-else class="w-3.5 h-3.5" />
                            CSV
                          </button>
                        </template>
                      </div>
                    </div>

                    <!-- Grand Total Banner -->
                    <div class="bg-slate-900 rounded-2xl p-4 sm:p-5 flex items-center justify-between text-white">
                      <div class="flex items-center gap-3">
                        <div class="w-9 h-9 bg-white/10 rounded-xl flex items-center justify-center">
                          <Banknote class="w-4 h-4 text-rose-400" />
                        </div>
                        <div>
                          <p class="font-black text-sm">TOTAL KESELURUHAN TRANSFER</p>
                          <p class="text-[11px] text-slate-400 font-medium">
                            {{ transferSummary.total_employees }} karyawan siap bayar
                          </p>
                        </div>
                      </div>
                      <p class="font-black text-base md:text-lg tabular-nums text-rose-400">
                        {{ formatIDR(transferSummary.total_transfer) }}
                      </p>
                    </div>
                  </div>
                </div>

                <!-- Empty State -->
                <div v-else class="py-20 flex flex-col items-center justify-center gap-3 text-slate-400">
                  <Building2 class="w-8 h-8" />
                  <p class="font-bold text-xs">Tidak ada data transfer untuk batch ini.</p>
                </div>
              </div>

              <!-- Modal Footer -->
              <div class="bg-white border-t border-slate-200 p-4 md:p-5 flex flex-wrap items-center justify-between gap-3 shrink-0">
                <Button
                  variant="outline"
                  @click="isTransferModalOpen = false"
                  class="rounded-xl h-10 px-5 font-bold text-slate-600 border-slate-200 hover:bg-slate-50 text-xs"
                >
                  Tutup
                </Button>

                <div class="flex flex-wrap items-center gap-2">
                  <button
                    v-if="transferSummary && !isTransferLoading"
                    @click="downloadTransferExcel(transferRun.id, 'ALL', 'GENERIC')"
                    :disabled="downloadingKey !== null"
                    class="inline-flex items-center gap-2 px-4 py-2 bg-emerald-600 hover:bg-emerald-700 disabled:opacity-50 text-white rounded-xl font-bold text-xs transition-all shadow-md shadow-emerald-600/20"
                  >
                    <Loader2 v-if="downloadingKey === 'ALL-GENERIC-EXCEL'" class="w-4 h-4 animate-spin" />
                    <FileSpreadsheet v-else class="w-4 h-4" />
                    Download Excel Mass Transfer (.xlsx)
                  </button>
                  <button
                    v-if="transferSummary && !isTransferLoading"
                    @click="downloadTransferFile(transferRun.id, 'ALL', 'GENERIC')"
                    :disabled="downloadingKey !== null"
                    class="inline-flex items-center gap-2 px-3.5 py-2 bg-slate-100 hover:bg-slate-200 disabled:opacity-50 text-slate-700 rounded-xl font-bold text-xs transition-all"
                  >
                    <Loader2 v-if="downloadingKey === 'ALL-GENERIC'" class="w-4 h-4 animate-spin" />
                    <Download v-else class="w-4 h-4" />
                    CSV Generic
                  </button>
                </div>
              </div>
            </div>
          </Transition>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>

<style scoped>
.no-scrollbar::-webkit-scrollbar {
  display: none;
}
.no-scrollbar {
  -ms-overflow-style: none;
  scrollbar-width: none;
}

.custom-scrollbar::-webkit-scrollbar {
  width: 6px;
}

.custom-scrollbar::-webkit-scrollbar-track {
  background: #f8fafc;
}

.custom-scrollbar::-webkit-scrollbar-thumb {
  background: #e2e8f0;
  border-radius: 10px;
}

.custom-scrollbar::-webkit-scrollbar-thumb:hover {
  background: #cbd5e1;
}
</style>
