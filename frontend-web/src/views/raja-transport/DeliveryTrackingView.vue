<script setup lang="ts">
import { MapPin } from 'lucide-vue-next'
import DeliveryTrackingTab from './DeliveryTrackingTab.vue'
</script>

<template>
  <div class="p-4 md:p-6 space-y-6 bg-[#f8fafc] min-h-screen">
    <!-- Header -->
    <div
      class="flex flex-col lg:flex-row justify-between items-start lg:items-center gap-4 bg-white p-6 rounded-3xl shadow-xs border border-slate-100"
    >
      <div class="flex items-center gap-3">
        <div class="p-2.5 rounded-2xl bg-amber-500/10 text-amber-600">
          <MapPin class="w-7 h-7" />
        </div>
        <div>
          <h1 class="text-2xl font-black tracking-tight text-slate-900">
            Live Tracking Pengiriman (GPS)
          </h1>
          <p class="text-xs md:text-sm text-slate-500 mt-0.5">
            Pelacakan posisi satelit armada logistik real-time TraxCenter Open API beserta kecepatan dan arah tempuh kendaraan.
          </p>
        </div>
      </div>
    </div>

    <!-- Main Content -->
    <DeliveryTrackingTab />
  </div>
</template>
