<script setup lang="ts">
import { ref } from 'vue'

const logs = ref([
  {
    id: '1',
    timestamp: '2026-04-15 08:12:00',
    actor: 'Ahmad Sofyan (Super Admin)',
    action: 'UPDATE',
    target: 'Pajak PTKP Settings',
    detail: 'Mengubah nominal PPh21 dari Rp54.000.000 ke Rp58.000.000',
  },
  {
    id: '2',
    timestamp: '2026-04-14 16:35:12',
    actor: 'Citra (Admin Nota)',
    action: 'APPROVE',
    target: 'Reimbursement #R-1002',
    detail: 'Menyetujui klaim transport Budi Santoso senilai Rp250.000',
  },
  {
    id: '3',
    timestamp: '2026-04-14 11:20:05',
    actor: 'Dika (HR Admin)',
    action: 'CREATE',
    target: 'Data Karyawan Baru',
    detail: 'Menambahkan staf operasional baru (ID: EMP-092)',
  },
])
</script>

<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-3xl font-bold tracking-tight text-slate-900">
          Audit Trail (Log Aktivitas)
        </h1>
        <p class="text-slate-500 mt-1">
          Pantau riwayat operasi kritis yang dilakukan oleh user manajerial untuk jaminan
          akuntabilitas.
        </p>
      </div>
    </div>

    <div class="bg-white rounded-xl border border-slate-200 shadow-sm overflow-hidden p-6 text-sm">
      <div
        class="space-y-8 relative before:absolute before:inset-0 before:ml-2.5 before:-translate-x-px md:before:mx-auto md:before:translate-x-0 before:h-full before:w-0.5 before:bg-gradient-to-b before:from-transparent before:via-slate-200 before:to-transparent"
      >
        <div
          v-for="log in logs"
          :key="log.id"
          class="relative flex items-center justify-between md:justify-normal md:odd:flex-row-reverse group is-active"
        >
          <!-- Icon -->
          <div
            class="flex items-center justify-center w-6 h-6 rounded-full border border-white bg-slate-200 text-slate-500 shadow shrink-0 md:order-1 md:group-odd:-translate-x-1/2 md:group-even:translate-x-1/2"
          >
            <svg class="w-3 h-3" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
              />
            </svg>
          </div>

          <!-- Card -->
          <div
            class="w-[calc(100%-2rem)] md:w-[calc(50%-2rem)] p-4 rounded-xl border border-slate-200 bg-white shadow-sm hover:border-slate-300 transition-colors"
          >
            <div class="flex items-center justify-between mb-1">
              <span class="font-bold text-slate-900">{{ log.actor }}</span>
              <span
                class="text-xs font-medium text-slate-500 bg-slate-100 px-2 py-0.5 rounded inline-block"
                >{{ log.action }}</span
              >
            </div>
            <p class="text-xs text-indigo-500 mb-2">{{ log.timestamp }}</p>
            <p class="text-slate-700 font-medium mb-1">Modifikasi: {{ log.target }}</p>
            <p class="text-slate-500">{{ log.detail }}</p>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
