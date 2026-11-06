<script setup lang="ts">
import { RotateCcw } from 'lucide-vue-next'
import BackhaulTab from './BackhaulTab.vue'
</script>

<template>
  <div class="p-4 md:p-6 space-y-6 bg-[#f8fafc] min-h-screen">
    <!-- Header -->
    <div
      class="flex flex-col lg:flex-row justify-between items-start lg:items-center gap-4 bg-white p-6 rounded-3xl shadow-xs border border-slate-100"
    >
      <div class="flex items-center gap-3">
        <div class="p-2.5 rounded-2xl bg-amber-500/10 text-amber-600">
          <RotateCcw class="w-7 h-7" />
        </div>
        <div>
          <h1 class="text-2xl font-black tracking-tight text-slate-900">
            Pengiriman Muatan Balik (Fee 10%)
          </h1>
          <p class="text-xs md:text-sm text-slate-500 mt-0.5">
            Persetujuan muatan balikan sopir ekspedisi dari kota tujuan, perhitungan insentif fee 10%, dan verifikasi bukti nota/surat jalan.
          </p>
        </div>
      </div>
    </div>

    <!-- Main Content -->
    <BackhaulTab />
  </div>
</template>
