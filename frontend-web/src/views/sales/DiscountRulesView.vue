<template>
  <div class="p-6">
    <div class="flex justify-between items-center mb-6">
      <h1 class="text-2xl font-bold text-gray-900 flex items-center flex-wrap gap-2">
        <span>Pengaturan Diskon Penjualan</span>
        <span v-if="masterDataStore.selectedBranchName" class="text-primary text-xl font-bold">
          — {{ masterDataStore.selectedBranchName }}
        </span>
      </h1>
      <button 
        @click="openCreateModal"
        class="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 transition"
      >
        + Buat Aturan Diskon
      </button>
    </div>

    <!-- Loading State -->
    <div v-if="loading" class="flex justify-center py-8">
      <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
    </div>

    <!-- Error State -->
    <div v-else-if="error" class="bg-red-50 text-red-600 p-4 rounded-md mb-6">
      {{ error }}
    </div>

    <!-- Data List -->
    <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
      <div 
        v-for="rule in rules" 
        :key="rule.id" 
        class="bg-white rounded-lg shadow-sm border p-5 flex flex-col"
      >
        <div class="flex justify-between items-start mb-4">
          <div>
            <h3 class="text-lg font-semibold text-gray-800">{{ rule.name }}</h3>
            <span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium mt-1"
                  :class="rule.is_active ? 'bg-green-100 text-green-800' : 'bg-gray-100 text-gray-800'">
              {{ rule.is_active ? 'Aktif' : 'Tidak Aktif' }}
            </span>
          </div>
          <div class="flex space-x-2">
            <button @click="editRule(rule)" class="text-gray-400 hover:text-blue-600">Edit</button>
            <button @click="deleteRule(rule.id)" class="text-gray-400 hover:text-red-600">Hapus</button>
          </div>
        </div>

        <div class="space-y-3 flex-1 text-sm text-gray-600">
          <p><strong>Metode Bayar:</strong> {{ rule.payment_method_required || 'Semua' }}</p>
          
          <div>
            <strong>Tingkatan Diskon (Tiers):</strong>
            <ul class="list-disc list-inside mt-1 ml-2">
              <li v-for="tier in rule.tiers" :key="tier.id">
                Min. {{ tier.min_qty }} Item &rarr; Diskon {{ tier.discount_percent }}%
              </li>
              <li v-if="!rule.tiers || rule.tiers.length === 0" class="text-gray-400 italic">
                Belum ada tier.
              </li>
            </ul>
          </div>
          
          <div>
            <strong>Pengecualian:</strong>
            <p v-if="rule.excluded_products?.length > 0" class="mt-1 ml-2">
              {{ rule.excluded_products.length }} Produk dikecualikan.
            </p>
            <p v-else class="text-gray-400 italic ml-2">Tidak ada</p>
          </div>
        </div>
      </div>
      
      <div v-if="rules.length === 0" class="col-span-full text-center py-12 text-gray-500 bg-gray-50 rounded-lg border border-dashed">
        Belum ada aturan diskon yang dibuat.
      </div>
    </div>
    <!-- Form Modal -->
    <div v-if="showModal" class="fixed inset-0 z-50 overflow-y-auto" aria-labelledby="modal-title" role="dialog" aria-modal="true">
      <div class="flex items-end justify-center min-h-screen pt-4 px-4 pb-20 text-center sm:block sm:p-0">
        <div class="fixed inset-0 bg-gray-500 bg-opacity-75 transition-opacity" @click="closeModal"></div>

        <span class="hidden sm:inline-block sm:align-middle sm:h-screen" aria-hidden="true">&#8203;</span>

        <div class="inline-block align-bottom bg-white rounded-lg text-left overflow-hidden shadow-xl transform transition-all sm:my-8 sm:align-middle sm:max-w-2xl sm:w-full">
          <div class="bg-white px-4 pt-5 pb-4 sm:p-6 sm:pb-4">
            <h3 class="text-lg leading-6 font-medium text-gray-900 mb-4" id="modal-title">
              {{ editingRule ? 'Edit Aturan Diskon' : 'Buat Aturan Diskon' }}
            </h3>
            
            <form @submit.prevent="saveRule" class="space-y-4">
              <div>
                <label class="block text-sm font-medium text-gray-700">Nama Aturan</label>
                <input v-model="form.name" required type="text" class="mt-1 block w-full border border-gray-300 rounded-md shadow-sm py-2 px-3 focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm">
              </div>

              <div>
                <label class="block text-sm font-medium text-gray-700">Metode Pembayaran (opsional)</label>
                <select v-model="form.payment_method_required" class="mt-1 block w-full border border-gray-300 rounded-md shadow-sm py-2 px-3 focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm">
                  <option value="">Semua Metode</option>
                  <option value="CASH">CASH</option>
                  <option value="TEMPO">TEMPO</option>
                </select>
              </div>

              <div class="flex items-center">
                <input id="is_active" v-model="form.is_active" type="checkbox" class="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded">
                <label for="is_active" class="ml-2 block text-sm text-gray-900">Aktifkan Aturan Ini</label>
              </div>

              <div class="pt-4 border-t border-gray-200">
                <div class="flex justify-between items-center mb-2">
                  <h4 class="text-sm font-medium text-gray-900">Tingkatan Diskon (Tiers)</h4>
                  <button type="button" @click="addTier" class="text-sm text-blue-600 hover:text-blue-800">+ Tambah Tier</button>
                </div>
                
                <div v-for="(tier, index) in form.tiers" :key="index" class="flex items-center space-x-2 mb-2">
                  <div class="flex-1">
                    <label class="block text-xs text-gray-500">Min. Item (Qty)</label>
                    <input v-model.number="tier.min_qty" type="number" required min="1" class="block w-full border border-gray-300 rounded-md shadow-sm py-1.5 px-3 sm:text-sm">
                  </div>
                  <div class="flex-1">
                    <label class="block text-xs text-gray-500">Diskon (%)</label>
                    <input v-model.number="tier.discount_percent" type="number" required min="0" step="0.1" class="block w-full border border-gray-300 rounded-md shadow-sm py-1.5 px-3 sm:text-sm">
                  </div>
                  <button type="button" @click="removeTier(index)" class="mt-5 text-red-500 hover:text-red-700 font-bold px-2">&times;</button>
                </div>
              </div>
            </form>
          </div>
          <div class="bg-gray-50 px-4 py-3 sm:px-6 sm:flex sm:flex-row-reverse">
            <button @click="saveRule" type="button" class="w-full inline-flex justify-center rounded-md border border-transparent shadow-sm px-4 py-2 bg-blue-600 text-base font-medium text-white hover:bg-blue-700 focus:outline-none sm:ml-3 sm:w-auto sm:text-sm">
              Simpan
            </button>
            <button @click="closeModal" type="button" class="mt-3 w-full inline-flex justify-center rounded-md border border-gray-300 shadow-sm px-4 py-2 bg-white text-base font-medium text-gray-700 hover:bg-gray-50 focus:outline-none sm:mt-0 sm:ml-3 sm:w-auto sm:text-sm">
              Batal
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import api from '@/api/axios'
import { useMasterDataStore } from '@/stores/masterData'

const masterDataStore = useMasterDataStore()

interface DiscountTier {
  id: string
  min_qty: number
  discount_percent: number
}

interface DiscountRule {
  id: string
  name: string
  payment_method_required: string
  is_active: boolean
  tiers: DiscountTier[]
  excluded_products: any[]
}

const rules = ref<DiscountRule[]>([])
const loading = ref(true)
const error = ref('')

const showModal = ref(false)
const editingRule = ref<DiscountRule | null>(null)
const form = ref<Partial<DiscountRule>>({
  name: '',
  payment_method_required: 'CASH',
  is_active: true,
  tiers: []
})

const fetchRules = async () => {
  loading.value = true
  error.value = ''
  try {
    const response = await api.get('/api/v1/discount-rules')
    rules.value = response.data
  } catch (err: any) {
    error.value = err?.response?.data?.error || 'Gagal mengambil data diskon'
  } finally {
    loading.value = false
  }
}

const openCreateModal = () => {
  editingRule.value = null
  form.value = {
    name: '',
    payment_method_required: 'CASH',
    is_active: true,
    tiers: [{ id: '', min_qty: 10, discount_percent: 1 }]
  }
  showModal.value = true
}

const editRule = (rule: DiscountRule) => {
  editingRule.value = rule
  // Clone to avoid reactive mutations before save
  form.value = JSON.parse(JSON.stringify(rule))
  if (!form.value.tiers) {
    form.value.tiers = []
  }
  showModal.value = true
}

const closeModal = () => {
  showModal.value = false
}

const addTier = () => {
  form.value.tiers!.push({ min_qty: 1, discount_percent: 0, id: '' })
}

const removeTier = (index: number) => {
  form.value.tiers!.splice(index, 1)
}

const saveRule = async () => {
  try {
    if (editingRule.value) {
      await api.put(`/api/v1/discount-rules/${editingRule.value.id}`, form.value)
    } else {
      await api.post('/api/v1/discount-rules', form.value)
    }
    showModal.value = false
    fetchRules()
  } catch (err: any) {
    alert('Gagal menyimpan data: ' + err?.response?.data?.error)
  }
}

const deleteRule = async (id: string) => {
  if (confirm('Yakin ingin menghapus aturan ini?')) {
    try {
      await api.delete(`/api/v1/discount-rules/${id}`)
      fetchRules()
    } catch (err: any) {
      alert('Gagal menghapus: ' + err?.response?.data?.error)
    }
  }
}

onMounted(() => {
  fetchRules()
})
</script>
