<script setup lang="ts">
import { ref, shallowRef, computed, onMounted, onUnmounted } from 'vue'
import L from 'leaflet'
import 'leaflet/dist/leaflet.css'
import apiClient from '@/api/axios'
import { vehicleApi } from '@/api/vehicle'
import { useRajaTransportStore } from '@/stores/rajaTransport'
import {
  Truck,
  MapPin,
  RefreshCw,
  Search,
  Radio,
  Navigation,
  Gauge,
  Compass,
  Clock,
  User,
  AlertTriangle,
  ExternalLink,
  ChevronRight,
  Filter,
} from 'lucide-vue-next'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { toast } from 'vue-sonner'

// Fix default Leaflet icon assets
import iconRetinaUrl from 'leaflet/dist/images/marker-icon-2x.png'
import iconUrl from 'leaflet/dist/images/marker-icon.png'
import shadowUrl from 'leaflet/dist/images/marker-shadow.png'

L.Icon.Default.mergeOptions({
  iconRetinaUrl,
  iconUrl,
  shadowUrl,
})

interface VehicleTrackingItem {
  id: string
  name: string
  license_plate: string
  model?: string
  type?: string
  status: string
  gps_imei?: string
  active_assignment?: any
  // Live GPS Data from TraxCenter
  lat?: number | null
  lon?: number | null
  speed?: number
  heading?: number
  is_online?: boolean
  last_updated?: Date
  loading_gps?: boolean
}

const transportStore = useRajaTransportStore()
const map = shallowRef<L.Map | null>(null)
const markers = shallowRef<{ [key: string]: L.Marker }>({})
const mapContainer = ref<HTMLElement | null>(null)

const vehicles = ref<VehicleTrackingItem[]>([])
const loading = ref(false)
const isAutoRefresh = ref(true)
const autoRefreshInterval = ref<any>(null)
const selectedVehicleId = ref<string | null>(null)
const searchQuery = ref('')
const filterStatus = ref<'ALL' | 'ONLINE' | 'ACTIVE_TRIP' | 'OFFLINE'>('ALL')

// Computed tracking list with search and filters
const filteredVehicles = computed(() => {
  return vehicles.value.filter((v) => {
    // Search
    if (searchQuery.value.trim()) {
      const q = searchQuery.value.toLowerCase()
      const matchPlate = v.license_plate?.toLowerCase().includes(q)
      const matchName = v.name?.toLowerCase().includes(q)
      const matchImei = v.gps_imei?.toLowerCase().includes(q)
      const matchDriver = v.active_assignment?.driver_name?.toLowerCase().includes(q)
      const matchDo = v.active_assignment?.do_number?.toLowerCase().includes(q)
      if (!matchPlate && !matchName && !matchImei && !matchDriver && !matchDo) return false
    }

    // Filter
    if (filterStatus.value === 'ONLINE') {
      return v.is_online === true
    }
    if (filterStatus.value === 'ACTIVE_TRIP') {
      return !!v.active_assignment
    }
    if (filterStatus.value === 'OFFLINE') {
      return !v.is_online
    }
    return true
  })
})

const onlineCount = computed(() => vehicles.value.filter((v) => v.is_online).length)
const activeTripCount = computed(() => vehicles.value.filter((v) => !!v.active_assignment).length)

// Create Truck Marker with Heading Rotation
const createTruckIcon = (heading = 0, isOnline = false, hasActiveTrip = false) => {
  const bgClass = isOnline
    ? hasActiveTrip
      ? 'bg-blue-600 text-white border-2 border-white ring-2 ring-blue-400'
      : 'bg-emerald-600 text-white border-2 border-white ring-2 ring-emerald-400'
    : 'bg-slate-400 text-white border-2 border-white ring-2 ring-slate-300'

  return L.divIcon({
    className: 'custom-trax-marker',
    html: `
      <div class="relative flex items-center justify-center">
        <div class="h-9 w-9 rounded-2xl ${bgClass} shadow-lg flex items-center justify-center transition-all duration-300 transform" style="transform: rotate(${heading}deg);">
          <svg class="h-5 w-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
            <polygon points="12 2 19 21 12 17 5 21 12 2"/>
          </svg>
        </div>
        ${
          isOnline
            ? '<span class="absolute -top-1 -right-1 flex h-3 w-3"><span class="animate-ping absolute inline-flex h-full w-full rounded-full bg-emerald-400 opacity-75"></span><span class="relative inline-flex rounded-full h-3 w-3 bg-emerald-500 border-2 border-white"></span></span>'
            : ''
        }
      </div>
    `,
    iconSize: [36, 36],
    iconAnchor: [18, 18],
    popupAnchor: [0, -18],
  })
}

// Fetch Vehicles & TraxCenter GPS
const fetchVehiclesAndGps = async () => {
  try {
    loading.value = true
    const res = await vehicleApi.getVehicles()
    const allVehicles: any[] = res.data || []

    // Fetch active trips from RajaTransport store if not fetched yet
    if (transportStore.trips.length === 0) {
      await transportStore.fetchTrips({ status: 'ACTIVE' })
    }

    // Map trips to vehicle active_assignment if not already present
    const mapped: VehicleTrackingItem[] = allVehicles.map((v) => {
      const activeTrip = transportStore.trips.find(
        (t) => t.vehicle_id === v.id && (t.status === 'DEPARTED' || t.status === 'SCHEDULED'),
      )

      let assignment = v.active_assignment
      if (!assignment && activeTrip) {
        assignment = {
          delivery_type: 'RAJA_TRANSPORT',
          type_label: 'Ekspedisi Antar Provinsi',
          do_number: activeTrip.trip_number || 'TRIP-EXP',
          driver_name: activeTrip.driver?.name || 'Driver Ekspedisi',
          origin_name: activeTrip.origin_city || 'Pusat',
          destination_name: activeTrip.destination_city || 'Tujuan',
        }
      }

      // Preserve existing live coordinates if vehicle already in list
      const existing = vehicles.value.find((ev) => ev.id === v.id)

      return {
        id: v.id,
        name: v.name,
        license_plate: v.license_plate,
        model: v.model,
        type: v.type,
        status: v.status,
        gps_imei: v.gps_imei,
        active_assignment: assignment,
        lat: existing?.lat ?? null,
        lon: existing?.lon ?? null,
        speed: existing?.speed ?? 0,
        heading: existing?.heading ?? 0,
        is_online: existing?.is_online ?? false,
        last_updated: existing?.last_updated,
        loading_gps: false,
      }
    })

    vehicles.value = mapped

    // Now trigger GPS polling for each vehicle with IMEI
    await fetchLiveTraxCenterAll()
  } catch (err) {
    console.error('Failed to load vehicles:', err)
  } finally {
    loading.value = false
  }
}

// Query TraxCenter open API via Backend Proxy
const fetchLiveTraxCenterAll = async () => {
  const promises = vehicles.value
    .filter((v) => !!v.gps_imei)
    .map(async (v) => {
      v.loading_gps = true
      try {
        const res = await apiClient.get(`/vehicles/traxcenter/${v.gps_imei}`)
        const data = res.data
        if (data && data.lat !== false && data.lat !== 'false' && data.lat !== undefined && data.lat !== null) {
          const latNum = parseFloat(data.lat)
          const lonNum = parseFloat(data.lon)
          if (!isNaN(latNum) && !isNaN(lonNum) && (latNum !== 0 || lonNum !== 0)) {
            v.lat = latNum
            v.lon = lonNum
            v.speed = typeof data.speed === 'number' ? data.speed : parseFloat(data.speed) || 0
            v.heading = typeof data.heading === 'number' ? data.heading : parseFloat(data.heading) || 0
            v.is_online = true
            v.last_updated = new Date()
          } else {
            v.is_online = false
          }
        } else {
          v.is_online = false
        }
      } catch (err) {
        v.is_online = false
      } finally {
        v.loading_gps = false
      }
    })

  await Promise.allSettled(promises)
  updateMapMarkers()
}

// Update Leaflet Map Markers
const updateMapMarkers = () => {
  if (!map.value) return

  const bounds = L.latLngBounds([])
  let hasValidCoords = false

  vehicles.value.forEach((v) => {
    if (v.lat != null && v.lon != null) {
      hasValidCoords = true
      const latlng = L.latLng(v.lat, v.lon)
      bounds.extend(latlng)

      const icon = createTruckIcon(v.heading || 0, v.is_online, !!v.active_assignment)

      const popupContent = `
        <div class="p-1 min-w-[220px] font-sans">
          <div class="flex items-center justify-between border-b pb-2 mb-2">
            <span class="font-mono font-black text-sm text-slate-900">${v.license_plate}</span>
            <span class="text-[10px] font-black uppercase px-2 py-0.5 rounded-full ${
              v.is_online ? 'bg-emerald-100 text-emerald-700' : 'bg-slate-100 text-slate-600'
            }">
              ${v.is_online ? 'LIVE ONLINE' : 'NO SIGNAL'}
            </span>
          </div>
          <div class="text-xs font-bold text-slate-700 mb-1">${v.name}</div>
          ${
            v.active_assignment
              ? `<div class="p-2 mb-2 rounded-xl bg-blue-50 border border-blue-100 text-[11px] font-medium text-blue-900">
                  <div class="font-bold">Trip: ${v.active_assignment.do_number}</div>
                  <div>Driver: ${v.active_assignment.driver_name}</div>
                  <div>Rute: ${v.active_assignment.origin_name} → ${v.active_assignment.destination_name}</div>
                </div>`
              : '<div class="text-[11px] text-slate-400 mb-2">Tidak ada penugasan aktif (Standby)</div>'
          }
          <div class="grid grid-cols-2 gap-2 text-[11px] bg-slate-50 p-2 rounded-xl border border-slate-100 mb-2">
            <div>
              <span class="text-slate-400 block text-[9px] uppercase font-bold">Kecepatan</span>
              <span class="font-black font-mono text-slate-800">${v.speed || 0} km/h</span>
            </div>
            <div>
              <span class="text-slate-400 block text-[9px] uppercase font-bold">Arah</span>
              <span class="font-black font-mono text-slate-800">${v.heading || 0}°</span>
            </div>
          </div>
          <div class="text-[10px] text-slate-400 flex items-center justify-between">
            <span class="font-mono">IMEI: ${v.gps_imei || '-'}</span>
            <a href="https://www.google.com/maps?q=${v.lat},${v.lon}" target="_blank" class="text-blue-600 font-bold hover:underline flex items-center gap-0.5">
              G-Maps
            </a>
          </div>
        </div>
      `

      const existingMarker = markers.value[v.id]
      if (existingMarker) {
        existingMarker.setLatLng(latlng)
        existingMarker.setIcon(icon)
        existingMarker.setPopupContent(popupContent)
      } else {
        const marker = L.marker(latlng, { icon })
          .addTo(map.value!)
          .bindPopup(popupContent)

        marker.on('click', () => {
          selectedVehicleId.value = v.id
        })

        markers.value[v.id] = marker
      }
    }
  })

  // Fit bounds if first load and has coordinates
  if (hasValidCoords && !selectedVehicleId.value) {
    map.value.fitBounds(bounds, { padding: [50, 50], maxZoom: 14 })
  }
}

const focusVehicle = (v: VehicleTrackingItem) => {
  selectedVehicleId.value = v.id
  if (v.lat != null && v.lon != null && map.value) {
    map.value.setView([v.lat, v.lon], 16, { animate: true })
    const marker = markers.value[v.id]
    if (marker) {
      marker.openPopup()
    }
  } else {
    toast.warning(`Armada ${v.license_plate} belum mengirim koordinat GPS.`)
  }
}

const toggleAutoRefresh = () => {
  isAutoRefresh.value = !isAutoRefresh.value
  if (isAutoRefresh.value) {
    startAutoRefresh()
    toast.success('Auto-refresh diaktifkan (tiap 15 detik)')
  } else {
    stopAutoRefresh()
    toast.info('Auto-refresh dijeda')
  }
}

const startAutoRefresh = () => {
  stopAutoRefresh()
  autoRefreshInterval.value = setInterval(() => {
    fetchLiveTraxCenterAll()
  }, 15000)
}

const stopAutoRefresh = () => {
  if (autoRefreshInterval.value) {
    clearInterval(autoRefreshInterval.value)
    autoRefreshInterval.value = null
  }
}

onMounted(() => {
  // Init Leaflet map
  map.value = L.map('trax-delivery-map').setView([-7.550687, 110.75848], 11) // Default Solo / Jateng area

  L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
    attribution: '&copy; OpenStreetMap contributors | TraxCenter GPS',
    maxZoom: 19,
  }).addTo(map.value!)

  fetchVehiclesAndGps()
  if (isAutoRefresh.value) {
    startAutoRefresh()
  }
})

onUnmounted(() => {
  stopAutoRefresh()
  if (map.value) {
    map.value.remove()
    map.value = null
  }
})
</script>

<template>
  <div class="space-y-6">
    <!-- Top Bar Controls -->
    <div class="flex flex-col lg:flex-row justify-between items-start lg:items-center gap-4 bg-white p-6 rounded-3xl border border-slate-200/80 shadow-sm">
      <div class="space-y-1">
        <div class="flex items-center gap-2">
          <Badge class="bg-amber-100 text-amber-800 border-amber-200 text-[10px] font-black uppercase px-2 py-0.5 rounded-full flex items-center gap-1.5">
            <Radio class="h-3 w-3 text-amber-600 animate-pulse" />
            TraxCenter Open API Live
          </Badge>
          <span class="text-xs text-slate-400 font-medium">Update real-time posisi armada via IMEI GPS</span>
        </div>
        <h2 class="text-xl font-black text-slate-900 flex items-center gap-2">
          Monitoring & Tracking Armada Ekspedisi
        </h2>
      </div>

      <div class="flex flex-wrap items-center gap-2.5">
        <!-- Live status chips -->
        <div class="flex items-center gap-2 bg-slate-50 border border-slate-200/80 px-3 py-1.5 rounded-2xl text-xs font-bold text-slate-700">
          <div class="flex items-center gap-1.5">
            <span class="h-2 w-2 rounded-full bg-emerald-500"></span>
            <span>{{ onlineCount }} Online</span>
          </div>
          <span class="text-slate-300">|</span>
          <div class="flex items-center gap-1.5 text-blue-600">
            <Truck class="h-3.5 w-3.5" />
            <span>{{ activeTripCount }} Sedang Berjalan</span>
          </div>
        </div>

        <Button
          variant="outline"
          size="sm"
          @click="toggleAutoRefresh"
          class="rounded-2xl text-xs font-bold h-10 px-4"
          :class="isAutoRefresh ? 'border-emerald-300 text-emerald-700 bg-emerald-50' : 'text-slate-600'"
        >
          <Clock class="h-3.5 w-3.5 mr-1.5" />
          {{ isAutoRefresh ? 'Auto-Refresh (15s)' : 'Auto-Refresh Jeda' }}
        </Button>

        <Button
          @click="fetchVehiclesAndGps"
          :disabled="loading"
          class="rounded-2xl text-xs font-bold h-10 px-4 bg-slate-900 text-white hover:bg-slate-800 shadow-sm"
        >
          <RefreshCw class="h-3.5 w-3.5 mr-1.5" :class="{ 'animate-spin': loading }" />
          Refresh Sekarang
        </Button>
      </div>
    </div>

    <!-- Main Grid: Map & Side Fleet List -->
    <div class="grid grid-cols-1 lg:grid-cols-12 gap-6">
      <!-- LEFT: Leaflet Map (8/12) -->
      <div class="lg:col-span-8 bg-white rounded-3xl p-3 border border-slate-200/80 shadow-sm relative overflow-hidden flex flex-col min-h-[580px]">
        <div id="trax-delivery-map" class="w-full h-full min-h-[560px] rounded-2xl z-0"></div>

        <!-- Floating map legend -->
        <div class="absolute bottom-6 left-6 bg-white/95 backdrop-blur-md p-3 rounded-2xl border border-slate-200/80 shadow-lg z-[400] text-xs font-medium space-y-1.5">
          <div class="text-[10px] font-black uppercase tracking-wider text-slate-400">Petunjuk Indikator</div>
          <div class="flex items-center gap-2">
            <div class="h-3 w-3 rounded-full bg-blue-600 border border-white"></div>
            <span class="text-slate-700 text-[11px] font-bold">Online & Trip Berjalan</span>
          </div>
          <div class="flex items-center gap-2">
            <div class="h-3 w-3 rounded-full bg-emerald-600 border border-white"></div>
            <span class="text-slate-700 text-[11px] font-bold">Online Standby</span>
          </div>
          <div class="flex items-center gap-2">
            <div class="h-3 w-3 rounded-full bg-slate-400 border border-white"></div>
            <span class="text-slate-500 text-[11px]">No Signal / Belum Ada GPS</span>
          </div>
        </div>
      </div>

      <!-- RIGHT: Fleet List & Search (4/12) -->
      <div class="lg:col-span-4 space-y-4">
        <div class="bg-white p-5 rounded-3xl border border-slate-200/80 shadow-sm space-y-3">
          <!-- Search box -->
          <div class="relative">
            <input
              v-model="searchQuery"
              type="text"
              placeholder="Cari plat, armada, driver, atau DO..."
              class="w-full pl-10 pr-4 py-2.5 bg-slate-50 border border-slate-200 rounded-2xl text-xs font-bold text-slate-800 placeholder:text-slate-400 focus:bg-white focus:border-amber-500 focus:ring-4 focus:ring-amber-500/10 outline-none transition-all"
            />
            <Search class="absolute left-3.5 top-1/2 -translate-y-1/2 h-4 w-4 text-slate-400 pointer-events-none" />
          </div>

          <!-- Filter status pills -->
          <div class="flex items-center gap-1.5 overflow-x-auto no-scrollbar pb-1">
            <button
              v-for="f in [
                { id: 'ALL', label: 'Semua' },
                { id: 'ONLINE', label: 'Online' },
                { id: 'ACTIVE_TRIP', label: 'Trip Aktif' },
                { id: 'OFFLINE', label: 'Offline' },
              ]"
              :key="f.id"
              @click="filterStatus = f.id as any"
              :class="[
                filterStatus === f.id
                  ? 'bg-slate-900 text-white font-black'
                  : 'bg-slate-100 text-slate-600 hover:bg-slate-200 font-bold',
              ]"
              class="px-3 py-1.5 rounded-xl text-[11px] transition-all whitespace-nowrap"
            >
              {{ f.label }}
            </button>
          </div>
        </div>

        <!-- Fleet Scrollable Cards -->
        <div class="space-y-3 max-h-[520px] overflow-y-auto pr-1">
          <div
            v-if="filteredVehicles.length === 0"
            class="bg-white p-8 rounded-3xl border border-slate-200/80 text-center space-y-2 text-slate-400"
          >
            <Truck class="h-8 w-8 mx-auto text-slate-300" />
            <p class="text-xs font-bold">Tidak ada armada yang sesuai filter</p>
          </div>

          <div
            v-for="v in filteredVehicles"
            :key="v.id"
            @click="focusVehicle(v)"
            class="bg-white p-4 rounded-3xl border transition-all cursor-pointer hover:shadow-md relative group"
            :class="[
              selectedVehicleId === v.id
                ? 'border-amber-500 ring-2 ring-amber-500/10 shadow-sm'
                : 'border-slate-200/80 hover:border-slate-300',
            ]"
          >
            <!-- Card Header -->
            <div class="flex items-start justify-between gap-2 mb-2">
              <div>
                <div class="flex items-center gap-2">
                  <span class="font-black text-sm font-mono text-slate-900">{{ v.license_plate }}</span>
                  <Badge variant="outline" class="text-[9px] font-bold px-1.5 py-0 rounded-md border-slate-200 text-slate-500 uppercase">
                    {{ v.type || 'TRUCK' }}
                  </Badge>
                </div>
                <p class="text-xs font-bold text-slate-600">{{ v.name }}</p>
              </div>

              <div class="flex items-center gap-1.5">
                <span
                  class="flex items-center gap-1 text-[10px] font-black px-2 py-0.5 rounded-full"
                  :class="v.is_online ? 'bg-emerald-50 text-emerald-700 border border-emerald-200' : 'bg-slate-100 text-slate-500'"
                >
                  <span class="h-1.5 w-1.5 rounded-full" :class="v.is_online ? 'bg-emerald-500 animate-ping' : 'bg-slate-400'"></span>
                  {{ v.is_online ? 'ONLINE' : 'NO SIGNAL' }}
                </span>
              </div>
            </div>

            <!-- Active Delivery Info -->
            <div
              v-if="v.active_assignment"
              class="mb-2 p-2.5 rounded-2xl bg-blue-50/80 border border-blue-100/80 text-[11px] space-y-1"
            >
              <div class="flex items-center justify-between text-blue-700 font-black text-[10px] uppercase">
                <span class="flex items-center gap-1">
                  <Truck class="h-3 w-3" />
                  {{ v.active_assignment.do_number }}
                </span>
                <span>{{ v.active_assignment.type_label }}</span>
              </div>
              <div class="font-bold text-slate-800 flex items-center justify-between">
                <span>Driver: {{ v.active_assignment.driver_name }}</span>
              </div>
              <div class="text-slate-600 text-[10px] truncate">
                Rute: {{ v.active_assignment.origin_name }} → {{ v.active_assignment.destination_name }}
              </div>
            </div>

            <!-- GPS Metrics Footer -->
            <div class="flex items-center justify-between pt-2 border-t border-slate-100 text-[11px]">
              <div class="flex items-center gap-3">
                <span class="text-slate-500 flex items-center gap-1">
                  <Gauge class="h-3.5 w-3.5 text-slate-400" />
                  <span class="font-bold font-mono text-slate-800">{{ v.speed || 0 }}</span> km/h
                </span>
                <span class="text-slate-500 flex items-center gap-1">
                  <Compass class="h-3.5 w-3.5 text-slate-400" />
                  <span class="font-bold font-mono text-slate-800">{{ v.heading || 0 }}°</span>
                </span>
              </div>

              <div class="flex items-center gap-1">
                <span v-if="v.gps_imei" class="text-[9px] font-mono text-slate-400" :title="'IMEI: ' + v.gps_imei">
                  {{ v.gps_imei.slice(-6) }}
                </span>
                <span v-else class="text-[9px] font-bold text-amber-600 bg-amber-50 px-1.5 py-0.5 rounded">
                  Tanpa IMEI
                </span>
                <ChevronRight class="h-4 w-4 text-slate-300 group-hover:text-amber-500 transition-colors" />
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
:deep(.custom-trax-marker) {
  background: transparent !important;
  border: none !important;
}
</style>
