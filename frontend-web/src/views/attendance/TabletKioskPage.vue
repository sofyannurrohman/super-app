<template>
  <div class="h-screen h-dvh max-h-screen overflow-hidden bg-gradient-to-br from-slate-50 via-white to-sky-50 flex flex-col justify-between p-3 sm:p-5 md:p-6 font-sans text-gray-800 select-none">
    <!-- Header -->
    <header class="flex items-center justify-between gap-2 sm:gap-4 shrink-0 pb-2 sm:pb-3 border-b border-slate-200/60">
      <div class="flex items-center gap-2 sm:gap-3">
        <div class="w-9 h-9 sm:w-11 sm:h-11 rounded-xl sm:rounded-2xl bg-primary/10 flex items-center justify-center text-primary font-bold shadow-sm">
          <i class="bi bi-person-badge text-lg sm:text-xl"></i>
        </div>
        <div>
          <h1 class="m-0 text-gray-900 text-lg sm:text-xl md:text-2xl font-bold tracking-tight leading-tight">
            Kehadiran Pekerja Harian
          </h1>
          <p class="m-0 text-[11px] sm:text-xs font-medium text-slate-500 hidden sm:block">
            {{ currentDate }} &bull; {{ currentTime }}
          </p>
        </div>
      </div>

      <div class="flex items-center gap-2">
        <div class="flex items-center gap-2 bg-white px-3 sm:px-4 py-1.5 rounded-full shadow-sm border border-slate-200/80">
          <label class="text-xs sm:text-sm font-medium text-gray-500 m-0 whitespace-nowrap hidden md:inline">Cabang:</label>
          <select 
            v-model="selectedBranch" 
            @change="saveBranch" 
            class="form-select text-xs sm:text-sm font-semibold text-gray-900 border-none bg-transparent outline-none cursor-pointer pr-4 focus:ring-0"
          >
            <option value="" disabled>Pilih Lokasi Cabang</option>
            <option v-for="b in branches" :key="b.id" :value="b.id">
              {{ b.name }}
            </option>
          </select>
        </div>

        <button 
          type="button" 
          @click="toggleFullscreen" 
          :title="isFullscreen ? 'Keluar Layar Penuh' : 'Layar Penuh (Kiosk)'"
          class="w-8 h-8 sm:w-9 sm:h-9 rounded-full bg-white shadow-sm border border-slate-200/80 flex items-center justify-center text-slate-600 hover:text-primary hover:bg-slate-50 active:scale-95 transition-all cursor-pointer"
        >
          <svg v-if="!isFullscreen" xmlns="http://www.w3.org/2000/svg" class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M4 8V4m0 0h4M4 4l5 5m11-5h-4m4 0v4m0-4l-5 5M4 16v4m0 0h4m-4 0l5-5m11 5l-5-5m5 5v-4m0 4h-4" />
          </svg>
          <svg v-else xmlns="http://www.w3.org/2000/svg" class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M9 9L4 4m0 0v4m0-4h4m6 5l5-5m0 0v4m0-4h-4m-7 11l-5 5m0 0v-4m0 4h4m11-5l5 5m0 0v-4m0 4h-4" />
          </svg>
        </button>
      </div>
    </header>

    <!-- Main Content Area -->
    <main class="flex-1 min-h-0 flex items-center justify-center py-2 sm:py-4">
      <div 
        v-if="selectedBranch" 
        class="w-full h-full max-w-5xl mx-auto flex flex-col sm:flex-row items-center justify-center gap-3 sm:gap-6 md:gap-10 lg:gap-14"
      >
        <!-- Camera Column with Face & Mask Detection Overlay -->
        <div class="flex-1 w-full max-w-[280px] sm:max-w-[320px] md:max-w-[360px] flex flex-col items-center justify-center">
          <div 
            class="w-full aspect-[4/3] sm:aspect-[3/4] max-h-[28vh] sm:max-h-[50vh] md:max-h-[55vh] bg-slate-900 rounded-2xl sm:rounded-3xl overflow-hidden shadow-xl sm:shadow-2xl relative border-4 shrink-0 transition-all duration-300"
            :class="[
              faceStatusCode === 'VALID' ? 'border-emerald-500 shadow-emerald-500/30' :
              faceStatusCode === 'MASK_DETECTED' ? 'border-rose-500 shadow-rose-500/40 animate-pulse' :
              faceStatusCode === 'MULTIPLE_FACES' ? 'border-amber-500 shadow-amber-500/30' :
              'border-white'
            ]"
          >
            <video ref="video" autoplay playsinline muted class="w-full h-full object-cover -scale-x-100"></video>
            <canvas ref="canvas" style="display: none;"></canvas>
            
            <!-- Face Guide Oval Box -->
            <div class="absolute inset-0 flex items-center justify-center pointer-events-none p-4">
              <div 
                class="w-36 h-48 sm:w-44 sm:h-56 rounded-[50%] border-2 border-dashed transition-all duration-300 flex items-center justify-center"
                :class="[
                  faceStatusCode === 'VALID' ? 'border-emerald-400/90 shadow-[0_0_20px_rgba(52,211,153,0.3)]' :
                  faceStatusCode === 'MASK_DETECTED' ? 'border-rose-500 shadow-[0_0_20px_rgba(244,63,94,0.4)]' :
                  faceStatusCode === 'MULTIPLE_FACES' ? 'border-amber-400/90 shadow-[0_0_15px_rgba(251,191,36,0.3)]' :
                  'border-white/50'
                ]"
              >
                <!-- Corner markers for face alignment -->
                <div class="w-full h-full relative opacity-60">
                  <div class="absolute top-2 left-4 w-3 h-3 border-t-2 border-l-2 border-white"></div>
                  <div class="absolute top-2 right-4 w-3 h-3 border-t-2 border-r-2 border-white"></div>
                  <div class="absolute bottom-2 left-4 w-3 h-3 border-b-2 border-l-2 border-white"></div>
                  <div class="absolute bottom-2 right-4 w-3 h-3 border-b-2 border-r-2 border-white"></div>
                </div>
              </div>
            </div>

            <!-- Top AI Face Status Badge -->
            <div class="absolute top-2 inset-x-0 flex justify-center px-2 pointer-events-none">
              <div 
                class="px-3 py-1 rounded-full text-[10px] sm:text-xs font-bold tracking-wide flex items-center gap-1.5 shadow-md backdrop-blur-md transition-all duration-200"
                :class="[
                  faceStatusCode === 'VALID' ? 'bg-emerald-600/90 text-white shadow-emerald-500/20' :
                  faceStatusCode === 'MASK_DETECTED' ? 'bg-rose-600/95 text-white shadow-rose-500/30' :
                  faceStatusCode === 'MULTIPLE_FACES' ? 'bg-amber-500/90 text-white shadow-amber-500/20' :
                  faceStatusCode === 'LOADING' ? 'bg-sky-600/90 text-white' :
                  'bg-slate-900/80 text-white/90'
                ]"
              >
                <i v-if="faceStatusCode === 'VALID'" class="bi bi-check-circle-fill text-white"></i>
                <i v-else-if="faceStatusCode === 'MASK_DETECTED'" class="bi bi-exclamation-octagon-fill text-white animate-bounce"></i>
                <i v-else-if="faceStatusCode === 'MULTIPLE_FACES'" class="bi bi-people-fill text-white"></i>
                <i v-else-if="faceStatusCode === 'LOADING'" class="bi bi-arrow-repeat animate-spin text-white"></i>
                <i v-else class="bi bi-person-bounding-box text-white/80"></i>
                
                <span>{{ faceStatusMessage }}</span>
              </div>
            </div>

            <!-- Worker Name Badge Overlay -->
            <transition name="fade">
              <div 
                v-if="workerName" 
                class="absolute bottom-0 inset-x-0 bg-gradient-to-t from-black/90 via-black/60 to-transparent pt-8 pb-3 px-4 text-center"
              >
                <span class="inline-block bg-emerald-500/90 text-white text-[10px] sm:text-xs font-bold px-2.5 py-0.5 rounded-full mb-1 tracking-wider uppercase">
                  Pekerja Ditemukan
                </span>
                <h2 class="m-0 text-base sm:text-xl md:text-2xl font-bold text-white drop-shadow-md truncate">
                  {{ workerName }}
                </h2>
              </div>
            </transition>
          </div>

          <!-- Helper caption with live validation hint -->
          <p 
            class="mt-2 sm:mt-3 text-[11px] sm:text-xs md:text-sm font-medium text-center py-1.5 px-3 sm:px-4 rounded-full shadow-sm border max-w-xs transition-colors"
            :class="[
              faceStatusCode === 'VALID' ? 'bg-emerald-50 text-emerald-700 border-emerald-200' :
              faceStatusCode === 'MASK_DETECTED' ? 'bg-rose-50 text-rose-700 border-rose-200 font-bold' :
              'bg-white/80 text-slate-500 border-slate-200/60'
            ]"
          >
            <span v-if="faceStatusCode === 'MASK_DETECTED'">
              <i class="bi bi-x-circle-fill text-rose-500 mr-1"></i> Buka masker sebelum melakukan absensi.
            </span>
            <span v-else-if="faceStatusCode === 'VALID'">
              <i class="bi bi-shield-check text-emerald-500 mr-1"></i> Wajah terverifikasi, silakan tekan tombol absen.
            </span>
            <span v-else>
              <i class="bi bi-camera-fill text-primary mr-1"></i> Pastikan wajah terlihat jelas tanpa penutup.
            </span>
          </p>
        </div>

        <!-- Keypad & Input Column -->
        <div class="flex-1 w-full max-w-[300px] sm:max-w-[340px] md:max-w-[380px] flex flex-col justify-center">
          <!-- PIN Indicator -->
          <div class="text-center mb-2 sm:mb-3">
            <label class="text-xs sm:text-sm md:text-base font-bold text-slate-700 block mb-1.5">
              Masukkan PIN (3 Digit)
            </label>
            
            <div class="flex justify-center gap-3 sm:gap-4 my-1">
              <div 
                v-for="i in 3" 
                :key="i"
                class="w-11 h-13 sm:w-13 sm:h-15 md:w-14 md:h-16 rounded-xl sm:rounded-2xl border-2 flex items-center justify-center text-2xl sm:text-3xl font-bold transition-all duration-200 shadow-sm"
                :class="pin.length >= i ? 'border-primary bg-primary/10 text-primary scale-105 shadow-primary/10' : 'border-slate-300 bg-white text-slate-400'"
              >
                {{ pin[i - 1] ? pin[i - 1] : '' }}
              </div>
            </div>
          </div>
          
          <!-- Keypad Numpad -->
          <div class="grid grid-cols-3 gap-2 sm:gap-2.5 md:gap-3 mb-2.5 sm:mb-4 place-items-center">
            <button 
              v-for="n in ['1','2','3','4','5','6','7','8','9']" 
              :key="n"
              type="button"
              class="w-13 h-13 sm:w-15 sm:h-15 md:w-16 md:h-16 text-2xl sm:text-3xl font-semibold rounded-xl sm:rounded-2xl border-none bg-white shadow hover:bg-slate-50 active:scale-95 active:bg-slate-100 active:shadow-inner text-slate-800 transition-all duration-150 flex items-center justify-center cursor-pointer"
              @click="appendPin(n)"
            >
              {{ n }}
            </button>
            
            <!-- Clear Button -->
            <button 
              type="button"
              class="w-13 h-13 sm:w-15 sm:h-15 md:w-16 md:h-16 text-xl sm:text-2xl font-bold rounded-xl sm:rounded-2xl border-none bg-rose-50 shadow hover:bg-rose-100 active:scale-95 active:shadow-inner text-rose-600 transition-all duration-150 flex items-center justify-center cursor-pointer"
              @click="clearPin()"
              title="Hapus Semua"
            >
              C
            </button>
            
            <!-- Digit 0 -->
            <button 
              type="button"
              class="w-13 h-13 sm:w-15 sm:h-15 md:w-16 md:h-16 text-2xl sm:text-3xl font-semibold rounded-xl sm:rounded-2xl border-none bg-white shadow hover:bg-slate-50 active:scale-95 active:bg-slate-100 active:shadow-inner text-slate-800 transition-all duration-150 flex items-center justify-center cursor-pointer"
              @click="appendPin('0')"
            >
              0
            </button>
            
            <!-- Backspace Button -->
            <button 
              type="button"
              class="w-13 h-13 sm:w-15 sm:h-15 md:w-16 md:h-16 rounded-xl sm:rounded-2xl border-none bg-rose-50 shadow hover:bg-rose-100 active:scale-95 active:shadow-inner text-rose-600 transition-all duration-150 flex items-center justify-center cursor-pointer"
              @click="deletePin()"
              title="Hapus 1 Digit"
            >
              <svg xmlns="http://www.w3.org/2000/svg" class="w-6 h-6 sm:w-7 sm:h-7" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <path d="M21 4H8l-7 8 7 8h13a2 2 0 0 0 2-2V6a2 2 0 0 0-2-2z"></path>
                <line x1="18" y1="9" x2="12" y2="15"></line>
                <line x1="12" y1="9" x2="18" y2="15"></line>
              </svg>
            </button>
          </div>

          <!-- Action Buttons (Masuk / Pulang) - Guarded with isFaceValid -->
          <div class="flex flex-row gap-2.5 sm:gap-3 w-full">
            <button 
              type="button"
              class="flex-1 text-sm sm:text-base font-bold py-2.5 sm:py-3.5 rounded-xl sm:rounded-2xl transition-all duration-200 border-none disabled:opacity-40 disabled:cursor-not-allowed disabled:transform-none disabled:shadow-none flex items-center justify-center gap-1.5 sm:gap-2 bg-emerald-500 text-white shadow-lg shadow-emerald-500/30 active:bg-emerald-600 active:translate-y-0.5 cursor-pointer" 
              :disabled="!isValid || isSubmitting || hasCheckedIn"
              @click="submitAttendance('check_in')"
            >
              <i class="bi bi-box-arrow-in-right text-base sm:text-lg"></i> MASUK
            </button>
            
            <button 
              type="button"
              class="flex-1 text-sm sm:text-base font-bold py-2.5 sm:py-3.5 rounded-xl sm:rounded-2xl transition-all duration-200 disabled:opacity-40 disabled:cursor-not-allowed disabled:transform-none disabled:shadow-none flex items-center justify-center gap-1.5 sm:gap-2 bg-white text-rose-600 border-2 border-rose-500 border-solid active:bg-rose-50 active:translate-y-0.5 cursor-pointer shadow-sm" 
              :disabled="!isValid || isSubmitting || !hasCheckedIn || hasClockedOut"
              @click="submitAttendance('clock_out')"
            >
              <i class="bi bi-box-arrow-right text-base sm:text-lg"></i> PULANG
            </button>
          </div>
        </div>
      </div>
      
      <!-- Placeholder when branch is not selected -->
      <div v-else class="flex flex-col text-center items-center justify-center p-6 max-w-lg mx-auto">
        <div class="bg-white p-8 sm:p-12 rounded-3xl shadow-xl border border-slate-100 flex flex-col items-center">
          <div class="w-20 h-20 rounded-full bg-primary/10 text-primary flex items-center justify-center mb-4">
            <i class="bi bi-shop text-4xl"></i>
          </div>
          <h3 class="text-xl sm:text-2xl font-bold text-gray-800 mb-2">Selamat Datang di Kiosk Absensi</h3>
          <p class="text-gray-500 text-sm sm:text-base m-0">Silakan pilih lokasi cabang di pojok kanan atas untuk memulai absensi.</p>
        </div>
      </div>
    </main>
    
    <!-- Footer info -->
    <footer class="text-center shrink-0 pt-1 border-t border-slate-200/40">
      <p class="m-0 text-[10px] sm:text-xs text-slate-400 font-medium">
        Kiosk Absensi Presensi Pekerja &bull; HRIS Karsa HR &bull; AI Face Detection Active
      </p>
    </footer>

    <!-- Notification Overlay -->
    <div class="fixed inset-0 bg-slate-900/40 backdrop-blur-sm flex justify-center items-center z-[9999] p-4" v-if="notification">
      <div class="text-center p-8 sm:p-12 rounded-3xl bg-white shadow-2xl animate-[popIn_0.35s_ease-out] max-w-[400px] w-full popup-card border border-slate-100">
        <div class="w-20 h-20 rounded-full flex items-center justify-center mx-auto mb-5 shadow-inner" :class="notification.type === 'success' ? 'bg-emerald-100 text-emerald-600' : 'bg-rose-100 text-rose-600'">
          <svg v-if="notification.type === 'success'" xmlns="http://www.w3.org/2000/svg" class="w-10 h-10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round">
            <polyline points="20 6 9 17 4 12"></polyline>
          </svg>
          <svg v-else xmlns="http://www.w3.org/2000/svg" class="w-10 h-10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round">
            <line x1="18" y1="6" x2="6" y2="18"></line>
            <line x1="6" y1="6" x2="18" y2="18"></line>
          </svg>
        </div>
        <h2 class="text-2xl font-bold mb-2 text-gray-900">{{ notification.title }}</h2>
        <p class="text-sm sm:text-base text-gray-600 m-0 leading-relaxed">{{ notification.message }}</p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import apiClient from '@/api/axios'
import { useFaceDetector } from '@/composables/useFaceDetector'
import { burnAttendanceWatermark, formatAttendanceDateTime } from '@/utils/watermark'

// State
const branches = ref<any[]>([])
const selectedBranch = ref(localStorage.getItem('kioskBranchId') || '')
const pin = ref('')
const workerName = ref('')
const hasCheckedIn = ref(false)
const hasClockedOut = ref(false)
const isCheckingPin = ref(false)
const isSubmitting = ref(false)
const notification = ref<{type: 'success'|'error', title: string, message: string} | null>(null)
const isFullscreen = ref(false)
const currentTime = ref('')
const currentDate = ref('')
let clockInterval: any = null

// Face & Mask Detector
const {
  isModelReady,
  isMaskDetected,
  isFaceValid,
  statusCode: faceStatusCode,
  statusMessage: faceStatusMessage,
  initDetector,
  startDetection,
  stopDetection
} = useFaceDetector()

// Camera refs
const video = ref<HTMLVideoElement | null>(null)
const canvas = ref<HTMLCanvasElement | null>(null)
let mediaStream: MediaStream | null = null

// Computed: Valid only if PIN, Worker, Branch, and Face are VALID
const isValid = computed(() => {
  return (
    pin.value.length === 3 &&
    workerName.value &&
    selectedBranch.value &&
    isFaceValid.value &&
    !isSubmitting.value
  )
})

// Clock updater
const updateClock = () => {
  const now = new Date()
  currentTime.value = now.toLocaleTimeString('id-ID', { hour: '2-digit', minute: '2-digit', second: '2-digit' }).replace(/\./g, ':')
  currentDate.value = now.toLocaleDateString('id-ID', { weekday: 'long', day: 'numeric', month: 'long', year: 'numeric' })
}

// Fullscreen toggle
const toggleFullscreen = () => {
  if (!document.fullscreenElement) {
    document.documentElement.requestFullscreen().catch((err) => {
      console.warn('Error attempting to enable fullscreen:', err)
    })
  } else {
    if (document.exitFullscreen) {
      document.exitFullscreen().catch((err) => {
        console.warn('Error attempting to exit fullscreen:', err)
      })
    }
  }
}

const handleFullscreenChange = () => {
  isFullscreen.value = !!document.fullscreenElement
}

// Methods
const saveBranch = () => {
  localStorage.setItem('kioskBranchId', selectedBranch.value)
}

const appendPin = (digit: string) => {
  if (pin.value.length < 3) {
    pin.value += digit
    if (pin.value.length === 3) {
      checkPin()
    }
  }
}

const clearPin = () => {
  pin.value = ''
  workerName.value = ''
  hasCheckedIn.value = false
  hasClockedOut.value = false
}

const deletePin = () => {
  if (pin.value.length > 0) {
    pin.value = pin.value.slice(0, -1)
    workerName.value = ''
    hasCheckedIn.value = false
    hasClockedOut.value = false
  }
}

const checkPin = async () => {
  if (!selectedBranch.value) {
    showNotification('error', 'Error', 'Pilih cabang terlebih dahulu.')
    clearPin()
    return
  }
  isCheckingPin.value = true
  try {
    const res = await apiClient.get(`/attendance/tablet/check-pin?pin=${pin.value}&branch_id=${selectedBranch.value}`)
    const data = res.data.data || res.data
    hasCheckedIn.value = !!data.has_checked_in
    hasClockedOut.value = !!data.has_clocked_out
    if (data.full_name) {
      workerName.value = data.full_name
    } else if (data.first_name) {
      workerName.value = `${data.first_name} ${data.last_name || ''}`.trim()
    } else {
      workerName.value = 'Nama Tidak Diketahui'
    }
  } catch (error: any) {
    const msg = error.response?.data?.message || 'PIN tidak ditemukan.'
    showNotification('error', 'Gagal', msg)
    setTimeout(() => {
      clearPin()
      notification.value = null
    }, 2000)
  } finally {
    isCheckingPin.value = false
  }
}

const initCamera = async () => {
  try {
    if (!navigator.mediaDevices || !navigator.mediaDevices.getUserMedia) {
      throw new Error("Browser tidak mendukung atau koneksi tidak aman.")
    }

    try {
      // Coba akses kamera depan terlebih dahulu
      mediaStream = await navigator.mediaDevices.getUserMedia({ 
        video: { facingMode: 'user', width: { ideal: 640 }, height: { ideal: 480 } } 
      })
    } catch (fallbackErr) {
      console.warn('Gagal mengakses kamera depan, mencoba kamera default...', fallbackErr)
      // Fallback: coba akses kamera apa saja yang tersedia
      mediaStream = await navigator.mediaDevices.getUserMedia({ 
        video: true 
      })
    }

    if (video.value) {
      video.value.srcObject = mediaStream
      video.value.onloadedmetadata = () => {
        if (video.value) {
          startDetection(video.value)
        }
      }
    }
  } catch (err: any) {
    console.error('Error accessing camera:', err)
    
    let errorMessage = 'Tidak dapat mengakses kamera perangkat. Pastikan tidak ada aplikasi lain yang menggunakan kamera.'
    
    if (err.name === 'NotAllowedError' || err.name === 'PermissionDeniedError') {
      errorMessage = 'Izin kamera ditolak. Periksa pengaturan izin situs di browser dan pengaturan izin Chrome di HP Anda.'
    } else if (err.name === 'NotFoundError' || err.name === 'DevicesNotFoundError') {
      errorMessage = 'Kamera tidak ditemukan atau rusak pada perangkat ini.'
    } else if (err.name === 'NotReadableError' || err.name === 'TrackStartError') {
      errorMessage = 'Kamera sedang digunakan oleh aplikasi lain. Mohon tutup aplikasi yang menggunakan kamera lalu muat ulang halaman.'
    } else if (err.name === 'OverconstrainedError' || err.name === 'ConstraintNotSatisfiedError') {
      errorMessage = 'Kamera perangkat tidak memenuhi spesifikasi yang dibutuhkan sistem.'
    } else if (err.message) {
      errorMessage = `Error: ${err.message}`
    }
    
    showNotification('error', 'Kamera Error', errorMessage)
  }
}

const stopCamera = () => {
  stopDetection()
  if (mediaStream) {
    mediaStream.getTracks().forEach(track => track.stop())
  }
}

const capturePhoto = async (): Promise<string> => {
  if (!video.value) return ''

  const branchObj = branches.value.find((b) => b.id === selectedBranch.value)
  const branchName = branchObj ? branchObj.name : 'Kiosk Tablet'

  try {
    return await burnAttendanceWatermark(video.value, {
      timeText: formatAttendanceDateTime(new Date()),
      branchName: branchName,
      addressText: `Presensi Kiosk Tablet - ${branchName}`,
      targetWidth: 720,
      cropSquare: true,
      quality: 0.8,
    })
  } catch (err) {
    console.error('Failed to burn watermark on kiosk photo:', err)
    if (!canvas.value) return ''
    const context = canvas.value.getContext('2d')
    if (!context) return ''
    canvas.value.width = video.value.videoWidth
    canvas.value.height = video.value.videoHeight
    context.drawImage(video.value, 0, 0, canvas.value.width, canvas.value.height)
    return canvas.value.toDataURL('image/jpeg', 0.7)
  }
}

const submitAttendance = async (type: 'check_in' | 'clock_out') => {
  if (!isValid.value) {
    if (!isFaceValid.value) {
      if (isMaskDetected.value) {
        showNotification('error', 'Masker Terdeteksi', 'Mohon lepaskan masker Anda untuk melakukan absensi.')
      } else {
        showNotification('error', 'Wajah Tidak Terdeteksi', 'Pastikan wajah Anda menghadap langsung ke kamera.')
      }
    }
    return
  }

  isSubmitting.value = true
  try {
    const photoBase64 = await capturePhoto()

    const payload = {
      branch_id: selectedBranch.value,
      pin: pin.value,
      attendance_type: type,
      photo_base64: photoBase64,
    }

    await apiClient.post('/attendance/tablet', payload)
    
    showNotification(
      'success', 
      'Berhasil!', 
      `Absen ${type === 'check_in' ? 'Masuk' : 'Pulang'} berhasil dicatat.`
    )
    
    setTimeout(() => {
      resetForm()
    }, 2000)
    
  } catch (error: any) {
    console.error(error)
    const msg = error.response?.data?.message || 'Terjadi kesalahan saat menyimpan absensi.'
    showNotification('error', 'Gagal', msg)
    
    setTimeout(() => {
      resetForm()
    }, 2000)
  }
}

const showNotification = (type: 'success' | 'error', title: string, message: string) => {
  notification.value = { type, title, message }
}

const resetForm = () => {
  clearPin()
  notification.value = null
  isSubmitting.value = false
}

const fetchData = async () => {
  try {
    const res = await apiClient.get('/branches')
    branches.value = res.data.data || []
  } catch (error) {
    console.error('Failed to fetch data:', error)
  }
}

onMounted(() => {
  fetchData()
  initDetector()
  initCamera()
  updateClock()
  clockInterval = setInterval(updateClock, 1000)
  document.addEventListener('fullscreenchange', handleFullscreenChange)
})

onUnmounted(() => {
  stopCamera()
  if (clockInterval) clearInterval(clockInterval)
  document.removeEventListener('fullscreenchange', handleFullscreenChange)
})
</script>

<style scoped>
@import url('https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700&display=swap');

.popup-card {
  animation: popIn 0.35s cubic-bezier(0.16, 1, 0.3, 1);
}

@keyframes popIn {
  0% { transform: scale(0.85) translateY(15px); opacity: 0; }
  100% { transform: scale(1) translateY(0); opacity: 1; }
}

.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.25s ease, transform 0.25s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
  transform: translateY(6px);
}
</style>
