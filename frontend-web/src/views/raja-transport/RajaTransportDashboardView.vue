<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useRajaTransportStore } from '@/stores/rajaTransport'
import {
  Truck,
  RotateCcw,
  Receipt,
  Wrench,
  Clock,
  MapPin,
  CheckCircle2,
  AlertTriangle,
  RefreshCw,
  ChevronRight,
  ArrowRight,
  Compass,
  ArrowUpRight,
  FileText,
  User,
  Calendar,
} from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'

const transportStore = useRajaTransportStore()
const router = useRouter()

const loading = ref(false)

async function refreshAll() {
  loading.value = true
  try {
    await Promise.all([
      transportStore.fetchDashboardStats(),
      transportStore.fetchTrips(),
      transportStore.fetchBackhauls(),
      transportStore.fetchExpenseClaims(),
    ])
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  refreshAll()
})

function formatCurrency(val: number) {
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    maximumFractionDigits: 0,
  }).format(val || 0)
}

function formatDate(d: string | undefined) {
  if (!d) return '-'
  return new Date(d).toLocaleDateString('id-ID', {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
  })
}

// Active trips filter
const activeTrips = computed(() => {
  return transportStore.trips.filter((t: any) => t.status === 'SCHEDULED' || t.status === 'IN_TRANSIT')
})

// Pending backhauls filter
const pendingBackhauls = computed(() => {
  return transportStore.backhauls.filter((b: any) => b.status === 'PENDING')
})

// Pending claims filter
const pendingClaims = computed(() => {
  return transportStore.expenseClaims.filter((c: any) => c.status === 'PENDING')
})

const navigationCards = [
  {
    title: 'Live Tracking Pengiriman (GPS)',
    description: 'Monitoring posisi armada secara real-time via satelit TraxCenter GPS.',
    path: '/rajatransport/tracking',
    icon: MapPin,
    color: 'amber',
    badge: 'Live Map',
    badgeClass: 'bg-amber-100 text-amber-800 border-amber-200',
    iconBg: 'bg-amber-500/10 text-amber-600',
  },
  {
    title: 'Riwayat & Penugasan Trip',
    description: 'Manajemen penugasan perjalanan antar provinsi, alokasi uang jalan & driver.',
    path: '/rajatransport/trips',
    icon: Truck,
    color: 'blue',
    badge: `${activeTrips.value.length} Berjalan`,
    badgeClass: 'bg-blue-100 text-blue-800 border-blue-200',
    iconBg: 'bg-blue-500/10 text-blue-600',
  },
  {
    title: 'Pengiriman Muatan Balik',
    description: 'Persetujuan fee 10% muatan balikan sopir ekspedisi dari kota tujuan.',
    path: '/rajatransport/backhauls',
    icon: RotateCcw,
    color: 'emerald',
    badge: `${pendingBackhauls.value.length} Pending`,
    badgeClass: 'bg-emerald-100 text-emerald-800 border-emerald-200',
    iconBg: 'bg-emerald-500/10 text-emerald-600',
  },
  {
    title: 'Klaim BBM & Pergantian Part',
    description: 'Verifikasi nota struk pengeluaran bensin, tol, dan part darurat armada.',
    path: '/rajatransport/claims',
    icon: Receipt,
    color: 'purple',
    badge: `${pendingClaims.value.length} Pending`,
    badgeClass: 'bg-purple-100 text-purple-800 border-purple-200',
    iconBg: 'bg-purple-500/10 text-purple-600',
  },
  {
    title: 'Kondisi & Perawatan Armada',
    description: 'Inspeksi checklist pra-jalan, kondisi ban, rem, oli, dan servis bengkel.',
    path: '/rajatransport/fleet-condition',
    icon: Wrench,
    color: 'rose',
    badge: `${transportStore.stats.fleets_in_maintenance_count || 0} Servis`,
    badgeClass: 'bg-rose-100 text-rose-800 border-rose-200',
    iconBg: 'bg-rose-500/10 text-rose-600',
  },
  {
    title: 'Absensi Driver Luar Kota',
    description: 'Presensi harian dan pencatatan lokasi sopir saat bertugas di perjalanan.',
    path: '/rajatransport/driver-attendance',
    icon: Clock,
    color: 'indigo',
    badge: 'Presensi',
    badgeClass: 'bg-indigo-100 text-indigo-800 border-indigo-200',
    iconBg: 'bg-indigo-500/10 text-indigo-600',
  },
]
</script>

<template>
  <div class="p-4 md:p-6 space-y-6 bg-[#f8fafc] min-h-screen">
    <!-- Header Banner -->
    <div
      class="flex flex-col md:flex-row justify-between items-start md:items-center gap-4 bg-gradient-to-r from-slate-900 via-slate-800 to-amber-950 p-8 rounded-[2rem] shadow-xl text-white"
    >
      <div class="space-y-1.5">
        <div class="flex items-center gap-2">
          <div class="p-2 rounded-2xl bg-amber-500/20 text-amber-400 border border-amber-500/30">
            <Truck class="h-6 w-6" />
          </div>
          <span class="text-xs font-black uppercase tracking-widest text-amber-400">
            Logistik & Ekspedisi Antar Provinsi
          </span>
        </div>
        <h1 class="text-3xl font-black tracking-tight text-white">Dashboard RajaTransport</h1>
        <p class="text-xs text-slate-300 max-w-2xl font-medium">
          Pusat kendali operasional ekspedisi armada, penugasan trip antar pulau/provinsi, monitoring GPS TraxCenter, pengiriman muatan kembali fee 10%, dan verifikasi klaim operasional.
        </p>
      </div>

      <div class="flex items-center gap-3">
        <Button
          variant="outline"
          @click="refreshAll"
          :disabled="loading"
          class="rounded-2xl border-white/20 text-black hover:bg-white/10 font-bold text-xs h-11"
        >
          <RefreshCw class="h-4 w-4 mr-2" :class="{ 'animate-spin': loading }" /> Segarkan Data
        </Button>
      </div>
    </div>

    <!-- KPI Metric Cards -->
    <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-5 gap-4">
      <!-- 1. Trip Aktif -->
      <div
        @click="router.push('/rajatransport/trips')"
        class="bg-white p-5 rounded-3xl border border-slate-200/80 shadow-sm space-y-2 cursor-pointer hover:shadow-md transition-all hover:border-blue-200 group"
      >
        <div class="flex items-center justify-between text-slate-400">
          <span class="text-[10px] font-black uppercase tracking-wider">Trip Berjalan</span>
          <div class="p-2 rounded-xl bg-blue-50 text-blue-600 group-hover:scale-110 transition-transform">
            <Truck class="h-4 w-4" />
          </div>
        </div>
        <div class="text-2xl font-black text-slate-900 font-mono">
          {{ transportStore.stats.active_trips_count || 0 }}
        </div>
        <div class="text-[10px] text-slate-400 font-bold flex items-center justify-between">
          <span>Armada di jalan / terjadwal</span>
          <ChevronRight class="h-3 w-3 text-slate-300 group-hover:text-blue-600" />
        </div>
      </div>

      <!-- 2. Trip Selesai -->
      <div
        @click="router.push('/rajatransport/trips')"
        class="bg-white p-5 rounded-3xl border border-slate-200/80 shadow-sm space-y-2 cursor-pointer hover:shadow-md transition-all hover:border-emerald-200 group"
      >
        <div class="flex items-center justify-between text-slate-400">
          <span class="text-[10px] font-black uppercase tracking-wider">Trip Sukses</span>
          <div class="p-2 rounded-xl bg-emerald-50 text-emerald-600 group-hover:scale-110 transition-transform">
            <CheckCircle2 class="h-4 w-4" />
          </div>
        </div>
        <div class="text-2xl font-black text-slate-900 font-mono">
          {{ transportStore.stats.completed_trips_count || 0 }}
        </div>
        <div class="text-[10px] text-slate-400 font-bold flex items-center justify-between">
          <span>Pengiriman terverifikasi</span>
          <ChevronRight class="h-3 w-3 text-slate-300 group-hover:text-emerald-600" />
        </div>
      </div>

      <!-- 3. Muatan Balik (Fee 10%) -->
      <div
        @click="router.push('/rajatransport/backhauls')"
        class="bg-white p-5 rounded-3xl border border-slate-200/80 shadow-sm space-y-2 cursor-pointer hover:shadow-md transition-all hover:border-amber-200 group"
      >
        <div class="flex items-center justify-between text-slate-400">
          <span class="text-[10px] font-black uppercase tracking-wider">Fee Balikan (10%)</span>
          <div class="p-2 rounded-xl bg-amber-50 text-amber-600 group-hover:scale-110 transition-transform">
            <RotateCcw class="h-4 w-4" />
          </div>
        </div>
        <div class="text-lg font-black text-amber-800 font-mono truncate">
          {{ formatCurrency(transportStore.stats.approved_backhaul_fee_total) }}
        </div>
        <div class="text-[10px] text-slate-400 font-bold flex items-center justify-between">
          <span>{{ transportStore.stats.pending_backhauls_count }} pending approval</span>
          <ChevronRight class="h-3 w-3 text-slate-300 group-hover:text-amber-600" />
        </div>
      </div>

      <!-- 4. Klaim BBM & Part Pending -->
      <div
        @click="router.push('/rajatransport/claims')"
        class="bg-white p-5 rounded-3xl border border-slate-200/80 shadow-sm space-y-2 cursor-pointer hover:shadow-md transition-all hover:border-purple-200 group"
      >
        <div class="flex items-center justify-between text-slate-400">
          <span class="text-[10px] font-black uppercase tracking-wider">Klaim Pending</span>
          <div class="p-2 rounded-xl bg-purple-50 text-purple-600 group-hover:scale-110 transition-transform">
            <Receipt class="h-4 w-4" />
          </div>
        </div>
        <div class="text-2xl font-black text-purple-900 font-mono">
          {{ transportStore.stats.pending_claims_count || 0 }}
        </div>
        <div class="text-[10px] text-slate-400 font-bold flex items-center justify-between">
          <span>Menunggu review bon nota</span>
          <ChevronRight class="h-3 w-3 text-slate-300 group-hover:text-purple-600" />
        </div>
      </div>

      <!-- 5. Armada Servis -->
      <div
        @click="router.push('/rajatransport/fleet-condition')"
        class="bg-white p-5 rounded-3xl border border-slate-200/80 shadow-sm space-y-2 cursor-pointer hover:shadow-md transition-all hover:border-rose-200 group"
      >
        <div class="flex items-center justify-between text-slate-400">
          <span class="text-[10px] font-black uppercase tracking-wider">Armada Servis</span>
          <div class="p-2 rounded-xl bg-rose-50 text-rose-600 group-hover:scale-110 transition-transform">
            <Wrench class="h-4 w-4" />
          </div>
        </div>
        <div class="text-2xl font-black text-rose-600 font-mono">
          {{ transportStore.stats.fleets_in_maintenance_count || 0 }}
        </div>
        <div class="text-[10px] text-slate-400 font-bold flex items-center justify-between">
          <span>Perbaikan & ganti part</span>
          <ChevronRight class="h-3 w-3 text-slate-300 group-hover:text-rose-600" />
        </div>
      </div>
    </div>

    <!-- Quick Navigation Shortcuts (Pintu Masuk Setiap Halaman) -->
    <div>
      <div class="flex items-center justify-between mb-3 px-1">
        <div>
          <h2 class="text-base font-black text-slate-900">Akses Modul Ekspedisi</h2>
          <p class="text-xs text-slate-500">Pilih menu untuk membuka pengelolaan data secara lengkap</p>
        </div>
      </div>

      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        <div
          v-for="card in navigationCards"
          :key="card.path"
          @click="router.push(card.path)"
          class="bg-white p-6 rounded-3xl border border-slate-200/80 shadow-xs hover:shadow-md transition-all cursor-pointer group hover:border-slate-300 flex flex-col justify-between"
        >
          <div class="space-y-3">
            <div class="flex items-center justify-between">
              <div :class="['p-3 rounded-2xl', card.iconBg]">
                <component :is="card.icon" class="w-6 h-6" />
              </div>
              <Badge :class="['text-[11px] font-bold rounded-xl px-2.5 py-0.5 border', card.badgeClass]">
                {{ card.badge }}
              </Badge>
            </div>
            <div>
              <h3 class="text-base font-black text-slate-900 group-hover:text-primary transition-colors flex items-center gap-1.5">
                {{ card.title }}
                <ArrowUpRight class="w-4 h-4 text-slate-300 group-hover:text-primary transition-all group-hover:translate-x-0.5 group-hover:-translate-y-0.5" />
              </h3>
              <p class="text-xs text-slate-500 mt-1 leading-relaxed">
                {{ card.description }}
              </p>
            </div>
          </div>

          <div class="pt-4 mt-4 border-t border-slate-100 flex items-center justify-between text-xs font-bold text-slate-600 group-hover:text-slate-900">
            <span>Buka Halaman</span>
            <ArrowRight class="w-4 h-4 text-slate-400 group-hover:text-primary transition-transform group-hover:translate-x-1" />
          </div>
        </div>
      </div>
    </div>

    <!-- Quick Overview Tables (Trip Aktif & Tindakan Tertunda) -->
    <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
      <!-- 1. Ringkasan Trip Aktif Sedang di Jalan -->
      <div class="bg-white p-6 rounded-3xl border border-slate-200/80 shadow-xs space-y-4">
        <div class="flex items-center justify-between">
          <div class="flex items-center gap-2.5">
            <div class="p-2 rounded-xl bg-blue-50 text-blue-600">
              <Truck class="w-5 h-5" />
            </div>
            <div>
              <h3 class="text-sm font-black text-slate-900">Trip Sedang Berjalan</h3>
              <p class="text-[11px] text-slate-400">Armada yang sedang menuju kota tujuan / dijadwalkan</p>
            </div>
          </div>
          <Button
            variant="ghost"
            size="sm"
            @click="router.push('/rajatransport/trips')"
            class="text-xs font-bold text-blue-600 hover:text-blue-700 hover:bg-blue-50 rounded-xl"
          >
            Lihat Semua <ArrowRight class="w-3.5 h-3.5 ml-1" />
          </Button>
        </div>

        <div v-if="activeTrips.length === 0" class="py-10 text-center text-slate-400 text-xs font-bold">
          Tidak ada trip armada yang sedang berjalan saat ini.
        </div>

        <div v-else class="space-y-2.5">
          <div
            v-for="trip in activeTrips.slice(0, 4)"
            :key="trip.id"
            @click="router.push('/rajatransport/trips')"
            class="p-3.5 rounded-2xl bg-slate-50/80 hover:bg-slate-100/80 border border-slate-200/60 transition-all cursor-pointer flex items-center justify-between"
          >
            <div class="space-y-1">
              <div class="flex items-center gap-2">
                <span class="font-mono font-black text-xs text-slate-900">{{ trip.trip_number }}</span>
                <Badge
                  :class="trip.status === 'IN_TRANSIT' ? 'bg-blue-100 text-blue-800' : 'bg-amber-100 text-amber-800'"
                  class="text-[10px] font-bold"
                >
                  {{ trip.status === 'IN_TRANSIT' ? 'Di Jalan' : 'Terjadwal' }}
                </Badge>
              </div>
              <div class="text-xs text-slate-600 font-medium flex items-center gap-2">
                <span>{{ trip.vehicle?.license_plate || '-' }}</span>
                <span class="text-slate-300">•</span>
                <span>Driver: {{ trip.driver?.first_name || 'Driver' }}</span>
              </div>
              <div class="text-[11px] text-slate-500 font-medium flex items-center gap-1">
                <MapPin class="w-3 h-3 text-slate-400" />
                <span>{{ trip.origin_city }} → {{ trip.destination_city }}</span>
              </div>
            </div>

            <ChevronRight class="w-4 h-4 text-slate-400" />
          </div>
        </div>
      </div>

      <!-- 2. Tindakan Tertunda (Pending Backhauls & Claims) -->
      <div class="bg-white p-6 rounded-3xl border border-slate-200/80 shadow-xs space-y-4">
        <div class="flex items-center justify-between">
          <div class="flex items-center gap-2.5">
            <div class="p-2 rounded-xl bg-amber-50 text-amber-600">
              <AlertTriangle class="w-5 h-5" />
            </div>
            <div>
              <h3 class="text-sm font-black text-slate-900">Persetujuan Menunggu Review</h3>
              <p class="text-[11px] text-slate-400">Pengajuan muatan balik dan klaim nota yang butuh approval</p>
            </div>
          </div>
        </div>

        <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
          <!-- Backhaul pending box -->
          <div
            @click="router.push('/rajatransport/backhauls')"
            class="p-4 rounded-2xl bg-amber-50/50 border border-amber-200/70 hover:bg-amber-50 transition-all cursor-pointer space-y-2"
          >
            <div class="flex items-center justify-between text-amber-800">
              <span class="text-[11px] font-black uppercase">Muatan Balik</span>
              <RotateCcw class="w-4 h-4 text-amber-600" />
            </div>
            <div class="text-2xl font-black text-amber-900 font-mono">
              {{ pendingBackhauls.length }}
            </div>
            <div class="text-[11px] text-amber-700 font-medium flex items-center justify-between">
              <span>Menunggu persetujuan fee 10%</span>
              <ArrowRight class="w-3.5 h-3.5 text-amber-600" />
            </div>
          </div>

          <!-- Claims pending box -->
          <div
            @click="router.push('/rajatransport/claims')"
            class="p-4 rounded-2xl bg-purple-50/50 border border-purple-200/70 hover:bg-purple-50 transition-all cursor-pointer space-y-2"
          >
            <div class="flex items-center justify-between text-purple-800">
              <span class="text-[11px] font-black uppercase">Klaim BBM & Part</span>
              <Receipt class="w-4 h-4 text-purple-600" />
            </div>
            <div class="text-2xl font-black text-purple-900 font-mono">
              {{ pendingClaims.length }}
            </div>
            <div class="text-[11px] text-purple-700 font-medium flex items-center justify-between">
              <span>Menunggu verifikasi bon/nota</span>
              <ArrowRight class="w-3.5 h-3.5 text-purple-600" />
            </div>
          </div>
        </div>

        <!-- Quick info note -->
        <div class="p-3.5 rounded-2xl bg-slate-50 border border-slate-200/60 text-xs text-slate-500 leading-relaxed flex items-center justify-between">
          <div class="flex items-center gap-2">
            <MapPin class="w-4 h-4 text-amber-600 shrink-0" />
            <span>Ingin memantau pergerakan armada langsung di peta TraxCenter?</span>
          </div>
          <Button
            size="sm"
            @click="router.push('/rajatransport/tracking')"
            class="rounded-xl text-xs font-bold bg-amber-600 hover:bg-amber-700 text-white shrink-0 ml-2"
          >
            Buka Peta GPS
          </Button>
        </div>
      </div>
    </div>
  </div>
</template>
