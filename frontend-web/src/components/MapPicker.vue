<script setup lang="ts">
import { ref, onMounted, watch, onUnmounted } from 'vue'
import L from 'leaflet'
import 'leaflet/dist/leaflet.css'
import 'leaflet-defaulticon-compatibility/dist/leaflet-defaulticon-compatibility.css'
import 'leaflet-defaulticon-compatibility'

const props = defineProps<{
  latitude: number
  longitude: number
  radius: number
  height?: string
}>()

const emit = defineEmits<{
  (e: 'update:location', lat: number, lng: number): void
  (e: 'update:address', address: string): void
}>()

const mapContainer = ref<HTMLElement | null>(null)
let map: L.Map | null = null
let marker: L.Marker | null = null
let circle: L.Circle | null = null

const initMap = () => {
  if (!mapContainer.value) return

  const initialLat = props.latitude || -6.2
  const initialLng = props.longitude || 106.816666

  map = L.map(mapContainer.value).setView([initialLat, initialLng], 15)

  L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
    attribution:
      '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors',
  }).addTo(map)

  marker = L.marker([initialLat, initialLng], {
    draggable: true,
  }).addTo(map)

  circle = L.circle([initialLat, initialLng], {
    radius: props.radius,
    color: '#3b82f6',
    fillColor: '#3b82f6',
    fillOpacity: 0.2,
  }).addTo(map)

  marker.on('dragend', (event: L.LeafletEvent) => {
    const position = (event.target as L.Marker).getLatLng()
    updateLocation(position.lat, position.lng)
  })

  map.on('click', (event: L.LeafletMouseEvent) => {
    const position = event.latlng
    marker?.setLatLng(position)
    updateLocation(position.lat, position.lng)
  })
}

const updateLocation = (lat: number, lng: number) => {
  circle?.setLatLng([lat, lng])
  emit('update:location', lat, lng)
  reverseGeocode(lat, lng)
}

const reverseGeocode = async (lat: number, lng: number) => {
  try {
    const response = await fetch(
      `https://nominatim.openstreetmap.org/reverse?format=json&lat=${lat}&lon=${lng}&zoom=18&addressdetails=1`,
    )
    const data = await response.json()
    if (data.display_name) {
      emit('update:address', data.display_name)
    }
  } catch (error) {
    console.error('Reverse geocoding failed:', error)
  }
}

watch(
  () => props.radius,
  (newRadius) => {
    circle?.setRadius(newRadius)
  },
)

watch(
  () => props.latitude,
  (newLat) => {
    if (marker && map) {
      const currentLatLng = marker.getLatLng()
      if (Math.abs(currentLatLng.lat - newLat) > 0.000001) {
        marker.setLatLng([newLat, props.longitude])
        circle?.setLatLng([newLat, props.longitude])
        map.panTo([newLat, props.longitude])
      }
    }
  },
)

watch(
  () => props.longitude,
  (newLng) => {
    if (marker && map) {
      const currentLatLng = marker.getLatLng()
      if (Math.abs(currentLatLng.lng - newLng) > 0.000001) {
        marker.setLatLng([props.latitude, newLng])
        circle?.setLatLng([props.latitude, newLng])
        map.panTo([props.latitude, newLng])
      }
    }
  },
)

onMounted(() => {
  setTimeout(() => {
    initMap()
  }, 200) // Small delay to ensure container is rendered in dialog
})

onUnmounted(() => {
  if (map) {
    map.remove()
  }
})
</script>

<template>
  <div class="relative w-full rounded-lg overflow-hidden border border-gray-200 shadow-inner">
    <div ref="mapContainer" :style="{ height: height || '300px' }"></div>
    <div
      class="absolute bottom-2 right-2 z-[1000] bg-white/90 backdrop-blur-sm px-2 py-1 rounded text-[10px] text-gray-500 font-medium border border-gray-200"
    >
      Klik peta atau geser pin untuk pilih lokasi
    </div>
  </div>
</template>

<style scoped>
:deep(.leaflet-container) {
  width: 100%;
  height: 100%;
}
</style>
