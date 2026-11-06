<script setup lang="ts">
import { ref, computed, watch, nextTick } from 'vue'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
} from '@/components/ui/dialog'
import { Badge } from '@/components/ui/badge'
import { Store, User, MapPin, MapPinOff, Navigation, Info, ExternalLink } from 'lucide-vue-next'
import { getDistance } from '@/utils/geolocation'
import L from 'leaflet'
import 'leaflet/dist/leaflet.css'
import 'leaflet-defaulticon-compatibility/dist/leaflet-defaulticon-compatibility.css'
import 'leaflet-defaulticon-compatibility'

const props = defineProps<{
  open: boolean
  orderData: any
}>()

const emit = defineEmits(['update:open'])

const mapContainer = ref<HTMLElement | null>(null)
let map: L.Map | null = null

const resolveCoord = (...vals: any[]): number | null => {
  for (const v of vals) {
    if (v !== undefined && v !== null && v !== '' && !isNaN(Number(v))) {
      const num = Number(v)
      if (num !== 0) return num
    }
  }
  return null
}

const salesLat = computed(() => {
  const d = props.orderData
  return resolveCoord(
    d?.latitude,
    d?.visit?.latitude,
    d?.raw?.latitude,
    d?.raw?.visit?.latitude,
    d?.sales_latitude
  )
})

const salesLng = computed(() => {
  const d = props.orderData
  return resolveCoord(
    d?.longitude,
    d?.visit?.longitude,
    d?.raw?.longitude,
    d?.raw?.visit?.longitude,
    d?.sales_longitude
  )
})

const storeLat = computed(() => {
  const d = props.orderData
  return resolveCoord(
    d?.store?.latitude,
    d?.visit?.store?.latitude,
    d?.raw?.store?.latitude,
    d?.raw?.visit?.store?.latitude,
    d?.store_latitude
  )
})

const storeLng = computed(() => {
  const d = props.orderData
  return resolveCoord(
    d?.store?.longitude,
    d?.visit?.store?.longitude,
    d?.raw?.store?.longitude,
    d?.raw?.visit?.store?.longitude,
    d?.store_longitude
  )
})

const hasValidCoordinates = computed(() => {
  return (
    salesLat.value !== null &&
    salesLng.value !== null &&
    storeLat.value !== null &&
    storeLng.value !== null
  )
})

const distanceMeters = computed(() => {
  if (!hasValidCoordinates.value || salesLat.value === null || salesLng.value === null || storeLat.value === null || storeLng.value === null) {
    return null
  }
  return getDistance(salesLat.value, salesLng.value, storeLat.value, storeLng.value)
})

const isOutsideRadius = computed(() => {
  if (distanceMeters.value === null) return false
  return distanceMeters.value > 500
})

const formattedDistance = computed(() => {
  if (distanceMeters.value === null) return '-'
  if (distanceMeters.value >= 1000) {
    return `${(distanceMeters.value / 1000).toFixed(2)} km`
  }
  return `${Math.round(distanceMeters.value)} meter`
})

const docNumber = computed(() => {
  const d = props.orderData
  return d?.receipt_no || d?.raw?.receipt_no || d?.so_number || d?.raw?.so_number || d?.id?.slice(0, 8)?.toUpperCase() || '-'
})

const storeName = computed(() => {
  const d = props.orderData
  return d?.store?.name || d?.raw?.store?.name || d?.store_name || d?.store || 'Toko'
})

const storeAddress = computed(() => {
  const d = props.orderData
  return d?.store?.address || d?.raw?.store?.address || d?.store_address || '-'
})

const openGoogleStreetView = (lat?: number | null, lng?: number | null) => {
  if (!lat || !lng) return
  window.open(`https://www.google.com/maps/@?api=1&map_action=pano&viewpoint=${lat},${lng}`, '_blank')
}

const openGoogleMaps = (lat?: number | null, lng?: number | null, queryName?: string) => {
  if (!lat || !lng) return
  const query = queryName ? encodeURIComponent(`${lat},${lng} (${queryName})`) : `${lat},${lng}`
  window.open(`https://www.google.com/maps/search/?api=1&query=${query}`, '_blank')
}

const salesmanName = computed(() => {
  const d = props.orderData
  const emp = d?.employee || d?.raw?.employee
  if (emp) {
    const fn = emp.first_name || ''
    const ln = emp.last_name || ''
    const name = `${fn} ${ln}`.trim()
    if (name) return name
  }
  return d?.salesman || 'Salesman'
})

const initMap = () => {
  if (!mapContainer.value || !props.orderData) return

  if (map) {
    try {
      map.remove()
    } catch (e) {
      console.warn('Leaflet map remove warning:', e)
    }
    map = null
  }

  // Clear any existing _leaflet_id to avoid "Map container is already initialized" error
  if (mapContainer.value && (mapContainer.value as any)._leaflet_id) {
    delete (mapContainer.value as any)._leaflet_id
  }

  if (!hasValidCoordinates.value || salesLat.value === null || salesLng.value === null || storeLat.value === null || storeLng.value === null) {
    return
  }

  const sLat = salesLat.value
  const sLng = salesLng.value
  const stLat = storeLat.value
  const stLng = storeLng.value

  map = L.map(mapContainer.value, {
    zoomControl: true,
  })

  L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
    attribution: '© OpenStreetMap contributors',
    maxZoom: 19,
  }).addTo(map)

  // Custom Icon: Store (Blue / Indigo Theme)
  const storeIcon = L.divIcon({
    className: 'custom-leaflet-marker-store',
    html: `
      <div style="position: relative; display: flex; flex-direction: column; align-items: center; cursor: pointer;">
        <div style="background: #2563eb; color: white; padding: 6px; border-radius: 12px; box-shadow: 0 4px 12px rgba(37, 99, 235, 0.4); border: 2px solid white; display: flex; align-items: center; justify-content: center; width: 36px; height: 36px;">
          <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><path d="m2 7 4.41-4.41A2 2 0 0 1 7.83 2h8.34a2 2 0 0 1 1.42.59L22 7"/><path d="M4 12v8a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2v-8"/><path d="M15 22v-4a2 2 0 0 0-2-2h-2a2 2 0 0 0-2 2v4"/><path d="M2 7h20"/><path d="M22 7v3a2 2 0 0 1-2 2v0a2.7 2.7 0 0 1-1.59-.63.7.7 0 0 0-.82 0A2.7 2.7 0 0 1 16 12a2.7 2.7 0 0 1-1.59-.63.7.7 0 0 0-.82 0A2.7 2.7 0 0 1 12 12a2.7 2.7 0 0 1-1.59-.63.7.7 0 0 0-.82 0A2.7 2.7 0 0 1 8 12a2.7 2.7 0 0 1-1.59-.63.7.7 0 0 0-.82 0A2.7 2.7 0 0 1 4 12v0a2 2 0 0 1-2-2V7"/></svg>
        </div>
        <div style="background: #1e3a8a; color: white; font-size: 9px; font-weight: 800; padding: 2px 6px; border-radius: 9999px; margin-top: -4px; border: 1px solid white; letter-spacing: 0.5px; text-transform: uppercase; white-space: nowrap; box-shadow: 0 2px 4px rgba(0,0,0,0.2);">
          🏢 Toko
        </div>
      </div>
    `,
    iconSize: [60, 50],
    iconAnchor: [30, 42],
    popupAnchor: [0, -42],
  })

  // Custom Icon: Salesman (Rose / Red Theme)
  const salesIcon = L.divIcon({
    className: 'custom-leaflet-marker-sales',
    html: `
      <div style="position: relative; display: flex; flex-direction: column; align-items: center; cursor: pointer;">
        <div style="background: #e11d48; color: white; padding: 6px; border-radius: 12px; box-shadow: 0 4px 12px rgba(225, 29, 72, 0.45); border: 2px solid white; display: flex; align-items: center; justify-content: center; width: 36px; height: 36px; animation: bounce-subtle 2s infinite;">
          <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><path d="M19 21v-2a4 4 0 0 0-4-4H9a4 4 0 0 0-4 4v2"/><circle cx="12" cy="7" r="4"/></svg>
        </div>
        <div style="background: #881337; color: white; font-size: 9px; font-weight: 800; padding: 2px 6px; border-radius: 9999px; margin-top: -4px; border: 1px solid white; letter-spacing: 0.5px; text-transform: uppercase; white-space: nowrap; box-shadow: 0 2px 4px rgba(0,0,0,0.2);">
          👤 Sales
        </div>
      </div>
    `,
    iconSize: [60, 50],
    iconAnchor: [30, 42],
    popupAnchor: [0, -42],
  })

  // 1. Radius Circle (500m tolerance around Store)
  L.circle([stLat, stLng], {
    radius: 500,
    color: '#3b82f6',
    fillColor: '#60a5fa',
    fillOpacity: 0.15,
    weight: 1.5,
    dashArray: '5, 5',
  }).addTo(map)

  // 2. Store Marker
  const storeMarker = L.marker([stLat, stLng], { icon: storeIcon }).addTo(map)
  storeMarker.bindPopup(`
    <div style="font-family: sans-serif; min-width: 180px; padding: 2px;">
      <div style="display: inline-block; background: #dbeafe; color: #1e40af; font-size: 10px; font-weight: 800; padding: 2px 6px; border-radius: 4px; text-transform: uppercase; margin-bottom: 4px;">
        📍 Master Lokasi Toko
      </div>
      <div style="font-size: 13px; font-weight: 700; color: #0f172a; line-height: 1.2;">${storeName.value}</div>
      <div style="font-size: 11px; color: #64748b; margin-top: 4px;">${storeAddress.value}</div>
      <div style="font-size: 10px; color: #94a3b8; font-family: monospace; margin-top: 4px;">${stLat.toFixed(6)}, ${stLng.toFixed(6)}</div>
    </div>
  `)

  // 3. Sales Marker
  const salesMarker = L.marker([sLat, sLng], { icon: salesIcon }).addTo(map)
  salesMarker.bindPopup(`
    <div style="font-family: sans-serif; min-width: 180px; padding: 2px;">
      <div style="display: inline-block; background: #ffe4e6; color: #9f1239; font-size: 10px; font-weight: 800; padding: 2px 6px; border-radius: 4px; text-transform: uppercase; margin-bottom: 4px;">
        👤 Titik Input Salesman
      </div>
      <div style="font-size: 13px; font-weight: 700; color: #0f172a; line-height: 1.2;">${salesmanName.value}</div>
      <div style="font-size: 11px; color: #64748b; margin-top: 4px;">No. Dokumen: <b>${docNumber.value}</b></div>
      <div style="font-size: 10px; color: #94a3b8; font-family: monospace; margin-top: 4px;">${sLat.toFixed(6)}, ${sLng.toFixed(6)}</div>
    </div>
  `)

  // 4. Connecting Polyline between Store and Sales
  const latlngs: L.LatLngExpression[] = [
    [sLat, sLng],
    [stLat, stLng],
  ]
  const lineColor = isOutsideRadius.value ? '#e11d48' : '#059669'
  const polyline = L.polyline(latlngs, {
    color: lineColor,
    weight: 3,
    dashArray: '6, 8',
    opacity: 0.9,
  }).addTo(map)

  polyline.bindTooltip(`Jarak: ${formattedDistance.value} (${isOutsideRadius.value ? 'Luar Radius' : 'Dalam Radius'})`, {
    permanent: true,
    direction: 'center',
    className: 'leaflet-distance-tooltip',
  })

  // Fit bounds precisely
  const bounds = L.latLngBounds([[stLat, stLng], [sLat, sLng]])
  map.fitBounds(bounds, { padding: [60, 60], maxZoom: 16 })
}

const triggerInitMap = async () => {
  await nextTick()
  setTimeout(() => {
    initMap()
    if (map) {
      map.invalidateSize()
    }
  }, 100)
  setTimeout(() => {
    if (map) {
      map.invalidateSize()
      const sLat = salesLat.value
      const sLng = salesLng.value
      const stLat = storeLat.value
      const stLng = storeLng.value
      if (sLat !== null && sLng !== null && stLat !== null && stLng !== null) {
        map.fitBounds(L.latLngBounds([[stLat, stLng], [sLat, sLng]]), { padding: [60, 60], maxZoom: 16 })
      }
    }
  }, 350)
}

watch(
  () => props.open,
  (newVal) => {
    if (newVal) {
      triggerInitMap()
    } else {
      if (map) {
        try {
          map.remove()
        } catch (e) {}
        map = null
      }
    }
  },
)

watch(
  () => props.orderData,
  () => {
    if (props.open) {
      triggerInitMap()
    }
  },
  { deep: true },
)
</script>

<template>
  <Dialog :open="open" @update:open="emit('update:open', $event)">
    <DialogContent class="max-w-4xl p-0 overflow-hidden rounded-[28px] border-0 shadow-2xl">
      <!-- Modal Header -->
      <div class="px-6 pt-6 pb-4 bg-white border-b border-slate-100 flex flex-col gap-3">
        <div class="flex items-start justify-between">
          <div>
            <DialogTitle class="text-lg font-black text-slate-900 flex items-center gap-2">
              <Navigation class="w-5 h-5 text-primary" />
              Verifikasi Lokasi & Radius Penerbitan
            </DialogTitle>
            <DialogDescription class="text-xs text-slate-500 mt-1">
              Perbandingan posisi koordinat fisik saat salesman menginput pesanan/faktur terhadap titik koordinat Master Toko.
            </DialogDescription>
          </div>
          <div class="text-right">
            <span class="text-[10px] font-black text-slate-400 uppercase tracking-widest block">No. Dokumen</span>
            <span class="text-xs font-mono font-black text-slate-800 bg-slate-100 px-2 py-0.5 rounded-md mt-0.5 inline-block">
              #{{ docNumber }}
            </span>
          </div>
        </div>

        <!-- Info Comparison Cards -->
        <div class="grid grid-cols-1 md:grid-cols-3 gap-3 pt-1">
          <!-- Store Card -->
          <div class="bg-blue-50/70 border border-blue-100 rounded-2xl p-3 flex items-start gap-3">
            <div class="p-2 bg-blue-600 text-white rounded-xl shadow-sm shrink-0">
              <Store class="w-4 h-4" />
            </div>
            <div class="min-w-0 flex-1">
              <div class="flex items-center justify-between">
                <span class="text-[9px] font-black text-blue-700 uppercase tracking-wider">Lokasi Master Toko</span>
                <span class="text-[9px] font-bold text-blue-600/80 bg-blue-100/70 px-1.5 py-0.2 rounded">Batas: 500m</span>
              </div>
              <p class="text-xs font-black text-slate-900 truncate mt-0.5" :title="storeName">{{ storeName }}</p>
              <p class="text-[10px] text-slate-500 truncate" :title="storeAddress">{{ storeAddress }}</p>
              <p v-if="storeLat && storeLng" class="text-[9px] font-mono text-blue-600 mt-0.5">
                {{ storeLat.toFixed(5) }}, {{ storeLng.toFixed(5) }}
              </p>
              <div v-if="storeLat && storeLng" class="mt-1.5 flex gap-1">
                <button
                  type="button"
                  @click="openGoogleMaps(storeLat, storeLng, storeName)"
                  class="text-[8.5px] font-black px-1.5 py-0.5 rounded bg-blue-100 hover:bg-blue-200 text-blue-800 flex items-center gap-0.5 cursor-pointer"
                >
                  <ExternalLink class="w-2.5 h-2.5" /> Maps Toko
                </button>
              </div>
            </div>
          </div>

          <!-- Sales Card -->
          <div class="bg-rose-50/70 border border-rose-100 rounded-2xl p-3 flex items-start gap-3">
            <div class="p-2 bg-rose-600 text-white rounded-xl shadow-sm shrink-0">
              <User class="w-4 h-4" />
            </div>
            <div class="min-w-0 flex-1">
              <span class="text-[9px] font-black text-rose-700 uppercase tracking-wider">Titik Input Sales</span>
              <p class="text-xs font-black text-slate-900 truncate mt-0.5" :title="salesmanName">{{ salesmanName }}</p>
              <p class="text-[10px] text-slate-500">Koordinat saat penerbitan</p>
              <p v-if="salesLat && salesLng" class="text-[9px] font-mono text-rose-600 mt-0.5">
                {{ salesLat.toFixed(5) }}, {{ salesLng.toFixed(5) }}
              </p>
              <div v-if="salesLat && salesLng" class="mt-1.5 flex gap-1">
                <button
                  type="button"
                  @click="openGoogleStreetView(salesLat, salesLng)"
                  class="text-[8.5px] font-black px-1.5 py-0.5 rounded bg-indigo-100 hover:bg-indigo-200 text-indigo-800 flex items-center gap-0.5 cursor-pointer"
                  title="Lihat Google Street View 360°"
                >
                  <Navigation class="w-2.5 h-2.5" /> Street View 360°
                </button>
                <button
                  type="button"
                  @click="openGoogleMaps(salesLat, salesLng, salesmanName)"
                  class="text-[8.5px] font-black px-1.5 py-0.5 rounded bg-rose-100 hover:bg-rose-200 text-rose-800 flex items-center gap-0.5 cursor-pointer"
                  title="Lihat Google Maps"
                >
                  <ExternalLink class="w-2.5 h-2.5" /> Maps
                </button>
              </div>
            </div>
          </div>

          <!-- Distance & Status Card -->
          <div 
            :class="[
              'rounded-2xl p-3 flex items-center justify-between border',
              isOutsideRadius ? 'bg-red-50/80 border-red-200' : 'bg-emerald-50/80 border-emerald-200'
            ]"
          >
            <div>
              <span class="text-[9px] font-black uppercase tracking-wider block" :class="isOutsideRadius ? 'text-red-700' : 'text-emerald-700'">
                Jarak Fisik Riil
              </span>
              <p class="text-base font-black mt-0.5" :class="isOutsideRadius ? 'text-red-900' : 'text-emerald-900'">
                {{ formattedDistance }}
              </p>
              <span class="text-[9px] font-medium text-slate-500 block mt-0.5">
                Toleransi radius: 500 meter
              </span>
            </div>
            <div class="text-right">
              <Badge 
                :class="`px-2.5 py-1 text-[10px] font-black uppercase tracking-wider shadow-sm border-0 ${isOutsideRadius ? 'bg-red-600 text-white hover:bg-red-700' : 'bg-emerald-600 text-white hover:bg-emerald-700'}`"
              >
                <component :is="isOutsideRadius ? MapPinOff : MapPin" class="w-3 h-3 mr-1" />
                {{ isOutsideRadius ? 'Luar Radius' : 'Sesuai Radius' }}
              </Badge>
            </div>
          </div>
        </div>
      </div>

      <!-- Map Viewport -->
      <div class="relative bg-slate-100 min-h-[420px]">
        <div 
          v-show="hasValidCoordinates" 
          class="h-[52vh] min-h-[420px] w-full z-0 relative" 
          ref="mapContainer"
        ></div>
        
        <!-- Empty / Invalid Coordinates State -->
        <div v-if="!hasValidCoordinates" class="h-[52vh] min-h-[420px] w-full flex flex-col items-center justify-center p-8 text-center bg-slate-50">
          <div class="w-14 h-14 rounded-2xl bg-amber-100 text-amber-600 flex items-center justify-center mb-3">
            <Info class="w-7 h-7" />
          </div>
          <h4 class="text-sm font-bold text-slate-800">Koordinat Lokasi Tidak Lengkap</h4>
          <p class="text-xs text-slate-500 max-w-md mt-1">
            Data transaksi atau master toko belum memiliki titik koordinat GPS yang valid sehingga peta tidak dapat dirender.
          </p>
          <div class="mt-4 flex gap-4 text-xs font-mono text-slate-600 bg-white p-3 rounded-xl border border-slate-200">
            <div>Toko: {{ storeLat || '0' }}, {{ storeLng || '0' }}</div>
            <div>|</div>
            <div>Sales: {{ salesLat || '0' }}, {{ salesLng || '0' }}</div>
          </div>
        </div>

        <!-- Floating Map Legend Indicator -->
        <div v-if="hasValidCoordinates" class="absolute bottom-3 left-3 z-[400] bg-white/95 backdrop-blur-sm px-3 py-2 rounded-xl shadow-lg border border-slate-200 text-[10px] font-bold text-slate-700 flex items-center gap-4">
          <div class="flex items-center gap-1.5">
            <span class="w-3 h-3 rounded-full bg-blue-600 border border-white shadow-sm inline-block"></span>
            <span>Titik Toko (Master)</span>
          </div>
          <div class="flex items-center gap-1.5">
            <span class="w-3 h-3 rounded-full bg-rose-600 border border-white shadow-sm inline-block"></span>
            <span>Titik Salesman (Input)</span>
          </div>
          <div class="flex items-center gap-1.5">
            <span class="w-3 h-3 rounded-full bg-blue-300/50 border border-blue-400 border-dashed inline-block"></span>
            <span>Zona Radius 500m</span>
          </div>
        </div>
      </div>
    </DialogContent>
  </Dialog>
</template>

<style>
/* Custom animations and tooltip styling for Leaflet */
@keyframes bounce-subtle {
  0%, 100% {
    transform: translateY(0);
  }
  50% {
    transform: translateY(-4px);
  }
}

.custom-leaflet-marker-store,
.custom-leaflet-marker-sales {
  background: transparent !important;
  border: none !important;
}

.leaflet-distance-tooltip {
  background: rgba(15, 23, 42, 0.9) !important;
  color: white !important;
  border: none !important;
  border-radius: 8px !important;
  font-size: 10px !important;
  font-weight: 700 !important;
  padding: 3px 8px !important;
  box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.2) !important;
}
</style>
