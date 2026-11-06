<script setup lang="ts">
import { ref, shallowRef, onMounted, onUnmounted, markRaw, toRaw, computed, watch } from 'vue'
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
import apiClient from '@/api/axios'
import {
  Truck,
  MapPin,
  Search,
  Navigation,
  CheckCircle2,
  AlertTriangle,
  RefreshCw,
  Phone,
  User,
  Layers,
  ChevronRight,
  Filter
} from 'lucide-vue-next'

interface Store {
  id: string
  name: string
  address: string
  latitude: number
  longitude: number
  assigned_employee?: {
    first_name: string
    last_name: string
  }
}

interface Warehouse {
  name: string
  address: string
  latitude: number
  longitude: number
}

interface StoreSequence {
  id: string
  name: string
  address: string
  latitude: number
  longitude: number
  sequence: number
  status: string
  delivered_at?: string
  is_target: boolean
}

interface TruckLocation {
  latitude: number
  longitude: number
  last_updated_at: string
  status_text: string
}

interface TrackingData {
  delivery_batch_id: string
  delivery_order_no: string
  status: string
  driver_name: string
  driver_phone: string
  vehicle_plate: string
  vehicle_model: string
  warehouse: Warehouse
  stores: StoreSequence[]
  truck_location: TruckLocation
}

const stores = ref<Store[]>([])
const searchQuery = ref('')
const selectedStoreId = ref('')
const isTracking = ref(false)
const isLoadingStores = ref(false)
const isLoadingTracking = ref(false)
const trackingError = ref('')
const trackingData = ref<TrackingData | null>(null)

// Cascading filters reactive states
const selectedSalesman = ref('')
const selectedProvince = ref('')
const selectedRegency = ref('')
const selectedDistrict = ref('')

// Leaflet Map Refs
const map = shallowRef<L.Map | null>(null)
const markers = shallowRef<L.Marker[]>([])
const routePolyline = shallowRef<L.Polyline | null>(null)

// Auto-refresh interval for real-time updates
let refreshInterval: number | null = null

// Address and Salesman parser helper
const parseStoreDetails = (store: Store) => {
  const address = store.address || ''
  
  // 1. Salesman name
  let salesman = 'Tanpa Salesman'
  if (store.assigned_employee) {
    salesman = `${store.assigned_employee.first_name} ${store.assigned_employee.last_name || ''}`.trim()
  }
  
  // 2. Location parsing (Provinsi, Kabupaten, Kecamatan)
  let province = ''
  let regency = ''
  let district = ''
  
  // Regex to match Indonesian location keywords
  const provMatch = address.match(/(?:provinsi|prov\.?)\s+([^,]+)/i)
  if (provMatch && provMatch[1]) province = provMatch[1].trim()
  
  const kabMatch = address.match(/(?:kabupaten|kab\.?|kota)\s+([^,]+)/i)
  if (kabMatch && kabMatch[1]) regency = kabMatch[1].trim()
  
  const kecMatch = address.match(/(?:kecamatan|kec\.?)\s+([^,]+)/i)
  if (kecMatch && kecMatch[1]) district = kecMatch[1].trim()
  
  // Comma split fallbacks if not matched by keywords
  const parts = address.split(',').map(p => p.trim())
  if (!province && parts.length > 0) {
    province = parts[parts.length - 1] || ''
  }
  if (!regency && parts.length > 1) {
    regency = parts[parts.length - 2] || ''
  }
  if (!district && parts.length > 2) {
    district = parts[parts.length - 3] || ''
  }
  
  // Standardize default fallbacks
  province = province || 'Lainnya'
  regency = regency || 'Lainnya'
  district = district || 'Lainnya'
  
  // Clean prefixes for cleaner dropdown items
  const cleanPrefixes = (s: string) => {
    return s.replace(/^(provinsi|prov\.?|kabupaten|kab\.?|kota|kecamatan|kec\.?)\s+/i, '').trim()
  }
  
  return {
    salesman,
    province: cleanPrefixes(province),
    regency: cleanPrefixes(regency),
    district: cleanPrefixes(district)
  }
}

// Parsed stores list
const parsedStores = computed(() => {
  return stores.value.map(store => {
    const parsed = parseStoreDetails(store)
    return {
      ...store,
      parsed
    }
  })
})

// Options for dropdown filters
const salesmanOptions = computed(() => {
  const set = new Set(parsedStores.value.map(s => s.parsed.salesman))
  return Array.from(set).sort()
})

const provinceOptions = computed(() => {
  let filtered = parsedStores.value
  if (selectedSalesman.value) {
    filtered = filtered.filter(s => s.parsed.salesman === selectedSalesman.value)
  }
  const set = new Set(filtered.map(s => s.parsed.province))
  return Array.from(set).sort()
})

const regencyOptions = computed(() => {
  let filtered = parsedStores.value
  if (selectedSalesman.value) {
    filtered = filtered.filter(s => s.parsed.salesman === selectedSalesman.value)
  }
  if (selectedProvince.value) {
    filtered = filtered.filter(s => s.parsed.province === selectedProvince.value)
  }
  const set = new Set(filtered.map(s => s.parsed.regency))
  return Array.from(set).sort()
})

const districtOptions = computed(() => {
  let filtered = parsedStores.value
  if (selectedSalesman.value) {
    filtered = filtered.filter(s => s.parsed.salesman === selectedSalesman.value)
  }
  if (selectedProvince.value) {
    filtered = filtered.filter(s => s.parsed.province === selectedProvince.value)
  }
  if (selectedRegency.value) {
    filtered = filtered.filter(s => s.parsed.regency === selectedRegency.value)
  }
  const set = new Set(filtered.map(s => s.parsed.district))
  return Array.from(set).sort()
})

// Cascading filters reset watches
watch(selectedSalesman, () => {
  selectedProvince.value = ''
  selectedRegency.value = ''
  selectedDistrict.value = ''
  selectedStoreId.value = ''
})

watch(selectedProvince, () => {
  selectedRegency.value = ''
  selectedDistrict.value = ''
  selectedStoreId.value = ''
})

watch(selectedRegency, () => {
  selectedDistrict.value = ''
  selectedStoreId.value = ''
})

watch(selectedDistrict, () => {
  selectedStoreId.value = ''
})

const filteredStores = computed(() => {
  let result = parsedStores.value
  
  if (selectedSalesman.value) {
    result = result.filter(s => s.parsed.salesman === selectedSalesman.value)
  }
  if (selectedProvince.value) {
    result = result.filter(s => s.parsed.province === selectedProvince.value)
  }
  if (selectedRegency.value) {
    result = result.filter(s => s.parsed.regency === selectedRegency.value)
  }
  if (selectedDistrict.value) {
    result = result.filter(s => s.parsed.district === selectedDistrict.value)
  }
  
  if (searchQuery.value) {
    const q = searchQuery.value.toLowerCase()
    result = result.filter(s => s.name.toLowerCase().includes(q) || s.address.toLowerCase().includes(q))
  }
  
  return result
})

const fetchStores = async () => {
  isLoadingStores.value = true
  try {
    const res = await apiClient.get('/public/delivery/stores')
    if (res.data && res.data.data) {
      stores.value = res.data.data
    }
  } catch (err: any) {
    console.error('Failed to fetch public stores:', err)
  } finally {
    isLoadingStores.value = false
  }
}

const trackShipment = async (autoRefresh = false) => {
  if (!selectedStoreId.value) return
  
  if (!autoRefresh) {
    isLoadingTracking.value = true
    trackingError.value = ''
  }
  
  try {
    const res = await apiClient.get(`/public/delivery/track?store_id=${selectedStoreId.value}`)
    if (res.data && res.data.data) {
      trackingData.value = res.data.data
      isTracking.value = true
      
      // Update map display
      await updateMap()
      
      // Setup auto-refresh if shipment is active
      setupAutoRefresh()
    }
  } catch (err: any) {
    console.error('Failed to track shipment:', err)
    const errMsg = err.response?.data?.message || err.response?.data?.error || 'Tidak ada pengiriman aktif untuk toko ini'
    trackingError.value = errMsg
    isTracking.value = false
    if (!autoRefresh) {
      trackingData.value = null
    }
  } finally {
    if (!autoRefresh) {
      isLoadingTracking.value = false
    }
  }
}

const setupAutoRefresh = () => {
  if (refreshInterval) {
    clearInterval(refreshInterval)
    refreshInterval = null
  }
  
  const status = trackingData.value?.status
  if (status === 'ON_DELIVERY' || status === 'PICKING') {
    refreshInterval = window.setInterval(() => {
      trackShipment(true)
    }, 30000)
  }
}

const getStatusLabel = (status?: string) => {
  if (!status) return 'Unknown'
  switch (status) {
    case 'WAITING_APPROVAL':
      return 'Menunggu Persetujuan'
    case 'SUPERVISOR_APPROVED':
      return 'Surat Jalan Diterbitkan'
    case 'PICKING':
      return 'Driver Mengambil Barang'
    case 'ON_DELIVERY':
      return 'Dalam Perjalanan'
    case 'COMPLETED':
      return 'Selesai'
    default:
      return status
  }
}

const getStatusBadgeClass = (status?: string) => {
  if (!status) return 'bg-slate-100 text-slate-800'
  switch (status) {
    case 'ON_DELIVERY':
      return 'bg-blue-600 text-white animate-pulse'
    case 'COMPLETED':
      return 'bg-emerald-500 text-white'
    case 'PICKING':
      return 'bg-amber-500 text-white'
    case 'SUPERVISOR_APPROVED':
      return 'bg-indigo-600 text-white'
    default:
      return 'bg-slate-500 text-white'
  }
}

// Leaflet Map Functions
const initMap = () => {
  if (map.value) return

  const mapInstance = L.map('tracking-map', {
    zoomControl: true
  }).setView([-6.2088, 106.8456], 11)
  
  map.value = markRaw(mapInstance)

  L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
    attribution: '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors'
  }).addTo(toRaw(map.value))
}

const updateMap = async () => {
  if (!map.value || !trackingData.value) return

  const mapInstance = toRaw(map.value)

  // Clear existing markers
  markers.value.forEach((m) => m.remove())
  markers.value = []

  // Clear existing path polyline
  if (routePolyline.value) {
    routePolyline.value.remove()
    routePolyline.value = null
  }

  const coordinates: L.LatLngExpression[] = []
  const data = trackingData.value

  // 1. Warehouse / Starting Point Marker
  if (data.warehouse && data.warehouse.latitude && data.warehouse.longitude) {
    const whCoords: L.LatLngExpression = [data.warehouse.latitude, data.warehouse.longitude]
    coordinates.push(whCoords)

    const whIcon = L.divIcon({
      className: 'wh-marker-icon',
      html: `
        <div class="flex items-center justify-center w-8 h-8 rounded-full bg-blue-600 border-2 border-white shadow-lg text-white font-bold text-xs">
          🏢
        </div>
      `,
      iconSize: [32, 32],
      iconAnchor: [16, 16]
    })

    const whMarker = L.marker(whCoords, { icon: whIcon })
      .addTo(mapInstance)
      .bindPopup(`<b>${data.warehouse.name}</b><br/>${data.warehouse.address}`)
    
    markers.value.push(whMarker)
  }

  // 2. Stores Sequence Markers
  data.stores.forEach((store) => {
    if (!store.latitude || !store.longitude) return
    const storeCoords: L.LatLngExpression = [store.latitude, store.longitude]
    coordinates.push(storeCoords)

    let bgColor = 'bg-slate-400'
    let labelText = store.sequence.toString()
    let pingElement = ''

    if (store.status === 'DELIVERED') {
      bgColor = 'bg-emerald-500'
    } else if (store.status === 'FAILED') {
      bgColor = 'bg-rose-500'
    } else if (store.status === 'PENDING') {
      if (store.is_target) {
        bgColor = 'bg-amber-500 border-dashed border-amber-600 scale-110'
        pingElement = `<div class="absolute w-full h-full rounded-full bg-amber-400 opacity-75 animate-ping"></div>`
      } else {
        bgColor = 'bg-slate-400'
      }
    }

    const storeIcon = L.divIcon({
      className: 'store-marker-icon',
      html: `
        <div class="relative flex items-center justify-center w-8 h-8">
          ${pingElement}
          <div class="relative flex items-center justify-center w-7 h-7 rounded-full ${bgColor} border-2 border-white shadow-md text-white font-bold text-xs">
            ${labelText}
          </div>
        </div>
      `,
      iconSize: [32, 32],
      iconAnchor: [16, 16]
    })

    const popupHtml = `
      <div class="p-1">
        <h4 class="font-bold text-slate-900 text-xs">${store.name}</h4>
        <p class="text-[10px] text-slate-500 mt-0.5">${store.address}</p>
        <div class="flex items-center gap-1.5 mt-2">
          <span class="text-[9px] px-1.5 py-0.5 rounded font-black text-white ${
            store.status === 'DELIVERED'
              ? 'bg-emerald-500'
              : store.status === 'FAILED'
              ? 'bg-rose-500'
              : 'bg-slate-500'
          }">${store.status}</span>
          <span class="text-[10px] font-bold text-slate-400">Urutan #${store.sequence}</span>
        </div>
      </div>
    `

    const storeMarker = L.marker(storeCoords, { icon: storeIcon })
      .addTo(mapInstance)
      .bindPopup(popupHtml)

    markers.value.push(storeMarker)
  })

  // 3. Truck Location Marker
  if (data.truck_location && data.truck_location.latitude && data.truck_location.longitude) {
    const truckCoords: L.LatLngExpression = [data.truck_location.latitude, data.truck_location.longitude]
    
    const isPulsing = data.status === 'ON_DELIVERY'
    const pulsingHtml = isPulsing 
      ? `<div class="absolute w-full h-full rounded-full bg-blue-400 opacity-60 animate-ping"></div>` 
      : ''

    const truckIcon = L.divIcon({
      className: 'truck-marker-icon',
      html: `
        <div class="relative flex items-center justify-center w-10 h-10">
          ${pulsingHtml}
          <div class="relative flex items-center justify-center w-8 h-8 rounded-full bg-slate-900 border-2 border-white shadow-xl text-white">
            🚚
          </div>
        </div>
      `,
      iconSize: [40, 40],
      iconAnchor: [20, 20]
    })

    const truckMarker = L.marker(truckCoords, { icon: truckIcon })
      .addTo(mapInstance)
      .bindPopup(`
        <div class="p-1">
          <div class="flex items-center gap-1.5">
            <span class="text-xs">🚚</span>
            <h4 class="font-black text-slate-900 text-xs">Lokasi Pengiriman</h4>
          </div>
          <p class="text-[10px] text-slate-600 mt-1 font-bold">${data.truck_location.status_text}</p>
          <p class="text-[9px] text-slate-400 mt-0.5">Diperbarui: ${new Date(data.truck_location.last_updated_at).toLocaleTimeString()}</p>
        </div>
      `)
    
    markers.value.push(truckMarker)
    
    mapInstance.panTo(truckCoords)
  }

  // 4. Draw Route Line
  if (coordinates.length > 1) {
    let routeCoords: L.LatLngExpression[] = coordinates
    try {
      const osrmCoords = coordinates.map((c: any) => `${c[1]},${c[0]}`).join(';')
      const response = await fetch(`https://router.project-osrm.org/route/v1/driving/${osrmCoords}?overview=full&geometries=geojson&steps=false`)
      if (response.ok) {
        const resData = await response.json()
        if (resData.code === 'Ok' && resData.routes && resData.routes.length > 0) {
          const geometry = resData.routes[0].geometry
          routeCoords = geometry.coordinates.map((c: any) => [c[1], c[0]])
        }
      }
    } catch (e) {
      console.error('Failed to fetch OSRM route, falling back to straight lines:', e)
    }

    const line = L.polyline(routeCoords, {
      color: '#2563eb',
      weight: 4,
      opacity: 0.8,
      dashArray: '8, 8',
      lineJoin: 'round'
    }).addTo(mapInstance)

    routePolyline.value = markRaw(line)

    mapInstance.fitBounds(line.getBounds(), {
      padding: [40, 40]
    })
  } else if (coordinates.length === 1) {
    const firstCoords = coordinates[0]
    if (firstCoords) {
      mapInstance.setView(firstCoords, 14)
    }
  }
}

const shareViaWhatsApp = () => {
  if (!selectedStoreId.value || !trackingData.value) return
  
  const storeName = trackingData.value.stores.find(s => s.is_target)?.name || 'Toko'
  const trackingUrl = `${window.location.origin}${window.location.pathname}?store_id=${selectedStoreId.value}`
  
  const text = `Halo, berikut adalah link pelacakan pengiriman untuk Toko *${storeName}* dengan Surat Jalan *#${trackingData.value.delivery_order_no}*: ${trackingUrl}`
  
  const encodedText = encodeURIComponent(text)
  const whatsappUrl = `https://api.whatsapp.com/send?text=${encodedText}`
  
  window.open(whatsappUrl, '_blank')
}

onMounted(() => {
  fetchStores()
  initMap()
  
  // Auto-load store tracking from URL query parameter
  const urlParams = new URLSearchParams(window.location.search)
  const storeIdParam = urlParams.get('store_id')
  if (storeIdParam) {
    selectedStoreId.value = storeIdParam
  }
})

onUnmounted(() => {
  if (refreshInterval) {
    clearInterval(refreshInterval)
  }
})

watch(selectedStoreId, (newId) => {
  if (newId) {
    trackShipment()
  } else {
    isTracking.value = false
    trackingData.value = null
  }
})
</script>

<template>
  <div class="min-h-screen bg-slate-50 text-slate-700 flex flex-col font-sans">
    <!-- Header -->
    <header class="bg-white border-b border-slate-200 py-4 px-6 sticky top-0 z-50 shadow-sm">
      <div class="max-w-7xl mx-auto flex justify-between items-center">
        <div class="flex items-center gap-3">
          <div class="w-10 h-10 rounded-xl bg-green-600 flex items-center justify-center text-white shadow-md shadow-blue-500/10">
            <Truck class="w-5 h-5" />
          </div>
          <div>
            <h1 class="text-base font-black tracking-tight text-slate-900 flex items-center gap-1">
              KARSA HR <span class="text-green-600 font-medium">Pengiriman</span>
            </h1>
            <p class="text-[10px] text-slate-500 font-bold uppercase tracking-wider">Layanan Untuk Pelanggan Pelacakan Pengiriman</p>
          </div>
        </div>
        
        <div class="flex items-center gap-2 text-xs text-slate-400">
          <span class="w-2 h-2 rounded-full bg-emerald-500 animate-pulse"></span>
          <span class="font-bold text-slate-500 uppercase tracking-widest text-[9px]">Sistem Online</span>
        </div>
      </div>
    </header>

    <!-- Main Container -->
    <main class="flex-1 max-w-7xl w-full mx-auto p-4 md:p-6 grid grid-cols-1 lg:grid-cols-12 gap-6 min-h-0">
      <!-- Left Column (Cascading Filters & Timeline details) -->
      <section class="lg:col-span-5 flex flex-col gap-6 overflow-y-auto max-h-[85vh] pr-1">
        
        <!-- Cascading Filter Card -->
        <div class="bg-white border border-slate-200 rounded-3xl p-5 shadow-sm flex flex-col gap-4">
          <div class="flex items-center gap-2 pb-3 border-b border-slate-100">
            <div class="w-6 h-6 rounded-lg bg-blue-50 flex items-center justify-center text-blue-600">
              <Filter class="w-3.5 h-3.5" />
            </div>
            <h3 class="text-xs font-black text-slate-800 uppercase tracking-wider">Cari & Filter Toko</h3>
          </div>
          
          <div class="grid grid-cols-1 md:grid-cols-2 gap-3">
            <!-- 1. Salesman Dropdown -->
            <div class="flex flex-col gap-1">
              <label class="text-[9px] font-black text-slate-400 uppercase tracking-widest">Filter Salesman</label>
              <select
                v-model="selectedSalesman"
                class="w-full px-3 py-2 text-xs bg-slate-50 border border-slate-200 rounded-xl focus:outline-none focus:border-blue-500 transition text-slate-700"
              >
                <option value="">Semua Salesman</option>
                <option v-for="opt in salesmanOptions" :key="opt" :value="opt">{{ opt }}</option>
              </select>
            </div>

            <!-- 2. Provinsi Dropdown -->
            <div class="flex flex-col gap-1">
              <label class="text-[9px] font-black text-slate-400 uppercase tracking-widest">Provinsi</label>
              <select
                v-model="selectedProvince"
                class="w-full px-3 py-2 text-xs bg-slate-50 border border-slate-200 rounded-xl focus:outline-none focus:border-blue-500 transition text-slate-700"
              >
                <option value="">Semua Provinsi</option>
                <option v-for="opt in provinceOptions" :key="opt" :value="opt">{{ opt }}</option>
              </select>
            </div>

            <!-- 3. Kabupaten Dropdown -->
            <div class="flex flex-col gap-1">
              <label class="text-[9px] font-black text-slate-400 uppercase tracking-widest">Kabupaten/Kota</label>
              <select
                v-model="selectedRegency"
                class="w-full px-3 py-2 text-xs bg-slate-50 border border-slate-200 rounded-xl focus:outline-none focus:border-blue-500 transition text-slate-700"
                :disabled="!selectedProvince"
              >
                <option value="">Semua Kabupaten/Kota</option>
                <option v-for="opt in regencyOptions" :key="opt" :value="opt">{{ opt }}</option>
              </select>
            </div>

            <!-- 4. Kecamatan Dropdown -->
            <div class="flex flex-col gap-1">
              <label class="text-[9px] font-black text-slate-400 uppercase tracking-widest">Kecamatan</label>
              <select
                v-model="selectedDistrict"
                class="w-full px-3 py-2 text-xs bg-slate-50 border border-slate-200 rounded-xl focus:outline-none focus:border-blue-500 transition text-slate-700"
                :disabled="!selectedRegency"
              >
                <option value="">Semua Kecamatan</option>
                <option v-for="opt in districtOptions" :key="opt" :value="opt">{{ opt }}</option>
              </select>
            </div>
          </div>

          <!-- 5. Search Store name input -->
          <div class="flex flex-col gap-1">
            <label class="text-[9px] font-black text-slate-400 uppercase tracking-widest">Cari Nama / Alamat Toko</label>
            <div class="relative w-full">
              <Search class="absolute left-3 top-2.5 w-3.5 h-3.5 text-slate-400" />
              <input
                v-model="searchQuery"
                type="text"
                placeholder="Ketik nama toko untuk mencari..."
                class="w-full pl-9 pr-4 py-2 text-xs bg-slate-50 border border-slate-200 rounded-xl focus:outline-none focus:border-blue-500 transition text-slate-700"
              />
            </div>
          </div>

          <!-- 6. Store Selector Dropdown -->
          <div class="flex flex-col gap-1 pt-2 border-t border-slate-100">
            <label class="text-[10px] font-black text-blue-600 uppercase tracking-wider">Pilih Toko untuk Lacak ({{ filteredStores.length }} Toko Cocok)</label>
            <select
              v-model="selectedStoreId"
              class="w-full px-4 py-2.5 text-xs bg-blue-50/30 border border-blue-200 font-bold rounded-xl focus:outline-none focus:border-blue-500 transition text-slate-800"
            >
              <option value="" selected>-- Pilih Toko yang Akan Dilacak --</option>
              <option v-for="s in filteredStores" :key="s.id" :value="s.id">
                {{ s.name }} ({{ s.parsed.regency }}, {{ s.parsed.province }})
              </option>
            </select>
          </div>
        </div>

        <!-- Initial Prompt / Default Screen -->
        <div v-if="!isTracking && !isLoadingTracking && !trackingError" class="bg-white border border-slate-200 rounded-3xl p-8 text-center flex flex-col items-center justify-center my-auto min-h-[300px] shadow-sm">
          <div class="w-16 h-16 rounded-full bg-blue-50 flex items-center justify-center text-blue-600 mb-4 animate-bounce">
            <Navigation class="w-8 h-8" />
          </div>
          <h3 class="text-base font-bold text-slate-900 mb-2">Lacak Pengiriman Anda</h3>
          <p class="text-xs text-slate-500 max-w-sm leading-relaxed mb-6">
            Gunakan filter di atas untuk menemukan toko Anda dan klik dropdown Toko untuk melacak surat jalan serta posisi armada driver pengiriman secara instan.
          </p>
          <div class="text-[10px] text-slate-400 font-bold bg-slate-50 border border-slate-100 px-4 py-2 rounded-xl">
            Sistem otomatis memperbarui peta setiap 30 detik
          </div>
        </div>

        <!-- Loading State Container -->
        <div v-if="isLoadingTracking" class="bg-white border border-slate-200 rounded-3xl p-12 text-center flex flex-col items-center justify-center min-h-[300px] shadow-sm">
          <RefreshCw class="w-8 h-8 text-blue-600 animate-spin mb-4" />
          <h4 class="text-xs font-bold text-slate-800">Menghubungkan ke Server...</h4>
          <p class="text-[11px] text-slate-400 mt-2">Sedang mengambil status surat jalan terbaru</p>
        </div>

        <!-- Tracking Error Message Container (Elegant Handle for 404) -->
        <div v-if="trackingError" class="bg-red-50/50 border border-red-100 rounded-3xl p-6 text-center flex flex-col items-center justify-center min-h-[300px] shadow-sm">
          <div class="w-12 h-12 rounded-full bg-red-100 flex items-center justify-center text-red-600 mb-3 animate-pulse">
            <AlertTriangle class="w-5 h-5" />
          </div>
          <h4 class="text-sm font-bold text-red-800">Tidak Ada Pengiriman Aktif</h4>
          <p class="text-xs text-red-600 mt-2 max-w-xs leading-relaxed">{{ trackingError }}</p>
          <div class="mt-5 text-[10px] text-slate-500 font-medium bg-white border border-slate-100 p-4 rounded-xl shadow-xs">
            Toko terpilih belum terdaftar dalam jadwal pengiriman driver hari ini, atau proses pengiriman telah selesai. Silakan hubungi Sales Agent Anda untuk informasi lebih lanjut.
          </div>
        </div>

        <!-- Real-time Tracking Details Display -->
        <div v-if="isTracking && trackingData" class="flex flex-col gap-6">
          
          <!-- Route header details -->
          <div class="bg-white border border-slate-200 rounded-3xl p-5 shadow-sm relative overflow-hidden">
            <div class="flex justify-between items-start gap-4">
              <div>
                <span class="text-[9px] font-black text-slate-400 uppercase tracking-widest">Nomor Surat Jalan</span>
                <h2 class="text-base font-black text-slate-900 mt-0.5">#{{ trackingData.delivery_order_no }}</h2>
              </div>
              <span :class="['px-3 py-1.5 rounded-xl text-[10px] font-black uppercase tracking-wider shadow-sm', getStatusBadgeClass(trackingData.status)]">
                {{ getStatusLabel(trackingData.status) }}
              </span>
            </div>

            <!-- Operational progress text -->
            <div class="mt-4 bg-blue-50/30 border border-blue-100 rounded-2xl p-3.5 flex items-start gap-3">
              <div class="w-8 h-8 rounded-lg bg-blue-100 flex items-center justify-center text-blue-600 flex-shrink-0">
                <Navigation class="w-4 h-4 animate-pulse" />
              </div>
              <div>
                <p class="text-[9px] font-bold text-blue-500 uppercase tracking-wider">Status Armada</p>
                <p class="text-xs text-slate-800 mt-0.5 font-bold leading-relaxed">{{ trackingData.truck_location.status_text }}</p>
                <p class="text-[9px] text-slate-400 mt-1">Status diperbarui: {{ new Date(trackingData.truck_location.last_updated_at).toLocaleTimeString() }}</p>
              </div>
            </div>

            <!-- Driver Card details -->
            <div class="mt-4 pt-4 border-t border-slate-100 grid grid-cols-2 gap-4">
              <div class="flex items-center gap-3">
                <div class="w-9 h-9 rounded-xl bg-slate-100 flex items-center justify-center text-slate-500">
                  <User class="w-4 h-4 text-blue-600" />
                </div>
                <div>
                  <p class="text-[9px] text-slate-400 font-bold uppercase">Driver</p>
                  <p class="text-xs font-bold text-slate-800 mt-0.5 truncate max-w-[120px]">{{ trackingData.driver_name }}</p>
                  <p class="text-[10px] text-slate-500 flex items-center gap-0.5" v-if="trackingData.driver_phone && trackingData.driver_phone !== '-'">
                    <Phone class="w-2.5 h-2.5" /> {{ trackingData.driver_phone }}
                  </p>
                </div>
              </div>

              <div class="flex items-center gap-3">
                <div class="w-9 h-9 rounded-xl bg-slate-100 flex items-center justify-center text-slate-500">
                  <Truck class="w-4 h-4 text-blue-600" />
                </div>
                <div>
                  <p class="text-[9px] text-slate-400 font-bold uppercase">Armada</p>
                  <p class="text-xs font-bold text-slate-800 mt-0.5">{{ trackingData.vehicle_plate }}</p>
                  <p class="text-[10px] text-slate-500 truncate max-w-[120px]">{{ trackingData.vehicle_model }}</p>
                </div>
              </div>
            </div>

            <!-- WhatsApp Share Button -->
            <button
              @click="shareViaWhatsApp"
              class="w-full mt-4 py-2.5 px-4 bg-emerald-600 hover:bg-emerald-700 active:scale-[0.98] text-white text-xs font-bold rounded-xl shadow-md shadow-emerald-600/10 flex items-center justify-center gap-2 transition-all duration-300"
            >
              <svg class="w-4 h-4 fill-current" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
                <path d="M.057 24l1.687-6.163c-1.041-1.804-1.588-3.849-1.587-5.946C.06 5.348 5.397.01 12.008.01c3.202.001 6.212 1.246 8.477 3.514 2.266 2.268 3.507 5.28 3.505 8.484-.004 6.657-5.34 11.997-11.953 11.997-2.005-.001-3.973-.502-5.724-1.455L0 24zm6.59-4.846c1.6.95 3.188 1.449 4.825 1.451 5.436 0 9.86-4.37 9.864-9.799.002-2.63-1.023-5.101-2.885-6.965C16.528 2.025 14.07 1 11.69 1 6.251 1 1.826 5.37 1.822 10.799c-.001 1.702.459 3.361 1.332 4.877L2.146 21.05l5.501-1.896zm12.353-6.426c-.328-.164-1.942-.959-2.242-1.069-.3-.11-.518-.164-.736.164-.218.328-.846 1.069-1.037 1.288-.19.219-.382.246-.71.082-.328-.164-1.386-.511-2.64-1.63-1-.893-1.676-2-1.874-2.328-.197-.329-.022-.507.142-.671.147-.148.328-.383.492-.574.164-.19.218-.328.328-.547.11-.219.055-.411-.027-.574-.082-.164-.736-1.777-1.01-2.434-.268-.646-.536-.558-.736-.569-.19-.01-.41-.01-.628-.01s-.573.082-.873.411c-.3.328-1.147 1.122-1.147 2.735 0 1.614 1.174 3.176 1.338 3.395.164.22 2.313 3.532 5.6 4.954.783.339 1.395.541 1.874.693.786.25 1.5.215 2.065.13.63-.095 1.943-.794 2.216-1.56.273-.767.273-1.424.19-1.56-.081-.136-.3-.218-.627-.382z"/>
              </svg>
              Bagikan Rute via WhatsApp
            </button>
          </div>

          <!-- Stops Timeline details -->
          <div class="bg-white border border-slate-200 rounded-3xl p-5 shadow-sm">
            <h3 class="text-xs font-black text-slate-800 uppercase tracking-wider mb-4 flex items-center gap-2">
              <Layers class="w-4 h-4 text-blue-600" /> Detail Rute Pengantaran (SJ)
            </h3>
            
            <div class="relative border-l border-slate-200 ml-4 pl-6 space-y-6">
              
              <!-- Warehouse start stop -->
              <div class="relative">
                <div class="absolute -left-[35px] top-0 w-6 h-6 rounded-full bg-blue-600 border-4 border-white shadow-sm flex items-center justify-center text-[10px]">
                  🏢
                </div>
                <div>
                  <div class="flex items-center gap-2">
                    <h4 class="text-xs font-black text-slate-800">{{ trackingData.warehouse.name }}</h4>
                    <span class="px-1.5 py-0.5 rounded text-[8px] font-black bg-blue-50 border border-blue-200 text-blue-600 uppercase">Gudang</span>
                  </div>
                  <p class="text-[10px] text-slate-400 mt-0.5">{{ trackingData.warehouse.address }}</p>
                </div>
              </div>

              <!-- Loop stop cards -->
              <div
                v-for="store in trackingData.stores"
                :key="store.id"
                :class="['relative transition-all', store.is_target ? 'bg-blue-50/20 border border-blue-100 rounded-2xl p-3 -mx-3 shadow-xs' : '']"
              >
                <!-- Timeline status dot -->
                <div
                  :class="[
                    'absolute top-0 w-6 h-6 rounded-full border-4 border-white shadow-sm flex items-center justify-center font-black text-[9px]',
                    store.is_target ? '-left-[47px]' : '-left-[35px]',
                    store.status === 'DELIVERED'
                      ? 'bg-emerald-500 text-white'
                      : store.status === 'FAILED'
                      ? 'bg-rose-500 text-white'
                      : 'bg-slate-300 text-slate-600'
                  ]"
                >
                  <span v-if="store.status === 'DELIVERED'">✓</span>
                  <span v-else-if="store.status === 'FAILED'">✗</span>
                  <span v-else>{{ store.sequence }}</span>
                </div>

                <div>
                  <div class="flex flex-wrap items-center gap-2">
                    <h4 :class="['text-xs font-black', store.is_target ? 'text-blue-600' : 'text-slate-800']">
                      {{ store.name }}
                    </h4>
                    <span
                      v-if="store.is_target"
                      class="px-1.5 py-0.5 rounded text-[8px] font-black bg-amber-50 border border-amber-200 text-amber-600 uppercase tracking-wider"
                    >
                      Toko Anda
                    </span>
                    <span
                      :class="[
                        'px-1.5 py-0.5 rounded text-[8px] font-black uppercase tracking-wider',
                        store.status === 'DELIVERED'
                          ? 'bg-emerald-50 border border-emerald-200 text-emerald-600'
                          : store.status === 'FAILED'
                          ? 'bg-rose-50 border border-rose-200 text-rose-600'
                          : 'bg-slate-100 border border-slate-200 text-slate-500'
                      ]"
                    >
                      {{ store.status }}
                    </span>
                  </div>
                  <p class="text-[10px] text-slate-400 mt-1 leading-normal">{{ store.address }}</p>
                  
                  <p v-if="store.delivered_at" class="text-[9px] text-emerald-600 mt-1 flex items-center gap-1 font-bold">
                    <CheckCircle2 class="w-3 h-3 text-emerald-500" />
                    Diterima pada {{ new Date(store.delivered_at).toLocaleTimeString() }}
                  </p>
                </div>
              </div>
              
            </div>
          </div>
        </div>
      </section>

      <!-- Right Column (Leaflet Maps) -->
      <section class="lg:col-span-7 bg-white border border-slate-200 rounded-3xl overflow-hidden shadow-sm h-[50vh] lg:h-[80vh] relative min-h-[350px]">
        <div id="tracking-map" class="w-full h-full z-10"></div>
        
        <div v-if="isLoadingTracking" class="absolute inset-0 bg-white/80 z-20 flex items-center justify-center backdrop-blur-xs">
          <div class="flex flex-col items-center">
            <RefreshCw class="w-8 h-8 text-blue-600 animate-spin mb-3" />
            <p class="text-xs font-bold text-slate-500">Menyinkronkan data peta...</p>
          </div>
        </div>
      </section>
    </main>

    <!-- Footer -->
    <footer class="bg-white border-t border-slate-200 py-4 px-6 text-center text-xs text-slate-400 font-bold mt-auto shadow-xs">
      <div class="max-w-7xl mx-auto flex flex-col sm:flex-row justify-between items-center gap-3">
        <p>&copy; {{ new Date().getFullYear() }} Perusahaan Demo. Hak Cipta Dilindungi.</p>
        <div class="flex items-center gap-4 text-slate-400">
          <a href="#" class="hover:text-blue-600 transition">Panduan Aplikasi</a>
          <span class="text-slate-200">|</span>
          <a href="#" class="hover:text-blue-600 transition">Hubungi Kami</a>
        </div>
      </div>
    </footer>
  </div>
</template>

<style scoped>
:deep(.leaflet-container) {
  background-color: #f8fafc;
}

/* Customize Leaflet elements to fit bright professional theme */
:deep(.leaflet-bar) {
  border: 1px solid #e2e8f0 !important;
  box-shadow: 0 4px 6px -1px rgb(0 0 0 / 0.05) !important;
}
:deep(.leaflet-bar a) {
  background-color: #ffffff !important;
  color: #475569 !important;
  border-bottom: 1px solid #e2e8f0 !important;
}
:deep(.leaflet-bar a:hover) {
  background-color: #f8fafc !important;
  color: #1e293b !important;
}
:deep(.leaflet-popup-content-wrapper) {
  background-color: #ffffff !important;
  border: 1px solid #e2e8f0 !important;
  color: #334155 !important;
  border-radius: 1rem !important;
  box-shadow: 0 10px 15px -3px rgb(0 0 0 / 0.05) !important;
}
:deep(.leaflet-popup-tip) {
  background-color: #ffffff !important;
  border-left: 1px solid #e2e8f0 !important;
  border-bottom: 1px solid #e2e8f0 !important;
}
</style>
