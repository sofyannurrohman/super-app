<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed, watch } from 'vue'
import { useRouter } from 'vue-router'
import { toast } from 'vue-sonner'
import {
  Clock,
  MapPin,
  Calendar,
  FileText,
  Send,
  CheckCircle2,
  History,
  AlertCircle,
  LogOut,
  Camera,
  Trash2,
  Paperclip,
  Eye,
  RefreshCw,
  AlertTriangle,
  User,
  ShieldCheck,
  Plane,
  Briefcase,
  MessageSquare,
  Check,
  X,
  FileEdit,
} from 'lucide-vue-next'
import axios from '@/api/axios'
import { useAuthStore } from '@/stores/auth'
import { detectAndCropFace } from '@/utils/faceCrop'
import {
  burnAttendanceWatermark,
  formatAttendanceDateTime,
  reverseGeocodeAddress,
} from '@/utils/watermark'
import L from 'leaflet'
import 'leaflet/dist/leaflet.css'
import 'leaflet-defaulticon-compatibility/dist/leaflet-defaulticon-compatibility.css'
import 'leaflet-defaulticon-compatibility'

const router = useRouter()
const authStore = useAuthStore()
const employeeProfile = ref<any>(null)
const leaveBalances = ref<any[]>([])
const leaveHistory = ref<any[]>([])
const attendanceHistory = ref<any[]>([])

const isLoading = ref(true)
const isSubmitting = ref(false)

// Tab state for History
const historyTab = ref<'attendance' | 'leave'>('attendance')
const currentTab = ref<'attendance' | 'leaves' | 'profile'>('attendance')

// Face register states
const isSubmittingFace = ref(false)
const capturedFacePhoto = ref<string | null>(null)
const capturedFaceFile = ref<File | null>(null)
const facePhotoInput = ref<HTMLInputElement | null>(null)

// Camera & Selfie states
const capturedSelfie = ref<string | null>(null)
const capturedSelfieFile = ref<File | null>(null)
const isProcessingSelfie = ref(false)
const selfieInput = ref<HTMLInputElement | null>(null)

// GPS states
const latitude = ref<number | null>(null)
const longitude = ref<number | null>(null)
const gpsError = ref<string | null>(null)
const isFetchingGps = ref(false)

// Out of town (Dinas Luar Kota) states
const isOutOfTown = ref(false)
const outOfTownDestination = ref('')

// Late Reason Modal states
const isLateReasonModalOpen = ref(false)
const lateAttendanceId = ref('')
const lateReasonInput = ref('')
const isSubmittingLateReason = ref(false)

// Map states
const portalMapContainer = ref<HTMLElement | null>(null)
let portalMap: L.Map | null = null
let userMarker: L.Marker | null = null
let officeMarker: L.Marker | null = null
let geofenceCircle: L.Circle | null = null

// Leave form states
const leaveForm = ref({
  leave_type_id: '',
  start_date: '',
  end_date: '',
  reason: '',
  attachment_file: null as File | null,
})

const liveTime = ref('')
const liveDate = ref('')

// Update live clock
const updateClock = () => {
  const now = new Date()
  liveTime.value = now.toLocaleTimeString('id-ID', {
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
  })
  liveDate.value = now.toLocaleDateString('id-ID', {
    weekday: 'long',
    day: 'numeric',
    month: 'long',
    year: 'numeric',
  })
}

// Fetch GPS location
const fetchGps = () => {
  isFetchingGps.value = true
  gpsError.value = null
  if (!navigator.geolocation) {
    gpsError.value = 'Geolokasi tidak didukung oleh browser Anda'
    isFetchingGps.value = false
    return
  }

  navigator.geolocation.getCurrentPosition(
    (position) => {
      latitude.value = position.coords.latitude
      longitude.value = position.coords.longitude
      isFetchingGps.value = false
    },
    (err) => {
      console.error(err)
      gpsError.value = 'Gagal mengambil lokasi. Pastikan GPS aktif dan izin lokasi diizinkan.'
      isFetchingGps.value = false
    },
    { enableHighAccuracy: true, timeout: 15000 },
  )
}

// Map initialization
const initPortalMap = () => {
  if (
    !portalMapContainer.value ||
    !latitude.value ||
    !longitude.value ||
    !employeeProfile.value?.branch
  )
    return

  const userLat = latitude.value
  const userLng = longitude.value
  const branch = employeeProfile.value.branch
  const officeLat = branch.latitude
  const officeLng = branch.longitude
  const radius = branch.radius_meter

  try {
    if (portalMap) {
      portalMap.remove()
      portalMap = null
    }

    portalMap = L.map(portalMapContainer.value, {
      zoomControl: false,
      attributionControl: false,
    }).setView([userLat, userLng], 16)

    L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
      maxZoom: 19,
    }).addTo(portalMap)

    L.control.zoom({ position: 'bottomright' }).addTo(portalMap)

    // Office marker
    officeMarker = L.marker([officeLat, officeLng]).addTo(portalMap)
    officeMarker.bindPopup(
      `<b>Kantor: ${branch.name || 'Cabang'}</b><br/>Radius Absensi: ${radius}m`,
    )

    // Geofence Circle (green if inside, red if outside)
    const inside = isWithinRadius.value
    geofenceCircle = L.circle([officeLat, officeLng], {
      radius: radius,
      color: inside ? '#10b981' : '#f43f5e',
      fillColor: inside ? '#10b981' : '#f43f5e',
      fillOpacity: 0.12,
      weight: 2,
    }).addTo(portalMap)

    // Pulse dot for User location
    const pulseHtml = `
      <div class="relative flex h-5 w-5">
        <span class="animate-ping absolute inline-flex h-full w-full rounded-full bg-blue-400 opacity-75"></span>
        <span class="relative inline-flex rounded-full h-5 w-5 bg-blue-600 border-2 border-white shadow-md"></span>
      </div>
    `
    const pulseIcon = L.divIcon({
      className: 'custom-user-marker',
      html: pulseHtml,
      iconSize: [20, 20],
      iconAnchor: [10, 10],
    })

    userMarker = L.marker([userLat, userLng], { icon: pulseIcon }).addTo(portalMap)
    userMarker.bindPopup('<b>Lokasi Anda Sekarang</b>')

    // Automatically zoom & fit bounds to show both markers
    const group = L.featureGroup([userMarker, officeMarker, geofenceCircle])
    portalMap.fitBounds(group.getBounds().pad(0.25))
  } catch (err) {
    console.error('Error rendering Leaflet Map:', err)
  }
}

// Distance calculation to branch office
const distanceToOffice = computed(() => {
  if (!latitude.value || !longitude.value || !employeeProfile.value?.branch) return null
  const branch = employeeProfile.value.branch
  return Math.round(
    calculateDistance(latitude.value, longitude.value, branch.latitude, branch.longitude),
  )
})

const isWithinRadius = computed(() => {
  if (distanceToOffice.value === null || !employeeProfile.value?.branch) return false
  return distanceToOffice.value <= employeeProfile.value.branch.radius_meter
})

const calculateDistance = (lat1: number, lon1: number, lat2: number, lon2: number) => {
  const R = 6371e3 // metres
  const phi1 = (lat1 * Math.PI) / 180
  const phi2 = (lat2 * Math.PI) / 180
  const deltaPhi = ((lat2 - lat1) * Math.PI) / 180
  const deltaLambda = ((lon2 - lon1) * Math.PI) / 180

  const a =
    Math.sin(deltaPhi / 2) * Math.sin(deltaPhi / 2) +
    Math.cos(phi1) * Math.cos(phi2) * Math.sin(deltaLambda / 2) * Math.sin(deltaLambda / 2)
  const c = 2 * Math.atan2(Math.sqrt(a), Math.sqrt(1 - a))

  return R * c
}

// Load initial data
const loadData = async () => {
  isLoading.value = true
  try {
    // 1. Get profile
    const profileRes = await axios.get('/employees/profile')
    employeeProfile.value = profileRes.data.data

    // 2. Get leave balances
    const balanceRes = await axios.get('/time-off/balances')
    leaveBalances.value = balanceRes.data.data || []

    // 3. Get leave history
    const leaveHistoryRes = await axios.get('/time-off/history')
    leaveHistory.value = leaveHistoryRes.data.data || []

    // 4. Get attendance history
    const attHistoryRes = await axios.get('/attendance/history')
    const historyData = attHistoryRes.data.data || []
    attendanceHistory.value = historyData.map((att: any) => ({
      ...att,
      clock_in_time: att.check_in,
      clock_out_time: att.check_out,
    }))
  } catch (err: any) {
    console.error(err)
    toast.error('Gagal memuat profil atau riwayat data')
  } finally {
    isLoading.value = false
  }
}

const handleLogout = () => {
  authStore.logout()
  router.push('/login')
}

// Selfie captured handlers
const triggerSelfieCapture = () => {
  selfieInput.value?.click()
}

const handleSelfieCaptured = async (e: any) => {
  const file = e.target.files[0]
  if (!file) return

  capturedSelfie.value = null
  capturedSelfieFile.value = file
  isProcessingSelfie.value = true

  const reader = new FileReader()
  reader.onload = async (event: any) => {
    const rawImage = event.target.result
    try {
      const branchName = employeeProfile.value?.branch?.name || 'Cabang Kantor'
      let locationAddress = ''
      if (isOutOfTown.value && outOfTownDestination.value.trim()) {
        locationAddress = `Dinas: ${outOfTownDestination.value.trim()}`
      } else if (latitude.value && longitude.value) {
        locationAddress = await reverseGeocodeAddress(latitude.value, longitude.value)
      } else {
        locationAddress = 'Lokasi Presensi Karyawan'
      }

      const watermarkedImage = await burnAttendanceWatermark(rawImage, {
        timeText: formatAttendanceDateTime(new Date()),
        branchName,
        addressText: locationAddress,
        targetWidth: 720,
        cropSquare: true,
        quality: 0.8,
      })
      capturedSelfie.value = watermarkedImage
    } catch (err) {
      console.error('Error applying watermark to selfie:', err)
      capturedSelfie.value = rawImage
    } finally {
      isProcessingSelfie.value = false
    }
  }
  reader.onerror = () => {
    capturedSelfie.value = null
    isProcessingSelfie.value = false
  }
  reader.readAsDataURL(file)
}

const removeSelfie = () => {
  capturedSelfie.value = null
  capturedSelfieFile.value = null
  isProcessingSelfie.value = false
  if (selfieInput.value) {
    selfieInput.value.value = ''
  }
}

// Face register handlers
const triggerFaceCapture = () => {
  facePhotoInput.value?.click()
}

const handleFacePhotoCaptured = (e: any) => {
  const file = e.target.files[0]
  if (!file) return
  capturedFaceFile.value = file

  const reader = new FileReader()
  reader.onload = async (event: any) => {
    const rawImage = event.target.result
    const croppedImage = await detectAndCropFace(rawImage)
    capturedFacePhoto.value = croppedImage
  }
  reader.readAsDataURL(file)
}

const removeFacePhoto = () => {
  capturedFacePhoto.value = null
  capturedFaceFile.value = null
  if (facePhotoInput.value) {
    facePhotoInput.value.value = ''
  }
}

const handleFaceRegisterSubmit = async () => {
  if (!capturedFacePhoto.value) return
  isSubmittingFace.value = true
  
  try {
    let base64Photo = capturedFacePhoto.value
    const commaIndex = base64Photo.indexOf(',')
    if (commaIndex !== -1) {
      base64Photo = base64Photo.substring(commaIndex + 1)
    }

    const mockEmbedding = Array.from({ length: 128 }, () => 0.0)

    const response = await axios.post('/users/face-register', {
      embedding: mockEmbedding,
      selfie: base64Photo
    })

    if (response.data.status === 'success' || response.status === 200) {
      toast.success('Pembaruan Wajah Berhasil', {
        description: 'Data biometrik wajah Anda telah diperbarui.'
      })
      if (employeeProfile.value) {
        employeeProfile.value.is_face_registered = true
      }
      removeFacePhoto()
    } else {
      toast.error('Gagal memperbarui wajah', {
        description: response.data.message || 'Terjadi kesalahan.'
      })
    }
  } catch (err: any) {
    console.error('Face registration failed', err)
    toast.error('Gagal memperbarui wajah', {
      description: err.response?.data?.message || 'Gangguan koneksi sistem.'
    })
  } finally {
    isSubmittingFace.value = false
  }
}

// Check in
const handleCheckIn = async () => {
  if (employeeProfile.value && !employeeProfile.value.is_face_registered) {
    toast.error('Gagal absen masuk', {
      description: 'Data wajah biometrik pembanding belum terdaftar. Silakan daftarkan wajah Anda di tab Profil & Akun.',
    })
    return
  }

  if (isProcessingSelfie.value) {
    toast.error('Foto selfie sedang diproses, silakan tunggu beberapa detik.')
    return
  }

  if (!capturedSelfie.value) {
    toast.error('Foto selfie diperlukan untuk absen masuk.')
    return
  }

  if (!latitude.value || !longitude.value) {
    toast.error('Lokasi GPS Anda diperlukan untuk absen masuk')
    return
  }

  if (isOutOfTown.value && !outOfTownDestination.value.trim()) {
    toast.error('Kota / Lokasi tujuan dinas wajib diisi untuk absen dinas luar kota.')
    return
  }

  isSubmitting.value = true
  try {
    let selfieBase64 = ''
    const commaIndex = capturedSelfie.value.indexOf(',')
    if (commaIndex !== -1) {
      selfieBase64 = capturedSelfie.value.substring(commaIndex + 1)
    } else {
      selfieBase64 = capturedSelfie.value
    }

    const payload = {
      latitude: latitude.value,
      longitude: longitude.value,
      selfie: selfieBase64,
      face_embedding: [], // Bypassed for web UI
      is_out_of_town: isOutOfTown.value,
      destination: isOutOfTown.value ? outOfTownDestination.value.trim() : '',
    }

    const res = await axios.post('/attendance/checkin', payload)
    const attData = res.data?.data

    toast.success('Absen masuk berhasil!')
    removeSelfie()
    await loadData()

    // If attendance status is LATE, prompt employee to enter late reason
    if (attData && attData.status === 'LATE') {
      lateAttendanceId.value = attData.id
      lateReasonInput.value = ''
      isLateReasonModalOpen.value = true
    }
  } catch (err: any) {
    toast.error(err.response?.data?.message || err.message || 'Gagal absen masuk')
  } finally {
    isSubmitting.value = false
  }
}

// Check out
const handleCheckOut = async () => {
  if (employeeProfile.value && !employeeProfile.value.is_face_registered) {
    toast.error('Gagal absen keluar', {
      description: 'Data wajah biometrik pembanding belum terdaftar. Silakan daftarkan wajah Anda di tab Profil & Akun.',
    })
    return
  }

  if (isProcessingSelfie.value) {
    toast.error('Foto selfie sedang diproses, silakan tunggu beberapa detik.')
    return
  }

  if (!capturedSelfie.value) {
    toast.error('Foto selfie diperlukan untuk absen keluar.')
    return
  }

  if (!latitude.value || !longitude.value) {
    toast.error('Lokasi GPS Anda diperlukan untuk absen keluar')
    return
  }

  isSubmitting.value = true
  try {
    let selfieBase64 = ''
    const commaIndex = capturedSelfie.value.indexOf(',')
    if (commaIndex !== -1) {
      selfieBase64 = capturedSelfie.value.substring(commaIndex + 1)
    } else {
      selfieBase64 = capturedSelfie.value
    }

    const payload = {
      latitude: latitude.value,
      longitude: longitude.value,
      selfie: selfieBase64,
      face_embedding: [], // Bypassed for web UI
      is_out_of_town: isOutOfTown.value,
      destination: isOutOfTown.value ? outOfTownDestination.value.trim() : '',
    }

    await axios.post('/attendance/checkout', payload)
    toast.success('Absen keluar berhasil!')
    removeSelfie()
    await loadData()
  } catch (err: any) {
    toast.error(err.response?.data?.message || err.message || 'Gagal absen keluar')
  } finally {
    isSubmitting.value = false
  }
}

// Submit late reason notes
const handleLateReasonSubmit = async () => {
  if (!lateReasonInput.value.trim()) {
    toast.error('Alasan keterlambatan wajib diisi.')
    return
  }

  if (!lateAttendanceId.value) {
    isLateReasonModalOpen.value = false
    return
  }

  isSubmittingLateReason.value = true
  try {
    await axios.put(`/attendance/my-notes/${lateAttendanceId.value}`, {
      notes: lateReasonInput.value.trim(),
    })
    toast.success('Catatan alasan keterlambatan berhasil disimpan!')
    isLateReasonModalOpen.value = false
    lateAttendanceId.value = ''
    lateReasonInput.value = ''
    await loadData()
  } catch (err: any) {
    toast.error(err.response?.data?.message || err.message || 'Gagal menyimpan alasan keterlambatan')
  } finally {
    isSubmittingLateReason.value = false
  }
}

// Submit leave request
const onLeaveFileChange = (e: any) => {
  const file = e.target.files[0]
  if (file) {
    leaveForm.value.attachment_file = file
  }
}

const handleLeaveSubmit = async () => {
  if (
    !leaveForm.value.leave_type_id ||
    !leaveForm.value.start_date ||
    !leaveForm.value.end_date ||
    !leaveForm.value.reason
  ) {
    toast.error('Semua kolom formulir cuti/izin wajib diisi')
    return
  }

  isSubmitting.value = true
  try {
    const formData = new FormData()
    formData.append('leave_type_id', leaveForm.value.leave_type_id)
    formData.append('start_date', leaveForm.value.start_date)
    formData.append('end_date', leaveForm.value.end_date)
    formData.append('reason', leaveForm.value.reason)
    if (leaveForm.value.attachment_file) {
      formData.append('attachment', leaveForm.value.attachment_file)
    }

    await axios.post('/time-off/request', formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
    })

    toast.success('Pengajuan cuti/izin berhasil diajukan!')
    leaveForm.value = {
      leave_type_id: '',
      start_date: '',
      end_date: '',
      reason: '',
      attachment_file: null,
    }

    // Clear input file element if possible
    const fileInput = document.getElementById('leave-attachment-input') as HTMLInputElement
    if (fileInput) fileInput.value = ''

    loadData()
  } catch (err: any) {
    toast.error(err.response?.data?.message || err.message || 'Gagal mengajukan izin')
  } finally {
    isSubmitting.value = false
  }
}

const getTodayAttendance = computed(() => {
  if (attendanceHistory.value.length === 0) return null

  const getLocalDateStr = (d: Date) => {
    const year = d.getFullYear()
    const month = String(d.getMonth() + 1).padStart(2, '0')
    const day = String(d.getDate()).padStart(2, '0')
    return `${year}-${month}-${day}`
  }

  const todayStr = getLocalDateStr(new Date())
  return attendanceHistory.value.find((att) => {
    const clockInStr = att.clock_in_time
      ? getLocalDateStr(new Date(att.clock_in_time))
      : ''
    return clockInStr === todayStr
  })
})

const getShiftText = computed(() => {
  if (employeeProfile.value?.default_shift) {
    const s = employeeProfile.value.default_shift
    return `${s.name} (${s.start_time.slice(0, 5)} - ${s.end_time.slice(0, 5)})`
  }
  return 'Regular Shift (09:00 - 18:00)'
})

// Watch GPS coordinate changes & Profile changes to update Map
watch(
  [latitude, longitude, employeeProfile],
  () => {
    if (latitude.value && longitude.value && employeeProfile.value?.branch) {
      setTimeout(() => {
        initPortalMap()
      }, 150)
    }
  },
  { deep: true },
)

onMounted(() => {
  updateClock()
  const timer = setInterval(updateClock, 1000)
  fetchGps()
  loadData()

  onUnmounted(() => {
    clearInterval(timer)
    if (portalMap) {
      portalMap.remove()
      portalMap = null
    }
  })
})

const formatDate = (dateStr: string) => {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleDateString('id-ID', {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
  })
}

const formatDateTime = (dateStr: string) => {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleTimeString('id-ID', { hour: '2-digit', minute: '2-digit' })
}

const thisMonthStats = computed(() => {
  const now = new Date()
  const currentYear = now.getFullYear()
  const currentMonth = now.getMonth()

  const checkedInCount = attendanceHistory.value.filter(
    (att) => att.status === 'ON_TIME' || att.status === 'LATE' || att.status === 'PRESENT'
  ).length

  const alphaCount = attendanceHistory.value.filter(
    (att) => att.status === 'ALFA' || att.status === 'ABSENT'
  ).length

  let leaveDaysCount = 0
  leaveHistory.value.forEach((leave) => {
    if (leave.status !== 'APPROVED') return

    const start = new Date(leave.start_date)
    const end = new Date(leave.end_date)
    
    const d = new Date(start)
    while (d <= end) {
      if (d.getFullYear() === currentYear && d.getMonth() === currentMonth) {
        leaveDaysCount++
      }
      d.setDate(d.getDate() + 1)
    }
  })

  return {
    checkedIn: checkedInCount,
    alpha: alphaCount,
    leave: leaveDaysCount
  }
})
</script>

<template>
  <div class="max-w-xl mx-auto px-4 sm:px-0 space-y-4 pb-20">
    <!-- Header/User Profile -->
    <div
      class="bg-gradient-to-br from-slate-900 via-slate-800 to-slate-900 text-white rounded-2xl sm:rounded-3xl p-4 sm:p-6 shadow-lg sm:shadow-xl relative overflow-hidden"
    >
      <!-- Glow effect -->
      <div class="absolute -top-12 -right-12 w-32 h-32 bg-primary/25 rounded-full blur-2xl"></div>

      <div class="flex items-center justify-between relative z-10">
        <div class="flex items-center gap-2.5 sm:gap-4">
          <div
            class="w-10 h-10 sm:w-14 sm:h-14 rounded-xl sm:rounded-2xl bg-primary flex items-center justify-center font-black text-sm sm:text-xl shadow-md border border-white/10 text-white"
          >
            {{ employeeProfile ? employeeProfile.first_name[0] : 'U' }}
          </div>
          <div>
            <h2 class="text-sm sm:text-lg font-extrabold tracking-tight">
              {{
                employeeProfile
                  ? `${employeeProfile.first_name} ${employeeProfile.last_name || ''}`
                  : 'Memuat...'
              }}
            </h2>
            <p class="text-[10px] sm:text-xs text-slate-300 font-semibold uppercase tracking-wider mt-0.5 sm:mt-1">
              {{ employeeProfile?.job_position?.title || 'Karyawan' }}
            </p>
          </div>
        </div>
        <div class="flex items-center gap-1.5 sm:gap-2">
          <!-- Back to Admin Dashboard for HR/Admin/Manager -->
          <router-link
            v-if="authStore.user?.role !== 'employee'"
            to="/"
            class="px-2.5 py-1.5 sm:px-3.5 sm:py-2 rounded-lg sm:rounded-xl bg-white/10 hover:bg-white/20 text-slate-250 hover:text-white transition-all active:scale-95 text-[10px] sm:text-xs font-black uppercase tracking-wider"
          >
            Admin
          </router-link>

          <button
            @click="handleLogout"
            class="p-2 sm:p-2.5 rounded-lg sm:rounded-xl bg-white/10 hover:bg-white/20 text-slate-200 hover:text-white transition-all active:scale-95"
          >
            <LogOut class="w-4 h-4 sm:w-5 sm:h-5" />
          </button>
        </div>
      </div>

      <div
        class="mt-4 pt-3 border-t border-white/10 grid grid-cols-2 gap-3 text-[11px] sm:text-xs relative z-10 text-slate-300"
      >
        <div>
          <span class="block text-[9px] sm:text-[10px] uppercase font-black text-slate-400">Kantor Cabang</span>
          <span class="font-bold text-white mt-0.5 block">{{
            employeeProfile?.branch?.name || 'Head Office'
          }}</span>
        </div>
        <div>
          <span class="block text-[9px] sm:text-[10px] uppercase font-black text-slate-400">Jadwal Shift</span>
          <span class="font-bold text-white mt-0.5 block">{{ getShiftText }}</span>
        </div>
      </div>

      <!-- Ringkasan Kehadiran Bulan Ini -->
      <div
        class="mt-3.5 pt-3 border-t border-white/10 grid grid-cols-3 gap-2 text-center relative z-10"
      >
        <div class="bg-white/5 border border-white/10 rounded-xl p-1.5 sm:p-2.5 space-y-0.5">
          <span class="block text-[8px] sm:text-[10px] font-black text-slate-400 uppercase tracking-wider">Hadir</span>
          <span class="block text-base sm:text-2xl font-black text-white leading-none py-0.5 sm:py-1">{{ thisMonthStats.checkedIn }}</span>
          <span class="block text-[8px] sm:text-[9px] font-bold text-slate-400">Hari</span>
        </div>
        <div class="bg-white/5 border border-white/10 rounded-xl p-1.5 sm:p-2.5 space-y-0.5">
          <span class="block text-[8px] sm:text-[10px] font-black text-slate-400 uppercase tracking-wider">Cuti/Izin</span>
          <span class="block text-base sm:text-2xl font-black text-white leading-none py-0.5 sm:py-1">{{ thisMonthStats.leave }}</span>
          <span class="block text-[8px] sm:text-[9px] font-bold text-slate-400">Hari</span>
        </div>
        <div class="bg-white/5 border border-white/10 rounded-xl p-1.5 sm:p-2.5 space-y-0.5">
          <span class="block text-[8px] sm:text-[10px] font-black text-slate-400 uppercase tracking-wider">Alfa</span>
          <span class="block text-base sm:text-2xl font-black text-white leading-none py-0.5 sm:py-1">{{ thisMonthStats.alpha }}</span>
          <span class="block text-[8px] sm:text-[9px] font-bold text-slate-400">Hari</span>
        </div>
      </div>
    </div>

    <!-- Tab Navigation (Mobile-like Bottom/Top Navigation) -->
    <div class="flex bg-slate-105 p-1 rounded-xl sm:p-1.5 sm:rounded-2xl text-[10px] sm:text-xs font-extrabold shadow-inner gap-0.5 sm:gap-1 relative z-10 bg-slate-100">
      <button
        @click="currentTab = 'attendance'"
        type="button"
        class="flex-1 flex flex-col sm:flex-row items-center justify-center gap-1 py-2 sm:py-3 rounded-lg sm:rounded-xl transition-all duration-300"
        :class="
          currentTab === 'attendance'
            ? 'bg-white text-primary shadow-md shadow-primary/5 ring-1 ring-slate-200/30'
            : 'text-slate-500 hover:text-slate-800'
        "
      >
        <Clock class="w-3.5 h-3.5 sm:w-4 sm:h-4" />
        <span>Absensi</span>
      </button>
      <button
        @click="currentTab = 'leaves'"
        type="button"
        class="flex-1 flex flex-col sm:flex-row items-center justify-center gap-1 py-2 sm:py-3 rounded-lg sm:rounded-xl transition-all duration-300"
        :class="
          currentTab === 'leaves'
            ? 'bg-white text-primary shadow-md shadow-primary/5 ring-1 ring-slate-200/30'
            : 'text-slate-500 hover:text-slate-800'
        "
      >
        <Calendar class="w-3.5 h-3.5 sm:w-4 sm:h-4" />
        <span>Cuti & Izin</span>
      </button>
      <button
        @click="currentTab = 'profile'"
        type="button"
        class="flex-1 flex flex-col sm:flex-row items-center justify-center gap-1 py-2 sm:py-3 rounded-lg sm:rounded-xl transition-all duration-300"
        :class="
          currentTab === 'profile'
            ? 'bg-white text-primary shadow-md shadow-primary/5 ring-1 ring-slate-200/30'
            : 'text-slate-500 hover:text-slate-800'
        "
      >
        <User class="w-3.5 h-3.5 sm:w-4 sm:h-4" />
        <span>Profil & Akun</span>
      </button>
    </div>

    <Transition name="fade-slide" mode="out-in">
      <div :key="currentTab" class="space-y-4">
        <!-- Content for TAB: Absensi (Attendance) -->
        <div v-if="currentTab === 'attendance'" class="space-y-4">
      <!-- Live Clock & Attendance Panel -->
      <div
        class="bg-white rounded-2xl sm:rounded-3xl border border-slate-200/60 p-4 sm:p-6 shadow-sm flex flex-col items-center justify-center text-center space-y-3 sm:space-y-4"
      >
        <Clock class="w-6 h-6 sm:w-8 sm:h-8 text-primary" />
        <div>
          <h3 class="text-2xl sm:text-3xl font-black text-slate-900 tracking-tight leading-none">
            {{ liveTime }}
          </h3>
          <p class="text-[10px] sm:text-xs text-slate-500 font-bold mt-1 sm:mt-1.5 uppercase tracking-wide">
            {{ liveDate }}
          </p>
        </div>

        <!-- GPS status indicator -->
        <div class="w-full pt-3 sm:pt-4 border-t border-slate-100 flex flex-col items-center space-y-1 sm:space-y-2">
          <div
            class="flex items-center gap-1 sm:gap-1.5 text-[10px] sm:text-xs font-bold"
            :class="latitude && longitude ? 'text-green-600' : 'text-red-500'"
          >
            <MapPin class="w-3.5 h-3.5 sm:w-4 sm:h-4 shrink-0" :class="{ 'animate-bounce': isFetchingGps }" />
            <span v-if="isFetchingGps">Mengambil Lokasi GPS...</span>
            <span v-else-if="latitude && longitude">GPS Aktif (Akurasi Tinggi)</span>
            <span v-else>GPS Tidak Aktif / Diizinkan</span>
          </div>

          <p v-if="latitude && longitude" class="text-[9px] sm:text-[11px] text-slate-400 font-medium leading-none">
            Koordinat: {{ latitude.toFixed(6) }}, {{ longitude.toFixed(6) }}
          </p>

          <!-- Distance and radius text -->
          <p v-if="distanceToOffice !== null" class="text-xs text-slate-600 font-bold">
            Jarak ke kantor:
            <span :class="isWithinRadius ? 'text-green-600' : 'text-red-600'"
              >{{ distanceToOffice }}m</span
            >
            (Maks: {{ employeeProfile?.branch?.radius_meter }}m)
          </p>

          <button
            v-if="!latitude || gpsError"
            @click="fetchGps"
            type="button"
            class="text-xs text-primary font-bold hover:underline flex items-center gap-1"
          >
            <RefreshCw class="w-3.5 h-3.5" />
            Muat Ulang Lokasi GPS
          </button>

          <!-- GPS Error and Instructions Alert -->
          <div
            v-if="gpsError"
            class="bg-amber-50 border border-amber-200 rounded-xl p-3 sm:p-4 text-left text-xs text-amber-800 w-full space-y-2 mt-2"
          >
            <div class="flex gap-2 items-start">
              <AlertTriangle class="w-4 h-4 sm:w-5 sm:h-5 text-amber-600 shrink-0 mt-0.5" />
              <div>
                <p class="font-bold text-amber-900 text-[11px] sm:text-xs">Akses Lokasi (GPS) Dibatasi</p>
                <p class="mt-0.5 text-amber-700 leading-relaxed font-medium text-[10px] sm:text-[11px]">
                  {{ gpsError }}
                </p>
              </div>
            </div>
            
            <div class="border-t border-amber-200/60 pt-2 text-[10px] sm:text-[11px] space-y-1.5 font-medium text-amber-800/90">
              <p class="font-bold text-amber-900 uppercase tracking-wide text-[9px]">Cara Mengaktifkan Izin Lokasi:</p>
              
              <!-- Safari iOS instructions -->
              <div class="flex gap-1 items-start">
                <span class="font-bold text-amber-900">1.</span>
                <p>
                  <strong>Safari (iPhone/iPad):</strong> Ketuk tombol <span class="bg-amber-100 px-1 py-0.5 rounded border border-amber-200 text-[9px] font-mono font-bold text-amber-900">aA</span> atau <span class="bg-amber-100 px-1 py-0.5 rounded border border-amber-200 text-[9px] font-mono font-bold text-amber-900">🧩</span> di sebelah kiri alamat web (URL) &rarr; Pilih <strong>Pengaturan Situs Web</strong> &rarr; Ubah izin <strong>Lokasi</strong> menjadi <strong>Izinkan</strong> (Allow) &rarr; Muat ulang halaman.
                </p>
              </div>

              <!-- Safari Mac instructions -->
              <div class="flex gap-1 items-start">
                <span class="font-bold text-amber-900">2.</span>
                <p>
                  <strong>Safari (MacBook/Mac):</strong> Klik menu <strong>Safari</strong> di kiri atas &rarr; <strong>Pengaturan...</strong> (Settings/Preferences) &rarr; Tab <strong>Situs Web</strong> &rarr; Pilih <strong>Lokasi</strong> di kolom kiri &rarr; Cari situs ini dan ubah menjadi <strong>Izinkan</strong> (Allow).
                </p>
              </div>

              <!-- Chrome/Android/PC instructions -->
              <div class="flex gap-1 items-start">
                <span class="font-bold text-amber-900">3.</span>
                <p>
                  <strong>Chrome / Browser Lain:</strong> Ketuk ikon gembok/pengaturan di kiri kolom alamat URL &rarr; Aktifkan/Izinkan <strong>Lokasi</strong> &rarr; Muat ulang halaman web.
                </p>
              </div>
              
              <!-- Device GPS settings -->
              <div class="flex gap-1 items-start">
                <span class="font-bold text-amber-900">4.</span>
                <p>
                  Pastikan <strong>GPS / Layanan Lokasi</strong> pada menu pengaturan utama HP atau laptop Anda dalam keadaan <strong>Aktif (ON)</strong>.
                </p>
              </div>
            </div>
          </div>

          <!-- Leaflet Map Container -->
          <div v-show="latitude && longitude" class="w-full mt-2.5">
            <div
              ref="portalMapContainer"
              class="h-32 sm:h-44 w-full rounded-xl sm:rounded-2xl overflow-hidden border border-slate-200 z-0"
            ></div>
            <span class="text-[9px] sm:text-[10px] text-slate-400 font-semibold mt-1 block">
              *Pin Biru: Lokasi Anda | Lingkaran Hijau: Geofence Kantor
            </span>
          </div>
        </div>

        <!-- Dinas Luar Kota Toggle & Destination Section -->
        <div class="w-full pt-3 sm:pt-4 border-t border-slate-100">
          <div
            class="p-3.5 sm:p-4 rounded-2xl border transition-all duration-200"
            :class="
              isOutOfTown
                ? 'bg-indigo-50/70 border-indigo-200 text-indigo-900 shadow-sm'
                : 'bg-slate-50/60 border-slate-200/80 text-slate-700'
            "
          >
            <div class="flex items-center justify-between gap-3">
              <div class="flex items-center gap-2.5 sm:gap-3">
                <div
                  class="w-8 h-8 sm:w-9 sm:h-9 rounded-xl flex items-center justify-center transition-colors"
                  :class="
                    isOutOfTown
                      ? 'bg-indigo-600 text-white shadow-sm'
                      : 'bg-slate-200 text-slate-600'
                  "
                >
                  <Plane class="w-4 h-4 sm:w-4.5 sm:h-4.5" />
                </div>
                <div>
                  <h4 class="text-xs sm:text-sm font-extrabold tracking-tight">
                    Dinas Luar Kota
                  </h4>
                  <p class="text-[10px] sm:text-[11px] text-slate-500 font-medium">
                    {{ isOutOfTown ? 'Bypass radius kantor cabang aktif' : 'Aktifkan jika bertugas di luar area kantor' }}
                  </p>
                </div>
              </div>

              <!-- Switch button -->
              <button
                type="button"
                @click="isOutOfTown = !isOutOfTown"
                class="relative inline-flex h-6 w-11 flex-shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none"
                :class="isOutOfTown ? 'bg-indigo-600' : 'bg-slate-300'"
              >
                <span
                  class="pointer-events-none inline-block h-5 w-5 transform rounded-full bg-white shadow ring-0 transition duration-200 ease-in-out"
                  :class="isOutOfTown ? 'translate-x-5' : 'translate-x-0'"
                />
              </button>
            </div>

            <!-- Destination Input if Dinas is Active -->
            <Transition
              enter-active-class="transition duration-200 ease-out"
              enter-from-class="transform -translate-y-2 opacity-0"
              enter-to-class="transform translate-y-0 opacity-100"
              leave-active-class="transition duration-150 ease-in"
              leave-from-class="transform translate-y-0 opacity-100"
              leave-to-class="transform -translate-y-2 opacity-0"
            >
              <div v-if="isOutOfTown" class="mt-3 pt-3 border-t border-indigo-200/60 space-y-1.5">
                <label class="text-[11px] font-bold text-indigo-900 uppercase tracking-wider block">
                  Kota / Lokasi Tujuan Dinas <span class="text-red-500">*</span>
                </label>
                <div class="relative">
                  <input
                    type="text"
                    v-model="outOfTownDestination"
                    placeholder="Contoh: Surabaya, Pabrik Karawang, dsb."
                    class="w-full h-10 px-3.5 rounded-xl border border-indigo-200 bg-white text-xs font-semibold text-slate-900 placeholder:text-slate-400 focus:outline-none focus:ring-2 focus:ring-indigo-500/30 focus:border-indigo-500"
                  />
                  <MapPin class="w-4 h-4 text-indigo-500 absolute right-3 top-3 pointer-events-none" />
                </div>
              </div>
            </Transition>
          </div>
        </div>

        <!-- Warning info if outside radius and not on business trip -->
        <div
          v-if="latitude && !isWithinRadius && !isOutOfTown"
          class="bg-red-50 border border-red-100 rounded-2xl p-4 flex gap-3 items-start text-left text-xs text-red-700 w-full"
        >
          <AlertTriangle class="w-5 h-5 text-red-500 shrink-0 mt-0.5" />
          <div>
            <p class="font-bold">Di Luar Radius Kantor</p>
            <p class="mt-0.5 text-red-600/90 leading-relaxed">
              Anda berada di luar jangkauan absensi kantor. Jika Anda sedang bertugas di luar kota/kantor, silakan aktifkan opsi <strong>Dinas Luar Kota</strong> di atas.
            </p>
          </div>
        </div>

        <!-- Warning info if face is not registered -->
        <div
          v-if="employeeProfile && !employeeProfile.is_face_registered"
          class="bg-amber-50 border border-amber-200 rounded-2xl p-4 flex gap-3 items-start text-left text-xs text-amber-800 w-full shadow-sm"
        >
          <AlertCircle class="w-5 h-5 text-amber-500 shrink-0 mt-0.5" />
          <div class="flex-1 space-y-1.5">
            <p class="font-black text-amber-900">Data Wajah Belum Terdaftar</p>
            <p class="text-amber-700 leading-relaxed">
              Anda belum meregistrasikan data wajah pembanding biometrik. Anda wajib meregistrasikan wajah terlebih dahulu sebelum dapat melakukan absen masuk atau absen keluar.
            </p>
            <button
              @click="currentTab = 'profile'"
              type="button"
              class="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-amber-600 hover:bg-amber-700 text-white font-extrabold text-[10px] uppercase tracking-wide transition-all active:scale-95 shadow-sm mt-1"
            >
              <Camera class="w-3.5 h-3.5" />
              Daftarkan Wajah Sekarang
            </button>
          </div>
        </div>

        <!-- Selfie Capture Section -->
        <div
          v-if="latitude && (isWithinRadius || isOutOfTown) && employeeProfile?.is_face_registered"
          class="w-full border-t border-slate-100 pt-3 sm:pt-4 flex flex-col items-center"
        >
          <span
            class="text-[10px] sm:text-xs font-bold text-slate-600 uppercase tracking-wider pl-1 mb-1.5 block self-start"
          >
            Foto Selfie Kehadiran
          </span>

          <!-- Camera capture picker container (using label for native Safari file trigger) -->
          <label
            for="selfie-file-input"
            class="group relative w-full h-28 sm:h-40 rounded-xl sm:rounded-2xl border-2 border-dashed flex flex-col items-center justify-center transition-all duration-200 overflow-hidden"
            :class="[
              isProcessingSelfie
                ? 'border-slate-200 bg-slate-50/80 cursor-wait pointer-events-none'
                : capturedSelfie
                  ? 'border-primary bg-slate-50 cursor-pointer'
                  : 'border-slate-300 hover:border-primary hover:bg-slate-50/50 bg-slate-50/20 cursor-pointer'
            ]"
          >
            <!-- Hidden Input (Safari-safe invisible styling instead of display:none) -->
            <input
              id="selfie-file-input"
              type="file"
              accept="image/*"
              capture="user"
              @change="handleSelfieCaptured"
              ref="selfieInput"
              style="opacity: 0; position: absolute; width: 1px; height: 1px; overflow: hidden; pointer-events: none;"
            />

            <!-- Loading Spinner State -->
            <template v-if="isProcessingSelfie">
              <div class="flex flex-col items-center text-center space-y-2 px-4">
                <div class="w-8 h-8 rounded-full border-4 border-slate-200 border-t-primary animate-spin"></div>
                <div>
                  <p class="text-xs font-black text-slate-700">Mendeteksi Wajah...</p>
                  <p class="text-[10px] text-slate-400 font-medium">Memproses foto selfie Anda</p>
                </div>
              </div>
            </template>

            <!-- Preview image if captured -->
            <template v-else-if="capturedSelfie">
              <img :src="capturedSelfie" class="w-full h-full object-cover" />

              <!-- Overlay controls -->
              <div
                class="absolute inset-0 bg-black/40 opacity-0 group-hover:opacity-100 transition-opacity duration-200 flex items-center justify-center gap-3"
              >
                <div
                  class="p-2.5 rounded-full bg-white/95 hover:bg-white text-slate-800 shadow-md font-bold text-xs flex items-center gap-1.5 active:scale-95 transition-transform cursor-pointer"
                >
                  <Camera class="w-4 h-4" />
                  Foto Ulang
                </div>
                <button
                  type="button"
                  @click.stop="removeSelfie"
                  class="p-2.5 rounded-full bg-red-600 hover:bg-red-700 text-white shadow-md font-bold text-xs flex items-center gap-1.5 active:scale-95 transition-transform"
                >
                  <Trash2 class="w-4 h-4" />
                  Hapus
                </button>
              </div>

              <!-- Quick indicator on bottom of photo -->
              <div
                class="absolute bottom-2 left-2 bg-black/60 text-white text-[10px] font-bold px-2 py-0.5 rounded-md backdrop-blur-sm"
              >
                Foto Siap
              </div>
            </template>

            <!-- Placeholder state if not captured -->
            <template v-else>
              <div class="flex flex-col items-center text-center space-y-2 px-4">
                <div
                  class="w-12 h-12 rounded-full bg-slate-100 flex items-center justify-center text-slate-500 group-hover:bg-primary/10 group-hover:text-primary transition-colors"
                >
                  <Camera class="w-6 h-6" />
                </div>
                <div>
                  <p
                    class="text-xs font-black text-slate-700 group-hover:text-primary transition-colors"
                  >
                    Ambil Foto Selfie
                  </p>
                  <p class="text-[10px] text-slate-400 font-medium">
                    Ketuk untuk mengambil foto dari kamera Anda
                  </p>
                </div>
              </div>
            </template>
          </label>
        </div>

        <!-- Action Buttons -->
        <div class="grid grid-cols-2 gap-3 sm:gap-4 w-full pt-3 sm:pt-4">
          <!-- Clock In Button -->
          <button
            @click="handleCheckIn"
            type="button"
            :disabled="
              isSubmitting ||
              (!isWithinRadius && !isOutOfTown) ||
              !employeeProfile?.is_face_registered ||
              !!getTodayAttendance?.clock_in_time ||
              !capturedSelfie ||
              isProcessingSelfie ||
              (isOutOfTown && !outOfTownDestination.trim())
            "
            class="flex flex-col items-center justify-center gap-1 sm:gap-2 h-14 sm:h-20 rounded-xl sm:rounded-2xl border text-[11px] sm:text-xs font-black transition-all shadow-sm"
            :class="[
              getTodayAttendance?.clock_in_time
                ? 'bg-green-50 border-green-200 text-green-700'
                : ((!isWithinRadius && !isOutOfTown) || !employeeProfile?.is_face_registered || !capturedSelfie || isProcessingSelfie || (isOutOfTown && !outOfTownDestination.trim()))
                  ? 'bg-slate-50 border-slate-200 text-slate-400 cursor-not-allowed'
                  : isOutOfTown
                    ? 'bg-indigo-600 border-indigo-600 hover:bg-indigo-700 text-white active:scale-95 shadow-indigo-600/20'
                    : 'bg-primary border-primary hover:bg-primary/95 text-white active:scale-95',
            ]"
          >
            <CheckCircle2 class="w-4 h-4 sm:w-5 sm:h-5" />
            <span>{{ getTodayAttendance?.clock_in_time ? 'SUDAH MASUK' : (isOutOfTown ? 'ABSEN DINAS MASUK' : 'ABSEN MASUK') }}</span>
            <span
              v-if="getTodayAttendance?.clock_in_time"
              class="text-[9px] sm:text-[10px] font-medium text-green-600"
            >
              {{ formatDateTime(getTodayAttendance.clock_in_time) }}
            </span>
          </button>

          <!-- Clock Out Button -->
          <button
            @click="handleCheckOut"
            type="button"
            :disabled="
              isSubmitting ||
              (!isWithinRadius && !isOutOfTown) ||
              !employeeProfile?.is_face_registered ||
              !getTodayAttendance?.clock_in_time ||
              !!getTodayAttendance?.clock_out_time ||
              !capturedSelfie ||
              isProcessingSelfie
            "
            class="flex flex-col items-center justify-center gap-1 sm:gap-2 h-14 sm:h-20 rounded-xl sm:rounded-2xl border text-[11px] sm:text-xs font-black transition-all shadow-sm"
            :class="[
              getTodayAttendance?.clock_out_time
                ? 'bg-blue-50 border-blue-200 text-blue-700'
                : (!getTodayAttendance?.clock_in_time || (!isWithinRadius && !isOutOfTown) || !employeeProfile?.is_face_registered || !capturedSelfie || isProcessingSelfie)
                  ? 'bg-slate-50 border-slate-200 text-slate-400 cursor-not-allowed'
                  : 'bg-slate-900 border-slate-900 hover:bg-slate-800 text-white active:scale-95',
            ]"
          >
            <LogOut class="w-4 h-4 sm:w-5 sm:h-5" />
            <span>{{ getTodayAttendance?.clock_out_time ? 'SUDAH PULANG' : (isOutOfTown ? 'ABSEN DINAS PULANG' : 'ABSEN PULANG') }}</span>
            <span
              v-if="getTodayAttendance?.clock_out_time"
              class="text-[9px] sm:text-[10px] font-medium text-blue-600"
            >
              {{ formatDateTime(getTodayAttendance.clock_out_time) }}
            </span>
          </button>
        </div>
      </div>



      <!-- Riwayat Kehadiran Terbaru -->
      <div class="bg-white rounded-2xl sm:rounded-3xl border border-slate-200/60 p-4 sm:p-6 shadow-sm space-y-4 sm:space-y-6">
        <h3 class="text-xs sm:text-sm font-black text-slate-800 uppercase tracking-widest flex items-center gap-2">
          <History class="w-4 h-4 text-slate-500" />
          Riwayat Kehadiran Terbaru
        </h3>
        <div class="space-y-3">
          <div
            v-for="att in attendanceHistory.slice(0, 5)"
            :key="att.id"
            class="flex items-center justify-between p-2.5 sm:p-4 border border-slate-100 rounded-xl sm:rounded-2xl hover:bg-slate-50/50 transition-colors gap-2"
          >
            <div class="flex items-center gap-2.5 sm:gap-3 min-w-0 flex-1">
              <!-- Selfie Thumbnail -->
              <div
                v-if="att.selfie_url"
                class="w-8 h-8 sm:w-10 sm:h-10 rounded-lg sm:rounded-xl overflow-hidden bg-slate-100 border border-slate-200 shrink-0"
              >
                <img :src="att.selfie_url" class="w-full h-full object-cover" />
              </div>
              <div
                v-else
                class="w-8 h-8 sm:w-10 sm:h-10 rounded-lg sm:rounded-xl bg-slate-100 border border-slate-200 flex items-center justify-center shrink-0 text-slate-400"
              >
                <Camera class="w-4 h-4" />
              </div>

              <div class="space-y-1 min-w-0 flex-1">
                <div class="flex items-center gap-2">
                  <span class="text-[11px] sm:text-xs font-bold text-slate-800 block truncate">{{
                    formatDate(att.clock_in_time)
                  }}</span>
                  <span
                    v-if="att.is_out_of_town || att.status === 'BUSINESS_TRIP'"
                    class="px-1.5 py-0.2 rounded-md bg-indigo-50 text-indigo-700 border border-indigo-200 text-[9px] font-black flex items-center gap-0.5"
                  >
                    <span>✈️</span> DINAS
                  </span>
                </div>

                <div class="flex flex-wrap gap-x-2 gap-y-0.5 text-[10px] sm:text-[11px] text-slate-500 font-medium">
                  <span class="truncate"
                    >In: {{ att.clock_in_time ? formatDateTime(att.clock_in_time) : '-' }}</span
                  >
                  <span class="truncate"
                    >Out: {{ att.clock_out_time ? formatDateTime(att.clock_out_time) : '-' }}</span
                  >
                </div>

                <!-- Destination chip if dinas -->
                <div v-if="att.destination" class="text-[10px] text-indigo-700 font-semibold flex items-center gap-1">
                  <span>📍</span>
                  <span class="truncate">{{ att.destination }}</span>
                </div>

                <!-- Notes / Late reason if present -->
                <div
                  v-if="att.notes"
                  class="mt-1 text-[10px] sm:text-[11px] text-amber-900 bg-amber-50/80 border border-amber-200/60 rounded-lg px-2 py-1 flex items-start gap-1"
                >
                  <FileEdit class="w-3 h-3 text-amber-600 shrink-0 mt-0.5" />
                  <span class="leading-tight"><strong class="font-bold">Catatan:</strong> {{ att.notes }}</span>
                </div>
              </div>
            </div>

            <div class="flex flex-col items-end gap-1">
              <span
                class="px-2 sm:px-3 py-0.5 sm:py-1 rounded-full text-[9px] sm:text-[10px] font-bold shrink-0 text-center"
                :class="[
                  att.is_out_of_town || att.status === 'BUSINESS_TRIP'
                    ? 'bg-indigo-50 text-indigo-700 border border-indigo-200'
                    : att.status === 'PRESENT' || att.status === 'ON_TIME'
                      ? 'bg-green-50 text-green-700 border border-green-150'
                      : att.status === 'LATE'
                        ? 'bg-orange-50 text-orange-700 border border-orange-150'
                        : 'bg-red-50 text-red-700 border border-red-150',
                ]"
              >
                {{
                  att.is_out_of_town || att.status === 'BUSINESS_TRIP'
                    ? 'DINAS LUAR KOTA'
                    : att.status === 'PRESENT' || att.status === 'ON_TIME'
                      ? 'TEPAT WAKTU'
                      : att.status === 'LATE'
                        ? 'TERLAMBAT'
                        : att.status
                }}
              </span>
            </div>
          </div>
          <div v-if="attendanceHistory.length === 0" class="text-xs text-slate-400 text-center py-6">
            Belum ada riwayat kehadiran terdaftar.
          </div>
        </div>
      </div>
    </div>

    <!-- Content for TAB: Cuti & Izin (Leaves) -->
    <div v-else-if="currentTab === 'leaves'" class="space-y-6">
      <!-- Leave Submission Form -->
      <div class="bg-white rounded-3xl border border-slate-200/60 p-5 sm:p-6 shadow-sm space-y-6">
        <div>
          <h3
            class="text-sm font-black text-slate-800 uppercase tracking-widest flex items-center gap-2"
          >
            <FileText class="w-4 h-4 text-slate-500" />
            Pengajuan Cuti / Surat Izin
          </h3>
          <p class="text-xs text-slate-400 mt-1">
            Ajukan permohonan surat izin sakit, izin dinas, atau cuti tahunan langsung dari sini.
          </p>
        </div>

        <div class="space-y-4">
          <div class="grid gap-1.5">
            <label class="text-xs font-bold text-slate-600 uppercase tracking-wider pl-1"
              >Tipe Cuti/Izin</label
            >
            <select
              v-model="leaveForm.leave_type_id"
              class="h-11 w-full rounded-xl border border-slate-200 bg-white px-3 text-xs focus:ring-1 focus:ring-primary focus:outline-none"
            >
              <option value="" disabled selected>Pilih Jenis Izin</option>
              <option v-for="b in leaveBalances" :key="b.leave_type_id" :value="b.leave_type_id">
                {{ b.leave_type_name }} (Sisa: {{ b.remaining }} hari)
              </option>
            </select>
          </div>

          <div class="grid grid-cols-2 gap-4">
            <div class="grid gap-1.5">
              <label class="text-xs font-bold text-slate-600 uppercase tracking-wider pl-1"
                >Tanggal Mulai</label
              >
              <input
                type="date"
                v-model="leaveForm.start_date"
                class="h-11 w-full rounded-xl border border-slate-200 bg-white px-3 text-xs focus:ring-1 focus:ring-primary focus:outline-none"
              />
            </div>
            <div class="grid gap-1.5">
              <label class="text-xs font-bold text-slate-600 uppercase tracking-wider pl-1"
                >Tanggal Selesai</label
              >
              <input
                type="date"
                v-model="leaveForm.end_date"
                class="h-11 w-full rounded-xl border border-slate-200 bg-white px-3 text-xs focus:ring-1 focus:ring-primary focus:outline-none"
              />
            </div>
          </div>

          <div class="grid gap-1.5">
            <label class="text-xs font-bold text-slate-600 uppercase tracking-wider pl-1"
              >Alasan Pengajuan</label
            >
            <textarea
              v-model="leaveForm.reason"
              rows="3"
              placeholder="Alasan cuti/sakit secara singkat..."
              class="w-full rounded-xl border border-slate-200 bg-white p-3 text-xs focus:ring-1 focus:ring-primary focus:outline-none"
            ></textarea>
          </div>

          <div class="grid gap-1.5">
            <label
              class="text-xs font-bold text-slate-600 uppercase tracking-wider pl-1 flex items-center gap-1"
            >
              Lampiran
              <span class="text-[10px] text-slate-400 font-normal lowercase"
                >(Surat dokter/Surat Tugas)</span
              >
            </label>
            <input
              type="file"
              id="leave-attachment-input"
              @change="onLeaveFileChange"
              accept=".jpg,.jpeg,.png,.pdf"
              class="h-11 w-full rounded-xl border border-slate-200 bg-white px-3 py-2 text-xs focus:ring-1 focus:ring-primary focus:outline-none file:mr-4 file:py-1 file:px-3 file:rounded-md file:border-0 file:text-xs file:font-semibold file:bg-primary/10 file:text-primary file:cursor-pointer"
            />
          </div>

          <button
            @click="handleLeaveSubmit"
            type="button"
            :disabled="isSubmitting"
            class="w-full h-11 bg-primary hover:bg-primary/95 text-white font-black rounded-xl text-xs shadow-md shadow-primary/20 flex items-center justify-center gap-2 active:scale-98 transition-all"
          >
            <Send class="w-4 h-4" />
            <span>{{ isSubmitting ? 'MENGIRIM...' : 'KIRIM PENGAJUAN' }}</span>
          </button>
        </div>
      </div>

      <!-- Riwayat Pengajuan Cuti & Izin -->
      <div class="bg-white rounded-3xl border border-slate-200/60 p-5 sm:p-6 shadow-sm space-y-6">
        <h3
          class="text-sm font-black text-slate-800 uppercase tracking-widest flex items-center gap-2"
        >
          <History class="w-4 h-4 text-slate-500" />
          Riwayat Pengajuan Cuti & Izin
        </h3>

        <div class="space-y-3">
          <div
            v-for="leave in leaveHistory.slice(0, 5)"
            :key="leave.id"
            class="p-3.5 sm:p-4 border border-slate-100 rounded-2xl bg-white hover:bg-slate-50/30 transition-all space-y-3"
          >
            <div class="flex items-start justify-between gap-3">
              <div class="space-y-0.5">
                <span class="text-xs font-bold text-slate-800 block">{{
                  leave.leave_type?.name || 'Izin/Cuti'
                }}</span>
                <span class="text-[11px] text-slate-400 font-bold block">
                  {{ formatDate(leave.start_date) }} - {{ formatDate(leave.end_date) }}
                </span>
              </div>

              <span
                class="px-3 py-1 rounded-full text-[10px] font-bold shrink-0 text-center"
                :class="[
                  leave.status === 'APPROVED'
                    ? 'bg-green-50 text-green-700 border border-green-200'
                    : leave.status === 'REJECTED'
                      ? 'bg-red-50 text-red-700 border border-red-200'
                      : 'bg-yellow-50 text-yellow-700 border border-yellow-200',
                ]"
              >
                {{
                  leave.status === 'APPROVED'
                    ? 'DISETUJUI'
                    : leave.status === 'REJECTED'
                      ? 'DITOLAK'
                      : 'PENDING'
                }}
              </span>
            </div>

            <div
              class="text-xs text-slate-600 bg-slate-50/50 p-2.5 rounded-xl border border-slate-100/50"
            >
              <span class="font-bold text-[10px] uppercase text-slate-450 block mb-0.5">Alasan</span>
              <p class="leading-relaxed">{{ leave.reason }}</p>
            </div>

            <!-- Rejection details if rejected -->
            <div
              v-if="leave.status === 'REJECTED' && leave.reject_reason"
              class="text-xs text-red-700 bg-red-50/50 p-2.5 rounded-xl border border-red-100"
            >
              <span class="font-bold text-[10px] uppercase text-red-500 block mb-0.5"
                >Alasan Penolakan</span
              >
              <p class="leading-relaxed">{{ leave.reject_reason }}</p>
            </div>

            <div
              class="flex items-center justify-between pt-2 border-t border-slate-100/60 text-[10px] text-slate-400"
            >
              <span>Diajukan: {{ formatDate(leave.created_at) }}</span>

              <!-- Attachment link -->
              <a
                v-if="leave.attachment_url"
                :href="leave.attachment_url"
                target="_blank"
                class="font-bold text-primary hover:underline flex items-center gap-1.5"
              >
                <Paperclip class="w-3.5 h-3.5" />
                Lihat Lampiran
              </a>
            </div>
          </div>

          <div v-if="leaveHistory.length === 0" class="text-xs text-slate-400 text-center py-6">
            Belum ada riwayat pengajuan cuti atau surat izin.
          </div>
        </div>
      </div>
    </div>

    <!-- Content for TAB: Profil & Akun (Profile) -->
    <div v-else-if="currentTab === 'profile'" class="space-y-6">
      <!-- Pembaruan Data Wajah Card -->
      <div class="bg-white rounded-3xl border border-slate-200/60 p-5 sm:p-6 shadow-sm space-y-6">
        <div class="flex items-center gap-3">
          <div class="w-10 h-10 rounded-xl bg-primary/10 text-primary flex items-center justify-center">
            <ShieldCheck class="w-5 h-5" />
          </div>
          <div>
            <h3 class="text-sm font-black text-slate-800 uppercase tracking-wider">Registrasi Wajah Biometrik</h3>
            <p class="text-[11px] text-slate-400 font-bold uppercase mt-0.5">
              Status: 
              <span v-if="employeeProfile?.is_face_registered" class="text-green-600 font-black">TERDAFTAR</span>
              <span v-else class="text-amber-500 font-black">BELUM TERDAFTAR</span>
            </p>
          </div>
        </div>

        <p class="text-xs text-slate-500 leading-relaxed">
          Gunakan kamera depan Anda untuk mengambil foto selfie terbaru sebagai data biometrik wajah. Pembaruan data wajah di sini akan digunakan saat memverifikasi absensi di HP/iPhone Anda.
        </p>

        <!-- Face Capture Section -->
        <div class="flex flex-col items-center gap-4 border-t border-slate-100 pt-4">
          <!-- Input file picker for Face Capture (Safari-safe invisible styling instead of display:none) -->
          <input
            id="face-photo-file-input"
            type="file"
            accept="image/*"
            capture="user"
            @change="handleFacePhotoCaptured"
            ref="facePhotoInput"
            style="opacity: 0; position: absolute; width: 1px; height: 1px; overflow: hidden; pointer-events: none;"
          />

          <!-- Preview if captured -->
          <div v-if="capturedFacePhoto" class="relative w-40 h-40 rounded-full border-4 border-primary overflow-hidden shadow-md group">
            <img :src="capturedFacePhoto" class="w-full h-full object-cover" />
            <div class="absolute inset-0 bg-black/40 opacity-0 group-hover:opacity-100 transition-opacity duration-200 flex items-center justify-center">
              <button
                type="button"
                @click="removeFacePhoto"
                class="p-2.5 rounded-full bg-red-600 hover:bg-red-700 text-white shadow"
              >
                <Trash2 class="w-4 h-4" />
              </button>
            </div>
          </div>

          <!-- Label for native Safari file trigger -->
          <label
            v-if="!capturedFacePhoto"
            for="face-photo-file-input"
            class="flex items-center gap-2 px-5 py-3 rounded-2xl border-2 border-dashed border-slate-300 text-slate-600 hover:border-primary hover:text-primary transition-all font-bold text-xs cursor-pointer"
          >
            <Camera class="w-4.5 h-4.5" />
            Ambil Foto Wajah Baru
          </label>

          <button
            v-else
            @click="handleFaceRegisterSubmit"
            :disabled="isSubmittingFace"
            type="button"
            class="w-full h-11 bg-primary hover:bg-primary/95 text-white font-black rounded-xl text-xs shadow-md shadow-primary/20 flex items-center justify-center gap-2 active:scale-98 transition-all"
          >
            <Send class="w-4 h-4" />
            <span>{{ isSubmittingFace ? 'MEMPROSES...' : 'SIMPAN DATA WAJAH' }}</span>
          </button>
        </div>
      </div>

      <!-- Profile Details Card -->
      <div class="bg-white rounded-3xl border border-slate-200/60 p-5 sm:p-6 shadow-sm space-y-4">
        <h3 class="text-sm font-black text-slate-800 uppercase tracking-widest flex items-center gap-2">
          <User class="w-4 h-4 text-slate-500" />
          Informasi Detail Karyawan
        </h3>
        <div class="space-y-3 pt-2 text-xs">
          <div class="flex justify-between py-2 border-b border-slate-100">
            <span class="text-slate-450 font-bold uppercase tracking-wider">Nama Lengkap</span>
            <span class="font-extrabold text-slate-800">{{ employeeProfile ? `${employeeProfile.first_name} ${employeeProfile.last_name || ''}` : '-' }}</span>
          </div>
          <div class="flex justify-between py-2 border-b border-slate-100">
            <span class="text-slate-450 font-bold uppercase tracking-wider">Nomor Induk Karyawan (NIK)</span>
            <span class="font-extrabold text-slate-800">{{ employeeProfile?.employee_id_number || '-' }}</span>
          </div>
          <div class="flex justify-between py-2 border-b border-slate-100">
            <span class="text-slate-450 font-bold uppercase tracking-wider">Kantor Cabang</span>
            <span class="font-extrabold text-slate-800">{{ employeeProfile?.branch?.name || '-' }}</span>
          </div>
          <div class="flex justify-between py-2 border-b border-slate-100">
            <span class="text-slate-450 font-bold uppercase tracking-wider">Departemen</span>
            <span class="font-extrabold text-slate-800">{{ employeeProfile?.department?.name || '-' }}</span>
          </div>
          <div class="flex justify-between py-2 border-b border-slate-100">
            <span class="text-slate-450 font-bold uppercase tracking-wider">Jabatan</span>
            <span class="font-extrabold text-slate-800">{{ employeeProfile?.job_position?.title || '-' }}</span>
          </div>
          <div class="flex justify-between py-2 border-b border-slate-100">
            <span class="text-slate-450 font-bold uppercase tracking-wider">Jadwal Shift Kerja</span>
            <span class="font-extrabold text-slate-800">{{ getShiftText }}</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</Transition>

  <!-- Modal Dialog: Alasan Keterlambatan Absensi Masuk (Non-dismissible) -->
  <Transition
    enter-active-class="transition duration-200 ease-out"
    enter-from-class="opacity-0"
    enter-to-class="opacity-100"
    leave-active-class="transition duration-150 ease-in"
    leave-from-class="opacity-100"
    leave-to-class="opacity-0"
  >
    <div
      v-if="isLateReasonModalOpen"
      class="fixed inset-0 z-50 bg-slate-900/60 backdrop-blur-sm flex items-end sm:items-center justify-center p-0 sm:p-4"
    >
      <div
        class="bg-white w-full sm:max-w-md rounded-t-3xl sm:rounded-3xl p-5 sm:p-6 shadow-2xl border border-slate-100 space-y-4 animate-in fade-in zoom-in-95 duration-200 max-h-[90vh] overflow-y-auto"
      >
        <!-- Modal Header -->
        <div class="flex items-start gap-3">
          <div
            class="w-11 h-11 rounded-2xl bg-amber-100 border border-amber-200 flex items-center justify-center text-amber-700 shrink-0"
          >
            <AlertTriangle class="w-6 h-6" />
          </div>
          <div>
            <h3 class="text-base sm:text-lg font-black text-slate-900 tracking-tight">
              Keterlambatan Presensi
            </h3>
            <p class="text-xs text-slate-500 font-medium mt-0.5">
              Waktu absen masuk melewati jadwal shift kerja. Mohon berikan alasan keterlambatan Anda.
            </p>
          </div>
        </div>

        <!-- Input Reason Textarea -->
        <div class="space-y-1.5 pt-2">
          <label class="text-xs font-bold text-slate-700 uppercase tracking-wider block">
            Ketik Alasan Keterlambatan <span class="text-red-500">*</span>
          </label>
          <textarea
            v-model="lateReasonInput"
            rows="3"
            :disabled="isSubmittingLateReason"
            placeholder="Tuliskan secara jelas alasan Anda terlambat absen masuk..."
            class="w-full rounded-2xl border border-slate-200 bg-slate-50/50 p-3.5 text-xs font-semibold text-slate-900 placeholder:text-slate-400 focus:bg-white focus:outline-none focus:ring-2 focus:ring-amber-500/30 focus:border-amber-500 transition-all leading-relaxed"
          ></textarea>
          <span class="text-[10px] text-slate-400 font-medium block">
            *Catatan ini akan tercatat dalam rekapan kehadiran Anda dan dapat ditinjau oleh HR/Manager.
          </span>
        </div>

        <!-- Action Button -->
        <div class="pt-2">
          <button
            @click="handleLateReasonSubmit"
            type="button"
            :disabled="isSubmittingLateReason || !lateReasonInput.trim()"
            class="w-full h-12 rounded-xl bg-amber-500 hover:bg-amber-600 disabled:bg-slate-200 disabled:text-slate-400 text-white font-black text-xs uppercase tracking-wider flex items-center justify-center gap-2 shadow-lg shadow-amber-500/20 active:scale-98 transition-all"
          >
            <div
              v-if="isSubmittingLateReason"
              class="w-4 h-4 rounded-full border-2 border-white/30 border-t-white animate-spin"
            ></div>
            <Send v-else class="w-4 h-4" />
            <span>{{ isSubmittingLateReason ? 'MENYIMPAN ALASAN...' : 'KIRIM CATATAN ALASAN' }}</span>
          </button>
        </div>
      </div>
    </div>
  </Transition>
</div>
</template>

<style scoped>
textarea {
  resize: none;
}

/* Tab Fade slide transition */
.fade-slide-enter-active,
.fade-slide-leave-active {
  transition: all 0.2s ease-out;
}
.fade-slide-enter-from {
  opacity: 0;
  transform: translateY(8px);
}
.fade-slide-leave-to {
  opacity: 0;
  transform: translateY(-8px);
}

/* Leaflet customized overlay styles */
:deep(.leaflet-control-zoom) {
  border: none !important;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08) !important;
}

:deep(.leaflet-bar a) {
  background-color: white !important;
  color: #334155 !important;
  border-bottom: 1px solid #f1f5f9 !important;
  width: 32px !important;
  height: 32px !important;
  line-height: 32px !important;
  transition: all 0.2s;
}

:deep(.leaflet-bar a:hover) {
  background-color: #f8fafc !important;
  color: #020617 !important;
}

:deep(.leaflet-popup-content-wrapper) {
  border-radius: 12px !important;
  box-shadow: 0 10px 15px -3px rgba(0, 0, 0, 0.05) !important;
  border: 1px solid #e2e8f0 !important;
  padding: 2px !important;
}

:deep(.leaflet-popup-content) {
  font-family: inherit !important;
  font-size: 11px !important;
  color: #334155 !important;
  margin: 8px 12px !important;
  line-height: 1.4 !important;
}

:deep(.leaflet-popup-tip) {
  box-shadow: 0 10px 15px -3px rgba(0, 0, 0, 0.05) !important;
}
</style>
