<script setup lang="ts">
import { ref, onMounted, computed, watch } from 'vue'
import apiClient from '@/api/axios'
import { toast } from 'vue-sonner'
import { useMasterDataStore } from '@/stores/masterData'
import {
  Search,
  History,
  Eye,
  X,
  FileSpreadsheet,
  AlertCircle,
  Activity,
  PlusCircle,
  RefreshCw,
  Trash2,
  Filter,
} from 'lucide-vue-next'
import { Button } from '@/components/ui/button'

const masterDataStore = useMasterDataStore()
const isLoading = ref(true)
const logsList = ref<any[]>([])
const searchQuery = ref('')
const selectedAction = ref('ALL')
const selectedEntity = ref('Store,SalesOrder,SalesTransaction')
const selectedLog = ref<any>(null)
const showDetailModal = ref(false)

const currentPage = ref(1)
const limit = ref(20)
const totalItems = ref(0)
const totalPages = computed(() => Math.ceil(totalItems.value / limit.value))

const entityOptions = [
  { value: 'Store,SalesOrder,SalesTransaction', label: 'Semua Entity Sales' },
  { value: 'Store', label: 'Toko (Customer)' },
  { value: 'SalesOrder', label: 'Pesanan Sales (SO)' },
  { value: 'SalesTransaction', label: 'Transaksi Penjualan' },
]

const actionOptions = [
  { value: 'ALL', label: 'Semua Aksi' },
  { value: 'CREATE', label: 'CREATE (Penambahan)' },
  { value: 'UPDATE', label: 'UPDATE (Perubahan)' },
  { value: 'DELETE', label: 'DELETE (Penghapusan)' },
]

// Stats counters
const stats = computed(() => {
  let createCount = 0
  let updateCount = 0
  let deleteCount = 0

  logsList.value.forEach((log) => {
    if (log.action === 'CREATE') createCount++
    else if (log.action === 'UPDATE') updateCount++
    else if (log.action === 'DELETE') deleteCount++
  })

  return {
    total: logsList.value.length,
    create: createCount,
    update: updateCount,
    delete: deleteCount,
  }
})

// Watch filters to reset page and fetch
watch([selectedAction, selectedEntity], () => {
  currentPage.value = 1
  fetchAuditLogs()
})

let searchTimeout: any = null
watch(searchQuery, () => {
  if (searchTimeout) clearTimeout(searchTimeout)
  searchTimeout = setTimeout(() => {
    currentPage.value = 1
    fetchAuditLogs()
  }, 500)
})

const fetchAuditLogs = async () => {
  isLoading.value = true
  try {
    const params = new URLSearchParams()
    if (selectedEntity.value !== 'ALL') params.append('entity', selectedEntity.value)
    if (selectedAction.value !== 'ALL') params.append('action', selectedAction.value)
    if (searchQuery.value) params.append('search', searchQuery.value)
    
    params.append('page', currentPage.value.toString())
    params.append('limit', limit.value.toString())

    const res = await apiClient.get(`/sales-marketing/audit-logs?${params.toString()}`)
    // backend response has success: true, message: ..., data: [...], meta: {...}
    logsList.value = res.data?.data || []
    totalItems.value = res.data?.meta?.total || 0
  } catch (error) {
    console.error('Failed to fetch audit logs:', error)
    toast.error('Gagal mengambil data log audit')
  } finally {
    isLoading.value = false
  }
}

const changePage = (page: number) => {
  if (page >= 1 && page <= totalPages.value) {
    currentPage.value = page
    fetchAuditLogs()
  }
}

const formatDate = (dateStr: string) => {
  if (!dateStr) return '-'
  try {
    const d = new Date(dateStr)
    return d.toLocaleString('id-ID', {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit',
    })
  } catch (e) {
    return dateStr
  }
}

const getActionBadgeClass = (action: string) => {
  switch (action) {
    case 'CREATE':
      return 'bg-emerald-50 text-emerald-700 border-emerald-200'
    case 'UPDATE':
      return 'bg-blue-50 text-blue-700 border-blue-200'
    case 'DELETE':
      return 'bg-rose-50 text-rose-700 border-rose-200'
    default:
      return 'bg-slate-50 text-slate-700 border-slate-200'
  }
}

const openDetail = (log: any) => {
  selectedLog.value = log
  showDetailModal.value = true
}

const closeDetail = () => {
  showDetailModal.value = false
  selectedLog.value = null
}

const parseJSON = (str: string) => {
  if (!str) return null
  try {
    // If it's already an object (which shouldn't happen but safe-guard)
    if (typeof str === 'object') return str
    return JSON.parse(str)
  } catch (e) {
    return null
  }
}

// Convert JSON object into a flat key-value list for friendly UI display
const getDiffItems = (oldVal: string, newVal: string) => {
  const oldObj = parseJSON(oldVal) || {}
  const newObj = parseJSON(newVal) || {}

  // Gather all unique keys
  const keys = Array.from(new Set([...Object.keys(oldObj), ...Object.keys(newObj)]))
  
  // Filter out system or huge objects if any
  const excludeKeys = ['created_at', 'updated_at', 'deleted_at', 'password', 'token']

  return keys
    .filter((k) => !excludeKeys.includes(k.toLowerCase()))
    .map((key) => {
      const oVal = oldObj[key]
      const nVal = newObj[key]
      
      let isChanged = false
      if (typeof oVal === 'object' || typeof nVal === 'object') {
        isChanged = JSON.stringify(oVal) !== JSON.stringify(nVal)
      } else {
        isChanged = oVal !== nVal
      }

      return {
        key,
        oldValue: oVal !== undefined ? (typeof oVal === 'object' ? JSON.stringify(oVal) : String(oVal)) : '-',
        newValue: nVal !== undefined ? (typeof nVal === 'object' ? JSON.stringify(nVal) : String(nVal)) : '-',
        isChanged,
      }
    })
}

// Watch selected branch company to refetch
watch(
  () => masterDataStore.selectedBranchId,
  () => {
    fetchAuditLogs()
  }
)

onMounted(() => {
  fetchAuditLogs()
})
</script>

<template>
  <div class="space-y-6">
    <!-- Header -->
    <div class="flex flex-col md:flex-row md:items-center md:justify-between gap-4">
      <div>
        <h1 class="text-3xl font-extrabold tracking-tight text-slate-900 flex items-center flex-wrap gap-2">
          <History class="w-8 h-8 text-indigo-600" />
          <span>Audit Trail & Activity Log</span>
          <span v-if="masterDataStore.selectedBranchName" class="text-primary text-2xl font-bold">
            — {{ masterDataStore.selectedBranchName }}
          </span>
        </h1>
        <p class="text-slate-500 mt-1">
          Pantau riwayat operasi kritis (CREATE, UPDATE, DELETE) pada modul Sales & Marketing untuk jaminan akuntabilitas data.
        </p>
      </div>
      
      <div class="flex items-center gap-2">
        <Button
          variant="outline"
          class="flex items-center gap-2 bg-white"
          @click="fetchAuditLogs"
        >
          <RefreshCw class="w-4 h-4" :class="{ 'animate-spin': isLoading }" />
          Refresh
        </Button>
      </div>
    </div>

    <!-- Stats Section -->
    <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
      <!-- Total Logs Card -->
      <div class="bg-white p-5 rounded-2xl border border-slate-200 shadow-sm flex items-center justify-between">
        <div class="space-y-1">
          <span class="text-xs font-semibold uppercase tracking-wider text-slate-400">Total Log Audit</span>
          <h3 class="text-2xl font-black text-slate-800">{{ totalItems }}</h3>
        </div>
        <div class="p-3 bg-indigo-50 text-indigo-600 rounded-xl">
          <Activity class="w-6 h-6" />
        </div>
      </div>

      <!-- Creates Card -->
      <div class="bg-white p-5 rounded-2xl border border-slate-200 shadow-sm flex items-center justify-between">
        <div class="space-y-1">
          <span class="text-xs font-semibold uppercase tracking-wider text-slate-400">Penambahan (CREATE)</span>
          <h3 class="text-2xl font-black text-emerald-600">{{ stats.create }}</h3>
        </div>
        <div class="p-3 bg-emerald-50 text-emerald-600 rounded-xl">
          <PlusCircle class="w-6 h-6" />
        </div>
      </div>

      <!-- Updates Card -->
      <div class="bg-white p-5 rounded-2xl border border-slate-200 shadow-sm flex items-center justify-between">
        <div class="space-y-1">
          <span class="text-xs font-semibold uppercase tracking-wider text-slate-400">Perubahan (UPDATE)</span>
          <h3 class="text-2xl font-black text-blue-600">{{ stats.update }}</h3>
        </div>
        <div class="p-3 bg-blue-50 text-blue-600 rounded-xl">
          <RefreshCw class="w-6 h-6" />
        </div>
      </div>

      <!-- Deletes Card -->
      <div class="bg-white p-5 rounded-2xl border border-slate-200 shadow-sm flex items-center justify-between">
        <div class="space-y-1">
          <span class="text-xs font-semibold uppercase tracking-wider text-slate-400">Penghapusan (DELETE)</span>
          <h3 class="text-2xl font-black text-rose-600">{{ stats.delete }}</h3>
        </div>
        <div class="p-3 bg-rose-50 text-rose-600 rounded-xl">
          <Trash2 class="w-6 h-6" />
        </div>
      </div>
    </div>

    <!-- Filters Area -->
    <div class="bg-white p-4 rounded-2xl border border-slate-200 shadow-sm flex flex-col md:flex-row gap-4 items-center justify-between">
      <!-- Search Input -->
      <div class="relative w-full md:w-80">
        <Search class="absolute left-3.5 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400" />
        <input
          v-model="searchQuery"
          type="text"
          placeholder="Cari user, deskripsi, IP..."
          class="w-full pl-10 pr-4 py-2 border border-slate-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500/20 focus:border-indigo-500 transition-all bg-slate-50/50"
        />
      </div>

      <!-- Dropdown Filters -->
      <div class="flex flex-wrap items-center gap-3 w-full md:w-auto">
        <!-- Action Filter -->
        <div class="flex items-center gap-2">
          <Filter class="w-4 h-4 text-slate-400 shrink-0" />
          <select
            v-model="selectedAction"
            class="px-3 py-2 border border-slate-200 rounded-xl text-sm bg-white focus:outline-none focus:ring-2 focus:ring-indigo-500/20 focus:border-indigo-500 transition-all cursor-pointer"
          >
            <option v-for="opt in actionOptions" :key="opt.value" :value="opt.value">
              {{ opt.label }}
            </option>
          </select>
        </div>

        <!-- Entity Filter -->
        <select
          v-model="selectedEntity"
          class="px-3 py-2 border border-slate-200 rounded-xl text-sm bg-white focus:outline-none focus:ring-2 focus:ring-indigo-500/20 focus:border-indigo-500 transition-all cursor-pointer"
        >
          <option v-for="opt in entityOptions" :key="opt.value" :value="opt.value">
            {{ opt.label }}
          </option>
        </select>
      </div>
    </div>

    <!-- Logs Table -->
    <div class="bg-white rounded-2xl border border-slate-200 shadow-sm overflow-hidden">
      <div v-if="isLoading" class="p-12 flex flex-col items-center justify-center gap-3 text-slate-500">
        <RefreshCw class="w-8 h-8 text-indigo-600 animate-spin" />
        <p class="text-sm font-medium">Memuat log audit...</p>
      </div>

      <div v-else-if="logsList.length === 0" class="p-12 flex flex-col items-center justify-center gap-2 text-slate-400">
        <AlertCircle class="w-10 h-10 text-slate-300" />
        <p class="text-sm font-medium">Tidak ada data log audit ditemukan</p>
      </div>

      <div v-else class="overflow-x-auto">
        <table class="w-full text-left border-collapse text-sm">
          <thead>
            <tr class="bg-slate-50/75 border-b border-slate-100 text-slate-600 font-semibold">
              <th class="p-4 w-44">Waktu</th>
              <th class="p-4">User Admin</th>
              <th class="p-4 w-28">Aksi</th>
              <th class="p-4">Target Objek</th>
              <th class="p-4">Deskripsi Perubahan</th>
              <th class="p-4 w-32 text-center">Aksi</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-100 text-slate-700">
            <tr v-for="log in logsList" :key="log.id" class="hover:bg-slate-50/50 transition-colors">
              <!-- Date Time -->
              <td class="p-4 font-medium text-slate-900 whitespace-nowrap">
                {{ formatDate(log.created_at) }}
              </td>
              <!-- User / IP -->
              <td class="p-4">
                <div class="font-semibold text-slate-800">{{ log.user_name || 'System' }}</div>
                <div class="text-xs text-slate-400 font-mono">{{ log.ip_address || '0.0.0.0' }}</div>
              </td>
              <!-- Action Badge -->
              <td class="p-4 whitespace-nowrap">
                <span
                  class="px-2.5 py-1 rounded-full text-xs font-bold border"
                  :class="getActionBadgeClass(log.action)"
                >
                  {{ log.action }}
                </span>
              </td>
              <!-- Entity Target -->
              <td class="p-4">
                <span class="font-medium text-indigo-600 bg-indigo-50/50 px-2 py-0.5 rounded text-xs">
                  {{ log.entity }}
                </span>
                <div class="text-xs text-slate-400 font-mono mt-1 select-all" title="Click to copy">
                  ID: {{ log.entity_id }}
                </div>
              </td>
              <!-- Description -->
              <td class="p-4">
                <p class="line-clamp-2 text-slate-600" :title="log.description">
                  {{ log.description }}
                </p>
              </td>
              <!-- Action Button -->
              <td class="p-4 text-center">
                <Button
                  size="sm"
                  variant="outline"
                  class="inline-flex items-center gap-1.5 bg-white border-slate-200 text-slate-600 hover:text-indigo-600 hover:border-indigo-200"
                  @click="openDetail(log)"
                >
                  <Eye class="w-3.5 h-3.5" />
                  Detail
                </Button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- Pagination Footer -->
      <div v-if="totalPages > 1" class="bg-slate-50 p-4 border-t border-slate-200 flex items-center justify-between">
        <div class="text-sm text-slate-500">
          Menampilkan <span class="font-black text-slate-900">{{ (currentPage - 1) * limit + 1 }}</span> sampai
          <span class="font-black text-slate-900">
            {{ Math.min(currentPage * limit, totalItems) }}
          </span>
          dari <span class="font-black text-slate-900">{{ totalItems }}</span> entri
        </div>
        
        <div class="flex items-center gap-1.5">
          <Button
            variant="outline"
            size="sm"
            class="h-8 bg-white border-slate-200 hover:bg-slate-50 hover:text-indigo-600 disabled:opacity-50"
            @click="changePage(currentPage - 1)"
            :disabled="currentPage === 1"
          >
            Sebelumnya
          </Button>

          <div class="hidden sm:flex items-center gap-1">
            <Button
              v-for="page in totalPages"
              :key="page"
              variant="outline"
              size="sm"
              class="w-8 h-8 p-0"
              :class="
                page === currentPage
                  ? 'bg-indigo-600 text-white border-indigo-600 hover:bg-indigo-700 hover:text-white'
                  : 'bg-white border-slate-200 text-slate-600 hover:border-indigo-200 hover:text-indigo-600'
              "
              @click="changePage(page)"
            >
              {{ page }}
            </Button>
          </div>

          <Button
            variant="outline"
            size="sm"
            class="h-8 bg-white border-slate-200 hover:bg-slate-50 hover:text-indigo-600 disabled:opacity-50"
            @click="changePage(currentPage + 1)"
            :disabled="currentPage === totalPages"
          >
            Selanjutnya
          </Button>
        </div>
      </div>
    </div>

    <!-- Detail Audit Modal -->
    <div v-if="showDetailModal && selectedLog" class="fixed inset-0 z-50 flex items-center justify-center p-4">
      <!-- Backdrop -->
      <div class="fixed inset-0 bg-slate-900/60 backdrop-blur-sm" @click="closeDetail" />

      <!-- Modal Card -->
      <div class="bg-white rounded-2xl shadow-2xl border border-slate-100 max-w-4xl w-full z-10 overflow-hidden animate-in fade-in-50 zoom-in-95 duration-200">
        <!-- Modal Header -->
        <div class="p-5 border-b border-slate-100 flex items-center justify-between bg-slate-50/50">
          <div class="flex items-center gap-3">
            <span
              class="px-3 py-1 rounded-full text-xs font-extrabold border"
              :class="getActionBadgeClass(selectedLog.action)"
            >
              {{ selectedLog.action }}
            </span>
            <h3 class="text-lg font-bold text-slate-900">
              Audit Detail - {{ selectedLog.entity }}
            </h3>
          </div>
          <button @click="closeDetail" class="text-slate-400 hover:text-slate-600 p-1.5 rounded-lg hover:bg-slate-100 transition-all">
            <X class="w-5 h-5" />
          </button>
        </div>

        <!-- Modal Body -->
        <div class="p-6 overflow-y-auto max-h-[75vh] space-y-6">
          <!-- Metadata Info Block -->
          <div class="grid grid-cols-1 md:grid-cols-3 gap-4 bg-slate-50 p-4 rounded-xl text-sm border border-slate-100">
            <div>
              <span class="text-xs text-slate-400 block uppercase font-bold tracking-wider">Pelaku (Admin)</span>
              <span class="font-semibold text-slate-800">{{ selectedLog.user_name }}</span>
            </div>
            <div>
              <span class="text-xs text-slate-400 block uppercase font-bold tracking-wider">Waktu Aktivitas</span>
              <span class="font-semibold text-slate-800">{{ formatDate(selectedLog.created_at) }}</span>
            </div>
            <div>
              <span class="text-xs text-slate-400 block uppercase font-bold tracking-wider">Alamat IP / Lokasi</span>
              <span class="font-mono text-slate-700">{{ selectedLog.ip_address || '0.0.0.0' }}</span>
            </div>
          </div>

          <!-- Description -->
          <div class="space-y-1">
            <span class="text-xs text-slate-400 block uppercase font-bold tracking-wider">Aktivitas</span>
            <p class="text-slate-800 font-medium text-base bg-indigo-50/30 p-3 rounded-lg border border-indigo-100/50">
              {{ selectedLog.description }}
            </p>
          </div>

          <!-- Changes Diff (Before & After comparison) -->
          <div class="space-y-3">
            <div class="flex items-center justify-between">
              <span class="text-xs text-slate-400 block uppercase font-bold tracking-wider">Perbandingan Data</span>
              <span class="text-xs text-indigo-500 font-medium">* Menampilkan perubahan atribut objek</span>
            </div>

            <!-- If there is old_value or new_value logs (UPDATE/DELETE/CREATE) -->
            <div v-if="selectedLog.old_value || selectedLog.new_value" class="border border-slate-150 rounded-xl overflow-hidden shadow-inner bg-slate-50">
              <!-- Check if it is JSON valid or plain text -->
              <div v-if="parseJSON(selectedLog.old_value || selectedLog.new_value)" class="divide-y divide-slate-200 max-h-96 overflow-y-auto">
                <table class="w-full text-left text-xs border-collapse">
                  <thead>
                    <tr class="bg-slate-100 text-slate-600 font-bold border-b border-slate-200">
                      <th class="p-3 w-1/4">Atribut / Kolom</th>
                      <th class="p-3 w-3/8 bg-rose-50/50">Nilai Sebelum (Old)</th>
                      <th class="p-3 w-3/8 bg-emerald-50/50">Nilai Sesudah (New)</th>
                    </tr>
                  </thead>
                  <tbody class="divide-y divide-slate-150 bg-white">
                    <tr
                      v-for="item in getDiffItems(selectedLog.old_value, selectedLog.new_value)"
                      :key="item.key"
                      :class="{ 'bg-yellow-50/40 font-medium': item.isChanged }"
                    >
                      <td class="p-3 font-mono text-slate-600 font-semibold">{{ item.key }}</td>
                      <!-- Old Value -->
                      <td
                        class="p-3 font-mono text-slate-700 whitespace-pre-wrap break-all"
                        :class="{ 'bg-rose-50/30 text-rose-800 line-through': item.isChanged && item.oldValue !== '-' }"
                      >
                        {{ item.oldValue }}
                      </td>
                      <!-- New Value -->
                      <td
                        class="p-3 font-mono text-slate-700 whitespace-pre-wrap break-all"
                        :class="{ 'bg-emerald-50/40 text-emerald-800 font-semibold': item.isChanged && item.newValue !== '-' }"
                      >
                        {{ item.newValue }}
                      </td>
                    </tr>
                  </tbody>
                </table>
              </div>

              <!-- Plain Text Diff block fallback -->
              <div v-else class="grid grid-cols-1 md:grid-cols-2 divide-y md:divide-y-0 md:divide-x divide-slate-200 font-mono text-xs p-4">
                <div class="pb-3 md:pb-0 md:pr-4">
                  <div class="text-rose-600 font-bold mb-2">SEBELUM:</div>
                  <pre class="bg-slate-800 text-slate-100 p-3 rounded-lg overflow-x-auto whitespace-pre-wrap">{{ selectedLog.old_value || 'None (Kosong)' }}</pre>
                </div>
                <div class="pt-3 md:pt-0 md:pl-4">
                  <div class="text-emerald-600 font-bold mb-2">SESUDAH:</div>
                  <pre class="bg-slate-800 text-slate-100 p-3 rounded-lg overflow-x-auto whitespace-pre-wrap">{{ selectedLog.new_value || 'None (Kosong)' }}</pre>
                </div>
              </div>
            </div>

            <div v-else class="text-center p-8 bg-slate-50 border border-dashed border-slate-200 rounded-xl text-slate-400 text-sm">
              Tidak ada data perbandingan sebelum dan sesudah untuk aksi ini
            </div>
          </div>
        </div>

        <!-- Modal Footer -->
        <div class="p-4 border-t border-slate-100 flex justify-end bg-slate-50/50">
          <Button @click="closeDetail" class="bg-indigo-600 hover:bg-indigo-700 text-white shadow">
            Tutup
          </Button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* Scrollbar adjustments */
.max-h-96::-webkit-scrollbar {
  width: 6px;
}
.max-h-96::-webkit-scrollbar-track {
  background: transparent;
}
.max-h-96::-webkit-scrollbar-thumb {
  background: #cbd5e1;
  border-radius: 3px;
}
.max-h-96::-webkit-scrollbar-thumb:hover {
  background: #94a3b8;
}
</style>
