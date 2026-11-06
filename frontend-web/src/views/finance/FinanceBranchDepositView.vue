<template>
  <div class="p-6">
    <div class="flex flex-col sm:flex-row justify-between items-start sm:items-center mb-6">
      <div>
        <h1 class="text-2xl font-bold text-gray-900">Validasi Setoran Cabang</h1>
        <p class="text-sm text-gray-500 mt-1">Verifikasi laporan setoran masuk dari Admin Cabang dan buat Jurnal</p>
      </div>
    </div>

    <!-- Table -->
    <div class="bg-white rounded-xl shadow overflow-hidden">
      <div class="overflow-x-auto">
        <table class="min-w-full divide-y divide-gray-200">
          <thead class="bg-gray-50">
            <tr>
              <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Cabang</th>
              <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Tanggal</th>
              <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Metode</th>
              <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Nominal</th>
              <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Status</th>
              <th scope="col" class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">Aksi</th>
            </tr>
          </thead>
          <tbody class="bg-white divide-y divide-gray-200">
            <!-- Simulated empty state -->
            <tr>
              <td colspan="6" class="px-6 py-8 text-center text-gray-500">
                Belum ada pengajuan setoran dari cabang. (Implementasi API menyusul)
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'

const approveDeposit = (id: string) => {
  if(confirm("Setujui setoran ini? Sistem akan otomatis menjurnal transaksi ini ke buku besar (Kas/Bank vs Piutang Cabang).")) {
    alert("Setoran disetujui (Simulasi)")
  }
}
</script>
