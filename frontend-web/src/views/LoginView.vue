<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { Mail, Lock, AlertCircle, Loader2, ShieldCheck, ChevronRight, X, KeyRound } from 'lucide-vue-next'
import { brand } from '@/config/brand'
import { toast } from 'vue-sonner'
import apiClient from '@/api/axios'

const email = ref('')
const password = ref('')
const rememberMe = ref(false)
const errorMsg = ref('')
const isLoading = ref(false)
const showForgotAccessModal = ref(false)
const forgotEmail = ref('')
const forgotErrorMsg = ref('')
const isForgotLoading = ref(false)

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()

const openForgotAccessModal = () => {
  forgotEmail.value = email.value
  forgotErrorMsg.value = ''
  showForgotAccessModal.value = true
}

const closeForgotAccessModal = () => {
  if (isForgotLoading.value) return
  showForgotAccessModal.value = false
  forgotErrorMsg.value = ''
}

onMounted(() => {
  if (route.query.registeredEmail) {
    email.value = (route.query.registeredEmail as string).trim().toLowerCase()
    toast.success('Pendaftaran Berhasil', {
      description: 'Silakan masukkan kata sandi Anda untuk masuk.',
    })
  }
  if (route.query.session === 'expired') {
    errorMsg.value = 'Sesi Anda telah berakhir. Silakan masuk kembali.'
    toast.error('Sesi Berakhir', {
      description: 'Silakan login ulang untuk melanjutkan.',
    })
  }

  // Load remembered credentials (email only, never store passwords)
  const savedRemember = localStorage.getItem('remember_me') === 'true'
  if (savedRemember) {
    rememberMe.value = true
    email.value = localStorage.getItem('saved_email') || ''
  }
  // Clear any legacy plain password from previous versions
  localStorage.removeItem('saved_password')
})

const handleLogin = async () => {
  isLoading.value = true
  errorMsg.value = ''

  const cleanEmail = email.value.trim().toLowerCase()
  const cleanPassword = password.value

  if (!cleanEmail) {
    errorMsg.value = 'Silakan masukkan email Anda.'
    isLoading.value = false
    return
  }

  try {
    const success = await authStore.login({ email: cleanEmail, password: cleanPassword })
    if (success) {
      if (rememberMe.value) {
        localStorage.setItem('remember_me', 'true')
        localStorage.setItem('saved_email', cleanEmail)
      } else {
        localStorage.removeItem('remember_me')
        localStorage.removeItem('saved_email')
      }
      localStorage.removeItem('saved_password')
      router.push('/')
    } else {
      errorMsg.value = 'Email atau sandi yang Anda masukkan tidak sesuai.'
    }
  } catch (err) {
    errorMsg.value = 'Gangguan koneksi sistem. Silakan coba lagi.'
  } finally {
    isLoading.value = false
  }
}

const handleForgotAccess = async () => {
  const targetEmail = forgotEmail.value.trim().toLowerCase()
  forgotErrorMsg.value = ''

  if (!targetEmail) {
    forgotErrorMsg.value = 'Masukkan email yang terdaftar.'
    return
  }

  if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(targetEmail)) {
    forgotErrorMsg.value = 'Format email tidak valid.'
    return
  }

  isForgotLoading.value = true
  try {
    await apiClient.post('/forgot-password', { email: targetEmail })
    toast.success('Email Terkirim', {
      description: 'Password sementara telah dikirim ke email Anda. Periksa kotak masuk atau folder spam.',
    })
    showForgotAccessModal.value = false
    email.value = targetEmail
    forgotEmail.value = ''
  } catch (err: any) {
    forgotErrorMsg.value =
      err?.response?.data?.message ||
      err?.response?.data?.error ||
      'Gagal memproses permintaan. Silakan coba lagi.'
  } finally {
    isForgotLoading.value = false
  }
}
</script>

<template>
  <div
    class="min-h-screen bg-[#FDFDFD] flex items-center justify-center p-4 sm:p-6 lg:p-8 relative overflow-y-auto font-jakarta"
  >
    <!-- Premium Background: Abstract Crimson Flows -->
    <div class="absolute inset-0 z-0 overflow-hidden pointer-events-none">
      <div
        class="absolute -top-[10%] -left-[10%] w-[50%] h-[50%] bg-primary/5 rounded-full blur-[120px] mix-blend-multiply transition-transform duration-1000 transform hover:scale-110"
      ></div>
      <div
        class="absolute -bottom-[10%] -right-[10%] w-[40%] h-[50%] bg-primary/3 rounded-full blur-[100px] mix-blend-multiply"
      ></div>
      <div
        class="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-full h-full opacity-[0.02] grayscale select-none"
      >
        <!-- Noise pattern -->
        <svg xmlns="http://www.w3.org/2000/svg" width="100%" height="100%">
          <filter id="n">
            <feTurbulence
              type="fractalNoise"
              baseFrequency="0.65"
              numOctaves="3"
              stitchTiles="stitch"
            />
          </filter>
          <rect width="100%" height="100%" filter="url(#n)" />
        </svg>
      </div>
    </div>

    <div
      class="w-full max-w-[1100px] grid lg:grid-cols-2 bg-white/40 backdrop-blur-3xl rounded-3xl sm:rounded-[40px] border border-white/50 shadow-[0_32px_80px_-20px_rgba(0,0,0,0.1)] overflow-hidden relative z-10 transition-all duration-700"
    >
      <!-- Left side: Branding & Hero -->
      <div
        class="hidden lg:flex flex-col justify-between p-12 bg-gradient-to-br from-primary via-[#0f766e] to-primary relative overflow-hidden"
      >
        <!-- Abstract Decoration -->
        <div class="absolute inset-0 opacity-10 pointer-events-none">
          <svg viewBox="0 0 400 400" xmlns="http://www.w3.org/2000/svg" class="w-full h-full">
            <path
              d="M0,400 Q200,300 400,400 T800,400"
              fill="none"
              stroke="white"
              stroke-width="2"
            />
            <circle
              cx="300"
              cy="100"
              r="150"
              fill="none"
              stroke="white"
              stroke-width="1"
              stroke-dasharray="10 10"
            />
          </svg>
        </div>

        <div class="relative z-10">
          <div class="flex items-center gap-3 mb-12">
            <div class="flex flex-col">
              <span class="text-2xl font-black text-white tracking-tight leading-none"
                >{{ brand.name }}</span
              >
              <span class="text-[10px] font-bold text-white/70 uppercase tracking-[0.2em] mt-1"
                >Portal HR & Operasional</span
              >
            </div>
          </div>

          <h2 class="text-5xl font-black text-white leading-[1.1] mb-6">
            Tim lebih rapi,<br /><span class="text-white/70">kerja lebih lancar.</span>
          </h2>
          <p class="text-white/70 text-lg font-medium max-w-sm leading-relaxed">
            Kelola karyawan, kehadiran, cuti, dan pekerjaan harian dalam satu tempat yang mudah digunakan.
          </p>
        </div>

        <div
          class="relative z-10 mt-12 bg-white/5 backdrop-blur-lg border border-white/10 p-6 rounded-[28px]"
        >
          <div class="flex items-center gap-4">
            <div class="bg-emerald-400/20 p-2.5 rounded-xl">
              <ShieldCheck class="w-6 h-6 text-white" />
            </div>
            <div>
              <p class="text-white font-bold text-sm">Akses sesuai peran</p>
              <p class="text-white/50 text-[12px]">Informasi tim hanya tersedia untuk pengguna yang berwenang.</p>
            </div>
          </div>
        </div>
      </div>

      <!-- Right side: Login Form -->
      <div class="p-5 sm:p-10 md:p-14 lg:p-16 flex flex-col justify-center bg-white/40">
        <div class="max-w-[400px] mx-auto w-full">
          <div class="lg:hidden flex justify-center mb-6 sm:mb-10">
            <img :src="brand.logoUrl" alt="Logo" class="w-12 h-12 sm:w-14 sm:h-14" />
          </div>

          <div class="mb-6 sm:mb-10 text-center lg:text-left">
            <h1 class="text-2xl sm:text-3xl font-black tracking-tight text-slate-900 leading-tight">
              Masuk ke Sistem
            </h1>
            <p class="text-slate-500 mt-2 sm:mt-3 text-xs sm:text-sm font-medium">
              Masuk untuk melihat aktivitas dan data perusahaan Anda.
            </p>
          </div>

          <form @submit.prevent="handleLogin" class="space-y-6">
            <div
              v-if="errorMsg"
              class="p-4 bg-primary/5 border border-primary/10 text-primary rounded-2xl text-[12px] font-bold flex items-center gap-3 animate-bounce-subtle"
            >
              <AlertCircle class="w-4 h-4 shrink-0" />
              {{ errorMsg }}
            </div>

            <div class="space-y-2.5">
              <label class="block text-xs font-black text-slate-400 uppercase tracking-widest ml-1"
                >Email</label
              >
              <div class="relative group">
                <Mail
                  class="absolute left-5 top-1/2 -translate-y-1/2 w-5 h-5 text-slate-400 group-focus-within:text-primary transition-colors duration-300"
                />
                <input
                  v-model="email"
                  type="email"
                  autocomplete="username"
                  autocapitalize="none"
                  autocorrect="off"
                  spellcheck="false"
                  placeholder="admin@demo.local"
                  class="w-full pl-14 pr-4 py-3.5 sm:py-4.5 rounded-2xl border border-slate-200 bg-slate-50/50 font-bold text-slate-910 outline-none transition-all duration-300 focus:bg-white focus:border-primary focus:ring-4 focus:ring-primary/5 placeholder:text-slate-300"
                  required
                />
              </div>
            </div>

            <div class="space-y-2.5">
              <div class="flex items-center justify-between ml-1">
                <label class="block text-xs font-black text-slate-400 uppercase tracking-widest"
                  >Sandi Akses</label
                >
                <a
                  href="#"
                  @click.prevent="openForgotAccessModal"
                  class="text-[11px] font-black text-primary hover:text-[#770000] uppercase tracking-widest transition-colors"
                  >Lupa Akses?</a
                >
              </div>
              <div class="relative group">
                <Lock
                  class="absolute left-5 top-1/2 -translate-y-1/2 w-5 h-5 text-slate-400 group-focus-within:text-primary transition-colors duration-300"
                />
                <input
                  v-model="password"
                  type="password"
                  autocomplete="current-password"
                  autocapitalize="none"
                  autocorrect="off"
                  spellcheck="false"
                  placeholder="••••••••"
                  class="w-full pl-14 pr-4 py-3.5 sm:py-4.5 rounded-2xl border border-slate-200 bg-slate-50/50 font-bold text-slate-910 outline-none transition-all duration-300 focus:bg-white focus:border-primary focus:ring-4 focus:ring-primary/5 placeholder:text-slate-300"
                  required
                />
              </div>
            </div>

            <div class="flex items-center gap-3 ml-1">
              <div class="relative flex items-center">
                <input
                  v-model="rememberMe"
                  type="checkbox"
                  id="remember"
                  class="w-5 h-5 rounded-lg border-slate-300 text-primary focus:ring-primary/20 accent-primary cursor-pointer"
                />
              </div>
              <label
                for="remember"
                class="text-sm font-bold text-slate-500 cursor-pointer select-none"
                >Ingat Saya</label
              >
            </div>

            <button
              type="submit"
              :disabled="isLoading"
              class="group relative w-full py-4 sm:py-5 px-6 bg-slate-900 text-white font-black text-sm rounded-2xl overflow-hidden transition-all duration-500 hover:bg-primary shadow-xl hover:shadow-primary/30 active:scale-[0.98] disabled:opacity-70 disabled:pointer-events-none"
            >
              <div class="relative z-10 flex items-center justify-center gap-2">
                <Loader2 v-if="isLoading" class="w-5 h-5 animate-spin" />
                <span v-else class="flex items-center gap-2 uppercase tracking-[0.1em]">
                  Mulai Sesi
                  <ChevronRight
                    class="w-4 h-4 transition-transform duration-300 group-hover:translate-x-1"
                  />
                </span>
              </div>
              <!-- Hover slide effect -->
              <div
                class="absolute inset-0 bg-primary translate-y-full group-hover:translate-y-0 transition-transform duration-500"
              ></div>
            </button>
          </form>



          <p
            v-if="brand.supportEmail"
            class="mt-8 sm:mt-12 text-center text-[10px] sm:text-[12px] font-bold text-slate-400 uppercase tracking-widest"
          >
            Support:
            <a
               :href="`mailto:${brand.supportEmail}`"
              class="text-slate-900 hover:text-primary transition-colors decoration-dashed underline underline-offset-4"
              >{{ brand.supportEmail }}</a
            >
          </p>
        </div>
      </div>
    </div>

    <!-- Legal footer -->
    <div
      class="absolute bottom-10 left-1/2 -translate-x-1/2 text-center pointer-events-none opacity-40"
    >
      <p class="text-[10px] font-black text-slate-400 uppercase tracking-[0.3em]">
        {{ brand.name }} &copy; 2026
      </p>
    </div>

    <Teleport to="body">
      <Transition name="modal">
        <div
          v-if="showForgotAccessModal"
          class="fixed inset-0 z-[1000] flex items-center justify-center p-4"
        >
          <div
            class="absolute inset-0 bg-slate-950/60 backdrop-blur-sm"
            @click="closeForgotAccessModal"
          />

          <div
            class="relative z-10 w-full max-w-md rounded-[32px] bg-white shadow-2xl border border-slate-100 overflow-hidden"
          >
            <div class="p-7 border-b border-slate-100 flex items-start justify-between gap-4">
              <div class="flex items-center gap-4">
                <div class="w-12 h-12 rounded-2xl bg-primary/10 text-primary flex items-center justify-center shrink-0">
                  <KeyRound class="w-6 h-6" />
                </div>
                <div>
                  <h2 class="text-xl font-black text-slate-900 tracking-tight">Lupa Akses</h2>
                  <p class="text-xs font-semibold text-slate-500 mt-1 leading-relaxed">
                    Masukkan email terdaftar untuk menerima password sementara.
                  </p>
                </div>
              </div>
              <button
                type="button"
                @click="closeForgotAccessModal"
                :disabled="isForgotLoading"
                class="p-2 rounded-xl text-slate-400 hover:bg-slate-100 hover:text-slate-900 transition-colors disabled:opacity-50"
              >
                <X class="w-5 h-5" />
              </button>
            </div>

            <form class="p-7 space-y-5" @submit.prevent="handleForgotAccess">
              <div
                v-if="forgotErrorMsg"
                class="p-4 bg-primary/5 border border-primary/10 text-primary rounded-2xl text-[12px] font-bold flex items-center gap-3"
              >
                <AlertCircle class="w-4 h-4 shrink-0" />
                {{ forgotErrorMsg }}
              </div>

              <div class="space-y-2.5">
                <label class="block text-xs font-black text-slate-400 uppercase tracking-widest ml-1">
                  Email Terdaftar
                </label>
                <div class="relative group">
                  <Mail
                    class="absolute left-5 top-1/2 -translate-y-1/2 w-5 h-5 text-slate-400 group-focus-within:text-primary transition-colors duration-300"
                  />
                  <input
                    v-model="forgotEmail"
                    type="email"
                    autocomplete="email"
                    autocapitalize="none"
                    autocorrect="off"
                    spellcheck="false"
                    placeholder="nama@perusahaan.com"
                    class="w-full pl-14 pr-4 py-4 rounded-2xl border border-slate-200 bg-slate-50/50 font-bold text-slate-900 outline-none transition-all duration-300 focus:bg-white focus:border-primary focus:ring-4 focus:ring-primary/5 placeholder:text-slate-300"
                    :disabled="isForgotLoading"
                    autofocus
                  />
                </div>
              </div>

              <div class="flex gap-3 pt-2">
                <button
                  type="button"
                  @click="closeForgotAccessModal"
                  :disabled="isForgotLoading"
                  class="flex-1 py-4 rounded-2xl border border-slate-200 text-xs font-black text-slate-600 hover:bg-slate-50 transition-colors disabled:opacity-50"
                >
                  BATAL
                </button>
                <button
                  type="submit"
                  :disabled="isForgotLoading"
                  class="flex-[1.5] py-4 rounded-2xl bg-slate-900 text-white text-xs font-black hover:bg-primary transition-colors flex items-center justify-center gap-2 disabled:opacity-70"
                >
                  <Loader2 v-if="isForgotLoading" class="w-4 h-4 animate-spin" />
                  <span>{{ isForgotLoading ? 'MENGIRIM...' : 'KIRIM PASSWORD' }}</span>
                </button>
              </div>
            </form>
          </div>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>

<style scoped>
@keyframes bounce-subtle {
  0%,
  100% {
    transform: translateY(0);
  }
  50% {
    transform: translateY(-3px);
  }
}
.animate-bounce-subtle {
  animation: bounce-subtle 2s infinite ease-in-out;
}

.pl-14 {
  padding-left: 3.5rem;
}

.py-4\.5 {
  padding-top: 1.125rem;
  padding-bottom: 1.125rem;
}

.py-3\.5 {
  padding-top: 0.875rem;
  padding-bottom: 0.875rem;
}

.modal-enter-active,
.modal-leave-active {
  transition: opacity 0.2s ease;
}

.modal-enter-from,
.modal-leave-to {
  opacity: 0;
}
</style>
