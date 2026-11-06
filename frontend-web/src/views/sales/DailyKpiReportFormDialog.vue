<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import apiClient from '@/api/axios'
import { toast } from 'vue-sonner'
import {
  X,
  Sparkles,
  Save,
  Send,
  Plus,
  Trash2,
  Calendar,
  AlertCircle,
  CheckCircle2,
  HelpCircle,
  TrendingUp,
  MapPin,
  Building2,
  User,
  Clock,
  Loader2,
} from 'lucide-vue-next'
import { Button } from '@/components/ui/button'

const props = defineProps<{
  open: boolean
  reportId?: string | null
  initialDate?: string
  initialBranchId?: string
}>()

const emit = defineEmits<{
  (e: 'update:open', value: boolean): void
  (e: 'saved'): void
}>()

const isLoading = ref(false)
const isPrefilling = ref(false)
const isSaving = ref(false)

const form = ref({
  id: '',
  report_date: new Date().toISOString().split('T')[0],
  branch_id: '',
  area_name: '',
  reporter_name: '',
  status: 'DRAFT',
  tomorrow_priority: '',
  notes: '',
  items: [] as any[],
  action_plans: [] as any[],
})

const formatCurrency = (val: number) => {
  if (isNaN(val)) return '0'
  return new Intl.NumberFormat('id-ID').format(val)
}

// Summary Calculation from Items
const summary = computed(() => {
  let targetCall = 0
  let realCall = 0
  let targetEC = 0
  let realEC = 0
  let targetSales = 0
  let realSales = 0

  for (const it of form.value.items) {
    targetCall += Number(it.target_call) || 0
    realCall += Number(it.realization_call) || 0
    targetEC += Number(it.target_ec) || 0
    realEC += Number(it.realization_ec) || 0
    targetSales += Number(it.target_sales) || 0
    realSales += Number(it.realization_sales) || 0
  }

  const capaiCall = targetCall > 0 ? (realCall / targetCall) * 100 : 0
  const capaiEC = targetEC > 0 ? (realEC / targetEC) * 100 : 0
  const capaiSales = targetSales > 0 ? (realSales / targetSales) * 100 : 0
  const diffSales = realSales - targetSales

  return {
    targetCall,
    realCall,
    capaiCall,
    targetEC,
    realEC,
    capaiEC,
    targetSales,
    realSales,
    capaiSales,
    diffSales,
  }
})

const getAchievementBadgeColor = (percent: number) => {
  if (percent >= 100) return 'bg-emerald-100 text-emerald-800 border-emerald-300'
  if (percent >= 80) return 'bg-amber-100 text-amber-800 border-amber-300'
  return 'bg-rose-100 text-rose-800 border-rose-300'
}

// Auto Prefill from System
const triggerAutoPrefill = async () => {
  isPrefilling.value = true
  try {
    let url = `/sales/daily-kpi-reports/prefill?date=${form.value.report_date}`
    if (form.value.branch_id) {
      url += `&branch_id=${form.value.branch_id}`
    }
    const res = await apiClient.get(url)
    const data = res.data.data
    if (data) {
      if (data.area_name && !form.value.area_name) {
        form.value.area_name = data.area_name
      }
      if (data.reporter_name) {
        form.value.reporter_name = data.reporter_name
      }

      if (data.items && data.items.length > 0) {
        form.value.items = data.items.map((it: any) => ({
          employee_id: it.employee_id,
          employee_name: it.employee_name,
          area_route: it.area_route || '-',
          target_call: it.target_call || 15,
          realization_call: it.realization_call || 0,
          target_ec: it.target_ec || 5,
          realization_ec: it.realization_ec || 0,
          target_sales: it.target_sales || 0,
          realization_sales: it.realization_sales || 0,
        }))
        toast.success(`Berhasil menarik data realisasi untuk ${data.items.length} sales`)
      } else {
        toast.info('Tidak ada data bawahan atau kunjungan terdeteksi pada tanggal ini.')
      }
    }
  } catch (err: any) {
    toast.error('Gagal menarik data realisasi otomatis: ' + (err.response?.data?.message || err.message))
  } finally {
    isPrefilling.value = false
  }
}

// Add empty row to Section A
const addItemRow = () => {
  form.value.items.push({
    employee_id: '',
    employee_name: '',
    area_route: '',
    target_call: 15,
    realization_call: 0,
    target_ec: 5,
    realization_ec: 0,
    target_sales: 0,
    realization_sales: 0,
  })
}

const removeItemRow = (index: number) => {
  form.value.items.splice(index, 1)
}

// Section B: Action Plan
const addActionPlanRow = (salesName = '') => {
  form.value.action_plans.push({
    employee_id: null,
    employee_name: salesName,
    reason_unachieved: '',
    field_obstacle: '',
    action_follow_up: '',
    pic_name: form.value.reporter_name || 'SPV',
    due_date: new Date(Date.now() + 86400000).toISOString().split('T')[0],
    status: 'OPEN',
    support_needed: '',
  })
}

const removeActionPlanRow = (index: number) => {
  form.value.action_plans.splice(index, 1)
}

// Quick Add all unachieved sales to Action Plan
const autoAddUnachievedToActionPlan = () => {
  const unachieved = form.value.items.filter((it) => {
    const capaiCall = it.target_call > 0 ? (it.realization_call / it.target_call) * 100 : 100
    const capaiEC = it.target_ec > 0 ? (it.realization_ec / it.target_ec) * 100 : 100
    const capaiSales = it.target_sales > 0 ? (it.realization_sales / it.target_sales) * 100 : 100
    return capaiCall < 100 || capaiEC < 100 || capaiSales < 100
  })

  if (unachieved.length === 0) {
    toast.info('Hebat! Seluruh sales telah mencapai 100% target.')
    return
  }

  let added = 0
  for (const s of unachieved) {
    const alreadyExists = form.value.action_plans.some((ap) => ap.employee_name === s.employee_name)
    if (!alreadyExists) {
      addActionPlanRow(s.employee_name)
      added++
    }
  }

  if (added > 0) {
    toast.success(`Ditambahkan ${added} sales yang belum capai target ke Evaluasi Bagian B.`)
  } else {
    toast.info('Semua sales yang belum capai target sudah ada di tabel evaluasi.')
  }
}

// Load existing report
const loadReport = async (id: string) => {
  isLoading.value = true
  try {
    const res = await apiClient.get(`/sales/daily-kpi-reports/${id}`)
    const r = res.data.data
    form.value = {
      id: r.id,
      report_date: r.report_date ? r.report_date.split('T')[0] : '',
      branch_id: r.branch_id || '',
      area_name: r.area_name || '',
      reporter_name: r.reporter_employee ? `${r.reporter_employee.first_name} ${r.reporter_employee.last_name || ''}`.trim() : '',
      status: r.status,
      tomorrow_priority: r.tomorrow_priority || '',
      notes: r.notes || '',
      items: (r.items || []).map((it: any) => ({
        ...it,
      })),
      action_plans: (r.action_plans || []).map((ap: any) => ({
        ...ap,
        due_date: ap.due_date ? ap.due_date.split('T')[0] : '',
      })),
    }
  } catch (err: any) {
    toast.error('Gagal memuat detail laporan: ' + (err.response?.data?.message || err.message))
  } finally {
    isLoading.value = false
  }
}

watch(
  () => props.open,
  (isOpen) => {
    if (isOpen) {
      if (props.reportId) {
        loadReport(props.reportId)
      } else {
        form.value = {
          id: '',
          report_date: props.initialDate || new Date().toISOString().split('T')[0],
          branch_id: props.initialBranchId || '',
          area_name: '',
          reporter_name: '',
          status: 'DRAFT',
          tomorrow_priority: '',
          notes: '',
          items: [],
          action_plans: [],
        }
        // Auto trigger prefill on new report
        triggerAutoPrefill()
      }
    }
  },
)

const saveReport = async (isDraft: boolean) => {
  if (form.value.items.length === 0) {
    toast.error('Laporan harus memiliki minimal 1 sales di Bagian A.')
    return
  }

  isSaving.value = true
  try {
    const payload = {
      report_date: form.value.report_date,
      branch_id: form.value.branch_id || null,
      area_name: form.value.area_name,
      is_draft: isDraft,
      tomorrow_priority: form.value.tomorrow_priority,
      notes: form.value.notes,
      items: form.value.items.map((it) => ({
        employee_id: it.employee_id,
        employee_name: it.employee_name,
        area_route: it.area_route,
        target_call: Number(it.target_call) || 0,
        realization_call: Number(it.realization_call) || 0,
        target_ec: Number(it.target_ec) || 0,
        realization_ec: Number(it.realization_ec) || 0,
        target_sales: Number(it.target_sales) || 0,
        realization_sales: Number(it.realization_sales) || 0,
      })),
      action_plans: form.value.action_plans.map((ap) => ({
        employee_id: ap.employee_id || null,
        employee_name: ap.employee_name,
        reason_unachieved: ap.reason_unachieved,
        field_obstacle: ap.field_obstacle,
        action_follow_up: ap.action_follow_up,
        pic_name: ap.pic_name,
        due_date: ap.due_date ? `${ap.due_date}T00:00:00Z` : null,
        status: ap.status || 'OPEN',
        support_needed: ap.support_needed,
      })),
    }

    if (form.value.id) {
      await apiClient.put(`/sales/daily-kpi-reports/${form.value.id}`, payload)
      toast.success(isDraft ? 'Draft laporan berhasil diperbarui' : 'Laporan berhasil disubmit ke atasan')
    } else {
      await apiClient.post('/sales/daily-kpi-reports', payload)
      toast.success(isDraft ? 'Draft laporan berhasil disimpan' : 'Laporan kinerja harian berhasil dikirim ke atasan')
    }

    emit('saved')
    emit('update:open', false)
  } catch (err: any) {
    toast.error('Gagal menyimpan laporan: ' + (err.response?.data?.message || err.message))
  } finally {
    isSaving.value = false
  }
}
</script>

<template>
  <div v-if="open" class="fixed inset-0 z-50 overflow-y-auto bg-slate-900/60 backdrop-blur-sm flex items-center justify-center p-3 sm:p-6 animate-in fade-in duration-200">
    <div class="bg-white rounded-3xl shadow-2xl border border-slate-200 w-full max-w-7xl max-h-[92vh] flex flex-col overflow-hidden">
      
      <!-- DIALOG HEADER -->
      <div class="px-6 py-5 border-b border-slate-200 flex items-center justify-between bg-slate-50/80">
        <div class="flex items-center gap-3">
          <div class="w-10 h-10 rounded-2xl bg-indigo-600 flex items-center justify-center text-white shadow-md shadow-indigo-200">
            <TrendingUp class="w-5 h-5" />
          </div>
          <div>
            <h2 class="text-xl font-black text-slate-800 tracking-tight flex items-center gap-2">
              <span>{{ form.id ? 'Edit Laporan KPI Sales Harian' : 'Buat Laporan KPI Sales Harian' }}</span>
              <span v-if="form.status === 'DRAFT'" class="text-xs font-bold px-2.5 py-0.5 rounded-full bg-slate-200 text-slate-700">DRAFT</span>
              <span v-else-if="form.status === 'SUBMITTED'" class="text-xs font-bold px-2.5 py-0.5 rounded-full bg-blue-100 text-blue-700">SUBMITTED</span>
            </h2>
            <p class="text-xs text-slate-500 font-medium">Format baku pelaporan kinerja harian SPV &rarr; ASM &rarr; NSM &rarr; GM</p>
          </div>
        </div>

        <div class="flex items-center gap-2">
          <Button
            variant="outline"
            size="sm"
            class="rounded-xl font-bold gap-2 text-indigo-600 border-indigo-200 hover:bg-indigo-50"
            :disabled="isPrefilling"
            @click="triggerAutoPrefill"
          >
            <Sparkles class="w-4 h-4" :class="{ 'animate-spin': isPrefilling }" />
            <span>Tarik Realisasi Otomatis</span>
          </Button>

          <button
            @click="emit('update:open', false)"
            class="p-2 rounded-xl text-slate-400 hover:text-slate-700 hover:bg-slate-200/60 transition-colors"
          >
            <X class="w-5 h-5" />
          </button>
        </div>
      </div>

      <!-- DIALOG BODY -->
      <div class="flex-1 overflow-y-auto p-6 space-y-8 bg-slate-50/30">
        
        <!-- HEADER METADATA -->
        <div class="grid grid-cols-1 md:grid-cols-4 gap-4 p-5 bg-white rounded-2xl border border-slate-200 shadow-xs">
          <div>
            <label class="block text-xs font-bold text-slate-500 uppercase tracking-wider mb-1.5 flex items-center gap-1.5">
              <Calendar class="w-3.5 h-3.5 text-indigo-500" /> Tanggal Laporan
            </label>
            <input
              type="date"
              v-model="form.report_date"
              @change="triggerAutoPrefill"
              class="w-full text-sm font-semibold text-slate-800 bg-slate-50 border border-slate-200 rounded-xl px-3 py-2 focus:ring-2 focus:ring-indigo-500 outline-none"
            />
          </div>

          <div>
            <label class="block text-xs font-bold text-slate-500 uppercase tracking-wider mb-1.5 flex items-center gap-1.5">
              <User class="w-3.5 h-3.5 text-indigo-500" /> Nama SPV / Pelapor
            </label>
            <input
              type="text"
              v-model="form.reporter_name"
              placeholder="Nama SPV..."
              class="w-full text-sm font-semibold text-slate-800 bg-slate-50 border border-slate-200 rounded-xl px-3 py-2 focus:ring-2 focus:ring-indigo-500 outline-none"
            />
          </div>

          <div>
            <label class="block text-xs font-bold text-slate-500 uppercase tracking-wider mb-1.5 flex items-center gap-1.5">
              <Building2 class="w-3.5 h-3.5 text-indigo-500" /> Cabang / Area
            </label>
            <input
              type="text"
              v-model="form.area_name"
              placeholder="Contoh: Denpasar Timur / Bali"
              class="w-full text-sm font-semibold text-slate-800 bg-slate-50 border border-slate-200 rounded-xl px-3 py-2 focus:ring-2 focus:ring-indigo-500 outline-none"
            />
          </div>

          <div class="bg-indigo-50/60 p-3 rounded-xl border border-indigo-100 flex flex-col justify-center">
            <span class="text-[11px] font-bold text-indigo-700 uppercase tracking-wider">Keterangan Warna Template:</span>
            <div class="flex items-center gap-3 mt-1 text-xs">
              <span class="inline-flex items-center gap-1 text-amber-900 font-semibold">
                <span class="w-3 h-3 rounded bg-amber-200 border border-amber-300 inline-block"></span> Diisi SPV
              </span>
              <span class="inline-flex items-center gap-1 text-slate-700 font-semibold">
                <span class="w-3 h-3 rounded bg-slate-200 border border-slate-300 inline-block"></span> Rumus Otomatis
              </span>
            </div>
          </div>
        </div>

        <!-- RINGKASAN TIM (KPI SUMMARY) -->
        <div class="space-y-3">
          <div class="flex items-center justify-between">
            <h3 class="text-sm font-extrabold text-slate-800 uppercase tracking-wider flex items-center gap-2">
              <span class="w-2.5 h-2.5 rounded-full bg-indigo-600"></span> Ringkasan Tim
            </h3>
            <span class="text-xs text-slate-400 italic">Nilai otomatis terakumulasi dari rincian sales di bawah</span>
          </div>

          <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
            <!-- Card CALL -->
            <div class="bg-white p-5 rounded-2xl border border-slate-200 shadow-xs relative overflow-hidden">
              <div class="flex justify-between items-start">
                <div>
                  <span class="text-xs font-bold text-slate-500 tracking-wider">CALL (KUNJUNGAN)</span>
                  <div class="mt-2 flex items-baseline gap-2">
                    <span class="text-3xl font-black text-slate-800">{{ summary.realCall }}</span>
                    <span class="text-xs font-semibold text-slate-400">/ Target {{ summary.targetCall }}</span>
                  </div>
                </div>
                <span :class="['text-xs font-black px-3 py-1 rounded-full border shadow-2xs', getAchievementBadgeColor(summary.capaiCall)]">
                  {{ summary.capaiCall.toFixed(1) }}%
                </span>
              </div>
              <div class="w-full bg-slate-100 h-2 rounded-full mt-4 overflow-hidden">
                <div class="bg-indigo-600 h-2 rounded-full transition-all duration-300" :style="{ width: `${Math.min(summary.capaiCall, 100)}%` }"></div>
              </div>
            </div>

            <!-- Card EC -->
            <div class="bg-white p-5 rounded-2xl border border-slate-200 shadow-xs relative overflow-hidden">
              <div class="flex justify-between items-start">
                <div>
                  <span class="text-xs font-bold text-slate-500 tracking-wider">EC (OUTLET ORDER)</span>
                  <div class="mt-2 flex items-baseline gap-2">
                    <span class="text-3xl font-black text-slate-800">{{ summary.realEC }}</span>
                    <span class="text-xs font-semibold text-slate-400">/ Target {{ summary.targetEC }}</span>
                  </div>
                </div>
                <span :class="['text-xs font-black px-3 py-1 rounded-full border shadow-2xs', getAchievementBadgeColor(summary.capaiEC)]">
                  {{ summary.capaiEC.toFixed(1) }}%
                </span>
              </div>
              <div class="w-full bg-slate-100 h-2 rounded-full mt-4 overflow-hidden">
                <div class="bg-blue-600 h-2 rounded-full transition-all duration-300" :style="{ width: `${Math.min(summary.capaiEC, 100)}%` }"></div>
              </div>
            </div>

            <!-- Card PENJUALAN -->
            <div class="bg-white p-5 rounded-2xl border border-slate-200 shadow-xs relative overflow-hidden">
              <div class="flex justify-between items-start">
                <div>
                  <span class="text-xs font-bold text-slate-500 tracking-wider">PENJUALAN (RP)</span>
                  <div class="mt-2 flex items-baseline gap-2">
                    <span class="text-2xl font-black text-emerald-600">Rp {{ formatCurrency(summary.realSales) }}</span>
                  </div>
                  <span class="text-xs font-medium text-slate-400">Target: Rp {{ formatCurrency(summary.targetSales) }}</span>
                </div>
                <span :class="['text-xs font-black px-3 py-1 rounded-full border shadow-2xs', getAchievementBadgeColor(summary.capaiSales)]">
                  {{ summary.capaiSales.toFixed(1) }}%
                </span>
              </div>
              <div class="mt-3 flex justify-between text-xs font-semibold">
                <span class="text-slate-500">Selisih:</span>
                <span :class="summary.diffSales >= 0 ? 'text-emerald-600 font-bold' : 'text-rose-600 font-bold'">
                  {{ summary.diffSales >= 0 ? '+' : '' }}Rp {{ formatCurrency(summary.diffSales) }}
                </span>
              </div>
            </div>
          </div>
        </div>

        <!-- BAGIAN A: TARGET DAN REALISASI PER SALES -->
        <div class="space-y-3">
          <div class="flex items-center justify-between flex-wrap gap-2">
            <div>
              <h3 class="text-sm font-extrabold text-slate-800 uppercase tracking-wider flex items-center gap-2">
                <span class="w-2.5 h-2.5 rounded-full bg-indigo-600"></span> A. Target dan Realisasi per Sales
              </h3>
              <p class="text-xs text-slate-400">Input rincian kinerja masing-masing salesman di wilayah Anda</p>
            </div>
            <Button size="sm" variant="outline" class="rounded-xl font-bold gap-1 text-slate-700" @click="addItemRow">
              <Plus class="w-4 h-4" /> Tambah Baris Sales
            </Button>
          </div>

          <div class="overflow-x-auto bg-white rounded-2xl border border-slate-200 shadow-xs">
            <table class="w-full text-xs text-left">
              <thead class="bg-slate-900 text-white font-extrabold uppercase tracking-wider text-[11px]">
                <tr>
                  <th class="py-3 px-3">Nama Sales</th>
                  <th class="py-3 px-3">Area / Rute</th>
                  <th class="py-3 px-2 text-center bg-slate-800">Target Call</th>
                  <th class="py-3 px-2 text-center bg-slate-800">Real Call</th>
                  <th class="py-3 px-2 text-center">Capai %</th>
                  <th class="py-3 px-2 text-center bg-slate-800">Target EC</th>
                  <th class="py-3 px-2 text-center bg-slate-800">Real EC</th>
                  <th class="py-3 px-2 text-center">Capai %</th>
                  <th class="py-3 px-3 text-right bg-slate-800">Target Jual (Rp)</th>
                  <th class="py-3 px-3 text-right bg-slate-800">Real Jual (Rp)</th>
                  <th class="py-3 px-2 text-center">Capai %</th>
                  <th class="py-3 px-3 text-right">Selisih (Rp)</th>
                  <th class="py-3 px-2 text-center">Aksi</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-slate-100">
                <tr v-if="form.items.length === 0">
                  <td colspan="13" class="py-8 text-center text-slate-400 font-medium">
                    Belum ada data sales. Klik <strong>"Tarik Realisasi Otomatis"</strong> atau <strong>"Tambah Baris Sales"</strong>.
                  </td>
                </tr>
                <tr v-for="(it, idx) in form.items" :key="idx" class="hover:bg-slate-50/60 transition-colors">
                  <!-- Nama Sales -->
                  <td class="p-2 min-w-[140px]">
                    <input
                      type="text"
                      v-model="it.employee_name"
                      placeholder="Nama Sales..."
                      class="w-full px-2 py-1.5 font-bold text-slate-800 bg-white border border-slate-200 rounded-lg text-xs outline-none focus:border-indigo-500"
                    />
                  </td>
                  <!-- Area Rute -->
                  <td class="p-2 min-w-[120px]">
                    <input
                      type="text"
                      v-model="it.area_route"
                      placeholder="Area / Rute..."
                      class="w-full px-2 py-1.5 text-slate-700 bg-white border border-slate-200 rounded-lg text-xs outline-none focus:border-indigo-500"
                    />
                  </td>
                  <!-- Target Call (Kuning muda) -->
                  <td class="p-2 w-18">
                    <input
                      type="number"
                      v-model.number="it.target_call"
                      min="0"
                      class="w-full text-center px-1.5 py-1.5 font-bold text-amber-950 bg-amber-50 border border-amber-200 rounded-lg text-xs outline-none focus:border-indigo-500"
                    />
                  </td>
                  <!-- Real Call (Kuning muda) -->
                  <td class="p-2 w-18">
                    <input
                      type="number"
                      v-model.number="it.realization_call"
                      min="0"
                      class="w-full text-center px-1.5 py-1.5 font-black text-amber-950 bg-amber-50 border border-amber-200 rounded-lg text-xs outline-none focus:border-indigo-500"
                    />
                  </td>
                  <!-- Capai Call % (Abu formula) -->
                  <td class="p-2 text-center w-18 bg-slate-50 font-black">
                    <span :class="['px-2 py-0.5 rounded-md text-[11px]', getAchievementBadgeColor(it.target_call > 0 ? (it.realization_call / it.target_call) * 100 : 0)]">
                      {{ (it.target_call > 0 ? (it.realization_call / it.target_call) * 100 : 0).toFixed(0) }}%
                    </span>
                  </td>
                  <!-- Target EC (Kuning muda) -->
                  <td class="p-2 w-18">
                    <input
                      type="number"
                      v-model.number="it.target_ec"
                      min="0"
                      class="w-full text-center px-1.5 py-1.5 font-bold text-amber-950 bg-amber-50 border border-amber-200 rounded-lg text-xs outline-none focus:border-indigo-500"
                    />
                  </td>
                  <!-- Real EC (Kuning muda) -->
                  <td class="p-2 w-18">
                    <input
                      type="number"
                      v-model.number="it.realization_ec"
                      min="0"
                      class="w-full text-center px-1.5 py-1.5 font-black text-amber-950 bg-amber-50 border border-amber-200 rounded-lg text-xs outline-none focus:border-indigo-500"
                    />
                  </td>
                  <!-- Capai EC % (Abu formula) -->
                  <td class="p-2 text-center w-18 bg-slate-50 font-black">
                    <span :class="['px-2 py-0.5 rounded-md text-[11px]', getAchievementBadgeColor(it.target_ec > 0 ? (it.realization_ec / it.target_ec) * 100 : 0)]">
                      {{ (it.target_ec > 0 ? (it.realization_ec / it.target_ec) * 100 : 0).toFixed(0) }}%
                    </span>
                  </td>
                  <!-- Target Sales (Kuning muda) -->
                  <td class="p-2 min-w-[130px]">
                    <input
                      type="number"
                      v-model.number="it.target_sales"
                      min="0"
                      class="w-full text-right px-2 py-1.5 font-semibold text-amber-950 bg-amber-50 border border-amber-200 rounded-lg text-xs outline-none focus:border-indigo-500"
                    />
                  </td>
                  <!-- Real Sales (Kuning muda) -->
                  <td class="p-2 min-w-[130px]">
                    <input
                      type="number"
                      v-model.number="it.realization_sales"
                      min="0"
                      class="w-full text-right px-2 py-1.5 font-black text-emerald-800 bg-amber-50 border border-amber-200 rounded-lg text-xs outline-none focus:border-indigo-500"
                    />
                  </td>
                  <!-- Capai Sales % (Abu formula) -->
                  <td class="p-2 text-center w-18 bg-slate-50 font-black">
                    <span :class="['px-2 py-0.5 rounded-md text-[11px]', getAchievementBadgeColor(it.target_sales > 0 ? (it.realization_sales / it.target_sales) * 100 : 0)]">
                      {{ (it.target_sales > 0 ? (it.realization_sales / it.target_sales) * 100 : 0).toFixed(0) }}%
                    </span>
                  </td>
                  <!-- Selisih Sales (Abu formula) -->
                  <td class="p-2 text-right min-w-[120px] bg-slate-50 font-bold" :class="(it.realization_sales - it.target_sales) >= 0 ? 'text-emerald-700' : 'text-rose-600'">
                    Rp {{ formatCurrency(it.realization_sales - it.target_sales) }}
                  </td>
                  <!-- Delete button -->
                  <td class="p-2 text-center">
                    <button
                      type="button"
                      @click="removeItemRow(idx)"
                      class="p-1 rounded text-slate-300 hover:text-rose-600 hover:bg-rose-50 transition-colors"
                    >
                      <Trash2 class="w-4 h-4" />
                    </button>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>

        <!-- BAGIAN B: EVALUASI DAN TINDAK LANJUT SPV -->
        <div class="space-y-3">
          <div class="flex items-center justify-between flex-wrap gap-2">
            <div>
              <h3 class="text-sm font-extrabold text-slate-800 uppercase tracking-wider flex items-center gap-2">
                <span class="w-2.5 h-2.5 rounded-full bg-indigo-600"></span> B. Evaluasi dan Tindak Lanjut SPV
              </h3>
              <p class="text-xs text-slate-400">Tuliskan analisa penyebab, hambatan lapangan, dan rencana aksi konkret untuk sales yang belum capai target</p>
            </div>
            <div class="flex items-center gap-2">
              <Button size="sm" variant="outline" class="rounded-xl font-bold gap-1 text-indigo-700 border-indigo-200 hover:bg-indigo-50" @click="autoAddUnachievedToActionPlan">
                <Sparkles class="w-3.5 h-3.5" /> Deteksi Sales Belum Capai Target
              </Button>
              <Button size="sm" variant="outline" class="rounded-xl font-bold gap-1 text-slate-700" @click="() => addActionPlanRow()">
                <Plus class="w-4 h-4" /> Tambah Evaluasi
              </Button>
            </div>
          </div>

          <div class="overflow-x-auto bg-white rounded-2xl border border-slate-200 shadow-xs">
            <table class="w-full text-xs text-left">
              <thead class="bg-slate-900 text-white font-extrabold uppercase tracking-wider text-[11px]">
                <tr>
                  <th class="py-3 px-3 min-w-[130px]">Nama Sales</th>
                  <th class="py-3 px-3 min-w-[180px]">Alasan Target Belum Tercapai</th>
                  <th class="py-3 px-3 min-w-[180px]">Kendala Lapangan</th>
                  <th class="py-3 px-3 min-w-[200px]">Tindak Lanjut SPV</th>
                  <th class="py-3 px-2 w-28">PIC</th>
                  <th class="py-3 px-2 w-32">Batas Waktu</th>
                  <th class="py-3 px-2 w-28 text-center">Status</th>
                  <th class="py-3 px-3 min-w-[180px]">Dukungan Diperlukan</th>
                  <th class="py-3 px-2 text-center">Aksi</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-slate-100">
                <tr v-if="form.action_plans.length === 0">
                  <td colspan="9" class="py-6 text-center text-slate-400">
                    Belum ada catatan evaluasi kendala. Klik <strong>"Deteksi Sales Belum Capai Target"</strong> atau <strong>"Tambah Evaluasi"</strong>.
                  </td>
                </tr>
                <tr v-for="(ap, idx) in form.action_plans" :key="idx" class="hover:bg-slate-50/60 transition-colors">
                  <td class="p-2">
                    <input
                      type="text"
                      v-model="ap.employee_name"
                      placeholder="Nama Sales..."
                      class="w-full px-2 py-1.5 font-bold text-slate-800 bg-white border border-slate-200 rounded-lg text-xs outline-none focus:border-indigo-500"
                    />
                  </td>
                  <td class="p-2">
                    <textarea
                      v-model="ap.reason_unachieved"
                      rows="2"
                      placeholder="Contoh: 4 outlet tidak order karena stok masih penuh"
                      class="w-full px-2 py-1.5 text-slate-700 bg-white border border-slate-200 rounded-lg text-xs outline-none focus:border-indigo-500"
                    ></textarea>
                  </td>
                  <td class="p-2">
                    <textarea
                      v-model="ap.field_obstacle"
                      rows="2"
                      placeholder="Contoh: Stok produk fast-moving kosong di depo"
                      class="w-full px-2 py-1.5 text-slate-700 bg-white border border-slate-200 rounded-lg text-xs outline-none focus:border-indigo-500"
                    ></textarea>
                  </td>
                  <td class="p-2">
                    <textarea
                      v-model="ap.action_follow_up"
                      rows="2"
                      placeholder="Contoh: SPV koordinasi percepatan restock gudang"
                      class="w-full px-2 py-1.5 font-medium text-slate-800 bg-white border border-slate-200 rounded-lg text-xs outline-none focus:border-indigo-500"
                    ></textarea>
                  </td>
                  <td class="p-2">
                    <input
                      type="text"
                      v-model="ap.pic_name"
                      placeholder="PIC..."
                      class="w-full px-2 py-1.5 text-slate-700 bg-white border border-slate-200 rounded-lg text-xs outline-none focus:border-indigo-500"
                    />
                  </td>
                  <td class="p-2">
                    <input
                      type="date"
                      v-model="ap.due_date"
                      class="w-full px-2 py-1.5 text-slate-700 bg-white border border-slate-200 rounded-lg text-xs outline-none focus:border-indigo-500"
                    />
                  </td>
                  <td class="p-2 text-center">
                    <select
                      v-model="ap.status"
                      class="w-full px-1.5 py-1.5 font-bold text-xs rounded-lg border border-slate-200 bg-white outline-none"
                    >
                      <option value="OPEN">OPEN</option>
                      <option value="IN_PROGRESS">IN PROGRESS</option>
                      <option value="DONE">DONE</option>
                    </select>
                  </td>
                  <td class="p-2">
                    <textarea
                      v-model="ap.support_needed"
                      rows="2"
                      placeholder="Contoh: Butuh approval alokasi stok dari ASM / Gudang"
                      class="w-full px-2 py-1.5 text-slate-700 bg-white border border-slate-200 rounded-lg text-xs outline-none focus:border-indigo-500"
                    ></textarea>
                  </td>
                  <td class="p-2 text-center">
                    <button
                      type="button"
                      @click="removeActionPlanRow(idx)"
                      class="p-1 rounded text-slate-300 hover:text-rose-600 hover:bg-rose-50 transition-colors"
                    >
                      <Trash2 class="w-4 h-4" />
                    </button>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>

        <!-- BAGIAN C: CATATAN DAN VALIDASI SPV -->
        <div class="space-y-4">
          <h3 class="text-sm font-extrabold text-slate-800 uppercase tracking-wider flex items-center gap-2">
            <span class="w-2.5 h-2.5 rounded-full bg-indigo-600"></span> C. Catatan dan Validasi SPV
          </h3>

          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div class="bg-white p-5 rounded-2xl border border-slate-200 shadow-xs">
              <label class="block text-xs font-bold text-slate-700 uppercase tracking-wider mb-2">
                Prioritas Besok:
              </label>
              <textarea
                v-model="form.tomorrow_priority"
                rows="4"
                placeholder="Tuliskan fokus area, target kunjungan prioritas, atau program khusus untuk esok hari..."
                class="w-full p-3 text-sm text-slate-800 bg-slate-50 border border-slate-200 rounded-xl outline-none focus:ring-2 focus:ring-indigo-500"
              ></textarea>
            </div>

            <div class="bg-white p-5 rounded-2xl border border-slate-200 shadow-xs">
              <label class="block text-xs font-bold text-slate-700 uppercase tracking-wider mb-2">
                Catatan Tambahan untuk Atasan (ASM / NSM):
              </label>
              <textarea
                v-model="form.notes"
                rows="4"
                placeholder="Catatan kondisi tim, cuaca, promo kompetitor, atau usulan khusus..."
                class="w-full p-3 text-sm text-slate-800 bg-slate-50 border border-slate-200 rounded-xl outline-none focus:ring-2 focus:ring-indigo-500"
              ></textarea>
            </div>
          </div>
        </div>

      </div>

      <!-- DIALOG FOOTER -->
      <div class="px-6 py-4 border-t border-slate-200 bg-white flex items-center justify-between flex-wrap gap-3">
        <div class="flex items-center gap-2 text-xs text-slate-400">
          <Clock class="w-4 h-4" />
          <span>Status: <strong>{{ form.status }}</strong></span>
        </div>

        <div class="flex items-center gap-2">
          <Button
            variant="outline"
            class="rounded-xl font-bold"
            @click="emit('update:open', false)"
          >
            Batal
          </Button>

          <Button
            variant="outline"
            class="rounded-xl font-bold gap-2 text-slate-700 border-slate-300 hover:bg-slate-50"
            :disabled="isSaving"
            @click="saveReport(true)"
          >
            <Save class="w-4 h-4" />
            <span>Simpan Draft</span>
          </Button>

          <Button
            class="rounded-xl font-extrabold gap-2 bg-indigo-600 hover:bg-indigo-700 text-white shadow-md shadow-indigo-200"
            :disabled="isSaving"
            @click="saveReport(false)"
          >
            <Send class="w-4 h-4" />
            <span>Kirim Laporan ke Atasan</span>
          </Button>
        </div>
      </div>

    </div>
  </div>
</template>
