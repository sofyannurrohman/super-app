<script setup lang="ts">
import { ref, onMounted, computed, h } from 'vue'
import {
  FileText,
  Plus,
  Search,
  Printer,
  Download,
  Eye,
  Calendar,
  User,
  Building2,
  FileCheck2,
  ArrowRight,
  ArrowLeft,
  X,
  CheckCircle2,
  Clock,
  Edit,
  AlertTriangle,
  AlertCircle,
  UploadCloud,
  History,
  ExternalLink,
  CalendarClock,
  Sparkles,
  FileUp,
  FolderArchive,
  FileCheck,
  Check,
} from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from '@/components/ui/dialog'
import DataTable from '@/components/DataTable.vue'
import PKWTRenderer from '@/components/PKWTRenderer.vue'
import apiClient from '@/api/axios'
import { useMasterDataStore } from '@/stores/masterData'
import { toast } from 'vue-sonner'

const masterData = useMasterDataStore()
const isLoading = ref(false)
const isSubmitting = ref(false)
const pkwtList = ref<any[]>([])
const isCreateModalOpen = ref(false)
const isPreviewModalOpen = ref(false)
const currentStep = ref(1)
const isEditMode = ref(false)
const selectedEpkwtId = ref<string | null>(null)

// History Modal
const isHistoryModalOpen = ref(false)
const historyEmployee = ref<any>(null)

// Quick Upload Modal
const isUploadModalOpen = ref(false)
const targetEpkwtForUpload = ref<any>(null)
const selectedDocumentFile = ref<File | null>(null)
const isUploadingDocument = ref(false)
const filePreviewUrl = ref<string | null>(null)

// Form State
const form = ref({
  documentNumber: '',
  agreementDay: '',
  agreementDate: new Date().toISOString().split('T')[0] ?? '',
  agreementPlace: 'Surakarta',

  // Pihak Pertama (Company)
  companyId: '',
  representativeName: 'Andi Muhammad Santoso',
  representativePosition: 'HRD',

  // Pihak Kedua (Employee)
  employeeId: '',

  // Contract details
  startDate: '',
  endDate: '',
  contractSequence: 1,
  jobPosition: '',
  jobPositionId: '',
  departmentId: '',
  probationMonths: 3,

  // Salary details
  basicSalary: 0,
  allowances: 0,
  notes: '',
  employeeSignatureUrl: '',
  hrSignatureUrl: '',
  documentUrl: '',
})

const hrSignatureFile = ref<File | null>(null)
const formDocumentFile = ref<File | null>(null)

const getFileFullUrl = (url: string | undefined) => {
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

const selectedEmployee = ref<any>(null)

const getDayName = (dateStr: string | undefined): string => {
  if (!dateStr) return '....................'
  const date = new Date(dateStr)
  const days = ['Minggu', 'Senin', 'Selasa', 'Rabu', 'Kamis', 'Jumat', 'Sabtu']
  return days[date.getDay()] || '....................'
}

const getInitials = (name: string): string => {
  if (!name) return ''
  return name
    .split(' ')
    .map((n) => n[0])
    .join('')
    .substring(0, 2)
    .toUpperCase()
}

const calculateDaysRemaining = (endDateStr: string | null | undefined): number | null => {
  if (!endDateStr) return null
  const end = new Date(endDateStr)
  end.setHours(23, 59, 59, 999)
  const now = new Date()
  const diffTime = end.getTime() - now.getTime()
  return Math.ceil(diffTime / (1000 * 60 * 60 * 24))
}

const setContractDurationMonths = (months: number) => {
  if (!form.value.startDate) {
    form.value.startDate = new Date().toISOString().split('T')[0] ?? ''
  }
  const start = new Date(form.value.startDate)
  const end = new Date(start)
  end.setMonth(end.getMonth() + months)
  // Minus 1 day for standard contract terms (e.g., Jan 1 - Dec 31)
  end.setDate(end.getDate() - 1)
  form.value.endDate = end.toISOString().split('T')[0] ?? ''
}

const fetchPKWT = async () => {
  isLoading.value = true
  try {
    const response = await apiClient.get('/epkwt')
    pkwtList.value = response.data.data.map((item: any) => {
      const empName = item.employee
        ? `${item.employee.first_name} ${item.employee.last_name || ''}`.trim()
        : `${item.first_name} ${item.last_name || ''}`.trim()
      const jobPos = item.job_position?.title || item.employee?.job_position?.title || ''
      const daysRemaining = calculateDaysRemaining(item.end_date)

      let expiryStatus = 'ACTIVE'
      if (!item.end_date) {
        expiryStatus = 'NO_END_DATE'
      } else if (daysRemaining !== null && daysRemaining < 0) {
        expiryStatus = 'EXPIRED'
      } else if (daysRemaining !== null && daysRemaining <= 7) {
        expiryStatus = 'CRITICAL'
      } else if (daysRemaining !== null && daysRemaining <= 30) {
        expiryStatus = 'EXPIRING_SOON'
      }

      return {
        id: item.id,
        documentNumber: item.document_number || 'BELUM DIISI',
        employeeId: item.employee_id,
        employeeName: empName,
        jobPosition: jobPos,
        contractSequence: item.contract_sequence || 1,
        startDate: item.start_date,
        endDate: item.end_date,
        documentUrl: item.document_url || '',
        daysRemaining: daysRemaining,
        expiryStatus: expiryStatus,
        agreementDate: item.agreement_date || item.created_at,
        status: item.status,
        originalData: item,
      }
    })
  } catch (error: any) {
    console.error(error)
    toast.error('Gagal memuat data EPKWT: ' + (error.response?.data?.error || error.message))
  } finally {
    isLoading.value = false
  }
}

onMounted(() => {
  fetchPKWT()
  masterData.fetchEmployees()
  masterData.fetchCompanies()
  masterData.fetchDepartments()
  masterData.fetchJobPositions()
})

const handleEmployeeChange = (id: any) => {
  const emp = masterData.employees.find((e) => e.id === id)
  if (emp) {
    selectedEmployee.value = emp
    form.value.jobPositionId = emp.job_position_id || ''
    form.value.departmentId = emp.department_id || ''
    form.value.jobPosition = emp.job_position?.title || ''
    form.value.basicSalary = emp.salary || 0

    // Auto calculate sequence based on past contracts count for this employee
    const pastCount = pkwtList.value.filter((p) => p.employeeId === id).length
    form.value.contractSequence = pastCount + 1
  }
}

const handleJobPositionChange = (id: any) => {
  const job = masterData.jobPositions.find((j) => j.id === id)
  if (job) {
    form.value.jobPosition = job.title
  }
}

const steps = [
  { id: 1, name: 'Informasi Dasar', icon: FileCheck2 },
  { id: 2, name: 'Pihak Pertama', icon: Building2 },
  { id: 3, name: 'Pihak Kedua', icon: User },
  { id: 4, name: 'Detail Kerja & Periode', icon: Clock },
  { id: 5, name: 'Gaji, Berkas & Review', icon: FileText },
]

const nextStep = () => {
  if (currentStep.value < 5) currentStep.value++
}

const prevStep = () => {
  if (currentStep.value > 1) currentStep.value--
}

const openCreateModal = () => {
  isEditMode.value = false
  selectedEpkwtId.value = null
  formDocumentFile.value = null
  hrSignatureFile.value = null
  form.value = {
    documentNumber: '',
    agreementDay: '',
    agreementDate: new Date().toISOString().split('T')[0] ?? '',
    agreementPlace: 'Surakarta',
    companyId: '',
    representativeName: 'Andi Muhammad Santoso',
    representativePosition: 'HRD',
    employeeId: '',
    startDate: new Date().toISOString().split('T')[0] ?? '',
    endDate: '',
    contractSequence: 1,
    jobPosition: '',
    jobPositionId: '',
    departmentId: '',
    probationMonths: 3,
    basicSalary: 0,
    allowances: 0,
    notes: '',
    employeeSignatureUrl: '',
    hrSignatureUrl: '',
    documentUrl: '',
  }
  setContractDurationMonths(12)
  selectedEmployee.value = null
  isCreateModalOpen.value = true
  currentStep.value = 1
}

const openEditModal = (row: any) => {
  const req = row.originalData
  isEditMode.value = true
  selectedEpkwtId.value = req.id
  formDocumentFile.value = null
  hrSignatureFile.value = null
  form.value = {
    documentNumber: req.document_number || '',
    agreementDay: getDayName(req.agreement_date || req.created_at),
    agreementDate: req.agreement_date ? req.agreement_date.split('T')[0] : new Date().toISOString().split('T')[0] ?? '',
    agreementPlace: 'Surakarta',
    companyId: req.employee?.company_id || req.employee?.branch?.company_id || '',
    representativeName: 'Andi Muhammad Santoso',
    representativePosition: 'HRD',
    employeeId: req.employee_id,
    startDate: req.start_date ? req.start_date.split('T')[0] : '',
    endDate: req.end_date ? req.end_date.split('T')[0] : '',
    contractSequence: req.contract_sequence || 1,
    jobPosition: req.job_position?.title || req.employee?.job_position?.title || '',
    jobPositionId: req.job_position_id || req.employee?.job_position_id || '',
    departmentId: req.department_id || req.employee?.department_id || '',
    probationMonths: req.probation_months ?? 3,
    basicSalary: req.salary || 0,
    allowances: 0,
    notes: req.notes || '',
    employeeSignatureUrl: req.employee_signature_url || '',
    hrSignatureUrl: req.hr_signature_url || '',
    documentUrl: req.document_url || '',
  }
  selectedEmployee.value = req.employee || {
    id: req.employee_id,
    first_name: req.first_name,
    last_name: req.last_name,
    identity_number: req.identity_number,
    address_ktp: req.address_ktp,
    birth_place: req.birth_place,
    birth_date: req.birth_date,
    gender: req.gender,
    phone_number: req.phone_number,
    address_residential: req.address_residential,
    marital_status: req.marital_status
  }
  isCreateModalOpen.value = true
  currentStep.value = 1
}

const openPreview = (row: any) => {
  const req = row.originalData
  form.value = {
    documentNumber: req.document_number || '',
    agreementDay: getDayName(req.agreement_date || req.created_at),
    agreementDate: req.agreement_date ? req.agreement_date.split('T')[0] : new Date().toISOString().split('T')[0] ?? '',
    agreementPlace: 'Surakarta',
    companyId: req.employee?.company_id || req.employee?.branch?.company_id || '',
    representativeName: 'Andi Muhammad Santoso',
    representativePosition: 'HRD',
    employeeId: req.employee_id,
    startDate: req.start_date ? req.start_date.split('T')[0] : '',
    endDate: req.end_date ? req.end_date.split('T')[0] : '',
    contractSequence: req.contract_sequence || 1,
    jobPosition: req.job_position?.title || req.employee?.job_position?.title || '',
    jobPositionId: req.job_position_id || req.employee?.job_position_id || '',
    departmentId: req.department_id || req.employee?.department_id || '',
    probationMonths: req.probation_months ?? 3,
    basicSalary: req.salary || 0,
    allowances: 0,
    notes: req.notes || '',
    employeeSignatureUrl: req.employee_signature_url || '',
    hrSignatureUrl: req.hr_signature_url || '',
    documentUrl: req.document_url || '',
  }
  selectedEmployee.value = req.employee || {
    id: req.employee_id,
    first_name: req.first_name,
    last_name: req.last_name,
    identity_number: req.identity_number,
    address_ktp: req.address_ktp,
    birth_place: req.birth_place,
    birth_date: req.birth_date,
    gender: req.gender,
    phone_number: req.phone_number,
    address_residential: req.address_residential,
    marital_status: req.marital_status
  }
  isPreviewModalOpen.value = true
}

const openHistoryModal = (row: any) => {
  historyEmployee.value = {
    id: row.employeeId,
    name: row.employeeName,
    jobPosition: row.jobPosition,
    identityNumber: row.originalData?.identity_number || row.originalData?.employee?.identity_number || '',
  }
  isHistoryModalOpen.value = true
}

const employeeHistoryList = computed(() => {
  if (!historyEmployee.value) return []
  return pkwtList.value
    .filter((item) => item.employeeId === historyEmployee.value.id)
    .sort((a, b) => (a.contractSequence || 1) - (b.contractSequence || 1))
})

const extendContractFromHistory = () => {
  isHistoryModalOpen.value = false
  openCreateModal()
  if (historyEmployee.value) {
    form.value.employeeId = historyEmployee.value.id
    handleEmployeeChange(historyEmployee.value.id)

    // Calculate next sequence and recommended start date
    const pastContracts = pkwtList.value.filter((i) => i.employeeId === historyEmployee.value.id)
    if (pastContracts.length > 0) {
      const sorted = [...pastContracts].sort((a, b) => (b.contractSequence || 1) - (a.contractSequence || 1))
      const latest = sorted[0]
      form.value.contractSequence = (latest.contractSequence || 1) + 1
      if (latest.originalData.end_date) {
        const nextStart = new Date(latest.originalData.end_date)
        nextStart.setDate(nextStart.getDate() + 1)
        form.value.startDate = nextStart.toISOString().split('T')[0] ?? ''
        setContractDurationMonths(12)
      }
    }
  }
}

const openQuickUploadModal = (row: any) => {
  targetEpkwtForUpload.value = row
  selectedDocumentFile.value = null
  filePreviewUrl.value = null
  isUploadModalOpen.value = true
}

const onFileSelected = (e: Event) => {
  const file = (e.target as HTMLInputElement).files?.[0] || null
  if (!file) {
    selectedDocumentFile.value = null
    filePreviewUrl.value = null
    return
  }

  // Validate size (15MB)
  if (file.size > 15 * 1024 * 1024) {
    toast.error('Ukuran file maksimal 15MB')
    return
  }

  selectedDocumentFile.value = file
  if (file.type.startsWith('image/')) {
    filePreviewUrl.value = URL.createObjectURL(file)
  } else {
    filePreviewUrl.value = null
  }
}

const handleQuickUpload = async () => {
  if (!targetEpkwtForUpload.value || !selectedDocumentFile.value) {
    toast.error('Pilih file dokumen terlebih dahulu')
    return
  }
  isUploadingDocument.value = true
  try {
    const formData = new FormData()
    formData.append('file', selectedDocumentFile.value)
    const res = await apiClient.post(`/epkwt/${targetEpkwtForUpload.value.id}/document`, formData)
    toast.success('Dokumen fisik PKWT berhasil diarsipkan!')
    isUploadModalOpen.value = false

    // Update locally in employeeHistoryList if history modal is currently open
    if (isHistoryModalOpen.value && res.data?.data?.url) {
      const foundInList = pkwtList.value.find((p) => p.id === targetEpkwtForUpload.value.id)
      if (foundInList) {
        foundInList.documentUrl = res.data.data.url
        foundInList.originalData.document_url = res.data.data.url
      }
    }
    fetchPKWT()
  } catch (err: any) {
    toast.error('Gagal mengunggah dokumen: ' + (err.response?.data?.error || err.message))
  } finally {
    isUploadingDocument.value = false
  }
}

const saveDocument = async (status: string) => {
  isSubmitting.value = true
  try {
    const payload: any = {
      document_number: form.value.documentNumber || null,
      job_position_id: form.value.jobPositionId || null,
      department_id: form.value.departmentId || null,
      salary: Number(form.value.basicSalary),
      probation_months: Number(form.value.probationMonths),
      contract_sequence: Number(form.value.contractSequence) || 1,
      start_date: form.value.startDate ? new Date(form.value.startDate).toISOString() : null,
      end_date: form.value.endDate ? new Date(form.value.endDate).toISOString() : null,
      agreement_date: form.value.agreementDate ? new Date(form.value.agreementDate).toISOString() : null,
      document_url: form.value.documentUrl || null,
      notes: form.value.notes || null,
      status: status,
    }

    let savedId = selectedEpkwtId.value

    if (isEditMode.value && selectedEpkwtId.value) {
      await apiClient.put(`/epkwt/${selectedEpkwtId.value}`, payload)
      toast.success(`EPKWT berhasil ${status === 'APPROVED_DRAFT' ? 'disetujui & diterbitkan' : 'diperbarui'}!`)
    } else {
      // 1. Create request
      const createPayload = {
        employee_id: form.value.employeeId,
        identity_number: selectedEmployee.value?.identity_number || '',
        first_name: selectedEmployee.value?.first_name || '',
        last_name: selectedEmployee.value?.last_name || null,
        phone_number: selectedEmployee.value?.phone_number || null,
        gender: selectedEmployee.value?.gender || null,
        birth_date: selectedEmployee.value?.birth_date || null,
        birth_place: selectedEmployee.value?.birth_place || null,
        address_ktp: selectedEmployee.value?.address_ktp || null,
        address_residential: selectedEmployee.value?.address_residential || null,
        marital_status: selectedEmployee.value?.marital_status || null,
        contract_sequence: Number(form.value.contractSequence) || 1,
        notes: form.value.notes || null,
      }
      const createRes = await apiClient.post('/epkwt', createPayload)
      const newReq = createRes.data.data

      if (newReq && newReq.id) {
        savedId = newReq.id
        await apiClient.put(`/epkwt/${newReq.id}`, payload)
      }
      toast.success('EPKWT berhasil dibuat!')
    }

    // 2. Upload document file if selected in form
    if (savedId && formDocumentFile.value) {
      const docFormData = new FormData()
      docFormData.append('file', formDocumentFile.value)
      await apiClient.post(`/epkwt/${savedId}/document`, docFormData)
    }

    // 3. Upload HR signature if selected
    if (savedId && hrSignatureFile.value) {
      const sigFormData = new FormData()
      sigFormData.append('file', hrSignatureFile.value)
      await apiClient.post(`/epkwt/${savedId}/hr-signature`, sigFormData)
    }

    isCreateModalOpen.value = false
    fetchPKWT()
  } catch (error: any) {
    console.error(error)
    toast.error('Gagal menyimpan EPKWT: ' + (error.response?.data?.error || error.message))
  } finally {
    isSubmitting.value = false
  }
}

// Columns definition
const columns = [
  {
    accessorKey: 'documentNumber',
    header: 'SURAT & URUTAN',
    cell: ({ row }: any) => {
      const seq = row.original.contractSequence || 1
      const isRenewal = seq > 1
      return h('div', { class: 'flex flex-col gap-1' }, [
        h('span', { class: 'font-bold text-slate-900 text-xs' }, row.original.documentNumber),
        h(
          'span',
          {
            class: `inline-flex items-center w-fit px-2 py-0.5 rounded-full text-[10px] font-black uppercase tracking-wider ${
              isRenewal
                ? 'bg-purple-100 text-purple-700 border border-purple-200'
                : 'bg-blue-100 text-blue-700 border border-blue-200'
            }`,
          },
          isRenewal ? `PKWT ${seq} (Perpanjangan)` : `PKWT ${seq} (Awal)`,
        ),
      ])
    },
  },
  {
    accessorKey: 'employeeName',
    header: 'KARYAWAN',
    cell: ({ row }: any) =>
      h('div', { class: 'flex flex-col' }, [
        h('span', { class: 'font-bold text-slate-900' }, row.original.employeeName),
        h('span', { class: 'text-xs text-slate-500 font-medium' }, row.original.jobPosition || '-'),
      ]),
  },
  {
    accessorKey: 'period',
    header: 'MASA BERLAKU & SISA HARI',
    cell: ({ row }: any) => {
      const startStr = row.original.startDate
        ? new Date(row.original.startDate).toLocaleDateString('id-ID', { day: 'numeric', month: 'short', year: 'numeric' })
        : '-'
      const endStr = row.original.endDate
        ? new Date(row.original.endDate).toLocaleDateString('id-ID', { day: 'numeric', month: 'short', year: 'numeric' })
        : '-'

      const days = row.original.daysRemaining
      let badgeColor = 'bg-slate-100 text-slate-600'
      let badgeLabel = 'Tanpa Batas'

      if (days !== null) {
        if (days < 0) {
          badgeColor = 'bg-red-100 text-red-700 font-black border border-red-200'
          badgeLabel = `Kedaluwarsa (${Math.abs(days)} hr lalu)`
        } else if (days <= 7) {
          badgeColor = 'bg-red-600 text-white font-black animate-pulse shadow-sm'
          badgeLabel = `Kritis: Sisa ${days} Hari`
        } else if (days <= 30) {
          badgeColor = 'bg-amber-100 text-amber-800 font-bold border border-amber-300'
          badgeLabel = `Segera Berakhir: ${days} Hari`
        } else {
          badgeColor = 'bg-emerald-100 text-emerald-700 font-bold'
          badgeLabel = `Sisa ${days} Hari`
        }
      }

      return h('div', { class: 'flex flex-col gap-1' }, [
        h('span', { class: 'text-xs font-semibold text-slate-700' }, `${startStr} s.d ${endStr}`),
        h(
          'span',
          {
            class: `inline-flex items-center w-fit px-2 py-0.5 rounded-full text-[10px] tracking-wide ${badgeColor}`,
          },
          badgeLabel,
        ),
      ])
    },
  },
  {
    accessorKey: 'document',
    header: 'BERKAS ARSIP',
    cell: ({ row }: any) => {
      const docUrl = row.original.documentUrl
      if (docUrl) {
        return h('div', { class: 'flex items-center gap-1.5' }, [
          h(
            'a',
            {
              href: getFileFullUrl(docUrl),
              target: '_blank',
              rel: 'noopener noreferrer',
              class: 'inline-flex items-center gap-1.5 px-3 py-1 rounded-xl bg-emerald-50 text-emerald-700 hover:bg-emerald-100 border border-emerald-200 text-xs font-bold transition-all shadow-sm',
            },
            [h(FileCheck, { class: 'w-3.5 h-3.5 text-emerald-600' }), 'Lihat Berkas'],
          ),
          h(
            Button,
            {
              variant: 'ghost',
              size: 'sm',
              class: 'h-7 w-7 p-0 rounded-lg text-slate-400 hover:text-slate-700',
              title: 'Perbarui / Ganti Berkas',
              onClick: () => openQuickUploadModal(row.original),
            },
            () => h(UploadCloud, { class: 'w-3.5 h-3.5' }),
          ),
        ])
      }
      return h(
        Button,
        {
          variant: 'outline',
          size: 'sm',
          class: 'h-7 px-2.5 rounded-xl text-[11px] font-bold text-amber-700 bg-amber-50/60 border border-amber-300 hover:bg-amber-100 hover:border-amber-400 gap-1.5 shadow-sm transition-all',
          onClick: () => openQuickUploadModal(row.original),
        },
        () => [h(UploadCloud, { class: 'w-3.5 h-3.5 text-amber-600' }), 'Upload Berkas'],
      )
    },
  },
  {
    accessorKey: 'status',
    header: 'STATUS',
    cell: ({ row }: any) => {
      let bgClass = 'bg-slate-100 text-slate-700'
      if (row.original.status === 'ACTIVE') {
        bgClass = 'bg-emerald-100 text-emerald-700'
      } else if (row.original.status === 'APPROVED_DRAFT') {
        bgClass = 'bg-blue-100 text-blue-700'
      } else if (row.original.status === 'PENDING_HR') {
        bgClass = 'bg-amber-100 text-amber-700'
      }
      return h(
        'span',
        {
          class: `px-2.5 py-1 rounded-full text-[10px] font-extrabold uppercase tracking-wider ${bgClass}`,
        },
        row.original.status,
      )
    },
  },
  {
    id: 'actions',
    header: 'AKSI',
    cell: ({ row }: any) => {
      const actions = [
        h(
          Button,
          {
            variant: 'ghost',
            size: 'sm',
            class: 'h-8 w-8 p-0 rounded-lg',
            title: 'Riwayat PKWT Karyawan',
            onClick: () => openHistoryModal(row.original),
          },
          () => h(History, { class: 'w-4 h-4 text-purple-600' }),
        ),
        h(
          Button,
          {
            variant: 'ghost',
            size: 'sm',
            class: 'h-8 w-8 p-0 rounded-lg',
            title: 'Pratinjau Surat',
            onClick: () => openPreview(row.original),
          },
          () => h(Eye, { class: 'w-4 h-4 text-slate-400 hover:text-slate-700' }),
        ),
      ]

      if (row.original.status !== 'ACTIVE') {
        actions.push(
          h(
            Button,
            {
              variant: 'ghost',
              size: 'sm',
              class: 'h-8 w-8 p-0 rounded-lg',
              title: 'Edit Kontrak',
              onClick: () => openEditModal(row.original),
            },
            () => h(Edit, { class: 'w-4 h-4 text-amber-500' }),
          ),
        )
      }

      actions.push(
        h(
          Button,
          {
            variant: 'ghost',
            size: 'sm',
            class: 'h-8 w-8 p-0 rounded-lg',
            title: 'Cetak Dokumen',
            onClick: () => {
              openPreview(row.original)
              setTimeout(() => printDocument(), 300)
            },
          },
          () => h(Printer, { class: 'w-4 h-4 text-primary' }),
        ),
      )

      return h('div', { class: 'flex items-center gap-1' }, actions)
    },
  },
]

const searchQuery = ref('')
const selectedYear = ref('all')
const activeFilterTab = ref<'all' | 'expiring_soon' | 'critical' | 'active' | 'expired' | 'missing_archive' | 'has_archive'>('all')

// Metrics
const stats = computed(() => {
  const activeContracts = pkwtList.value.filter((i) => i.status === 'ACTIVE')
  const totalActive = activeContracts.length

  const expiringSoonCount = activeContracts.filter(
    (i) => i.daysRemaining !== null && i.daysRemaining >= 0 && i.daysRemaining <= 30,
  ).length

  const criticalCount = activeContracts.filter(
    (i) => i.daysRemaining !== null && i.daysRemaining >= 0 && i.daysRemaining <= 7,
  ).length

  const expiredCount = pkwtList.value.filter(
    (i) => i.daysRemaining !== null && i.daysRemaining < 0,
  ).length

  const unarchivedCount = pkwtList.value.filter(
    (i) => i.status !== 'PENDING_HR' && !i.documentUrl,
  ).length

  const archivedCount = pkwtList.value.filter(
    (i) => i.status !== 'PENDING_HR' && !!i.documentUrl,
  ).length

  return {
    totalActive,
    expiringSoonCount,
    criticalCount,
    expiredCount,
    unarchivedCount,
    archivedCount,
  }
})

const filteredPkwtList = computed(() => {
  return pkwtList.value
    .filter((item) => item.status !== 'PENDING_HR')
    .filter((item) => {
      // Expiry filter tab
      if (activeFilterTab.value === 'expiring_soon') {
        return item.daysRemaining !== null && item.daysRemaining >= 0 && item.daysRemaining <= 30
      }
      if (activeFilterTab.value === 'critical') {
        return item.daysRemaining !== null && item.daysRemaining >= 0 && item.daysRemaining <= 7
      }
      if (activeFilterTab.value === 'expired') {
        return item.daysRemaining !== null && item.daysRemaining < 0
      }
      if (activeFilterTab.value === 'active') {
        return item.status === 'ACTIVE'
      }
      if (activeFilterTab.value === 'missing_archive') {
        return !item.documentUrl
      }
      if (activeFilterTab.value === 'has_archive') {
        return !!item.documentUrl
      }
      return true
    })
    .filter((item) => {
      const matchSearch =
        searchQuery.value.trim() === '' ||
        item.employeeName.toLowerCase().includes(searchQuery.value.toLowerCase()) ||
        item.documentNumber.toLowerCase().includes(searchQuery.value.toLowerCase())

      let matchYear = true
      if (selectedYear.value !== 'all') {
        const year = new Date(item.agreementDate).getFullYear().toString()
        matchYear = year === selectedYear.value
      }

      return matchSearch && matchYear
    })
})

const pendingRequestList = computed(() => {
  return pkwtList.value
    .filter((item) => item.status === 'PENDING_HR')
    .filter((item) => {
      const matchSearch =
        searchQuery.value.trim() === '' ||
        item.employeeName.toLowerCase().includes(searchQuery.value.toLowerCase()) ||
        item.documentNumber.toLowerCase().includes(searchQuery.value.toLowerCase())

      let matchYear = true
      if (selectedYear.value !== 'all') {
        const year = new Date(item.agreementDate).getFullYear().toString()
        matchYear = year === selectedYear.value
      }

      return matchSearch && matchYear
    })
})

const printDocument = () => {
  const printContent = document.getElementById('pkwt-print-area')
  if (!printContent) return

  const printWindow = window.open('', '_blank')
  if (!printWindow) return

  printWindow.document.write('<html><head><title>Perjanjian Kerja</title>')
  printWindow.document.write(`<base href="${window.location.origin}/">`)
  const styles = document.querySelectorAll('style')
  styles.forEach((s) => printWindow.document.write(s.outerHTML))
  const links = document.querySelectorAll('link[rel="stylesheet"]')
  links.forEach((l) => printWindow.document.write(l.outerHTML))

  printWindow.document.write('</head><body>')
  printWindow.document.write(printContent.innerHTML)
  printWindow.document.write('</body></html>')
  printWindow.document.close()

  setTimeout(() => {
    printWindow.print()
  }, 500)
}
</script>

<template>
  <div class="space-y-8 pb-12 animate-in fade-in duration-700">
    <!-- Premium Header -->
    <div class="flex flex-col md:flex-row md:items-center justify-between gap-6">
      <div>
        <h1 class="text-3xl font-extrabold text-slate-900 tracking-tight flex items-center gap-3">
          <FileText class="w-8 h-8 text-primary" />
          Manajemen Surat EPKWT
        </h1>
        <p class="text-[14px] font-medium text-slate-500 mt-1.5 uppercase tracking-widest flex items-center gap-2">
          Perjanjian Kerja Waktu Tertentu, Berkas & Notifikasi Masa Berlaku
        </p>
      </div>

      <div class="flex items-center gap-3">
        <Button
          @click="openCreateModal"
          class="gap-2 shadow-xl shadow-primary/20 rounded-2xl h-12 px-6 font-black bg-primary hover:bg-primary/90 transition-all hover:scale-105 active:scale-95"
        >
          <Plus class="w-5 h-5" /> BUAT SURAT PKWT
        </Button>
      </div>
    </div>

    <!-- Alert Banner If Contracts are Expiring -->
    <div
      v-if="stats.criticalCount > 0 || stats.expiringSoonCount > 0"
      class="p-5 rounded-3xl border transition-all flex flex-col md:flex-row items-start md:items-center justify-between gap-4 shadow-sm"
      :class="stats.criticalCount > 0 ? 'bg-red-50 border-red-200 text-red-900' : 'bg-amber-50 border-amber-200 text-amber-900'"
    >
      <div class="flex items-center gap-4">
        <div
          class="w-12 h-12 rounded-2xl flex items-center justify-center shrink-0"
          :class="stats.criticalCount > 0 ? 'bg-red-500 text-white animate-bounce' : 'bg-amber-500 text-white'"
        >
          <AlertTriangle class="w-6 h-6" />
        </div>
        <div>
          <h3 class="font-extrabold text-base">
            {{ stats.criticalCount > 0 ? 'Peringatan Kritis: Kontrak Segera Berakhir!' : 'Pemberitahuan Masa Berlaku Kontrak' }}
          </h3>
          <p class="text-xs mt-0.5 opacity-90 leading-relaxed">
            Terdapat <span class="font-black">{{ stats.expiringSoonCount }} kontrak</span> yang akan habis masa berlakunya dalam 30 hari ke depan
            <span v-if="stats.criticalCount > 0"> (termasuk <span class="font-black underline">{{ stats.criticalCount }} kontrak kritis &le; 7 hari</span>)</span>.
            Segera evaluasi untuk perpanjangan (PKWT berikutnya) atau persiapan serah terima.
          </p>
        </div>
      </div>
      <div class="flex items-center gap-2 self-end md:self-center shrink-0">
        <Button
          size="sm"
          variant="outline"
          @click="activeFilterTab = stats.criticalCount > 0 ? 'critical' : 'expiring_soon'"
          :class="stats.criticalCount > 0 ? 'bg-white text-red-700 border-red-300 font-bold' : 'bg-white text-amber-700 border-amber-300 font-bold'"
        >
          Lihat Kontrak &rarr;
        </Button>
      </div>
    </div>

    <!-- Metric Stats Cards -->
    <div class="grid grid-cols-2 lg:grid-cols-5 gap-4">
      <div
        @click="activeFilterTab = 'all'"
        class="bg-white p-5 rounded-3xl border border-slate-100 shadow-sm hover:shadow-md cursor-pointer transition-all"
        :class="activeFilterTab === 'all' ? 'ring-2 ring-primary border-primary/20' : ''"
      >
        <div class="flex items-center justify-between text-slate-500">
          <span class="text-xs font-black uppercase tracking-wider">Kontrak Aktif</span>
          <FileCheck2 class="w-5 h-5 text-primary" />
        </div>
        <p class="text-3xl font-black text-slate-900 mt-2">{{ stats.totalActive }}</p>
        <p class="text-[11px] font-bold text-slate-400 mt-1">Total kontrak berjalan</p>
      </div>

      <div
        @click="activeFilterTab = 'missing_archive'"
        class="bg-white p-5 rounded-3xl border border-slate-100 shadow-sm hover:shadow-md cursor-pointer transition-all"
        :class="activeFilterTab === 'missing_archive' ? 'ring-2 ring-amber-500 border-amber-500/20' : ''"
      >
        <div class="flex items-center justify-between text-amber-600">
          <span class="text-xs font-black uppercase tracking-wider">Belum Ada Berkas</span>
          <FolderArchive class="w-5 h-5 text-amber-500" />
        </div>
        <p class="text-3xl font-black text-amber-600 mt-2">{{ stats.unarchivedCount }}</p>
        <p class="text-[11px] font-bold text-amber-700/70 mt-1">Perlu diunggah arsipnya</p>
      </div>

      <div
        @click="activeFilterTab = 'expiring_soon'"
        class="bg-white p-5 rounded-3xl border border-slate-100 shadow-sm hover:shadow-md cursor-pointer transition-all"
        :class="activeFilterTab === 'expiring_soon' ? 'ring-2 ring-amber-500 border-amber-500/20' : ''"
      >
        <div class="flex items-center justify-between text-amber-600">
          <span class="text-xs font-black uppercase tracking-wider">Habis &le; 30 Hari</span>
          <AlertTriangle class="w-5 h-5 text-amber-500" />
        </div>
        <p class="text-3xl font-black text-amber-600 mt-2">{{ stats.expiringSoonCount }}</p>
        <p class="text-[11px] font-bold text-amber-700/70 mt-1">Butuh evaluasi</p>
      </div>

      <div
        @click="activeFilterTab = 'critical'"
        class="bg-white p-5 rounded-3xl border border-slate-100 shadow-sm hover:shadow-md cursor-pointer transition-all"
        :class="activeFilterTab === 'critical' ? 'ring-2 ring-red-500 border-red-500/20' : ''"
      >
        <div class="flex items-center justify-between text-red-600">
          <span class="text-xs font-black uppercase tracking-wider">Kritis &le; 7 Hari</span>
          <AlertCircle class="w-5 h-5 text-red-500 animate-pulse" />
        </div>
        <p class="text-3xl font-black text-red-600 mt-2">{{ stats.criticalCount }}</p>
        <p class="text-[11px] font-bold text-red-700/70 mt-1">Segera putuskan status</p>
      </div>

      <div
        @click="activeFilterTab = 'expired'"
        class="bg-white p-5 rounded-3xl border border-slate-100 shadow-sm hover:shadow-md cursor-pointer transition-all"
        :class="activeFilterTab === 'expired' ? 'ring-2 ring-slate-700 border-slate-700/20' : ''"
      >
        <div class="flex items-center justify-between text-slate-500">
          <span class="text-xs font-black uppercase tracking-wider">Kedaluwarsa</span>
          <CalendarClock class="w-5 h-5 text-slate-400" />
        </div>
        <p class="text-3xl font-black text-slate-700 mt-2">{{ stats.expiredCount }}</p>
        <p class="text-[11px] font-bold text-slate-400 mt-1">Lewat batas akhir</p>
      </div>
    </div>

    <!-- Quick Tab Filter Buttons -->
    <div class="flex items-center gap-2 overflow-x-auto pb-1">
      <button
        @click="activeFilterTab = 'all'"
        class="px-4 py-2 rounded-2xl text-xs font-bold transition-all whitespace-nowrap"
        :class="activeFilterTab === 'all' ? 'bg-slate-900 text-white shadow-md' : 'bg-white text-slate-600 hover:bg-slate-100'"
      >
        Semua Kontrak ({{ pkwtList.filter(i => i.status !== 'PENDING_HR').length }})
      </button>
      <button
        @click="activeFilterTab = 'missing_archive'"
        class="px-4 py-2 rounded-2xl text-xs font-bold transition-all whitespace-nowrap flex items-center gap-1.5"
        :class="activeFilterTab === 'missing_archive' ? 'bg-amber-600 text-white shadow-md' : 'bg-white text-amber-700 hover:bg-amber-50 border border-amber-200'"
      >
        <FolderArchive class="w-3.5 h-3.5" /> Belum Ada Berkas Arsip ({{ stats.unarchivedCount }})
      </button>
      <button
        @click="activeFilterTab = 'has_archive'"
        class="px-4 py-2 rounded-2xl text-xs font-bold transition-all whitespace-nowrap flex items-center gap-1.5"
        :class="activeFilterTab === 'has_archive' ? 'bg-emerald-600 text-white shadow-md' : 'bg-white text-emerald-700 hover:bg-emerald-50 border border-emerald-200'"
      >
        <FileCheck class="w-3.5 h-3.5" /> Sudah Diarsipkan ({{ stats.archivedCount }})
      </button>
      <button
        @click="activeFilterTab = 'expiring_soon'"
        class="px-4 py-2 rounded-2xl text-xs font-bold transition-all whitespace-nowrap flex items-center gap-1.5"
        :class="activeFilterTab === 'expiring_soon' ? 'bg-amber-500 text-white shadow-md' : 'bg-white text-amber-700 hover:bg-amber-50 border border-amber-200'"
      >
        <AlertTriangle class="w-3.5 h-3.5" /> Berakhir &le; 30 Hari ({{ stats.expiringSoonCount }})
      </button>
      <button
        @click="activeFilterTab = 'critical'"
        class="px-4 py-2 rounded-2xl text-xs font-bold transition-all whitespace-nowrap flex items-center gap-1.5"
        :class="activeFilterTab === 'critical' ? 'bg-red-600 text-white shadow-md' : 'bg-white text-red-700 hover:bg-red-50 border border-red-200'"
      >
        <AlertCircle class="w-3.5 h-3.5" /> Kritis &le; 7 Hari ({{ stats.criticalCount }})
      </button>
      <button
        @click="activeFilterTab = 'expired'"
        class="px-4 py-2 rounded-2xl text-xs font-bold transition-all whitespace-nowrap flex items-center gap-1.5"
        :class="activeFilterTab === 'expired' ? 'bg-slate-700 text-white shadow-md' : 'bg-white text-slate-600 hover:bg-slate-100'"
      >
        <CalendarClock class="w-3.5 h-3.5" /> Kedaluwarsa ({{ stats.expiredCount }})
      </button>
    </div>

    <!-- Search & Filters -->
    <div class="grid grid-cols-1 md:grid-cols-4 gap-4">
      <div class="md:col-span-2 relative">
        <Search class="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-slate-400" />
        <Input
          v-model="searchQuery"
          placeholder="Cari nomor surat atau nama karyawan..."
          class="pl-12 h-14 rounded-2xl border-none bg-white shadow-xl shadow-slate-100/50 font-medium"
        />
      </div>
      <Select v-model="selectedYear">
        <SelectTrigger
          class="h-14 rounded-2xl border-none bg-white shadow-xl shadow-slate-100/50 font-bold px-6"
        >
          <SelectValue placeholder="Semua Tahun" />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="all">Semua Tahun</SelectItem>
          <SelectItem value="2026">2026</SelectItem>
          <SelectItem value="2025">2025</SelectItem>
        </SelectContent>
      </Select>
      <Button
        variant="outline"
        class="h-14 rounded-2xl border-2 border-slate-50 bg-white shadow-xl shadow-slate-100/50 text-slate-600 font-bold gap-2"
      >
        <Download class="w-5 h-5" /> EXPORT DATA
      </Button>
    </div>

    <!-- Two Column Dashboard Layout -->
    <div class="grid grid-cols-1 lg:grid-cols-3 gap-8">
      <!-- Left Column: Processed EPKWT Table (takes 2 cols) -->
      <div class="lg:col-span-2 space-y-4">
        <div class="flex items-center justify-between px-2">
          <h2 class="text-lg font-bold text-slate-800 flex items-center gap-2">
            <FileText class="w-5 h-5 text-primary" />
            Daftar Kontrak PKWT
          </h2>
          <span class="text-xs text-slate-500 font-bold uppercase tracking-widest bg-slate-100 px-3 py-1 rounded-full">
            {{ filteredPkwtList.length }} Kontrak
          </span>
        </div>
        <div
          class="bg-white rounded-[2.5rem] border border-slate-200 shadow-2xl shadow-slate-100/50 overflow-hidden"
        >
          <DataTable :data="filteredPkwtList" :columns="columns" :isLoading="isLoading" />
        </div>
      </div>

      <!-- Right Column: Pending EPKWT Requests Action Queue (takes 1 col) -->
      <div class="space-y-4">
        <div class="flex items-center justify-between px-2">
          <h2 class="text-lg font-bold text-slate-800 flex items-center gap-2">
            <Clock class="w-5 h-5 text-amber-500" />
            Antrean Pengajuan Baru
          </h2>
          <span class="text-xs text-white font-bold uppercase tracking-widest bg-amber-500 px-3 py-1 rounded-full animate-pulse">
            {{ pendingRequestList.length }} Butuh Tindakan
          </span>
        </div>

        <div class="space-y-4 max-h-[60vh] overflow-y-auto pr-2 custom-scrollbar">
          <div v-if="pendingRequestList.length === 0" class="bg-white rounded-3xl p-8 border border-slate-100 shadow-xl text-center space-y-3">
            <CheckCircle2 class="w-12 h-12 text-emerald-500 mx-auto" />
            <div>
              <p class="font-bold text-slate-800 text-sm">Semua Beres!</p>
              <p class="text-xs text-slate-400 mt-1">Tidak ada pengajuan pembuatan PKWT yang menunggu tindakan.</p>
            </div>
          </div>

          <div v-for="item in pendingRequestList" :key="item.id" class="bg-white rounded-3xl p-6 border border-slate-100 shadow-xl hover:shadow-2xl transition-all border-l-4 border-l-amber-500 hover:-translate-y-1 duration-300">
            <div class="flex items-start justify-between gap-2">
              <div class="flex items-center gap-3">
                <div class="w-10 h-10 rounded-xl bg-amber-50 flex items-center justify-center text-amber-600 font-extrabold text-sm">
                  {{ getInitials(item.employeeName) }}
                </div>
                <div>
                  <h4 class="font-extrabold text-slate-800 text-sm leading-tight">{{ item.employeeName }}</h4>
                  <p class="text-[10px] text-slate-400 font-bold mt-1 uppercase tracking-wider">NIK: {{ item.originalData.identity_number }}</p>
                </div>
              </div>
            </div>

            <div class="mt-4 pt-4 border-t border-slate-50 flex items-center justify-between">
              <div class="flex flex-col">
                <span class="text-[9px] text-slate-400 font-bold uppercase tracking-widest">Tanggal Pengajuan</span>
                <span class="text-xs font-semibold text-slate-600 mt-0.5">{{ new Date(item.agreementDate).toLocaleDateString('id-ID', { day: 'numeric', month: 'short', year: 'numeric' }) }}</span>
              </div>

              <Button @click="openEditModal(item)" class="h-10 rounded-xl font-bold bg-amber-500 hover:bg-amber-600 text-white text-xs gap-1.5 shadow-lg shadow-amber-500/20 px-4">
                TINDAK LANJUT <ArrowRight class="w-3.5 h-3.5" />
              </Button>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Quick Upload Modal -->
    <Dialog v-model:open="isUploadModalOpen">
      <DialogContent class="sm:max-w-lg rounded-[2.5rem] p-8 shadow-2xl border-none">
        <DialogHeader>
          <div class="w-12 h-12 rounded-2xl bg-primary/10 flex items-center justify-center mb-2">
            <UploadCloud class="w-6 h-6 text-primary" />
          </div>
          <DialogTitle class="text-xl font-black text-slate-900">
            Arsipkan Dokumen Fisik PKWT
          </DialogTitle>
          <DialogDescription class="text-xs text-slate-500 font-medium">
            Unggah dokumen kontrak yang telah ditandatangani basah oleh karyawan & pihak perusahaan (PDF / Scan Gambar).
          </DialogDescription>
        </DialogHeader>

        <div v-if="targetEpkwtForUpload" class="space-y-5 my-2">
          <div class="p-4 bg-slate-50 rounded-2xl border border-slate-100 flex items-center justify-between">
            <div>
              <p class="text-[10px] text-slate-400 font-black uppercase tracking-wider">Karyawan & Urutan Kontrak</p>
              <p class="text-sm font-extrabold text-slate-900 mt-0.5">{{ targetEpkwtForUpload.employeeName }}</p>
              <p class="text-xs text-slate-500 font-semibold mt-0.5">
                {{ targetEpkwtForUpload.documentNumber }} &bull; PKWT {{ targetEpkwtForUpload.contractSequence }}
              </p>
            </div>
            <span
              v-if="targetEpkwtForUpload.documentUrl"
              class="px-2.5 py-1 rounded-full text-[10px] font-black uppercase tracking-wider bg-emerald-100 text-emerald-800 border border-emerald-200"
            >
              Sudah Ada Berkas
            </span>
            <span
              v-else
              class="px-2.5 py-1 rounded-full text-[10px] font-black uppercase tracking-wider bg-amber-100 text-amber-800 border border-amber-200"
            >
              Belum Diarsipkan
            </span>
          </div>

          <!-- Dropzone file input -->
          <div class="space-y-2">
            <label class="text-xs font-black text-slate-700 uppercase tracking-widest block ml-1">
              Pilih Berkas Dokumen (PDF, JPG, PNG &bull; Maks 15MB)
            </label>
            <div
              class="relative border-2 border-dashed rounded-2xl p-6 text-center transition-all group hover:border-primary/60 hover:bg-primary/5 cursor-pointer"
              :class="selectedDocumentFile ? 'border-primary bg-primary/5' : 'border-slate-200 bg-slate-50/50'"
            >
              <input
                type="file"
                accept=".pdf,image/jpeg,image/png,image/jpg"
                @change="onFileSelected"
                class="absolute inset-0 w-full h-full opacity-0 cursor-pointer z-10"
              />

              <div v-if="!selectedDocumentFile" class="flex flex-col items-center justify-center space-y-2">
                <div class="w-10 h-10 rounded-full bg-slate-100 flex items-center justify-center text-slate-500 group-hover:scale-110 transition-transform">
                  <FileUp class="w-5 h-5 text-primary" />
                </div>
                <p class="text-xs font-bold text-slate-700">Klik atau geser file ke sini</p>
                <p class="text-[11px] text-slate-400">Mendukung format PDF, JPG, atau PNG hingga 15MB</p>
              </div>

              <div v-else class="flex items-center gap-4 text-left">
                <div v-if="filePreviewUrl" class="w-14 h-14 rounded-xl overflow-hidden border bg-white shrink-0">
                  <img :src="filePreviewUrl" class="w-full h-full object-cover" alt="Preview" />
                </div>
                <div v-else class="w-12 h-12 rounded-xl bg-primary/10 flex items-center justify-center shrink-0">
                  <FileText class="w-6 h-6 text-primary" />
                </div>
                <div class="flex-1 min-w-0">
                  <p class="text-xs font-extrabold text-slate-900 truncate">{{ selectedDocumentFile.name }}</p>
                  <p class="text-[11px] text-slate-500 font-medium">
                    {{ (selectedDocumentFile.size / (1024 * 1024)).toFixed(2) }} MB
                  </p>
                </div>
                <button
                  type="button"
                  @click.stop="selectedDocumentFile = null; filePreviewUrl = null"
                  class="p-1.5 rounded-lg text-slate-400 hover:text-red-500 hover:bg-red-50 z-20 relative"
                  title="Hapus pilihan"
                >
                  <X class="w-4 h-4" />
                </button>
              </div>
            </div>
          </div>

          <div v-if="targetEpkwtForUpload.documentUrl" class="p-3 bg-slate-100 rounded-xl flex items-center justify-between text-xs">
            <span class="text-slate-600 font-medium">Berkas tersimpan sebelumnya:</span>
            <a
              :href="getFileFullUrl(targetEpkwtForUpload.documentUrl)"
              target="_blank"
              rel="noopener noreferrer"
              class="font-bold text-primary hover:underline inline-flex items-center gap-1"
            >
              <ExternalLink class="w-3.5 h-3.5" /> Lihat Dokumen Lama
            </a>
          </div>
        </div>

        <DialogFooter class="mt-6 flex items-center justify-end gap-3">
          <Button variant="outline" @click="isUploadModalOpen = false" class="rounded-xl px-5 font-bold">
            Batal
          </Button>
          <Button
            @click="handleQuickUpload"
            :disabled="isUploadingDocument || !selectedDocumentFile"
            class="rounded-xl font-black bg-primary hover:bg-primary/90 text-white px-6 gap-2 shadow-lg shadow-primary/20"
          >
            <Check class="w-4 h-4" v-if="!isUploadingDocument" />
            {{ isUploadingDocument ? 'Mengunggah & Mengarsipkan...' : 'Simpan ke Arsip' }}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>

    <!-- History Modal -->
    <Dialog v-model:open="isHistoryModalOpen">
      <DialogContent class="sm:max-w-3xl rounded-[2.5rem] p-8 max-h-[90vh] flex flex-col shadow-2xl border-none">
        <div class="flex items-start justify-between border-b pb-4 shrink-0">
          <div>
            <div class="flex items-center gap-2">
              <History class="w-6 h-6 text-purple-600" />
              <h2 class="text-xl font-black text-slate-900">Riwayat Arsip & Kontrak PKWT</h2>
            </div>
            <p v-if="historyEmployee" class="text-xs text-slate-500 font-semibold mt-1">
              Karyawan: <span class="text-slate-900 font-extrabold">{{ historyEmployee.name }}</span> (NIK: {{ historyEmployee.identityNumber || '-' }})
            </p>
          </div>
          <Button
            @click="extendContractFromHistory"
            size="sm"
            class="rounded-xl font-bold bg-purple-600 hover:bg-purple-700 text-white text-xs gap-1.5 shadow-md shadow-purple-600/20"
          >
            <Sparkles class="w-3.5 h-3.5" /> + Perpanjang Kontrak
          </Button>
        </div>

        <div class="flex-1 overflow-y-auto py-6 space-y-4">
          <div v-if="employeeHistoryList.length === 0" class="text-center py-10 text-slate-400 text-xs font-bold">
            Belum ada riwayat kontrak PKWT untuk karyawan ini.
          </div>

          <div
            v-for="item in employeeHistoryList"
            :key="item.id"
            class="p-5 rounded-2xl border transition-all space-y-3"
            :class="item.status === 'ACTIVE' ? 'bg-purple-50/50 border-purple-200' : 'bg-white border-slate-100 shadow-sm'"
          >
            <div class="flex items-center justify-between">
              <div class="flex items-center gap-2">
                <span
                  class="px-2.5 py-0.5 rounded-full text-xs font-black uppercase tracking-wider"
                  :class="item.contractSequence > 1 ? 'bg-purple-100 text-purple-700' : 'bg-blue-100 text-blue-700'"
                >
                  PKWT {{ item.contractSequence }} {{ item.contractSequence > 1 ? '(Perpanjangan)' : '(Kontrak Awal)' }}
                </span>
                <span class="text-xs font-bold text-slate-800">{{ item.documentNumber }}</span>
              </div>
              <div class="flex items-center gap-2">
                <span
                  v-if="item.documentUrl"
                  class="px-2.5 py-0.5 rounded-full text-[10px] font-extrabold uppercase bg-emerald-100 text-emerald-800 border border-emerald-200"
                >
                  Arsip Ada
                </span>
                <span
                  v-else
                  class="px-2.5 py-0.5 rounded-full text-[10px] font-extrabold uppercase bg-amber-100 text-amber-800 border border-amber-200"
                >
                  Belum Ada Arsip
                </span>
                <span
                  class="px-2 py-0.5 rounded-full text-[10px] font-extrabold uppercase"
                  :class="item.status === 'ACTIVE' ? 'bg-emerald-100 text-emerald-700' : 'bg-slate-100 text-slate-600'"
                >
                  {{ item.status }}
                </span>
              </div>
            </div>

            <div class="grid grid-cols-2 md:grid-cols-4 gap-3 text-xs bg-white/80 p-3 rounded-xl border border-slate-50">
              <div>
                <p class="text-[10px] text-slate-400 font-bold uppercase">Masa Berlaku</p>
                <p class="font-bold text-slate-800 mt-0.5">
                  {{ item.startDate ? new Date(item.startDate).toLocaleDateString('id-ID') : '-' }} s.d
                  {{ item.endDate ? new Date(item.endDate).toLocaleDateString('id-ID') : '-' }}
                </p>
              </div>
              <div>
                <p class="text-[10px] text-slate-400 font-bold uppercase">Sisa Masa</p>
                <p class="font-bold mt-0.5" :class="item.daysRemaining !== null && item.daysRemaining < 0 ? 'text-red-600' : 'text-slate-800'">
                  {{ item.daysRemaining !== null ? (item.daysRemaining < 0 ? `Kedaluwarsa` : `${item.daysRemaining} Hari`) : '-' }}
                </p>
              </div>
              <div>
                <p class="text-[10px] text-slate-400 font-bold uppercase">Jabatan</p>
                <p class="font-bold text-slate-800 mt-0.5">{{ item.jobPosition || '-' }}</p>
              </div>
              <div>
                <p class="text-[10px] text-slate-400 font-bold uppercase">Gaji Pokok</p>
                <p class="font-bold text-slate-800 mt-0.5">Rp {{ (item.originalData.salary || 0).toLocaleString('id-ID') }}</p>
              </div>
            </div>

            <div class="flex items-center justify-between pt-1">
              <span class="text-[11px] text-slate-400">
                Tgl Perjanjian: {{ new Date(item.agreementDate).toLocaleDateString('id-ID') }}
              </span>
              <div class="flex items-center gap-2">
                <a
                  v-if="item.documentUrl"
                  :href="getFileFullUrl(item.documentUrl)"
                  target="_blank"
                  rel="noopener noreferrer"
                  class="inline-flex items-center gap-1 text-xs font-bold text-emerald-700 bg-emerald-50 px-2.5 py-1 rounded-lg border border-emerald-200 hover:bg-emerald-100 transition-colors"
                >
                  <FileCheck class="w-3.5 h-3.5" /> Buka Arsip
                </a>
                <Button
                  size="sm"
                  variant="outline"
                  class="h-8 text-xs font-bold gap-1 text-slate-700 rounded-lg hover:border-primary hover:text-primary"
                  @click="openQuickUploadModal(item)"
                >
                  <UploadCloud class="w-3.5 h-3.5 text-primary" /> {{ item.documentUrl ? 'Ganti Arsip' : 'Upload Arsip' }}
                </Button>
                <Button
                  size="sm"
                  variant="ghost"
                  class="h-8 text-xs font-bold gap-1 text-primary rounded-lg"
                  @click="openPreview(item)"
                >
                  <Printer class="w-3.5 h-3.5" /> Cetak Surat
                </Button>
              </div>
            </div>
          </div>
        </div>

        <DialogFooter class="border-t pt-4 shrink-0">
          <Button variant="outline" @click="isHistoryModalOpen = false" class="rounded-xl">Tutup</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>

    <!-- Create PKWT Wizard Modal -->
    <Dialog v-model:open="isCreateModalOpen">
      <DialogContent
        class="sm:max-w-4xl rounded-[2.5rem] p-0 overflow-hidden border-none shadow-2xl max-h-[95vh] flex flex-col"
      >
        <!-- Progress Header -->
        <div class="bg-slate-900 p-8 text-white shrink-0">
          <div class="flex items-center justify-between mb-8">
            <div>
              <h2 class="text-2xl font-black">
                {{ isEditMode ? 'Edit Dokumen PKWT' : 'Wizard Pembuatan PKWT' }}
              </h2>
              <p class="text-slate-400 text-xs font-bold uppercase tracking-widest mt-1">
                Lengkapi data untuk menjana dokumen kontrak kerja
              </p>
            </div>
            <button
              @click="isCreateModalOpen = false"
              class="p-2 rounded-full bg-white/10 hover:bg-white/20 transition-colors"
            >
              <X class="w-5 h-5" />
            </button>
          </div>

          <div class="flex items-center justify-between relative px-2">
            <div
              v-for="step in steps"
              :key="step.id"
              class="flex flex-col items-center gap-2 relative z-10 group cursor-pointer"
              @click="step.id < currentStep ? (currentStep = step.id) : null"
            >
              <div
                :class="[
                  'w-10 h-10 rounded-xl flex items-center justify-center border-2 transition-all duration-500',
                  currentStep === step.id
                    ? 'bg-primary border-primary shadow-lg shadow-primary/30 scale-110'
                    : currentStep > step.id
                      ? 'bg-emerald-500 border-emerald-500'
                      : 'bg-slate-800 border-slate-700 text-slate-500',
                ]"
              >
                <component :is="currentStep > step.id ? CheckCircle2 : step.icon" class="w-5 h-5" />
              </div>
              <span
                :class="[
                  'text-[10px] font-black uppercase tracking-tighter transition-all duration-500 text-center max-w-[80px]',
                  currentStep === step.id ? 'text-primary' : 'text-slate-500',
                ]"
                >{{ step.name }}</span
              >
            </div>
            <!-- Connector Line -->
            <div class="absolute top-5 left-10 right-10 h-[2px] bg-slate-800 -z-0">
              <div
                class="h-full bg-primary transition-all duration-700"
                :style="{ width: `${(currentStep - 1) * 25}%` }"
              ></div>
            </div>
          </div>
        </div>

        <!-- Form Content -->
        <div class="flex-1 overflow-y-auto p-10 bg-slate-50/50">
          <!-- Step 1: Basic Info -->
          <div
            v-if="currentStep === 1"
            class="space-y-8 animate-in fade-in slide-in-from-right-4 duration-500"
          >
            <div class="bg-white p-8 rounded-3xl shadow-sm border border-slate-100 space-y-6">
              <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
                <div class="space-y-2">
                  <label class="text-xs font-black text-slate-500 uppercase tracking-widest ml-1"
                    >Nomor Surat</label
                  >
                  <Input
                    v-model="form.documentNumber"
                    placeholder="Kosongkan untuk generate otomatis..."
                    class="h-12 rounded-xl bg-slate-50 border-none font-bold"
                  />
                </div>
                <div class="space-y-2">
                  <label class="text-xs font-black text-slate-500 uppercase tracking-widest ml-1"
                    >Tempat Perjanjian</label
                  >
                  <Input
                    v-model="form.agreementPlace"
                    class="h-12 rounded-xl bg-slate-50 border-none font-bold"
                  />
                </div>
              </div>
              <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
                <div class="space-y-2">
                  <label class="text-xs font-black text-slate-500 uppercase tracking-widest ml-1"
                    >Hari</label
                  >
                  <Input
                    v-model="form.agreementDay"
                    placeholder="Contoh: Senin"
                    class="h-12 rounded-xl bg-slate-50 border-none font-bold"
                  />
                </div>
                <div class="space-y-2">
                  <label class="text-xs font-black text-slate-500 uppercase tracking-widest ml-1"
                    >Tanggal Perjanjian</label
                  >
                  <Input
                    v-model="form.agreementDate"
                    type="date"
                    class="h-12 rounded-xl bg-slate-50 border-none font-bold"
                  />
                </div>
              </div>
            </div>
          </div>

          <!-- Step 2: Pihak Pertama -->
          <div
            v-if="currentStep === 2"
            class="space-y-8 animate-in fade-in slide-in-from-right-4 duration-500"
          >
            <div class="bg-white p-8 rounded-3xl shadow-sm border border-slate-100 space-y-6">
              <div class="space-y-2">
                <label class="text-xs font-black text-slate-500 uppercase tracking-widest ml-1"
                  >Perusahaan (Pihak Pertama)</label
                >
                <Select v-model="form.companyId">
                  <SelectTrigger class="h-12 rounded-xl bg-slate-50 border-none font-bold">
                    <SelectValue placeholder="Pilih Perusahaan" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem v-for="c in masterData.companies" :key="c.id" :value="c.id">{{
                      c.name
                    }}</SelectItem>
                  </SelectContent>
                </Select>
              </div>
              <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
                <div class="space-y-2">
                  <label class="text-xs font-black text-slate-500 uppercase tracking-widest ml-1"
                    >Nama Perwakilan</label
                  >
                  <Input
                    v-model="form.representativeName"
                    class="h-12 rounded-xl bg-slate-50 border-none font-bold"
                  />
                </div>
                <div class="space-y-2">
                  <label class="text-xs font-black text-slate-500 uppercase tracking-widest ml-1"
                    >Jabatan Perwakilan</label
                  >
                  <Input
                    v-model="form.representativePosition"
                    class="h-12 rounded-xl bg-slate-50 border-none font-bold"
                  />
                </div>
              </div>
            </div>
          </div>

          <!-- Step 3: Pihak Kedua -->
          <div
            v-if="currentStep === 3"
            class="space-y-8 animate-in fade-in slide-in-from-right-4 duration-500"
          >
            <div class="bg-white p-8 rounded-3xl shadow-sm border border-slate-100 space-y-6">
              <div class="space-y-2">
                <label class="text-xs font-black text-slate-500 uppercase tracking-widest ml-1"
                  >Pilih Karyawan (Pihak Kedua)</label
                >
                <Select v-model="form.employeeId" @update:modelValue="handleEmployeeChange" :disabled="isEditMode">
                  <SelectTrigger class="h-12 rounded-xl bg-slate-50 border-none font-bold">
                    <SelectValue placeholder="Cari Karyawan..." />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem v-for="e in masterData.employees" :key="e.id" :value="e.id"
                      >{{ e.first_name }} {{ e.last_name }}</SelectItem
                    >
                  </SelectContent>
                </Select>
              </div>

              <div
                v-if="selectedEmployee"
                class="p-6 bg-slate-50 rounded-2xl border border-slate-100 grid grid-cols-2 gap-4"
              >
                <div class="flex flex-col">
                  <span class="text-[10px] font-black text-slate-400 uppercase tracking-widest"
                    >NIK</span
                  >
                  <span class="font-bold text-slate-700">{{
                    selectedEmployee.identity_number || '-'
                  }}</span>
                </div>
                <div class="flex flex-col">
                  <span class="text-[10px] font-black text-slate-400 uppercase tracking-widest"
                    >Alamat (KTP)</span
                  >
                  <span class="font-bold text-slate-700 text-xs">{{
                    selectedEmployee.address_ktp || '-'
                  }}</span>
                </div>
                <div class="flex flex-col">
                  <span class="text-[10px] font-black text-slate-400 uppercase tracking-widest"
                    >Tempat, Tgl Lahir</span
                  >
                  <span class="font-bold text-slate-700"
                    >{{ selectedEmployee.birth_place }},
                    {{
                      selectedEmployee.birth_date
                        ? new Date(selectedEmployee.birth_date).toLocaleDateString('id-ID')
                        : '-'
                    }}</span
                  >
                </div>
                <div class="flex flex-col">
                  <span class="text-[10px] font-black text-slate-400 uppercase tracking-widest"
                    >Jenis Kelamin</span
                  >
                  <span class="font-bold text-slate-700">{{ selectedEmployee.gender }}</span>
                </div>
              </div>
            </div>
          </div>

          <!-- Step 4: Detail Kerja & Periode -->
          <div
            v-if="currentStep === 4"
            class="space-y-8 animate-in fade-in slide-in-from-right-4 duration-500"
          >
            <div class="bg-white p-8 rounded-3xl shadow-sm border border-slate-100 space-y-6">
              <!-- Urutan Kontrak -->
              <div class="p-4 bg-purple-50/60 rounded-2xl border border-purple-100 flex flex-col md:flex-row md:items-center justify-between gap-4">
                <div>
                  <label class="text-xs font-black text-purple-900 uppercase tracking-widest">
                    Status Urutan PKWT ke-Berapa
                  </label>
                  <p class="text-xs text-purple-700 mt-0.5">
                    Otomatis dihitung dari jumlah kontrak yang sudah dimiliki karyawan ini.
                  </p>
                </div>
                <div class="flex items-center gap-3">
                  <Select v-model="form.contractSequence">
                    <SelectTrigger class="w-48 h-11 rounded-xl bg-white border border-purple-200 font-black text-purple-900">
                      <SelectValue placeholder="Pilih Urutan" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem :value="1">PKWT 1 (Awal)</SelectItem>
                      <SelectItem :value="2">PKWT 2 (Perpanjangan 1)</SelectItem>
                      <SelectItem :value="3">PKWT 3 (Perpanjangan 2)</SelectItem>
                      <SelectItem :value="4">PKWT 4 (Perpanjangan 3)</SelectItem>
                      <SelectItem :value="5">PKWT 5</SelectItem>
                    </SelectContent>
                  </Select>
                </div>
              </div>

              <!-- Masa Kontrak -->
              <div class="space-y-3">
                <div class="flex items-center justify-between">
                  <label class="text-xs font-black text-slate-500 uppercase tracking-widest ml-1">
                    Masa Berlaku Kontrak
                  </label>
                  <div class="flex items-center gap-1.5">
                    <span class="text-[10px] font-bold text-slate-400 uppercase">Set Cepat:</span>
                    <button
                      type="button"
                      @click="setContractDurationMonths(3)"
                      class="px-2 py-1 rounded-lg bg-slate-100 hover:bg-slate-200 text-[11px] font-bold text-slate-700"
                    >
                      3 Bulan
                    </button>
                    <button
                      type="button"
                      @click="setContractDurationMonths(6)"
                      class="px-2 py-1 rounded-lg bg-slate-100 hover:bg-slate-200 text-[11px] font-bold text-slate-700"
                    >
                      6 Bulan
                    </button>
                    <button
                      type="button"
                      @click="setContractDurationMonths(12)"
                      class="px-2 py-1 rounded-lg bg-primary/10 hover:bg-primary/20 text-[11px] font-bold text-primary"
                    >
                      1 Tahun
                    </button>
                    <button
                      type="button"
                      @click="setContractDurationMonths(24)"
                      class="px-2 py-1 rounded-lg bg-slate-100 hover:bg-slate-200 text-[11px] font-bold text-slate-700"
                    >
                      2 Tahun
                    </button>
                  </div>
                </div>

                <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
                  <div class="space-y-2">
                    <label class="text-xs font-bold text-slate-600 ml-1">Tanggal Mulai</label>
                    <Input
                      v-model="form.startDate"
                      type="date"
                      class="h-12 rounded-xl bg-slate-50 border-none font-bold"
                    />
                  </div>
                  <div class="space-y-2">
                    <label class="text-xs font-bold text-slate-600 ml-1">Tanggal Berakhir</label>
                    <Input
                      v-model="form.endDate"
                      type="date"
                      class="h-12 rounded-xl bg-slate-50 border-none font-bold"
                    />
                  </div>
                </div>
              </div>

              <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
                <div class="space-y-2">
                  <label class="text-xs font-black text-slate-500 uppercase tracking-widest ml-1"
                    >Departemen</label
                  >
                  <Select v-model="form.departmentId">
                    <SelectTrigger class="h-12 rounded-xl bg-slate-50 border-none font-bold">
                      <SelectValue placeholder="Pilih Departemen" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem v-for="d in masterData.departments" :key="d.id" :value="d.id">
                        {{ d.name }}
                      </SelectItem>
                    </SelectContent>
                  </Select>
                </div>
                <div class="space-y-2">
                  <label class="text-xs font-black text-slate-500 uppercase tracking-widest ml-1"
                    >Lama Probation (Bulan)</label
                  >
                  <Input
                    v-model="form.probationMonths"
                    type="number"
                    min="0"
                    class="h-12 rounded-xl bg-slate-50 border-none font-bold"
                  />
                </div>
              </div>

              <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
                <div class="space-y-2">
                  <label class="text-xs font-black text-slate-500 uppercase tracking-widest ml-1"
                    >Jabatan di Kontrak</label
                  >
                  <Select v-model="form.jobPositionId" @update:modelValue="handleJobPositionChange">
                    <SelectTrigger class="h-12 rounded-xl bg-slate-50 border-none font-bold">
                      <SelectValue placeholder="Pilih Jabatan" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem v-for="j in masterData.jobPositions" :key="j.id" :value="j.id">
                        {{ j.title }}
                      </SelectItem>
                    </SelectContent>
                  </Select>
                </div>
                <div class="space-y-2">
                  <label class="text-xs font-black text-slate-500 uppercase tracking-widest ml-1"
                    >Tanda Tangan HRD (Opsional)</label
                  >
                  <div class="flex items-center gap-4">
                    <Input
                      type="file"
                      accept="image/*"
                      @change="(e: Event) => hrSignatureFile = (e.target as HTMLInputElement).files?.[0] || null"
                      class="h-12 rounded-xl bg-slate-50 border-none font-bold pt-3"
                    />
                    <img v-if="form.hrSignatureUrl" :src="getFileFullUrl(form.hrSignatureUrl)" class="h-12 w-auto object-contain border rounded bg-white" alt="HR Sig" />
                  </div>
                </div>
              </div>
            </div>
          </div>

          <!-- Step 5: Gaji, Berkas & Review -->
          <div
            v-if="currentStep === 5"
            class="space-y-8 animate-in fade-in slide-in-from-right-4 duration-500"
          >
            <div class="bg-white p-8 rounded-3xl shadow-sm border border-slate-100 space-y-6">
              <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
                <div class="space-y-2">
                  <label class="text-xs font-black text-slate-500 uppercase tracking-widest ml-1"
                    >Gaji Pokok (Rp)</label
                  >
                  <Input
                    v-model="form.basicSalary"
                    type="number"
                    class="h-12 rounded-xl bg-slate-50 border-none font-bold"
                  />
                </div>
                <div class="space-y-2">
                  <label class="text-xs font-black text-slate-500 uppercase tracking-widest ml-1"
                    >Tunjangan Lainnya (Rp)</label
                  >
                  <Input
                    v-model="form.allowances"
                    type="number"
                    class="h-12 rounded-xl bg-slate-50 border-none font-bold"
                  />
                </div>
              </div>

              <div class="p-6 bg-primary/5 rounded-2xl border border-primary/10">
                <div class="flex justify-between items-center">
                  <span class="text-sm font-black text-slate-600"
                    >Total Take Home Pay (Estimasi)</span
                  >
                  <span class="text-xl font-black text-primary"
                    >Rp
                    {{
                      (Number(form.basicSalary) + Number(form.allowances)).toLocaleString('id-ID')
                    }}</span
                  >
                </div>
              </div>

              <!-- Upload Berkas Scan Fisik -->
              <div class="p-6 bg-emerald-50/50 rounded-2xl border border-emerald-100 space-y-3">
                <div class="flex items-center gap-2">
                  <FileUp class="w-5 h-5 text-emerald-600" />
                  <div>
                    <h4 class="text-xs font-black text-emerald-900 uppercase tracking-wider">
                      Upload Dokumen Fisik / Scan EPKWT (Opsional)
                    </h4>
                    <p class="text-[11px] text-emerald-700">
                      Lampirkan berkas kontrak bertanda tangan basah atau scan PDF/Foto resmi.
                    </p>
                  </div>
                </div>

                <div class="flex items-center gap-4">
                  <Input
                    type="file"
                    accept=".pdf,image/*"
                    @change="(e: Event) => formDocumentFile = (e.target as HTMLInputElement).files?.[0] || null"
                    class="h-12 rounded-xl bg-white border border-emerald-200 font-bold pt-3 text-xs"
                  />
                  <a
                    v-if="form.documentUrl"
                    :href="getFileFullUrl(form.documentUrl)"
                    target="_blank"
                    rel="noopener noreferrer"
                    class="px-3 py-2 rounded-xl bg-emerald-100 text-emerald-800 text-xs font-bold whitespace-nowrap hover:bg-emerald-200"
                  >
                    Lihat Berkas Lama
                  </a>
                </div>
              </div>

              <div class="space-y-2">
                <label class="text-xs font-black text-slate-500 uppercase tracking-widest ml-1"
                  >Catatan Khusus Kontrak (Opsional)</label
                >
                <Input
                  v-model="form.notes"
                  placeholder="Keterangan perpanjangan atau catatan klausul khusus..."
                  class="h-12 rounded-xl bg-slate-50 border-none font-bold"
                />
              </div>
            </div>

            <div
              class="flex items-center gap-4 p-6 bg-amber-50 rounded-2xl border border-amber-100"
            >
              <div
                class="w-10 h-10 bg-amber-500 rounded-xl flex items-center justify-center shrink-0"
              >
                <Printer class="w-5 h-5 text-white" />
              </div>
              <p class="text-xs font-bold text-amber-800 leading-relaxed">
                Setelah disetujui & diterbitkan, sistem akan menyimpan draf kontrak dan karyawan dapat menandatanganinya secara digital atau mengunduh surat resmi.
              </p>
            </div>
          </div>
        </div>

        <!-- Footer -->
        <div
          class="p-8 bg-white border-t border-slate-100 flex items-center justify-between shrink-0"
        >
          <Button
            variant="ghost"
            @click="prevStep"
            :disabled="currentStep === 1"
            class="rounded-2xl h-12 px-8 font-black text-slate-400 gap-2"
          >
            <ArrowLeft class="w-4 h-4" /> KEMBALI
          </Button>

          <div class="flex items-center gap-3">
            <Button
              variant="outline"
              @click="isCreateModalOpen = false"
              class="rounded-2xl h-12 px-8 font-black text-slate-500 border-none"
              >BATAL</Button
            >

            <Button
              v-if="currentStep < 5"
              @click="nextStep"
              class="rounded-2xl h-12 px-10 font-black bg-slate-900 hover:bg-slate-800 text-white gap-2 shadow-xl shadow-slate-200"
            >
              LANJUT <ArrowRight class="w-4 h-4" />
            </Button>

            <div v-else class="flex items-center gap-3">
              <Button
                @click="saveDocument('PENDING_HR')"
                :disabled="isSubmitting"
                variant="outline"
                class="rounded-2xl h-12 px-6 font-black text-slate-700 border-2 border-slate-200"
              >
                {{ isSubmitting ? 'MENYIMPAN...' : 'SIMPAN DRAFT' }}
              </Button>
              <Button
                @click="saveDocument('APPROVED_DRAFT')"
                :disabled="isSubmitting"
                class="rounded-2xl h-12 px-10 font-black bg-primary hover:bg-primary/90 text-white gap-2 shadow-xl shadow-primary/20"
              >
                {{ isSubmitting ? 'MENYIMPAN...' : 'SETUJUI & TERBITKAN' }} <Printer class="w-4 h-4" />
              </Button>
            </div>
          </div>
        </div>
      </DialogContent>
    </Dialog>

    <!-- Preview Modal -->
    <Dialog v-model:open="isPreviewModalOpen">
      <DialogContent
        class="sm:max-w-5xl rounded-[2.5rem] p-0 overflow-hidden border-none shadow-2xl max-h-[95vh] flex flex-col bg-slate-100"
      >
        <div
          class="bg-white p-6 border-b border-slate-200 flex items-center justify-between shrink-0"
        >
          <div>
            <h2 class="text-xl font-black">Pratinjau Dokumen PKWT</h2>
            <p class="text-slate-500 text-[10px] font-bold uppercase tracking-widest">
              Pastikan data sudah sesuai sebelum dicetak
            </p>
          </div>
          <div class="flex items-center gap-3">
            <Button
              variant="outline"
              @click="isPreviewModalOpen = false"
              class="rounded-xl h-10 px-6 font-bold border-2 border-slate-100"
              >TUTUP</Button
            >
            <Button
              @click="printDocument"
              class="rounded-xl h-10 px-6 font-black bg-primary hover:bg-primary/90 text-white gap-2 shadow-lg shadow-primary/20"
            >
              CETAK SEKARANG <Printer class="w-4 h-4" />
            </Button>
          </div>
        </div>

        <div class="flex-1 overflow-y-auto p-12 flex justify-center bg-slate-200/50">
          <div id="pkwt-print-area">
            <PKWTRenderer
              :data="{
                documentNumber: form.documentNumber,
                agreementDay: form.agreementDay,
                agreementDate: form.agreementDate,
                agreementPlace: form.agreementPlace,
                representativeName: form.representativeName,
                representativePosition: form.representativePosition,
                employee: selectedEmployee,
                startDate: form.startDate,
                endDate: form.endDate,
                jobPosition: form.jobPosition,
                basicSalary: Number(form.basicSalary),
                allowances: Number(form.allowances),
                employeeSignatureUrl: form.employeeSignatureUrl,
              }"
            />
          </div>
        </div>
      </DialogContent>
    </Dialog>
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
</style>
