<template>
  <div class="p-6">
    <h1 class="text-2xl font-bold text-gray-800 mb-6">Banner Request Kanban (Sales Booster)</h1>

    <div class="flex gap-6 overflow-x-auto pb-4">
      <!-- Requested Column -->
      <div class="flex-shrink-0 w-80 bg-gray-100 rounded-lg p-4 shadow">
        <h3 class="font-bold text-gray-700 mb-3 flex justify-between items-center">
          Diajukan <span class="bg-gray-200 text-gray-600 text-xs px-2 py-1 rounded-full">2</span>
        </h3>
        <div class="space-y-3">
          <!-- Card -->
          <div
            class="bg-white p-4 rounded shadow-sm border-l-4 border-yellow-400 cursor-move hover:shadow-md transition"
          >
            <div class="flex justify-between items-start mb-2">
              <span class="text-xs font-semibold bg-blue-100 text-blue-800 px-2 py-1 rounded"
                >Spanduk 3x1</span
              >
              <span class="text-xs text-gray-500">2 hari lalu</span>
            </div>
            <h4 class="font-bold text-gray-800 text-sm">Toko Makmur Jaya</h4>
            <p class="text-xs text-gray-600 mt-1">Requested by: Sales Booster Anton</p>
          </div>

          <div
            class="bg-white p-4 rounded shadow-sm border-l-4 border-yellow-400 cursor-move hover:shadow-md transition"
          >
            <div class="flex justify-between items-start mb-2">
              <span class="text-xs font-semibold bg-purple-100 text-purple-800 px-2 py-1 rounded"
                >Neon Box</span
              >
              <span class="text-xs text-gray-500">1 hari lalu</span>
            </div>
            <h4 class="font-bold text-gray-800 text-sm">Toko Berkah Abadi</h4>
            <p class="text-xs text-gray-600 mt-1">Requested by: Sales Booster Anton</p>
          </div>
        </div>
      </div>

      <!-- In Progress Column -->
      <div class="flex-shrink-0 w-80 bg-gray-100 rounded-lg p-4 shadow">
        <h3 class="font-bold text-gray-700 mb-3 flex justify-between items-center">
          Proses Produksi
          <span class="bg-blue-200 text-blue-800 text-xs px-2 py-1 rounded-full">1</span>
        </h3>
        <div class="space-y-3">
          <div
            class="bg-white p-4 rounded shadow-sm border-l-4 border-blue-500 cursor-move hover:shadow-md transition"
          >
            <div class="flex justify-between items-start mb-2">
              <span class="text-xs font-semibold bg-blue-100 text-blue-800 px-2 py-1 rounded"
                >Spanduk 2x1</span
              >
              <span class="text-xs text-gray-500">Diupdate 5 jam lalu</span>
            </div>
            <h4 class="font-bold text-gray-800 text-sm">Toko Laris Manis</h4>
            <p class="text-xs text-gray-600 mt-1">Vendor: CV Sentosa Print</p>
          </div>
        </div>
      </div>

      <!-- Ready to Install Column -->
      <div class="flex-shrink-0 w-80 bg-gray-100 rounded-lg p-4 shadow">
        <h3 class="font-bold text-gray-700 mb-3 flex justify-between items-center">
          Siap Pasang
          <span class="bg-orange-200 text-orange-800 text-xs px-2 py-1 rounded-full">0</span>
        </h3>
        <div
          class="space-y-3 min-h-[100px] border-2 border-dashed border-gray-300 rounded flex items-center justify-center text-gray-400 text-sm"
        >
          Tarik card ke sini
        </div>
      </div>

      <!-- Done Column -->
      <div class="flex-shrink-0 w-80 bg-gray-100 rounded-lg p-4 shadow">
        <h3 class="font-bold text-gray-700 mb-3 flex justify-between items-center">
          Selesai Terpasang
          <span class="bg-green-200 text-green-800 text-xs px-2 py-1 rounded-full">1</span>
        </h3>
        <div class="space-y-3">
          <div class="bg-white p-4 rounded shadow-sm border-l-4 border-green-500 opacity-75">
            <div class="flex justify-between items-start mb-2">
              <span class="text-xs font-semibold bg-green-100 text-green-800 px-2 py-1 rounded"
                >Spanduk 4x1</span
              >
              <span class="text-xs text-gray-500">Minggu lalu</span>
            </div>
            <h4 class="font-bold text-gray-800 text-sm line-through decoration-gray-400">
              Warung Kopi Sinar
            </h4>
            <p class="text-xs text-gray-600 mt-1">Dipasang oleh: Tim Kurir Budi</p>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
// Logic for drag and drop kanban board
</script>
