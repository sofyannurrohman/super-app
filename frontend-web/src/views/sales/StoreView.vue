<script setup lang="ts">
import { ref, onMounted, computed, watch, nextTick } from 'vue'
import { useRouter } from 'vue-router'
import apiClient from '@/api/axios'
import { regionApi } from '@/api/region'
import { toast } from 'vue-sonner'
import { onClickOutside } from '@vueuse/core'

const router = useRouter()
import {
  Plus,
  Search,
  Store,
  User,
  Phone,
  MapPin,
  Pencil,
  Trash2,
  Save,
  X,
  CheckCircle2,
  AlertCircle,
  MoreVertical,
  Navigation,
  Globe,
  ImagePlus,
  Banknote,
  AlertTriangle,
  History,
  Package,
  FileText,
  Calendar,
  Activity,
  Info,
  ChevronDown,
  FileDown,
  Loader2,
  ShieldCheck,
  LocateFixed,
  Camera,
  ClipboardList,
  MessageSquare,
  Send,
  RefreshCw,
  Clock,
  ShoppingBag,
  Sparkles,
  TrendingUp,
  Check
} from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import { useMasterDataStore } from '@/stores/masterData'

const masterDataStore = useMasterDataStore()
import L from 'leaflet'
import 'leaflet/dist/leaflet.css'
import 'leaflet-defaulticon-compatibility'
import 'leaflet-defaulticon-compatibility/dist/leaflet-defaulticon-compatibility.css'

import markerIcon2x from 'leaflet/dist/images/marker-icon-2x.png'
import markerIcon from 'leaflet/dist/images/marker-icon.png'
import markerShadow from 'leaflet/dist/images/marker-shadow.png'

delete (L.Icon.Default.prototype as any)._getIconUrl
L.Icon.Default.mergeOptions({
  iconRetinaUrl: markerIcon2x,
  iconUrl: markerIcon,
  shadowUrl: markerShadow
})
import StoreImportModal from '@/components/sales/StoreImportModal.vue'
import { Upload } from 'lucide-vue-next'


// Audit States
const showAuditModal = ref(false)
const isSubmittingAudit = ref(false)
const selectedStoreForAudit = ref<any>(null)
const auditForm = ref({
  status: 'VERIFIED_REAL',
  notes: '',
  photo_proof_url: ''
})

const auditPhotoFile = ref<File | null>(null)
const auditPhotoPreview = ref<string | null>(null)

const handleAuditPhotoChange = (e: Event) => {
  const file = (e.target as HTMLInputElement).files?.[0]
  if (!file) return
  auditPhotoFile.value = file
  const reader = new FileReader()
  reader.onload = (e) => {
    auditPhotoPreview.value = e.target?.result as string
  }
  reader.readAsDataURL(file)
}

const openAuditModal = (store: any) => {
  selectedStoreForAudit.value = store
  auditForm.value = {
    status: store.verification_status || 'VERIFIED_REAL',
    notes: '',
    photo_proof_url: ''
  }
  auditPhotoFile.value = null
  auditPhotoPreview.value = null
  showAuditModal.value = true
}

const submitAudit = async () => {
  if (!auditForm.value.notes) {
    toast.error('Catatan audit wajib diisi')
    return
  }

  isSubmittingAudit.value = true
  try {
    const formData = new FormData()
    formData.append('status', auditForm.value.status)
    formData.append('notes', auditForm.value.notes)
    if (auditPhotoFile.value) {
      formData.append('photo', auditPhotoFile.value)
    }

    await apiClient.post(`/stores/${selectedStoreForAudit.value.id}/audit`, formData, {
      headers: { 'Content-Type': 'multipart/form-data' }
    })
    
    toast.success('Status verifikasi toko berhasil disimpan')
    showAuditModal.value = false
    await fetchStores()
  } catch (error: any) {
    console.error('Error submitting audit:', error)
    toast.error(error.response?.data?.error || 'Gagal menyimpan verifikasi toko')
  } finally {
    isSubmittingAudit.value = false
  }
}

// === Webcam Capture Logic ===
const showWebcamModal = ref(false)
const videoRef = ref<HTMLVideoElement | null>(null)
const mediaStream = ref<MediaStream | null>(null)
const webcamTarget = ref<'audit' | 'store' | null>(null)

const openWebcam = async (target: 'audit' | 'store') => {
  webcamTarget.value = target
  showWebcamModal.value = true
  
  try {
    const stream = await navigator.mediaDevices.getUserMedia({ 
      video: { facingMode: 'environment' }
    })
    mediaStream.value = stream
    await nextTick()
    if (videoRef.value) {
      videoRef.value.srcObject = stream
    }
  } catch (error) {
    toast.error('Gagal mengakses kamera. Pastikan izin kamera diberikan.')
    closeWebcam()
  }
}

const closeWebcam = () => {
  if (mediaStream.value) {
    mediaStream.value.getTracks().forEach(track => track.stop())
    mediaStream.value = null
  }
  showWebcamModal.value = false
}

const captureWebcam = () => {
  if (!videoRef.value) return
  const canvas = document.createElement('canvas')
  canvas.width = videoRef.value.videoWidth
  canvas.height = videoRef.value.videoHeight
  const ctx = canvas.getContext('2d')
  if (!ctx) return
  
  ctx.drawImage(videoRef.value, 0, 0)
  canvas.toBlob((blob) => {
    if (!blob) return
    const file = new File([blob], `capture-${Date.now()}.jpg`, { type: 'image/jpeg' })
    const previewUrl = URL.createObjectURL(file)
    
    if (webcamTarget.value === 'audit') {
      auditPhotoFile.value = file
      auditPhotoPreview.value = previewUrl
    } else if (webcamTarget.value === 'store') {
      selectedFile.value = file
      imagePreview.value = previewUrl
    }
    
    closeWebcam()
  }, 'image/jpeg', 0.8)
}
// ==========================

const isLoading = ref(true)
const storesList = ref<any[]>([])
const totalStores = ref(0)
const storeFetchError = ref<string | null>(null)
const salesmen = ref<any[]>([])
const filteredSalesmen = computed(() => {
  return salesmen.value.filter((emp: any) => {
    const branchId = masterDataStore.selectedBranchId
    const empBranchId = emp.operating_branch_id || emp.branch_id
    return !branchId || branchId === 'ALL_BRANCHES' || empBranchId === branchId
  })
})
const searchQuery = ref('')
const selectedStatus = ref('ACTIVE')
const selectedCategory = ref('ALL')
const selectedCustomerType = ref('ALL')
const selectedSalesman = ref('ALL')
const selectedCompleteness = ref('ALL')

const totalNooToday = computed(() => {
  return storesList.value.filter((s: any) => s.is_noo_today).length
})

const totalEcToday = computed(() => {
  return storesList.value.filter((s: any) => s.is_effective_call || (s.today_order_count && s.today_order_count > 0)).length
})

const totalReceivableSum = computed(() => {
  return storesList.value.reduce((acc: number, s: any) => acc + (s.total_receivable || 0), 0)
})

const customerTypeOptions = [
  { value: 'WARUNG', label: 'Warung' },
  { value: 'TOKO', label: 'Toko' },
  { value: 'GROSIR', label: 'Grosir' },
  { value: 'AGEN', label: 'Agen' },
  { value: 'DISTRIBUTOR', label: 'Distributor' },
  { value: 'PASAR', label: 'Pasar' },
  { value: 'MINIMARKET', label: 'Minimarket' },
  { value: 'SUPERMARKET', label: 'Supermarket' },
  { value: 'HOREKA', label: 'Horeka' },
  { value: 'LAINNYA', label: 'Lainnya' },
]

const horekaSubTypes = [
  { value: 'HOTEL', label: 'Hotel' },
  { value: 'RESTORAN', label: 'Restoran' },
  { value: 'CAFE', label: 'Café' },
  { value: 'KATERING', label: 'Katering' },
  { value: 'RUMAH_MAKAN', label: 'Rumah Makan' },
  { value: 'LAINNYA', label: 'Lainnya' },
]

const frequencyOptions = [
  { value: 'F1', label: 'F1 (1x Seminggu)', desc: 'Kunjungan rutin dilakukan 1 kali setiap minggu (Seminggu sekali)' },
  { value: 'F2', label: 'F2 (2 Minggu Sekali)', desc: 'Kunjungan rutin dilakukan 1 kali setiap 2 minggu (Dua minggu sekali)' },
  { value: 'F4', label: 'F4 (4 Minggu Sekali / Bulanan)', desc: 'Kunjungan rutin dilakukan 1 kali setiap 4 minggu (Sebulan sekali)' },
]

const dayOptions = [
  { value: 1, label: 'Senin', short: 'Sen' },
  { value: 2, label: 'Selasa', short: 'Sel' },
  { value: 3, label: 'Rabu', short: 'Rab' },
  { value: 4, label: 'Kamis', short: 'Kam' },
  { value: 5, label: 'Jumat', short: 'Jum' },
  { value: 6, label: 'Sabtu', short: 'Sab' },
  { value: 7, label: 'Minggu', short: 'Min' },
]

const formatVisitFrequencyBadge = (freq: string) => {
  if (!freq) return 'F1'
  return freq
}

const formatVisitDaysSummary = (days: number[]) => {
  if (!days || days.length === 0) return ''
  const sorted = [...days].sort((a, b) => a - b)
  return sorted.map((d) => dayOptions.find((opt) => opt.value === d)?.short || d).join(', ')
}

const formatCustomerTypeBadge = (store: any) => {
  const type = store.customer_type || 'TOKO'
  const opt = customerTypeOptions.find((o) => o.value === type)
  const typeLabel = opt ? opt.label : type

  if (type === 'HOREKA' && store.sub_customer_type) {
    const subOpt = horekaSubTypes.find((s) => s.value === store.sub_customer_type)
    return `${typeLabel} (${subOpt ? subOpt.label : store.sub_customer_type})`
  }
  if (type === 'DISTRIBUTOR' && store.distribution_city) {
    return `${typeLabel} (${store.distribution_city})`
  }
  return typeLabel
}

const setCompletenessFilter = (val: string) => {
  selectedCompleteness.value = val
}

const isIncompleteActive = computed(() => {
  return ['INCOMPLETE', 'MISSING_PHONE', 'MISSING_LOCATION', 'MISSING_PHOTO', 'MISSING_OWNER', 'MISSING_SALESMAN', 'MISSING_ADDRESS'].includes(selectedCompleteness.value)
})

const showModal = ref(false)
const isEditing = ref(false)
const map = ref<any>(null)
const marker = ref<any>(null)
const isGeocoding = ref(false)
const isMapLoading = ref(false)
const isGettingLocation = ref(false)
const showImportModal = ref(false)

const getMissingFields = (store: any): { key: string; label: string }[] => {
  const missing: { key: string; label: string }[] = []
  
  // 1. Phone number
  const phone = (store.phone_number || '').trim()
  if (!phone || phone === '-' || phone === '0') {
    missing.push({ key: 'phone', label: 'No. Telepon / WA' })
  }
  
  // 2. Owner name
  const owner = (store.owner_name || '').trim()
  if (!owner || owner === '-') {
    missing.push({ key: 'owner', label: 'Nama Pemilik' })
  }
  
  // 3. Photo URL
  if (!store.photo_url || !store.photo_url.trim()) {
    missing.push({ key: 'photo', label: 'Foto Toko' })
  }
  
  // 4. GPS Coordinates
  const lat = Number(store.latitude || 0)
  const lng = Number(store.longitude || 0)
  if (lat === 0 && lng === 0) {
    missing.push({ key: 'location', label: 'Titik Lokasi (GPS)' })
  }
  
  // 5. Address / Administrative region
  const address = (store.address || '').trim()
  const regency = (store.regency || '').trim()
  if (!address) {
    missing.push({ key: 'address', label: 'Alamat Toko' })
  } else if (!regency && !store.province) {
    missing.push({ key: 'region', label: 'Wilayah (Kec./Kab.)' })
  }
  
  // 6. Assigned Salesman
  if (!store.assigned_employee_id && !store.assigned_employee) {
    missing.push({ key: 'salesman', label: 'Salesman PIC' })
  }
  
  return missing
}

const followUpSalesViaWhatsApp = (store: any) => {
  const missing = getMissingFields(store)
  const salesman = store.assigned_employee
  const salesPhone = salesman?.phone_number ? salesman.phone_number.replace(/[^0-9]/g, '') : ''
  
  const missingListText = missing.map((m, idx) => `  ${idx + 1}. ${m.label}`).join('\n')
  const salesName = salesman ? `${salesman.first_name || ''} ${salesman.last_name || ''}`.trim() : 'Sales'
  const storeIdShort = store.id ? store.id.split('-')[0].toUpperCase() : '-'
  
  let message = `Halo Mas/Mbak ${salesName},\n\n`
  message += `Mohon bantuannya untuk melengkapi data toko customer berikut:\n`
  message += `🏪 *Nama Toko*: ${store.name || '-'}\n`
  message += `🆔 *ID Toko*: ${storeIdShort}\n`
  message += `📍 *Alamat*: ${store.address || '-'}\n\n`
  message += `⚠️ *Data yang masih perlu dilengkapi*:\n${missingListText}\n\n`
  message += `Mohon bantu update melalui aplikasi ya saat kunjungan berikutnya. Terima kasih banyak! 🙏`

  if (salesPhone) {
    let formattedPhone = salesPhone
    if (formattedPhone.startsWith('0')) {
      formattedPhone = '62' + formattedPhone.substring(1)
    } else if (!formattedPhone.startsWith('62')) {
      formattedPhone = '62' + formattedPhone
    }
    const url = `https://wa.me/${formattedPhone}?text=${encodeURIComponent(message)}`
    window.open(url, '_blank')
  } else {
    navigator.clipboard.writeText(message)
    toast.info('Nomor WhatsApp salesman belum tercatat di sistem. Format pesan follow-up berhasil disalin ke clipboard.')
  }
}

const handleImportSuccess = () => {
  showImportModal.value = false
  fetchStores()
}

const getCurrentLocation = () => {
  if (!navigator.geolocation) {
    toast.error('Browser Anda tidak mendukung fitur lokasi GPS.')
    return
  }

  isGettingLocation.value = true
  navigator.geolocation.getCurrentPosition(
    (position) => {
      const lat = position.coords.latitude
      const lng = position.coords.longitude
      
      currentStore.value.latitude = lat
      currentStore.value.longitude = lng

      if (map.value && marker.value) {
        marker.value.setLatLng([lat, lng])
        map.value.setView([lat, lng], 16)
      }
      
      toast.success('Lokasi saat ini berhasil didapatkan.')
      isGettingLocation.value = false
    },
    (error) => {
      console.error('Error mendapatkan lokasi:', error)
      toast.error('Gagal mendapatkan lokasi. Pastikan izin akses lokasi (GPS) diaktifkan.')
      isGettingLocation.value = false
    },
    {
      enableHighAccuracy: true,
      timeout: 10000,
      maximumAge: 0
    }
  )
}

// Salesman Dropdown States
const salesmanDropdownRef = ref<HTMLElement | null>(null)
const salesmanSearchQuery = ref('')
const isSalesmanDropdownOpen = ref(false)

onClickOutside(salesmanDropdownRef, () => {
  isSalesmanDropdownOpen.value = false
})

const filteredSalesmenForDropdown = computed(() => {
  if (!salesmanSearchQuery.value) return filteredSalesmen.value
  const query = salesmanSearchQuery.value.toLowerCase()
  return filteredSalesmen.value.filter((s: any) => {
    const firstName = (s.first_name || '').toLowerCase()
    const lastName = (s.last_name || '').toLowerCase()
    const fullName = `${firstName} ${lastName}`.trim()
    return fullName.includes(query)
  })
})

const selectSalesman = (s: any) => {
  currentStore.value.assigned_employee_id = s.id
  salesmanSearchQuery.value = `${s.first_name} ${s.last_name}`
  isSalesmanDropdownOpen.value = false
}

// Pagination states
const currentPage = ref(1)
const itemsPerPage = ref(10)

// Image Preview states
const showImagePreview = ref(false)
const previewImageUrl = ref('')

const openImagePreview = (url: string) => {
  if (!url) return
  previewImageUrl.value = url
  showImagePreview.value = true
}

// History Order States
const showHistoryModal = ref(false)
const isHistoryLoading = ref(false)
const historyOrders = ref<any[]>([])
const currentHistoryStore = ref<any>(null)
const historyCurrentPage = ref(1)
const historyItemsPerPage = ref(10)
const historyTotal = ref(0)
const historyTotalPages = computed(() => Math.ceil(historyTotal.value / historyItemsPerPage.value))
const expandedOrderIds = ref<string[]>([])

const toggleExpandOrder = (id: string) => {
  if (expandedOrderIds.value.includes(id)) {
    expandedOrderIds.value = expandedOrderIds.value.filter(i => i !== id)
  } else {
    expandedOrderIds.value.push(id)
  }
}

const fetchStoreHistory = async (storeId: string) => {
  isHistoryLoading.value = true
  try {
    const res = await apiClient.get(`/admin/sales-orders/store/${storeId}/orders?page=${historyCurrentPage.value}&limit=${historyItemsPerPage.value}`)
    historyOrders.value = res.data.data || []
    historyTotal.value = res.data.meta?.total || historyOrders.value.length
  } catch (error) {
    toast.error('Gagal memuat riwayat pesanan')
  } finally {
    isHistoryLoading.value = false
  }
}

const openHistoryModal = (store: any) => {
  currentHistoryStore.value = store
  historyCurrentPage.value = 1
  historyOrders.value = []
  historyTotal.value = 0
  expandedOrderIds.value = []
  showHistoryModal.value = true
  fetchStoreHistory(store.id)
}

const closeHistoryModal = () => {
  showHistoryModal.value = false
  currentHistoryStore.value = null
}

// === Audit History Logic ===
const showAuditHistoryModal = ref(false)
const currentAuditHistoryStore = ref<any>(null)
const auditHistories = ref<any[]>([])
const isAuditHistoryLoading = ref(false)

const openAuditHistoryModal = async (store: any) => {
  currentAuditHistoryStore.value = store
  showAuditHistoryModal.value = true
  isAuditHistoryLoading.value = true
  auditHistories.value = []
  
  try {
    const res = await apiClient.get(`/stores/${store.id}/audits`)
    auditHistories.value = res.data.data || []
  } catch (error) {
    toast.error('Gagal memuat riwayat verifikasi')
  } finally {
    isAuditHistoryLoading.value = false
  }
}

const closeAuditHistoryModal = () => {
  showAuditHistoryModal.value = false
  currentAuditHistoryStore.value = null
  auditHistories.value = []
}
// === Global Audit History Logic ===
const showGlobalAuditHistoryModal = ref(false)
const globalAuditHistories = ref<any[]>([])
const isGlobalAuditHistoryLoading = ref(false)
const globalAuditPage = ref(1)
const hasMoreGlobalAudits = ref(true)

const fetchGlobalAudits = async (loadMore = false) => {
  if (isGlobalAuditHistoryLoading.value) return
  if (!loadMore) {
    globalAuditPage.value = 1
    globalAuditHistories.value = []
    hasMoreGlobalAudits.value = true
  }

  isGlobalAuditHistoryLoading.value = true
  try {
    const params = new URLSearchParams({
      page: globalAuditPage.value.toString(),
      limit: '20'
    })

    if (masterDataStore.selectedBranchCompanyId) {
      params.append('company_id', masterDataStore.selectedBranchCompanyId)
    }
    if (masterDataStore.selectedBranchId && masterDataStore.selectedBranchId !== 'ALL_BRANCHES') {
      params.append('branch_id', masterDataStore.selectedBranchId)
    }

    const res = await apiClient.get(`/stores/audits/all?${params.toString()}`)
    const newData = res.data.data || []
    
    if (loadMore) {
      globalAuditHistories.value.push(...newData)
    } else {
      globalAuditHistories.value = newData
    }
    
    hasMoreGlobalAudits.value = globalAuditPage.value < (res.data.meta?.total_pages || 1)
  } catch (error) {
    toast.error('Gagal memuat rekap verifikasi global')
  } finally {
    isGlobalAuditHistoryLoading.value = false
  }
}

const openGlobalAuditHistoryModal = () => {
  showGlobalAuditHistoryModal.value = true
  fetchGlobalAudits()
}

const closeGlobalAuditHistoryModal = () => {
  showGlobalAuditHistoryModal.value = false
}

const loadMoreGlobalAudits = () => {
  if (hasMoreGlobalAudits.value) {
    globalAuditPage.value++
    fetchGlobalAudits(true)
  }
}
// =========================

const historyNextPage = () => {
  if (historyCurrentPage.value < historyTotalPages.value) {
    historyCurrentPage.value++
    if (currentHistoryStore.value) {
      fetchStoreHistory(currentHistoryStore.value.id)
    }
  }
}

const historyPrevPage = () => {
  if (historyCurrentPage.value > 1) {
    historyCurrentPage.value--
    if (currentHistoryStore.value) {
      fetchStoreHistory(currentHistoryStore.value.id)
    }
  }
}

const getStatusClass = (status: string) => {
  switch (status) {
    case 'DELIVERED':
    case 'CONFIRMED':
      return 'bg-emerald-50 text-emerald-600 border-emerald-100'
    case 'WAITING_STOCK':
      return 'bg-amber-50 text-amber-600 border-amber-100'
    case 'CANCELLED':
    case 'REJECTED':
      return 'bg-red-50 text-red-600 border-red-100'
    default:
      return 'bg-slate-50 text-slate-600 border-slate-100'
  }
}

const formatDate = (dateString: string) => {
  if (!dateString) return '-'
  return new Date(dateString).toLocaleDateString('id-ID', {
    day: 'numeric',
    month: 'short',
    year: 'numeric'
  })
}

const formatTime = (dateString: string) => {
  if (!dateString) return '-'
  return new Date(dateString).toLocaleTimeString('id-ID', {
    hour: '2-digit',
    minute: '2-digit',
  }) + ' WIB'
}

const formatDateTime = (dateString: string) => {
  if (!dateString) return '-'
  const d = new Date(dateString)
  return `${d.toLocaleDateString('id-ID', { day: 'numeric', month: 'short', year: 'numeric' })}, ${d.toLocaleTimeString('id-ID', { hour: '2-digit', minute: '2-digit' })} WIB`
}


// === Region State & Methods ===
const provinces = ref<any[]>([])
const regencies = ref<any[]>([])
const districts = ref<any[]>([])
const villages = ref<any[]>([])
const selectedProvinceId = ref<string | null>(null)
const selectedRegencyId = ref<string | null>(null)
const selectedDistrictId = ref<string | null>(null)
const selectedVillageId = ref<string | null>(null)
const isLoadingRegions = ref(false)

const loadProvinces = async () => {
  isLoadingRegions.value = true
  try {
    provinces.value = await regionApi.getProvinces()
  } catch (e) {
    console.error('Failed to load provinces:', e)
  } finally {
    isLoadingRegions.value = false
  }
}

const onProvinceChange = async () => {
  regencies.value = []
  districts.value = []
  villages.value = []
  selectedRegencyId.value = null
  selectedDistrictId.value = null
  selectedVillageId.value = null
  currentStore.value.province = provinces.value.find((p) => p.id === selectedProvinceId.value)?.name || ''
  currentStore.value.regency = ''
  currentStore.value.district = ''
  currentStore.value.village = ''
  
  if (!selectedProvinceId.value) return
  isLoadingRegions.value = true
  try {
    regencies.value = await regionApi.getRegencies(selectedProvinceId.value)
  } catch (e) {
    console.error('Failed to load regencies:', e)
  } finally {
    isLoadingRegions.value = false
  }
}

const onRegencyChange = async () => {
  districts.value = []
  villages.value = []
  selectedDistrictId.value = null
  selectedVillageId.value = null
  currentStore.value.regency = regencies.value.find((r) => r.id === selectedRegencyId.value)?.name || ''
  currentStore.value.district = ''
  currentStore.value.village = ''

  if (!selectedRegencyId.value) return
  isLoadingRegions.value = true
  try {
    districts.value = await regionApi.getDistricts(selectedRegencyId.value)
  } catch (e) {
    console.error('Failed to load districts:', e)
  } finally {
    isLoadingRegions.value = false
  }
}

const onDistrictChange = async () => {
  villages.value = []
  selectedVillageId.value = null
  currentStore.value.district = districts.value.find((d) => d.id === selectedDistrictId.value)?.name || ''
  currentStore.value.village = ''

  if (!selectedDistrictId.value) return
  isLoadingRegions.value = true
  try {
    villages.value = await regionApi.getVillages(selectedDistrictId.value)
  } catch (e) {
    console.error('Failed to load villages:', e)
  } finally {
    isLoadingRegions.value = false
  }
}

const onVillageChange = () => {
  currentStore.value.village = villages.value.find((v) => v.id === selectedVillageId.value)?.name || ''
}
// =============================

const currentStore = ref<{
  id: string
  name: string
  owner_name: string
  phone_number: string
  address: string
  latitude: number
  longitude: number
  assigned_employee_id: string
  is_active: boolean
  category: string
  customer_type: string
  sub_customer_type: string
  distribution_city: string
  visit_frequency: string
  visit_days: number[]
  photo_url: string
  npwp: string
  nik: string
  tax_name: string
  province: string
  regency: string
  district: string
  village: string
}>({
  id: '',
  name: '',
  owner_name: '',
  phone_number: '',
  address: '',
  latitude: 0,
  longitude: 0,
  assigned_employee_id: '',
  is_active: true,
  category: 'TOKO_LAMA',
  customer_type: 'TOKO',
  sub_customer_type: '',
  distribution_city: '',
  visit_frequency: 'F1',
  visit_days: [1, 2, 3, 4, 5, 6],
  photo_url: '',
  npwp: '',
  nik: '',
  tax_name: '',
  province: '',
  regency: '',
  district: '',
  village: '',
})

const selectedFile = ref<File | null>(null)
const imagePreview = ref<string | null>(null)

const handleFileChange = (e: any) => {
  const file = e.target.files[0]
  if (file) {
    selectedFile.value = file
    imagePreview.value = URL.createObjectURL(file)
  }
}

const getErrorMessage = (error: any, defaultMsg: string): string => {
  if (error?.code === 'ECONNABORTED' || (error?.message && error.message.includes('timeout'))) {
    return 'Waktu permintaan habis (timeout). Database/server sedang sibuk memproses data.'
  }
  if (typeof navigator !== 'undefined' && !navigator.onLine) {
    return 'Koneksi internet terputus. Periksa sambungan Wi-Fi atau paket data Anda.'
  }
  if (error?.response) {
    const status = error.response.status
    if (status === 403) {
      return 'Akses ditolak: Akun Anda tidak memiliki izin untuk mengakses data ini.'
    }
    if (status === 401) {
      return 'Sesi login telah berakhir. Silakan login kembali.'
    }
    if (status === 500) {
      return error.response.data?.error || 'Terjadi gangguan internal pada server (HTTP 500).'
    }
    if (status === 502 || status === 503 || status === 504) {
      return 'Server sedang sibuk atau tidak merespons (Gateway Timeout/Error).'
    }
    if (error.response.data?.error) {
      return error.response.data.error
    }
    if (error.response.data?.message) {
      return error.response.data.message
    }
  }
  return defaultMsg
}

let searchTimeout: any = null
const isBackgroundSyncing = ref(false)
const isUsingOfflineCache = ref(false)
let currentAbortController: AbortController | null = null

const fetchStores = async (forceBackground = false) => {
  if (!masterDataStore.selectedBranchId) return
  
  const cacheKey = `stores_cache_${masterDataStore.selectedBranchId}_${currentPage.value}_${selectedCategory.value}_${selectedCustomerType.value}_${selectedSalesman.value}_${selectedCompleteness.value}_${selectedStatus.value}_${masterDataStore.selectedSalesType}_${searchQuery.value}`

  // 1. Instant Cache Load (Stale-While-Revalidate)
  const cachedData = localStorage.getItem(cacheKey)
  if (cachedData) {
    try {
      const parsed = JSON.parse(cachedData)
      storesList.value = parsed.list || []
      totalStores.value = parsed.total || 0
      isUsingOfflineCache.value = false
      isLoading.value = false // Instant rendering for slow devices
      isBackgroundSyncing.value = true
    } catch (e) {
      if (!forceBackground && storesList.value.length === 0) {
        isLoading.value = true
      }
    }
  } else {
    if (!forceBackground && storesList.value.length === 0) {
      isLoading.value = true
    }
  }

  // 2. Cancel previous pending request to prevent bandwidth clogging
  if (currentAbortController) {
    currentAbortController.abort()
  }
  currentAbortController = new AbortController()

  storeFetchError.value = null

  try {
    const params = new URLSearchParams({
      page: currentPage.value.toString(),
      limit: itemsPerPage.value.toString(),
      search: searchQuery.value,
      status: selectedStatus.value,
      company_id: masterDataStore.selectedBranchCompanyId || '',
      sales_type: masterDataStore.selectedSalesType,
    })

    if (masterDataStore.selectedBranchId && masterDataStore.selectedBranchId !== 'ALL_BRANCHES') {
      params.append('branch_id', masterDataStore.selectedBranchId)
    }

    if (selectedCategory.value !== 'ALL') {
      params.append('category', selectedCategory.value)
    }
    if (selectedCustomerType.value !== 'ALL') {
      params.append('customer_type', selectedCustomerType.value)
    }
    if (selectedSalesman.value !== 'ALL') {
      params.append('assigned_employee_id', selectedSalesman.value)
    }
    if (selectedCompleteness.value !== 'ALL') {
      params.append('completeness', selectedCompleteness.value)
    }

    const res = await apiClient.get(`/stores?${params.toString()}`, {
      signal: currentAbortController.signal,
    })

    storesList.value = res.data.data || []
    totalStores.value = res.data.meta?.total || storesList.value.length
    storeFetchError.value = null
    isUsingOfflineCache.value = false
    
    // Save to Cache
    try {
      localStorage.setItem(cacheKey, JSON.stringify({
        list: storesList.value,
        total: totalStores.value,
        savedAt: Date.now(),
      }))
    } catch (e) {
      console.warn('Failed to cache stores data in localStorage:', e)
    }
  } catch (error: any) {
    // If request was canceled by AbortController, ignore error
    if (error?.name === 'CanceledError' || error?.code === 'ERR_CANCELED' || error?.name === 'AbortError') {
      return
    }

    console.error('Failed to fetch stores:', error)
    const errorMsg = getErrorMessage(error, 'Gagal memuat data toko.')

    // If we already have cached data rendered, don't wipe it out!
    if (storesList.value.length > 0) {
      isUsingOfflineCache.value = true
      // Don't spam toast if user already has data, show friendly warning once
      toast.warning('Koneksi lambat atau terputus. Menampilkan data toko tersimpan.')
    } else {
      storeFetchError.value = errorMsg
      storesList.value = []
      totalStores.value = 0
      toast.error(errorMsg)
    }
  } finally {
    isLoading.value = false
    isBackgroundSyncing.value = false
  }
}

let salesmenAbortController: AbortController | null = null

const fetchSalesmen = async () => {
  if (!masterDataStore.selectedBranchId) return
  
  const cacheKey = `salesmen_cache_${masterDataStore.selectedBranchId}_${masterDataStore.selectedSalesType}`
  
  // Instant load salesmen from cache
  const cachedData = localStorage.getItem(cacheKey)
  if (cachedData) {
    try {
      salesmen.value = JSON.parse(cachedData)
    } catch (e) {}
  }

  if (salesmenAbortController) {
    salesmenAbortController.abort()
  }
  salesmenAbortController = new AbortController()

  try {
    const res = await apiClient.get(
      `/employees?limit=all&is_sales=true&branch_id=${masterDataStore.selectedBranchId !== 'ALL_BRANCHES' ? masterDataStore.selectedBranchId : ''}&sales_type=${masterDataStore.selectedSalesType}`,
      { signal: salesmenAbortController.signal }
    )
    if (res.data?.data) {
      salesmen.value = res.data.data || []
      try {
        localStorage.setItem(cacheKey, JSON.stringify(salesmen.value))
      } catch (e) {}
    }
  } catch (error: any) {
    if (error?.name === 'CanceledError' || error?.code === 'ERR_CANCELED' || error?.name === 'AbortError') {
      return
    }
    console.error('Failed to fetch salesmen:', error)
    // If no salesmen in cache, show error
    if (salesmen.value.length === 0) {
      const errorMsg = getErrorMessage(error, 'Gagal memuat filter sales.')
      toast.error(errorMsg)
    }
  }
}

const initMap = async () => {
  if (map.value) return

  await nextTick()
  const mapElement = document.getElementById('store-map')
  if (!mapElement) return

  // Small delay to ensure modal transition finished
  setTimeout(() => {
    const el = document.getElementById('store-map')
    if (!el || map.value) return

    const defaultLat = currentStore.value.latitude || -6.2088
    const defaultLng = currentStore.value.longitude || 106.8456

    map.value = L.map('store-map', {
      zoomControl: true,
      scrollWheelZoom: true,
    }).setView([defaultLat, defaultLng], 15)

    L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
      attribution: '&copy; OpenStreetMap contributors',
    }).addTo(map.value)

    marker.value = L.marker([defaultLat, defaultLng], {
      draggable: true,
    }).addTo(map.value)

    marker.value.on('dragend', (e: any) => {
      const pos = e.target.getLatLng()
      currentStore.value.latitude = pos.lat
      currentStore.value.longitude = pos.lng
    })

    map.value.on('click', (e: any) => {
      const pos = e.latlng
      marker.value.setLatLng(pos)
      currentStore.value.latitude = pos.lat
      currentStore.value.longitude = pos.lng
    })

    // Force invalidation to fix grey tiles
    setTimeout(() => {
      map.value?.invalidateSize()
    }, 400)
  }, 100)
}

// Watch showModal to trigger map init and cleanup
watch(showModal, (newVal) => {
  if (newVal) {
    initMap()
  } else {
    if (map.value) {
      map.value.remove()
      map.value = null
      marker.value = null
    }
  }
})

const reverseGeocode = async (lat: number, lng: number) => {
  if (lat === 0 && lng === 0) return
  isGeocoding.value = true
  try {
    const res = await fetch(
      `https://nominatim.openstreetmap.org/reverse?format=json&lat=${lat}&lon=${lng}`,
    )
    const data = await res.json()
    if (data.display_name) {
      currentStore.value.address = data.display_name
    }
  } catch (error) {
    console.error('Reverse geocoding failed:', error)
  } finally {
    isGeocoding.value = false
  }
}

const toggleVisitDay = (day: number) => {
  if (!currentStore.value.visit_days) {
    currentStore.value.visit_days = []
  }
  const index = currentStore.value.visit_days.indexOf(day)
  if (index > -1) {
    currentStore.value.visit_days.splice(index, 1)
  } else {
    currentStore.value.visit_days.push(day)
  }
}

const openAddModal = () => {
  isMapLoading.value = true
  isEditing.value = false
  currentStore.value = {
    id: '',
    name: '',
    owner_name: '',
    phone_number: '',
    address: '',
    latitude: -6.2088,
    longitude: 106.8456,
    assigned_employee_id: '',
    is_active: true,
    category: 'TOKO_LAMA',
    customer_type: 'TOKO',
    sub_customer_type: '',
    distribution_city: '',
    visit_frequency: 'F1',
    visit_days: [1, 2, 3, 4, 5, 6],
    photo_url: '',
    npwp: '',
    nik: '',
    tax_name: '',
    province: '',
    regency: '',
    district: '',
    village: ''
  }
  
  selectedProvinceId.value = null
  selectedRegencyId.value = null
  selectedDistrictId.value = null
  selectedVillageId.value = null
  regencies.value = []
  districts.value = []
  villages.value = []
  loadProvinces()

  salesmanSearchQuery.value = ''
  selectedFile.value = null
  imagePreview.value = null
  showModal.value = true
  setTimeout(() => {
    isMapLoading.value = false
  }, 800)
}

const handleEdit = async (store: any) => {
  isMapLoading.value = true
  isEditing.value = true
  currentStore.value = {
    id: store.id,
    name: store.name,
    owner_name: store.owner_name,
    phone_number: store.phone_number,
    address: store.address,
    latitude: store.latitude,
    longitude: store.longitude,
    assigned_employee_id: store.assigned_employee_id || '',
    is_active: store.is_active,
    category: store.category || 'TOKO_LAMA',
    customer_type: store.customer_type || 'TOKO',
    sub_customer_type: store.sub_customer_type || '',
    distribution_city: store.distribution_city || '',
    visit_frequency: store.visit_frequency || 'F1',
    visit_days: Array.isArray(store.visit_days) ? [...store.visit_days] : [],
    photo_url: store.photo_url || '',
    npwp: store.npwp || '',
    nik: store.nik || '',
    tax_name: store.tax_name || '',
    province: store.province || '',
    regency: store.regency || '',
    district: store.district || '',
    village: store.village || ''
  }
  
  // Load region data sequentially based on existing names
  selectedProvinceId.value = null
  selectedRegencyId.value = null
  selectedDistrictId.value = null
  selectedVillageId.value = null
  regencies.value = []
  districts.value = []
  villages.value = []
  
  await loadProvinces()
  
  if (currentStore.value.province) {
    const prov = provinces.value.find((p) => p.name === currentStore.value.province)
    if (prov) {
      selectedProvinceId.value = prov.id
      isLoadingRegions.value = true
      try {
        regencies.value = await regionApi.getRegencies(prov.id)
        if (currentStore.value.regency) {
          const reg = regencies.value.find((r) => r.name === currentStore.value.regency)
          if (reg) {
            selectedRegencyId.value = reg.id
            districts.value = await regionApi.getDistricts(reg.id)
            if (currentStore.value.district) {
              const dist = districts.value.find((d) => d.name === currentStore.value.district)
              if (dist) {
                selectedDistrictId.value = dist.id
                villages.value = await regionApi.getVillages(dist.id)
                if (currentStore.value.village) {
                  const vil = villages.value.find((v) => v.name === currentStore.value.village)
                  if (vil) selectedVillageId.value = vil.id
                }
              }
            }
          }
        }
      } catch(e) {}
      finally {
        isLoadingRegions.value = false
      }
    }
  }

  
  const emp = salesmen.value.find(s => s.id === store.assigned_employee_id)
  if (emp) {
    salesmanSearchQuery.value = `${emp.first_name} ${emp.last_name}`
  } else {
    salesmanSearchQuery.value = ''
  }

  selectedFile.value = null
  imagePreview.value = null
  showModal.value = true
  setTimeout(() => {
    isMapLoading.value = false
  }, 800)
}

const handleDelete = async (id: string) => {
  if (
    !confirm(
      'Yakin ingin menghapus toko ini? Semua riwayat penjualan akan tetap ada namun toko tidak akan muncul di daftar aktif.',
    )
  )
    return
  try {
    await apiClient.delete(`/stores/${id}`)
    toast.success('Toko berhasil dihapus')
    fetchStores()
  } catch (error) {
    toast.error('Gagal menghapus toko')
  }
}

const closeModal = () => {
  showModal.value = false
  map.value = null
}

const handleSubmit = async () => {
  try {
    const formData = new FormData()
    formData.append('name', currentStore.value.name)
    formData.append('owner_name', currentStore.value.owner_name || '')
    formData.append('phone_number', currentStore.value.phone_number || '')
    formData.append('address', currentStore.value.address || '')
    formData.append('latitude', String(currentStore.value.latitude))
    formData.append('longitude', String(currentStore.value.longitude))
    if (currentStore.value.assigned_employee_id) {
      formData.append('assigned_employee_id', currentStore.value.assigned_employee_id)
    } else {
      formData.append('assigned_employee_id', '')
    }
    formData.append('is_active', String(currentStore.value.is_active))
    formData.append('category', currentStore.value.category || 'TOKO_LAMA')
    formData.append('customer_type', currentStore.value.customer_type || 'TOKO')
    formData.append('sub_customer_type', currentStore.value.customer_type === 'HOREKA' ? (currentStore.value.sub_customer_type || '') : '')
    formData.append('distribution_city', currentStore.value.customer_type === 'DISTRIBUTOR' ? (currentStore.value.distribution_city || '') : '')
    formData.append('visit_frequency', currentStore.value.visit_frequency || 'F1')
    if (currentStore.value.visit_days && currentStore.value.visit_days.length > 0) {
      currentStore.value.visit_days.forEach((day: number) => {
        formData.append('visit_days[]', String(day))
      })
    }
    formData.append('company_id', masterDataStore.selectedBranchCompanyId)

    if (currentStore.value.npwp) formData.append('npwp', currentStore.value.npwp)
    if (currentStore.value.nik) formData.append('nik', currentStore.value.nik)
    if (currentStore.value.tax_name) formData.append('tax_name', currentStore.value.tax_name)
    if (currentStore.value.province) formData.append('province', currentStore.value.province)
    if (currentStore.value.regency) formData.append('regency', currentStore.value.regency)
    if (currentStore.value.district) formData.append('district', currentStore.value.district)
    if (currentStore.value.village) formData.append('village', currentStore.value.village)

    if (selectedFile.value) {
      formData.append('photo', selectedFile.value)
    }

    if (isEditing.value) {
      await apiClient.put(`/stores/${currentStore.value.id}`, formData)
      toast.success('Data toko diperbarui')
    } else {
      await apiClient.post('/stores', formData)
      toast.success('Toko baru berhasil terdaftar')
    }
    showModal.value = false
    fetchStores()
  } catch (error: any) {
    console.error('Gagal menyimpan data toko:', error)
    toast.error(
      error.response?.data?.error ||
      error.response?.data?.message ||
      error.message ||
      'Gagal menyimpan data toko'
    )
  }
}

watch(
  () => currentStore.value.latitude,
  (newVal) => {
    if (isMapLoading.value) return
    if (newVal && currentStore.value.longitude && !isGeocoding.value) {
      reverseGeocode(newVal, currentStore.value.longitude)
    }
  },
)

watch(
  () => currentStore.value.longitude,
  (newVal) => {
    if (isMapLoading.value) return
    if (newVal && currentStore.value.latitude && !isGeocoding.value) {
      reverseGeocode(currentStore.value.latitude, newVal)
    }
  },
)

watch(searchQuery, () => {
  if (searchTimeout) clearTimeout(searchTimeout)
  searchTimeout = setTimeout(() => {
    currentPage.value = 1
    fetchStores()
  }, 500)
})

watch(selectedStatus, () => {
  currentPage.value = 1
  fetchStores()
})

watch(selectedCategory, () => {
  currentPage.value = 1
  fetchStores()
})

watch(selectedCustomerType, () => {
  currentPage.value = 1
  fetchStores()
})

watch(selectedSalesman, () => {
  currentPage.value = 1
  fetchStores()
})

watch(selectedCompleteness, () => {
  currentPage.value = 1
  fetchStores()
})

watch([currentPage, itemsPerPage], () => {
  fetchStores()
})

const totalPages = computed(() => Math.ceil(totalStores.value / itemsPerPage.value))

const nextPage = () => {
  if (currentPage.value < totalPages.value) currentPage.value++
}
const prevPage = () => {
  if (currentPage.value > 1) currentPage.value--
}

let isComponentMounted = false

onMounted(() => {
  isComponentMounted = true
  fetchStores()
  fetchSalesmen()
})

watch(
  () => masterDataStore.selectedBranchId,
  () => {
    if (!isComponentMounted) return
    selectedSalesman.value = 'ALL'
    if (currentStore.value) {
      currentStore.value.assigned_employee_id = ''
    }
    currentPage.value = 1
    fetchStores()
    fetchSalesmen()
  },
)

watch(
  () => masterDataStore.selectedSalesType,
  () => {
    if (!isComponentMounted) return
    selectedSalesman.value = 'ALL'
    currentPage.value = 1
    fetchStores()
    fetchSalesmen()
  },
)
const formatCurrency = (amount: number) => {
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    minimumFractionDigits: 0,
  }).format(amount)
}

// ── Export Excel ────────────────────────────────────────────────────────────
const isExporting = ref(false)

const handleExportExcel = async () => {
  isExporting.value = true
  try {
    const params = new URLSearchParams({
      search: searchQuery.value,
      status: selectedStatus.value,
      company_id: masterDataStore.selectedBranchCompanyId || '',
      sales_type: masterDataStore.selectedSalesType,
    })

    if (masterDataStore.selectedBranchId && masterDataStore.selectedBranchId !== 'ALL_BRANCHES') {
      params.append('branch_id', masterDataStore.selectedBranchId)
    }

    if (selectedCategory.value !== 'ALL') {
      params.append('category', selectedCategory.value)
    }
    if (selectedCustomerType.value !== 'ALL') {
      params.append('customer_type', selectedCustomerType.value)
    }
    if (selectedSalesman.value !== 'ALL') {
      params.append('assigned_employee_id', selectedSalesman.value)
    }
    if (selectedCompleteness.value !== 'ALL') {
      params.append('completeness', selectedCompleteness.value)
    }

    const res = await apiClient.get(`/stores/export?${params.toString()}`, {
      responseType: 'blob',
    })

    const url = window.URL.createObjectURL(new Blob([res.data]))
    const link = document.createElement('a')
    link.href = url
    const date = new Date().toISOString().split('T')[0]
    link.setAttribute('download', `kelola_toko_${date}.xlsx`)
    document.body.appendChild(link)
    link.click()
    link.remove()
    window.URL.revokeObjectURL(url)

    toast.success('Export Excel berhasil diunduh!')
  } catch {
    toast.error('Gagal mengekspor data toko')
  } finally {
    isExporting.value = false
  }
}
</script>

<template>
  <div class="space-y-8 pb-10">
    <div class="flex flex-col lg:flex-row lg:items-center justify-between gap-4 sm:gap-6">
      <div>
        <h1 class="text-2xl sm:text-3xl font-extrabold text-slate-900 tracking-tight flex items-center flex-wrap gap-2">
          <span>{{ masterDataStore.selectedSalesType === 'sales_mt' ? 'Mitra Modern Trade (MT)' : 'Toko Customer' }}</span>
          <span v-if="masterDataStore.selectedBranchName" class="text-primary text-xl sm:text-2xl font-bold">
            — {{ masterDataStore.selectedBranchName }}
          </span>
        </h1>
        <p class="text-xs sm:text-[15px] font-medium text-slate-500 mt-1 sm:mt-1.5 flex items-center gap-2">
          <Store class="w-4 h-4 text-primary shrink-0" />
          {{ masterDataStore.selectedSalesType === 'sales_mt' ? 'Manajemen data mitra dan titik distribusi Key Account.' : 'Manajemen data pelanggan dan titik distribusi outlet.' }}
        </p>
      </div>
      <div class="grid grid-cols-2 sm:flex sm:flex-wrap items-center gap-2 sm:gap-3 w-full lg:w-auto">
        <Button
          @click="router.push('/sales/tracking-noo-ec')"
          class="bg-purple-600 hover:bg-purple-700 text-white rounded-xl sm:rounded-2xl px-3 sm:px-5 py-2.5 sm:py-5 font-bold text-xs sm:text-sm shadow-md sm:shadow-lg shadow-purple-600/20 flex items-center justify-center gap-1.5 sm:gap-2 transition-all h-auto cursor-pointer"
        >
          <TrendingUp class="w-4 h-4 sm:w-5 sm:h-5 shrink-0" />
          <span>Tracking NOO & EC</span>
        </Button>
        <Button
          @click="openGlobalAuditHistoryModal"
          class="bg-indigo-600 hover:bg-indigo-700 text-white rounded-xl sm:rounded-2xl px-3 sm:px-5 py-2.5 sm:py-5 font-bold text-xs sm:text-sm shadow-md sm:shadow-lg shadow-indigo-600/20 flex items-center justify-center gap-1.5 sm:gap-2 transition-all h-auto cursor-pointer"
        >
          <ClipboardList class="w-4 h-4 sm:w-5 sm:h-5 shrink-0" />
          <span>Rekap Verifikasi</span>
        </Button>
        <Button
          @click="showImportModal = true"
          class="bg-blue-600 hover:bg-blue-700 text-white rounded-xl sm:rounded-2xl px-3 sm:px-5 py-2.5 sm:py-5 font-bold text-xs sm:text-sm shadow-md sm:shadow-lg shadow-blue-600/20 flex items-center justify-center gap-1.5 sm:gap-2 transition-all h-auto cursor-pointer"
        >
          <Upload class="w-4 h-4 sm:w-5 sm:h-5 shrink-0" />
          <span>Import Excel</span>
        </Button>
        <Button
          @click="handleExportExcel"
          :disabled="isExporting"
          class="bg-emerald-600 hover:bg-emerald-700 disabled:opacity-60 text-white rounded-xl sm:rounded-2xl px-3 sm:px-5 py-2.5 sm:py-5 font-bold text-xs sm:text-sm shadow-md sm:shadow-lg shadow-emerald-600/20 flex items-center justify-center gap-1.5 sm:gap-2 transition-all h-auto cursor-pointer"
        >
          <Loader2 v-if="isExporting" class="w-4 h-4 sm:w-5 sm:h-5 animate-spin shrink-0" />
          <FileDown v-else class="w-4 h-4 sm:w-5 sm:h-5 shrink-0" />
          <span>{{ isExporting ? 'Mengekspor...' : 'Export Excel' }}</span>
        </Button>
        <Button
          @click="openAddModal"
          class="bg-primary hover:bg-primary/90 text-white rounded-xl sm:rounded-2xl px-3 sm:px-6 py-2.5 sm:py-5 font-bold text-xs sm:text-sm shadow-md sm:shadow-lg shadow-primary/20 flex items-center justify-center gap-1.5 sm:gap-2 h-auto cursor-pointer"
        >
          <Plus class="w-4 h-4 sm:w-5 sm:h-5 shrink-0" />
          <span class="truncate">{{ masterDataStore.selectedSalesType === 'sales_mt' ? 'Daftar Mitra' : 'Daftar Toko' }}</span>
        </Button>
      </div>
    </div>

    <!-- KPI Quick Summary Cards -->
    <div class="grid grid-cols-2 lg:grid-cols-4 gap-3 sm:gap-4">
      <!-- Card 1: Total Toko -->
      <div
        @click="selectedCategory = 'ALL'"
        class="bg-white p-4 sm:p-5 rounded-2xl sm:rounded-3xl border border-slate-200 shadow-sm hover:border-slate-300 transition-all cursor-pointer group"
        :class="selectedCategory === 'ALL' ? 'ring-2 ring-primary/20 border-primary/40' : ''"
      >
        <div class="flex items-center justify-between">
          <span class="text-[11px] sm:text-xs font-bold text-slate-500 uppercase tracking-wider">
            {{ selectedStatus === 'ACTIVE' ? 'Total Toko Aktif' : selectedStatus === 'INACTIVE' ? 'Total Toko Inaktif' : 'Total Seluruh Toko' }}
          </span>
          <div class="w-8 h-8 rounded-xl bg-slate-100 flex items-center justify-center text-slate-600 group-hover:bg-slate-200 transition-colors">
            <Store class="w-4 h-4" />
          </div>
        </div>
        <p class="text-xl sm:text-2xl font-black text-slate-900 mt-2">{{ totalStores }}</p>
        <p class="text-[10px] text-slate-400 font-medium mt-0.5">
          {{ selectedStatus === 'ACTIVE' ? 'Seluruh outlet aktif terdaftar' : selectedStatus === 'INACTIVE' ? 'Outlet nonaktif / duplikat' : 'Total seluruh outlet' }}
        </p>
      </div>

      <!-- Card 2: NOO Hari Ini -->
      <div
        @click="selectedCategory = 'NOO_TODAY'"
        class="bg-gradient-to-br from-blue-50/70 to-indigo-50/30 p-4 sm:p-5 rounded-2xl sm:rounded-3xl border border-blue-200/80 shadow-sm hover:border-blue-300 transition-all cursor-pointer group"
        :class="selectedCategory === 'NOO_TODAY' ? 'ring-2 ring-blue-500/30 border-blue-400' : ''"
      >
        <div class="flex items-center justify-between">
          <span class="text-[11px] sm:text-xs font-bold text-blue-700 uppercase tracking-wider flex items-center gap-1">
            <Sparkles class="w-3.5 h-3.5 text-blue-600" />
            NOO Hari Ini
          </span>
          <div class="w-8 h-8 rounded-xl bg-blue-100 flex items-center justify-center text-blue-700 group-hover:bg-blue-200 transition-colors">
            <Plus class="w-4 h-4" />
          </div>
        </div>
        <p class="text-xl sm:text-2xl font-black text-blue-900 mt-2">{{ totalNooToday }}</p>
        <p class="text-[10px] text-blue-600 font-bold mt-0.5 flex items-center gap-1">
          <Clock class="w-2.5 h-2.5" />
          Toko diinput hari ini
        </p>
      </div>

      <!-- Card 3: EC (Effective Call) Hari Ini -->
      <div
        @click="selectedCategory = 'EC_TODAY'"
        class="bg-gradient-to-br from-emerald-50/70 to-teal-50/30 p-4 sm:p-5 rounded-2xl sm:rounded-3xl border border-emerald-200/80 shadow-sm hover:border-emerald-300 transition-all cursor-pointer group"
        :class="selectedCategory === 'EC_TODAY' ? 'ring-2 ring-emerald-500/30 border-emerald-400' : ''"
      >
        <div class="flex items-center justify-between">
          <span class="text-[11px] sm:text-xs font-bold text-emerald-700 uppercase tracking-wider flex items-center gap-1">
            <ShoppingBag class="w-3.5 h-3.5 text-emerald-600" />
            EC Hari Ini
          </span>
          <div class="w-8 h-8 rounded-xl bg-emerald-100 flex items-center justify-center text-emerald-700 group-hover:bg-emerald-200 transition-colors">
            <TrendingUp class="w-4 h-4" />
          </div>
        </div>
        <p class="text-xl sm:text-2xl font-black text-emerald-900 mt-2">{{ totalEcToday }}</p>
        <p class="text-[10px] text-emerald-600 font-bold mt-0.5">Toko memiliki orderan hari ini</p>
      </div>

      <!-- Card 4: Total Piutang Aktif -->
      <div
        class="bg-white p-4 sm:p-5 rounded-2xl sm:rounded-3xl border border-slate-200 shadow-sm"
      >
        <div class="flex items-center justify-between">
          <span class="text-[11px] sm:text-xs font-bold text-slate-500 uppercase tracking-wider">Total Piutang</span>
          <div class="w-8 h-8 rounded-xl bg-rose-50 flex items-center justify-center text-rose-600">
            <Banknote class="w-4 h-4" />
          </div>
        </div>
        <p class="text-lg sm:text-xl font-black text-rose-600 mt-2 truncate">{{ formatCurrency(totalReceivableSum) }}</p>
        <p class="text-[10px] text-slate-400 font-medium mt-0.5">Piutang berjalan outlet</p>
      </div>
    </div>

    <!-- Filters -->
    <div
      class="bg-white p-4 sm:p-5 rounded-2xl sm:rounded-3xl border border-slate-200 shadow-sm space-y-4"
    >
      <!-- Top Row: Search, Category, Salesman, Status -->
      <div class="flex flex-col lg:flex-row gap-3 sm:gap-4 items-center flex-wrap">
        <div class="relative flex-1 w-full min-w-[200px]">
          <Search class="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 sm:w-5 sm:h-5 text-slate-400" />
          <input
            v-model="searchQuery"
            type="text"
            placeholder="Cari Nama Toko, Owner, atau Alamat..."
            class="w-full pl-11 sm:pl-12 pr-4 py-3 sm:py-3.5 bg-slate-50 border-none rounded-xl sm:rounded-2xl focus:ring-2 focus:ring-primary/20 transition-all font-medium text-xs sm:text-sm text-slate-900"
          />
        </div>
        <select
          v-model="selectedCategory"
          class="px-4 sm:px-5 py-3 sm:py-3.5 bg-slate-50 border-none rounded-xl sm:rounded-2xl font-bold text-xs sm:text-sm text-slate-700 outline-none focus:ring-2 focus:ring-primary/20 w-full lg:w-auto cursor-pointer"
        >
          <option value="ALL">Semua Status Toko</option>
          <option value="NOO_TODAY">⚡ NOO Hari Ini (Input Hari Ini)</option>
          <option value="EC_TODAY">🛒 EC Hari Ini (Ada Orderan)</option>
          <option value="NON_EC_TODAY">⏳ Non-EC Hari Ini (Belum Ada Order)</option>
          <option value="TOKO_BARU">Toko Baru (NOO - Semua)</option>
          <option value="TOKO_LAMA">Toko Lama (Existing)</option>
        </select>
        <select
          v-model="selectedCustomerType"
          class="px-5 py-3.5 bg-slate-50 border-none rounded-2xl font-bold text-slate-700 outline-none focus:ring-2 focus:ring-primary/20 w-full lg:w-auto cursor-pointer"
        >
          <option value="ALL">Semua Kategori Pelanggan</option>
          <option v-for="opt in customerTypeOptions" :key="opt.value" :value="opt.value">
            {{ opt.label }}
          </option>
        </select>
        <select
          v-model="selectedSalesman"
          class="px-4 sm:px-5 py-3 sm:py-3.5 bg-slate-50 border-none rounded-xl sm:rounded-2xl font-bold text-xs sm:text-sm text-slate-700 outline-none focus:ring-2 focus:ring-primary/20 w-full lg:w-auto cursor-pointer"
        >
          <option value="ALL">Semua Salesman</option>
          <option v-for="s in filteredSalesmen" :key="s.id" :value="s.id">
            {{ s.first_name }} {{ s.last_name }} ({{ s.job_position?.title || 'Sales' }})
          </option>
        </select>
        <div class="grid grid-cols-3 gap-1.5 sm:gap-2 w-full lg:w-auto">
          <button
            @click="selectedStatus = 'ACTIVE'"
            class="px-3 sm:px-4 py-2.5 sm:py-3 rounded-xl sm:rounded-2xl text-xs sm:text-sm font-bold transition-all cursor-pointer text-center flex items-center justify-center gap-1.5"
            :class="
              selectedStatus === 'ACTIVE'
                ? 'bg-emerald-600 text-white shadow-md sm:shadow-lg shadow-emerald-600/20 ring-2 ring-emerald-500'
                : 'bg-slate-50 text-slate-600 hover:bg-slate-100 border border-slate-200/60'
            "
          >
            <span class="w-2 h-2 rounded-full bg-white animate-pulse" v-if="selectedStatus === 'ACTIVE'"></span>
            Aktif
          </button>
          <button
            @click="selectedStatus = 'INACTIVE'"
            class="px-3 sm:px-4 py-2.5 sm:py-3 rounded-xl sm:rounded-2xl text-xs sm:text-sm font-bold transition-all cursor-pointer text-center flex items-center justify-center gap-1.5"
            :class="
              selectedStatus === 'INACTIVE'
                ? 'bg-amber-600 text-white shadow-md sm:shadow-lg shadow-amber-600/20 ring-2 ring-amber-500'
                : 'bg-slate-50 text-slate-600 hover:bg-slate-100 border border-slate-200/60'
            "
          >
            <span class="w-2 h-2 rounded-full bg-white" v-if="selectedStatus === 'INACTIVE'"></span>
            Inaktif
          </button>
          <button
            @click="selectedStatus = 'ALL'"
            class="px-3 sm:px-4 py-2.5 sm:py-3 rounded-xl sm:rounded-2xl text-xs sm:text-sm font-bold transition-all cursor-pointer text-center"
            :class="
              selectedStatus === 'ALL'
                ? 'bg-slate-900 text-white shadow-md sm:shadow-lg shadow-slate-900/20 ring-2 ring-slate-900'
                : 'bg-slate-50 text-slate-600 hover:bg-slate-100 border border-slate-200/60'
            "
          >
            Semua
          </button>
        </div>
      </div>

      <!-- Bottom Row: Tombol Quick Filter Kelengkapan Data -->
      <div class="pt-3 border-t border-slate-100 flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3 flex-wrap">
        <div class="flex items-center gap-2 flex-wrap w-full sm:w-auto">
          <span class="text-[11px] sm:text-xs font-black text-slate-400 uppercase tracking-wider mr-1 w-full sm:w-auto">Filter Kelengkapan:</span>

          <!-- Tombol Semua Data -->
          <button
            @click="setCompletenessFilter('ALL')"
            class="px-3.5 py-2 sm:px-4 sm:py-2.5 rounded-xl sm:rounded-2xl text-xs font-black transition-all flex items-center gap-1.5 cursor-pointer shadow-sm"
            :class="
              selectedCompleteness === 'ALL'
                ? 'bg-slate-900 text-white shadow-slate-900/20 ring-2 ring-slate-900'
                : 'bg-slate-50 text-slate-600 hover:bg-slate-100 border border-slate-200/60'
            "
          >
            Semua Data
          </button>

          <!-- Tombol Data Kurang Lengkap -->
          <button
            @click="setCompletenessFilter('INCOMPLETE')"
            class="px-3.5 py-2 sm:px-4 sm:py-2.5 rounded-xl sm:rounded-2xl text-xs font-black transition-all flex items-center gap-1.5 cursor-pointer shadow-sm"
            :class="
              isIncompleteActive
                ? 'bg-amber-500 text-white shadow-amber-500/25 ring-2 ring-amber-500'
                : 'bg-amber-50 text-amber-800 hover:bg-amber-100 border border-amber-200'
            "
          >
            <AlertTriangle class="w-3.5 h-3.5" />
            Data Kurang Lengkap
          </button>

          <!-- Sub-Dropdown saat Data Kurang Lengkap aktif -->
          <select
            v-if="isIncompleteActive"
            v-model="selectedCompleteness"
            class="px-3 py-2 bg-amber-50 border border-amber-300 rounded-xl sm:rounded-2xl text-xs font-bold text-amber-900 outline-none focus:ring-2 focus:ring-amber-500 transition-all cursor-pointer w-full sm:w-auto"
          >
            <option value="INCOMPLETE">Semua Jenis Kekurangan</option>
            <option value="MISSING_PHONE">Tanpa No. Telepon / WA</option>
            <option value="MISSING_LOCATION">Tanpa Titik Lokasi GPS</option>
            <option value="MISSING_PHOTO">Tanpa Foto Toko</option>
            <option value="MISSING_OWNER">Tanpa Nama Pemilik</option>
            <option value="MISSING_SALESMAN">Tanpa Salesman PIC</option>
            <option value="MISSING_ADDRESS">Tanpa Alamat / Wilayah</option>
          </select>

          <!-- Tombol Data Lengkap -->
          <button
            @click="setCompletenessFilter('COMPLETE')"
            class="px-3.5 py-2 sm:px-4 sm:py-2.5 rounded-xl sm:rounded-2xl text-xs font-black transition-all flex items-center gap-1.5 cursor-pointer shadow-sm"
            :class="
              selectedCompleteness === 'COMPLETE'
                ? 'bg-emerald-600 text-white shadow-emerald-600/25 ring-2 ring-emerald-600'
                : 'bg-emerald-50 text-emerald-800 hover:bg-emerald-100 border border-emerald-200'
            "
          >
            <CheckCircle2 class="w-3.5 h-3.5" />
            Data Lengkap
          </button>
        </div>
      </div>
    </div>

    <!-- Store DataTable & Mobile Cards -->
    <div
      class="bg-white rounded-2xl sm:rounded-[32px] border border-slate-200 shadow-sm overflow-hidden flex flex-col relative"
    >
      <!-- Background Sync Indicator & Offline Cache Banner -->
      <div
        v-if="isBackgroundSyncing"
        class="bg-indigo-50/90 border-b border-indigo-100 px-4 py-2 flex items-center justify-between text-xs text-indigo-700 font-semibold transition-all animate-pulse"
      >
        <div class="flex items-center gap-2">
          <Loader2 class="w-3.5 h-3.5 animate-spin text-indigo-600" />
          <span>Memperbarui data toko terbaru di latar belakang...</span>
        </div>
      </div>
      <div
        v-else-if="isUsingOfflineCache"
        class="bg-amber-50/90 border-b border-amber-200 px-4 py-2 flex items-center justify-between text-xs text-amber-800 font-semibold transition-all"
      >
        <div class="flex items-center gap-2">
          <AlertCircle class="w-4 h-4 text-amber-600 shrink-0" />
          <span>Menampilkan data toko tersimpan (Mode Hemat Jaringan).</span>
        </div>
        <button
          @click="fetchStores(false)"
          class="text-[11px] font-bold text-amber-900 underline hover:text-amber-950 cursor-pointer"
        >
          Coba Sinkron Ulang
        </button>
      </div>

      <!-- DESKTOP TABLE VIEW (hidden on mobile, visible on lg+) -->
      <div class="hidden lg:block overflow-x-auto no-scrollbar">
        <table class="w-full border-collapse text-left">
          <thead class="bg-slate-50/50 border-b border-slate-100">
            <tr>
              <th class="p-6 text-[10px] font-black text-slate-400 uppercase tracking-widest">
                Toko / Outlet
              </th>
              <th class="p-6 text-[10px] font-black text-slate-400 uppercase tracking-widest whitespace-nowrap">
                Waktu Registrasi
              </th>
              <th class="p-6 text-[10px] font-black text-slate-400 uppercase tracking-widest text-center whitespace-nowrap">
                Status EC (Hari Ini)
              </th>
              <th class="p-6 text-[10px] font-black text-slate-400 uppercase tracking-widest">
                Pemilik
              </th>
              <th class="p-6 text-[10px] font-black text-slate-400 uppercase tracking-widest">
                Kontak
              </th>
              <th class="p-6 text-[10px] font-black text-slate-400 uppercase tracking-widest">
                Salesman
              </th>
              <th class="p-6 text-[10px] font-black text-slate-400 uppercase tracking-widest">
                Wilayah / Alamat
              </th>
              <th
                class="p-6 text-[10px] font-black text-slate-400 uppercase tracking-widest text-right"
              >
                Total Piutang
              </th>
              <th
                class="p-6 text-[10px] font-black text-slate-400 uppercase tracking-widest text-center"
              >
                Status
              </th>
              <th
                class="p-6 text-[10px] font-black text-slate-400 uppercase tracking-widest text-center"
              >
                Verifikasi
              </th>
              <th
                class="p-6 text-[10px] font-black text-slate-400 uppercase tracking-widest text-right"
              >
                Aksi
              </th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-50">
            <tr v-if="isLoading" v-for="i in 5" :key="i" class="animate-pulse">
              <td v-for="j in 11" :key="j" class="p-6">
                <div class="h-4 bg-slate-100 rounded-lg w-full"></div>
              </td>
            </tr>

            <tr v-else-if="storeFetchError && storesList.length === 0">
              <td colspan="11" class="py-20 text-center">
                <div class="flex flex-col items-center gap-3 max-w-md mx-auto">
                  <div class="w-12 h-12 rounded-2xl bg-red-50 flex items-center justify-center text-red-500">
                    <AlertTriangle class="w-6 h-6" />
                  </div>
                  <div>
                    <h3 class="font-bold text-slate-800 text-sm">Gagal Memuat Data Toko</h3>
                    <p class="text-xs text-slate-500 mt-1">{{ storeFetchError }}</p>
                  </div>
                  <Button
                    @click="fetchStores"
                    size="sm"
                    variant="outline"
                    class="mt-2 font-bold text-xs gap-2 border-slate-200 hover:bg-slate-50 cursor-pointer"
                  >
                    <RefreshCw class="w-3.5 h-3.5" />
                    Coba Lagi
                  </Button>
                </div>
              </td>
            </tr>

            <tr v-else-if="storesList.length === 0">
              <td colspan="11" class="py-24 text-center">
                <div class="flex flex-col items-center gap-3">
                  <Store class="w-12 h-12 text-slate-200" />
                  <p class="font-bold text-slate-400">Tidak ada toko ditemukan.</p>
                </div>
              </td>
            </tr>

            <tr
              v-for="store in storesList"
              :key="store.id"
              class="group hover:bg-slate-50/50 transition-all"
            >
              <td class="p-6">
                <div class="flex items-center gap-4">
                  <div
                    v-if="store.photo_url"
                    class="w-10 h-10 rounded-xl overflow-hidden border border-slate-100 shrink-0 cursor-pointer hover:ring-2 hover:ring-primary/50 transition-all"
                    @click="openImagePreview(store.photo_url)"
                  >
                    <img :src="store.photo_url" class="w-full h-full object-cover" />
                  </div>
                  <div
                    class="w-10 h-10 rounded-xl bg-slate-100 flex items-center justify-center shrink-0"
                    v-else
                  >
                    <Store class="w-5 h-5 text-slate-400" />
                  </div>
                  <div>
                    <p class="text-sm font-black text-slate-900 leading-none">{{ store.name }}</p>
                    <div class="flex items-center gap-2 mt-1">
                      <p class="text-[10px] font-bold text-slate-400 uppercase tracking-tighter">
                        ID: {{ store.id.split('-')[0] }}
                      </p>
                      <span
                        :class="[
                          'text-[8px] font-black px-2 py-0.5 rounded-md uppercase tracking-widest',
                          store.category === 'TOKO_BARU'
                            ? 'bg-blue-50 text-blue-600'
                            : 'bg-amber-50 text-amber-600',
                        ]"
                      >
                        {{ store.category === 'TOKO_BARU' ? 'BARU' : 'LAMA' }}
                      </span>
                      <span
                        v-if="store.customer_type"
                        class="text-[8px] font-black px-2 py-0.5 rounded-md bg-purple-50 text-purple-700 uppercase tracking-wider border border-purple-100"
                      >
                        {{ formatCustomerTypeBadge(store) }}
                      </span>
                      <!-- Badge Frekuensi Kunjungan -->
                      <span
                        class="text-[8px] font-black px-2 py-0.5 rounded-md uppercase tracking-wider border"
                        :class="
                          store.visit_frequency === 'F1'
                            ? 'bg-emerald-50 text-emerald-700 border-emerald-200'
                            : store.visit_frequency === 'F2'
                              ? 'bg-cyan-50 text-cyan-700 border-cyan-200'
                              : 'bg-indigo-50 text-indigo-700 border-indigo-200'
                        "
                        :title="`Jadwal: ${formatVisitDaysSummary(store.visit_days) || 'Semua hari'}`"
                      >
                        {{ store.visit_frequency || 'F1' }}<span v-if="formatVisitDaysSummary(store.visit_days)" class="ml-1 text-[7px] opacity-75">({{ formatVisitDaysSummary(store.visit_days) }})</span>
                      </span>
                    </div>

                    <!-- Data Completeness Status Badge -->
                    <div class="mt-1.5 flex items-center">
                      <div
                        v-if="getMissingFields(store).length > 0"
                        class="missing-fields-tooltip-wrapper inline-flex"
                      >
                        <span
                          class="inline-flex items-center gap-1 text-[9px] font-black px-2 py-0.5 rounded-md bg-amber-50 text-amber-700 border border-amber-200/80 cursor-help hover:bg-amber-100 transition-all"
                        >
                          <AlertTriangle class="w-2.5 h-2.5 text-amber-600 shrink-0" />
                          Data Kurang ({{ getMissingFields(store).length }})
                        </span>
                        <div class="missing-fields-tooltip">
                          <div class="missing-fields-tooltip-arrow"></div>
                          <p class="text-[10px] font-black text-amber-400 mb-1 flex items-center gap-1">
                            <AlertCircle class="w-3 h-3 text-amber-400 shrink-0" />
                            Data Perlu Dilengkapi:
                          </p>
                          <ul class="text-[10px] font-semibold text-slate-200 space-y-0.5 list-disc pl-3">
                            <li v-for="m in getMissingFields(store)" :key="m.key">{{ m.label }}</li>
                          </ul>
                        </div>
                      </div>
                      <span
                        v-else
                        class="inline-flex items-center gap-1 text-[9px] font-black px-2 py-0.5 rounded-md bg-emerald-50 text-emerald-700 border border-emerald-200/80"
                      >
                        <CheckCircle2 class="w-2.5 h-2.5 text-emerald-600 shrink-0" />
                        Data Lengkap
                      </span>
                    </div>
                  </div>
                </div>
              </td>
              <!-- Kolom Waktu Registrasi -->
              <td class="p-6 whitespace-nowrap">
                <div class="flex flex-col">
                  <div class="flex items-center gap-1.5 text-xs font-bold text-slate-800">
                    <Calendar class="w-3.5 h-3.5 text-slate-400 shrink-0" />
                    <span>{{ formatDate(store.created_at) }}</span>
                  </div>
                  <div class="flex items-center gap-1 text-[11px] font-medium text-slate-500 mt-0.5">
                    <Clock class="w-3 h-3 text-slate-400 shrink-0" />
                    <span>{{ formatTime(store.created_at) }}</span>
                  </div>
                  <div v-if="store.is_noo_today" class="mt-1.5">
                    <span class="inline-flex items-center gap-1 text-[9px] font-black px-2 py-0.5 rounded-md bg-blue-600 text-white shadow-sm shadow-blue-500/20 uppercase tracking-wider">
                      <Sparkles class="w-2.5 h-2.5" /> NOO HARI INI
                    </span>
                  </div>
                </div>
              </td>
              <!-- Kolom Status EC / Pesanan -->
              <td class="p-6 text-center whitespace-nowrap">
                <div class="flex flex-col items-center justify-center gap-1">
                  <div v-if="store.is_effective_call || (store.today_order_count && store.today_order_count > 0)">
                    <span class="inline-flex items-center gap-1 px-2.5 py-1 text-[10px] font-black bg-emerald-100 text-emerald-800 rounded-full uppercase tracking-wider border border-emerald-200">
                      <ShoppingBag class="w-3 h-3 text-emerald-700 shrink-0" />
                      EC (ADA PESANAN)
                    </span>
                    <p class="text-[10px] font-black text-emerald-700 mt-1">
                      {{ store.today_order_count || 1 }} Order • {{ formatCurrency(store.today_order_amount || 0) }}
                    </p>
                  </div>
                  <div v-else>
                    <span class="inline-flex items-center gap-1 px-2.5 py-1 text-[10px] font-bold bg-slate-100 text-slate-600 rounded-full uppercase tracking-wider border border-slate-200">
                      Non-EC
                    </span>
                    <p class="text-[9px] font-semibold text-slate-400 mt-0.5">Belum Ada Order</p>
                  </div>
                </div>
              </td>
              <td class="p-6">
                <p class="text-sm font-bold text-slate-700">{{ store.owner_name || '-' }}</p>
              </td>
              <td class="p-6">
                <div class="flex items-center gap-2">
                  <Phone class="w-3.5 h-3.5 text-slate-300" />
                  <p class="text-sm font-bold text-slate-700">{{ store.phone_number || '-' }}</p>
                </div>
              </td>
              <td class="p-6">
                <div class="flex items-center gap-2">
                  <div
                    class="w-7 h-7 rounded-lg bg-primary/10 flex items-center justify-center text-primary shrink-0"
                  >
                    <User class="w-3.5 h-3.5" />
                  </div>
                  <p class="text-sm font-bold text-slate-700">
                    {{ store.assigned_employee?.first_name || 'Unassigned' }}
                  </p>
                </div>
              </td>
              <td class="p-6 max-w-xs">
                <div class="address-tooltip-wrapper">
                  <div class="flex items-center gap-1.5 cursor-default">
                    <MapPin class="w-3 h-3 text-slate-300 shrink-0" />
                    <p class="text-xs font-medium text-slate-500 line-clamp-1 leading-relaxed">
                      {{ store.address || '-' }}
                    </p>
                  </div>
                  <div v-if="store.address" class="address-tooltip">
                    <div class="address-tooltip-arrow"></div>
                    <p class="text-[11px] font-semibold text-white leading-relaxed">
                      {{ store.address }}
                    </p>
                  </div>
                  <p v-if="store.province" class="text-[10px] font-bold text-primary mt-1">
                    {{ store.village ? store.village + ', ' : '' }}{{ store.district ? store.district + ', ' : '' }}{{ store.regency ? store.regency + ', ' : '' }}{{ store.province }}
                  </p>
                </div>
              </td>
              <td class="p-6 text-right">
                <div class="flex flex-col items-end">
                  <p
                    :class="[
                      'text-sm font-black',
                      store.total_receivable > 0 ? 'text-rose-600' : 'text-slate-900',
                    ]"
                  >
                    {{ formatCurrency(store.total_receivable || 0) }}
                  </p>
                  <div v-if="store.overdue_count > 0" class="flex items-center gap-1 mt-1">
                    <AlertTriangle class="w-3 h-3 text-rose-500 animate-pulse" />
                    <p class="text-[9px] font-bold text-rose-500 uppercase">
                      {{ store.overdue_count }} Nota Jatuh Tempo
                    </p>
                  </div>
                  <p
                    v-else-if="store.total_receivable > 0"
                    class="text-[9px] font-bold text-orange-500 uppercase mt-1"
                  >
                    Belum Lunas
                  </p>
                  <p v-else class="text-[9px] font-bold text-emerald-500 uppercase mt-1">Lunas</p>
                </div>
              </td>
              <td class="p-6">
                <div class="flex justify-center">
                  <span
                    :class="[
                      'text-[10px] font-black px-3 py-1 rounded-full uppercase tracking-wider border',
                      store.is_active
                        ? 'bg-emerald-50 text-emerald-600 border-emerald-100'
                        : 'bg-red-50 text-red-600 border-red-100',
                    ]"
                  >
                    {{ store.is_active ? 'AKTIF' : 'NON-AKTIF' }}
                  </span>
                </div>
              </td>
              <td class="p-6">
                <div class="flex flex-col items-center justify-center gap-1">
                  <span v-if="store.verification_status === 'VERIFIED_REAL'" 
                        class="px-2.5 py-1 text-[10px] font-black bg-emerald-100 text-emerald-700 rounded-full flex items-center gap-1 uppercase tracking-wider border border-emerald-200">
                    <CheckCircle2 class="w-3 h-3" /> TERVERIFIKASI
                  </span>
                  <span v-else-if="store.verification_status === 'FICTITIOUS'" 
                        class="px-2.5 py-1 text-[10px] font-black bg-red-100 text-red-700 rounded-full flex items-center gap-1 uppercase tracking-wider border border-red-200">
                    <AlertCircle class="w-3 h-3" /> FIKTIF
                  </span>
                  <span v-else 
                        class="px-2.5 py-1 text-[10px] font-black bg-gray-100 text-gray-600 rounded-full uppercase tracking-wider border border-gray-200">
                    BELUM VERIF
                  </span>
                  <p v-if="store.verification_status === 'VERIFIED_REAL' && store.latest_verifier_role" class="text-[9px] font-bold text-slate-400 capitalize">
                    oleh {{ store.latest_verifier_role }}
                  </p>
                </div>
              </td>
              <td class="p-6 text-right">
                <div class="flex justify-end gap-1">
                  <button
                    v-if="getMissingFields(store).length > 0"
                    @click="followUpSalesViaWhatsApp(store)"
                    class="p-2 hover:bg-emerald-50 rounded-xl transition-all text-slate-400 hover:text-emerald-600 relative group/wa cursor-pointer"
                    :title="store.assigned_employee?.phone_number ? `Follow Up ${store.assigned_employee.first_name || 'Sales'} via WhatsApp` : 'Salin Pesan Follow Up Sales (Data Kurang)'"
                  >
                    <MessageSquare class="w-5 h-5 text-emerald-600" />
                  </button>
                  <button
                    @click="openAuditModal(store)"
                    class="p-2 hover:bg-blue-50 rounded-xl transition-all text-slate-400 hover:text-blue-600 cursor-pointer"
                    title="Verifikasi Keaslian Toko"
                  >
                    <ShieldCheck class="w-5 h-5" />
                  </button>
                  <button
                    @click="openAuditHistoryModal(store)"
                    class="p-2 hover:bg-emerald-50 rounded-xl transition-all text-slate-400 hover:text-emerald-600 cursor-pointer"
                    title="Riwayat Verifikasi"
                  >
                    <Activity class="w-5 h-5" />
                  </button>
                  <button
                    @click="openHistoryModal(store)"
                    class="p-2 hover:bg-indigo-50 rounded-xl transition-all text-slate-400 hover:text-indigo-600 cursor-pointer"
                    title="Riwayat Pesanan"
                  >
                    <History class="w-5 h-5" />
                  </button>
                  <button
                    @click="handleEdit(store)"
                    class="p-2 hover:bg-slate-100 rounded-xl transition-all text-slate-400 hover:text-primary cursor-pointer"
                    title="Edit Toko"
                  >
                    <Pencil class="w-5 h-5" />
                  </button>
                  <button
                    @click="handleDelete(store.id)"
                    class="p-2 hover:bg-red-50 rounded-xl transition-all text-red-400 hover:text-red-600 cursor-pointer"
                    title="Hapus Toko"
                  >
                    <Trash2 class="w-5 h-5" />
                  </button>
                  <a
                    :href="`https://www.google.com/maps/search/?api=1&query=${store.latitude},${store.longitude}`"
                    target="_blank"
                    class="p-2 hover:bg-slate-100 rounded-xl transition-all text-slate-400 hover:text-blue-500 inline-flex"
                    title="Lihat Lokasi"
                  >
                    <Navigation class="w-5 h-5" />
                  </a>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- MOBILE CARD LIST VIEW (visible on mobile <lg, hidden on desktop) -->
      <div class="block lg:hidden divide-y divide-slate-100">
        <!-- Skeleton loader for mobile -->
        <div v-if="isLoading" class="p-4 space-y-3 animate-pulse">
          <div v-for="i in 3" :key="i" class="bg-slate-50 p-4 rounded-2xl border border-slate-100 space-y-3">
            <div class="h-4 bg-slate-200 rounded w-1/2"></div>
            <div class="h-3 bg-slate-200 rounded w-3/4"></div>
            <div class="h-8 bg-slate-200 rounded w-full"></div>
          </div>
        </div>

        <!-- Error state for mobile -->
        <div v-else-if="storeFetchError && storesList.length === 0" class="py-12 px-4 text-center">
          <div class="flex flex-col items-center gap-3 max-w-xs mx-auto">
            <div class="w-12 h-12 rounded-2xl bg-red-50 flex items-center justify-center text-red-500">
              <AlertTriangle class="w-6 h-6" />
            </div>
            <div>
              <h3 class="font-bold text-slate-800 text-sm">Gagal Memuat Data Toko</h3>
              <p class="text-xs text-slate-500 mt-1">{{ storeFetchError }}</p>
            </div>
            <Button
              @click="fetchStores"
              size="sm"
              variant="outline"
              class="mt-2 font-bold text-xs gap-2 border-slate-200 hover:bg-slate-50 cursor-pointer"
            >
              <RefreshCw class="w-3.5 h-3.5" />
              Coba Lagi
            </Button>
          </div>
        </div>

        <!-- Empty state for mobile -->
        <div v-else-if="storesList.length === 0" class="py-16 px-4 text-center">
          <div class="flex flex-col items-center gap-3">
            <Store class="w-12 h-12 text-slate-200" />
            <p class="font-bold text-slate-400 text-sm">Tidak ada toko ditemukan.</p>
          </div>
        </div>

        <!-- Store Cards -->
        <div
          v-else
          v-for="store in storesList"
          :key="store.id"
          class="p-4 sm:p-5 space-y-3.5 hover:bg-slate-50/50 transition-colors"
        >
          <!-- Top Info: Photo, Name, Badges -->
          <div class="flex items-start gap-3">
            <div
              v-if="store.photo_url"
              class="w-12 h-12 rounded-2xl overflow-hidden border border-slate-100 shrink-0 cursor-pointer ring-1 ring-slate-200"
              @click="openImagePreview(store.photo_url)"
            >
              <img :src="store.photo_url" class="w-full h-full object-cover" />
            </div>
            <div
              v-else
              class="w-12 h-12 rounded-2xl bg-slate-100 flex items-center justify-center shrink-0 border border-slate-200/60 text-slate-400"
            >
              <Store class="w-6 h-6" />
            </div>

            <div class="flex-1 min-w-0">
              <div class="flex items-center justify-between gap-2">
                <h3 class="text-sm font-black text-slate-900 truncate leading-snug">{{ store.name }}</h3>
                <span
                  :class="[
                    'text-[9px] font-black px-2 py-0.5 rounded-full uppercase tracking-wider border shrink-0',
                    store.is_active
                      ? 'bg-emerald-50 text-emerald-600 border-emerald-100'
                      : 'bg-red-50 text-red-600 border-red-100',
                  ]"
                >
                  {{ store.is_active ? 'AKTIF' : 'NON-AKTIF' }}
                </span>
              </div>

              <div class="flex flex-wrap items-center gap-1.5 mt-1">
                <span class="text-[9px] font-bold text-slate-400 uppercase">
                  ID: {{ store.id.split('-')[0] }}
                </span>
                <span class="text-slate-300">•</span>
                <span
                  :class="[
                    'text-[8px] font-black px-1.5 py-0.5 rounded uppercase tracking-wider',
                    store.category === 'TOKO_BARU'
                      ? 'bg-blue-50 text-blue-600'
                      : 'bg-amber-50 text-amber-600',
                  ]"
                >
                  {{ store.category === 'TOKO_BARU' ? 'BARU' : 'LAMA' }}
                </span>
                <span
                  v-if="store.customer_type"
                  class="text-[8px] font-black px-1.5 py-0.5 rounded bg-purple-50 text-purple-700 uppercase tracking-wider border border-purple-100"
                >
                  {{ formatCustomerTypeBadge(store) }}
                </span>
                <span class="text-slate-300">•</span>
                <!-- Badge Frekuensi Kunjungan Mobile -->
                <span
                  class="text-[8px] font-black px-1.5 py-0.5 rounded uppercase tracking-wider border"
                  :class="
                    store.visit_frequency === 'F1'
                      ? 'bg-emerald-50 text-emerald-700 border-emerald-200'
                      : store.visit_frequency === 'F2'
                        ? 'bg-cyan-50 text-cyan-700 border-cyan-200'
                        : 'bg-indigo-50 text-indigo-700 border-indigo-200'
                  "
                >
                  {{ store.visit_frequency || 'F1' }}<span v-if="formatVisitDaysSummary(store.visit_days)" class="ml-1 text-[7px] opacity-75">({{ formatVisitDaysSummary(store.visit_days) }})</span>
                </span>
                <span class="text-slate-300">•</span>
                <span v-if="store.is_noo_today" class="text-[8px] font-black px-1.5 py-0.5 rounded bg-blue-600 text-white flex items-center gap-0.5 uppercase">
                  <Sparkles class="w-2.5 h-2.5" /> NOO Hari Ini
                </span>
                <span class="text-slate-300">•</span>
                <!-- EC status badge on mobile -->
                <span
                  v-if="store.is_effective_call || (store.today_order_count && store.today_order_count > 0)"
                  class="text-[8px] font-black px-1.5 py-0.5 rounded bg-emerald-100 text-emerald-800 flex items-center gap-0.5 uppercase border border-emerald-200"
                >
                  <ShoppingBag class="w-2.5 h-2.5 text-emerald-700" /> EC: {{ store.today_order_count || 1 }} Order
                </span>
                <span
                  v-else
                  class="text-[8px] font-semibold px-1.5 py-0.5 rounded bg-slate-100 text-slate-500 uppercase"
                >
                  Non-EC
                </span>
                <span class="text-slate-300">•</span>
                <!-- Verification status badge -->
                <span
                  v-if="store.verification_status === 'VERIFIED_REAL'"
                  class="text-[8px] font-black px-1.5 py-0.5 rounded bg-emerald-100 text-emerald-700 flex items-center gap-0.5 uppercase"
                >
                  <CheckCircle2 class="w-2.5 h-2.5" /> Terverifikasi
                </span>
                <span
                  v-else-if="store.verification_status === 'FICTITIOUS'"
                  class="text-[8px] font-black px-1.5 py-0.5 rounded bg-red-100 text-red-700 flex items-center gap-0.5 uppercase"
                >
                  <AlertCircle class="w-2.5 h-2.5" /> Fiktif
                </span>
                <span
                  v-else
                  class="text-[8px] font-black px-1.5 py-0.5 rounded bg-slate-100 text-slate-600 uppercase"
                >
                  Belum Verif
                </span>
              </div>
            </div>
          </div>

          <!-- Completeness status if missing fields -->
          <div
            v-if="getMissingFields(store).length > 0"
            class="bg-amber-50/80 border border-amber-200/80 rounded-xl p-2.5 text-xs text-amber-900"
          >
            <div class="flex items-center gap-1.5 font-bold text-[11px] text-amber-800">
              <AlertTriangle class="w-3.5 h-3.5 text-amber-600 shrink-0" />
              <span>Data Kurang ({{ getMissingFields(store).length }} item):</span>
            </div>
            <p class="text-[10px] text-amber-700 mt-1 pl-5">
              {{ getMissingFields(store).map(m => m.label).join(', ') }}
            </p>
          </div>

          <!-- Store Info Grid -->
          <div class="grid grid-cols-2 gap-2.5 bg-slate-50/70 p-3 rounded-2xl border border-slate-100 text-xs">
            <div>
              <span class="text-[9px] font-black text-slate-400 uppercase tracking-wider block">Waktu Registrasi</span>
              <span class="font-bold text-slate-800 truncate block mt-0.5 flex items-center gap-1">
                <Clock class="w-3 h-3 text-slate-400 shrink-0" />
                {{ formatDate(store.created_at) }}, {{ formatTime(store.created_at) }}
              </span>
            </div>
            <div>
              <span class="text-[9px] font-black text-slate-400 uppercase tracking-wider block">Pesanan Hari Ini</span>
              <span class="font-bold text-emerald-700 truncate block mt-0.5">
                {{ store.is_effective_call || (store.today_order_count && store.today_order_count > 0) ? `${store.today_order_count || 1} Order (${formatCurrency(store.today_order_amount || 0)})` : 'Belum Ada' }}
              </span>
            </div>
            <div>
              <span class="text-[9px] font-black text-slate-400 uppercase tracking-wider block">Pemilik</span>
              <span class="font-bold text-slate-700 truncate block mt-0.5">{{ store.owner_name || '-' }}</span>
            </div>
            <div>
              <span class="text-[9px] font-black text-slate-400 uppercase tracking-wider block">Telepon / WA</span>
              <span class="font-bold text-slate-700 truncate block mt-0.5">{{ store.phone_number || '-' }}</span>
            </div>
            <div>
              <span class="text-[9px] font-black text-slate-400 uppercase tracking-wider block">Salesman</span>
              <span class="font-bold text-slate-700 truncate block mt-0.5">{{ store.assigned_employee?.first_name || 'Unassigned' }}</span>
            </div>
            <div>
              <span class="text-[9px] font-black text-slate-400 uppercase tracking-wider block">Total Piutang</span>
              <span
                :class="[
                  'font-black block mt-0.5',
                  store.total_receivable > 0 ? 'text-rose-600' : 'text-slate-800',
                ]"
              >
                {{ formatCurrency(store.total_receivable || 0) }}
              </span>
            </div>
            <div class="col-span-2 pt-1 border-t border-slate-200/60">
              <span class="text-[9px] font-black text-slate-400 uppercase tracking-wider block">Alamat / Wilayah</span>
              <p class="font-medium text-slate-600 text-[11px] mt-0.5 line-clamp-2 leading-relaxed">
                {{ store.address || '-' }}
              </p>
              <p v-if="store.province" class="text-[10px] font-bold text-primary mt-0.5">
                {{ store.village ? store.village + ', ' : '' }}{{ store.district ? store.district + ', ' : '' }}{{ store.regency ? store.regency + ', ' : '' }}{{ store.province }}
              </p>
            </div>
          </div>

          <!-- Mobile Action Toolbar (All 7 buttons accessible and responsive) -->
          <div class="pt-1 flex flex-wrap items-center gap-1.5 justify-end">
            <!-- WA Followup -->
            <button
              v-if="getMissingFields(store).length > 0"
              @click="followUpSalesViaWhatsApp(store)"
              class="px-2.5 py-2 bg-emerald-50 hover:bg-emerald-100 text-emerald-700 rounded-xl text-xs font-bold transition-all flex items-center gap-1.5 border border-emerald-200 cursor-pointer"
              title="Follow Up WA"
            >
              <MessageSquare class="w-4 h-4 text-emerald-600 shrink-0" />
              <span class="text-[11px]">Follow Up WA</span>
            </button>

            <!-- Verifikasi -->
            <button
              @click="openAuditModal(store)"
              class="px-2.5 py-2 bg-blue-50 hover:bg-blue-100 text-blue-700 rounded-xl text-xs font-bold transition-all flex items-center gap-1.5 border border-blue-200 cursor-pointer"
              title="Verifikasi Keaslian Toko"
            >
              <ShieldCheck class="w-4 h-4 text-blue-600 shrink-0" />
              <span class="text-[11px]">Verifikasi</span>
            </button>

            <!-- Riwayat Verifikasi -->
            <button
              @click="openAuditHistoryModal(store)"
              class="p-2 bg-emerald-50 hover:bg-emerald-100 text-emerald-700 rounded-xl transition-all border border-emerald-200 cursor-pointer"
              title="Riwayat Verifikasi"
            >
              <Activity class="w-4 h-4" />
            </button>

            <!-- Riwayat Pesanan -->
            <button
              @click="openHistoryModal(store)"
              class="p-2 bg-indigo-50 hover:bg-indigo-100 text-indigo-700 rounded-xl transition-all border border-indigo-200 cursor-pointer"
              title="Riwayat Pesanan"
            >
              <History class="w-4 h-4" />
            </button>

            <!-- Edit -->
            <button
              @click="handleEdit(store)"
              class="p-2 bg-slate-100 hover:bg-slate-200 text-slate-700 rounded-xl transition-all border border-slate-200 cursor-pointer"
              title="Edit Toko"
            >
              <Pencil class="w-4 h-4" />
            </button>

            <!-- Hapus -->
            <button
              @click="handleDelete(store.id)"
              class="p-2 bg-red-50 hover:bg-red-100 text-red-600 rounded-xl transition-all border border-red-200 cursor-pointer"
              title="Hapus Toko"
            >
              <Trash2 class="w-4 h-4" />
            </button>

            <!-- Maps -->
            <a
              :href="`https://www.google.com/maps/search/?api=1&query=${store.latitude},${store.longitude}`"
              target="_blank"
              class="p-2 bg-slate-100 hover:bg-slate-200 text-blue-600 rounded-xl transition-all border border-slate-200 inline-flex items-center justify-center"
              title="Buka Google Maps"
            >
              <Navigation class="w-4 h-4" />
            </a>
          </div>
        </div>
      </div>

      <!-- Pagination Controls -->
      <div
        v-if="totalStores > 0"
        class="flex flex-col sm:flex-row sm:items-center justify-between gap-3 sm:gap-4 bg-white p-4 sm:p-6 border-t border-slate-100"
      >
        <div class="flex items-center justify-between sm:justify-start gap-2 w-full sm:w-auto">
          <span class="text-xs font-bold text-slate-500">Baris per hal:</span>
          <select
            v-model="itemsPerPage"
            @change="currentPage = 1"
            class="bg-slate-50 border border-slate-200 rounded-lg text-xs font-bold px-2 py-1 outline-none cursor-pointer"
          >
            <option :value="10">10</option>
            <option :value="20">20</option>
            <option :value="50">50</option>
            <option :value="100">100</option>
          </select>
        </div>
        <span class="text-xs font-bold text-slate-500 text-center sm:text-left">
          Menampilkan {{ (currentPage - 1) * itemsPerPage + 1 }} -
          {{ Math.min(currentPage * itemsPerPage, totalStores) }} dari {{ totalStores }} data
        </span>
        <div class="flex items-center justify-center gap-2 w-full sm:w-auto">
          <button
            @click="prevPage"
            :disabled="currentPage === 1"
            class="px-3 sm:px-4 py-1.5 sm:py-2 rounded-xl text-xs font-black border border-slate-200 disabled:opacity-50 hover:bg-slate-50 transition-colors cursor-pointer"
          >
            Prev
          </button>
          <span class="text-xs font-black text-slate-700 px-2"
            >Hal {{ currentPage }} / {{ totalPages }}</span
          >
          <button
            @click="nextPage"
            :disabled="currentPage === totalPages"
            class="px-3 sm:px-4 py-1.5 sm:py-2 rounded-xl text-xs font-black border border-slate-200 disabled:opacity-50 hover:bg-slate-50 transition-colors cursor-pointer"
          >
            Next
          </button>
        </div>
      </div>
    </div>

    <!-- Store Modal -->
    <div v-if="showModal" class="fixed inset-0 z-50 flex items-center justify-center p-4">
      <div
        class="absolute inset-0 bg-slate-900/60 backdrop-blur-sm"
        @click="showModal = false"
      ></div>
      <div
        class="bg-white rounded-[32px] w-full max-w-2xl relative z-10 shadow-2xl flex flex-col max-h-[90vh] overflow-hidden animate-in fade-in zoom-in duration-300"
      >
        <!-- Header -->
        <div
          class="px-8 py-6 border-b border-slate-100 flex items-center justify-between bg-white shrink-0"
        >
          <div>
            <h3 class="text-2xl font-black text-slate-900 tracking-tight">
              {{ isEditing ? 'Edit Data Pelanggan / Toko' : 'Pendaftaran Pelanggan / Toko' }}
            </h3>
            <p class="text-[10px] font-black text-slate-400 uppercase tracking-widest mt-0.5">
              Informasi Outlet Customer & Kategori Pelanggan
            </p>
          </div>
          <button
            @click="closeModal"
            class="p-2 hover:bg-slate-100 rounded-xl transition-all text-slate-400 hover:text-slate-900"
          >
            <X class="w-6 h-6" />
          </button>
        </div>

        <!-- Body (Scrollable) -->
        <div class="p-8 overflow-y-auto space-y-6 custom-scrollbar">
          <!-- Visual Status Pendaftaran Selector: Toko Lama (Default) vs Toko Baru -->
          <div class="space-y-2">
            <label class="text-[10px] font-black text-slate-400 uppercase tracking-widest ml-1">
              Status Pendaftaran Toko
            </label>
            <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
              <button
                type="button"
                @click="currentStore.category = 'TOKO_LAMA'"
                class="p-4 rounded-2xl border-2 text-left transition-all flex items-start gap-3 cursor-pointer"
                :class="
                  currentStore.category === 'TOKO_LAMA'
                    ? 'border-primary bg-primary/5 text-primary shadow-sm ring-2 ring-primary/20'
                    : 'border-slate-200 bg-slate-50/50 text-slate-600 hover:bg-slate-100/60'
                "
              >
                <div
                  class="w-9 h-9 rounded-xl flex items-center justify-center shrink-0 transition-colors"
                  :class="
                    currentStore.category === 'TOKO_LAMA'
                      ? 'bg-primary text-white'
                      : 'bg-slate-200 text-slate-500'
                  "
                >
                  <Store class="w-4 h-4" />
                </div>
                <div>
                  <div class="flex items-center gap-1.5">
                    <p class="text-sm font-black leading-tight">Toko Lama</p>
                    <span class="text-[9px] font-black px-1.5 py-0.5 bg-amber-100 text-amber-800 rounded">Existing</span>
                  </div>
                  <p class="text-[11px] font-medium text-slate-500 mt-0.5">
                    Pelanggan / mitra lama yang sudah ada sebelumnya
                  </p>
                </div>
              </button>

              <button
                type="button"
                @click="currentStore.category = 'TOKO_BARU'"
                class="p-4 rounded-2xl border-2 text-left transition-all flex items-start gap-3 cursor-pointer"
                :class="
                  currentStore.category === 'TOKO_BARU'
                    ? 'border-blue-600 bg-blue-50/50 text-blue-700 shadow-sm ring-2 ring-blue-500/20'
                    : 'border-slate-200 bg-slate-50/50 text-slate-600 hover:bg-slate-100/60'
                "
              >
                <div
                  class="w-9 h-9 rounded-xl flex items-center justify-center shrink-0 transition-colors"
                  :class="
                    currentStore.category === 'TOKO_BARU'
                      ? 'bg-blue-600 text-white'
                      : 'bg-slate-200 text-slate-500'
                  "
                >
                  <Plus class="w-4 h-4" />
                </div>
                <div>
                  <div class="flex items-center gap-1.5">
                    <p class="text-sm font-black leading-tight">Toko Baru</p>
                    <span class="text-[9px] font-black px-1.5 py-0.5 bg-blue-100 text-blue-800 rounded">NOO</span>
                  </div>
                  <p class="text-[11px] font-medium text-slate-500 mt-0.5">
                    Outlet / pelanggan baru hasil akuisisi salesman
                  </p>
                </div>
              </button>
            </div>
          </div>

          <!-- Klasifikasi Kategori Pelanggan -->
          <div class="space-y-4 p-5 bg-slate-50/80 rounded-2xl border border-slate-100">
            <div class="space-y-2">
              <label class="text-[10px] font-black text-slate-400 uppercase tracking-widest ml-1">
                Kategori Pelanggan
              </label>
              <select
                v-model="currentStore.customer_type"
                class="w-full bg-white border-2 border-slate-200 rounded-2xl px-5 py-4 font-bold text-slate-900 focus:border-primary/40 focus:ring-0 transition-all cursor-pointer shadow-sm"
              >
                <option v-for="opt in customerTypeOptions" :key="opt.value" :value="opt.value">
                  {{ opt.label }}
                </option>
              </select>
            </div>

            <!-- Conditional HOREKA Sub-Category -->
            <div
              v-if="currentStore.customer_type === 'HOREKA'"
              class="space-y-2 animate-in fade-in slide-in-from-top-2 duration-200"
            >
              <label class="text-[10px] font-black text-amber-700 uppercase tracking-widest ml-1 flex items-center gap-1">
                <span>🍽️</span> Sub-Kategori Horeka (Hotel, Restoran, Café, dsb.)
              </label>
              <select
                v-model="currentStore.sub_customer_type"
                class="w-full bg-amber-50/70 border-2 border-amber-300 rounded-2xl px-5 py-4 font-bold text-slate-900 focus:bg-white focus:border-amber-500 focus:ring-0 transition-all cursor-pointer"
              >
                <option value="">-- Pilih Jenis Horeka --</option>
                <option v-for="sub in horekaSubTypes" :key="sub.value" :value="sub.value">
                  {{ sub.label }}
                </option>
              </select>
            </div>

            <!-- Conditional DISTRIBUTOR City -->
            <div
              v-if="currentStore.customer_type === 'DISTRIBUTOR'"
              class="space-y-2 animate-in fade-in slide-in-from-top-2 duration-200"
            >
              <label class="text-[10px] font-black text-blue-700 uppercase tracking-widest ml-1 flex items-center gap-1">
                <span>🏙️</span> Wilayah / Kota Distribusi
              </label>
              <input
                v-model="currentStore.distribution_city"
                type="text"
                placeholder="Contoh: Semarang, Solo, Surabaya, DIY..."
                class="w-full bg-blue-50/70 border-2 border-blue-300 rounded-2xl px-5 py-4 font-bold text-slate-900 focus:bg-white focus:border-blue-500 focus:ring-0 transition-all"
              />
            </div>
          </div>

          <!-- Frekuensi & Jadwal Kunjungan Sales -->
          <div class="space-y-4 p-5 bg-emerald-50/50 rounded-2xl border border-emerald-100/80">
            <div class="flex items-center justify-between">
              <label class="text-[10px] font-black text-emerald-800 uppercase tracking-widest ml-1 flex items-center gap-1.5">
                <RefreshCw class="w-3.5 h-3.5 text-emerald-600" />
                Rencana Frekuensi Kunjungan Sales
              </label>
              <span class="text-[9px] font-black px-2 py-0.5 rounded bg-emerald-100 text-emerald-800 uppercase tracking-wider">
                Visit Plan
              </span>
            </div>

            <div class="space-y-2">
              <select
                v-model="currentStore.visit_frequency"
                class="w-full bg-white border-2 border-emerald-200 rounded-2xl px-5 py-4 font-bold text-slate-900 focus:border-emerald-500 focus:ring-0 transition-all cursor-pointer shadow-sm"
              >
                <option v-for="fOpt in frequencyOptions" :key="fOpt.value" :value="fOpt.value">
                  {{ fOpt.label }}
                </option>
              </select>
              <p class="text-[11px] font-medium text-emerald-700/80 px-1">
                {{ frequencyOptions.find(f => f.value === currentStore.visit_frequency)?.desc || 'Tentukan ritme frekuensi kunjungan salesman ke toko ini.' }}
              </p>
            </div>

            <!-- Hari Kunjungan Rutin -->
            <div class="space-y-2 pt-2 border-t border-emerald-100">
              <label class="text-[10px] font-black text-emerald-800 uppercase tracking-widest ml-1 flex items-center justify-between">
                <span>Hari Kunjungan Rutin</span>
                <span class="text-[9px] font-bold text-emerald-600 lowercase">
                  {{ (currentStore.visit_days || []).length }} hari terpilih
                </span>
              </label>
              <div class="grid grid-cols-4 sm:grid-cols-7 gap-2">
                <button
                  v-for="d in dayOptions"
                  :key="d.value"
                  type="button"
                  @click="toggleVisitDay(d.value)"
                  class="py-2.5 px-2 rounded-xl text-xs font-black transition-all flex items-center justify-center cursor-pointer border"
                  :class="
                    (currentStore.visit_days || []).includes(d.value)
                      ? 'bg-emerald-600 text-white border-emerald-600 shadow-sm shadow-emerald-600/20'
                      : 'bg-white text-slate-600 border-slate-200 hover:bg-emerald-50/50'
                  "
                >
                  {{ d.short }}
                </button>
              </div>
            </div>
          </div>

          <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div class="space-y-2">
              <label class="text-[10px] font-black text-slate-400 uppercase tracking-widest ml-1"
                >Nama Toko / Outlet</label
              >
              <input
                v-model="currentStore.name"
                type="text"
                placeholder="Masukkan nama toko / pelanggan..."
                class="w-full bg-slate-50 border-2 border-transparent rounded-2xl px-5 py-4 font-bold text-slate-900 focus:bg-white focus:border-primary/20 focus:ring-0 transition-all"
              />
            </div>
            <div class="space-y-2">
              <label class="text-[10px] font-black text-slate-400 uppercase tracking-widest ml-1"
                >Nama Pemilik / PIC</label
              >
              <input
                v-model="currentStore.owner_name"
                type="text"
                placeholder="Nama pemilik toko..."
                class="w-full bg-slate-50 border-2 border-transparent rounded-2xl px-5 py-4 font-bold text-slate-900 focus:bg-white focus:border-primary/20 focus:ring-0 transition-all"
              />
            </div>
          </div>

          <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div class="space-y-2">
              <label class="text-[10px] font-black text-slate-400 uppercase tracking-widest ml-1"
                >Nomor Telepon</label
              >
              <input
                v-model="currentStore.phone_number"
                type="text"
                placeholder="0812..."
                class="w-full bg-slate-50 border-2 border-transparent rounded-2xl px-5 py-4 font-bold text-slate-900 focus:bg-white focus:border-primary/20 focus:ring-0 transition-all"
              />
            </div>
            <div class="space-y-2 relative" ref="salesmanDropdownRef">
              <label class="text-[10px] font-black text-slate-400 uppercase tracking-widest ml-1"
                >Salesman Penanggung Jawab</label
              >
              <div class="relative">
                <input
                  v-model="salesmanSearchQuery"
                  @focus="isSalesmanDropdownOpen = true"
                  @input="currentStore.assigned_employee_id = ''"
                  type="text"
                  placeholder="Ketik untuk mencari salesman..."
                  class="w-full bg-slate-50 border-2 border-transparent rounded-2xl px-5 py-4 font-bold text-slate-900 focus:bg-white focus:border-primary/20 focus:ring-0 transition-all pr-12"
                />
                <ChevronDown class="absolute right-4 top-1/2 -translate-y-1/2 w-5 h-5 text-slate-400 pointer-events-none" />
              </div>
              
              <div 
                v-if="isSalesmanDropdownOpen"
                class="absolute z-20 w-full mt-2 bg-white border border-slate-100 shadow-xl rounded-2xl max-h-60 overflow-y-auto custom-scrollbar"
              >
                <ul class="py-2">
                  <li 
                    v-if="filteredSalesmenForDropdown.length === 0"
                    class="px-5 py-3 text-sm text-slate-400 font-bold"
                  >
                    Tidak ada salesman ditemukan.
                  </li>
                  <li
                    v-for="s in filteredSalesmenForDropdown"
                    :key="s.id"
                    @click="selectSalesman(s)"
                    class="px-5 py-3 text-sm font-bold text-slate-700 hover:bg-slate-50 cursor-pointer transition-colors"
                    :class="{'bg-primary/5 text-primary': currentStore.assigned_employee_id === s.id}"
                  >
                    {{ s.first_name }} {{ s.last_name }} ({{ s.job_position?.title || 'Sales' }})
                  </li>
                </ul>
              </div>
            </div>
          </div>

          <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div class="space-y-2">
              <label class="text-[10px] font-black text-slate-400 uppercase tracking-widest ml-1"
                >NIK</label
              >
              <input
                v-model="currentStore.nik"
                type="text"
                placeholder="Nomor Induk Kependudukan (opsional)"
                class="w-full bg-slate-50 border-2 border-transparent rounded-2xl px-5 py-4 font-bold text-slate-900 focus:bg-white focus:border-primary/20 focus:ring-0 transition-all"
              />
            </div>
            <div class="space-y-2">
              <label class="text-[10px] font-black text-slate-400 uppercase tracking-widest ml-1"
                >NPWP</label
              >
              <input
                v-model="currentStore.npwp"
                type="text"
                placeholder="Nomor Pokok Wajib Pajak (opsional)"
                class="w-full bg-slate-50 border-2 border-transparent rounded-2xl px-5 py-4 font-bold text-slate-900 focus:bg-white focus:border-primary/20 focus:ring-0 transition-all"
              />
            </div>
          </div>
          <div class="space-y-2">
            <label class="text-[10px] font-black text-slate-400 uppercase tracking-widest ml-1"
              >Nama Pajak (Tax Name)</label
            >
            <input
              v-model="currentStore.tax_name"
              type="text"
              placeholder="Nama untuk keperluan pajak (opsional)"
              class="w-full bg-slate-50 border-2 border-transparent rounded-2xl px-5 py-4 font-bold text-slate-900 focus:bg-white focus:border-primary/20 focus:ring-0 transition-all"
            />
          </div>

          <div class="space-y-2">
            <label class="text-[10px] font-black text-slate-400 uppercase tracking-widest ml-1"
              >Foto Toko (Tampak Depan)</label
            >
            <div class="flex items-center gap-4">
              <div
                v-if="imagePreview || currentStore.photo_url"
                class="w-24 h-24 rounded-2xl overflow-hidden border-2 border-slate-100 shrink-0 bg-slate-50"
              >
                <img
                  :src="imagePreview || currentStore.photo_url"
                  class="w-full h-full object-cover"
                />
              </div>
              <div class="flex flex-col gap-2 flex-1">
                <button
                  type="button"
                  @click="openWebcam('store')"
                  class="h-11 rounded-xl border-2 border-dashed border-slate-200 hover:border-primary/40 hover:bg-primary/5 transition-all cursor-pointer flex items-center justify-center gap-2"
                >
                  <Camera class="w-4 h-4 text-slate-500" />
                  <span class="text-xs font-bold text-slate-600">Buka Kamera</span>
                </button>
                <label
                  class="h-11 rounded-xl border-2 border-dashed border-slate-200 hover:border-primary/40 hover:bg-primary/5 transition-all cursor-pointer flex items-center justify-center gap-2"
                >
                  <ImagePlus class="w-4 h-4 text-slate-500" />
                  <span class="text-xs font-bold text-slate-600">Pilih dari Galeri</span>
                  <input type="file" class="hidden" accept="image/*" @change="handleFileChange" />
                </label>
              </div>
            </div>
          </div>

          <div class="space-y-4">
            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div class="space-y-2">
                <label class="text-[10px] font-black text-slate-400 uppercase tracking-widest ml-1">Provinsi</label>
                <select
                  v-model="selectedProvinceId"
                  @change="onProvinceChange"
                  :disabled="isLoadingRegions"
                  class="w-full bg-slate-50 border-2 border-transparent rounded-2xl px-5 py-4 font-bold text-slate-700 focus:bg-white focus:border-primary/20 focus:ring-0 transition-all outline-none"
                >
                  <option :value="null">Pilih Provinsi</option>
                  <option v-for="prov in provinces" :key="prov.id" :value="prov.id">{{ prov.name }}</option>
                </select>
              </div>
              <div class="space-y-2">
                <label class="text-[10px] font-black text-slate-400 uppercase tracking-widest ml-1">Kabupaten/Kota</label>
                <select
                  v-model="selectedRegencyId"
                  @change="onRegencyChange"
                  :disabled="!selectedProvinceId || isLoadingRegions"
                  class="w-full bg-slate-50 border-2 border-transparent rounded-2xl px-5 py-4 font-bold text-slate-700 focus:bg-white focus:border-primary/20 focus:ring-0 transition-all outline-none disabled:opacity-50"
                >
                  <option :value="null">Pilih Kabupaten/Kota</option>
                  <option v-for="reg in regencies" :key="reg.id" :value="reg.id">{{ reg.name }}</option>
                </select>
              </div>
              <div class="space-y-2">
                <label class="text-[10px] font-black text-slate-400 uppercase tracking-widest ml-1">Kecamatan</label>
                <select
                  v-model="selectedDistrictId"
                  @change="onDistrictChange"
                  :disabled="!selectedRegencyId || isLoadingRegions"
                  class="w-full bg-slate-50 border-2 border-transparent rounded-2xl px-5 py-4 font-bold text-slate-700 focus:bg-white focus:border-primary/20 focus:ring-0 transition-all outline-none disabled:opacity-50"
                >
                  <option :value="null">Pilih Kecamatan</option>
                  <option v-for="dist in districts" :key="dist.id" :value="dist.id">{{ dist.name }}</option>
                </select>
              </div>
              <div class="space-y-2">
                <label class="text-[10px] font-black text-slate-400 uppercase tracking-widest ml-1">Desa/Kelurahan</label>
                <select
                  v-model="selectedVillageId"
                  @change="onVillageChange"
                  :disabled="!selectedDistrictId || isLoadingRegions"
                  class="w-full bg-slate-50 border-2 border-transparent rounded-2xl px-5 py-4 font-bold text-slate-700 focus:bg-white focus:border-primary/20 focus:ring-0 transition-all outline-none disabled:opacity-50"
                >
                  <option :value="null">Pilih Desa/Kelurahan</option>
                  <option v-for="vil in villages" :key="vil.id" :value="vil.id">{{ vil.name }}</option>
                </select>
              </div>
            </div>

            <div class="space-y-2">
              <label class="text-[10px] font-black text-slate-400 uppercase tracking-widest ml-1"
                >Alamat Lengkap</label
              >
            <textarea
              v-model="currentStore.address"
              placeholder="Masukkan alamat lengkap toko..."
              class="w-full bg-slate-50 border-2 border-transparent rounded-2xl px-5 py-4 font-bold text-slate-900 focus:bg-white focus:border-primary/20 focus:ring-0 transition-all h-24 resize-none"
            ></textarea>
          </div>
          </div>

          <div class="space-y-4">
            <div class="flex items-center justify-between">
              <label class="text-[10px] font-black text-slate-400 uppercase tracking-widest ml-1"
                >Titik Lokasi (Map Picker)</label
              >
              <div class="flex items-center gap-2">
                <div
                  v-if="isGeocoding"
                  class="flex items-center gap-2 text-[10px] font-black text-primary animate-pulse"
                >
                  <Globe class="w-3 h-3 animate-spin" /> MENCARI ALAMAT...
                </div>
                
                <button 
                  type="button"
                  @click="getCurrentLocation"
                  :disabled="isGettingLocation"
                  class="flex items-center gap-1.5 px-3 py-1.5 text-[10px] font-black uppercase tracking-wider rounded-lg border transition-all"
                  :class="isGettingLocation ? 'bg-slate-100 text-slate-400 border-slate-200' : 'bg-blue-50 text-blue-600 border-blue-200 hover:bg-blue-100'"
                >
                  <Loader2 v-if="isGettingLocation" class="w-3 h-3 animate-spin" />
                  <LocateFixed v-else class="w-3 h-3" />
                  {{ isGettingLocation ? 'Mengambil...' : 'Lokasi Saat Ini' }}
                </button>
              </div>
            </div>
            <div
              id="store-map"
              class="w-full h-64 rounded-3xl border border-slate-100 shadow-inner z-10 overflow-hidden"
            ></div>

            <div class="grid grid-cols-1 md:grid-cols-2 gap-6 mt-4">
              <div class="space-y-2">
                <label class="text-[10px] font-black text-slate-400 uppercase tracking-widest ml-1"
                  >Latitude</label
                >
                <input
                  v-model.number="currentStore.latitude"
                  type="number"
                  step="any"
                  class="w-full bg-slate-50 border-2 border-transparent rounded-2xl px-5 py-4 font-bold text-slate-900 focus:bg-white focus:border-primary/20 focus:ring-0 transition-all"
                />
              </div>
              <div class="space-y-2">
                <label class="text-[10px] font-black text-slate-400 uppercase tracking-widest ml-1"
                  >Longitude</label
                >
                <input
                  v-model.number="currentStore.longitude"
                  type="number"
                  step="any"
                  class="w-full bg-slate-50 border-2 border-transparent rounded-2xl px-5 py-4 font-bold text-slate-900 focus:bg-white focus:border-primary/20 focus:ring-0 transition-all"
                />
              </div>
            </div>
          </div>

          <div class="flex items-center gap-3 bg-slate-50 p-4 rounded-2xl border border-slate-100">
            <input
              type="checkbox"
              v-model="currentStore.is_active"
              id="is_active"
              class="w-5 h-5 rounded-lg border-2 border-slate-300 text-primary focus:ring-primary/20 transition-all"
            />
            <label for="is_active" class="text-sm font-bold text-slate-700 cursor-pointer"
              >Status Toko Aktif</label
            >
          </div>
        </div>

        <!-- Footer -->
        <div class="p-4 sm:p-8 border-t border-slate-100 flex flex-col-reverse sm:flex-row gap-3 sm:gap-4 bg-white shrink-0">
          <button
            @click="closeModal"
            class="w-full sm:flex-1 py-3 sm:py-4 text-xs font-black text-slate-400 hover:text-slate-900 transition-all uppercase tracking-widest cursor-pointer"
          >
            BATAL
          </button>
          <button
            @click="handleSubmit"
            class="w-full sm:flex-[2] bg-primary text-white rounded-2xl py-3.5 sm:py-4 text-xs font-black shadow-lg shadow-primary/20 hover:scale-[1.02] active:scale-[0.98] transition-all flex items-center justify-center gap-2 uppercase tracking-widest cursor-pointer"
          >
            <Save class="w-4 h-4 sm:w-5 sm:h-5" /> SIMPAN DATA TOKO
          </button>
        </div>
      </div>
    </div>

    <!-- Image Preview Modal -->
    <div v-if="showImagePreview" class="fixed inset-0 z-[60] flex items-center justify-center p-4">
      <div
        class="absolute inset-0 bg-slate-900/80 backdrop-blur-sm"
        @click="showImagePreview = false"
      ></div>
      <div
        class="relative z-10 w-full max-w-3xl flex flex-col items-center animate-in fade-in zoom-in duration-300"
      >
        <button
          @click="showImagePreview = false"
          class="absolute -top-12 right-0 p-2 bg-white/10 hover:bg-white/20 text-white rounded-full backdrop-blur-md transition-all"
        >
          <X class="w-6 h-6" />
        </button>
        <img
          :src="previewImageUrl"
          class="w-full max-h-[85vh] object-contain rounded-2xl shadow-2xl"
        />
      </div>
    </div>

    <!-- History Order Modal -->
    <div v-if="showHistoryModal" class="fixed inset-0 z-50 flex items-center justify-center p-4">
      <div
        class="absolute inset-0 bg-slate-900/60 backdrop-blur-sm"
        @click="closeHistoryModal"
      ></div>
      <div
        class="bg-white rounded-[32px] w-full max-w-4xl relative z-10 shadow-2xl flex flex-col max-h-[90vh] overflow-hidden animate-in fade-in zoom-in duration-300"
      >
        <div
          class="px-8 py-6 border-b border-slate-100 flex items-center justify-between bg-white shrink-0"
        >
          <div>
            <h3 class="text-2xl font-black text-slate-900 tracking-tight">
              Riwayat Pesanan
            </h3>
            <p class="text-[10px] font-black text-slate-400 uppercase tracking-widest mt-0.5">
              {{ currentHistoryStore?.name }}
            </p>
          </div>
          <button
            @click="closeHistoryModal"
            class="p-2 hover:bg-slate-100 rounded-xl transition-all text-slate-400 hover:text-slate-900"
          >
            <X class="w-6 h-6" />
          </button>
        </div>

        <div class="flex-1 overflow-y-auto p-8 custom-scrollbar">
          <div v-if="isHistoryLoading" class="flex flex-col items-center justify-center py-20 gap-4">
            <div class="w-10 h-10 border-4 border-primary border-t-transparent rounded-full animate-spin"></div>
            <p class="text-sm font-bold text-slate-400">Memuat riwayat pesanan...</p>
          </div>
          <div v-else-if="historyOrders.length === 0" class="flex flex-col items-center justify-center py-20 gap-4">
            <History class="w-12 h-12 text-slate-200" />
            <p class="text-sm font-bold text-slate-400">Belum ada riwayat pesanan.</p>
          </div>
          <div v-else class="space-y-4">
            <div
              v-for="order in historyOrders"
              :key="order.id"
              class="bg-white rounded-2xl border border-slate-200 shadow-sm overflow-hidden"
            >
              <!-- Order Header -->
              <div 
                class="p-5 flex flex-wrap items-center justify-between gap-4 cursor-pointer hover:bg-slate-50 transition-colors"
                @click="toggleExpandOrder(order.id)"
              >
                <div class="flex items-center gap-4">
                  <div class="w-10 h-10 rounded-xl bg-indigo-50 flex items-center justify-center shrink-0">
                    <Package class="w-5 h-5 text-indigo-500" />
                  </div>
                  <div>
                    <h4 class="text-sm font-black text-slate-900">{{ order.so_number }}</h4>
                    <div class="flex items-center gap-3 mt-1">
                      <div class="flex items-center gap-1">
                        <Calendar class="w-3 h-3 text-slate-400" />
                        <span class="text-[11px] font-semibold text-slate-500">{{ formatDate(order.created_at) }}</span>
                      </div>
                      <span class="text-slate-300">•</span>
                      <div class="flex items-center gap-1">
                        <User class="w-3 h-3 text-slate-400" />
                        <span class="text-[11px] font-semibold text-slate-500">{{ order.employee?.first_name || '-' }}</span>
                      </div>
                    </div>
                  </div>
                </div>

                <div class="flex items-center gap-6">
                  <div class="text-right">
                    <p class="text-[10px] font-black text-slate-400 uppercase tracking-widest">Total Belanja</p>
                    <p class="text-sm font-black text-slate-900">{{ formatCurrency(order.total_amount) }}</p>
                  </div>
                  <span
                    :class="['text-[10px] font-black px-3 py-1 rounded-full uppercase tracking-wider border', getStatusClass(order.status)]"
                  >
                    {{ order.status }}
                  </span>
                </div>
              </div>

              <!-- Order Items (Expanded) -->
              <div v-if="expandedOrderIds.includes(order.id)" class="border-t border-slate-100 bg-slate-50/50 p-5">
                <div class="space-y-3">
                  <h5 class="text-[10px] font-black text-slate-400 uppercase tracking-widest mb-3">Rincian Barang</h5>
                  
                  <div v-if="!order.items || order.items.length === 0" class="text-xs font-medium text-slate-500 italic">
                    Tidak ada item.
                  </div>

                  <div 
                    v-for="item in order.items" 
                    :key="item.id"
                    class="flex items-center justify-between bg-white p-3 rounded-xl border border-slate-100 shadow-sm"
                  >
                    <div class="flex items-center gap-3">
                      <div class="w-8 h-8 rounded-lg bg-slate-100 flex items-center justify-center">
                        <FileText class="w-4 h-4 text-slate-400" />
                      </div>
                      <div>
                        <p class="text-xs font-bold text-slate-900">{{ item.product?.name || 'Produk Tidak Diketahui' }}</p>
                        <p class="text-[10px] font-semibold text-slate-500 mt-0.5">
                          {{ item.quantity }} x {{ formatCurrency(item.unit_price) }}
                        </p>
                      </div>
                    </div>
                    <p class="text-xs font-black text-slate-900">{{ formatCurrency(item.total_price) }}</p>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>

        <div v-if="historyTotal > 0" class="px-4 sm:px-8 py-3 sm:py-4 border-t border-slate-100 bg-slate-50 flex flex-col sm:flex-row items-center justify-between gap-3 shrink-0">
          <span class="text-xs font-bold text-slate-500 text-center sm:text-left">
            Menampilkan {{ (historyCurrentPage - 1) * historyItemsPerPage + 1 }} -
            {{ Math.min(historyCurrentPage * historyItemsPerPage, historyTotal) }} dari {{ historyTotal }} data
          </span>
          <div class="flex items-center gap-2">
            <button
              @click="historyPrevPage"
              :disabled="historyCurrentPage === 1"
              class="px-3 py-1.5 rounded-lg text-xs font-black border border-slate-200 bg-white disabled:opacity-50 hover:bg-slate-50 transition-colors cursor-pointer"
            >
              Prev
            </button>
            <span class="text-xs font-black text-slate-700 px-2">{{ historyCurrentPage }} / {{ historyTotalPages }}</span>
            <button
              @click="historyNextPage"
              :disabled="historyCurrentPage === historyTotalPages"
              class="px-3 py-1.5 rounded-lg text-xs font-black border border-slate-200 bg-white disabled:opacity-50 hover:bg-slate-50 transition-colors cursor-pointer"
            >
              Next
            </button>
          </div>
        </div>
      </div>
    </div>
    <!-- Modal Verifikasi Toko -->
    <div v-if="showAuditModal" class="fixed inset-0 z-50 flex items-center justify-center p-4">
      <div class="absolute inset-0 bg-slate-900/60 backdrop-blur-sm" @click="showAuditModal = false"></div>
      
      <div class="bg-white rounded-[32px] w-full max-w-lg relative z-10 shadow-2xl flex flex-col max-h-[90vh] overflow-hidden animate-in fade-in zoom-in duration-300">
        
        <!-- Header -->
        <div class="px-6 sm:px-8 py-5 sm:py-6 border-b border-slate-100 flex items-center justify-between bg-white shrink-0">
          <div class="flex items-center gap-3">
            <div class="w-10 h-10 rounded-full bg-blue-50 flex items-center justify-center ring-4 ring-blue-50/50">
              <ShieldCheck class="w-5 h-5 text-blue-600" />
            </div>
            <div>
              <h3 class="text-xl font-black text-slate-900 tracking-tight">Verifikasi Toko</h3>
              <p class="text-[10px] font-black text-slate-400 uppercase tracking-widest mt-0.5">Validasi Keaslian Outlet</p>
            </div>
          </div>
          <button @click="showAuditModal = false" class="p-2 hover:bg-slate-100 rounded-xl transition-all text-slate-400 hover:text-slate-900 cursor-pointer">
            <X class="w-6 h-6" />
          </button>
        </div>

        <!-- Body -->
        <div class="p-6 sm:p-8 overflow-y-auto space-y-5 sm:space-y-6 custom-scrollbar">
          <div class="p-4 bg-slate-50 rounded-2xl border border-slate-100">
            <p class="text-xs font-black text-slate-400 uppercase tracking-widest">Toko yang diverifikasi</p>
            <p class="text-sm font-black text-slate-900 mt-1">{{ selectedStoreForAudit?.name }}</p>
            <p class="text-xs font-medium text-slate-500 mt-1 flex items-center gap-1.5 line-clamp-1"><MapPin class="w-3.5 h-3.5"/> {{ selectedStoreForAudit?.address || 'Tidak ada alamat' }}</p>
          </div>
          
          <div class="space-y-3">
            <label class="block text-[10px] font-black text-slate-400 uppercase tracking-widest ml-1">Hasil Verifikasi <span class="text-rose-500">*</span></label>
            <div class="grid grid-cols-2 gap-3 sm:gap-4">
              <label :class="['relative flex flex-col items-center justify-center p-3 sm:p-4 cursor-pointer rounded-2xl border-2 transition-all', auditForm.status === 'VERIFIED_REAL' ? 'border-emerald-500 bg-emerald-50/50 shadow-sm shadow-emerald-500/20' : 'border-slate-100 hover:border-slate-200 bg-white']">
                <input type="radio" v-model="auditForm.status" value="VERIFIED_REAL" class="sr-only">
                <CheckCircle2 :class="['w-7 h-7 sm:w-8 sm:h-8 mb-2', auditForm.status === 'VERIFIED_REAL' ? 'text-emerald-500' : 'text-slate-300']"/>
                <span :class="['text-xs sm:text-sm font-black uppercase tracking-wide', auditForm.status === 'VERIFIED_REAL' ? 'text-emerald-700' : 'text-slate-500']">REAL (Valid)</span>
              </label>
              
              <label :class="['relative flex flex-col items-center justify-center p-3 sm:p-4 cursor-pointer rounded-2xl border-2 transition-all', auditForm.status === 'FICTITIOUS' ? 'border-rose-500 bg-rose-50/50 shadow-sm shadow-rose-500/20' : 'border-slate-100 hover:border-slate-200 bg-white']">
                <input type="radio" v-model="auditForm.status" value="FICTITIOUS" class="sr-only">
                <AlertCircle :class="['w-7 h-7 sm:w-8 sm:h-8 mb-2', auditForm.status === 'FICTITIOUS' ? 'text-rose-500' : 'text-slate-300']"/>
                <span :class="['text-xs sm:text-sm font-black uppercase tracking-wide', auditForm.status === 'FICTITIOUS' ? 'text-rose-700' : 'text-slate-500']">FIKTIF (Palsu)</span>
              </label>
            </div>
          </div>

          <div class="space-y-2">
            <label class="block text-[10px] font-black text-slate-400 uppercase tracking-widest ml-1">Catatan Verifikasi <span class="text-rose-500">*</span></label>
            <textarea v-model="auditForm.notes" rows="4" 
                      class="w-full bg-slate-50 border-2 border-transparent rounded-2xl px-4 sm:px-5 py-3 sm:py-4 font-bold text-slate-900 focus:bg-white focus:border-primary/20 focus:ring-0 transition-all resize-none text-xs sm:text-sm" 
                      placeholder="Tuliskan alasan atau temuan di lapangan..."></textarea>
          </div>

          <div class="space-y-2">
            <label class="block text-[10px] font-black text-slate-400 uppercase tracking-widest ml-1">Foto Bukti Kunjungan (Opsional)</label>
            <div class="flex items-center gap-3 sm:gap-4">
              <div v-if="auditPhotoPreview" class="w-20 h-20 sm:w-24 sm:h-24 rounded-2xl overflow-hidden border-2 border-slate-100 shrink-0 bg-slate-50">
                <img :src="auditPhotoPreview" class="w-full h-full object-cover" />
              </div>
              <div class="flex flex-col gap-2 flex-1">
                <button type="button" @click="openWebcam('audit')" class="h-10 sm:h-11 rounded-xl border-2 border-dashed border-slate-200 hover:border-primary/40 hover:bg-primary/5 transition-all cursor-pointer flex items-center justify-center gap-2">
                  <Camera class="w-4 h-4 text-slate-500" />
                  <span class="text-xs font-bold text-slate-600">Buka Kamera</span>
                </button>
                <label class="h-10 sm:h-11 rounded-xl border-2 border-dashed border-slate-200 hover:border-primary/40 hover:bg-primary/5 transition-all cursor-pointer flex items-center justify-center gap-2">
                  <ImagePlus class="w-4 h-4 text-slate-500" />
                  <span class="text-xs font-bold text-slate-600">Pilih dari Galeri</span>
                  <input type="file" class="hidden" accept="image/*" @change="handleAuditPhotoChange" />
                </label>
              </div>
            </div>
          </div>
        </div>

        <!-- Footer -->
        <div class="px-6 sm:px-8 py-4 sm:py-6 border-t border-slate-100 bg-white shrink-0 flex flex-col-reverse sm:flex-row gap-2 sm:gap-4 justify-end">
          <Button variant="outline" @click="showAuditModal = false" class="w-full sm:w-auto rounded-xl px-6 py-4 sm:py-5 font-bold border-slate-200 text-slate-600 hover:bg-slate-50 hover:text-slate-900 cursor-pointer">
            Batal
          </Button>
          <Button @click="submitAudit" :disabled="isSubmittingAudit" class="w-full sm:w-auto bg-blue-600 hover:bg-blue-700 text-white rounded-xl px-6 sm:px-8 py-4 sm:py-5 font-bold shadow-lg shadow-blue-600/20 cursor-pointer">
            <Loader2 v-if="isSubmittingAudit" class="w-5 h-5 mr-2 animate-spin" />
            <ShieldCheck v-else class="w-5 h-5 mr-2" />
            Simpan Verifikasi
          </Button>
        </div>
      </div>
    </div>
    
    <!-- Import Modal -->
    <StoreImportModal
      :show="showImportModal"
      @close="showImportModal = false"
      @success="handleImportSuccess"
    />

    <!-- Audit History Modal -->
    <div v-if="showAuditHistoryModal" class="fixed inset-0 z-50 flex items-center justify-center p-4 sm:p-6 bg-slate-900/40 backdrop-blur-sm">
      <div class="bg-white rounded-[2rem] w-full max-w-2xl shadow-2xl flex flex-col max-h-[90vh] overflow-hidden" v-on-click-outside="closeAuditHistoryModal">
        <!-- Header -->
        <div class="px-8 py-6 border-b border-slate-100 flex items-center justify-between bg-white shrink-0">
          <div>
            <h2 class="text-xl font-black text-slate-800">Riwayat Verifikasi Toko</h2>
            <p class="text-sm font-medium text-slate-500 mt-1">{{ currentAuditHistoryStore?.name }}</p>
          </div>
          <button @click="closeAuditHistoryModal" class="p-2 hover:bg-slate-100 rounded-xl transition-colors">
            <X class="w-5 h-5 text-slate-400" />
          </button>
        </div>

        <!-- Body -->
        <div class="p-8 overflow-y-auto custom-scrollbar flex-1 bg-slate-50">
          <div v-if="isAuditHistoryLoading" class="flex flex-col items-center justify-center py-12">
            <Loader2 class="w-8 h-8 text-primary animate-spin mb-4" />
            <p class="text-sm font-bold text-slate-400">Memuat riwayat verifikasi...</p>
          </div>

          <div v-else-if="auditHistories.length === 0" class="flex flex-col items-center justify-center py-12 text-center">
            <div class="w-16 h-16 rounded-full bg-slate-100 flex items-center justify-center mb-4">
              <Activity class="w-8 h-8 text-slate-300" />
            </div>
            <h3 class="text-sm font-bold text-slate-600 mb-1">Belum Ada Riwayat</h3>
            <p class="text-xs font-medium text-slate-400">Toko ini belum pernah diverifikasi oleh auditor.</p>
          </div>

          <div v-else class="space-y-6">
            <div v-for="(audit, index) in auditHistories" :key="audit.id" class="relative pl-6">
              <!-- Timeline line -->
              <div v-if="index !== auditHistories.length - 1" class="absolute left-2.5 top-6 bottom-[-1.5rem] w-0.5 bg-slate-200"></div>
              
              <!-- Timeline dot -->
              <div class="absolute left-1 top-1.5 w-3.5 h-3.5 rounded-full border-2 border-white shadow-sm"
                   :class="audit.status === 'VERIFIED_REAL' ? 'bg-emerald-500' : 'bg-rose-500'"></div>

              <div class="bg-white rounded-2xl p-5 shadow-sm border border-slate-100">
                <div class="flex items-center justify-between mb-3">
                  <div class="flex items-center gap-2">
                    <div class="w-8 h-8 rounded-full bg-blue-100 flex items-center justify-center text-blue-600 font-bold text-xs">
                      {{ audit.auditor?.first_name?.charAt(0) || 'A' }}
                    </div>
                    <div>
                      <p class="text-sm font-bold text-slate-800">{{ audit.auditor?.first_name }} {{ audit.auditor?.last_name }}</p>
                      <p class="text-xs font-medium text-slate-400 capitalize">{{ audit.auditor?.job_position?.title || 'Auditor' }} &bull; {{ new Date(audit.created_at).toLocaleString('id-ID') }}</p>
                    </div>
                  </div>
                  <span :class="['px-3 py-1 text-[10px] font-black uppercase tracking-widest rounded-lg', audit.status === 'VERIFIED_REAL' ? 'bg-emerald-100 text-emerald-700' : 'bg-rose-100 text-rose-700']">
                    {{ audit.status === 'VERIFIED_REAL' ? 'TERVERIFIKASI' : 'FIKTIF' }}
                  </span>
                </div>
                
                <p class="text-sm font-medium text-slate-600 bg-slate-50 p-4 rounded-xl mb-4">{{ audit.notes || 'Tidak ada catatan.' }}</p>
                
                <div v-if="audit.photo_proof_url" class="mt-2">
                  <p class="text-[10px] font-black text-slate-400 uppercase tracking-widest mb-2">Foto Bukti</p>
                  <a :href="audit.photo_proof_url" target="_blank" class="block w-24 h-24 rounded-xl overflow-hidden border border-slate-200 hover:border-blue-400 transition-colors">
                    <img :src="audit.photo_proof_url" class="w-full h-full object-cover" />
                  </a>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Global Audit History Modal -->
    <div v-if="showGlobalAuditHistoryModal" class="fixed inset-0 z-50 flex items-center justify-center p-4 sm:p-6 bg-slate-900/40 backdrop-blur-sm">
      <div class="bg-white rounded-[2rem] w-full max-w-4xl shadow-2xl flex flex-col max-h-[90vh] overflow-hidden" v-on-click-outside="closeGlobalAuditHistoryModal">
        <!-- Header -->
        <div class="px-8 py-6 border-b border-slate-100 flex items-center justify-between bg-white shrink-0">
          <div>
            <h2 class="text-xl font-black text-slate-800">Rekap Riwayat Verifikasi</h2>
            <p class="text-sm font-medium text-slate-500 mt-1">Daftar semua toko yang telah divalidasi</p>
          </div>
          <button @click="closeGlobalAuditHistoryModal" class="p-2 hover:bg-slate-100 rounded-xl transition-colors">
            <X class="w-5 h-5 text-slate-400" />
          </button>
        </div>

        <!-- Body -->
        <div class="p-0 overflow-y-auto custom-scrollbar flex-1 bg-slate-50 relative">
          <div v-if="isGlobalAuditHistoryLoading && globalAuditHistories.length === 0" class="flex flex-col items-center justify-center py-12">
            <Loader2 class="w-8 h-8 text-primary animate-spin mb-4" />
            <p class="text-sm font-bold text-slate-400">Memuat rekap verifikasi...</p>
          </div>

          <div v-else-if="globalAuditHistories.length === 0" class="flex flex-col items-center justify-center py-12 text-center">
            <div class="w-16 h-16 rounded-full bg-slate-100 flex items-center justify-center mb-4">
              <ClipboardList class="w-8 h-8 text-slate-300" />
            </div>
            <h3 class="text-sm font-bold text-slate-600 mb-1">Belum Ada Rekap</h3>
            <p class="text-xs font-medium text-slate-400">Belum ada satupun toko yang diverifikasi.</p>
          </div>

          <div v-else class="p-6">
            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div v-for="audit in globalAuditHistories" :key="audit.id" class="bg-white rounded-2xl p-5 shadow-sm border border-slate-200 hover:shadow-md transition-all flex flex-col">
                <div class="flex justify-between items-start mb-3 border-b border-slate-100 pb-3">
                  <div>
                    <h3 class="text-sm font-black text-slate-800 line-clamp-1">{{ audit.store?.name }}</h3>
                    <p class="text-[11px] font-medium text-slate-500 flex items-center gap-1 mt-1">
                      <MapPin class="w-3 h-3 text-slate-400 shrink-0" />
                      <span class="line-clamp-1">{{ audit.store?.address || '-' }}</span>
                    </p>
                  </div>
                  <span :class="['shrink-0 px-2.5 py-1 text-[9px] font-black uppercase tracking-widest rounded-lg', audit.status === 'VERIFIED_REAL' ? 'bg-emerald-100 text-emerald-700' : 'bg-rose-100 text-rose-700']">
                    {{ audit.status === 'VERIFIED_REAL' ? 'TERVERIFIKASI' : 'FIKTIF' }}
                  </span>
                </div>
                
                <div class="flex items-center gap-3 mb-4">
                  <div class="w-8 h-8 rounded-full bg-blue-100 flex items-center justify-center text-blue-600 font-bold text-xs shrink-0">
                    {{ audit.auditor?.first_name?.charAt(0) || 'A' }}
                  </div>
                  <div>
                    <p class="text-xs font-bold text-slate-800">{{ audit.auditor?.first_name }} {{ audit.auditor?.last_name }}</p>
                    <p class="text-[10px] font-medium text-slate-400 capitalize">{{ audit.auditor?.job_position?.title || 'Auditor' }} &bull; {{ new Date(audit.created_at).toLocaleString('id-ID') }}</p>
                  </div>
                </div>

                <div class="flex gap-4 mt-auto">
                  <div class="flex-1 bg-slate-50 rounded-xl p-3">
                    <p class="text-[10px] font-bold text-slate-400 uppercase tracking-wider mb-1">Catatan</p>
                    <p class="text-xs font-medium text-slate-600 line-clamp-3">{{ audit.notes || '-' }}</p>
                  </div>
                  <div v-if="audit.photo_proof_url" class="shrink-0 w-16 h-16 rounded-xl overflow-hidden border border-slate-200">
                    <a :href="audit.photo_proof_url" target="_blank" class="block w-full h-full">
                      <img :src="audit.photo_proof_url" class="w-full h-full object-cover hover:scale-110 transition-transform" />
                    </a>
                  </div>
                </div>
              </div>
            </div>

            <div class="mt-8 flex justify-center pb-4">
              <Button v-if="hasMoreGlobalAudits" @click="loadMoreGlobalAudits" :disabled="isGlobalAuditHistoryLoading" variant="outline" class="rounded-full px-6 text-sm font-bold border-slate-200 text-slate-600">
                <Loader2 v-if="isGlobalAuditHistoryLoading" class="w-4 h-4 mr-2 animate-spin" />
                Muat Lebih Banyak
              </Button>
              <p v-else-if="globalAuditHistories.length > 0" class="text-xs font-bold text-slate-400">Semua riwayat telah dimuat</p>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Webcam Capture Modal -->
    <div v-if="showWebcamModal" class="fixed inset-0 z-[60] flex items-center justify-center bg-black/90 backdrop-blur-md">
      <div class="flex flex-col items-center w-full max-w-lg p-4">
        <div class="w-full relative bg-black rounded-2xl overflow-hidden aspect-[3/4] sm:aspect-video border border-slate-700 shadow-2xl">
          <video ref="videoRef" autoplay playsinline class="w-full h-full object-cover"></video>
        </div>
        
        <div class="mt-8 flex items-center gap-6">
          <button @click="closeWebcam" class="w-12 h-12 rounded-full bg-slate-800 text-white flex items-center justify-center hover:bg-slate-700 transition-colors">
            <X class="w-5 h-5" />
          </button>
          <button @click="captureWebcam" class="w-20 h-20 rounded-full bg-white border-4 border-slate-300 flex items-center justify-center shadow-lg active:scale-95 transition-all">
            <div class="w-16 h-16 rounded-full border-2 border-slate-900"></div>
          </button>
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

/* Address Tooltip */
.address-tooltip-wrapper {
  position: relative;
  display: inline-block;
  max-width: 200px;
}

.address-tooltip {
  position: absolute;
  bottom: calc(100% + 10px);
  left: 50%;
  transform: translateX(-50%) translateY(6px);
  background: #0f172a;
  color: white;
  padding: 10px 14px;
  border-radius: 12px;
  width: 260px;
  max-width: 260px;
  box-shadow:
    0 8px 32px rgba(0, 0, 0, 0.25),
    0 2px 8px rgba(0, 0, 0, 0.15);
  z-index: 9999;
  opacity: 0;
  pointer-events: none;
  transition:
    opacity 0.2s ease,
    transform 0.2s ease;
  white-space: normal;
  word-break: break-word;
}

.address-tooltip-arrow {
  position: absolute;
  bottom: -6px;
  left: 50%;
  transform: translateX(-50%);
  width: 12px;
  height: 6px;
  background: #0f172a;
  clip-path: polygon(0 0, 100% 0, 50% 100%);
}

.address-tooltip-wrapper:hover .address-tooltip {
  opacity: 1;
  transform: translateX(-50%) translateY(0);
}

/* Missing Fields Tooltip */
.missing-fields-tooltip-wrapper {
  position: relative;
  display: inline-block;
}

.missing-fields-tooltip {
  position: absolute;
  bottom: calc(100% + 8px);
  left: 0;
  transform: translateY(6px);
  background: #0f172a;
  color: white;
  padding: 10px 12px;
  border-radius: 12px;
  width: 220px;
  max-width: 220px;
  box-shadow:
    0 10px 30px rgba(0, 0, 0, 0.3),
    0 2px 8px rgba(0, 0, 0, 0.2);
  z-index: 9999;
  opacity: 0;
  pointer-events: none;
  transition:
    opacity 0.2s ease,
    transform 0.2s ease;
  white-space: normal;
  word-break: break-word;
}

.missing-fields-tooltip-arrow {
  position: absolute;
  bottom: -6px;
  left: 24px;
  width: 12px;
  height: 6px;
  background: #0f172a;
  clip-path: polygon(0 0, 100% 0, 50% 100%);
}

.missing-fields-tooltip-wrapper:hover .missing-fields-tooltip {
  opacity: 1;
  transform: translateY(0);
}
</style>
