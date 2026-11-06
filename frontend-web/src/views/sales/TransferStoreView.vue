<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ArrowLeftRight, CheckCircle2, User, Loader2, Store } from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import apiClient from '@/api/axios'
import { toast } from 'vue-sonner'
import { useMasterDataStore } from '@/stores/masterData'

const masterDataStore = useMasterDataStore()

const isLoading = ref(false)
const isTransferring = ref(false)
const salesmen = ref<any[]>([])
const fromEmployeeId = ref('')
const toEmployeeId = ref('')
const fromEmployeeStores = ref(0)

const filteredSalesmen = computed(() => {
  return salesmen.value.filter((emp: any) => {
    const isActive = emp.user?.is_active !== false && emp.employment_status !== 'RESIGNED' && emp.employment_status !== 'TERMINATED'
    
    const branchId = masterDataStore.selectedBranchId
    const empBranchId = emp.operating_branch_id || emp.branch_id
    const belongsToBranch = !branchId || branchId === 'ALL_BRANCHES' || empBranchId === branchId

    const jobTitle = (emp.job_position?.title || '').toLowerCase()
    const isSales = /\b(sales|motoris|booster|marketing|task order|take order|to|mt|kam|modern trade)\b/i.test(jobTitle)

    return isActive && belongsToBranch && isSales
  })
})

const fetchSalesmen = async () => {
  isLoading.value = true
  try {
    const res = await apiClient.get(
      `/employees?limit=all&is_sales=true&company_id=${masterDataStore.selectedBranchCompanyId || ''}&sales_type=${masterDataStore.selectedSalesType}`,
    )
    if (res.data?.data) {
      salesmen.value = res.data.data
    }
  } catch (error) {
    console.error('Failed to fetch salesmen:', error)
    toast.error('Gagal mengambil data salesman')
  } finally {
    isLoading.value = false
  }
}

const checkFromStores = async () => {
  if (!fromEmployeeId.value) {
    fromEmployeeStores.value = 0
    return
  }
  try {
    const res = await apiClient.get(`/stores?assigned_employee_id=${fromEmployeeId.value}&limit=1&company_id=${masterDataStore.selectedBranchCompanyId || ''}`)
    fromEmployeeStores.value = res.data.meta?.total || 0
  } catch (error) {
    console.error('Failed to count stores:', error)
  }
}

const processTransfer = async () => {
  if (!fromEmployeeId.value || !toEmployeeId.value) {
    toast.error('Pilih salesman asal dan tujuan terlebih dahulu')
    return
  }
  if (fromEmployeeId.value === toEmployeeId.value) {
    toast.error('Salesman asal dan tujuan tidak boleh sama')
    return
  }
  if (fromEmployeeStores.value === 0) {
    toast.error('Salesman asal tidak memiliki toko untuk ditransfer')
    return
  }
  
  if (!confirm(`Apakah Anda yakin ingin mentransfer ${fromEmployeeStores.value} toko?`)) {
    return
  }

  isTransferring.value = true
  try {
    const res = await apiClient.post('/stores/transfer', {
      from_employee_id: fromEmployeeId.value,
      to_employee_id: toEmployeeId.value
    })
    
    toast.success(res.data.message || `Berhasil mentransfer ${res.data.data?.stores_transferred} toko`)
    fromEmployeeId.value = ''
    toEmployeeId.value = ''
    fromEmployeeStores.value = 0
  } catch (error: any) {
    console.error('Transfer failed:', error)
    toast.error(error.response?.data?.error || 'Gagal melakukan transfer toko')
  } finally {
    isTransferring.value = false
  }
}

onMounted(() => {
  fetchSalesmen()
})
</script>

<template>
  <div class="space-y-6">
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
      <div>
        <h1 class="text-2xl font-bold tracking-tight text-slate-900 flex items-center flex-wrap gap-2">
          <span>Transfer Toko</span>
          <span v-if="masterDataStore.selectedBranchName" class="text-primary text-xl font-bold">
            — {{ masterDataStore.selectedBranchName }}
          </span>
        </h1>
        <p class="text-sm text-slate-500 mt-1">Pindahkan kepemilikan seluruh toko dari satu salesman ke salesman lainnya.</p>
      </div>
    </div>

    <div class="bg-white rounded-2xl border border-slate-200/60 shadow-sm overflow-hidden p-6 max-w-3xl">
      <div v-if="isLoading" class="flex justify-center p-8">
        <Loader2 class="w-8 h-8 animate-spin text-primary" />
      </div>
      
      <div v-else class="space-y-6">
        <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
          <!-- Asal -->
          <div class="space-y-3 p-4 bg-slate-50 rounded-xl border border-slate-100">
            <label class="text-sm font-semibold text-slate-700 flex items-center gap-2">
              <User class="w-4 h-4 text-slate-500" />
              Salesman Asal
            </label>
            <select
              v-model="fromEmployeeId"
              @change="checkFromStores"
              class="w-full text-sm border-slate-200 rounded-xl focus:border-primary focus:ring-primary py-2.5 px-3 bg-white"
            >
              <option value="">-- Pilih Salesman Asal --</option>
              <option v-for="emp in filteredSalesmen" :key="emp.id" :value="emp.id" :disabled="emp.id === toEmployeeId">
                {{ emp.first_name }} {{ emp.last_name }} ({{ emp.job_position?.title || 'No Role' }})
              </option>
            </select>
            
            <div v-if="fromEmployeeId" class="mt-2 text-sm text-slate-600 flex items-center gap-2 bg-white p-2 rounded-lg border border-slate-200">
              <Store class="w-4 h-4 text-primary" />
              Total Toko: <span class="font-bold text-slate-900">{{ fromEmployeeStores }}</span>
            </div>
          </div>
          
          <!-- Tujuan -->
          <div class="space-y-3 p-4 bg-primary/5 rounded-xl border border-primary/10">
            <label class="text-sm font-semibold text-primary flex items-center gap-2">
              <CheckCircle2 class="w-4 h-4" />
              Salesman Tujuan
            </label>
            <select
              v-model="toEmployeeId"
              class="w-full text-sm border-primary/20 rounded-xl focus:border-primary focus:ring-primary py-2.5 px-3 bg-white"
            >
              <option value="">-- Pilih Salesman Tujuan --</option>
              <option v-for="emp in filteredSalesmen" :key="emp.id" :value="emp.id" :disabled="emp.id === fromEmployeeId">
                {{ emp.first_name }} {{ emp.last_name }} ({{ emp.job_position?.title || 'No Role' }})
              </option>
            </select>
          </div>
        </div>

        <div class="flex justify-end pt-4 border-t border-slate-100">
          <Button 
            @click="processTransfer" 
            :disabled="!fromEmployeeId || !toEmployeeId || fromEmployeeStores === 0 || isTransferring"
            class="bg-primary hover:bg-primary/90 text-white gap-2 px-6"
          >
            <Loader2 v-if="isTransferring" class="w-4 h-4 animate-spin" />
            <ArrowLeftRight v-else class="w-4 h-4" />
            {{ isTransferring ? 'Memproses Transfer...' : 'Proses Transfer Toko' }}
          </Button>
        </div>
      </div>
    </div>
  </div>
</template>
