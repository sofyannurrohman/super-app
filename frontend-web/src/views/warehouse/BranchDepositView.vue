<template>
  <div class="p-6">
    <div class="flex flex-col sm:flex-row justify-between items-start sm:items-center mb-6">
      <div>
        <h1 class="text-2xl font-bold text-gray-900">Setoran Harian Cabang</h1>
        <p class="text-sm text-gray-500 mt-1">Kelola dan laporkan setoran tunai/transfer ke Pusat</p>
      </div>
      <button 
        @click="showCreateModal = true"
        class="mt-4 sm:mt-0 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors flex items-center"
      >
        <span class="mr-2">➕</span> Buat Laporan Setoran
      </button>
    </div>

    <!-- Stats/Filters can go here -->

    <!-- Table -->
    <div class="bg-white rounded-xl shadow overflow-hidden">
      <div class="overflow-x-auto">
        <table class="min-w-full divide-y divide-gray-200">
          <thead class="bg-gray-50">
            <tr>
              <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Tanggal</th>
              <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Metode</th>
              <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Nominal</th>
              <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Catatan</th>
              <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Status</th>
            </tr>
          </thead>
          <tbody class="bg-white divide-y divide-gray-200">
            <!-- Simulated empty state -->
            <tr>
              <td colspan="5" class="px-6 py-8 text-center text-gray-500">
                Data setoran cabang akan muncul di sini. (Implementasi API menyusul)
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Create Modal Placeholder -->
    <div v-if="showCreateModal" class="fixed inset-0 z-50 overflow-y-auto flex items-center justify-center">
      <div class="fixed inset-0 bg-black bg-opacity-50 transition-opacity" @click="showCreateModal = false"></div>
      <div class="relative bg-white rounded-xl shadow-xl w-full max-w-md mx-4 p-6 overflow-hidden">
        <h3 class="text-lg font-bold text-gray-900 mb-4">Input Setoran Baru</h3>
        
        <div class="space-y-4">
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Tanggal Setor</label>
            <input type="date" v-model="formSetoran.deposit_date" @change="fetchDailyBalance" class="w-full border-gray-300 rounded-lg shadow-sm focus:ring-blue-500 focus:border-blue-500" />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Metode Setoran</label>
            <select v-model="formSetoran.payment_method" @change="fetchDailyBalance" class="w-full border-gray-300 rounded-lg shadow-sm focus:ring-blue-500 focus:border-blue-500">
              <option value="CASH">Tunai (Cash)</option>
              <option value="TRANSFER">Transfer Bank</option>
            </select>
          </div>
          
          <div v-if="formSetoran.payment_method === 'CASH'" class="p-3 bg-blue-50 rounded-lg text-sm mb-2">
            <p class="text-blue-800 font-medium">Total Validasi Admin Nota (Sistem): Rp {{ systemAmount.toLocaleString() }}</p>
          </div>

          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Nominal (Rp)</label>
            <input type="number" v-model="formSetoran.amount" placeholder="Contoh: 5000000" class="w-full border-gray-300 rounded-lg shadow-sm focus:ring-blue-500 focus:border-blue-500" />
          </div>
          
          <div v-if="formSetoran.payment_method === 'CASH' && formSetoran.amount > 0 && discrepancy !== 0" class="p-3 bg-yellow-50 rounded-lg text-sm mt-2">
            <p class="text-yellow-800 font-medium">⚠️ Peringatan: Terdapat selisih Rp {{ discrepancy.toLocaleString() }}</p>
            <p class="text-yellow-700 mt-1">Nominal setoran tidak sama dengan sistem. Anda tetap bisa mengirim laporan ini, namun selisih akan dicatat.</p>
          </div>

          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1 mt-2">Bukti Setor/Transfer (URL/Foto)</label>
            <input type="file" class="w-full text-sm text-gray-500 file:mr-4 file:py-2 file:px-4 file:rounded-full file:border-0 file:text-sm file:font-semibold file:bg-blue-50 file:text-blue-700 hover:file:bg-blue-100" />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Catatan Tambahan</label>
            <textarea v-model="formSetoran.notes" rows="2" class="w-full border-gray-300 rounded-lg shadow-sm focus:ring-blue-500 focus:border-blue-500"></textarea>
          </div>
        </div>

        <div class="mt-6 flex justify-end space-x-3">
          <button @click="showCreateModal = false" class="px-4 py-2 bg-gray-100 text-gray-700 rounded-lg hover:bg-gray-200">
            Batal
          </button>
          <button @click="submitDeposit" class="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700">
            Kirim Laporan
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'

const showCreateModal = ref(false)
const systemAmount = ref(0)
const formSetoran = ref({
  deposit_date: new Date().toISOString().split('T')[0],
  amount: 0,
  payment_method: 'CASH',
  notes: ''
})

const discrepancy = computed(() => {
  return formSetoran.value.amount - systemAmount.value
})

const fetchDailyBalance = async () => {
  if (formSetoran.value.payment_method !== 'CASH') return;
  // TODO: call API with current branch_id and date
  // const res = await api.get(`/branch-deposits/daily-balance?branch_id=${branchID}&date=${formSetoran.value.deposit_date}`)
  // systemAmount.value = res.data.data.total_admin_nota
  
  // Simulated API response for demonstration
  systemAmount.value = 5000000;
}

const submitDeposit = () => {
  // TODO: Call API SubmitDeposit
  alert(`Simulasi berhasil mengirim laporan setoran!\nActual: ${formSetoran.value.amount}\nSystem: ${systemAmount.value}\nSelisih: ${discrepancy.value}`)
  showCreateModal.value = false
}
</script>
