<template>
  <div class="p-6 space-y-6 bg-slate-50 min-h-screen">
    <div class="bg-white rounded-2xl p-6 shadow-sm">
      <h1 class="text-2xl font-black text-slate-900">Request Distributor</h1>
      <p class="text-sm text-slate-500 mt-1">Buat request stok distributor dan pantau status approvalnya.</p>
    </div>

    <div class="grid lg:grid-cols-3 gap-6">
      <form class="bg-white rounded-2xl p-5 shadow-sm space-y-4" @submit.prevent="submit">
        <h2 class="font-black text-slate-900">Buat Request</h2>
        <label class="block text-xs font-bold text-slate-600">Distributor</label>
        <select v-model="form.distributor_store_id" required class="w-full rounded-xl border p-2 text-sm">
          <option value="">Pilih distributor</option>
          <option v-for="store in distributors" :key="store.id" :value="store.id">{{ store.name }}</option>
        </select>
        <label class="block text-xs font-bold text-slate-600">Produk</label>
        <select v-model="form.product_id" required class="w-full rounded-xl border p-2 text-sm">
          <option value="">Pilih produk</option>
          <option v-for="product in products" :key="product.id" :value="product.id">{{ product.name }}</option>
        </select>
        <label class="block text-xs font-bold text-slate-600">Jumlah</label>
        <input v-model.number="form.quantity" type="number" min="1" required class="w-full rounded-xl border p-2 text-sm" />
        <textarea v-model="form.notes" placeholder="Catatan (opsional)" class="w-full rounded-xl border p-2 text-sm" rows="3" />
        <button :disabled="loading" class="w-full rounded-xl bg-slate-900 text-white py-2.5 text-sm font-bold disabled:opacity-50">
          {{ loading ? 'Mengirim...' : 'Kirim Request' }}
        </button>
      </form>

      <div class="lg:col-span-2 bg-white rounded-2xl p-5 shadow-sm">
        <div class="flex justify-between items-center mb-4">
          <h2 class="font-black text-slate-900">Request Saya</h2>
          <button class="text-xs font-bold text-blue-600" @click="load">Refresh</button>
        </div>
        <div v-if="requests.length === 0" class="text-sm text-slate-400 py-10 text-center">Belum ada request distributor.</div>
        <div v-for="request in requests" :key="request.id" class="border-b py-3 flex justify-between gap-3 text-sm">
          <div><p class="font-bold">{{ request.product?.name || 'Produk' }}</p><p class="text-xs text-slate-500">{{ request.quantity }} {{ request.unit || 'PCS' }}</p></div>
          <span class="h-fit rounded-full bg-amber-100 px-2 py-1 text-[10px] font-black text-amber-700">{{ request.status }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import apiClient from '@/api/axios'
import { useMasterDataStore } from '@/stores/masterData'
import { useSalesTransferStore } from '@/stores/salesTransfer'

const master = useMasterDataStore()
const transferStore = useSalesTransferStore()
const loading = ref(false)
const products = ref<any[]>([])
const requests = ref<any[]>([])
const form = reactive({ distributor_store_id: '', product_id: '', quantity: 1, notes: '' })
const distributors = computed(() => (master.stores || []).filter((store: any) => store.customer_type === 'DISTRIBUTOR'))

async function load() { requests.value = await transferStore.fetchMyDistributorRequests() || [] }
async function submit() {
  loading.value = true
  try { await transferStore.createDistributorRequest({ ...form, type: 'TRANSFER', unit: 'PCS' }); form.quantity = 1; form.notes = ''; await load() }
  finally { loading.value = false }
}
onMounted(async () => {
  await master.fetchStores()
  const response = await apiClient.get('/factory/products')
  products.value = response.data?.data || response.data || []
  await load()
})
</script>
