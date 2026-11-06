<script setup lang="ts">
import { ref, computed, onMounted, watch, h } from 'vue'
import {
  Pencil,
  Trash2,
  Megaphone,
  Bell,
  Calendar,
  AlertTriangle,
  Users,
  CheckCircle2,
  Clock,
  Search,
  Check,
  UserCheck,
  Briefcase,
  RefreshCw,
  X,
  BellRing,
  Send,
  Heart,
  MessageSquare,
  Share2,
  Image as ImageIcon,
  Sparkles,
  ExternalLink,
  Upload,
  Video,
  Film,
  Play,
} from 'lucide-vue-next'
import apiClient from '@/api/axios'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Switch } from '@/components/ui/switch'
import { Label } from '@/components/ui/label'
import { toast } from 'vue-sonner'
import { resolvePhotoUrl } from '@/utils/imageUrl'

// ─── STATE MANAGEMENT ───────────────────────────────────────────
const activeTab = ref<'general' | 'direct'>('direct')
const isLoading = ref(true)
const isSubmitting = ref(false)
const announcements = ref<any[]>([])
const allEmployees = ref<any[]>([])

// Modal Create/Edit
const isModalOpen = ref(false)
const isEditMode = ref(false)

// Modal Confirmation Tracking
const isTrackingModalOpen = ref(false)
const isLoadingTracking = ref(false)
const trackingSummary = ref<any>(null)
const trackingFilter = ref<'all' | 'confirmed' | 'pending'>('all')
const trackingSearch = ref('')

// Form State
const currentAnnouncement = ref({
  id: '',
  title: '',
  content: '',
  category: 'urgent_call',
  is_active: true,
  media_type: 'none',
  image_url: '',
  media_urls_text: '',
  external_url: '',
  target_type: 'individual', // 'broadcast' | 'role' | 'individual'
  target_roles: [] as string[],
  target_user_ids: [] as string[],
  is_popup: true,
  popup_expires_at: '',
  track_downloads: false,
  track_shares: false,
})

// Employee & Role search in modal
const employeeSearchQuery = ref('')
const roleSearchQuery = ref('')
const roleTargetSubTab = ref<'positions' | 'roles'>('positions')

// Database Master Data
const jobPositions = ref<any[]>([])
const accessRoles = ref<any[]>([])

const categories = [
  { value: 'social_media', label: 'Postingan Medsos & Video', icon: Share2, color: 'text-purple-500' },
  { value: 'promo', label: 'Promosi & Brosur Katalog', icon: Sparkles, color: 'text-emerald-500' },
  { value: 'urgent_call', label: 'Panggilan Menghadap (Urgent)', icon: AlertTriangle, color: 'text-red-500' },
  { value: 'event', label: 'Undangan Meeting & Acara', icon: Calendar, color: 'text-blue-500' },
  { value: 'policy', label: 'Kebijakan Perusahaan', icon: Bell, color: 'text-indigo-500' },
  { value: 'info', label: 'Informasi Umum', icon: Megaphone, color: 'text-amber-500' },
]

// ─── COMPUTED: DYNAMIC TARGET ROLES & POSITIONS (FROM DB) ────────
const dynamicJobPositions = computed(() => {
  const counts: Record<string, number> = {}
  allEmployees.value.forEach((emp) => {
    const title = emp.job_position?.title
    if (title) {
      counts[title] = (counts[title] || 0) + 1
    }
  })

  const list = jobPositions.value.map((pos) => ({
    value: pos.title,
    label: pos.title,
    count: counts[pos.title] || 0,
    department: pos.department?.name,
  }))

  const existingTitles = new Set(list.map((p) => p.value.toLowerCase()))
  Object.keys(counts).forEach((title) => {
    if (!existingTitles.has(title.toLowerCase())) {
      list.push({
        value: title,
        label: title,
        count: counts[title] || 0,
        department: undefined,
      })
    }
  })

  return list.sort((a, b) => b.count - a.count || a.label.localeCompare(b.label))
})

const dynamicAccessRoles = computed(() => {
  const counts: Record<string, number> = {}
  allEmployees.value.forEach((emp) => {
    const r = emp.user?.role || emp.role
    if (r) {
      counts[r] = (counts[r] || 0) + 1
    }
  })

  const list = accessRoles.value.map((r) => {
    const roleKey = (r.name || '').toLowerCase().replace(/\s+/g, '_')
    return {
      value: roleKey || r.name,
      label: r.name,
      count: counts[roleKey] || counts[r.name] || 0,
    }
  })

  const standardRoles = [
    { value: 'supervisor_sales', label: 'Supervisor Sales' },
    { value: 'employee', label: 'Sales / Staff Karyawan' },
    { value: 'admin_gudang', label: 'Admin Gudang' },
    { value: 'supervisor_gudang', label: 'Supervisor Gudang' },
    { value: 'admin_pabrik', label: 'Admin Pabrik' },
    { value: 'operator', label: 'Operator Produksi' },
    { value: 'admin_kasir', label: 'Admin Kasir' },
    { value: 'cashier', label: 'Kasir Toko' },
    { value: 'cashier_pusat', label: 'Kasir Pusat' },
    { value: 'customer_support', label: 'Customer Support' },
    { value: 'admin_nota', label: 'Admin Nota' },
    { value: 'admin_designer', label: 'Admin Designer' },
    { value: 'validator_desain', label: 'Validator Desain' },
    { value: 'admin_purchasing', label: 'Admin Purchasing' },
    { value: 'admin_marketplace', label: 'Admin Marketplace' },
    { value: 'admin_operasional', label: 'Admin Operasional' },
    { value: 'auditor', label: 'Auditor' },
    { value: 'manager', label: 'Manager' },
  ]

  const existingKeys = new Set(list.map((r) => r.value.toLowerCase()))
  standardRoles.forEach((sr) => {
    if (!existingKeys.has(sr.value.toLowerCase())) {
      list.push({
        value: sr.value,
        label: sr.label,
        count: counts[sr.value] || 0,
      })
    }
  })

  return list.sort((a, b) => b.count - a.count || a.label.localeCompare(b.label))
})

const filteredJobPositions = computed(() => {
  if (!roleSearchQuery.value.trim()) return dynamicJobPositions.value
  const q = roleSearchQuery.value.toLowerCase()
  return dynamicJobPositions.value.filter((p) => p.label.toLowerCase().includes(q))
})

const filteredAccessRoles = computed(() => {
  if (!roleSearchQuery.value.trim()) return dynamicAccessRoles.value
  const q = roleSearchQuery.value.toLowerCase()
  return dynamicAccessRoles.value.filter((r) => r.label.toLowerCase().includes(q))
})

// ─── COMPUTED ───────────────────────────────────────────────────
const generalAnnouncements = computed(() => {
  return announcements.value.filter(
    (a) => a.category !== 'urgent_call' && a.category !== 'event'
  )
})

const directAnnouncements = computed(() => {
  return announcements.value.filter(
    (a) => a.is_popup || a.category === 'urgent_call' || a.category === 'event' || a.target_type !== 'broadcast'
  )
})

const filteredEmployees = computed(() => {
  if (!employeeSearchQuery.value.trim()) return allEmployees.value.slice(0, 30)
  const q = employeeSearchQuery.value.toLowerCase()
  return allEmployees.value.filter((emp) => {
    const fullName = `${emp.first_name || ''} ${emp.last_name || ''}`.toLowerCase()
    const nik = (emp.employee_id_number || '').toLowerCase()
    const dept = (emp.department?.name || '').toLowerCase()
    const job = (emp.job_position?.title || '').toLowerCase()
    return fullName.includes(q) || nik.includes(q) || dept.includes(q) || job.includes(q)
  })
})

const filteredRecipients = computed(() => {
  if (!trackingSummary.value || !trackingSummary.value.recipients) return []
  let list = trackingSummary.value.recipients

  if (trackingFilter.value === 'confirmed') {
    list = list.filter((r: any) => r.is_confirmed)
  } else if (trackingFilter.value === 'pending') {
    list = list.filter((r: any) => !r.is_confirmed)
  }

  if (trackingSearch.value.trim()) {
    const q = trackingSearch.value.toLowerCase()
    list = list.filter((r: any) => {
      const name = (r.employee_name || '').toLowerCase()
      const nik = (r.employee_id_number || '').toLowerCase()
      const dept = (r.department_name || '').toLowerCase()
      const job = (r.job_title || '').toLowerCase()
      const role = (r.role || '').toLowerCase()
      return name.includes(q) || nik.includes(q) || dept.includes(q) || job.includes(q) || role.includes(q)
    })
  }

  return list
})

// ─── API METHODS ────────────────────────────────────────────────
const fetchAnnouncements = async () => {
  isLoading.value = true
  try {
    const response = await apiClient.get('/admin/announcements')
    announcements.value = response.data.data || []
  } catch (error) {
    console.error('Failed to fetch announcements:', error)
    toast.error('Gagal mengambil data pengumuman')
  } finally {
    isLoading.value = false
  }
}

const fetchEmployees = async () => {
  try {
    const res = await apiClient.get('/employees?limit=all&status=active')
    allEmployees.value = res.data.data || []
  } catch (error) {
    console.error('Failed to fetch employees:', error)
  }
}

const fetchJobPositions = async () => {
  try {
    const res = await apiClient.get('/job-positions')
    jobPositions.value = res.data.data || []
  } catch (error) {
    console.error('Failed to fetch job positions:', error)
  }
}

const fetchAccessRoles = async () => {
  try {
    const res = await apiClient.get('/access-roles')
    accessRoles.value = res.data.data || []
  } catch (error) {
    console.error('Failed to fetch access roles:', error)
  }
}

const selectedRolesCountEmployees = computed(() => {
  if (currentAnnouncement.value.target_roles.length === 0) return 0
  const selected = new Set(currentAnnouncement.value.target_roles.map((r) => r.toLowerCase()))
  const matchedUserIds = new Set<string>()

  allEmployees.value.forEach((emp) => {
    const jobTitle = (emp.job_position?.title || '').toLowerCase()
    const userRole = (emp.user?.role || emp.role || '').toLowerCase()
    const uId = emp.user_id || emp.user?.id || emp.id

    if (selected.has(jobTitle) || selected.has(userRole)) {
      matchedUserIds.add(uId)
    }
  })

  return matchedUserIds.size
})

const selectAllFilteredRoles = () => {
  const currentList = roleTargetSubTab.value === 'positions' ? filteredJobPositions.value : filteredAccessRoles.value
  currentList.forEach((item) => {
    if (!currentAnnouncement.value.target_roles.includes(item.value)) {
      currentAnnouncement.value.target_roles.push(item.value)
    }
  })
}

const clearSelectedRoles = () => {
  currentAnnouncement.value.target_roles = []
}

const openAddModal = (mode: 'general' | 'direct' = 'direct') => {
  isEditMode.value = false
  currentAnnouncement.value = {
    id: '',
    title: '',
    content: '',
    category: mode === 'direct' ? 'urgent_call' : 'policy',
    is_active: true,
    media_type: 'none',
    image_url: '',
    media_urls_text: '',
    external_url: '',
    target_type: mode === 'direct' ? 'individual' : 'broadcast',
    target_roles: [],
    target_user_ids: [],
    is_popup: mode === 'direct',
    popup_expires_at: '',
    track_downloads: false,
    track_shares: false,
  }
  employeeSearchQuery.value = ''
  isModalOpen.value = true
}

const openEditModal = (ann: any) => {
  isEditMode.value = true
  const roleTargets: string[] = []
  const userTargets: string[] = []

  if (ann.targets && Array.isArray(ann.targets)) {
    ann.targets.forEach((t: any) => {
      if (t.target_role) roleTargets.push(t.target_role)
      if (t.target_user_id) userTargets.push(t.target_user_id)
    })
  }

  const mediaUrls = Array.isArray(ann.media_urls) ? ann.media_urls : []

  currentAnnouncement.value = {
    id: ann.id,
    title: ann.title,
    content: ann.content,
    category: ann.category || 'policy',
    is_active: ann.is_active,
    media_type: ann.media_type || 'none',
    image_url: ann.image_url || '',
    media_urls_text: mediaUrls.join('\n'),
    external_url: ann.external_url || '',
    target_type: ann.target_type || 'broadcast',
    target_roles: roleTargets,
    target_user_ids: userTargets,
    is_popup: ann.is_popup ?? false,
    popup_expires_at: ann.popup_expires_at ? ann.popup_expires_at.substring(0, 16) : '',
    track_downloads: ann.track_downloads ?? false,
    track_shares: ann.track_shares ?? false,
  }
  employeeSearchQuery.value = ''
  isModalOpen.value = true
}

const toggleRoleSelection = (role: string) => {
  const idx = currentAnnouncement.value.target_roles.indexOf(role)
  if (idx > -1) {
    currentAnnouncement.value.target_roles.splice(idx, 1)
  } else {
    currentAnnouncement.value.target_roles.push(role)
  }
}

const toggleEmployeeSelection = (userId: string) => {
  if (!userId) return
  const idx = currentAnnouncement.value.target_user_ids.indexOf(userId)
  if (idx > -1) {
    currentAnnouncement.value.target_user_ids.splice(idx, 1)
  } else {
    currentAnnouncement.value.target_user_ids.push(userId)
  }
}

const selectAllFilteredEmployees = () => {
  filteredEmployees.value.forEach((emp) => {
    const uId = emp.user_id || emp.user?.id
    if (uId && !currentAnnouncement.value.target_user_ids.includes(uId)) {
      currentAnnouncement.value.target_user_ids.push(uId)
    }
  })
}

const clearSelectedEmployees = () => {
  currentAnnouncement.value.target_user_ids = []
}

// Media Upload & GDrive Helpers
const isUploadingMedia = ref(false)
const mediaUploadProgress = ref(0)
const uploadStatusText = ref('')
const mediaInputRef = ref<HTMLInputElement | null>(null)
const isPreviewVideoModalOpen = ref(false)
const selectedVideoToPreview = ref<{ url: string; title: string } | null>(null)

const getMediaUrl = (url?: string) => {
  if (!url) return ''
  const trimmed = url.trim()
  if (trimmed.startsWith('http://') || trimmed.startsWith('https://') || trimmed.startsWith('blob:')) {
    return trimmed
  }
  const apiBase = (import.meta.env.VITE_API_BASE_URL || '').replace(/\/api\/?$/, '')
  if (trimmed.startsWith('/')) {
    return `${apiBase}${trimmed}`
  }
  return `${apiBase}/${trimmed}`
}

const openVideoPreview = (ann: any) => {
  if (!ann || !ann.image_url) return
  selectedVideoToPreview.value = {
    url: ann.image_url,
    title: ann.title || 'Pratinjau Video Pengumuman',
  }
  isPreviewVideoModalOpen.value = true
}

const isGDriveUrl = (url?: string) => {
  if (!url) return false
  return url.includes('drive.google.com') || url.includes('/file/d/')
}

const extractGDriveId = (url?: string) => {
  if (!url) return ''
  const match = url.match(/\/file\/d\/([a-zA-Z0-9_-]+)/) || url.match(/id=([a-zA-Z0-9_-]+)/)
  return match ? match[1] : ''
}

const getGDriveEmbedUrl = (url?: string) => {
  const id = extractGDriveId(url)
  return id ? `https://drive.google.com/file/d/${id}/preview` : (url || '')
}

const isVideoAnnouncement = (ann: any) => {
  if (!ann) return false
  const mediaType = (ann.media_type || '').toLowerCase()
  const url = (ann.image_url || '').toLowerCase()
  return (
    mediaType === 'video' ||
    mediaType === 'video_url' ||
    url.endsWith('.mp4') ||
    url.endsWith('.mov') ||
    url.endsWith('.webm') ||
    url.endsWith('.m4v') ||
    isGDriveUrl(url)
  )
}

const triggerMediaUpload = () => {
  mediaInputRef.value?.click()
}

const handleMediaFileUpload = async (event: Event) => {
  const target = event.target as HTMLInputElement
  const file = target.files?.[0]
  if (!file) return

  // Validasi format & ukuran: Gambar 15MB, Video 500MB
  const isVideo = file.type.startsWith('video/') || /\.(mp4|mov|webm|m4v)$/i.test(file.name)
  const isImage = file.type.startsWith('image/') || /\.(jpg|jpeg|png|webp|gif)$/i.test(file.name)

  if (!isVideo && !isImage) {
    toast.error('Hanya berkas gambar (JPG, PNG, WebP) atau video (MP4, MOV, WebM) yang didukung.')
    target.value = ''
    return
  }

  if (isImage && file.size > 15 * 1024 * 1024) {
    toast.error('Ukuran gambar maksimal 15MB.')
    target.value = ''
    return
  }
  if (isVideo && file.size > 500 * 1024 * 1024) {
    toast.error('Ukuran video maksimal 500MB.')
    target.value = ''
    return
  }

  isUploadingMedia.value = true
  mediaUploadProgress.value = 0
  uploadStatusText.value = 'Menyiapkan berkas...'

  try {
    let uploaded: any = null

    // Jika video berukuran > 20MB, gunakan mekanisme Chunked Upload (20MB per chunk)
    // agar lolos 100% dari batasan 100MB Cloudflare dan aman dari timeout.
    if (isVideo && file.size > 20 * 1024 * 1024) {
      const chunkSize = 20 * 1024 * 1024 // 20MB per chunk
      const totalChunks = Math.ceil(file.size / chunkSize)
      const uploadID = 'up_' + Date.now() + '_' + Math.random().toString(36).substring(2, 9)

      for (let i = 0; i < totalChunks; i++) {
        const start = i * chunkSize
        const end = Math.min(file.size, start + chunkSize)
        const chunkBlob = file.slice(start, end)

        uploadStatusText.value = `Mengunggah potongan ${i + 1} dari ${totalChunks} (Cloudflare bypass)...`

        const chunkForm = new FormData()
        chunkForm.append('upload_id', uploadID)
        chunkForm.append('chunk_index', i.toString())
        chunkForm.append('total_chunks', totalChunks.toString())
        chunkForm.append('file_name', file.name)
        chunkForm.append('chunk', chunkBlob, file.name)

        const res = await apiClient.post('/admin/announcements/upload-chunk', chunkForm, {
          headers: { 'Content-Type': 'multipart/form-data' },
          timeout: 600000,
        })

        mediaUploadProgress.value = Math.round(((i + 1) / totalChunks) * 100)
        if (i === totalChunks - 1) {
          uploadStatusText.value = 'Menggabungkan & menyimpan ke Google Drive...'
          uploaded = res.data?.data
        }
      }
    } else {
      // Single upload untuk gambar atau video kecil (<= 20MB)
      uploadStatusText.value = 'Mengunggah ke server...'
      const formData = new FormData()
      formData.append('media', file)

      const res = await apiClient.post('/admin/announcements/upload', formData, {
        headers: { 'Content-Type': 'multipart/form-data' },
        timeout: 600000,
        onUploadProgress: (progressEvent) => {
          if (progressEvent.total) {
            mediaUploadProgress.value = Math.round((progressEvent.loaded * 100) / progressEvent.total)
          }
        },
      })
      uploaded = res.data?.data
    }

    if (uploaded?.url) {
      if (uploaded.media_type === 'video') {
        currentAnnouncement.value.media_type = 'video'
        currentAnnouncement.value.image_url = uploaded.url
        if (uploaded.is_gdrive) {
          toast.success('Video berhasil diunggah ke Google Drive dan siap ditonton karyawan!')
        } else {
          toast.success('Video berhasil diunggah!')
        }
      } else {
        if (!currentAnnouncement.value.image_url) {
          currentAnnouncement.value.image_url = uploaded.url
          currentAnnouncement.value.media_type = 'image'
        } else {
          // Tambahkan ke multi-image URLs
          const existing = currentAnnouncement.value.media_urls_text
            ? currentAnnouncement.value.media_urls_text.split('\n').filter(Boolean)
            : []
          existing.push(uploaded.url)
          currentAnnouncement.value.media_urls_text = existing.join('\n')
          currentAnnouncement.value.media_type = 'multi_image'
        }
        toast.success('Gambar berhasil diunggah!')
      }
    }
  } catch (err: any) {
    if (err.response?.status === 413) {
      toast.error('Ukuran berkas melebihi batas yang diizinkan proxy Cloudflare (Error 413: Payload Too Large). Pastikan ukuran video maksimal 500MB.')
    } else {
      toast.error('Gagal mengunggah media: ' + (err.response?.data?.message || err.message))
    }
  } finally {
    isUploadingMedia.value = false
    uploadStatusText.value = ''
    target.value = ''
  }
}

const removeUploadedMedia = (urlToRemove: string) => {
  if (currentAnnouncement.value.image_url === urlToRemove) {
    currentAnnouncement.value.image_url = ''
    if (currentAnnouncement.value.media_type === 'video') {
      currentAnnouncement.value.media_type = 'none'
    }
  }
  const lines = currentAnnouncement.value.media_urls_text
    .split('\n')
    .map((s) => s.trim())
    .filter((s) => s && s !== urlToRemove)
  currentAnnouncement.value.media_urls_text = lines.join('\n')
  if (!currentAnnouncement.value.image_url && lines.length > 0) {
    currentAnnouncement.value.image_url = lines[0] ?? ''
  }
  if (!currentAnnouncement.value.image_url && lines.length === 0) {
    currentAnnouncement.value.media_type = 'none'
  }
}

// Deteksi otomatis tautan Google Drive atau file video yang dimasukkan secara manual
watch(
  () => currentAnnouncement.value.image_url,
  (newUrl) => {
    if (newUrl) {
      const trimmed = newUrl.trim()
      if (isGDriveUrl(trimmed) || /\.(mp4|mov|webm|m4v)(\?.*)?$/i.test(trimmed)) {
        currentAnnouncement.value.media_type = 'video'
      }
    }
  }
)

const saveAnnouncement = async () => {
  if (!currentAnnouncement.value.title.trim() || !currentAnnouncement.value.content.trim()) {
    toast.error('Judul dan isi pengumuman wajib diisi')
    return
  }

  if (currentAnnouncement.value.target_type === 'role' && currentAnnouncement.value.target_roles.length === 0) {
    toast.error('Pilih minimal satu jabatan/role target penerima')
    return
  }

  if (currentAnnouncement.value.target_type === 'individual' && currentAnnouncement.value.target_user_ids.length === 0) {
    toast.error('Pilih minimal satu karyawan target penerima')
    return
  }

  isSubmitting.value = true
  try {
    const parsedMediaUrls = currentAnnouncement.value.media_urls_text
      .split('\n')
      .map((s) => s.trim())
      .filter((s) => s.length > 0)

    const imgTrimmed = currentAnnouncement.value.image_url.trim()
    let effectiveMediaType = currentAnnouncement.value.media_type || 'none'
    const isVid =
      /\.(mp4|mov|webm|m4v)(\?.*)?$/i.test(imgTrimmed) ||
      isGDriveUrl(imgTrimmed) ||
      effectiveMediaType === 'video'
    if (isVid && imgTrimmed) {
      effectiveMediaType = 'video'
    } else if (effectiveMediaType === 'none') {
      if (parsedMediaUrls.length > 1) {
        effectiveMediaType = 'multi_image'
      } else if (parsedMediaUrls.length === 1 || imgTrimmed) {
        effectiveMediaType = 'image'
      } else if (currentAnnouncement.value.external_url.trim()) {
        effectiveMediaType = 'social_media'
      }
    }

    const payload: any = {
      title: currentAnnouncement.value.title,
      content: currentAnnouncement.value.content,
      category: currentAnnouncement.value.category,
      is_active: currentAnnouncement.value.is_active,
      media_type: effectiveMediaType,
      target_type: currentAnnouncement.value.target_type,
      is_popup: currentAnnouncement.value.is_popup,
      track_downloads: currentAnnouncement.value.track_downloads,
      track_shares: currentAnnouncement.value.track_shares,
      image_url: imgTrimmed,
      media_urls: parsedMediaUrls,
      external_url: currentAnnouncement.value.external_url.trim(),
    }

    if (currentAnnouncement.value.target_type === 'role') {
      payload.target_roles = currentAnnouncement.value.target_roles
    } else if (currentAnnouncement.value.target_type === 'individual') {
      payload.target_user_ids = currentAnnouncement.value.target_user_ids
    }

    if (currentAnnouncement.value.popup_expires_at) {
      payload.popup_expires_at = new Date(currentAnnouncement.value.popup_expires_at).toISOString()
    }

    if (isEditMode.value) {
      await apiClient.put(`/announcements/${currentAnnouncement.value.id}`, payload)
      toast.success('Pengumuman berhasil diperbarui!')
    } else {
      await apiClient.post('/announcements', payload)
      toast.success('Pengumuman & Direct Notifikasi FCM berhasil dikirimkan!')
    }
    isModalOpen.value = false
    fetchAnnouncements()
  } catch (error: any) {
    toast.error('Gagal menyimpan pengumuman: ' + (error.response?.data?.message || error.message))
  } finally {
    isSubmitting.value = false
  }
}

const deleteAnnouncement = async (id: string) => {
  if (!confirm('Apakah Anda yakin ingin menghapus pengumuman ini?')) return
  try {
    await apiClient.delete(`/announcements/${id}`)
    toast.success('Pengumuman berhasil dihapus!')
    fetchAnnouncements()
  } catch (error: any) {
    toast.error('Gagal menghapus: ' + (error.response?.data?.message || error.message))
  }
}

// ─── RESEND FCM (MANUAL TRIGGER) ────────────────────────────────
const isResendingFCM = ref<Record<string, boolean>>({})
const isResendModalOpen = ref(false)
const selectedResendAnnouncement = ref<any>(null)

const openResendModal = (ann: any) => {
  selectedResendAnnouncement.value = ann
  isResendModalOpen.value = true
}

const resendFCM = async (
  annId: string,
  mode: 'all' | 'pending_only' | 'single',
  userId?: string,
  userName?: string
) => {
  const loadingKey = mode === 'single' ? `${annId}_${userId}` : `${annId}_${mode}`
  isResendingFCM.value[loadingKey] = true
  try {
    const payload: any = { mode }
    if (userId) {
      payload.target_user_id = userId
    }
    const res = await apiClient.post(`/admin/announcements/${annId}/resend-fcm`, payload)
    const count = res.data.data?.recipients_count ?? 0

    if (mode === 'single') {
      toast.success(`Notifikasi FCM berhasil dikirim ke ${userName || 'karyawan'}!`)
    } else if (mode === 'all') {
      toast.success(`Notifikasi FCM berhasil disiarkan ulang ke seluruh target (${count > 0 ? count + ' penerima' : 'semua'})!`)
    } else {
      if (count > 0) {
        toast.success(`Notifikasi FCM berhasil dikirim ulang ke ${count} karyawan yang belum konfirmasi!`)
      } else {
        toast.info('Semua karyawan target telah melakukan konfirmasi.')
      }
    }

    isResendModalOpen.value = false

    if (isTrackingModalOpen.value && trackingSummary.value?.announcement_id === annId) {
      const resTrack = await apiClient.get(`/admin/announcements/${annId}/confirmations`)
      trackingSummary.value = resTrack.data.data
    }
  } catch (error: any) {
    console.error('Failed to resend FCM:', error)
    toast.error('Gagal mengirim ulang FCM: ' + (error.response?.data?.message || error.message))
  } finally {
    isResendingFCM.value[loadingKey] = false
  }
}

// ─── TRACKING MODAL ─────────────────────────────────────────────
const openTrackingModal = async (announcementId: string) => {
  isTrackingModalOpen.value = true
  isLoadingTracking.value = true
  trackingFilter.value = 'all'
  trackingSearch.value = ''
  try {
    const res = await apiClient.get(`/admin/announcements/${announcementId}/confirmations`)
    trackingSummary.value = res.data.data
  } catch (error: any) {
    console.error('Failed to fetch confirmations:', error)
    toast.error('Gagal mengambil data konfirmasi: ' + (error.response?.data?.message || error.message))
    isTrackingModalOpen.value = false
  } finally {
    isLoadingTracking.value = false
  }
}

const formatDate = (dateStr?: string) => {
  if (!dateStr) return '-'
  const d = new Date(dateStr)
  return new Intl.DateTimeFormat('id-ID', {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  }).format(d)
}

// ─── COMMENTS MODAL & MODERATION ─────────────────────────────────
const isCommentsModalOpen = ref(false)
const selectedAnnouncementForComments = ref<any>(null)
const commentsList = ref<any[]>([])
const isLoadingComments = ref(false)
const newCommentText = ref('')
const isSubmittingComment = ref(false)
const isDeletingComment = ref<Record<string, boolean>>({})

const openCommentsModal = async (ann: any) => {
  selectedAnnouncementForComments.value = ann
  isCommentsModalOpen.value = true
  newCommentText.value = ''
  await fetchComments(ann.id)
}

const fetchComments = async (announcementId: string) => {
  isLoadingComments.value = true
  try {
    const res = await apiClient.get(`/announcements/${announcementId}/comments?limit=100`)
    commentsList.value = res.data.data?.comments || []
    if (selectedAnnouncementForComments.value) {
      selectedAnnouncementForComments.value.comments_count = res.data.data?.total ?? commentsList.value.length
    }
  } catch (error: any) {
    console.error('Failed to fetch comments:', error)
    toast.error('Gagal memuat komentar: ' + (error.response?.data?.message || error.message))
  } finally {
    isLoadingComments.value = false
  }
}

const submitComment = async () => {
  if (!newCommentText.value.trim() || !selectedAnnouncementForComments.value) return
  isSubmittingComment.value = true
  try {
    const annId = selectedAnnouncementForComments.value.id
    await apiClient.post(`/announcements/${annId}/comments`, {
      content: newCommentText.value.trim(),
    })
    toast.success('Komentar / respon admin berhasil dikirim!')
    newCommentText.value = ''
    await fetchComments(annId)
    const ann = announcements.value.find((a) => a.id === annId)
    if (ann) {
      ann.comments_count = (ann.comments_count || 0) + 1
    }
  } catch (error: any) {
    console.error('Failed to submit comment:', error)
    toast.error('Gagal mengirim komentar: ' + (error.response?.data?.message || error.message))
  } finally {
    isSubmittingComment.value = false
  }
}

const deleteCommentItem = async (commentId: string) => {
  if (!confirm('Apakah Anda yakin ingin menghapus komentar ini?')) return
  if (!selectedAnnouncementForComments.value) return
  const annId = selectedAnnouncementForComments.value.id
  isDeletingComment.value[commentId] = true
  try {
    await apiClient.delete(`/announcements/${annId}/comments/${commentId}`)
    toast.success('Komentar berhasil dihapus')
    commentsList.value = commentsList.value.filter((c) => c.id !== commentId)
    const ann = announcements.value.find((a) => a.id === annId)
    if (ann && ann.comments_count > 0) {
      ann.comments_count -= 1
    }
    if (selectedAnnouncementForComments.value && selectedAnnouncementForComments.value.comments_count > 0) {
      selectedAnnouncementForComments.value.comments_count -= 1
    }
  } catch (error: any) {
    console.error('Failed to delete comment:', error)
    toast.error('Gagal menghapus komentar: ' + (error.response?.data?.message || error.message))
  } finally {
    isDeletingComment.value[commentId] = false
  }
}

onMounted(() => {
  fetchAnnouncements()
  fetchEmployees()
  fetchJobPositions()
  fetchAccessRoles()
})
</script>

<template>
  <div class="space-y-6">
    <!-- Header Section -->
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4 bg-white p-6 rounded-2xl border border-slate-100 shadow-sm">
      <div>
        <div class="flex items-center gap-2">
          <div class="p-2.5 rounded-xl bg-primary/10 text-primary">
            <Megaphone class="w-6 h-6" />
          </div>
          <div>
            <h1 class="text-xl sm:text-2xl font-black text-gray-900 tracking-tight">
              Manajemen Pengumuman & Panggilan Direct
            </h1>
            <p class="text-xs sm:text-sm text-gray-500 font-medium">
              Siarkan informasi dan instruksi panggilan menghadap/meeting dengan notifikasi FCM & konfirmasi pop-up.
            </p>
          </div>
        </div>
      </div>
      <div class="flex flex-wrap items-center gap-2.5">
        <Button
          @click="openAddModal('direct')"
          class="gap-2 rounded-xl bg-red-600 hover:bg-red-700 text-white font-bold shadow-lg shadow-red-600/20 text-xs sm:text-sm h-10 px-4"
        >
          <AlertTriangle class="w-4 h-4" />
          Buat Panggilan / Pop-up
        </Button>
        <Button
          @click="openAddModal('general')"
          variant="outline"
          class="gap-2 rounded-xl font-bold text-xs sm:text-sm h-10 px-4 border-slate-200"
        >
          <Megaphone class="w-4 h-4 text-primary" />
          Pengumuman Biasa
        </Button>
      </div>
    </div>

    <!-- Tabs Navigation -->
    <div class="flex items-center gap-2 border-b border-slate-200 pb-2">
      <button
        @click="activeTab = 'direct'"
        :class="[
          'flex items-center gap-2 px-5 py-2.5 rounded-xl font-bold text-xs sm:text-sm transition-all',
          activeTab === 'direct'
            ? 'bg-primary text-white shadow-md shadow-primary/25'
            : 'text-slate-600 hover:text-slate-900 hover:bg-slate-100',
        ]"
      >
        <UserCheck class="w-4 h-4" />
        Direct Panggilan & Konfirmasi (Pop-up)
        <span
          class="ml-1.5 px-2 py-0.5 rounded-full text-[11px] font-black"
          :class="activeTab === 'direct' ? 'bg-white/20 text-white' : 'bg-slate-200 text-slate-700'"
        >
          {{ directAnnouncements.length }}
        </span>
      </button>

      <button
        @click="activeTab = 'general'"
        :class="[
          'flex items-center gap-2 px-5 py-2.5 rounded-xl font-bold text-xs sm:text-sm transition-all',
          activeTab === 'general'
            ? 'bg-primary text-white shadow-md shadow-primary/25'
            : 'text-slate-600 hover:text-slate-900 hover:bg-slate-100',
        ]"
      >
        <Megaphone class="w-4 h-4" />
        Pengumuman Umum & Kebijakan
        <span
          class="ml-1.5 px-2 py-0.5 rounded-full text-[11px] font-black"
          :class="activeTab === 'general' ? 'bg-white/20 text-white' : 'bg-slate-200 text-slate-700'"
        >
          {{ generalAnnouncements.length }}
        </span>
      </button>
    </div>

    <!-- TAB 1: DIRECT PANGGILAN & KONFIRMASI -->
    <div v-if="activeTab === 'direct'" class="space-y-4">
      <div class="bg-amber-50/70 border border-amber-200/80 rounded-2xl p-4 flex items-start gap-3">
        <AlertTriangle class="w-5 h-5 text-amber-600 shrink-0 mt-0.5" />
        <div class="text-xs text-amber-900 leading-relaxed">
          <strong class="font-bold">Mekanisme Direct Panggilan:</strong> Pengumuman di tab ini dikirimkan langsung ke HP target karyawan (Sales, Driver, SPV, dll) dengan notifikasi FCM berdering. Karyawan akan melihat <strong>Modal Pop-up Wajib Konfirmasi</strong> saat membuka app dan mengklik tombol "SAYA MENGERTI / KONFIRMASI". Anda dapat memantau data historis siapa saja yang telah konfirmasi secara real-time.
        </div>
      </div>

      <div class="bg-white rounded-2xl border border-slate-100 shadow-sm overflow-hidden">
        <div class="overflow-x-auto">
          <table class="w-full text-left text-sm text-slate-600">
            <thead class="bg-slate-50/80 text-xs uppercase font-extrabold text-slate-500 tracking-wider border-b border-slate-100">
              <tr>
                <th class="py-4 px-6">Judul Panggilan & Kategori</th>
                <th class="py-4 px-6">Target Penerima</th>
                <th class="py-4 px-6">Status Konfirmasi</th>
                <th class="py-4 px-6">Pop-up & Tanggal</th>
                <th class="py-4 px-6 text-right">Aksi</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-100">
              <tr v-if="isLoading" class="text-center">
                <td colspan="5" class="py-12 text-slate-400">
                  <div class="flex items-center justify-center gap-2">
                    <RefreshCw class="w-5 h-5 animate-spin text-primary" />
                    Memuat data panggilan...
                  </div>
                </td>
              </tr>
              <tr v-else-if="directAnnouncements.length === 0" class="text-center">
                <td colspan="5" class="py-12 text-slate-400">
                  <div class="flex flex-col items-center justify-center gap-2">
                    <UserCheck class="w-10 h-10 text-slate-300" />
                    <p class="font-medium">Belum ada direct panggilan atau pengumuman ber-popup.</p>
                    <Button @click="openAddModal('direct')" size="sm" class="mt-2 rounded-xl">
                      Buat Panggilan Baru
                    </Button>
                  </div>
                </td>
              </tr>
              <tr
                v-for="ann in directAnnouncements"
                :key="ann.id"
                class="hover:bg-slate-50/80 transition-colors"
              >
                <!-- Judul & Kategori -->
                <td class="py-4 px-6">
                  <div class="flex items-start gap-3 max-w-[340px]">
                    <div
                      v-if="ann.image_url"
                      class="shrink-0 w-12 h-12 rounded-xl overflow-hidden border border-slate-200 bg-slate-100 flex items-center justify-center transition-transform hover:scale-105"
                      :class="isVideoAnnouncement(ann) ? 'cursor-pointer' : ''"
                      @click="isVideoAnnouncement(ann) ? openVideoPreview(ann) : null"
                    >
                      <div v-if="isVideoAnnouncement(ann)" class="w-full h-full bg-slate-900 flex flex-col items-center justify-center text-white group hover:bg-slate-800" title="Klik untuk Putar Video">
                        <Play class="w-5 h-5 text-indigo-400 fill-indigo-400" />
                        <span class="text-[8px] font-black tracking-tighter uppercase mt-0.5">Video</span>
                      </div>
                      <img v-else :src="getMediaUrl(ann.image_url)" class="w-full h-full object-cover" alt="Media" />
                    </div>
                    <div class="flex flex-col gap-1 min-w-0">
                      <div class="flex items-center gap-1.5 flex-wrap">
                        <span
                          v-if="ann.category === 'social_media'"
                          class="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-[10px] font-black uppercase tracking-wider bg-purple-100 text-purple-700"
                        >
                          <Share2 class="w-3 h-3" /> Postingan Medsos
                        </span>
                        <span
                          v-else-if="ann.category === 'promo'"
                          class="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-[10px] font-black uppercase tracking-wider bg-emerald-100 text-emerald-700"
                        >
                          <Sparkles class="w-3 h-3" /> Promo & Brosur
                        </span>
                        <span
                          v-else-if="ann.category === 'urgent_call'"
                          class="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-[10px] font-black uppercase tracking-wider bg-red-100 text-red-700"
                        >
                          <AlertTriangle class="w-3 h-3" /> Panggilan Menghadap
                        </span>
                        <span
                          v-else-if="ann.category === 'event'"
                          class="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-[10px] font-black uppercase tracking-wider bg-blue-100 text-blue-700"
                        >
                          <Calendar class="w-3 h-3" /> Undangan Meeting
                        </span>
                        <span
                          v-else
                          class="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-[10px] font-black uppercase tracking-wider bg-slate-100 text-slate-700"
                        >
                          {{ ann.category || 'INFO' }}
                        </span>

                        <span
                          v-if="ann.likes_count && ann.likes_count > 0"
                          class="inline-flex items-center gap-1 px-1.5 py-0.5 rounded-full text-[10px] font-bold bg-red-50 text-red-600 border border-red-200"
                        >
                          <Heart class="w-3 h-3 fill-red-500 text-red-500" /> {{ ann.likes_count }}
                        </span>

                        <span
                          v-if="ann.comments_count && ann.comments_count > 0"
                          @click="openCommentsModal(ann)"
                          class="inline-flex items-center gap-1 px-1.5 py-0.5 rounded-full text-[10px] font-bold bg-blue-50 text-blue-600 border border-blue-200 cursor-pointer hover:bg-blue-100 transition-colors"
                          title="Lihat Komentar"
                        >
                          <MessageSquare class="w-3 h-3 text-blue-500" /> {{ ann.comments_count }}
                        </span>
                      </div>
                      <span class="font-bold text-gray-900 text-[14px] line-clamp-1">
                        {{ ann.title }}
                      </span>
                      <span class="text-xs text-gray-500 line-clamp-1">
                        {{ ann.content }}
                      </span>
                    </div>
                  </div>
                </td>

                <!-- Target Penerima -->
                <td class="py-4 px-6">
                  <div class="flex flex-col gap-1">
                    <span
                      v-if="ann.target_type === 'broadcast'"
                      class="inline-flex items-center gap-1.5 text-xs font-bold text-indigo-700 bg-indigo-50 px-2.5 py-1 rounded-lg w-max"
                    >
                      <Users class="w-3.5 h-3.5" /> Semua Karyawan (Broadcast)
                    </span>
                    <span
                      v-else-if="ann.target_type === 'role'"
                      class="inline-flex items-center gap-1.5 text-xs font-bold text-purple-700 bg-purple-50 px-2.5 py-1 rounded-lg w-max"
                    >
                      <Briefcase class="w-3.5 h-3.5" />
                      Role Tertentu ({{ ann.targets?.length || 0 }} Role)
                    </span>
                    <span
                      v-else-if="ann.target_type === 'individual'"
                      class="inline-flex items-center gap-1.5 text-xs font-bold text-amber-700 bg-amber-50 px-2.5 py-1 rounded-lg w-max"
                    >
                      <UserCheck class="w-3.5 h-3.5" />
                      {{ ann.total_targets || ann.targets?.length || 0 }} Karyawan Tertentu
                    </span>
                  </div>
                </td>

                <!-- Status Konfirmasi -->
                <td class="py-4 px-6">
                  <div class="flex flex-col gap-1.5 min-w-[160px]">
                    <div class="flex items-center justify-between text-xs font-bold">
                      <span class="text-slate-700">
                        {{ ann.confirmed_count || 0 }} / {{ ann.total_targets || 0 }} Konfirmasi
                      </span>
                      <span
                        :class="[
                          'font-black',
                          (ann.confirmed_count || 0) >= (ann.total_targets || 1)
                            ? 'text-emerald-600'
                            : 'text-primary',
                        ]"
                      >
                        {{
                          ann.total_targets > 0
                            ? Math.round(((ann.confirmed_count || 0) / ann.total_targets) * 100)
                            : 0
                        }}%
                      </span>
                    </div>
                    <div class="w-full bg-slate-100 h-2 rounded-full overflow-hidden">
                      <div
                        class="h-full bg-primary rounded-full transition-all"
                        :style="{
                          width: `${
                            ann.total_targets > 0
                              ? Math.min(100, Math.round(((ann.confirmed_count || 0) / ann.total_targets) * 100))
                              : 0
                          }%`,
                        }"
                      ></div>
                    </div>
                  </div>
                </td>

                <!-- Pop-up & Tanggal -->
                <td class="py-4 px-6">
                  <div class="flex flex-col gap-0.5 text-xs">
                    <span
                      v-if="ann.is_popup"
                      class="font-black text-red-600 flex items-center gap-1"
                    >
                      <CheckCircle2 class="w-3 h-3" /> Pop-up Aktif
                    </span>
                    <span class="text-slate-400">
                      {{ formatDate(ann.created_at) }}
                    </span>
                  </div>
                </td>

                <!-- Aksi -->
                <td class="py-4 px-6 text-right">
                  <div class="flex items-center justify-end gap-1.5">
                    <Button
                      @click="openCommentsModal(ann)"
                      variant="outline"
                      size="sm"
                      class="h-8 px-2.5 rounded-lg text-xs font-bold gap-1 text-blue-600 border-blue-200 bg-blue-50/70 hover:bg-blue-100"
                      title="Lihat & Moderasi Komentar"
                    >
                      <MessageSquare class="w-3.5 h-3.5" />
                      <span v-if="ann.comments_count">{{ ann.comments_count }}</span>
                      <span v-else>Komentar</span>
                    </Button>
                    <Button
                      @click="openResendModal(ann)"
                      variant="outline"
                      size="sm"
                      class="h-8 px-2.5 rounded-lg text-xs font-bold gap-1 text-amber-600 border-amber-200 bg-amber-50/70 hover:bg-amber-100"
                      title="Kirim Ulang Notifikasi FCM"
                    >
                      <BellRing class="w-3.5 h-3.5" />
                      Kirim FCM
                    </Button>
                    <Button
                      @click="openTrackingModal(ann.id)"
                      variant="outline"
                      size="sm"
                      class="h-8 px-3 rounded-lg text-xs font-bold gap-1.5 bg-primary/5 hover:bg-primary/10 text-primary border-primary/20"
                    >
                      <Users class="w-3.5 h-3.5" />
                      Tracking
                    </Button>
                    <Button
                      @click="openEditModal(ann)"
                      variant="ghost"
                      size="sm"
                      class="h-8 w-8 p-0 rounded-lg text-slate-500 hover:text-primary"
                    >
                      <Pencil class="w-4 h-4" />
                    </Button>
                    <Button
                      @click="deleteAnnouncement(ann.id)"
                      variant="ghost"
                      size="sm"
                      class="h-8 w-8 p-0 rounded-lg text-red-500 hover:bg-red-50"
                    >
                      <Trash2 class="w-4 h-4" />
                    </Button>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>

    <!-- TAB 2: PENGUMUMAN UMUM & KEBIJAKAN -->
    <div v-else-if="activeTab === 'general'" class="space-y-4">
      <div class="bg-white rounded-2xl border border-slate-100 shadow-sm overflow-hidden">
        <div class="overflow-x-auto">
          <table class="w-full text-left text-sm text-slate-600">
            <thead class="bg-slate-50/80 text-xs uppercase font-extrabold text-slate-500 tracking-wider border-b border-slate-100">
              <tr>
                <th class="py-4 px-6">Judul Pengumuman</th>
                <th class="py-4 px-6">Kategori</th>
                <th class="py-4 px-6">Status</th>
                <th class="py-4 px-6">Tanggal Dibuat</th>
                <th class="py-4 px-6 text-right">Aksi</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-100">
              <tr v-if="isLoading" class="text-center">
                <td colspan="5" class="py-12 text-slate-400">
                  <div class="flex items-center justify-center gap-2">
                    <RefreshCw class="w-5 h-5 animate-spin text-primary" />
                    Memuat data pengumuman...
                  </div>
                </td>
              </tr>
              <tr v-else-if="generalAnnouncements.length === 0" class="text-center">
                <td colspan="5" class="py-12 text-slate-400">
                  Belum ada pengumuman umum.
                </td>
              </tr>
              <tr
                v-for="ann in generalAnnouncements"
                :key="ann.id"
                class="hover:bg-slate-50/80 transition-colors"
              >
                <td class="py-4 px-6">
                  <div class="flex items-start gap-3 max-w-[420px]">
                    <div
                      v-if="ann.image_url"
                      class="shrink-0 w-12 h-12 rounded-xl overflow-hidden border border-slate-200 bg-slate-100 flex items-center justify-center transition-transform hover:scale-105"
                      :class="isVideoAnnouncement(ann) ? 'cursor-pointer' : ''"
                      @click="isVideoAnnouncement(ann) ? openVideoPreview(ann) : null"
                    >
                      <div v-if="isVideoAnnouncement(ann)" class="w-full h-full bg-slate-900 flex flex-col items-center justify-center text-white group hover:bg-slate-800" title="Klik untuk Putar Video">
                        <Play class="w-5 h-5 text-indigo-400 fill-indigo-400" />
                        <span class="text-[8px] font-black tracking-tighter uppercase mt-0.5">Video</span>
                      </div>
                      <img v-else :src="getMediaUrl(ann.image_url)" class="w-full h-full object-cover" alt="Media" />
                    </div>
                    <div class="flex flex-col min-w-0">
                      <div class="flex items-center gap-1.5 flex-wrap mb-0.5">
                        <button
                          v-if="isVideoAnnouncement(ann)"
                          @click="openVideoPreview(ann)"
                          type="button"
                          class="inline-flex items-center gap-1 px-1.5 py-0.5 rounded text-[10px] font-black bg-indigo-100 text-indigo-700 hover:bg-indigo-200 transition"
                        >
                          <Play class="w-2.5 h-2.5 fill-indigo-700" /> Tonton Video
                        </button>
                      </div>
                      <span
                        class="font-bold text-gray-900 line-clamp-1"
                        :class="isVideoAnnouncement(ann) ? 'cursor-pointer hover:text-primary transition-colors' : ''"
                        @click="isVideoAnnouncement(ann) ? openVideoPreview(ann) : null"
                      >
                        {{ ann.title }}
                      </span>
                      <span class="text-xs text-gray-500 line-clamp-2 mt-0.5">{{ ann.content }}</span>
                      <a
                        v-if="ann.external_url"
                        :href="ann.external_url"
                        target="_blank"
                        class="inline-flex items-center gap-1 text-[11px] text-purple-600 font-bold hover:underline mt-1 truncate"
                      >
                        <ExternalLink class="w-3 h-3 shrink-0" />
                        {{ ann.external_url }}
                      </a>
                    </div>
                  </div>
                </td>
                <td class="py-4 px-6">
                  <div class="flex items-center gap-1.5 flex-wrap">
                    <span
                      v-if="ann.category === 'social_media'"
                      class="px-2.5 py-1 rounded-full text-xs font-bold uppercase tracking-wider bg-purple-100 text-purple-700 flex items-center gap-1"
                    >
                      <Share2 class="w-3 h-3" /> Medsos
                    </span>
                    <span
                      v-else-if="ann.category === 'promo'"
                      class="px-2.5 py-1 rounded-full text-xs font-bold uppercase tracking-wider bg-emerald-100 text-emerald-700 flex items-center gap-1"
                    >
                      <Sparkles class="w-3 h-3" /> Promo
                    </span>
                    <span
                      v-else
                      class="px-2.5 py-1 rounded-full text-xs font-bold uppercase tracking-wider bg-slate-100 text-slate-700"
                    >
                      {{ ann.category || 'POLICY' }}
                    </span>

                    <span
                      v-if="ann.likes_count && ann.likes_count > 0"
                      class="inline-flex items-center gap-1 px-1.5 py-0.5 rounded-full text-[10px] font-bold bg-red-50 text-red-600 border border-red-200"
                    >
                      <Heart class="w-3 h-3 fill-red-500 text-red-500" /> {{ ann.likes_count }}
                    </span>

                    <span
                      v-if="ann.comments_count && ann.comments_count > 0"
                      @click="openCommentsModal(ann)"
                      class="inline-flex items-center gap-1 px-1.5 py-0.5 rounded-full text-[10px] font-bold bg-blue-50 text-blue-600 border border-blue-200 cursor-pointer hover:bg-blue-100 transition-colors"
                      title="Lihat Komentar"
                    >
                      <MessageSquare class="w-3 h-3 text-blue-500" /> {{ ann.comments_count }}
                    </span>

                    <span
                      v-if="ann.is_popup"
                      class="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-[10px] font-black uppercase tracking-wider bg-red-100 text-red-700"
                    >
                      <CheckCircle2 class="w-3 h-3" /> Pop-up On
                    </span>
                  </div>
                </td>
                <td class="py-4 px-6">
                  <span
                    :class="[
                      'px-2.5 py-0.5 rounded-full text-xs font-bold',
                      ann.is_active ? 'bg-green-100 text-green-700' : 'bg-gray-100 text-gray-600',
                    ]"
                  >
                    {{ ann.is_active ? 'AKTIF' : 'NON-AKTIF' }}
                  </span>
                </td>
                <td class="py-4 px-6 text-xs text-slate-500">
                  {{ formatDate(ann.created_at) }}
                </td>
                <td class="py-4 px-6 text-right">
                  <div class="flex items-center justify-end gap-1">
                    <Button
                      @click="openCommentsModal(ann)"
                      variant="outline"
                      size="sm"
                      class="h-8 px-2.5 rounded-lg text-xs font-bold gap-1 text-blue-600 border-blue-200 bg-blue-50/70 hover:bg-blue-100"
                      title="Lihat & Moderasi Komentar"
                    >
                      <MessageSquare class="w-3.5 h-3.5" />
                      <span v-if="ann.comments_count">{{ ann.comments_count }}</span>
                      <span v-else>Komentar</span>
                    </Button>
                    <Button
                      @click="openResendModal(ann)"
                      variant="outline"
                      size="sm"
                      class="h-8 px-2.5 rounded-lg text-xs font-bold gap-1 text-amber-600 border-amber-200 bg-amber-50/70 hover:bg-amber-100"
                      title="Kirim Ulang Notifikasi FCM"
                    >
                      <BellRing class="w-3 h-3" />
                      Kirim FCM
                    </Button>
                    <Button
                      v-if="ann.is_popup"
                      @click="openTrackingModal(ann.id)"
                      variant="outline"
                      size="sm"
                      class="h-8 px-2.5 rounded-lg text-xs font-bold gap-1 text-primary border-primary/20 bg-primary/5 hover:bg-primary/10"
                    >
                      <Users class="w-3 h-3" />
                      Tracking
                    </Button>
                    <Button
                      @click="openEditModal(ann)"
                      variant="ghost"
                      size="sm"
                      class="h-8 w-8 p-0 rounded-lg text-slate-500 hover:text-primary"
                    >
                      <Pencil class="w-4 h-4" />
                    </Button>
                    <Button
                      @click="deleteAnnouncement(ann.id)"
                      variant="ghost"
                      size="sm"
                      class="h-8 w-8 p-0 rounded-lg text-red-500 hover:bg-red-50"
                    >
                      <Trash2 class="w-4 h-4" />
                    </Button>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>

    <!-- ─── MODAL BUAT / EDIT PENGUMUMAN ───────────────────────────── -->
    <Dialog v-model:open="isModalOpen">
      <DialogContent class="sm:max-w-2xl rounded-3xl p-0 overflow-hidden border-none shadow-2xl max-h-[92vh] flex flex-col">
        <!-- Header -->
        <div class="bg-primary p-6 md:p-8 text-white relative shrink-0">
          <DialogHeader>
            <DialogTitle class="text-xl md:text-2xl font-black">
              {{ isEditMode ? 'Edit Pengumuman / Panggilan' : 'Buat Direct Panggilan & Pengumuman' }}
            </DialogTitle>
            <DialogDescription class="text-white/80 mt-1 font-medium text-xs border-l-2 border-white/30 pl-3">
              Siarkan instruksi atau pengumuman dengan opsi modal pop-up wajib konfirmasi dan notifikasi FCM.
            </DialogDescription>
          </DialogHeader>
        </div>

        <!-- Body -->
        <div class="p-6 md:p-8 bg-white space-y-6 overflow-y-auto flex-1 custom-scrollbar">
          <!-- Judul -->
          <div class="space-y-1.5">
            <Label class="text-xs font-bold text-gray-700 uppercase tracking-wider">
              Judul Pengumuman / Instruksi Panggilan <span class="text-red-500">*</span>
            </Label>
            <Input
              v-model="currentAnnouncement.title"
              placeholder="e.g. Panggilan Menghadap Manager Sales / Evaluasi Target"
              class="h-11 rounded-xl"
            />
          </div>

          <!-- Kategori & Status Aktif -->
          <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <div class="space-y-1.5">
              <Label class="text-xs font-bold text-gray-700 uppercase tracking-wider">Kategori</Label>
              <Select v-model="currentAnnouncement.category">
                <SelectTrigger class="h-11 rounded-xl">
                  <SelectValue placeholder="Pilih kategori" />
                </SelectTrigger>
                <SelectContent>
                  <SelectGroup>
                    <SelectItem v-for="cat in categories" :key="cat.value" :value="cat.value">
                      <div class="flex items-center gap-2 font-medium">
                        <component :is="cat.icon" :class="['w-4 h-4', cat.color]" />
                        {{ cat.label }}
                      </div>
                    </SelectItem>
                  </SelectGroup>
                </SelectContent>
              </Select>
            </div>

            <div class="space-y-1.5">
              <Label class="text-xs font-bold text-gray-700 uppercase tracking-wider">Status Tayang</Label>
              <div
                @click="currentAnnouncement.is_active = !currentAnnouncement.is_active"
                class="flex items-center justify-between p-2.5 rounded-xl bg-slate-50 border border-slate-200 cursor-pointer hover:bg-slate-100 transition-colors"
              >
                <span class="text-xs font-medium text-slate-700 select-none">
                  Tampilkan ke Karyawan
                </span>
                <button
                  type="button"
                  class="relative inline-flex h-6 w-11 shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none"
                  :class="currentAnnouncement.is_active ? 'bg-primary' : 'bg-slate-300'"
                >
                  <span
                    class="pointer-events-none inline-block h-5 w-5 transform rounded-full bg-white shadow-md ring-0 transition duration-200 ease-in-out"
                    :class="currentAnnouncement.is_active ? 'translate-x-5' : 'translate-x-0'"
                  />
                </button>
              </div>
            </div>
          </div>

          <!-- MEDIA & LINK SOSMED (FOTO / VIDEO / PROMOSI) - POSISI ATAS -->
          <div class="p-4 rounded-2xl bg-indigo-50/60 border border-indigo-200/80 space-y-4">
            <div class="flex items-center justify-between">
              <span class="text-xs font-black text-indigo-950 flex items-center gap-1.5 uppercase tracking-wider">
                <ImageIcon class="w-4 h-4 text-indigo-600" />
                Media Gambar, Video & Tautan Media Sosial
              </span>
              <span class="text-[10px] font-bold text-indigo-600 bg-indigo-100/80 px-2 py-0.5 rounded-md">
                Superapp Feed & Sosmed
              </span>
            </div>

            <!-- Upload File Langsung (Gambar / Video) -->
            <div class="p-3.5 rounded-xl bg-white border border-dashed border-indigo-300 space-y-3">
              <input
                ref="mediaInputRef"
                type="file"
                accept="image/jpeg,image/png,image/webp,video/mp4,video/quicktime"
                class="hidden"
                @change="handleMediaFileUpload"
              />

              <div class="flex flex-col sm:flex-row items-center justify-between gap-3">
                <div class="flex items-center gap-3">
                  <div class="w-10 h-10 rounded-xl bg-indigo-100 text-indigo-600 flex items-center justify-center shrink-0">
                    <Upload class="w-5 h-5" />
                  </div>
                  <div>
                    <h5 class="text-xs font-bold text-slate-800">Unggah Gambar atau Video Pengumuman</h5>
                    <p class="text-[11px] text-slate-500">Mendukung format JPG/PNG/WebP (Maks 15MB) & Video MP4/MOV/WebM hingga 500MB (Otomatis Chunked Upload & Google Drive)</p>
                  </div>
                </div>

                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  class="rounded-xl border-indigo-200 text-indigo-700 hover:bg-indigo-50 h-9 font-bold text-xs shrink-0"
                  :disabled="isUploadingMedia"
                  @click="triggerMediaUpload"
                >
                  <Upload v-if="!isUploadingMedia" class="w-3.5 h-3.5 mr-1.5" />
                  <RefreshCw v-else class="w-3.5 h-3.5 mr-1.5 animate-spin" />
                  {{ isUploadingMedia ? `Mengunggah (${mediaUploadProgress}%)` : 'Pilih Berkas' }}
                </Button>
              </div>

              <!-- Status Teks & Upload Progress Bar -->
              <div v-if="isUploadingMedia" class="space-y-1.5">
                <div v-if="uploadStatusText" class="text-[11px] font-bold text-indigo-700 flex items-center gap-1.5 animate-pulse">
                  <RefreshCw class="w-3 h-3 animate-spin" />
                  <span>{{ uploadStatusText }}</span>
                </div>
                <div class="w-full bg-slate-100 rounded-full h-1.5 overflow-hidden">
                  <div
                    class="bg-indigo-600 h-1.5 rounded-full transition-all duration-300"
                    :style="{ width: `${mediaUploadProgress}%` }"
                  ></div>
                </div>
              </div>

              <!-- Preview Media yang Diunggah -->
              <div v-if="currentAnnouncement.image_url" class="pt-2 border-t border-slate-100">
                <div class="text-[11px] font-bold text-slate-700 mb-2 flex items-center justify-between">
                  <span>Pratinjau Media Terpasang:</span>
                  <span
                    class="px-2 py-0.5 rounded text-[10px] uppercase font-black"
                    :class="currentAnnouncement.media_type === 'video' ? 'bg-amber-100 text-amber-800' : 'bg-emerald-100 text-emerald-800'"
                  >
                    {{ currentAnnouncement.media_type === 'video' ? (isGDriveUrl(currentAnnouncement.image_url) ? 'Video (Google Drive)' : 'Video File') : 'Gambar' }}
                  </span>
                </div>

                <!-- Video Preview (Google Drive Iframe or Native Video) -->
                <div v-if="currentAnnouncement.media_type === 'video'" class="relative rounded-xl overflow-hidden bg-black max-w-md shadow-md">
                  <iframe
                    v-if="isGDriveUrl(currentAnnouncement.image_url)"
                    :src="getGDriveEmbedUrl(currentAnnouncement.image_url)"
                    class="w-full h-56 rounded-xl border-0"
                    allow="autoplay; fullscreen"
                  ></iframe>
                  <video
                    v-else
                    :src="getMediaUrl(currentAnnouncement.image_url)"
                    controls
                    preload="metadata"
                    class="w-full max-h-48 object-contain"
                  ></video>
                  <button
                    type="button"
                    @click="removeUploadedMedia(currentAnnouncement.image_url)"
                    class="absolute top-2 right-2 bg-red-600/90 hover:bg-red-700 text-white p-1.5 rounded-lg text-xs transition z-10"
                    title="Hapus Video"
                  >
                    <Trash2 class="w-3.5 h-3.5" />
                  </button>
                </div>

                <!-- Image Preview (Single / Multi) -->
                <div v-else class="flex flex-wrap gap-2">
                  <div class="relative w-20 h-20 rounded-xl overflow-hidden border border-slate-200 group bg-slate-50">
                    <img :src="getMediaUrl(currentAnnouncement.image_url)" class="w-full h-full object-cover" />
                    <button
                      type="button"
                      @click="removeUploadedMedia(currentAnnouncement.image_url)"
                      class="absolute top-1 right-1 bg-red-600/90 hover:bg-red-700 text-white p-0.5 rounded-md opacity-90 group-hover:opacity-100 transition"
                      title="Hapus Gambar"
                    >
                      <Trash2 class="w-3 h-3" />
                    </button>
                  </div>

                  <!-- Additional Images from media_urls_text -->
                  <template v-for="(imgUrl, idx) in currentAnnouncement.media_urls_text.split('\n').filter(Boolean)" :key="idx">
                    <div
                      v-if="imgUrl !== currentAnnouncement.image_url"
                      class="relative w-20 h-20 rounded-xl overflow-hidden border border-slate-200 group bg-slate-50"
                    >
                      <img :src="imgUrl" class="w-full h-full object-cover" />
                      <button
                        type="button"
                        @click="removeUploadedMedia(imgUrl)"
                        class="absolute top-1 right-1 bg-red-600/90 hover:bg-red-700 text-white p-0.5 rounded-md opacity-90 group-hover:opacity-100 transition"
                        title="Hapus Gambar"
                      >
                        <Trash2 class="w-3 h-3" />
                      </button>
                    </div>
                  </template>
                </div>
              </div>
            </div>

            <!-- URL Gambar / Video Manual (Opsional / Advanced) -->
            <div class="space-y-1.5">
              <Label class="text-[11px] font-bold text-indigo-900">
                URL Gambar Utama / File Video (Terisi otomatis saat unggah berkas)
              </Label>
              <Input
                v-model="currentAnnouncement.image_url"
                placeholder="https://... atau /uploads/announcements/..."
                class="h-9 text-xs rounded-xl bg-white border-indigo-200"
              />
            </div>

            <!-- URL Multi Gambar (Carousel) -->
            <div class="space-y-1.5">
              <Label class="text-[11px] font-bold text-indigo-900">
                URL Gambar Tambahan / Multi-Foto Carousel (Opsional, 1 URL per baris)
              </Label>
              <textarea
                v-model="currentAnnouncement.media_urls_text"
                placeholder="https://example.com/foto1.jpg&#10;https://example.com/foto2.jpg"
                rows="2"
                class="flex w-full rounded-xl border border-indigo-200 bg-white px-3 py-2 text-xs placeholder:text-gray-400 focus:outline-none focus:ring-2 focus:ring-primary transition-all resize-none"
              ></textarea>
            </div>

            <!-- Tautan Sosmed Eksternal -->
            <div class="space-y-1.5">
              <Label class="text-[11px] font-bold text-indigo-900 flex items-center gap-1">
                <Share2 class="w-3.5 h-3.5 text-purple-600" />
                Tautan Postingan Asli Medsos (Instagram / TikTok / YouTube)
              </Label>
              <Input
                v-model="currentAnnouncement.external_url"
                placeholder="https://instagram.com/p/... atau https://tiktok.com/@contoh/video/..."
                class="h-9 text-xs rounded-xl bg-white border-indigo-200"
              />
            </div>
          </div>

          <!-- TARGET PENERIMA (TARGETING) -->
          <div class="space-y-3 p-4 rounded-2xl bg-slate-50 border border-slate-200/80">
            <div class="flex items-center justify-between">
              <Label class="text-xs font-black text-gray-800 uppercase tracking-wider flex items-center gap-1.5">
                <Users class="w-4 h-4 text-primary" /> Target Penerima Pengumuman
              </Label>
            </div>

            <!-- Targeting Mode Switcher -->
            <div class="grid grid-cols-3 gap-2">
              <button
                type="button"
                @click="currentAnnouncement.target_type = 'individual'"
                :class="[
                  'py-2 px-3 rounded-xl text-xs font-bold border transition-all flex items-center justify-center gap-1.5',
                  currentAnnouncement.target_type === 'individual'
                    ? 'bg-primary text-white border-primary shadow-sm'
                    : 'bg-white text-slate-600 border-slate-200 hover:bg-slate-100',
                ]"
              >
                <UserCheck class="w-3.5 h-3.5" /> Karyawan Spesifik
              </button>

              <button
                type="button"
                @click="currentAnnouncement.target_type = 'role'"
                :class="[
                  'py-2 px-3 rounded-xl text-xs font-bold border transition-all flex items-center justify-center gap-1.5',
                  currentAnnouncement.target_type === 'role'
                    ? 'bg-primary text-white border-primary shadow-sm'
                    : 'bg-white text-slate-600 border-slate-200 hover:bg-slate-100',
                ]"
              >
                <Briefcase class="w-3.5 h-3.5" /> Per Jabatan/Role
              </button>

              <button
                type="button"
                @click="currentAnnouncement.target_type = 'broadcast'"
                :class="[
                  'py-2 px-3 rounded-xl text-xs font-bold border transition-all flex items-center justify-center gap-1.5',
                  currentAnnouncement.target_type === 'broadcast'
                    ? 'bg-primary text-white border-primary shadow-sm'
                    : 'bg-white text-slate-600 border-slate-200 hover:bg-slate-100',
                ]"
              >
                <Users class="w-3.5 h-3.5" /> Semua Karyawan
              </button>
            </div>

            <!-- Individual Employee Selector -->
            <div v-if="currentAnnouncement.target_type === 'individual'" class="space-y-3 pt-2">
              <div class="flex items-center justify-between gap-2">
                <div class="relative flex-1">
                  <Search class="w-4 h-4 text-slate-400 absolute left-3 top-1/2 -translate-y-1/2" />
                  <Input
                    v-model="employeeSearchQuery"
                    placeholder="Cari nama karyawan, NIK, atau divisi..."
                    class="h-9 pl-9 text-xs rounded-xl bg-white border-slate-200"
                  />
                </div>
                <div class="flex items-center gap-1.5">
                  <Button
                    type="button"
                    variant="outline"
                    size="sm"
                    @click="selectAllFilteredEmployees"
                    class="h-9 text-[11px] font-bold rounded-xl px-2.5"
                  >
                    Pilih Semua
                  </Button>
                  <Button
                    type="button"
                    variant="ghost"
                    size="sm"
                    @click="clearSelectedEmployees"
                    class="h-9 text-[11px] font-bold text-red-600 hover:bg-red-50 rounded-xl px-2.5"
                  >
                    Reset
                  </Button>
                </div>
              </div>

              <!-- Badges Karyawan Terpilih -->
              <div class="flex items-center justify-between text-xs font-bold text-slate-500">
                <span>{{ currentAnnouncement.target_user_ids.length }} Karyawan Dipilih</span>
              </div>

              <!-- Employee Checklist Box -->
              <div class="max-h-48 overflow-y-auto space-y-1.5 pr-1 custom-scrollbar">
                <div
                  v-for="emp in filteredEmployees"
                  :key="emp.id"
                  @click="toggleEmployeeSelection(emp.user_id || emp.user?.id)"
                  :class="[
                    'flex items-center justify-between p-2.5 rounded-xl border cursor-pointer transition-all',
                    currentAnnouncement.target_user_ids.includes(emp.user_id || emp.user?.id)
                      ? 'bg-primary/10 border-primary text-primary font-bold shadow-xs'
                      : 'bg-white border-slate-200 hover:border-slate-300 text-slate-700',
                  ]"
                >
                  <div class="flex items-center gap-2.5">
                    <div
                      class="w-7 h-7 rounded-full flex items-center justify-center font-bold text-xs"
                      :class="currentAnnouncement.target_user_ids.includes(emp.user_id || emp.user?.id) ? 'bg-primary text-white' : 'bg-slate-200 text-slate-600'"
                    >
                      {{ (emp.first_name || 'U').charAt(0) }}
                    </div>
                    <div class="flex flex-col">
                      <span class="text-xs leading-tight">
                        {{ emp.first_name }} {{ emp.last_name || '' }}
                      </span>
                      <span class="text-[10px] text-slate-400 font-normal">
                        NIK: {{ emp.employee_id_number || '-' }} • {{ emp.job_position?.title || emp.department?.name || 'Staff' }}
                      </span>
                    </div>
                  </div>
                  <div
                    class="w-5 h-5 rounded-md border flex items-center justify-center"
                    :class="currentAnnouncement.target_user_ids.includes(emp.user_id || emp.user?.id) ? 'bg-primary border-primary text-white' : 'border-slate-300'"
                  >
                    <Check v-if="currentAnnouncement.target_user_ids.includes(emp.user_id || emp.user?.id)" class="w-3.5 h-3.5 stroke-[3]" />
                  </div>
                </div>
              </div>
            </div>

            <!-- Role & Jabatan Selector (Database-Driven) -->
            <div v-else-if="currentAnnouncement.target_type === 'role'" class="space-y-3 pt-2">
              <!-- Sub-tabs & Search Header -->
              <div class="flex flex-col sm:flex-row items-stretch sm:items-center justify-between gap-2">
                <div class="flex items-center gap-1 p-1 bg-slate-200/70 rounded-xl">
                  <button
                    type="button"
                    @click="roleTargetSubTab = 'positions'"
                    :class="[
                      'px-3 py-1 rounded-lg text-xs font-bold transition-all flex items-center gap-1.5',
                      roleTargetSubTab === 'positions'
                        ? 'bg-white text-primary shadow-xs'
                        : 'text-slate-600 hover:text-slate-900',
                    ]"
                  >
                    <Briefcase class="w-3.5 h-3.5" />
                    Jabatan Karyawan ({{ dynamicJobPositions.length }})
                  </button>
                  <button
                    type="button"
                    @click="roleTargetSubTab = 'roles'"
                    :class="[
                      'px-3 py-1 rounded-lg text-xs font-bold transition-all flex items-center gap-1.5',
                      roleTargetSubTab === 'roles'
                        ? 'bg-white text-primary shadow-xs'
                        : 'text-slate-600 hover:text-slate-900',
                    ]"
                  >
                    <Users class="w-3.5 h-3.5" />
                    Role Akses ({{ dynamicAccessRoles.length }})
                  </button>
                </div>

                <div class="flex items-center gap-1.5">
                  <div class="relative flex-1 sm:w-48">
                    <Search class="w-3.5 h-3.5 text-slate-400 absolute left-2.5 top-1/2 -translate-y-1/2" />
                    <Input
                      v-model="roleSearchQuery"
                      placeholder="Cari jabatan/role..."
                      class="h-8 pl-8 text-xs rounded-xl bg-white border-slate-200"
                    />
                  </div>
                  <Button
                    type="button"
                    variant="outline"
                    size="sm"
                    @click="selectAllFilteredRoles"
                    class="h-8 text-[11px] font-bold rounded-xl px-2.5"
                  >
                    Pilih Semua
                  </Button>
                  <Button
                    type="button"
                    variant="ghost"
                    size="sm"
                    @click="clearSelectedRoles"
                    class="h-8 text-[11px] font-bold text-red-600 hover:bg-red-50 rounded-xl px-2.5"
                  >
                    Reset
                  </Button>
                </div>
              </div>

              <!-- Selected Counter -->
              <div class="flex items-center justify-between text-xs font-bold text-slate-500 px-0.5">
                <span>
                  {{ currentAnnouncement.target_roles.length }} Target Dipilih
                  <span v-if="selectedRolesCountEmployees > 0" class="text-primary font-black ml-1">
                    (Mencakup ±{{ selectedRolesCountEmployees }} Karyawan Aktif)
                  </span>
                </span>
              </div>

              <!-- Job Positions Grid -->
              <div
                v-if="roleTargetSubTab === 'positions'"
                class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-2 max-h-48 overflow-y-auto pr-1 custom-scrollbar"
              >
                <div
                  v-if="filteredJobPositions.length === 0"
                  class="col-span-full py-8 text-center text-xs text-slate-400"
                >
                  Tidak ada jabatan yang sesuai pencarian.
                </div>
                <div
                  v-for="pos in filteredJobPositions"
                  :key="pos.value"
                  @click="toggleRoleSelection(pos.value)"
                  :class="[
                    'flex items-center justify-between p-2.5 rounded-xl border text-xs cursor-pointer transition-all',
                    currentAnnouncement.target_roles.includes(pos.value)
                      ? 'bg-primary/10 border-primary text-primary font-bold'
                      : 'bg-white border-slate-200 hover:border-slate-300 text-slate-700',
                  ]"
                >
                  <div class="flex items-center gap-2 min-w-0">
                    <div
                      class="w-4 h-4 rounded border flex items-center justify-center shrink-0"
                      :class="currentAnnouncement.target_roles.includes(pos.value) ? 'bg-primary border-primary text-white' : 'border-slate-300'"
                    >
                      <Check v-if="currentAnnouncement.target_roles.includes(pos.value)" class="w-3 h-3 stroke-[3]" />
                    </div>
                    <span class="truncate">{{ pos.label }}</span>
                  </div>
                  <span
                    class="px-1.5 py-0.5 rounded-md text-[10px] font-black shrink-0 ml-1.5"
                    :class="pos.count > 0 ? 'bg-primary/10 text-primary' : 'bg-slate-100 text-slate-400'"
                  >
                    {{ pos.count }}
                  </span>
                </div>
              </div>

              <!-- Access Roles Grid -->
              <div
                v-else
                class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-2 max-h-48 overflow-y-auto pr-1 custom-scrollbar"
              >
                <div
                  v-if="filteredAccessRoles.length === 0"
                  class="col-span-full py-8 text-center text-xs text-slate-400"
                >
                  Tidak ada role akses yang sesuai pencarian.
                </div>
                <div
                  v-for="role in filteredAccessRoles"
                  :key="role.value"
                  @click="toggleRoleSelection(role.value)"
                  :class="[
                    'flex items-center justify-between p-2.5 rounded-xl border text-xs cursor-pointer transition-all',
                    currentAnnouncement.target_roles.includes(role.value)
                      ? 'bg-primary/10 border-primary text-primary font-bold'
                      : 'bg-white border-slate-200 hover:border-slate-300 text-slate-700',
                  ]"
                >
                  <div class="flex items-center gap-2 min-w-0">
                    <div
                      class="w-4 h-4 rounded border flex items-center justify-center shrink-0"
                      :class="currentAnnouncement.target_roles.includes(role.value) ? 'bg-primary border-primary text-white' : 'border-slate-300'"
                    >
                      <Check v-if="currentAnnouncement.target_roles.includes(role.value)" class="w-3 h-3 stroke-[3]" />
                    </div>
                    <span class="truncate">{{ role.label }}</span>
                  </div>
                  <span
                    class="px-1.5 py-0.5 rounded-md text-[10px] font-black shrink-0 ml-1.5"
                    :class="role.count > 0 ? 'bg-primary/10 text-primary' : 'bg-slate-100 text-slate-400'"
                  >
                    {{ role.count }}
                  </span>
                </div>
              </div>
            </div>

            <!-- Broadcast notice -->
            <div v-else-if="currentAnnouncement.target_type === 'broadcast'" class="p-3 bg-indigo-50/70 border border-indigo-100 rounded-xl text-xs text-indigo-900">
              Pengumuman ini akan disiarkan ke <strong>seluruh karyawan aktif</strong> di perusahaan.
            </div>
          </div>

          <!-- POP-UP KONFIRMASI SETTING -->
          <div class="p-4 rounded-2xl bg-red-50/60 border border-red-200/80 space-y-3">
            <div
              @click="currentAnnouncement.is_popup = !currentAnnouncement.is_popup"
              class="flex items-center justify-between cursor-pointer select-none"
            >
              <div>
                <span class="text-xs font-black text-red-900 flex items-center gap-1.5">
                  <AlertTriangle class="w-4 h-4 text-red-600" />
                  Wajibkan Modal Pop-up di Mobile App (Direct Confirmation)
                </span>
                <p class="text-[11px] text-red-700/80 mt-0.5">
                  Jika aktif, karyawan akan otomatis melihat modal pop-up blocking dan harus menekan tombol "OK / KONFIRMASI".
                </p>
              </div>
              <button
                type="button"
                class="relative inline-flex h-6 w-11 shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none ml-4"
                :class="currentAnnouncement.is_popup ? 'bg-red-600' : 'bg-slate-300'"
              >
                <span
                  class="pointer-events-none inline-block h-5 w-5 transform rounded-full bg-white shadow-md ring-0 transition duration-200 ease-in-out"
                  :class="currentAnnouncement.is_popup ? 'translate-x-5' : 'translate-x-0'"
                />
              </button>
            </div>

            <div v-if="currentAnnouncement.is_popup" class="pt-2 border-t border-red-200/60 grid grid-cols-1 sm:grid-cols-2 gap-3">
              <div class="space-y-1">
                <Label class="text-[11px] font-bold text-red-900">Batas Waktu Pop-up (Opsional)</Label>
                <Input
                  type="datetime-local"
                  v-model="currentAnnouncement.popup_expires_at"
                  class="h-9 text-xs rounded-xl bg-white border-red-200"
                />
              </div>
            </div>
          </div>

          <!-- PELACAKAN UNDUHAN & PEMBAGIAN (TRACKING ON/OFF) -->
          <div class="p-4 rounded-2xl bg-sky-50/70 border border-sky-200/80 space-y-3">
            <div class="flex items-center justify-between">
              <span class="text-xs font-black text-sky-950 flex items-center gap-1.5 uppercase tracking-wider">
                <Sparkles class="w-4 h-4 text-sky-600" />
                Opsi Pelacakan Interaksi Karyawan (Audit Log)
              </span>
              <span class="text-[10px] font-bold text-sky-700 bg-sky-100 px-2 py-0.5 rounded-md">
                Tracking ON / OFF
              </span>
            </div>

            <!-- Toggle 1: Track Downloads -->
            <div
              @click="currentAnnouncement.track_downloads = !currentAnnouncement.track_downloads"
              class="flex items-center justify-between p-3 rounded-xl bg-white border border-sky-100 cursor-pointer select-none hover:bg-sky-50/40 transition"
            >
              <div>
                <span class="text-xs font-bold text-slate-800 flex items-center gap-1.5">
                  📥 Lacak Unduhan Media (Track Downloads)
                </span>
                <p class="text-[11px] text-slate-500 mt-0.5">
                  Catat nama karyawan dan waktu ketika mengunduh gambar/video pengumuman ini ke ponsel.
                </p>
              </div>
              <button
                type="button"
                class="relative inline-flex h-6 w-11 shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none ml-4"
                :class="currentAnnouncement.track_downloads ? 'bg-sky-600' : 'bg-slate-300'"
              >
                <span
                  class="pointer-events-none inline-block h-5 w-5 transform rounded-full bg-white shadow-md ring-0 transition duration-200 ease-in-out"
                  :class="currentAnnouncement.track_downloads ? 'translate-x-5' : 'translate-x-0'"
                />
              </button>
            </div>

            <!-- Toggle 2: Track Shares -->
            <div
              @click="currentAnnouncement.track_shares = !currentAnnouncement.track_shares"
              class="flex items-center justify-between p-3 rounded-xl bg-white border border-sky-100 cursor-pointer select-none hover:bg-sky-50/40 transition"
            >
              <div>
                <span class="text-xs font-bold text-slate-800 flex items-center gap-1.5">
                  📤 Lacak Pembagian (Track Shares)
                </span>
                <p class="text-[11px] text-slate-500 mt-0.5">
                  Catat nama karyawan dan waktu ketika membagikan pengumuman ini ke WhatsApp / aplikasi lain.
                </p>
              </div>
              <button
                type="button"
                class="relative inline-flex h-6 w-11 shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none ml-4"
                :class="currentAnnouncement.track_shares ? 'bg-sky-600' : 'bg-slate-300'"
              >
                <span
                  class="pointer-events-none inline-block h-5 w-5 transform rounded-full bg-white shadow-md ring-0 transition duration-200 ease-in-out"
                  :class="currentAnnouncement.track_shares ? 'translate-x-5' : 'translate-x-0'"
                />
              </button>
            </div>
          </div>

          <!-- Isi Pengumuman -->
          <div class="space-y-1.5">
            <Label class="text-xs font-bold text-gray-700 uppercase tracking-wider">
              Isi Pengumuman / Deskripsi & Caption <span class="text-red-500">*</span>
            </Label>
            <textarea
              v-model="currentAnnouncement.content"
              placeholder="Tuliskan detail pengumuman, instruksi jam & ruangan meeting, atau keperluan menghadap..."
              rows="5"
              class="flex w-full rounded-xl border border-slate-200 bg-transparent px-3 py-2.5 text-sm placeholder:text-gray-400 focus:outline-none focus:ring-2 focus:ring-primary focus:border-transparent transition-all resize-none"
            ></textarea>
          </div>
        </div>

        <!-- Footer -->
        <DialogFooter class="p-6 md:p-8 pt-4 bg-slate-50/80 border-t border-slate-100 flex items-center justify-end gap-3 shrink-0">
          <Button
            variant="ghost"
            @click="isModalOpen = false"
            :disabled="isSubmitting"
            class="rounded-xl h-11 px-5 font-bold text-slate-500 hover:bg-slate-200"
          >
            BATAL
          </Button>
          <Button
            @click="saveAnnouncement"
            :disabled="isSubmitting"
            class="bg-primary hover:bg-primary/90 text-white rounded-xl h-11 px-6 font-black shadow-lg shadow-primary/20"
          >
            {{ isSubmitting ? 'MENGIRIM NOTIFIKASI FCM...' : isEditMode ? 'SIMPAN PERUBAHAN' : 'KIRIM PENGUMUMAN & FCM' }}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>

    <!-- ─── MODAL TRACKING RIWAYAT KONFIRMASI ──────────────────────── -->
    <Dialog v-model:open="isTrackingModalOpen">
      <DialogContent class="sm:max-w-3xl rounded-3xl p-0 overflow-hidden border-none shadow-2xl max-h-[92vh] flex flex-col">
        <!-- Header -->
        <div class="bg-slate-900 p-6 md:p-8 text-white relative shrink-0">
          <div class="flex items-start justify-between gap-4">
            <div>
              <div class="flex items-center gap-2">
                <span class="px-2.5 py-0.5 rounded-full text-[10px] font-black uppercase tracking-wider bg-primary text-white">
                  {{ trackingSummary?.category || 'DIRECT' }}
                </span>
                <span class="text-xs text-slate-400">
                  Diterbitkan {{ formatDate(trackingSummary?.created_at) }}
                </span>
              </div>
              <h2 class="text-lg md:text-xl font-black text-white mt-1 line-clamp-1">
                {{ trackingSummary?.title || 'Data Historis Konfirmasi Pengumuman' }}
              </h2>
            </div>
            <button
              @click="isTrackingModalOpen = false"
              class="p-2 rounded-xl bg-white/10 hover:bg-white/20 text-white transition-all"
            >
              <X class="w-5 h-5" />
            </button>
          </div>

          <!-- Summary Metric Cards -->
          <div class="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-5 gap-3 mt-6">
            <div class="bg-white/10 rounded-2xl p-3 border border-white/10">
              <span class="text-[10px] font-extrabold uppercase text-slate-300 tracking-wider">Total Target</span>
              <div class="text-xl md:text-2xl font-black text-white mt-0.5">
                {{ trackingSummary?.total_targets || 0 }}
              </div>
            </div>

            <div class="bg-emerald-500/20 rounded-2xl p-3 border border-emerald-500/30">
              <span class="text-[10px] font-extrabold uppercase text-emerald-300 tracking-wider">Konfirmasi</span>
              <div class="text-xl md:text-2xl font-black text-emerald-400 mt-0.5">
                {{ trackingSummary?.confirmed_count || 0 }}
                <span class="text-xs font-bold opacity-80">({{ Math.round(trackingSummary?.confirmation_rate || 0) }}%)</span>
              </div>
            </div>

            <div class="bg-amber-500/20 rounded-2xl p-3 border border-amber-500/30">
              <span class="text-[10px] font-extrabold uppercase text-amber-300 tracking-wider">Belum Konfirmasi</span>
              <div class="text-xl md:text-2xl font-black text-amber-400 mt-0.5">
                {{ trackingSummary?.pending_count || 0 }}
              </div>
            </div>

            <!-- Metric Download -->
            <div class="bg-sky-500/20 rounded-2xl p-3 border border-sky-500/30">
              <span class="text-[10px] font-extrabold uppercase text-sky-300 tracking-wider flex items-center gap-1">
                📥 Diunduh
              </span>
              <div class="text-xl md:text-2xl font-black text-sky-400 mt-0.5">
                {{ trackingSummary?.downloaded_count || 0 }}
              </div>
            </div>

            <!-- Metric Share -->
            <div class="bg-purple-500/20 rounded-2xl p-3 border border-purple-500/30">
              <span class="text-[10px] font-extrabold uppercase text-purple-300 tracking-wider flex items-center gap-1">
                📤 Dibagikan
              </span>
              <div class="text-xl md:text-2xl font-black text-purple-400 mt-0.5">
                {{ trackingSummary?.shared_count || 0 }}
              </div>
            </div>
          </div>
        </div>

        <!-- Filter & Search Controls -->
        <div class="p-4 md:px-8 bg-slate-50 border-b border-slate-200 flex flex-col sm:flex-row items-center justify-between gap-3 shrink-0">
          <!-- Filter Buttons & Mass FCM Resend -->
          <div class="flex flex-wrap items-center gap-1.5 w-full sm:w-auto">
            <button
              @click="trackingFilter = 'all'"
              :class="[
                'px-3 py-1.5 rounded-xl text-xs font-bold transition-all',
                trackingFilter === 'all'
                  ? 'bg-slate-900 text-white'
                  : 'bg-white text-slate-600 border border-slate-200 hover:bg-slate-100',
              ]"
            >
              Semua ({{ trackingSummary?.total_targets || 0 }})
            </button>
            <button
              @click="trackingFilter = 'confirmed'"
              :class="[
                'px-3 py-1.5 rounded-xl text-xs font-bold transition-all',
                trackingFilter === 'confirmed'
                  ? 'bg-emerald-600 text-white'
                  : 'bg-white text-slate-600 border border-slate-200 hover:bg-slate-100',
              ]"
            >
              Sudah Konfirmasi ({{ trackingSummary?.confirmed_count || 0 }})
            </button>
            <button
              @click="trackingFilter = 'pending'"
              :class="[
                'px-3 py-1.5 rounded-xl text-xs font-bold transition-all',
                trackingFilter === 'pending'
                  ? 'bg-amber-600 text-white'
                  : 'bg-white text-slate-600 border border-slate-200 hover:bg-slate-100',
              ]"
            >
              Belum Konfirmasi ({{ trackingSummary?.pending_count || 0 }})
            </button>

            <!-- Mass Reminder Button -->
            <Button
              v-if="(trackingSummary?.pending_count || 0) > 0"
              @click="resendFCM(trackingSummary.announcement_id, 'pending_only')"
              :disabled="isResendingFCM[`${trackingSummary.announcement_id}_pending_only`]"
              size="sm"
              class="h-8 px-3 ml-1 bg-amber-600 hover:bg-amber-700 text-white rounded-xl text-xs font-bold gap-1.5 shadow-sm"
            >
              <RefreshCw v-if="isResendingFCM[`${trackingSummary.announcement_id}_pending_only`]" class="w-3.5 h-3.5 animate-spin" />
              <BellRing v-else class="w-3.5 h-3.5" />
              Ingatkan Semua Belum Konfirmasi ({{ trackingSummary.pending_count }})
            </Button>
          </div>

          <!-- Search Input -->
          <div class="relative w-full sm:w-64">
            <Search class="w-4 h-4 text-slate-400 absolute left-3 top-1/2 -translate-y-1/2" />
            <Input
              v-model="trackingSearch"
              placeholder="Cari nama / NIK..."
              class="h-9 pl-9 text-xs rounded-xl bg-white border-slate-200"
            />
          </div>
        </div>

        <!-- Recipients Table List -->
        <div class="p-4 md:p-8 bg-white flex-1 overflow-y-auto custom-scrollbar">
          <div v-if="isLoadingTracking" class="py-16 text-center text-slate-400">
            <RefreshCw class="w-6 h-6 animate-spin mx-auto text-primary mb-2" />
            Memuat data historis konfirmasi...
          </div>
          <div v-else-if="filteredRecipients.length === 0" class="py-16 text-center text-slate-400">
            Tidak ada data karyawan yang sesuai filter.
          </div>
          <div v-else class="space-y-2">
            <div
              v-for="rec in filteredRecipients"
              :key="rec.user_id"
              class="flex flex-col sm:flex-row items-start sm:items-center justify-between p-3.5 rounded-2xl border border-slate-100 hover:border-slate-200 bg-slate-50/50 hover:bg-slate-50 gap-3 transition-all"
            >
              <div class="flex items-center gap-3">
                <div
                  class="w-10 h-10 rounded-full flex items-center justify-center font-black text-sm shrink-0"
                  :class="rec.is_confirmed ? 'bg-emerald-100 text-emerald-700' : 'bg-amber-100 text-amber-700'"
                >
                  {{ (rec.employee_name || 'U').charAt(0) }}
                </div>
                <div>
                  <div class="flex items-center gap-2 flex-wrap">
                    <span class="font-bold text-gray-900 text-sm">
                      {{ rec.employee_name }}
                    </span>
                    <span class="text-[10px] uppercase px-2 py-0.5 rounded-md font-extrabold bg-slate-200 text-slate-700">
                      {{ rec.role || 'Karyawan' }}
                    </span>
                  </div>
                  <div class="text-xs text-slate-500 mt-0.5">
                    NIK: {{ rec.employee_id_number || '-' }} • {{ rec.job_title || rec.department_name || '-' }}
                  </div>
                </div>
              </div>

              <!-- Status Kolom: Konfirmasi, Download, Share -->
              <div class="flex flex-wrap sm:flex-nowrap items-center gap-3 w-full sm:w-auto justify-between sm:justify-end">
                <!-- Activity Badges (Download & Share) -->
                <div class="flex items-center gap-1.5">
                  <!-- Download Status -->
                  <span
                    v-if="trackingSummary?.track_downloads"
                    class="inline-flex items-center gap-1 px-2.5 py-1 rounded-lg text-[10px] font-bold"
                    :class="rec.is_downloaded ? 'bg-sky-100 text-sky-800 border border-sky-200' : 'bg-slate-100 text-slate-400'"
                    :title="rec.is_downloaded ? `Diunduh pada ${formatDate(rec.downloaded_at)}` : 'Belum diunduh'"
                  >
                    📥 {{ rec.is_downloaded ? 'Diunduh' : 'Belum Unduh' }}
                  </span>

                  <!-- Share Status -->
                  <span
                    v-if="trackingSummary?.track_shares"
                    class="inline-flex items-center gap-1 px-2.5 py-1 rounded-lg text-[10px] font-bold"
                    :class="rec.is_shared ? 'bg-purple-100 text-purple-800 border border-purple-200' : 'bg-slate-100 text-slate-400'"
                    :title="rec.is_shared ? `Dibagikan pada ${formatDate(rec.shared_at)}` : 'Belum dibagikan'"
                  >
                    📤 {{ rec.is_shared ? 'Dibagikan' : 'Belum Share' }}
                  </span>
                </div>

                <!-- Confirmation Status -->
                <div class="text-right flex flex-col items-end gap-0.5 shrink-0">
                  <span
                    v-if="rec.is_confirmed"
                    class="inline-flex items-center gap-1.5 px-3 py-1 rounded-full text-xs font-black bg-emerald-100 text-emerald-800"
                  >
                    <CheckCircle2 class="w-3.5 h-3.5 text-emerald-600" />
                    Dikonfirmasi
                  </span>
                  <div v-else class="flex items-center gap-2">
                    <span
                      class="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-xs font-black bg-amber-100 text-amber-800"
                    >
                      <Clock class="w-3.5 h-3.5 text-amber-600" />
                      Menunggu
                    </span>
                    <Button
                      @click="resendFCM(trackingSummary.announcement_id, 'single', rec.user_id, rec.employee_name)"
                      :disabled="isResendingFCM[`${trackingSummary.announcement_id}_${rec.user_id}`]"
                      variant="outline"
                      size="sm"
                      class="h-7 px-2.5 rounded-lg text-[11px] font-bold gap-1 text-amber-700 border-amber-300 bg-amber-50 hover:bg-amber-100 shadow-none"
                      title="Kirim notifikasi FCM langsung ke HP karyawan ini"
                    >
                      <RefreshCw v-if="isResendingFCM[`${trackingSummary.announcement_id}_${rec.user_id}`]" class="w-3 h-3 animate-spin" />
                      <BellRing v-else class="w-3 h-3" />
                      Ping FCM
                    </Button>
                  </div>
                  <div class="text-[11px] text-slate-400">
                    {{ rec.acknowledged_at ? formatDate(rec.acknowledged_at) : 'Belum buka pop-up' }}
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- Modal Footer -->
        <DialogFooter class="p-4 md:px-8 bg-slate-50 border-t border-slate-100 flex items-center justify-between shrink-0">
          <span class="text-xs text-slate-400">
            Data otomatis ter-update saat karyawan klik konfirmasi pada aplikasi mobile.
          </span>
          <Button
            @click="isTrackingModalOpen = false"
            class="rounded-xl h-10 px-5 font-bold"
          >
            Tutup
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>

    <!-- ─── MODAL DIALOG: PILIHAN KIRIM ULANG FCM ──────────────────────── -->
    <Dialog :open="isResendModalOpen" @update:open="isResendModalOpen = $event">
      <DialogContent class="sm:max-w-md p-0 overflow-hidden rounded-3xl border-0 shadow-2xl">
        <div class="bg-gradient-to-r from-amber-600 to-amber-700 p-6 text-white">
          <div class="flex items-center gap-3">
            <div class="p-2.5 rounded-2xl bg-white/10 backdrop-blur-sm">
              <BellRing class="w-6 h-6 text-amber-200" />
            </div>
            <div>
              <DialogTitle class="text-lg font-black text-white">
                Kirim Ulang Notifikasi FCM
              </DialogTitle>
              <DialogDescription class="text-xs text-amber-100 mt-0.5">
                Kirimkan push notifikasi ke HP karyawan untuk pengumuman ini
              </DialogDescription>
            </div>
          </div>
        </div>

        <div class="p-6 space-y-4 bg-white">
          <div class="p-3.5 rounded-2xl bg-slate-50 border border-slate-200">
            <span class="text-[10px] uppercase tracking-wider font-extrabold text-slate-400">Judul Pengumuman:</span>
            <div class="text-sm font-bold text-gray-900 mt-0.5 line-clamp-2">
              {{ selectedResendAnnouncement?.title }}
            </div>
          </div>

          <div class="space-y-2.5">
            <span class="text-xs font-bold text-gray-700 uppercase tracking-wider">
              Pilih Target Penerima Notifikasi:
            </span>

            <!-- Option 1: Pending Only -->
            <div
              @click="resendFCM(selectedResendAnnouncement?.id, 'pending_only')"
              class="p-4 rounded-2xl border-2 border-amber-200 bg-amber-50/50 hover:bg-amber-100/70 hover:border-amber-400 cursor-pointer transition-all flex items-start gap-3.5 group"
            >
              <div class="p-2 rounded-xl bg-amber-200 text-amber-900 mt-0.5 group-hover:scale-105 transition-transform">
                <Clock class="w-5 h-5" />
              </div>
              <div class="flex-1">
                <div class="flex items-center gap-2">
                  <span class="font-bold text-sm text-amber-950">Hanya yang Belum Konfirmasi</span>
                  <span class="px-2 py-0.5 rounded-full text-[10px] font-black uppercase tracking-wider bg-amber-200 text-amber-900">
                    Rekomendasi
                  </span>
                </div>
                <p class="text-xs text-amber-800/80 mt-1">
                  Mencegah spam ke karyawan yang sudah patuh membaca. Hanya mengirim ke HP karyawan yang belum klik konfirmasi.
                </p>
              </div>
            </div>

            <!-- Option 2: All Targets -->
            <div
              @click="resendFCM(selectedResendAnnouncement?.id, 'all')"
              class="p-4 rounded-2xl border border-slate-200 bg-slate-50/60 hover:bg-slate-100/80 hover:border-slate-300 cursor-pointer transition-all flex items-start gap-3.5 group"
            >
              <div class="p-2 rounded-xl bg-slate-200 text-slate-700 mt-0.5 group-hover:scale-105 transition-transform">
                <Users class="w-5 h-5" />
              </div>
              <div class="flex-1">
                <span class="font-bold text-sm text-gray-900">Kirim ke Seluruh Target Penerima</span>
                <p class="text-xs text-slate-500 mt-1">
                  Menyiarkan ulang notifikasi ke semua karyawan target tanpa memandang status konfirmasi saat ini.
                </p>
              </div>
            </div>
          </div>
        </div>

        <DialogFooter class="p-4 bg-slate-50 border-t border-slate-100 flex justify-end">
          <Button
            @click="isResendModalOpen = false"
            variant="outline"
            class="rounded-xl h-10 px-5 font-bold"
          >
            Batal
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>

    <!-- MODAL KOMENTAR & DISKUSI -->
    <Dialog v-model:open="isCommentsModalOpen">
      <DialogContent class="sm:max-w-[650px] p-0 overflow-hidden border-0 rounded-3xl shadow-2xl bg-white max-h-[90vh] flex flex-col">
        <!-- Header -->
        <DialogHeader class="p-5 bg-gradient-to-r from-blue-900 via-indigo-900 to-slate-900 text-white shrink-0">
          <div class="flex items-center justify-between">
            <div class="flex items-center gap-3">
              <div class="p-2.5 rounded-2xl bg-white/10 text-white backdrop-blur-md">
                <MessageSquare class="w-5 h-5 text-blue-300" />
              </div>
              <div>
                <DialogTitle class="text-base sm:text-lg font-black text-white flex items-center gap-2">
                  Komentar & Diskusi
                  <span class="px-2 py-0.5 rounded-full text-xs font-bold bg-blue-500/30 text-blue-200 border border-blue-400/30">
                    {{ commentsList.length }} Komentar
                  </span>
                </DialogTitle>
                <DialogDescription class="text-xs text-blue-200/80 line-clamp-1 mt-0.5">
                  {{ selectedAnnouncementForComments?.title }}
                </DialogDescription>
              </div>
            </div>
          </div>
        </DialogHeader>

        <!-- Body / Comments List -->
        <div class="flex-1 overflow-y-auto p-5 space-y-4 custom-scrollbar bg-slate-50/50 min-h-[250px] max-h-[420px]">
          <div v-if="isLoadingComments" class="flex flex-col items-center justify-center py-12 text-slate-400 gap-2">
            <RefreshCw class="w-6 h-6 animate-spin text-primary" />
            <span class="text-xs font-bold">Memuat komentar...</span>
          </div>

          <div v-else-if="commentsList.length === 0" class="flex flex-col items-center justify-center py-12 text-center text-slate-400">
            <div class="w-14 h-14 rounded-2xl bg-slate-100 flex items-center justify-center mb-3">
              <MessageSquare class="w-7 h-7 text-slate-300" />
            </div>
            <p class="font-bold text-slate-700 text-sm">Belum Ada Komentar</p>
            <p class="text-xs text-slate-400 mt-0.5">Jadilah yang pertama memberikan respon atau klarifikasi resmi.</p>
          </div>

          <div v-else class="space-y-3">
            <div
              v-for="comment in commentsList"
              :key="comment.id"
              class="p-4 rounded-2xl bg-white border border-slate-100 shadow-sm hover:border-slate-200 transition-all flex items-start gap-3.5 group"
            >
              <!-- Avatar / Initials -->
              <div class="shrink-0 w-10 h-10 rounded-xl overflow-hidden border border-slate-200 bg-gradient-to-br from-indigo-500 to-purple-600 text-white flex items-center justify-center font-bold text-xs shadow-sm">
                <img
                  v-if="comment.user_avatar"
                  :src="resolvePhotoUrl(comment.user_avatar)"
                  class="w-full h-full object-cover"
                  alt="Avatar"
                />
                <span v-else>
                  {{ (comment.user_name || 'U').slice(0, 2).toUpperCase() }}
                </span>
              </div>

              <!-- Comment Body -->
              <div class="flex-1 min-w-0">
                <div class="flex items-center justify-between gap-2 flex-wrap">
                  <div class="flex items-center gap-2">
                    <span class="font-black text-sm text-slate-900">
                      {{ comment.user_name || 'Karyawan' }}
                    </span>
                    <span
                      v-if="comment.user_role"
                      class="px-2 py-0.5 rounded-md text-[10px] font-extrabold uppercase tracking-wider bg-slate-100 text-slate-600 border border-slate-200"
                    >
                      {{ comment.user_role }}
                    </span>
                  </div>
                  <span class="text-[11px] text-slate-400 font-medium">
                    {{ formatDate(comment.created_at) }}
                  </span>
                </div>

                <div class="text-xs sm:text-sm text-slate-700 mt-1.5 leading-relaxed whitespace-pre-wrap break-words bg-slate-50/70 p-2.5 rounded-xl border border-slate-100">
                  {{ comment.content }}
                </div>
              </div>

              <!-- Delete Action for Admin -->
              <button
                @click="deleteCommentItem(comment.id)"
                :disabled="isDeletingComment[comment.id]"
                class="opacity-60 group-hover:opacity-100 hover:bg-red-50 p-1.5 rounded-lg text-slate-400 hover:text-red-600 transition-all shrink-0"
                title="Hapus / Moderasi Komentar"
              >
                <RefreshCw v-if="isDeletingComment[comment.id]" class="w-4 h-4 animate-spin text-red-500" />
                <Trash2 v-else class="w-4 h-4" />
              </button>
            </div>
          </div>
        </div>

        <!-- Footer / Input Form -->
        <div class="p-4 bg-white border-t border-slate-100 shrink-0">
          <form @submit.prevent="submitComment" class="flex items-center gap-2">
            <Input
              v-model="newCommentText"
              placeholder="Tulis balasan atau klarifikasi resmi admin..."
              class="flex-1 rounded-xl h-11 px-4 text-xs sm:text-sm border-slate-200 bg-slate-50/50 focus:bg-white"
              :disabled="isSubmittingComment"
            />
            <Button
              type="submit"
              :disabled="!newCommentText.trim() || isSubmittingComment"
              class="rounded-xl h-11 px-5 font-bold gap-1.5 shrink-0 bg-primary hover:bg-primary/90 text-white shadow-md shadow-primary/20 text-xs sm:text-sm"
            >
              <RefreshCw v-if="isSubmittingComment" class="w-4 h-4 animate-spin" />
              <Send v-else class="w-4 h-4" />
              Kirim
            </Button>
          </form>
        </div>
      </DialogContent>
    </Dialog>

    <!-- ─── MODAL PRATINJAU VIDEO PENGUMUMAN ───────────────────────────── -->
    <Dialog v-model:open="isPreviewVideoModalOpen">
      <DialogContent class="sm:max-w-2xl p-0 overflow-hidden border-0 rounded-3xl shadow-2xl bg-black flex flex-col text-white">
        <!-- Header -->
        <div class="px-5 py-4 bg-slate-900 flex items-center justify-between border-b border-white/10 shrink-0">
          <div class="flex items-center gap-2.5 min-w-0 pr-3">
            <div class="p-1.5 rounded-lg bg-indigo-500/20 text-indigo-400 shrink-0">
              <Film class="w-4 h-4" />
            </div>
            <h3 class="font-bold text-sm text-white truncate">
              {{ selectedVideoToPreview?.title || 'Pratinjau Video' }}
            </h3>
          </div>
          <button
            @click="isPreviewVideoModalOpen = false"
            class="p-1.5 rounded-xl bg-white/10 hover:bg-white/20 text-white transition-all shrink-0"
          >
            <X class="w-4 h-4" />
          </button>
        </div>

        <!-- Video Player Body -->
        <div class="bg-black flex items-center justify-center p-2 min-h-[320px] max-h-[70vh]">
          <iframe
            v-if="selectedVideoToPreview?.url && isGDriveUrl(selectedVideoToPreview.url)"
            :src="getGDriveEmbedUrl(selectedVideoToPreview.url)"
            class="w-full h-80 rounded-xl border-0"
            allow="autoplay; fullscreen"
          ></iframe>
          <video
            v-else-if="selectedVideoToPreview?.url"
            :src="getMediaUrl(selectedVideoToPreview.url)"
            controls
            autoplay
            playsinline
            class="w-full max-h-[65vh] rounded-xl object-contain bg-black"
          ></video>
        </div>

        <!-- Footer / Action Link -->
        <div class="px-5 py-3 bg-slate-900 border-t border-white/10 flex items-center justify-between gap-3 text-xs shrink-0">
          <span class="text-slate-400 text-[11px] truncate">
            {{ selectedVideoToPreview?.url }}
          </span>
          <a
            v-if="selectedVideoToPreview?.url"
            :href="getMediaUrl(selectedVideoToPreview.url)"
            target="_blank"
            download
            class="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-white/10 hover:bg-white/20 text-white font-bold transition text-xs shrink-0"
          >
            <ExternalLink class="w-3.5 h-3.5" />
            Buka File Langsung
          </a>
        </div>
      </DialogContent>
    </Dialog>
  </div>
</template>

<style scoped>
.custom-scrollbar::-webkit-scrollbar {
  width: 6px;
  height: 6px;
}
.custom-scrollbar::-webkit-scrollbar-track {
  background: #f1f5f9;
  border-radius: 8px;
}
.custom-scrollbar::-webkit-scrollbar-thumb {
  background: #cbd5e1;
  border-radius: 8px;
}
.custom-scrollbar::-webkit-scrollbar-thumb:hover {
  background: #94a3b8;
}
</style>
