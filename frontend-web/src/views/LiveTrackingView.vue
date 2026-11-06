<template>
  <div class="p-8 space-y-8 bg-[#f8fafc] min-h-screen">
    <!-- Header -->
    <div
      class="flex flex-col md:flex-row justify-between items-start md:items-center gap-4 bg-white p-8 rounded-[2rem] shadow-sm border border-slate-100"
    >
      <div>
        <h1 class="text-3xl font-black tracking-tight text-slate-900">Live Tracking Armada</h1>
        <p class="text-slate-500 mt-1">
          Pantau pergerakan kendaraan secara real-time dari GPS Tracker.
        </p>
      </div>
      <div class="flex items-center gap-3">
        <Button
          @click="$router.push('/rajatransport/tracking')"
          class="rounded-2xl h-12 px-6 font-bold bg-amber-600 hover:bg-amber-700 text-white shadow-lg shadow-amber-500/20"
        >
          <Truck class="mr-2 h-4 w-4" /> Tracking Armada
        </Button>
        <Button
          @click="$router.push('/vehicles')"
          variant="outline"
          class="rounded-2xl h-12 px-8 font-bold"
        >
          <ArrowLeft class="mr-2 h-4 w-4" /> Kembali ke Kendaraan
        </Button>
      </div>
    </div>

    <div class="bg-white rounded-[2.5rem] p-4 shadow-sm border border-slate-100 min-h-[600px] relative">
      <div id="map" class="w-full h-[600px] rounded-[2rem] z-0"></div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, shallowRef, onMounted, onUnmounted } from 'vue';
import { useRouter } from 'vue-router';
import { ArrowLeft, Truck } from 'lucide-vue-next';
import L from 'leaflet';
import 'leaflet/dist/leaflet.css';
import axios from 'axios';
import { toast } from 'vue-sonner';

// Fix leaflet default icon issue
import iconRetinaUrl from 'leaflet/dist/images/marker-icon-2x.png';
import iconUrl from 'leaflet/dist/images/marker-icon.png';
import shadowUrl from 'leaflet/dist/images/marker-shadow.png';

L.Icon.Default.mergeOptions({
  iconRetinaUrl,
  iconUrl,
  shadowUrl,
});

const router = useRouter();
const map = shallowRef<L.Map | null>(null);
const markers = shallowRef<{ [key: string]: L.Marker }>({});
const intervalId = ref<any>(null);

const fetchLocations = async () => {
  try {
    const res = await axios.get('/vehicles/locations/live');
    if (res.data) {
      updateMap(res.data);
    }
  } catch (error) {
    console.error('Error fetching live locations:', error);
  }
};

const updateMap = (locations: any[]) => {
  if (!map.value) return;
  
  const bounds = L.latLngBounds([]);

  locations.forEach(loc => {
    const latlng = L.latLng(loc.latitude, loc.longitude);
    const vehicleName = loc.vehicle?.name || 'Kendaraan';
    const plate = loc.vehicle?.license_plate || '-';
    
    // Icon based on ACC status
    const iconColor = loc.is_acc_on ? 'green' : 'red';
    
    // Create HTML for custom marker
    const customIcon = L.divIcon({
      className: 'custom-div-icon',
      html: `<div style="background-color: ${iconColor}; border-radius: 50%; width: 16px; height: 16px; border: 2px solid white; box-shadow: 0 0 4px rgba(0,0,0,0.4);"></div>`,
      iconSize: [16, 16],
      iconAnchor: [8, 8]
    });

    const popupContent = `
      <div class="font-sans">
        <div class="font-black text-sm">${plate}</div>
        <div class="text-xs text-slate-500">${vehicleName}</div>
        <div class="mt-2 text-xs font-bold text-slate-700">Mesin: ${loc.is_acc_on ? 'ON' : 'OFF'}</div>
        <div class="text-xs text-slate-500">Kecepatan: ${loc.speed} km/h</div>
        <div class="text-[10px] text-slate-400 mt-1">Update: ${new Date(loc.timestamp).toLocaleTimeString()}</div>
      </div>
    `;

    const existingMarker = markers.value[loc.vehicle_id];
    if (existingMarker) {
      existingMarker.setLatLng(latlng);
      existingMarker.setPopupContent(popupContent);
    } else {
      const marker = L.marker(latlng, { icon: customIcon })
        .addTo(map.value!)
        .bindPopup(popupContent);
      markers.value[loc.vehicle_id] = marker;
    }
    bounds.extend(latlng);
  });

  if (locations.length > 0 && map.value) {
    // Only fit bounds on first load or if requested
    // map.value.fitBounds(bounds, { padding: [50, 50], maxZoom: 15 });
  }
};

onMounted(() => {
  // Initialize map
  map.value = L.map('map').setView([-6.200000, 106.816666], 11); // Default Jakarta

  L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
    attribution: '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors'
  }).addTo(map.value!);

  fetchLocations();
  
  // Refresh every 10 seconds
  intervalId.value = setInterval(fetchLocations, 10000);
});

onUnmounted(() => {
  if (intervalId.value) {
    clearInterval(intervalId.value);
  }
  if (map.value) {
    map.value.remove();
  }
});
</script>

<style>
/* Leaflet map styling overrides if needed */
.leaflet-popup-content-wrapper {
  border-radius: 1rem;
  box-shadow: 0 10px 15px -3px rgb(0 0 0 / 0.1), 0 4px 6px -4px rgb(0 0 0 / 0.1);
}
</style>
