<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import {
  User,
  Mail,
  Lock,
  Briefcase,
  MapPin,
  Clock,
  Camera,
  Trash2,
  ArrowLeft,
  Loader2,
  ChevronRight,
  ShieldCheck,
} from 'lucide-vue-next'
import { brand } from '@/config/brand'
import apiClient from '@/api/axios'
import { toast } from 'vue-sonner'
import { detectAndCropFace } from '@/utils/faceCrop'

const router = useRouter()

// Form States
const name = ref('')
const email = ref('')
const password = ref('')
const selectedJobPositionId = ref('')
const selectedBranchId = ref('')
const selectedShiftId = ref('')

// Dropdown Lists
const jobPositions = ref<any[]>([])
const branches = ref<any[]>([])
const shifts = ref<any[]>([])

// Loading states
const isLoadingData = ref(true)
const isSubmitting = ref(false)
const errorMsg = ref('')

// Camera / Selfie states
const capturedSelfie = ref<string | null>(null)
const selfieInput = ref<HTMLInputElement | null>(null)

// Fetch master data on load (public endpoints)
const loadMasterData = async () => {
  isLoadingData.value = true
  try {
    const [positionsRes, branchesRes, shiftsRes] = await Promise.all([
      apiClient.get('/job-positions'),
      apiClient.get('/branches'),
      apiClient.get('/shifts'),
    ])
    jobPositions.value = positionsRes.data.data || []
    branches.value = branchesRes.data.data || []
    shifts.value = shiftsRes.data.data || []
  } catch (err) {
    console.error('Failed to load register options:', err)
    toast.error('Gagal memuat data pilihan jabatan/kantor')
  } finally {
    isLoadingData.value = false
  }
}

// Camera controls
const triggerSelfie = () => {
  selfieInput.value?.click()
}

const handleSelfieCaptured = (e: any) => {
  const file = e.target.files[0]
  if (!file) return

  const reader = new FileReader()
  reader.onload = async (event: any) => {
    try {
      capturedSelfie.value = await detectAndCropFace(event.target.result)
    } catch (err) {
      console.error('Error processing selfie image:', err)
      capturedSelfie.value = event.target.result
    }
  }
  reader.readAsDataURL(file)
}

const removeSelfie = () => {
  capturedSelfie.value = null
  if (selfieInput.value) {
    selfieInput.value.value = ''
  }
}

// Handle Register Submit
const handleRegister = async () => {
  if (
    !name.value ||
    !email.value ||
    !password.value ||
    !selectedJobPositionId.value ||
    !selectedBranchId.value
  ) {
    errorMsg.value = 'Semua kolom wajib diisi kecuali Shift & Selfie'
    toast.error('Gagal', { description: 'Mohon lengkapi formulir pendaftaran.' })
    return
  }

  isSubmitting.value = true
  errorMsg.value = ''

  try {
    let selfieBase64 = ''
    if (capturedSelfie.value) {
      const commaIndex = capturedSelfie.value.indexOf(',')
      if (commaIndex !== -1) {
        selfieBase64 = capturedSelfie.value.substring(commaIndex + 1)
      } else {
        selfieBase64 = capturedSelfie.value
      }
    }

    const systemGeneratedNik = `REG-${Date.now()}`
    const cleanEmail = email.value.trim().toLowerCase()
    const cleanPassword = password.value.trim()

    await apiClient.post('/register', {
      name: name.value.trim(),
      email: cleanEmail,
      employee_id: systemGeneratedNik,
      password: cleanPassword,
      job_position_id: selectedJobPositionId.value,
      branch_id: selectedBranchId.value,
      shift_id: selectedShiftId.value || undefined,
      face_embedding: selfieBase64 ? Array.from({ length: 128 }, () => 0.0) : [],
      selfie: selfieBase64 || undefined,
    })

    toast.success('Pendaftaran Berhasil!', {
      description: 'Akun Anda berhasil didaftarkan. Silakan login.',
    })

    // Redirect to login with prefilled email
    router.push({ path: '/login', query: { registeredEmail: cleanEmail } })
  } catch (err: any) {
    console.error(err)
    errorMsg.value =
      err.response?.data?.message || err.message || 'Gagal melakukan pendaftaran akun.'
    toast.error('Pendaftaran Gagal', { description: errorMsg.value })
  } finally {
    isSubmitting.value = false
  }
}

onMounted(() => {
  loadMasterData()
})
</script>

<template>
  <div
    class="min-h-screen bg-[#FDFDFD] flex items-center justify-center p-4 sm:p-6 lg:p-8 relative overflow-y-auto font-jakarta"
  >
    <!-- Crimson background decorative blobs -->
    <div class="absolute inset-0 z-0 overflow-hidden pointer-events-none">
      <div
        class="absolute -top-[10%] -left-[10%] w-[50%] h-[50%] bg-primary/5 rounded-full blur-[120px] mix-blend-multiply"
      ></div>
      <div
        class="absolute -bottom-[10%] -right-[10%] w-[40%] h-[50%] bg-primary/3 rounded-full blur-[100px] mix-blend-multiply"
      ></div>
    </div>

    <!-- Main Container Card -->
    <div
      class="w-full max-w-5xl grid lg:grid-cols-12 bg-white/40 backdrop-blur-3xl rounded-3xl sm:rounded-[40px] border border-white/50 shadow-[0_32px_80px_-20px_rgba(0,0,0,0.1)] overflow-hidden relative z-10"
    >
      <!-- Left side: Information/Hero -->
      <div
        class="hidden lg:flex lg:col-span-5 flex-col justify-between p-10 bg-gradient-to-br from-primary via-[#0f766e] to-primary text-white relative overflow-hidden"
      >
        <!-- Abstract shape -->
        <div class="absolute inset-0 opacity-10 pointer-events-none">
          <svg viewBox="0 0 400 400" xmlns="http://www.w3.org/2000/svg" class="w-full h-full">
            <circle
              cx="200"
              cy="200"
              r="120"
              fill="none"
              stroke="white"
              stroke-width="1.5"
              stroke-dasharray="8 8"
            />
          </svg>
        </div>

        <div class="relative z-10 space-y-6">
          <div class="flex items-center gap-3">
            <div class="bg-white/10 backdrop-blur-md p-2.5 rounded-xl border border-white/20">
              <img :src="brand.logoUrl" alt="Logo" class="w-8 h-8" />
            </div>
            <div>
              <span class="text-lg font-black tracking-tight block">Portal Pendaftaran</span>
              <span class="text-[9px] font-bold text-white/70 uppercase tracking-[0.1em]"
                >Akun Karyawan Baru</span
              >
            </div>
          </div>

          <div class="pt-8">
            <h2 class="text-3xl font-black leading-tight">Daftarkan Akun<br />Secara Mandiri</h2>
            <p class="text-white/70 text-sm mt-3 leading-relaxed">
              Silakan lengkapi formulir registrasi karyawan. Akun Anda akan dibuat dan Anda bisa
              langsung melakukan absensi masuk melalui portal web.
            </p>
          </div>
        </div>

        <div
          class="relative z-10 bg-white/5 backdrop-blur-lg border border-white/10 p-5 rounded-2xl"
        >
          <div class="flex items-center gap-3">
            <ShieldCheck class="w-5 h-5 text-white shrink-0" />
            <div class="text-xs">
              <p class="font-bold">Akses Instan Absensi</p>
              <p class="text-white/60 mt-0.5">
                Setelah pendaftaran selesai, Anda dapat langsung melakukan clock-in/out.
              </p>
            </div>
          </div>
        </div>
      </div>

      <!-- Right side: Form -->
      <div class="lg:col-span-7 p-5 sm:p-8 md:p-10 bg-white/30 flex flex-col justify-center">
        <div class="w-full max-w-lg mx-auto">
          <!-- Back to Login button -->
          <button
            @click="router.push('/login')"
            class="flex items-center gap-1.5 text-xs font-bold text-slate-500 hover:text-slate-900 transition-colors uppercase tracking-widest mb-4 sm:mb-6"
          >
            <ArrowLeft class="w-4 h-4" />
            Kembali ke Login
          </button>

          <div class="mb-5 sm:mb-8">
            <h1 class="text-2xl font-black text-slate-900 leading-tight">Pendaftaran Karyawan</h1>
            <p class="text-slate-500 mt-1.5 text-xs font-medium">
              Lengkapi biodata kerja Anda secara akurat.
            </p>
          </div>

          <!-- Loader while loading job positions/branches -->
          <div
            v-if="isLoadingData"
            class="flex flex-col items-center justify-center py-12 space-y-3"
          >
            <Loader2 class="w-8 h-8 text-primary animate-spin" />
            <p class="text-xs text-slate-400 font-bold uppercase tracking-wider">
              Memuat Data Kantor & Jabatan...
            </p>
          </div>

          <form v-else @submit.prevent="handleRegister" class="space-y-3 sm:space-y-4">
            <div
              v-if="errorMsg"
              class="p-4 bg-primary/5 border border-primary/10 text-primary rounded-2xl text-xs font-bold"
            >
              {{ errorMsg }}
            </div>

            <!-- Full Name & Email -->
            <div class="grid grid-cols-1 md:grid-cols-2 gap-3 sm:gap-4">
              <div class="space-y-1.5">
                <label
                  class="block text-[10px] font-black text-slate-400 uppercase tracking-widest pl-1"
                  >Nama Lengkap</label
                >
                <div class="relative">
                  <User class="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-450" />
                  <input
                    v-model="name"
                    type="text"
                    placeholder="Contoh: Budi Pekerti"
                    class="w-full pl-11 pr-3 py-2.5 sm:py-3 rounded-xl border border-slate-200 bg-slate-50/50 text-xs font-bold outline-none focus:bg-white focus:border-primary transition-all placeholder:text-slate-350"
                    required
                  />
                </div>
              </div>
              <div class="space-y-1.5">
                <label
                  class="block text-[10px] font-black text-slate-400 uppercase tracking-widest pl-1"
                  >Email</label
                >
                <div class="relative">
                  <Mail class="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-450" />
                  <input
                    v-model="email"
                    type="email"
                    autocomplete="email"
                    autocapitalize="none"
                    autocorrect="off"
                    spellcheck="false"
                    placeholder="nama@email.com"
                    class="w-full pl-11 pr-3 py-2.5 sm:py-3 rounded-xl border border-slate-200 bg-slate-50/50 text-xs font-bold outline-none focus:bg-white focus:border-primary transition-all placeholder:text-slate-355"
                    required
                  />
                </div>
              </div>
            </div>

            <!-- Password Field -->
            <div class="space-y-1.5">
              <label
                class="block text-[10px] font-black text-slate-400 uppercase tracking-widest pl-1"
                >Kata Sandi (min 6 karakter)</label
              >
              <div class="relative">
                <Lock class="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-450" />
                <input
                  v-model="password"
                  type="password"
                  autocomplete="new-password"
                  autocapitalize="none"
                  autocorrect="off"
                  spellcheck="false"
                  placeholder="Buat sandi aman..."
                  class="w-full pl-11 pr-3 py-2.5 sm:py-3 rounded-xl border border-slate-200 bg-slate-50/50 text-xs font-bold outline-none focus:bg-white focus:border-primary transition-all placeholder:text-slate-355"
                  required
                />
              </div>
            </div>

            <!-- Job Position & Branch -->
            <div class="grid grid-cols-1 md:grid-cols-2 gap-3 sm:gap-4">
              <div class="space-y-1.5">
                <label
                  class="block text-[10px] font-black text-slate-400 uppercase tracking-widest pl-1"
                  >Posisi Pekerjaan</label
                >
                <div class="relative">
                  <Briefcase
                    class="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-450"
                  />
                  <select
                    v-model="selectedJobPositionId"
                    class="w-full pl-11 pr-3 py-2.5 sm:py-3 rounded-xl border border-slate-200 bg-slate-50/50 text-xs font-bold outline-none focus:bg-white focus:border-primary transition-all"
                    required
                  >
                    <option value="" disabled selected>Pilih Jabatan</option>
                    <option v-for="jp in jobPositions" :key="jp.id" :value="jp.id">
                      {{ jp.title }}
                    </option>
                  </select>
                </div>
              </div>

              <div class="space-y-1.5">
                <label
                  class="block text-[10px] font-black text-slate-400 uppercase tracking-widest pl-1"
                  >Cabang</label
                >
                <div class="relative">
                  <MapPin class="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-450" />
                  <select
                    v-model="selectedBranchId"
                    class="w-full pl-11 pr-3 py-2.5 sm:py-3 rounded-xl border border-slate-200 bg-slate-50/50 text-xs font-bold outline-none focus:bg-white focus:border-primary transition-all"
                    required
                  >
                    <option value="" disabled selected>Pilih Kantor Cabang</option>
                    <option v-for="b in branches" :key="b.id" :value="b.id">{{ b.name }}</option>
                  </select>
                </div>
              </div>
            </div>

            <!-- Shift (Optional) -->
            <div class="space-y-1.5">
              <label
                class="block text-[10px] font-black text-slate-400 uppercase tracking-widest pl-1"
                >Shift Kerja</label
              >
              <div class="relative">
                <Clock class="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-450" />
                <select
                  v-model="selectedShiftId"
                  class="w-full pl-11 pr-3 py-2.5 sm:py-3 rounded-xl border border-slate-200 bg-slate-50/50 text-xs font-bold outline-none focus:bg-white focus:border-primary transition-all"
                >
                  <option value="" selected>Pilih Shift Kerja</option>
                  <option v-for="s in shifts" :key="s.id" :value="s.id">
                    {{ s.name }} ({{ s.start_time.slice(0, 5) }} - {{ s.end_time.slice(0, 5) }})
                  </option>
                </select>
              </div>
            </div>

            <!-- Optional Registration Selfie Photo -->
            <div class="space-y-1.5 pt-2">
              <label
                class="block text-[10px] font-black text-slate-400 uppercase tracking-widest pl-1"
                >Ambil Selfie Wajah</label
              >
              <label
                for="register-selfie-input"
                class="group relative w-full h-24 sm:h-32 rounded-xl border-2 border-dashed border-slate-200 hover:border-primary hover:bg-slate-50/40 bg-slate-50/10 flex flex-col items-center justify-center cursor-pointer transition-all overflow-hidden"
              >
                <!-- Hidden Input (Safari-safe invisible styling instead of display:none) -->
                <input
                  id="register-selfie-input"
                  type="file"
                  accept="image/*"
                  capture="user"
                  @change="handleSelfieCaptured"
                  ref="selfieInput"
                  style="opacity: 0; position: absolute; width: 1px; height: 1px; overflow: hidden; pointer-events: none;"
                />

                <template v-if="capturedSelfie">
                  <img :src="capturedSelfie" class="w-full h-full object-cover" />
                  <div
                    class="absolute inset-0 bg-black/40 opacity-0 group-hover:opacity-100 flex items-center justify-center gap-2 transition-opacity"
                  >
                    <div
                      class="px-2.5 py-1.5 rounded-lg bg-white text-slate-800 text-[10px] font-bold flex items-center gap-1 shadow-sm cursor-pointer"
                    >
                      <Camera class="w-3.5 h-3.5" />Ulangi
                    </div>
                    <button
                      type="button"
                      @click.stop="removeSelfie"
                      class="px-2.5 py-1.5 rounded-lg bg-red-600 text-white text-[10px] font-bold flex items-center gap-1 shadow-sm"
                    >
                      <Trash2 class="w-3.5 h-3.5" />Hapus
                    </button>
                  </div>
                </template>
                <template v-else>
                  <div class="flex flex-col items-center text-center space-y-1">
                    <Camera
                      class="w-5 h-5 text-slate-400 group-hover:text-primary transition-colors"
                    />
                    <p class="text-[11px] font-bold text-slate-600 group-hover:text-primary">
                      Unggah / Ambil Foto Wajah
                    </p>
                    <p class="text-[9px] text-slate-400">Untuk pendaftaran identitas kehadiran</p>
                  </div>
                </template>
              </label>
            </div>

            <!-- Submit Button -->
            <button
              type="submit"
              :disabled="isSubmitting"
              class="group relative w-full py-3 sm:py-4 bg-slate-900 text-white font-black text-xs rounded-xl overflow-hidden shadow-lg transition-all hover:bg-primary active:scale-[0.98] disabled:opacity-75 disabled:pointer-events-none mt-2"
            >
              <div class="relative z-10 flex items-center justify-center gap-2">
                <Loader2 v-if="isSubmitting" class="w-4 h-4 animate-spin" />
                <span v-else class="flex items-center gap-1.5 tracking-widest text-xs">
                  Daftarkan Akun
                  <ChevronRight class="w-4 h-4 transition-transform group-hover:translate-x-0.5" />
                </span>
              </div>
              <div
                class="absolute inset-0 bg-primary translate-y-full group-hover:translate-y-0 transition-transform duration-350"
              ></div>
            </button>
          </form>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.pl-11 {
  padding-left: 2.75rem;
}
</style>
