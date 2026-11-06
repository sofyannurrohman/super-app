<template>
  <div class="p-6 space-y-6 bg-slate-50 min-h-screen">
    <div class="bg-white rounded-2xl p-6 shadow-sm">
      <h1 class="text-2xl font-black text-slate-900">Approval Request Distributor</h1>
      <p class="text-sm text-slate-500 mt-1">Khusus Super Admin dan SPV Gudang.</p>
    </div>
    <div class="bg-white rounded-2xl shadow-sm overflow-hidden">
      <div v-if="loading" class="p-10 text-center text-sm text-slate-400">Memuat request...</div>
      <div v-else-if="requests.length === 0" class="p-10 text-center text-sm text-slate-400">Tidak ada request pending.</div>
      <div v-for="request in requests" :key="request.id" class="p-5 border-b flex flex-col md:flex-row md:items-center justify-between gap-4">
        <div>
          <p class="font-black text-slate-900">{{ request.distributor_store?.name || 'Distributor' }}</p>
          <p class="text-sm text-slate-600">{{ request.product?.name || 'Produk' }} · {{ request.quantity }} {{ request.unit || 'PCS' }}</p>
          <p class="text-xs text-slate-400 mt-1">{{ request.notes || 'Tanpa catatan' }}</p>
        </div>
        <div class="flex gap-2">
          <button class="rounded-xl border border-rose-200 px-4 py-2 text-xs font-bold text-rose-700" @click="decide(request.id, 'reject')">Tolak</button>
          <button class="rounded-xl bg-emerald-600 px-4 py-2 text-xs font-bold text-white" @click="decide(request.id, 'approve')">Setujui</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import apiClient from '@/api/axios'
const requests = ref<any[]>([])
const loading = ref(false)
async function load() {
  loading.value = true
  try { const res = await apiClient.get('/sales-transfers/supervisor/pending', { params: { request_source: 'DISTRIBUTOR' } }); requests.value = res.data?.data || [] }
  finally { loading.value = false }
}
async function decide(id: string, action: 'approve' | 'reject') {
  await apiClient.patch(`/sales-transfers/${id}/${action}-supervisor`)
  await load()
}
onMounted(load)
</script>
