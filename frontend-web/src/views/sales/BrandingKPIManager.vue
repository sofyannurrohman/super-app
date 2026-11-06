<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import apiClient from '@/api/axios'
import { toast } from 'vue-sonner'
import {
  Target,
  Award,
  PenTool,
  UserCircle,
  RefreshCw,
  CheckCircle2,
  XCircle,
  Calendar,
  Sparkles,
} from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import { useMasterDataStore } from '@/stores/masterData'

const masterDataStore = useMasterDataStore()

const isLoading = ref(true)
const activeTab = ref<'editor' | 'operator'>('editor')
const selectedDate = ref(new Date().toISOString().split('T')[0])

const editorKPIs = ref<any[]>([])
const operatorKPIs = ref<any[]>([])

const setToday = () => {
  selectedDate.value = new Date().toISOString().split('T')[0]
  fetchData()
}

const setYesterday = () => {
  const d = new Date()
  d.setDate(d.getDate() - 1)
  selectedDate.value = d.toISOString().split('T')[0]
  fetchData()
}

const fetchEditorKPIs = async () => {
  try {
    const res = await apiClient.get(`/branding/editor/performance?date=${selectedDate.value}`)
    if (res.data?.data) {
      editorKPIs.value = res.data.data
    } else {
      editorKPIs.value = []
    }
  } catch (error) {
    console.error('Gagal mengambil KPI Editor', error)
    toast.error('Gagal mengambil KPI Editor')
  }
}

const fetchOperatorKPIs = async () => {
  try {
    const res = await apiClient.get('/branding/operator/performance')
    if (res.data?.data) {
      operatorKPIs.value = res.data.data
    } else {
      operatorKPIs.value = []
    }
  } catch (error) {
    console.error('Gagal mengambil KPI Operator', error)
    toast.error('Gagal mengambil KPI Operator')
  }
}

const fetchData = async () => {
  isLoading.value = true
  if (activeTab.value === 'editor') {
    await fetchEditorKPIs()
  } else {
    await fetchOperatorKPIs()
  }
  isLoading.value = false
}

watch(() => masterDataStore.selectedBrandingBranchId, () => {
  fetchData()
})

onMounted(() => {
  fetchData()
})

const switchTab = (tab: 'editor' | 'operator') => {
  activeTab.value = tab
  fetchData()
}

const formatCurrency = (val: number) => {
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    maximumFractionDigits: 0,
  }).format(val || 0)
}
</script>

<template>
  <div class="space-y-8 pb-10">
    <!-- Header -->
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-6">
      <div>
        <h1 class="text-3xl font-extrabold text-slate-900 tracking-tight flex items-center flex-wrap gap-2">
          <Target class="w-8 h-8 text-primary" />
          <span>Branding KPI Manager</span>
          <span v-if="masterDataStore.selectedBrandingBranchName" class="text-primary text-2xl font-bold">
            — {{ masterDataStore.selectedBrandingBranchName }}
          </span>
        </h1>
        <p class="text-[15px] font-medium text-slate-500 mt-1.5 flex items-center gap-2">
          <Award class="w-4 h-4 text-amber-500" />
          Pantau produktivitas harian tim Desain/Editor dan Operator branding secara detail.
        </p>
      </div>
      <div class="flex items-center gap-3">
        <Button
          @click="fetchData"
          class="bg-white border-2 border-slate-100 text-slate-700 hover:bg-slate-50 rounded-2xl px-4 py-6 shadow-sm transition-all font-bold flex items-center gap-2"
        >
          <RefreshCw class="w-5 h-5" :class="{ 'animate-spin': isLoading }" /> Refresh
        </Button>
      </div>
    </div>

    <!-- Tabs & Date Selector -->
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b border-slate-200 pb-1">
      <div class="flex items-center gap-4">
        <button
          @click="switchTab('editor')"
          class="pb-3 px-2 font-bold text-sm transition-all relative"
          :class="activeTab === 'editor' ? 'text-primary' : 'text-slate-400 hover:text-slate-600'"
        >
          <div class="flex items-center gap-2">
            <PenTool class="w-4 h-4" />
            KPI Tim Desain (Editor)
          </div>
          <div
            v-if="activeTab === 'editor'"
            class="absolute bottom-0 left-0 w-full h-0.5 bg-primary rounded-t-full"
          ></div>
        </button>
        <button
          @click="switchTab('operator')"
          class="pb-3 px-2 font-bold text-sm transition-all relative"
          :class="activeTab === 'operator' ? 'text-primary' : 'text-slate-400 hover:text-slate-600'"
        >
          <div class="flex items-center gap-2">
            <UserCircle class="w-4 h-4" />
            KPI Operator Pasang
          </div>
          <div
            v-if="activeTab === 'operator'"
            class="absolute bottom-0 left-0 w-full h-0.5 bg-primary rounded-t-full"
          ></div>
        </button>
      </div>

      <!-- Date Selector for Editor Evaluation -->
      <div v-if="activeTab === 'editor'" class="flex items-center gap-2 pb-2">
        <div class="flex items-center gap-1.5 bg-white border border-slate-200 rounded-xl px-3 py-1.5 shadow-sm text-xs font-bold text-slate-700">
          <Calendar class="w-3.5 h-3.5 text-primary" />
          <span>Tgl Evaluasi:</span>
          <input
            type="date"
            v-model="selectedDate"
            @change="fetchData"
            class="bg-transparent font-black text-slate-900 focus:outline-none cursor-pointer"
          />
        </div>
        <button
          @click="setToday"
          class="px-2.5 py-1.5 text-xs font-bold bg-slate-100 hover:bg-slate-200 rounded-lg text-slate-700 transition-colors"
        >
          Hari Ini
        </button>
        <button
          @click="setYesterday"
          class="px-2.5 py-1.5 text-xs font-bold bg-slate-100 hover:bg-slate-200 rounded-lg text-slate-700 transition-colors"
        >
          Kemarin
        </button>
      </div>
    </div>

    <!-- Content -->
    <div class="bg-white rounded-[32px] border border-slate-200 shadow-sm overflow-hidden">
      <!-- Editor KPI Table -->
      <div v-if="activeTab === 'editor'" class="overflow-x-auto">
        <table class="w-full text-left border-collapse">
          <thead>
            <tr class="bg-slate-50/80 uppercase text-[10px] font-black text-slate-400 tracking-[0.15em]">
              <th class="px-8 py-5">Nama Editor</th>
              <th class="px-8 py-5">Desain Baru (Tgl: {{ selectedDate }})</th>
              <th class="px-8 py-5">Revisi (Tgl: {{ selectedDate }})</th>
              <th class="px-8 py-5">Total Selesai (Tgl: {{ selectedDate }})</th>
              <th class="px-8 py-5">Total Akumulasi</th>
              <th class="px-8 py-5">Status</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-100">
            <tr v-if="isLoading" v-for="i in 3" :key="i" class="animate-pulse">
              <td v-for="j in 6" :key="j" class="px-8 py-6">
                <div class="h-4 bg-slate-100 rounded-lg w-full"></div>
              </td>
            </tr>
            <tr v-else-if="editorKPIs.length === 0">
              <td colspan="6" class="py-12 text-center text-sm font-bold text-slate-400">
                Belum ada data KPI Editor untuk tanggal ini.
              </td>
            </tr>
            <tr v-else v-for="editor in editorKPIs" :key="editor.id" class="group hover:bg-slate-50/50 transition-all">
              <td class="px-8 py-6">
                <p class="font-extrabold text-slate-900 leading-none">{{ editor.name }}</p>
                <p class="text-[11px] font-bold text-slate-400 mt-1.5">{{ editor.email }}</p>
              </td>
              <td class="px-8 py-6">
                <p class="text-sm font-black text-slate-900">{{ editor.harian_baru || 0 }}</p>
              </td>
              <td class="px-8 py-6">
                <p class="text-sm font-black text-amber-600">{{ editor.harian_revisi || 0 }}</p>
              </td>
              <td class="px-8 py-6">
                <span class="inline-flex items-center gap-1.5 px-3 py-1 rounded-xl text-xs font-black bg-indigo-50 text-indigo-700 border border-indigo-100">
                  <Sparkles class="w-3 h-3 text-indigo-500" />
                  {{ (editor.harian_baru || 0) + (editor.harian_revisi || 0) }} desain
                </span>
              </td>
              <td class="px-8 py-6">
                <p class="text-sm font-black text-primary">{{ editor.bulanan || 0 }}</p>
              </td>
              <td class="px-8 py-6">
                <span
                  :class="[
                    'text-[10px] font-black px-3 py-1 rounded-full uppercase tracking-wider border inline-flex items-center gap-1',
                    editor.is_active
                      ? 'bg-emerald-50 text-emerald-600 border-emerald-100'
                      : 'bg-slate-50 text-slate-500 border-slate-200'
                  ]"
                >
                  <CheckCircle2 v-if="editor.is_active" class="w-3 h-3" />
                  <XCircle v-else class="w-3 h-3" />
                  {{ editor.is_active ? 'AKTIF' : 'TIDAK AKTIF' }}
                </span>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- Operator KPI Table -->
      <div v-else class="overflow-x-auto">
        <table class="w-full text-left border-collapse">
          <thead>
            <tr class="bg-slate-50/80 uppercase text-[10px] font-black text-slate-400 tracking-[0.15em]">
              <th class="px-8 py-5">Nama Operator</th>
              <th class="px-8 py-5">Area Operasi</th>
              <th class="px-8 py-5">Banner Terpasang</th>
              <th class="px-8 py-5">Total Komisi</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-100">
            <tr v-if="isLoading" v-for="i in 3" :key="i" class="animate-pulse">
              <td v-for="j in 4" :key="j" class="px-8 py-6">
                <div class="h-4 bg-slate-100 rounded-lg w-full"></div>
              </td>
            </tr>
            <tr v-else-if="operatorKPIs.length === 0">
              <td colspan="4" class="py-12 text-center text-sm font-bold text-slate-400">
                Belum ada data KPI Operator.
              </td>
            </tr>
            <tr v-else v-for="operator in operatorKPIs" :key="operator.id" class="group hover:bg-slate-50/50 transition-all">
              <td class="px-8 py-6">
                <p class="font-extrabold text-slate-900 leading-none">{{ operator.name }}</p>
                <p class="text-[11px] font-bold text-slate-400 mt-1.5">{{ operator.email }}</p>
              </td>
              <td class="px-8 py-6">
                <p class="text-sm font-bold text-slate-600">{{ operator.area || '-' }}</p>
              </td>
              <td class="px-8 py-6">
                <p class="text-sm font-black text-primary">{{ operator.banner_terpasang || 0 }}</p>
              </td>
              <td class="px-8 py-6">
                <p class="text-sm font-black text-emerald-600">{{ formatCurrency(operator.total_komisi) }}</p>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>
