<script setup lang="ts">
import { ref, watch, computed } from 'vue'
import apiClient from '@/api/axios'
import { toast } from 'vue-sonner'
import {
  X,
  FileDown,
  CheckCircle2,
  AlertCircle,
  TrendingUp,
  Building2,
  Calendar,
  User,
  Clock,
  MessageSquare,
  ShieldAlert,
} from 'lucide-vue-next'
import { Button } from '@/components/ui/button'

const props = defineProps<{
  open: boolean
  reportId: string | null
}>()

const emit = defineEmits<{
  (e: 'update:open', value: boolean): void
  (e: 'reviewed'): void
}>()

const isLoading = ref(false)
const isSubmitting = ref(false)
const report = ref<any>(null)
const feedback = ref('')

const formatCurrency = (val: number) => {
  if (isNaN(val)) return '0'
  return new Intl.NumberFormat('id-ID').format(val)
}

const getAchievementBadgeColor = (percent: number) => {
  if (percent >= 100) return 'bg-emerald-100 text-emerald-800 border-emerald-300'
  if (percent >= 80) return 'bg-amber-100 text-amber-800 border-amber-300'
  return 'bg-rose-100 text-rose-800 border-rose-300'
}

const loadReport = async (id: string) => {
  isLoading.value = true
  try {
    const res = await apiClient.get(`/sales/daily-kpi-reports/${id}`)
    report.value = res.data.data
    feedback.value = report.value?.reviewer_feedback || ''
  } catch (err: any) {
    toast.error('Gagal memuat detail laporan: ' + (err.response?.data?.message || err.message))
  } finally {
    isLoading.value = false
  }
}

watch(
  () => props.open,
  (isOpen) => {
    if (isOpen && props.reportId) {
      loadReport(props.reportId)
    } else {
      report.value = null
      feedback.value = ''
    }
  },
)

const submitReview = async (status: 'APPROVED' | 'REVISED') => {
  if (!props.reportId) return

  if (status === 'REVISED' && !feedback.value.trim()) {
    toast.error('Mohon tuliskan catatan/alasan revisi pada kolom evaluasi atasan.')
    return
  }

  isSubmitting.value = true
  try {
    await apiClient.post(`/sales/daily-kpi-reports/${props.reportId}/review`, {
      status,
      feedback: feedback.value,
    })

    toast.success(status === 'APPROVED' ? 'Laporan berhasil disetujui (Approved)' : 'Laporan dikembalikan ke pembuat untuk revisi')
    emit('reviewed')
    emit('update:open', false)
  } catch (err: any) {
    toast.error('Gagal memproses review: ' + (err.response?.data?.message || err.message))
  } finally {
    isSubmitting.value = false
  }
}

const downloadExcel = () => {
  if (!props.reportId) return
  window.open(`${apiClient.defaults.baseURL}/sales/daily-kpi-reports/${props.reportId}/export?token=${localStorage.getItem('token') || ''}`, '_blank')
}
</script>

<template>
  <div v-if="open" class="fixed inset-0 z-50 overflow-y-auto bg-slate-900/60 backdrop-blur-sm flex items-center justify-center p-3 sm:p-6 animate-in fade-in duration-200">
    <div class="bg-white rounded-3xl shadow-2xl border border-slate-200 w-full max-w-6xl max-h-[92vh] flex flex-col overflow-hidden">
      
      <!-- HEADER -->
      <div class="px-6 py-5 border-b border-slate-200 flex items-center justify-between bg-slate-50/80">
        <div class="flex items-center gap-3">
          <div class="w-10 h-10 rounded-2xl bg-indigo-600 flex items-center justify-center text-white shadow-md shadow-indigo-200">
            <TrendingUp class="w-5 h-5" />
          </div>
          <div>
            <h2 class="text-xl font-black text-slate-800 tracking-tight flex items-center gap-2">
              <span>Review Laporan KPI Sales Harian</span>
              <span v-if="report" class="text-xs font-bold px-2.5 py-0.5 rounded-full"
                :class="{
                  'bg-emerald-100 text-emerald-800': report.status === 'APPROVED',
                  'bg-amber-100 text-amber-800': report.status === 'SUBMITTED',
                  'bg-rose-100 text-rose-800': report.status === 'REVISED',
                  'bg-slate-200 text-slate-700': report.status === 'DRAFT',
                }"
              >
                {{ report.status }}
              </span>
            </h2>
            <p class="text-xs text-slate-500 font-medium">Validasi dan berikan evaluasi kinerja bawahan Anda</p>
          </div>
        </div>

        <div class="flex items-center gap-2">
          <Button
            variant="outline"
            size="sm"
            class="rounded-xl font-bold gap-2 text-slate-700 border-slate-300 hover:bg-slate-50"
            @click="downloadExcel"
          >
            <FileDown class="w-4 h-4 text-emerald-600" />
            <span>Download Excel</span>
          </Button>

          <button
            @click="emit('update:open', false)"
            class="p-2 rounded-xl text-slate-400 hover:text-slate-700 hover:bg-slate-200/60 transition-colors"
          >
            <X class="w-5 h-5" />
          </button>
        </div>
      </div>

      <!-- BODY -->
      <div v-if="report" class="flex-1 overflow-y-auto p-6 space-y-8 bg-slate-50/30">
        
        <!-- HEADER METADATA -->
        <div class="grid grid-cols-1 md:grid-cols-3 gap-4 p-5 bg-white rounded-2xl border border-slate-200 shadow-xs">
          <div>
            <span class="text-xs font-bold text-slate-400 uppercase tracking-wider block">Tanggal Laporan</span>
            <span class="text-base font-extrabold text-slate-800 mt-1 block flex items-center gap-1.5">
              <Calendar class="w-4 h-4 text-indigo-500" />
              {{ report.report_date ? new Date(report.report_date).toLocaleDateString('id-ID', { dateStyle: 'full' }) : '-' }}
            </span>
          </div>

          <div>
            <span class="text-xs font-bold text-slate-400 uppercase tracking-wider block">Pelapor / Supervisor</span>
            <span class="text-base font-extrabold text-slate-800 mt-1 block flex items-center gap-1.5">
              <User class="w-4 h-4 text-indigo-500" />
              {{ report.reporter_employee ? `${report.reporter_employee.first_name} ${report.reporter_employee.last_name || ''}`.trim() : '-' }}
              <span class="text-xs px-2 py-0.5 rounded-md bg-indigo-50 text-indigo-700 font-bold ml-1">
                {{ report.reporter_role || 'SPV' }}
              </span>
            </span>
          </div>

          <div>
            <span class="text-xs font-bold text-slate-400 uppercase tracking-wider block">Cabang / Area</span>
            <span class="text-base font-extrabold text-slate-800 mt-1 block flex items-center gap-1.5">
              <Building2 class="w-4 h-4 text-indigo-500" />
              {{ report.area_name || (report.branch ? report.branch.name : '-') }}
            </span>
          </div>
        </div>

        <!-- RINGKASAN TIM -->
        <div class="space-y-3">
          <h3 class="text-sm font-extrabold text-slate-800 uppercase tracking-wider flex items-center gap-2">
            <span class="w-2.5 h-2.5 rounded-full bg-indigo-600"></span> Ringkasan Tim
          </h3>

          <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
            <!-- CALL -->
            <div class="bg-white p-5 rounded-2xl border border-slate-200 shadow-xs">
              <div class="flex justify-between items-start">
                <div>
                  <span class="text-xs font-bold text-slate-500">CALL (KUNJUNGAN)</span>
                  <div class="mt-2 flex items-baseline gap-2">
                    <span class="text-3xl font-black text-slate-800">{{ report.summary_realization_call }}</span>
                    <span class="text-xs font-semibold text-slate-400">/ Target {{ report.summary_target_call }}</span>
                  </div>
                </div>
                <span :class="['text-xs font-black px-3 py-1 rounded-full border shadow-2xs', getAchievementBadgeColor(report.summary_achievement_call_percent)]">
                  {{ Number(report.summary_achievement_call_percent || 0).toFixed(1) }}%
                </span>
              </div>
            </div>

            <!-- EC -->
            <div class="bg-white p-5 rounded-2xl border border-slate-200 shadow-xs">
              <div class="flex justify-between items-start">
                <div>
                  <span class="text-xs font-bold text-slate-500">EC (OUTLET ORDER)</span>
                  <div class="mt-2 flex items-baseline gap-2">
                    <span class="text-3xl font-black text-slate-800">{{ report.summary_realization_ec }}</span>
                    <span class="text-xs font-semibold text-slate-400">/ Target {{ report.summary_target_ec }}</span>
                  </div>
                </div>
                <span :class="['text-xs font-black px-3 py-1 rounded-full border shadow-2xs', getAchievementBadgeColor(report.summary_achievement_ec_percent)]">
                  {{ Number(report.summary_achievement_ec_percent || 0).toFixed(1) }}%
                </span>
              </div>
            </div>

            <!-- PENJUALAN -->
            <div class="bg-white p-5 rounded-2xl border border-slate-200 shadow-xs">
              <div class="flex justify-between items-start">
                <div>
                  <span class="text-xs font-bold text-slate-500">PENJUALAN (RP)</span>
                  <div class="mt-2 flex items-baseline gap-2">
                    <span class="text-2xl font-black text-emerald-600">Rp {{ formatCurrency(report.summary_realization_sales) }}</span>
                  </div>
                  <span class="text-xs font-medium text-slate-400">Target: Rp {{ formatCurrency(report.summary_target_sales) }}</span>
                </div>
                <span :class="['text-xs font-black px-3 py-1 rounded-full border shadow-2xs', getAchievementBadgeColor(report.summary_achievement_sales_percent)]">
                  {{ Number(report.summary_achievement_sales_percent || 0).toFixed(1) }}%
                </span>
              </div>
            </div>
          </div>
        </div>

        <!-- BAGIAN A: TARGET DAN REALISASI PER SALES -->
        <div class="space-y-3">
          <h3 class="text-sm font-extrabold text-slate-800 uppercase tracking-wider flex items-center gap-2">
            <span class="w-2.5 h-2.5 rounded-full bg-indigo-600"></span> A. Target dan Realisasi per Sales
          </h3>

          <div class="overflow-x-auto bg-white rounded-2xl border border-slate-200 shadow-xs">
            <table class="w-full text-xs text-left">
              <thead class="bg-slate-900 text-white font-extrabold uppercase tracking-wider text-[11px]">
                <tr>
                  <th class="py-3 px-3">Nama Sales</th>
                  <th class="py-3 px-3">Area / Rute</th>
                  <th class="py-3 px-2 text-center">Target Call</th>
                  <th class="py-3 px-2 text-center">Real Call</th>
                  <th class="py-3 px-2 text-center">Capai %</th>
                  <th class="py-3 px-2 text-center">Target EC</th>
                  <th class="py-3 px-2 text-center">Real EC</th>
                  <th class="py-3 px-2 text-center">Capai %</th>
                  <th class="py-3 px-3 text-right">Target Jual (Rp)</th>
                  <th class="py-3 px-3 text-right">Real Jual (Rp)</th>
                  <th class="py-3 px-2 text-center">Capai %</th>
                  <th class="py-3 px-3 text-right">Selisih (Rp)</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-slate-100">
                <tr v-for="(it, idx) in report.items" :key="idx" class="hover:bg-slate-50/60">
                  <td class="p-3 font-bold text-slate-800">{{ it.employee_name }}</td>
                  <td class="p-3 text-slate-600">{{ it.area_route }}</td>
                  <td class="p-3 text-center font-semibold">{{ it.target_call }}</td>
                  <td class="p-3 text-center font-black text-indigo-600">{{ it.realization_call }}</td>
                  <td class="p-3 text-center">
                    <span :class="['px-2 py-0.5 rounded-md text-[11px] font-black', getAchievementBadgeColor(it.achievement_call_percent)]">
                      {{ Number(it.achievement_call_percent || 0).toFixed(0) }}%
                    </span>
                  </td>
                  <td class="p-3 text-center font-semibold">{{ it.target_ec }}</td>
                  <td class="p-3 text-center font-black text-blue-600">{{ it.realization_ec }}</td>
                  <td class="p-3 text-center">
                    <span :class="['px-2 py-0.5 rounded-md text-[11px] font-black', getAchievementBadgeColor(it.achievement_ec_percent)]">
                      {{ Number(it.achievement_ec_percent || 0).toFixed(0) }}%
                    </span>
                  </td>
                  <td class="p-3 text-right font-medium text-slate-500">Rp {{ formatCurrency(it.target_sales) }}</td>
                  <td class="p-3 text-right font-black text-emerald-700">Rp {{ formatCurrency(it.realization_sales) }}</td>
                  <td class="p-3 text-center">
                    <span :class="['px-2 py-0.5 rounded-md text-[11px] font-black', getAchievementBadgeColor(it.achievement_sales_percent)]">
                      {{ Number(it.achievement_sales_percent || 0).toFixed(0) }}%
                    </span>
                  </td>
                  <td class="p-3 text-right font-bold" :class="it.diff_sales >= 0 ? 'text-emerald-700' : 'text-rose-600'">
                    {{ it.diff_sales >= 0 ? '+' : '' }}Rp {{ formatCurrency(it.diff_sales) }}
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>

        <!-- BAGIAN B: EVALUASI DAN TINDAK LANJUT SPV -->
        <div class="space-y-3">
          <h3 class="text-sm font-extrabold text-slate-800 uppercase tracking-wider flex items-center gap-2">
            <span class="w-2.5 h-2.5 rounded-full bg-indigo-600"></span> B. Evaluasi dan Tindak Lanjut SPV
          </h3>

          <div class="overflow-x-auto bg-white rounded-2xl border border-slate-200 shadow-xs">
            <table class="w-full text-xs text-left">
              <thead class="bg-slate-900 text-white font-extrabold uppercase tracking-wider text-[11px]">
                <tr>
                  <th class="py-3 px-3">Nama Sales</th>
                  <th class="py-3 px-3">Alasan Target Belum Tercapai</th>
                  <th class="py-3 px-3">Kendala Lapangan</th>
                  <th class="py-3 px-3">Tindak Lanjut SPV</th>
                  <th class="py-3 px-2">PIC</th>
                  <th class="py-3 px-2">Batas Waktu</th>
                  <th class="py-3 px-2 text-center">Status</th>
                  <th class="py-3 px-3">Dukungan Diperlukan</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-slate-100">
                <tr v-if="!report.action_plans || report.action_plans.length === 0">
                  <td colspan="8" class="py-6 text-center text-slate-400 font-medium">
                    Seluruh target tercapai / tidak ada catatan kendala khusus.
                  </td>
                </tr>
                <tr v-for="(ap, idx) in report.action_plans" :key="idx" class="hover:bg-slate-50/60">
                  <td class="p-3 font-bold text-slate-800">{{ ap.employee_name }}</td>
                  <td class="p-3 text-slate-700">{{ ap.reason_unachieved || '-' }}</td>
                  <td class="p-3 text-slate-700">{{ ap.field_obstacle || '-' }}</td>
                  <td class="p-3 font-semibold text-slate-900">{{ ap.action_follow_up || '-' }}</td>
                  <td class="p-3 text-slate-600">{{ ap.pic_name || '-' }}</td>
                  <td class="p-3 text-slate-600">{{ ap.due_date ? new Date(ap.due_date).toLocaleDateString('id-ID') : '-' }}</td>
                  <td class="p-3 text-center">
                    <span class="px-2 py-0.5 rounded-full text-[10px] font-extrabold"
                      :class="{
                        'bg-emerald-100 text-emerald-800': ap.status === 'DONE',
                        'bg-blue-100 text-blue-800': ap.status === 'IN_PROGRESS',
                        'bg-slate-100 text-slate-700': ap.status === 'OPEN',
                      }"
                    >
                      {{ ap.status }}
                    </span>
                  </td>
                  <td class="p-3 text-amber-900 font-medium bg-amber-50/40">{{ ap.support_needed || '-' }}</td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>

        <!-- BAGIAN C: PRIORITAS BESOK & CATATAN PELAPOR -->
        <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div class="bg-white p-5 rounded-2xl border border-slate-200 shadow-xs">
            <span class="text-xs font-bold text-slate-400 uppercase tracking-wider block mb-2">Prioritas Besok</span>
            <p class="text-sm text-slate-800 whitespace-pre-line">{{ report.tomorrow_priority || '-' }}</p>
          </div>

          <div class="bg-white p-5 rounded-2xl border border-slate-200 shadow-xs">
            <span class="text-xs font-bold text-slate-400 uppercase tracking-wider block mb-2">Catatan Tambahan SPV</span>
            <p class="text-sm text-slate-800 whitespace-pre-line">{{ report.notes || '-' }}</p>
          </div>
        </div>

        <!-- FORM EVALUASI & FEEDBACK ATASAN -->
        <div class="bg-indigo-50/50 p-6 rounded-3xl border border-indigo-200/80 shadow-xs space-y-4">
          <div class="flex items-center gap-2 text-indigo-900 font-extrabold text-sm">
            <MessageSquare class="w-4 h-4 text-indigo-600" />
            <span>Catatan Evaluasi / Feedback Atasan (ASM / NSM / GM):</span>
          </div>

          <textarea
            v-model="feedback"
            rows="3"
            placeholder="Tuliskan arahan, respon atas kendala, atau feedback untuk SPV/pelapor..."
            class="w-full p-4 text-sm font-medium text-slate-800 bg-white border border-indigo-200 rounded-2xl outline-none focus:ring-2 focus:ring-indigo-500 shadow-2xs"
          ></textarea>

          <div v-if="report.reviewer_employee" class="text-xs text-indigo-700 flex items-center gap-2">
            <CheckCircle2 class="w-4 h-4 text-emerald-600" />
            <span>Terakhir ditinjau oleh <strong>{{ report.reviewer_employee.first_name }} {{ report.reviewer_employee.last_name || '' }}</strong> pada {{ report.reviewed_at ? new Date(report.reviewed_at).toLocaleString('id-ID') : '-' }}</span>
          </div>
        </div>

      </div>

      <!-- FOOTER -->
      <div class="px-6 py-4 border-t border-slate-200 bg-white flex items-center justify-between flex-wrap gap-3">
        <Button
          variant="outline"
          class="rounded-xl font-bold"
          @click="emit('update:open', false)"
        >
          Tutup
        </Button>

        <div class="flex items-center gap-2">
          <Button
            variant="outline"
            class="rounded-xl font-bold text-rose-600 border-rose-200 hover:bg-rose-50"
            :disabled="isSubmitting"
            @click="submitReview('REVISED')"
          >
            Minta Revisi ke SPV
          </Button>

          <Button
            class="rounded-xl font-extrabold gap-2 bg-emerald-600 hover:bg-emerald-700 text-white shadow-md shadow-emerald-200"
            :disabled="isSubmitting"
            @click="submitReview('APPROVED')"
          >
            <CheckCircle2 class="w-4 h-4" />
            <span>Setujui Laporan (Approve)</span>
          </Button>
        </div>
      </div>

    </div>
  </div>
</template>
