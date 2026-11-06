<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { toast } from 'vue-sonner'
import employeeRequestApi, {
  type EmployeeRequest,
  type EmployeeRequestCandidate,
  type EmployeeRequestSummary,
  type CandidatePayload,
} from '@/api/employeeRequest'
import { useAuthStore } from '@/stores/auth'
import { useMasterDataStore } from '@/stores/masterData'
import { compressImageFile, formatFileSize } from '@/utils/imageCompressor'
import {
  UserPlus,
  Users,
  Clock,
  CheckCircle2,
  XCircle,
  AlertCircle,
  Search,
  Filter,
  Plus,
  Calendar,
  Building2,
  MapPin,
  Briefcase,
  GraduationCap,
  FileText,
  TrendingUp,
  Eye,
  Check,
  X,
  Edit,
  Trash2,
  ChevronRight,
  Sparkles,
  Info,
  ShieldCheck,
  Award,
  UploadCloud,
  FileCheck,
  ExternalLink,
  Phone,
  Paperclip,
} from 'lucide-vue-next'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import ConfirmationDialog from '@/components/ui/ConfirmationDialog.vue'

const authStore = useAuthStore()
const masterDataStore = useMasterDataStore()

// State
const requests = ref<EmployeeRequest[]>([])
const summary = ref<EmployeeRequestSummary>({
  total_requests: 0,
  pending_requests: 0,
  waiting_hr_approval: 0,
  waiting_director_approval: 0,
  approved_requests: 0,
  in_progress_requests: 0,
  fulfilled_requests: 0,
  rejected_requests: 0,
  total_headcount_needed: 0,
  total_headcount_fulfilled: 0,
})
const isLoading = ref(false)
const isSubmitting = ref(false)

// Filters
const activeStatusTab = ref('ALL')
const selectedBranchFilter = ref('ALL')
const selectedDepartmentFilter = ref('ALL')
const searchQuery = ref('')

const statusTabs = [
  { key: 'ALL', label: 'Semua Status' },
  { key: 'WAITING_HR', label: 'Menunggu Review HR' },
  { key: 'WAITING_DIRECTOR', label: 'Menunggu Direktur' },
  { key: 'APPROVED', label: 'Disetujui Penuh' },
  { key: 'IN_PROGRESS', label: 'Dalam Rekrutmen' },
  { key: 'FULFILLED', label: 'Selesai Terpenuhi' },
  { key: 'REJECTED', label: 'Ditolak' },
]

// Modal Form State (Create / Edit)
const isFormModalOpen = ref(false)
const isEditMode = ref(false)
const formId = ref<string | null>(null)
const formData = ref({
  department_id: '',
  branch_id: '',
  custom_branch_name: '',
  job_position_id: '',
  custom_position_title: '',
  employment_type: 'FULL_TIME',
  request_reason: 'NEW_HEADCOUNT',
  replacement_for_name: '',
  headcount_needed: 1,
  min_education: 'SMA/SMK',
  gender_preference: 'ANY',
  min_age: 20,
  max_age: 35,
  min_experience_years: 1,
  required_skills: '',
  job_description: '',
  notes: '',
  target_date: '',
})

// Modal Detail State
const isDetailModalOpen = ref(false)
const selectedRequest = ref<EmployeeRequest | null>(null)

// Modal Approval State
const isHRApproveModalOpen = ref(false)
const isDirectorApproveModalOpen = ref(false)
const isRejectModalOpen = ref(false)
const approvalNotes = ref('')
const rejectReason = ref('')
const rejectRole = ref<'HR_MANAGER' | 'DIRECTOR'>('HR_MANAGER')

// Modal Progress & Candidate Input State (Bulk supported)
interface CandidateFormItem {
  candidate_name: string
  candidate_phone: string
  candidate_email: string
  join_date: string
  document_url: string
  document_name: string
  document_type: string
  notes: string
  isUploading?: boolean
}

const isProgressModalOpen = ref(false)
const progressStatus = ref('IN_PROGRESS')
const progressRecruitmentNotes = ref('')
const candidateList = ref<CandidateFormItem[]>([])

// Document Preview Modal State
const isDocPreviewOpen = ref(false)
const previewDocUrl = ref('')
const previewDocName = ref('')
const previewDocType = ref('')

// Confirmation Delete State
const isDeleteConfirmOpen = ref(false)
const itemToDeleteId = ref<string | null>(null)

// Role & Permissions Detection
const userRole = computed(() => authStore.user?.role || '')
const userJobTitle = computed(() => {
  const emp = authStore.user?.employee
  return (emp?.job_position?.title || '').toLowerCase()
})

const isSuperAdmin = computed(() => userRole.value === 'superadmin')

const isHRManager = computed(() => {
  if (isSuperAdmin.value) return true
  if (userRole.value === 'hr_admin' || userRole.value === 'HR') return true
  const title = userJobTitle.value
  return (
    title.includes('hr') ||
    title.includes('human resource') ||
    title.includes('personalia') ||
    title.includes('recruiter')
  )
})

const isDirector = computed(() => {
  if (isSuperAdmin.value) return true
  const title = userJobTitle.value
  return (
    title.includes('direktur') ||
    title.includes('director') ||
    title.includes('direksi') ||
    title.includes('ceo') ||
    title.includes('coo') ||
    title.includes('cfo') ||
    title.includes('komisaris') ||
    userRole.value === 'manager'
  )
})

// Fetch Data
const fetchData = async () => {
  isLoading.value = true
  try {
    const params: any = {}
    if (activeStatusTab.value !== 'ALL') {
      params.status = activeStatusTab.value
    }
    if (selectedBranchFilter.value !== 'ALL') {
      params.branch_id = selectedBranchFilter.value
    }
    if (selectedDepartmentFilter.value !== 'ALL') {
      params.department_id = selectedDepartmentFilter.value
    }
    if (searchQuery.value.trim()) {
      params.search = searchQuery.value.trim()
    }

    const [resList, resSummary] = await Promise.all([
      employeeRequestApi.getEmployeeRequests(params),
      employeeRequestApi.getEmployeeRequestSummary({
        branch_id: selectedBranchFilter.value !== 'ALL' ? selectedBranchFilter.value : undefined,
        department_id: selectedDepartmentFilter.value !== 'ALL' ? selectedDepartmentFilter.value : undefined,
      }),
    ])

    requests.value = resList.data.data || []
    if (resSummary.data.data) {
      summary.value = resSummary.data.data
    }

    if (selectedRequest.value) {
      const updated = requests.value.find((r) => r.id === selectedRequest.value?.id)
      if (updated) selectedRequest.value = updated
    }
  } catch (error: any) {
    toast.error('Gagal memuat data permintaan karyawan: ' + (error.response?.data?.error || error.message))
  } finally {
    isLoading.value = false
  }
}

// Watchers
watch([activeStatusTab, selectedBranchFilter, selectedDepartmentFilter], () => {
  fetchData()
})

let searchTimeout: any = null
watch(searchQuery, () => {
  clearTimeout(searchTimeout)
  searchTimeout = setTimeout(() => {
    fetchData()
  }, 350)
})

onMounted(async () => {
  if (masterDataStore.branches.length === 0) masterDataStore.fetchBranches()
  if (masterDataStore.departments.length === 0) masterDataStore.fetchDepartments()
  if (masterDataStore.jobPositions.length === 0) masterDataStore.fetchJobPositions()
  await fetchData()
})

// Open Create Modal
const openCreateModal = () => {
  isEditMode.value = false
  formId.value = null

  const defaultDate = new Date()
  defaultDate.setDate(defaultDate.getDate() + 14)
  const defaultDateStr = defaultDate.toISOString().split('T')[0] || ''

  formData.value = {
    department_id: authStore.user?.department_id || (masterDataStore.departments[0]?.id || ''),
    branch_id: authStore.user?.branch_id || (masterDataStore.branches[0]?.id || ''),
    custom_branch_name: '',
    job_position_id: '',
    custom_position_title: '',
    employment_type: 'FULL_TIME',
    request_reason: 'NEW_HEADCOUNT',
    replacement_for_name: '',
    headcount_needed: 1,
    min_education: 'SMA/SMK',
    gender_preference: 'ANY',
    min_age: 20,
    max_age: 35,
    min_experience_years: 1,
    required_skills: '',
    job_description: '',
    notes: '',
    target_date: defaultDateStr,
  }
  isFormModalOpen.value = true
}

// Open Edit Modal
const openEditModal = (req: EmployeeRequest) => {
  isEditMode.value = true
  formId.value = req.id

  let initialBranchId = req.branch_id || ''
  if (!initialBranchId && req.custom_branch_name) {
    initialBranchId = 'OTHER'
  }

  formData.value = {
    department_id: req.department_id || '',
    branch_id: initialBranchId,
    custom_branch_name: req.custom_branch_name || '',
    job_position_id: req.job_position_id || '',
    custom_position_title: req.custom_position_title || '',
    employment_type: req.employment_type || 'FULL_TIME',
    request_reason: req.request_reason || 'NEW_HEADCOUNT',
    replacement_for_name: req.replacement_for_name || '',
    headcount_needed: req.headcount_needed || 1,
    min_education: req.min_education || 'SMA/SMK',
    gender_preference: req.gender_preference || 'ANY',
    min_age: req.min_age || 18,
    max_age: req.max_age || 45,
    min_experience_years: req.min_experience_years || 0,
    required_skills: req.required_skills || '',
    job_description: req.job_description || '',
    notes: req.notes || '',
    target_date: (req.target_date ? req.target_date.split('T')[0] : '') || '',
  }
  isFormModalOpen.value = true
}

// Submit Form
const handleSubmitForm = async () => {
  if (!formData.value.target_date) {
    toast.error('Tanggal target/deadline karyawan ready wajib diisi')
    return
  }
  if (!formData.value.job_position_id && !formData.value.custom_position_title) {
    toast.error('Pilih jabatan atau tuliskan judul posisi yang dibutuhkan')
    return
  }
  if (!formData.value.branch_id) {
    toast.error('Pilih cabang penempatan atau pilih Lainnya')
    return
  }
  if (formData.value.branch_id === 'OTHER' && !formData.value.custom_branch_name.trim()) {
    toast.error('Tuliskan nama lokasi / cabang penempatan kustom')
    return
  }
  if (formData.value.headcount_needed <= 0) {
    toast.error('Jumlah kebutuhan karyawan minimal 1 orang')
    return
  }

  isSubmitting.value = true
  try {
    const payload: any = { ...formData.value }
    if (!payload.job_position_id) delete payload.job_position_id
    if (!payload.department_id) delete payload.department_id

    if (payload.branch_id === 'OTHER') {
      delete payload.branch_id
      payload.custom_branch_name = payload.custom_branch_name.trim()
    } else {
      payload.custom_branch_name = ''
    }

    if (isEditMode.value && formId.value) {
      await employeeRequestApi.updateEmployeeRequest(formId.value, payload)
      toast.success('Permintaan karyawan berhasil diperbarui')
    } else {
      await employeeRequestApi.createEmployeeRequest(payload)
      toast.success('Permintaan karyawan berhasil diajukan!')
    }
    isFormModalOpen.value = false
    fetchData()
  } catch (err: any) {
    toast.error('Gagal menyimpan: ' + (err.response?.data?.error || err.message))
  } finally {
    isSubmitting.value = false
  }
}

// Open Detail Modal
const openDetailModal = (req: EmployeeRequest) => {
  selectedRequest.value = req
  isDetailModalOpen.value = true
}

// Open HR Approval Modal
const openHRApproveModal = (req: EmployeeRequest) => {
  selectedRequest.value = req
  approvalNotes.value = ''
  isHRApproveModalOpen.value = true
}

const handleHRApproveConfirm = async () => {
  if (!selectedRequest.value) return
  isSubmitting.value = true
  try {
    await employeeRequestApi.approveByHR(selectedRequest.value.id, approvalNotes.value)
    toast.success(`Persetujuan HR Manager untuk ${selectedRequest.value.request_number} berhasil disimpan`)
    isHRApproveModalOpen.value = false
    fetchData()
  } catch (err: any) {
    toast.error('Gagal menyetujui: ' + (err.response?.data?.error || err.message))
  } finally {
    isSubmitting.value = false
  }
}

// Open Director Approval Modal
const openDirectorApproveModal = (req: EmployeeRequest) => {
  selectedRequest.value = req
  approvalNotes.value = ''
  isDirectorApproveModalOpen.value = true
}

const handleDirectorApproveConfirm = async () => {
  if (!selectedRequest.value) return
  isSubmitting.value = true
  try {
    await employeeRequestApi.approveByDirector(selectedRequest.value.id, approvalNotes.value)
    toast.success(`Persetujuan Direktur untuk ${selectedRequest.value.request_number} berhasil disimpan`)
    isDirectorApproveModalOpen.value = false
    fetchData()
  } catch (err: any) {
    toast.error('Gagal menyetujui: ' + (err.response?.data?.error || err.message))
  } finally {
    isSubmitting.value = false
  }
}

// Open Rejection Modal
const openRejectModal = (req: EmployeeRequest, role: 'HR_MANAGER' | 'DIRECTOR' = 'HR_MANAGER') => {
  selectedRequest.value = req
  rejectReason.value = ''
  rejectRole.value = role
  isRejectModalOpen.value = true
}

const handleRejectConfirm = async () => {
  if (!selectedRequest.value) return
  if (!rejectReason.value.trim()) {
    toast.error('Alasan penolakan wajib diisi')
    return
  }
  isSubmitting.value = true
  try {
    await employeeRequestApi.rejectEmployeeRequest(selectedRequest.value.id, rejectReason.value, rejectRole.value)
    toast.success(`Pengajuan ${selectedRequest.value.request_number} telah ditolak`)
    isRejectModalOpen.value = false
    fetchData()
  } catch (err: any) {
    toast.error('Gagal menolak: ' + (err.response?.data?.error || err.message))
  } finally {
    isSubmitting.value = false
  }
}

// Open Progress & Candidate Bulk Input Modal
const openProgressModal = (req: EmployeeRequest) => {
  selectedRequest.value = req
  progressStatus.value = req.status === 'APPROVED' ? 'IN_PROGRESS' : req.status
  progressRecruitmentNotes.value = req.recruitment_notes || ''

  const todayStr = new Date().toISOString().split('T')[0] || ''

  // Pre-fill initial candidates slots if headcount not yet filled
  const existingCount = req.candidates?.length || 0
  const neededRemaining = Math.max(1, req.headcount_needed - existingCount)

  candidateList.value = []
  for (let i = 0; i < neededRemaining; i++) {
    candidateList.value.push({
      candidate_name: '',
      candidate_phone: '',
      candidate_email: '',
      join_date: todayStr,
      document_url: '',
      document_name: '',
      document_type: '',
      notes: '',
      isUploading: false,
    })
  }

  isProgressModalOpen.value = true
}

const addCandidateRow = () => {
  const todayStr = new Date().toISOString().split('T')[0] || ''
  candidateList.value.push({
    candidate_name: '',
    candidate_phone: '',
    candidate_email: '',
    join_date: todayStr,
    document_url: '',
    document_name: '',
    document_type: '',
    notes: '',
    isUploading: false,
  })
}

const removeCandidateRow = (index: number) => {
  if (candidateList.value.length > 1) {
    candidateList.value.splice(index, 1)
  }
}

// File Upload Handler for Candidate CV / Photo
const handleFileUpload = async (event: Event, index: number) => {
  const target = event.target as HTMLInputElement
  if (!target.files || target.files.length === 0) return

  let file = target.files[0]
  if (!file) return

  const item = candidateList.value[index]
  if (!item) return

  item.isUploading = true
  try {
    // If it is an image, compress client-side to save bandwidth & VPS storage
    if (file.type.startsWith('image/')) {
      const { compressedFile, originalSize, compressedSize } = await compressImageFile(file, 1600, 1600, 0.8)
      file = compressedFile
      if (compressedSize < originalSize) {
        toast.info(`Foto dioptimasi: ${formatFileSize(originalSize)} → ${formatFileSize(compressedSize)}`)
      }
    }

    const res = await employeeRequestApi.uploadCandidateDocument(file)
    item.document_url = res.data.document_url
    item.document_name = res.data.document_name
    item.document_type = res.data.document_type
    toast.success(`Dokumen ${res.data.document_name} berhasil diunggah`)
  } catch (err: any) {
    toast.error('Gagal mengunggah file: ' + (err.response?.data?.error || err.message))
    target.value = ''
  } finally {
    item.isUploading = false
  }
}

const removeCandidateFile = (index: number) => {
  const item = candidateList.value[index]
  if (item) {
    item.document_url = ''
    item.document_name = ''
    item.document_type = ''
  }
}

const handleProgressSubmit = async () => {
  if (!selectedRequest.value) return

  // Filter valid candidate rows that have at least candidate_name
  const validCandidates: CandidatePayload[] = candidateList.value
    .filter((c) => c.candidate_name.trim() !== '')
    .map((c) => ({
      candidate_name: c.candidate_name.trim(),
      candidate_phone: c.candidate_phone.trim(),
      candidate_email: c.candidate_email.trim(),
      join_date: c.join_date,
      document_url: c.document_url,
      document_name: c.document_name,
      document_type: c.document_type,
      notes: c.notes.trim(),
    }))

  if (progressStatus.value === 'FULFILLED' && validCandidates.length === 0) {
    toast.error('Mohon isi minimal 1 data nama karyawan baru yang diterima untuk status Selesai Terpenuhi')
    return
  }

  isSubmitting.value = true
  try {
    await employeeRequestApi.updateProgress(selectedRequest.value.id, {
      status: progressStatus.value,
      recruitment_notes: progressRecruitmentNotes.value,
      candidates: progressStatus.value === 'FULFILLED' && validCandidates.length > 0 ? validCandidates : undefined,
    })
    toast.success('Data progres dan status rekrutmen berhasil disimpan')
    isProgressModalOpen.value = false
    fetchData()
  } catch (err: any) {
    toast.error('Gagal menyimpan: ' + (err.response?.data?.error || err.message))
  } finally {
    isSubmitting.value = false
  }
}

// Delete Candidate from Detail Modal
const handleDeleteCandidate = async (candidateId: string) => {
  if (!selectedRequest.value) return
  if (!confirm('Apakah Anda yakin ingin menghapus data karyawan ini dari pengajuan?')) return

  isSubmitting.value = true
  try {
    const res = await employeeRequestApi.deleteCandidate(selectedRequest.value.id, candidateId)
    selectedRequest.value = res.data.data
    toast.success('Data karyawan berhasil dihapus')
    fetchData()
  } catch (err: any) {
    toast.error('Gagal menghapus kandidat: ' + (err.response?.data?.error || err.message))
  } finally {
    isSubmitting.value = false
  }
}

// Open Document Preview
const openDocPreview = (cand: EmployeeRequestCandidate) => {
  if (!cand.document_url) return
  previewDocUrl.value = cand.document_url
  previewDocName.value = cand.document_name || 'Dokumen CV / Foto'
  previewDocType.value = cand.document_type || 'PDF'
  isDocPreviewOpen.value = true
}

// Open Delete Confirm
const openDeleteDialog = (id: string) => {
  itemToDeleteId.value = id
  isDeleteConfirmOpen.value = true
}

const handleDeleteConfirm = async () => {
  if (!itemToDeleteId.value) return
  isSubmitting.value = true
  try {
    await employeeRequestApi.deleteEmployeeRequest(itemToDeleteId.value)
    toast.success('Permintaan karyawan berhasil dihapus')
    isDeleteConfirmOpen.value = false
    fetchData()
  } catch (err: any) {
    toast.error('Gagal menghapus: ' + (err.response?.data?.error || err.message))
  } finally {
    isSubmitting.value = false
  }
}

// Helpers
const formatDate = (d?: string) => {
  if (!d) return '-'
  return new Date(d).toLocaleDateString('id-ID', {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
  })
}

const getPositionTitle = (req: EmployeeRequest) => {
  if (req.job_position?.title) return req.job_position.title
  if (req.custom_position_title) return req.custom_position_title
  return 'Posisi Karyawan'
}

const getBranchName = (req?: EmployeeRequest | null) => {
  if (!req) return 'Pusat'
  if (req.branch?.name) return req.branch.name
  if (req.custom_branch_name) return req.custom_branch_name
  return 'Pusat'
}

const getReasonLabel = (reason: string) => {
  switch (reason) {
    case 'NEW_HEADCOUNT':
      return 'Penambahan Tim Baru'
    case 'REPLACEMENT':
      return 'Penggantian (Replacement)'
    case 'PROJECT_BASED':
      return 'Kebutuhan Proyek'
    case 'WORKLOAD_INCREASE':
      return 'Beban Kerja Meningkat'
    default:
      return 'Lainnya'
  }
}

const getEmploymentTypeLabel = (type: string) => {
  switch (type) {
    case 'FULL_TIME':
      return 'Tetap (PKWTT)'
    case 'CONTRACT':
      return 'Kontrak (PKWT)'
    case 'DAILY':
      return 'Pekerja Harian'
    case 'INTERNSHIP':
      return 'Magang / Internship'
    default:
      return type
  }
}

const getStatusBadge = (status: string) => {
  switch (status) {
    case 'PENDING':
      return { label: 'Menunggu Approval', class: 'bg-amber-50 text-amber-700 border-amber-200' }
    case 'WAITING_DIRECTOR_APPROVAL':
      return { label: 'Menunggu Direktur', class: 'bg-indigo-50 text-indigo-700 border-indigo-200' }
    case 'WAITING_HR_APPROVAL':
      return { label: 'Menunggu HR', class: 'bg-amber-50 text-amber-700 border-amber-200' }
    case 'APPROVED':
      return { label: 'Disetujui Penuh', class: 'bg-blue-50 text-blue-700 border-blue-200' }
    case 'IN_PROGRESS':
      return { label: 'Dalam Rekrutmen', class: 'bg-purple-50 text-purple-700 border-purple-200' }
    case 'FULFILLED':
      return { label: 'Selesai Terpenuhi', class: 'bg-emerald-50 text-emerald-700 border-emerald-200' }
    case 'REJECTED':
      return { label: 'Ditolak', class: 'bg-rose-50 text-rose-700 border-rose-200' }
    case 'CANCELLED':
      return { label: 'Dibatalkan', class: 'bg-slate-50 text-slate-600 border-slate-200' }
    default:
      return { label: status, class: 'bg-slate-50 text-slate-700 border-slate-200' }
  }
}

const isDeadlineNearOrPassed = (targetDateStr: string, status: string) => {
  if (status === 'FULFILLED' || status === 'CANCELLED' || status === 'REJECTED') return 'normal'
  const target = new Date(targetDateStr).getTime()
  const now = new Date().getTime()
  const diffDays = Math.ceil((target - now) / (1000 * 60 * 60 * 24))
  if (diffDays < 0) return 'passed'
  if (diffDays <= 3) return 'urgent'
  return 'normal'
}
</script>

<template>
  <div class="space-y-6 pb-12">
    <!-- Header Section -->
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
      <div>
        <div class="flex items-center gap-2.5">
          <div class="p-2.5 bg-primary/10 rounded-2xl text-primary shadow-xs">
            <UserPlus class="w-6 h-6" />
          </div>
          <div>
            <h1 class="text-2xl font-black text-slate-900 tracking-tight">Permintaan Karyawan</h1>
            <p class="text-xs font-semibold text-slate-500 mt-0.5">
              Pengajuan kebutuhan tenaga kerja per Kepala Divisi, persetujuan HR & Direktur, serta pelaporan data CV karyawan baru.
            </p>
          </div>
        </div>
      </div>

      <div class="flex items-center gap-3">
        <a
          href="/login"
          target="_blank"
          rel="noopener noreferrer"
          class="inline-flex items-center gap-2 rounded-xl border border-slate-200 bg-white hover:bg-slate-50 text-slate-700 px-4 py-2.5 font-bold text-xs shadow-xs transition-all hover:scale-[1.02]"
          title="Buka Portal eRecruitment"
        >
          <Briefcase class="w-4 h-4 text-primary" />
          <span>Portal eRecruitment</span>
          <ExternalLink class="w-3.5 h-3.5 text-slate-400" />
        </a>

        <Button
          @click="openCreateModal"
          class="rounded-xl px-4 py-2.5 h-auto bg-primary hover:bg-primary/90 text-white font-bold text-xs shadow-md shadow-primary/20 flex items-center gap-2 transition-all hover:scale-[1.02]"
        >
          <Plus class="w-4 h-4" />
          <span>Ajukan Permintaan</span>
        </Button>
      </div>
    </div>

    <!-- Summary KPI Cards -->
    <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-5 gap-3.5">
      <!-- Total Permintaan -->
      <div class="bg-white p-4 rounded-2xl border border-slate-200/80 shadow-xs flex items-center justify-between">
        <div>
          <p class="text-[10px] font-extrabold uppercase tracking-wider text-slate-400">Total Pengajuan</p>
          <h3 class="text-xl font-black text-slate-900 mt-1">{{ summary.total_requests }}</h3>
          <p class="text-[10px] font-semibold text-slate-500 mt-0.5">Semua pengajuan</p>
        </div>
        <div class="w-10 h-10 rounded-xl bg-slate-100 flex items-center justify-center text-slate-600">
          <FileText class="w-5 h-5" />
        </div>
      </div>

      <!-- Menunggu Review HR -->
      <div class="bg-white p-4 rounded-2xl border border-amber-200/80 bg-amber-50/20 shadow-xs flex items-center justify-between">
        <div>
          <p class="text-[10px] font-extrabold uppercase tracking-wider text-amber-600">Menunggu HR</p>
          <h3 class="text-xl font-black text-amber-700 mt-1">{{ summary.waiting_hr_approval }}</h3>
          <p class="text-[10px] font-semibold text-amber-600/80 mt-0.5">Tahap 1 Review HR</p>
        </div>
        <div class="w-10 h-10 rounded-xl bg-amber-100/80 flex items-center justify-center text-amber-600">
          <Clock class="w-5 h-5" />
        </div>
      </div>

      <!-- Menunggu Review Direktur -->
      <div class="bg-white p-4 rounded-2xl border border-indigo-200/80 bg-indigo-50/20 shadow-xs flex items-center justify-between">
        <div>
          <p class="text-[10px] font-extrabold uppercase tracking-wider text-indigo-600">Menunggu Direktur</p>
          <h3 class="text-xl font-black text-indigo-700 mt-1">{{ summary.waiting_director_approval }}</h3>
          <p class="text-[10px] font-semibold text-indigo-600/80 mt-0.5">Tahap 2 Otorisasi Direksi</p>
        </div>
        <div class="w-10 h-10 rounded-xl bg-indigo-100/80 flex items-center justify-center text-indigo-600">
          <ShieldCheck class="w-5 h-5" />
        </div>
      </div>

      <!-- Disetujui / Rekrutmen -->
      <div class="bg-white p-4 rounded-2xl border border-blue-200/80 bg-blue-50/20 shadow-xs flex items-center justify-between">
        <div>
          <p class="text-[10px] font-extrabold uppercase tracking-wider text-blue-600">Dalam Rekrutmen</p>
          <h3 class="text-xl font-black text-blue-700 mt-1">
            {{ summary.approved_requests + summary.in_progress_requests }}
          </h3>
          <p class="text-[10px] font-semibold text-blue-600/80 mt-0.5">Sourcing & interview</p>
        </div>
        <div class="w-10 h-10 rounded-xl bg-blue-100/80 flex items-center justify-center text-blue-600">
          <TrendingUp class="w-5 h-5" />
        </div>
      </div>

      <!-- Total Headcount Pemenuhan -->
      <div class="bg-white p-4 rounded-2xl border border-emerald-200/80 bg-emerald-50/20 shadow-xs flex flex-col justify-between">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-[10px] font-extrabold uppercase tracking-wider text-emerald-700">Pemenuhan Headcount</p>
            <h3 class="text-xl font-black text-emerald-800 mt-1">
              {{ summary.total_headcount_fulfilled }} / {{ summary.total_headcount_needed }}
              <span class="text-[11px] font-bold text-emerald-600">Orang</span>
            </h3>
          </div>
          <div class="w-10 h-10 rounded-xl bg-emerald-100/80 flex items-center justify-center text-emerald-700">
            <Users class="w-5 h-5" />
          </div>
        </div>
        <!-- Mini Progress Bar -->
        <div class="w-full bg-emerald-100 rounded-full h-1.5 mt-2.5 overflow-hidden">
          <div
            class="bg-emerald-600 h-full rounded-full transition-all duration-500"
            :style="{
              width: summary.total_headcount_needed > 0
                ? `${Math.min(100, Math.round((summary.total_headcount_fulfilled / summary.total_headcount_needed) * 100))}%`
                : '0%',
            }"
          ></div>
        </div>
      </div>
    </div>

    <!-- Filters & Search Bar -->
    <div class="bg-white p-4 rounded-2xl border border-slate-200/80 shadow-xs space-y-3">
      <!-- Status Tabs -->
      <div class="flex items-center gap-1.5 overflow-x-auto no-scrollbar pb-1">
        <button
          v-for="tab in statusTabs"
          :key="tab.key"
          @click="activeStatusTab = tab.key"
          class="px-3 py-1.5 rounded-xl text-xs font-bold whitespace-nowrap transition-all"
          :class="[
            activeStatusTab === tab.key
              ? 'bg-slate-900 text-white shadow-xs'
              : 'text-slate-600 hover:bg-slate-100 hover:text-slate-900',
          ]"
        >
          {{ tab.label }}
        </button>
      </div>

      <div class="flex flex-col sm:flex-row items-center gap-3 pt-2 border-t border-slate-100">
        <!-- Search Input -->
        <div class="relative flex-1 w-full">
          <Search class="absolute left-3.5 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400" />
          <Input
            v-model="searchQuery"
            placeholder="Cari posisi, no. pengajuan, nama pemohon, jabatan..."
            class="pl-10 h-10 rounded-xl bg-slate-50/70 border-slate-200 text-xs font-semibold focus:bg-white transition-all w-full"
          />
        </div>

        <!-- Filter Cabang -->
        <div class="w-full sm:w-48">
          <select
            v-model="selectedBranchFilter"
            class="w-full h-10 rounded-xl bg-slate-50/70 border border-slate-200 px-3 text-xs font-semibold text-slate-700 focus:outline-none focus:ring-2 focus:ring-primary/20"
          >
            <option value="ALL">Semua Cabang</option>
            <option v-for="b in masterDataStore.branches" :key="b.id" :value="b.id">
              {{ b.name }}
            </option>
          </select>
        </div>

        <!-- Filter Departemen -->
        <div class="w-full sm:w-48">
          <select
            v-model="selectedDepartmentFilter"
            class="w-full h-10 rounded-xl bg-slate-50/70 border border-slate-200 px-3 text-xs font-semibold text-slate-700 focus:outline-none focus:ring-2 focus:ring-primary/20"
          >
            <option value="ALL">Semua Divisi</option>
            <option v-for="d in masterDataStore.departments" :key="d.id" :value="d.id">
              {{ d.name }}
            </option>
          </select>
        </div>
      </div>
    </div>

    <!-- Data Table -->
    <div class="bg-white rounded-2xl border border-slate-200/80 shadow-xs overflow-hidden">
      <div v-if="isLoading" class="p-12 text-center text-slate-400 text-xs font-semibold flex flex-col items-center gap-3">
        <div class="w-8 h-8 border-3 border-primary/30 border-t-primary rounded-full animate-spin"></div>
        <span>Memuat data permintaan karyawan...</span>
      </div>

      <div v-else-if="requests.length === 0" class="p-12 text-center text-slate-400 flex flex-col items-center gap-3">
        <div class="w-14 h-14 rounded-2xl bg-slate-100 flex items-center justify-center text-slate-400">
          <UserPlus class="w-7 h-7" />
        </div>
        <div>
          <p class="text-sm font-bold text-slate-700">Belum ada pengajuan permintaan karyawan</p>
          <p class="text-xs text-slate-400 mt-0.5">Klik tombol 'Ajukan Permintaan' untuk membuat pengajuan baru.</p>
        </div>
      </div>

      <div v-else class="overflow-x-auto">
        <table class="w-full text-left text-xs border-collapse">
          <thead>
            <tr class="bg-slate-50/80 text-slate-500 font-extrabold uppercase tracking-wider border-b border-slate-200/70">
              <th class="py-3.5 px-4">No. Pengajuan</th>
              <th class="py-3.5 px-4">Pemohon & Jabatan</th>
              <th class="py-3.5 px-4">Posisi & Divisi</th>
              <th class="py-3.5 px-4">Headcount</th>
              <th class="py-3.5 px-4">Target Ready</th>
              <th class="py-3.5 px-4">Persetujuan Ganda</th>
              <th class="py-3.5 px-4 text-right">Aksi</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-100 font-medium text-slate-700">
            <tr
              v-for="req in requests"
              :key="req.id"
              class="hover:bg-slate-50/80 transition-colors"
            >
              <!-- No Pengajuan -->
              <td class="py-3.5 px-4">
                <div class="flex flex-col">
                  <span class="font-extrabold text-slate-900 font-mono text-[13px]">{{ req.request_number }}</span>
                  <span class="text-[10px] text-slate-400">{{ formatDate(req.created_at) }}</span>
                </div>
              </td>

              <!-- Pemohon & Jabatan (Jelas) -->
              <td class="py-3.5 px-4">
                <div class="flex flex-col gap-0.5">
                  <span class="font-bold text-slate-900">{{ req.requester_name }}</span>
                  <span class="text-[11px] font-semibold text-primary">
                    {{ req.requester_job_title || req.requester?.job_position?.title || 'Kepala Divisi' }}
                  </span>
                  <span class="text-[10px] text-slate-400">
                    {{ req.department?.name || 'Divisi' }} • {{ getBranchName(req) }}
                  </span>
                </div>
              </td>

              <!-- Posisi & Kebutuhan -->
              <td class="py-3.5 px-4">
                <div class="flex flex-col gap-0.5">
                  <span class="font-extrabold text-slate-900 text-sm">{{ getPositionTitle(req) }}</span>
                  <div class="flex items-center gap-1.5 text-[10px]">
                    <span class="px-2 py-0.5 rounded-md bg-slate-100 text-slate-700 font-semibold">
                      {{ getEmploymentTypeLabel(req.employment_type) }}
                    </span>
                    <span class="text-slate-400 font-medium">
                      {{ getReasonLabel(req.request_reason) }}
                    </span>
                  </div>
                </div>
              </td>

              <!-- Headcount Tracker & New Hires Count -->
              <td class="py-3.5 px-4">
                <div class="flex flex-col gap-1">
                  <div class="flex items-center gap-1 font-bold">
                    <span class="text-emerald-600">{{ req.headcount_fulfilled }}</span>
                    <span class="text-slate-400">/</span>
                    <span class="text-slate-800">{{ req.headcount_needed }}</span>
                    <span class="text-[10px] text-slate-500 font-normal">Orang</span>
                  </div>
                  <!-- Mini Progress Bar -->
                  <div class="w-20 bg-slate-100 rounded-full h-1.5 overflow-hidden">
                    <div
                      class="h-full rounded-full transition-all duration-300"
                      :class="req.headcount_fulfilled >= req.headcount_needed ? 'bg-emerald-500' : 'bg-primary'"
                      :style="{
                        width: req.headcount_needed > 0
                          ? `${Math.min(100, Math.round((req.headcount_fulfilled / req.headcount_needed) * 100))}%`
                          : '0%',
                      }"
                    ></div>
                  </div>
                  <span v-if="req.candidates && req.candidates.length > 0" class="text-[9px] font-bold text-indigo-600 flex items-center gap-0.5 mt-0.5">
                    <Paperclip class="w-2.5 h-2.5" /> {{ req.candidates.length }} CV Terlampir
                  </span>
                </div>
              </td>

              <!-- Target Ready Deadline -->
              <td class="py-3.5 px-4">
                <div class="flex flex-col">
                  <span
                    class="font-bold text-[12px] flex items-center gap-1"
                    :class="{
                      'text-rose-600': isDeadlineNearOrPassed(req.target_date, req.status) === 'passed',
                      'text-amber-600': isDeadlineNearOrPassed(req.target_date, req.status) === 'urgent',
                      'text-slate-700': isDeadlineNearOrPassed(req.target_date, req.status) === 'normal',
                    }"
                  >
                    <Calendar class="w-3.5 h-3.5" />
                    {{ formatDate(req.target_date) }}
                  </span>
                  <span
                    v-if="isDeadlineNearOrPassed(req.target_date, req.status) === 'passed'"
                    class="text-[10px] font-bold text-rose-500"
                  >
                    Lewat Deadline
                  </span>
                  <span
                    v-else-if="isDeadlineNearOrPassed(req.target_date, req.status) === 'urgent'"
                    class="text-[10px] font-bold text-amber-500"
                  >
                    Mendekati Deadline
                  </span>
                </div>
              </td>

              <!-- Dual-Approval Status Badges -->
              <td class="py-3.5 px-4">
                <div class="flex flex-col gap-1.5">
                  <!-- Overall Status Badge -->
                  <div>
                    <span
                      class="px-2 py-0.5 rounded-full text-[10px] font-black border uppercase"
                      :class="getStatusBadge(req.status).class"
                    >
                      {{ getStatusBadge(req.status).label }}
                    </span>
                  </div>

                  <!-- Two-Tier Approval Indicators -->
                  <div class="flex flex-col gap-1 text-[10px]">
                    <!-- HR Approval Indicator -->
                    <div class="flex items-center gap-1">
                      <span class="font-bold text-slate-500 w-14">HR Mgr:</span>
                      <span
                        v-if="req.hr_status === 'APPROVED'"
                        class="text-emerald-700 font-bold bg-emerald-50 border border-emerald-200 px-1.5 py-0.2 rounded"
                      >
                        ✓ Disetujui
                      </span>
                      <span
                        v-else-if="req.hr_status === 'REJECTED'"
                        class="text-rose-700 font-bold bg-rose-50 border border-rose-200 px-1.5 py-0.2 rounded"
                      >
                        ✗ Ditolak
                      </span>
                      <span
                        v-else
                        class="text-amber-700 font-semibold bg-amber-50 border border-amber-200 px-1.5 py-0.2 rounded"
                      >
                        ⏳ Menunggu
                      </span>
                    </div>

                    <!-- Director Approval Indicator -->
                    <div class="flex items-center gap-1">
                      <span class="font-bold text-slate-500 w-14">Direktur:</span>
                      <span
                        v-if="req.director_status === 'APPROVED'"
                        class="text-emerald-700 font-bold bg-emerald-50 border border-emerald-200 px-1.5 py-0.2 rounded"
                      >
                        ✓ Disetujui
                      </span>
                      <span
                        v-else-if="req.director_status === 'REJECTED'"
                        class="text-rose-700 font-bold bg-rose-50 border border-rose-200 px-1.5 py-0.2 rounded"
                      >
                        ✗ Ditolak
                      </span>
                      <span
                        v-else
                        class="text-amber-700 font-semibold bg-amber-50 border border-amber-200 px-1.5 py-0.2 rounded"
                      >
                        ⏳ Menunggu
                      </span>
                    </div>
                  </div>
                </div>
              </td>

              <!-- Actions -->
              <td class="py-3.5 px-4 text-right">
                <div class="flex items-center justify-end gap-1.5 flex-wrap">
                  <!-- Detail Button -->
                  <Button
                    @click="openDetailModal(req)"
                    variant="ghost"
                    size="sm"
                    class="h-8 px-2 rounded-lg text-slate-600 hover:text-slate-900 hover:bg-slate-100 text-xs font-semibold"
                    title="Lihat Detail & Karyawan Masuk"
                  >
                    <Eye class="w-3.5 h-3.5 mr-1" />
                    <span>Detail</span>
                  </Button>

                  <!-- HR Approval Button (If HR has not approved yet & not rejected/fulfilled) -->
                  <template v-if="req.hr_status === 'PENDING' && req.status !== 'REJECTED' && req.status !== 'CANCELLED' && isHRManager">
                    <Button
                      @click="openHRApproveModal(req)"
                      variant="outline"
                      size="sm"
                      class="h-8 px-2.5 rounded-lg border-amber-300 text-amber-800 hover:bg-amber-50 text-xs font-bold"
                    >
                      <Check class="w-3.5 h-3.5 mr-1 text-amber-600" />
                      <span>Setujui (HR)</span>
                    </Button>
                    <Button
                      @click="openRejectModal(req, 'HR_MANAGER')"
                      variant="outline"
                      size="sm"
                      class="h-8 px-2 rounded-lg border-rose-300 text-rose-700 hover:bg-rose-50 text-xs font-bold"
                    >
                      <X class="w-3.5 h-3.5 text-rose-600" />
                    </Button>
                  </template>

                  <!-- Director Approval Button (If Director has not approved yet & not rejected/fulfilled) -->
                  <template v-if="req.director_status === 'PENDING' && req.status !== 'REJECTED' && req.status !== 'CANCELLED' && isDirector">
                    <Button
                      @click="openDirectorApproveModal(req)"
                      variant="outline"
                      size="sm"
                      class="h-8 px-2.5 rounded-lg border-indigo-300 text-indigo-800 hover:bg-indigo-50 text-xs font-bold"
                    >
                      <ShieldCheck class="w-3.5 h-3.5 mr-1 text-indigo-600" />
                      <span>Setujui (Direktur)</span>
                    </Button>
                    <Button
                      v-if="req.hr_status !== 'PENDING'"
                      @click="openRejectModal(req, 'DIRECTOR')"
                      variant="outline"
                      size="sm"
                      class="h-8 px-2 rounded-lg border-rose-300 text-rose-700 hover:bg-rose-50 text-xs font-bold"
                    >
                      <X class="w-3.5 h-3.5 text-rose-600" />
                    </Button>
                  </template>

                  <!-- HR Progress Update & Candidate Entry (When FULLY APPROVED or IN_PROGRESS or FULFILLED) -->
                  <template v-if="(req.status === 'APPROVED' || req.status === 'IN_PROGRESS' || req.status === 'FULFILLED') && isHRManager">
                    <Button
                      @click="openProgressModal(req)"
                      variant="outline"
                      size="sm"
                      class="h-8 px-2.5 rounded-lg border-primary/30 text-primary hover:bg-primary/5 text-xs font-bold"
                    >
                      <TrendingUp class="w-3.5 h-3.5 mr-1" />
                      <span>Input Karyawan / Progres</span>
                    </Button>
                  </template>

                  <!-- Edit / Delete for Requester when PENDING -->
                  <template v-if="req.status === 'PENDING' && req.hr_status === 'PENDING'">
                    <Button
                      @click="openEditModal(req)"
                      variant="ghost"
                      size="icon"
                      class="h-8 w-8 rounded-lg text-slate-400 hover:text-slate-700 hover:bg-slate-100"
                      title="Edit Permintaan"
                    >
                      <Edit class="w-3.5 h-3.5" />
                    </Button>
                    <Button
                      @click="openDeleteDialog(req.id)"
                      variant="ghost"
                      size="icon"
                      class="h-8 w-8 rounded-lg text-slate-400 hover:text-rose-600 hover:bg-rose-50"
                      title="Hapus Permintaan"
                    >
                      <Trash2 class="w-3.5 h-3.5" />
                    </Button>
                  </template>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- MODAL: Form Buat / Edit Permintaan Karyawan -->
    <Dialog v-model:open="isFormModalOpen">
      <DialogContent class="max-w-2xl max-h-[90vh] overflow-y-auto rounded-3xl p-6">
        <DialogHeader class="pb-4 border-b border-slate-100">
          <div class="flex items-center gap-3">
            <div class="p-2.5 rounded-2xl bg-primary/10 text-primary">
              <UserPlus class="w-5 h-5" />
            </div>
            <div>
              <DialogTitle class="text-lg font-black text-slate-900">
                {{ isEditMode ? 'Edit Permintaan Karyawan' : 'Form Pengajuan Kebutuhan Karyawan' }}
              </DialogTitle>
              <DialogDescription class="text-xs text-slate-500 mt-0.5">
                Pengajuan akan diteruskan untuk verifikasi persetujuan HR Manager dan otorisasi Direktur.
              </DialogDescription>
            </div>
          </div>
        </DialogHeader>

        <div class="space-y-6 py-4">
          <!-- Section 1: Divisi & Posisi -->
          <div class="space-y-4">
            <div class="flex items-center gap-2 text-xs font-bold text-slate-800 uppercase tracking-wider">
              <Building2 class="w-4 h-4 text-primary" />
              <span>1. Informasi Penempatan & Posisi</span>
            </div>

            <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
              <!-- Departemen / Divisi -->
              <div class="space-y-1.5">
                <Label class="text-xs font-bold text-slate-700">Divisi / Departemen *</Label>
                <select
                  v-model="formData.department_id"
                  class="w-full h-10 rounded-xl bg-slate-50 border border-slate-200 px-3 text-xs font-semibold text-slate-800 focus:bg-white focus:outline-none focus:ring-2 focus:ring-primary/20"
                >
                  <option value="" disabled>Pilih Divisi</option>
                  <option v-for="d in masterDataStore.departments" :key="d.id" :value="d.id">
                    {{ d.name }}
                  </option>
                </select>
              </div>

              <!-- Cabang -->
              <div class="space-y-1.5">
                <Label class="text-xs font-bold text-slate-700">Cabang Penempatan *</Label>
                <select
                  v-model="formData.branch_id"
                  class="w-full h-10 rounded-xl bg-slate-50 border border-slate-200 px-3 text-xs font-semibold text-slate-800 focus:bg-white focus:outline-none focus:ring-2 focus:ring-primary/20"
                >
                  <option value="" disabled>Pilih Cabang</option>
                  <option v-for="b in masterDataStore.branches" :key="b.id" :value="b.id">
                    {{ b.name }}
                  </option>
                  <option value="OTHER">Lainnya (Tulis Lokasi Kustom)...</option>
                </select>
              </div>

              <!-- Lokasi Kustom (Jika memilih Lainnya) -->
              <div v-if="formData.branch_id === 'OTHER'" class="space-y-1.5 sm:col-span-2">
                <Label class="text-xs font-bold text-slate-700">Nama / Lokasi Penempatan Kustom *</Label>
                <Input
                  v-model="formData.custom_branch_name"
                  placeholder="Contoh: Gudang Transit Kendal / Proyek Pabrik Baru / Remote"
                  class="h-10 rounded-xl bg-slate-50 border-slate-200 text-xs font-semibold focus:bg-white"
                />
              </div>

              <!-- Posisi Master Data -->
              <div class="space-y-1.5">
                <Label class="text-xs font-bold text-slate-700">Jabatan (Master Data)</Label>
                <select
                  v-model="formData.job_position_id"
                  class="w-full h-10 rounded-xl bg-slate-50 border border-slate-200 px-3 text-xs font-semibold text-slate-800 focus:bg-white focus:outline-none focus:ring-2 focus:ring-primary/20"
                >
                  <option value="">-- Pilih dari Master atau Tulis Kustom --</option>
                  <option v-for="j in masterDataStore.jobPositions" :key="j.id" :value="j.id">
                    {{ j.title }}
                  </option>
                </select>
              </div>

              <!-- Posisi Kustom (Jika tidak ada di master) -->
              <div class="space-y-1.5">
                <Label class="text-xs font-bold text-slate-700">Judul Posisi Kustom (Opsional)</Label>
                <Input
                  v-model="formData.custom_position_title"
                  placeholder="Contoh: Senior UI/UX Designer"
                  class="h-10 rounded-xl bg-slate-50 border-slate-200 text-xs font-semibold focus:bg-white"
                />
              </div>

              <!-- Alasan Kebutuhan -->
              <div class="space-y-1.5">
                <Label class="text-xs font-bold text-slate-700">Alasan Kebutuhan *</Label>
                <select
                  v-model="formData.request_reason"
                  class="w-full h-10 rounded-xl bg-slate-50 border border-slate-200 px-3 text-xs font-semibold text-slate-800 focus:bg-white focus:outline-none focus:ring-2 focus:ring-primary/20"
                >
                  <option value="NEW_HEADCOUNT">Penambahan Tim Baru (New Headcount)</option>
                  <option value="REPLACEMENT">Penggantian Karyawan Keluar (Replacement)</option>
                  <option value="PROJECT_BASED">Kebutuhan Proyek Tertentu</option>
                  <option value="WORKLOAD_INCREASE">Peningkatan Beban Kerja Operasional</option>
                  <option value="OTHER">Alasan Lainnya</option>
                </select>
              </div>

              <!-- Tipe Pekerjaan -->
              <div class="space-y-1.5">
                <Label class="text-xs font-bold text-slate-700">Tipe Hubungan Kerja *</Label>
                <select
                  v-model="formData.employment_type"
                  class="w-full h-10 rounded-xl bg-slate-50 border border-slate-200 px-3 text-xs font-semibold text-slate-800 focus:bg-white focus:outline-none focus:ring-2 focus:ring-primary/20"
                >
                  <option value="FULL_TIME">Karyawan Tetap (PKWTT)</option>
                  <option value="CONTRACT">Karyawan Kontrak (PKWT)</option>
                  <option value="DAILY">Pekerja Harian Lepas</option>
                  <option value="INTERNSHIP">Magang / Internship</option>
                </select>
              </div>

              <!-- Nama Karyawan Diganti (Jika Replacement) -->
              <div v-if="formData.request_reason === 'REPLACEMENT'" class="space-y-1.5 sm:col-span-2">
                <Label class="text-xs font-bold text-slate-700">Nama Karyawan yang Digantikan</Label>
                <Input
                  v-model="formData.replacement_for_name"
                  placeholder="Tuliskan nama karyawan yang resign/mutasi..."
                  class="h-10 rounded-xl bg-slate-50 border-slate-200 text-xs font-semibold focus:bg-white"
                />
              </div>
            </div>
          </div>

          <!-- Section 2: Jumlah & Deadline Target -->
          <div class="space-y-4 pt-4 border-t border-slate-100">
            <div class="flex items-center gap-2 text-xs font-bold text-slate-800 uppercase tracking-wider">
              <Clock class="w-4 h-4 text-primary" />
              <span>2. Jumlah Kebutuhan & Batas Waktu</span>
            </div>

            <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
              <!-- Jumlah Kebutuhan -->
              <div class="space-y-1.5">
                <Label class="text-xs font-bold text-slate-700">Total Kebutuhan Karyawan (Orang) *</Label>
                <Input
                  v-model.number="formData.headcount_needed"
                  type="number"
                  min="1"
                  max="100"
                  class="h-10 rounded-xl bg-slate-50 border-slate-200 text-xs font-bold focus:bg-white"
                />
              </div>

              <!-- Target Date Ready -->
              <div class="space-y-1.5">
                <Label class="text-xs font-bold text-slate-700">Deadline Karyawan Harus Ready (Mulai Kerja) *</Label>
                <Input
                  v-model="formData.target_date"
                  type="date"
                  class="h-10 rounded-xl bg-slate-50 border-slate-200 text-xs font-semibold focus:bg-white"
                />
              </div>
            </div>
          </div>

          <!-- Section 3: Spesifikasi & Kualifikasi Karyawan -->
          <div class="space-y-4 pt-4 border-t border-slate-100">
            <div class="flex items-center gap-2 text-xs font-bold text-slate-800 uppercase tracking-wider">
              <GraduationCap class="w-4 h-4 text-primary" />
              <span>3. Spesifikasi & Kualifikasi Kandidat</span>
            </div>

            <div class="grid grid-cols-1 sm:grid-cols-3 gap-4">
              <!-- Pendidikan Min -->
              <div class="space-y-1.5">
                <Label class="text-xs font-bold text-slate-700">Pendidikan Minimal</Label>
                <select
                  v-model="formData.min_education"
                  class="w-full h-10 rounded-xl bg-slate-50 border border-slate-200 px-3 text-xs font-semibold text-slate-800 focus:bg-white"
                >
                  <option value="SMP">SMP Sederajat</option>
                  <option value="SMA/SMK">SMA / SMK Sederajat</option>
                  <option value="D3">Diploma (D3)</option>
                  <option value="S1">Sarjana (S1 / D4)</option>
                  <option value="S2">Magister (S2)</option>
                  <option value="TIDAK_DIBATASI">Tidak Dibatasi</option>
                </select>
              </div>

              <!-- Gender -->
              <div class="space-y-1.5">
                <Label class="text-xs font-bold text-slate-700">Jenis Kelamin</Label>
                <select
                  v-model="formData.gender_preference"
                  class="w-full h-10 rounded-xl bg-slate-50 border border-slate-200 px-3 text-xs font-semibold text-slate-800 focus:bg-white"
                >
                  <option value="ANY">Laki-laki / Perempuan (Bebas)</option>
                  <option value="MALE">Laki-laki</option>
                  <option value="FEMALE">Perempuan</option>
                </select>
              </div>

              <!-- Pengalaman Kerja -->
              <div class="space-y-1.5">
                <Label class="text-xs font-bold text-slate-700">Pengalaman Kerja Min</Label>
                <select
                  v-model.number="formData.min_experience_years"
                  class="w-full h-10 rounded-xl bg-slate-50 border border-slate-200 px-3 text-xs font-semibold text-slate-800 focus:bg-white"
                >
                  <option :value="0">Fresh Graduate / Bebas</option>
                  <option :value="1">Minimal 1 Tahun</option>
                  <option :value="2">Minimal 2 Tahun</option>
                  <option :value="3">Minimal 3 Tahun</option>
                  <option :value="5">Minimal 5+ Tahun</option>
                </select>
              </div>
            </div>

            <!-- Rentang Usia -->
            <div class="grid grid-cols-2 gap-4">
              <div class="space-y-1.5">
                <Label class="text-xs font-bold text-slate-700">Usia Minimal (Tahun)</Label>
                <Input
                  v-model.number="formData.min_age"
                  type="number"
                  min="17"
                  max="60"
                  class="h-10 rounded-xl bg-slate-50 border-slate-200 text-xs font-semibold focus:bg-white"
                />
              </div>
              <div class="space-y-1.5">
                <Label class="text-xs font-bold text-slate-700">Usia Maksimal (Tahun)</Label>
                <Input
                  v-model.number="formData.max_age"
                  type="number"
                  min="18"
                  max="65"
                  class="h-10 rounded-xl bg-slate-50 border-slate-200 text-xs font-semibold focus:bg-white"
                />
              </div>
            </div>

            <!-- Keahlian Khusus -->
            <div class="space-y-1.5">
              <Label class="text-xs font-bold text-slate-700">Keahlian Khusus / Kualifikasi Teknis *</Label>
              <textarea
                v-model="formData.required_skills"
                rows="3"
                placeholder="Contoh: Menguasai Excel tingkat lanjut, memiliki SIM B1, mampu operasikan mesin..."
                class="w-full rounded-xl bg-slate-50 border border-slate-200 p-3 text-xs font-medium text-slate-800 focus:bg-white focus:outline-none focus:ring-2 focus:ring-primary/20"
              ></textarea>
            </div>

            <!-- Deskripsi Pekerjaan -->
            <div class="space-y-1.5">
              <Label class="text-xs font-bold text-slate-700">Deskripsi Tugas / Tanggung Jawab Utama</Label>
              <textarea
                v-model="formData.job_description"
                rows="3"
                placeholder="Rincian garis besar pekerjaan yang akan dilakukan kandidat..."
                class="w-full rounded-xl bg-slate-50 border border-slate-200 p-3 text-xs font-medium text-slate-800 focus:bg-white focus:outline-none focus:ring-2 focus:ring-primary/20"
              ></textarea>
            </div>

            <!-- Catatan Tambahan -->
            <div class="space-y-1.5">
              <Label class="text-xs font-bold text-slate-700">Catatan Tambahan untuk Tim HR / Direksi</Label>
              <Input
                v-model="formData.notes"
                placeholder="Catatan tambahan misal: prioritas mendesak, jam kerja khusus, dll."
                class="h-10 rounded-xl bg-slate-50 border-slate-200 text-xs font-semibold focus:bg-white"
              />
            </div>
          </div>
        </div>

        <DialogFooter class="pt-4 border-t border-slate-100 flex items-center justify-end gap-2">
          <Button
            @click="isFormModalOpen = false"
            variant="outline"
            class="rounded-xl h-10 px-4 text-xs font-bold"
          >
            Batal
          </Button>
          <Button
            @click="handleSubmitForm"
            :disabled="isSubmitting"
            class="rounded-xl h-10 px-6 text-xs font-bold bg-primary hover:bg-primary/90 text-white shadow-md shadow-primary/20"
          >
            <span v-if="isSubmitting">Menyimpan...</span>
            <span v-else>{{ isEditMode ? 'Simpan Perubahan' : 'Kirim Pengajuan' }}</span>
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>

    <!-- MODAL: Detail Lengkap Permintaan Karyawan & Laporan Karyawan Masuk -->
    <Dialog v-model:open="isDetailModalOpen">
      <DialogContent class="max-w-4xl max-h-[90vh] overflow-y-auto rounded-3xl p-6">
        <DialogHeader class="pb-4 border-b border-slate-100">
          <div class="flex items-center justify-between">
            <div class="flex items-center gap-3">
              <div class="p-2.5 rounded-2xl bg-primary/10 text-primary">
                <FileText class="w-5 h-5" />
              </div>
              <div>
                <DialogTitle class="text-lg font-black text-slate-900">
                  Detail Permintaan Tenaga Kerja & Laporan
                </DialogTitle>
                <p class="text-xs text-slate-500 font-mono mt-0.5">
                  {{ selectedRequest?.request_number }} • Diajukan {{ formatDate(selectedRequest?.created_at) }}
                </p>
              </div>
            </div>
            <span
              v-if="selectedRequest"
              class="px-3 py-1 rounded-full text-xs font-black border uppercase"
              :class="getStatusBadge(selectedRequest.status).class"
            >
              {{ getStatusBadge(selectedRequest.status).label }}
            </span>
          </div>
        </DialogHeader>

        <div v-if="selectedRequest" class="space-y-6 py-4">
          <!-- Overview Card -->
          <div class="bg-slate-50/90 rounded-2xl p-4 border border-slate-200/80 grid grid-cols-2 sm:grid-cols-4 gap-4">
            <div>
              <span class="text-[10px] font-extrabold uppercase text-slate-400">Posisi Dibutuhkan</span>
              <p class="text-sm font-black text-slate-900 mt-0.5">{{ getPositionTitle(selectedRequest) }}</p>
            </div>
            <div>
              <span class="text-[10px] font-extrabold uppercase text-slate-400">Divisi & Cabang</span>
              <p class="text-xs font-bold text-slate-800 mt-0.5">
                {{ selectedRequest.department?.name || 'Umum' }} ({{ getBranchName(selectedRequest) }})
              </p>
            </div>
            <div>
              <span class="text-[10px] font-extrabold uppercase text-slate-400">Total Kebutuhan</span>
              <p class="text-xs font-black text-emerald-700 mt-0.5">
                {{ selectedRequest.headcount_fulfilled }} / {{ selectedRequest.headcount_needed }} Orang
              </p>
            </div>
            <div>
              <span class="text-[10px] font-extrabold uppercase text-slate-400">Target Ready</span>
              <p class="text-xs font-bold text-rose-700 mt-0.5 flex items-center gap-1">
                <Calendar class="w-3.5 h-3.5" />
                {{ formatDate(selectedRequest.target_date) }}
              </p>
            </div>
          </div>

          <!-- Section: Identitas Pemohon (Jelas) -->
          <div class="bg-white rounded-2xl p-4 border border-slate-200/80 space-y-2">
            <h4 class="text-xs font-extrabold text-slate-900 uppercase tracking-wider flex items-center gap-1.5">
              <Users class="w-4 h-4 text-primary" />
              Identitas Pemohon (Requester)
            </h4>
            <div class="grid grid-cols-1 sm:grid-cols-3 gap-3 text-xs pt-1">
              <div>
                <span class="text-[10px] text-slate-400 font-bold block">Nama Pemohon:</span>
                <span class="font-extrabold text-slate-900 text-sm">{{ selectedRequest.requester_name }}</span>
              </div>
              <div>
                <span class="text-[10px] text-slate-400 font-bold block">Jabatan Pemohon:</span>
                <span class="font-bold text-primary">
                  {{ selectedRequest.requester_job_title || selectedRequest.requester?.job_position?.title || 'Kepala Divisi' }}
                </span>
              </div>
              <div>
                <span class="text-[10px] text-slate-400 font-bold block">Divisi Penempatan:</span>
                <span class="font-semibold text-slate-700">{{ selectedRequest.department?.name || '-' }}</span>
              </div>
            </div>
          </div>

          <!-- Section: LAPORAN KARYAWAN BARU YANG DITERIMA (HIRED CANDIDATES & CV) -->
          <div class="bg-white rounded-2xl p-5 border-2 border-primary/20 bg-primary/[0.01] shadow-xs space-y-3">
            <div class="flex items-center justify-between">
              <div class="flex items-center gap-2">
                <div class="p-1.5 bg-primary/10 rounded-xl text-primary">
                  <FileCheck class="w-4 h-4" />
                </div>
                <div>
                  <h4 class="text-xs font-extrabold text-slate-900 uppercase tracking-wider">
                    Laporan Karyawan Baru yang Diterima (Hired Candidates & CV)
                  </h4>
                  <p class="text-[11px] text-slate-500 font-medium">
                    Daftar karyawan yang telah bergabung untuk memenuhi pengajuan ini beserta dokumen CV & Foto.
                  </p>
                </div>
              </div>

              <Button
                v-if="isHRManager && (selectedRequest.status === 'APPROVED' || selectedRequest.status === 'IN_PROGRESS' || selectedRequest.status === 'FULFILLED')"
                @click="openProgressModal(selectedRequest)"
                size="sm"
                variant="outline"
                class="h-8 rounded-xl text-xs font-bold border-primary text-primary hover:bg-primary/5 flex items-center gap-1.5"
              >
                <Plus class="w-3.5 h-3.5" />
                <span>+ Input / Tambah Karyawan</span>
              </Button>
            </div>

            <!-- List Candidates -->
            <div v-if="selectedRequest.candidates && selectedRequest.candidates.length > 0" class="grid grid-cols-1 sm:grid-cols-2 gap-3 pt-2">
              <div
                v-for="(cand, idx) in selectedRequest.candidates"
                :key="cand.id"
                class="bg-slate-50/80 p-4 rounded-2xl border border-slate-200/80 hover:bg-white hover:border-primary/40 transition-all flex flex-col justify-between space-y-3"
              >
                <div>
                  <div class="flex items-start justify-between gap-2">
                    <div class="flex items-center gap-2">
                      <div class="w-7 h-7 rounded-xl bg-primary text-white font-extrabold text-xs flex items-center justify-center">
                        {{ idx + 1 }}
                      </div>
                      <div>
                        <h5 class="font-extrabold text-slate-900 text-sm leading-tight">{{ cand.candidate_name }}</h5>
                        <span class="text-[10px] text-slate-400 font-medium flex items-center gap-1 mt-0.5">
                          <Calendar class="w-3 h-3 text-emerald-600" />
                          Mulai Masuk: <strong>{{ formatDate(cand.join_date) }}</strong>
                        </span>
                      </div>
                    </div>

                    <!-- Delete Candidate (For HR only) -->
                    <button
                      v-if="isHRManager"
                      @click="handleDeleteCandidate(cand.id)"
                      class="text-slate-300 hover:text-rose-600 p-1 rounded-lg transition-colors"
                      title="Hapus data karyawan ini"
                    >
                      <Trash2 class="w-3.5 h-3.5" />
                    </button>
                  </div>

                  <div v-if="cand.candidate_phone" class="mt-2 text-[11px] text-slate-600 flex items-center gap-1.5">
                    <Phone class="w-3 h-3 text-slate-400" />
                    <span>{{ cand.candidate_phone }}</span>
                  </div>

                  <div v-if="cand.notes" class="mt-1.5 text-[11px] text-slate-500 italic bg-white p-2 rounded-xl border border-slate-100">
                    "{{ cand.notes }}"
                  </div>
                </div>

                <!-- Document Attachment Button -->
                <div class="pt-2 border-t border-slate-200/60 flex items-center justify-between">
                  <div v-if="cand.document_url" class="flex items-center gap-2">
                    <span
                      class="px-2 py-0.5 rounded-md text-[9px] font-black uppercase"
                      :class="cand.document_type === 'PDF' ? 'bg-red-50 text-red-700 border border-red-200' : 'bg-blue-50 text-blue-700 border border-blue-200'"
                    >
                      {{ cand.document_type || 'FILE' }}
                    </span>
                    <span class="text-xs font-semibold text-slate-700 truncate max-w-[140px]" :title="cand.document_name">
                      {{ cand.document_name || 'Dokumen CV' }}
                    </span>
                  </div>
                  <span v-else class="text-[10px] text-slate-400 italic">Tanpa lampiran file</span>

                  <a
                    v-if="cand.document_url"
                    :href="cand.document_url"
                    target="_blank"
                    class="px-2.5 py-1 rounded-xl bg-primary text-white text-[11px] font-bold hover:bg-primary/90 flex items-center gap-1 shadow-xs transition-all"
                  >
                    <Eye class="w-3 h-3" />
                    <span>Lihat CV / Dokumen</span>
                  </a>
                </div>
              </div>
            </div>

            <div v-else class="text-center py-6 bg-slate-50 rounded-2xl border border-dashed border-slate-200 text-slate-400 text-xs">
              <Users class="w-6 h-6 mx-auto mb-1.5 text-slate-300" />
              <p class="font-semibold text-slate-600">Belum ada data final karyawan yang dicatat</p>
              <p class="text-[11px] text-slate-400 mt-0.5">Tim HR dapat menginput data nama karyawan dan file CV saat proses rekrutmen selesai.</p>
            </div>
          </div>

          <!-- Section: Dual Approval Status Cards -->
          <div class="space-y-3">
            <h4 class="text-xs font-extrabold text-slate-900 uppercase tracking-wider flex items-center gap-1.5">
              <ShieldCheck class="w-4 h-4 text-primary" />
              Status Persetujuan Berjenjang (Dual-Tier Approval)
            </h4>

            <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
              <!-- Card 1: HR Manager Approval -->
              <div
                class="rounded-2xl p-4 border text-xs flex flex-col justify-between space-y-3"
                :class="{
                  'bg-emerald-50/50 border-emerald-200': selectedRequest.hr_status === 'APPROVED',
                  'bg-rose-50/50 border-rose-200': selectedRequest.hr_status === 'REJECTED',
                  'bg-amber-50/40 border-amber-200': selectedRequest.hr_status === 'PENDING' || !selectedRequest.hr_status,
                }"
              >
                <div class="space-y-1">
                  <div class="flex items-center justify-between">
                    <span class="font-extrabold text-slate-800 flex items-center gap-1.5">
                      <Award class="w-4 h-4 text-amber-600" />
                      1. Persetujuan HR Manager
                    </span>
                    <span
                      class="px-2 py-0.5 rounded-full font-bold text-[10px] uppercase"
                      :class="{
                        'bg-emerald-100 text-emerald-800': selectedRequest.hr_status === 'APPROVED',
                        'bg-rose-100 text-rose-800': selectedRequest.hr_status === 'REJECTED',
                        'bg-amber-100 text-amber-800': selectedRequest.hr_status === 'PENDING' || !selectedRequest.hr_status,
                      }"
                    >
                      {{ selectedRequest.hr_status || 'PENDING' }}
                    </span>
                  </div>

                  <p v-if="selectedRequest.hr_approver" class="text-slate-600 mt-1">
                    Disetujui oleh: <strong class="text-slate-900">{{ selectedRequest.hr_approver.first_name }}</strong>
                    <span class="text-[10px] text-slate-400 block">{{ formatDate(selectedRequest.hr_approval_date) }}</span>
                  </p>
                  <p v-else class="text-slate-400 italic">Menunggu verifikasi dan persetujuan HR Manager.</p>

                  <div v-if="selectedRequest.hr_approval_notes" class="mt-2 pt-2 border-t border-slate-200/50">
                    <span class="text-[10px] text-slate-400 font-bold block">Catatan HR:</span>
                    <p class="text-slate-700 italic">"{{ selectedRequest.hr_approval_notes }}"</p>
                  </div>
                </div>
              </div>

              <!-- Card 2: Director Approval -->
              <div
                class="rounded-2xl p-4 border text-xs flex flex-col justify-between space-y-3"
                :class="{
                  'bg-emerald-50/50 border-emerald-200': selectedRequest.director_status === 'APPROVED',
                  'bg-rose-50/50 border-rose-200': selectedRequest.director_status === 'REJECTED',
                  'bg-indigo-50/40 border-indigo-200': selectedRequest.director_status === 'PENDING' || !selectedRequest.director_status,
                }"
              >
                <div class="space-y-1">
                  <div class="flex items-center justify-between">
                    <span class="font-extrabold text-slate-800 flex items-center gap-1.5">
                      <ShieldCheck class="w-4 h-4 text-indigo-600" />
                      2. Persetujuan Direktur
                    </span>
                    <span
                      class="px-2 py-0.5 rounded-full font-bold text-[10px] uppercase"
                      :class="{
                        'bg-emerald-100 text-emerald-800': selectedRequest.director_status === 'APPROVED',
                        'bg-rose-100 text-rose-800': selectedRequest.director_status === 'REJECTED',
                        'bg-indigo-100 text-indigo-800': selectedRequest.director_status === 'PENDING' || !selectedRequest.director_status,
                      }"
                    >
                      {{ selectedRequest.director_status || 'PENDING' }}
                    </span>
                  </div>

                  <p v-if="selectedRequest.director_approver" class="text-slate-600 mt-1">
                    Disetujui oleh: <strong class="text-slate-900">{{ selectedRequest.director_approver.first_name }}</strong>
                    <span class="text-[10px] text-slate-400 block">{{ formatDate(selectedRequest.director_approval_date) }}</span>
                  </p>
                  <p v-else class="text-slate-400 italic">Menunggu otorisasi final dari salah satu Direktur.</p>

                  <div v-if="selectedRequest.director_approval_notes" class="mt-2 pt-2 border-t border-slate-200/50">
                    <span class="text-[10px] text-slate-400 font-bold block">Catatan Direktur:</span>
                    <p class="text-slate-700 italic">"{{ selectedRequest.director_approval_notes }}"</p>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <!-- Section: Kualifikasi & Spesifikasi -->
          <div class="space-y-3">
            <h4 class="text-xs font-extrabold text-slate-900 uppercase tracking-wider flex items-center gap-1.5">
              <GraduationCap class="w-4 h-4 text-primary" />
              Kualifikasi & Spesifikasi Karyawan
            </h4>
            <div class="bg-white rounded-2xl p-4 border border-slate-200/80 grid grid-cols-2 sm:grid-cols-4 gap-4 text-xs">
              <div>
                <span class="text-[10px] text-slate-400 font-bold block">Pendidikan Min</span>
                <span class="font-bold text-slate-800">{{ selectedRequest.min_education || 'Tidak dibatasi' }}</span>
              </div>
              <div>
                <span class="text-[10px] text-slate-400 font-bold block">Jenis Kelamin</span>
                <span class="font-bold text-slate-800">
                  {{ selectedRequest.gender_preference === 'MALE' ? 'Laki-laki' : selectedRequest.gender_preference === 'FEMALE' ? 'Perempuan' : 'Bebas' }}
                </span>
              </div>
              <div>
                <span class="text-[10px] text-slate-400 font-bold block">Rentang Usia</span>
                <span class="font-bold text-slate-800">{{ selectedRequest.min_age }} - {{ selectedRequest.max_age }} Tahun</span>
              </div>
              <div>
                <span class="text-[10px] text-slate-400 font-bold block">Pengalaman Min</span>
                <span class="font-bold text-slate-800">{{ selectedRequest.min_experience_years }} Tahun</span>
              </div>
            </div>

            <!-- Keahlian & Deskripsi -->
            <div class="space-y-2">
              <div class="bg-slate-50 p-3.5 rounded-2xl border border-slate-200/60">
                <span class="text-[10px] font-extrabold text-slate-400 uppercase tracking-wider block mb-1">
                  Keahlian & Kemampuan Khusus:
                </span>
                <p class="text-xs text-slate-700 whitespace-pre-line font-medium leading-relaxed">
                  {{ selectedRequest.required_skills || 'Tidak ada spesifikasi khusus tertulis.' }}
                </p>
              </div>

              <div v-if="selectedRequest.job_description" class="bg-slate-50 p-3.5 rounded-2xl border border-slate-200/60">
                <span class="text-[10px] font-extrabold text-slate-400 uppercase tracking-wider block mb-1">
                  Deskripsi Tanggung Jawab:
                </span>
                <p class="text-xs text-slate-700 whitespace-pre-line font-medium leading-relaxed">
                  {{ selectedRequest.job_description }}
                </p>
              </div>
            </div>
          </div>
        </div>

        <DialogFooter class="pt-4 border-t border-slate-100 flex items-center justify-end">
          <Button
            @click="isDetailModalOpen = false"
            class="rounded-xl h-10 px-6 text-xs font-bold bg-slate-900 text-white"
          >
            Tutup
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>

    <!-- MODAL: Setujui (HR Manager) -->
    <Dialog v-model:open="isHRApproveModalOpen">
      <DialogContent class="max-w-md rounded-3xl p-6">
        <DialogHeader class="pb-2">
          <div class="flex items-center gap-3">
            <div class="p-2.5 rounded-2xl bg-amber-100 text-amber-800">
              <Award class="w-5 h-5" />
            </div>
            <div>
              <DialogTitle class="text-base font-black text-slate-900">
                Persetujuan HR Manager (Tahap 1)
              </DialogTitle>
              <DialogDescription class="text-xs text-slate-500 mt-0.5">
                Konfirmasi kelayakan kualifikasi posisi dan ketersediaan alokasi tim sebelum diteruskan ke Direktur.
              </DialogDescription>
            </div>
          </div>
        </DialogHeader>

        <div class="space-y-3 py-3">
          <div class="space-y-1.5">
            <Label class="text-xs font-bold text-slate-700">Catatan HR Manager (Opsional)</Label>
            <textarea
              v-model="approvalNotes"
              rows="3"
              placeholder="Tambahkan catatan verifikasi atau rekomendasi untuk Direktur..."
              class="w-full rounded-xl bg-slate-50 border border-slate-200 p-3 text-xs font-medium text-slate-800 focus:bg-white focus:outline-none focus:ring-2 focus:ring-amber-500/20"
            ></textarea>
          </div>
        </div>

        <DialogFooter class="pt-3 border-t border-slate-100 flex items-center justify-end gap-2">
          <Button
            @click="isHRApproveModalOpen = false"
            variant="outline"
            class="rounded-xl h-9 px-4 text-xs font-bold"
          >
            Batal
          </Button>
          <Button
            @click="handleHRApproveConfirm"
            :disabled="isSubmitting"
            class="rounded-xl h-9 px-5 text-xs font-bold bg-amber-600 hover:bg-amber-700 text-white"
          >
            <span v-if="isSubmitting">Memproses...</span>
            <span v-else>Setujui sebagai HR</span>
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>

    <!-- MODAL: Setujui (Direktur) -->
    <Dialog v-model:open="isDirectorApproveModalOpen">
      <DialogContent class="max-w-md rounded-3xl p-6">
        <DialogHeader class="pb-2">
          <div class="flex items-center gap-3">
            <div class="p-2.5 rounded-2xl bg-indigo-100 text-indigo-800">
              <ShieldCheck class="w-5 h-5" />
            </div>
            <div>
              <DialogTitle class="text-base font-black text-slate-900">
                Otorisasi Persetujuan Direktur (Tahap 2)
              </DialogTitle>
              <DialogDescription class="text-xs text-slate-500 mt-0.5">
                Memberikan otorisasi final Direksi untuk pembukaan lowongan dan penambahan tenaga kerja.
              </DialogDescription>
            </div>
          </div>
        </DialogHeader>

        <div class="space-y-3 py-3">
          <div class="space-y-1.5">
            <Label class="text-xs font-bold text-slate-700">Catatan Direktur (Opsional)</Label>
            <textarea
              v-model="approvalNotes"
              rows="3"
              placeholder="Tambahkan arahan atau catatan otorisasi Direksi..."
              class="w-full rounded-xl bg-slate-50 border border-slate-200 p-3 text-xs font-medium text-slate-800 focus:bg-white focus:outline-none focus:ring-2 focus:ring-indigo-500/20"
            ></textarea>
          </div>
        </div>

        <DialogFooter class="pt-3 border-t border-slate-100 flex items-center justify-end gap-2">
          <Button
            @click="isDirectorApproveModalOpen = false"
            variant="outline"
            class="rounded-xl h-9 px-4 text-xs font-bold"
          >
            Batal
          </Button>
          <Button
            @click="handleDirectorApproveConfirm"
            :disabled="isSubmitting"
            class="rounded-xl h-9 px-5 text-xs font-bold bg-indigo-600 hover:bg-indigo-700 text-white"
          >
            <span v-if="isSubmitting">Memproses...</span>
            <span v-else>Setujui sebagai Direktur</span>
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>

    <!-- MODAL: Tolak Pengajuan -->
    <Dialog v-model:open="isRejectModalOpen">
      <DialogContent class="max-w-md rounded-3xl p-6">
        <DialogHeader class="pb-2">
          <div class="flex items-center gap-3">
            <div class="p-2.5 rounded-2xl bg-rose-100 text-rose-700">
              <XCircle class="w-5 h-5" />
            </div>
            <div>
              <DialogTitle class="text-base font-black text-slate-900">
                Tolak Permintaan Karyawan
              </DialogTitle>
              <DialogDescription class="text-xs text-slate-500 mt-0.5">
                Tolak sebagai: <strong>{{ rejectRole === 'DIRECTOR' ? 'Direktur' : 'HR Manager' }}</strong>. Berikan alasan tertulis untuk Kepala Divisi pemohon.
              </DialogDescription>
            </div>
          </div>
        </DialogHeader>

        <div class="space-y-3 py-3">
          <div class="space-y-1.5">
            <Label class="text-xs font-bold text-slate-700">Alasan Penolakan *</Label>
            <textarea
              v-model="rejectReason"
              rows="3"
              placeholder="Contoh: Kuota anggaran divisi telah penuh, atau alokasi dialihkan ke kuartal berikutnya..."
              class="w-full rounded-xl bg-slate-50 border border-slate-200 p-3 text-xs font-medium text-slate-800 focus:bg-white focus:outline-none focus:ring-2 focus:ring-rose-500/20"
            ></textarea>
          </div>
        </div>

        <DialogFooter class="pt-3 border-t border-slate-100 flex items-center justify-end gap-2">
          <Button
            @click="isRejectModalOpen = false"
            variant="outline"
            class="rounded-xl h-9 px-4 text-xs font-bold"
          >
            Batal
          </Button>
          <Button
            @click="handleRejectConfirm"
            :disabled="isSubmitting"
            class="rounded-xl h-9 px-5 text-xs font-bold bg-rose-600 hover:bg-rose-700 text-white"
          >
            <span v-if="isSubmitting">Memproses...</span>
            <span v-else>Tolak Pengajuan</span>
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>

    <!-- MODAL: Input Karyawan Baru & CV (Bulk Supported) & Update Progres -->
    <Dialog v-model:open="isProgressModalOpen">
      <DialogContent class="max-w-3xl max-h-[90vh] overflow-y-auto rounded-3xl p-6">
        <DialogHeader class="pb-3 border-b border-slate-100">
          <div class="flex items-center gap-3">
            <div class="p-2.5 rounded-2xl bg-primary/10 text-primary">
              <TrendingUp class="w-5 h-5" />
            </div>
            <div>
              <DialogTitle class="text-base font-black text-slate-900">
                Input Data Karyawan Baru & Dokumen CV (Bulk)
              </DialogTitle>
              <DialogDescription class="text-xs text-slate-500 mt-0.5">
                Kirim data nama, tanggal join, dan unggah CV / Foto kandidat (PDF, JPG, PNG) untuk laporan Kepala Divisi.
              </DialogDescription>
            </div>
          </div>
        </DialogHeader>

        <div v-if="selectedRequest" class="space-y-5 py-4">
          <!-- Position & Headcount Info Banner -->
          <div class="p-3.5 bg-slate-50 rounded-2xl border border-slate-200/80 flex items-center justify-between text-xs">
            <div>
              <span class="text-slate-400 font-bold block">Posisi yang Diajukan:</span>
              <span class="font-extrabold text-slate-900 text-sm">{{ getPositionTitle(selectedRequest) }}</span>
            </div>
            <div class="text-right">
              <span class="text-slate-400 font-bold block">Total Kebutuhan:</span>
              <span class="font-black text-emerald-700 text-sm">
                {{ selectedRequest.headcount_fulfilled }} / {{ selectedRequest.headcount_needed }} Orang Terpenuhi
              </span>
            </div>
          </div>

          <!-- 1. Dropdown Pilihan Status Tahapan Rekrutmen -->
          <div class="p-4 rounded-2xl bg-primary/5 border border-primary/20 space-y-3">
            <div class="grid grid-cols-1 sm:grid-cols-2 gap-3 items-center">
              <div class="space-y-1">
                <Label class="text-xs font-black uppercase tracking-wider text-slate-800">
                  Status Tahapan Rekrutmen *
                </Label>
                <p class="text-[11px] text-slate-500 font-medium">
                  Tentukan apakah masih dalam proses seleksi atau telah selesai terpenuhi.
                </p>
              </div>

              <div>
                <select
                  v-model="progressStatus"
                  class="w-full h-10 rounded-xl bg-white border-2 border-primary/30 px-3 text-xs font-bold text-slate-800 focus:outline-none focus:ring-2 focus:ring-primary/20 shadow-xs"
                >
                  <option value="IN_PROGRESS">Sedang Berjalan (Sourcing / Screening / Interview)</option>
                  <option value="FULFILLED">Selesai Terpenuhi (Karyawan Terpilih Siap Masuk)</option>
                  <option value="CANCELLED">Dibatalkan</option>
                </select>
              </div>
            </div>

            <!-- Catatan Progres / Perkembangan -->
            <div class="space-y-1.5 pt-2 border-t border-primary/10">
              <Label class="text-xs font-bold text-slate-700">
                Catatan Perkembangan / Keterangan Seleksi
              </Label>
              <textarea
                v-model="progressRecruitmentNotes"
                rows="2"
                :placeholder="progressStatus === 'IN_PROGRESS' ? 'Contoh: Sedang memproses 3 kandidat untuk interview user dan tes teknis pada minggu ini...' : 'Contoh: Karyawan telah menandatangani kontrak dan siap bergabung pada tanggal yang ditentukan.'"
                class="w-full rounded-xl bg-white border border-slate-200 p-3 text-xs font-medium text-slate-800 focus:outline-none focus:ring-2 focus:ring-primary/20"
              ></textarea>
            </div>
          </div>

          <!-- 2. KONDISI: JIKA STATUS SEDANG BERJALAN (IN_PROGRESS) -->
          <div v-if="progressStatus === 'IN_PROGRESS'" class="p-4 rounded-2xl bg-amber-50/60 border border-amber-200 text-xs space-y-2">
            <div class="flex items-center gap-2 text-amber-900 font-bold">
              <Clock class="w-4 h-4 text-amber-600" />
              <span>Proses Rekrutmen Sedang Berjalan</span>
            </div>
            <p class="text-slate-600 text-[11px] leading-relaxed">
              Anda sedang dalam tahap pencarian, seleksi berkas, atau interview kandidat. Cukup perbarui catatan perkembangan di atas untuk menginformasikan Kepala Divisi pemohon. 
              <strong>Form pengisian data final karyawan & upload CV akan muncul saat status diubah menjadi "Selesai Terpenuhi".</strong>
            </p>
          </div>

          <!-- 3. KONDISI: JIKA STATUS SELESAI TERPENUHI (FULFILLED) -->
          <div v-else-if="progressStatus === 'FULFILLED'" class="space-y-4">
            <div class="flex items-center justify-between">
              <div class="flex items-center gap-2">
                <Users class="w-4 h-4 text-primary" />
                <span class="text-xs font-black uppercase tracking-wider text-slate-900">
                  Data Karyawan Baru yang Diterima ({{ candidateList.length }} Orang)
                </span>
              </div>
              <Button
                @click="addCandidateRow"
                size="sm"
                variant="outline"
                class="h-8 rounded-xl text-xs font-bold border-primary text-primary hover:bg-primary/5 flex items-center gap-1.5"
              >
                <Plus class="w-3.5 h-3.5" />
                <span>+ Tambah Karyawan Lain</span>
              </Button>
            </div>

            <!-- Candidate Row Items -->
            <div class="space-y-3">
              <div
                v-for="(cand, idx) in candidateList"
                :key="idx"
                class="p-4 rounded-2xl bg-white border-2 border-slate-200/80 shadow-xs space-y-3 relative"
              >
                <div class="flex items-center justify-between">
                  <span class="text-xs font-extrabold text-slate-800 bg-slate-100 px-2.5 py-0.5 rounded-lg">
                    Karyawan #{{ idx + 1 }}
                  </span>
                  <button
                    v-if="candidateList.length > 1"
                    @click="removeCandidateRow(idx)"
                    class="text-slate-400 hover:text-rose-600 text-xs font-bold flex items-center gap-1"
                  >
                    <Trash2 class="w-3.5 h-3.5" />
                    <span>Hapus Baris</span>
                  </button>
                </div>

                <div class="grid grid-cols-1 sm:grid-cols-3 gap-3">
                  <!-- Nama Karyawan -->
                  <div class="space-y-1 sm:col-span-2">
                    <Label class="text-xs font-bold text-slate-700">Nama Lengkap Karyawan Baru *</Label>
                    <Input
                      v-model="cand.candidate_name"
                      placeholder="Nama lengkap kandidat yang diterima..."
                      class="h-9 rounded-xl bg-slate-50 border-slate-200 text-xs font-semibold focus:bg-white"
                    />
                  </div>

                  <!-- Tanggal Join -->
                  <div class="space-y-1">
                    <Label class="text-xs font-bold text-slate-700">Tanggal Mulai Masuk *</Label>
                    <Input
                      v-model="cand.join_date"
                      type="date"
                      class="h-9 rounded-xl bg-slate-50 border-slate-200 text-xs font-semibold focus:bg-white"
                    />
                  </div>

                  <!-- No. HP -->
                  <div class="space-y-1">
                    <Label class="text-xs font-bold text-slate-700">Nomor Telepon / WA</Label>
                    <Input
                      v-model="cand.candidate_phone"
                      placeholder="Contoh: 08123456789"
                      class="h-9 rounded-xl bg-slate-50 border-slate-200 text-xs font-semibold focus:bg-white"
                    />
                  </div>

                  <!-- Catatan Singkat -->
                  <div class="space-y-1 sm:col-span-2">
                    <Label class="text-xs font-bold text-slate-700">Catatan Kandidat (Opsional)</Label>
                    <Input
                      v-model="cand.notes"
                      placeholder="Contoh: Hasil tes teknis memuaskan, background perbankan..."
                      class="h-9 rounded-xl bg-slate-50 border-slate-200 text-xs font-semibold focus:bg-white"
                    />
                  </div>
                </div>

                <!-- File Uploader for CV / Photo -->
                <div class="pt-2 border-t border-slate-100">
                  <Label class="text-xs font-bold text-slate-700 block mb-1.5">
                    Unggah CV / Foto Profil / Dokumen (PDF, JPG, JPEG, PNG)
                  </Label>

                  <!-- If file is uploaded -->
                  <div
                    v-if="cand.document_url"
                    class="flex items-center justify-between p-2.5 rounded-xl bg-emerald-50 border border-emerald-200 text-xs"
                  >
                    <div class="flex items-center gap-2">
                      <FileCheck class="w-4 h-4 text-emerald-600 shrink-0" />
                      <span class="font-bold text-emerald-900 truncate max-w-[280px]">
                        {{ cand.document_name || 'Dokumen CV Terunggah' }}
                      </span>
                    </div>
                    <div class="flex items-center gap-2">
                      <a
                        :href="cand.document_url"
                        target="_blank"
                        class="text-xs font-bold text-emerald-700 hover:underline flex items-center gap-0.5"
                      >
                        <Eye class="w-3 h-3" /> Preview
                      </a>
                      <button
                        @click="removeCandidateFile(idx)"
                        class="text-rose-600 hover:text-rose-800 p-1 font-bold text-xs"
                        title="Hapus file"
                      >
                        <X class="w-3.5 h-3.5" />
                      </button>
                    </div>
                  </div>

                  <!-- File input drop area -->
                  <div v-else class="flex items-center gap-2">
                    <label
                      class="flex-1 flex items-center justify-center gap-2 py-2 px-3 rounded-xl border border-dashed border-slate-300 bg-slate-50 hover:bg-slate-100 cursor-pointer transition-colors text-xs font-semibold text-slate-600"
                    >
                      <UploadCloud class="w-4 h-4 text-slate-400" />
                      <span v-if="cand.isUploading">Mengunggah file...</span>
                      <span v-else>Pilih File CV / Foto (PDF, JPG, PNG)</span>
                      <input
                        type="file"
                        accept=".pdf,.jpg,.jpeg,.png"
                        class="hidden"
                        :disabled="cand.isUploading"
                        @change="(e) => handleFileUpload(e, idx)"
                      />
                    </label>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>

        <DialogFooter class="pt-3 border-t border-slate-100 flex items-center justify-end gap-2">
          <Button
            @click="isProgressModalOpen = false"
            variant="outline"
            class="rounded-xl h-9 px-4 text-xs font-bold"
          >
            Batal
          </Button>
          <Button
            @click="handleProgressSubmit"
            :disabled="isSubmitting"
            class="rounded-xl h-9 px-5 text-xs font-bold bg-primary hover:bg-primary/90 text-white"
          >
            <span v-if="isSubmitting">Menyimpan...</span>
            <span v-else>Simpan Perubahan</span>
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>

    <!-- Confirmation Dialog Hapus -->
    <ConfirmationDialog
      v-model:open="isDeleteConfirmOpen"
      title="Hapus Permintaan Karyawan"
      description="Apakah Anda yakin ingin menghapus pengajuan permintaan karyawan ini? Tindakan ini tidak dapat dibatalkan."
      variant="danger"
      confirmText="Hapus Pengajuan"
      cancelText="Batal"
      :isLoading="isSubmitting"
      @confirm="handleDeleteConfirm"
      @cancel="isDeleteConfirmOpen = false"
    />
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
</style>
