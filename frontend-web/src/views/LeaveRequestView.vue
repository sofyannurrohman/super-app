<script setup lang="ts">
import { ref, h, onMounted } from 'vue'
import { toast } from 'vue-sonner'
import apiClient from '@/api/axios'
import DataTable from '@/components/DataTable.vue'
import ConfirmationDialog from '@/components/ui/ConfirmationDialog.vue'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'

const leaveRequests = ref<any[]>([])
const isLoading = ref(true)
const activeFilter = ref('ALL')

const statusFilters = ['ALL', 'PENDING', 'APPROVED', 'REJECTED']

// Dialog States
const showApproveDialog = ref(false)
const showRejectDialog = ref(false)
const selectedRequestId = ref<string | null>(null)
const rejectReason = ref('')
const isProcessing = ref(false)

const openApproveDialog = (id: string) => {
  selectedRequestId.value = id
  showApproveDialog.value = true
}

const openRejectDialog = (id: string) => {
  selectedRequestId.value = id
  rejectReason.value = ''
  showRejectDialog.value = true
}

const handleApproveConfirm = async () => {
  if (!selectedRequestId.value) return
  isProcessing.value = true
  try {
    await apiClient.put(`/time-off/manage/${selectedRequestId.value}/approve`, { status: 'APPROVED', reject_reason: '' })
    toast.success('Pengajuan cuti/izin berhasil disetujui!')
    fetchLeaveRequests()
  } catch (err: any) {
    toast.error('Gagal menyetujui: ' + (err.response?.data?.error || err.message))
  } finally {
    isProcessing.value = false
    showApproveDialog.value = false
    selectedRequestId.value = null
  }
}

const handleRejectConfirm = async () => {
  if (!selectedRequestId.value) return
  isProcessing.value = true
  try {
    await apiClient.put(`/time-off/manage/${selectedRequestId.value}/approve`, { status: 'REJECTED', reject_reason: rejectReason.value })
    toast.success('Pengajuan cuti/izin berhasil ditolak!')
    fetchLeaveRequests()
  } catch (err: any) {
    toast.error('Gagal menolak: ' + (err.response?.data?.error || err.message))
  } finally {
    isProcessing.value = false
    showRejectDialog.value = false
    selectedRequestId.value = null
  }
}

const fetchLeaveRequests = async () => {
  isLoading.value = true
  try {
    const params = activeFilter.value !== 'ALL' ? `?status=${activeFilter.value}` : ''
    const res = await apiClient.get(`/time-off/manage${params}`)
    leaveRequests.value = res.data.data
  } catch (e) {
    console.error(e)
  } finally {
    isLoading.value = false
  }
}

const handleFilterChange = (filter: string) => {
  activeFilter.value = filter
  fetchLeaveRequests()
}

onMounted(fetchLeaveRequests)

const columns = [
  {
    id: 'employee',
    header: 'KARYAWAN',
    cell: ({ row }: any) => {
      const emp = row.original.employee
      return h(
        'span',
        { class: 'font-bold text-gray-900' },
        emp ? `${emp.first_name} ${emp.last_name || ''}`.trim() : '-',
      )
    },
  },
  {
    id: 'leaveType',
    header: 'JENIS CUTI',
    cell: ({ row }: any) => {
      const lt = row.original.leave_type
      return h('span', { class: 'text-gray-700' }, lt?.name || '-')
    },
  },
  {
    id: 'duration',
    header: 'PERIODE',
    cell: ({ row }: any) => {
      const r = row.original
      const fmt = (d: string) => new Date(d).toLocaleDateString('id-ID', { dateStyle: 'medium' })
      return h(
        'span',
        { class: 'text-[13px] text-gray-600' },
        `${fmt(r.start_date)} - ${fmt(r.end_date)}`,
      )
    },
  },
  {
    accessorKey: 'reason',
    header: 'ALASAN',
    cell: (info: any) =>
      h(
        'span',
        {
          class: 'text-gray-500 text-[13px] max-w-[160px] truncate inline-block',
          title: info.getValue(),
        },
        info.getValue() || '-',
      ),
  },
  {
    accessorKey: 'created_at',
    id: 'created_at',
    header: 'DIBUAT PADA',
    cell: ({ row }: any) => {
      const d = row.original.created_at
      if (!d) return h('span', { class: 'text-gray-400 text-[13px]' }, '-')
      const date = new Date(d)
      const dateStr = date.toLocaleDateString('id-ID', { day: '2-digit', month: 'short', year: 'numeric' })
      const timeStr = date.toLocaleTimeString('id-ID', { hour: '2-digit', minute: '2-digit' })
      return h('div', { class: 'flex flex-col' }, [
        h('span', { class: 'text-gray-700 text-[13px] font-medium' }, dateStr),
        h('span', { class: 'text-gray-500 text-[11px]' }, timeStr),
      ])
    },
  },
  {
    accessorKey: 'status',
    header: 'STATUS',
    cell: ({ row, getValue }: any) => {
      const s = getValue() as string
      const cls =
        s === 'APPROVED'
          ? 'border-green-200 text-green-700 bg-green-50'
          : s === 'REJECTED'
            ? 'border-red-200 text-red-600 bg-red-50'
            : 'border-yellow-200 text-yellow-700 bg-yellow-50'

      const approver = row.original.approver
      const approverName = approver ? `${approver.first_name || ''} ${approver.last_name || ''}`.trim() : ''
      const rejectReason = row.original.reject_reason || ''

      return h('div', { class: 'flex flex-col items-start gap-1 max-w-[200px]' }, [
        h('span', { class: `border ${cls} px-2.5 py-0.5 rounded-full text-[11px] font-bold` }, s),
        s === 'APPROVED' && approverName
          ? h('span', { class: 'text-[10px] text-gray-500 font-medium' }, `✓ ${approverName}`)
          : null,
        s === 'REJECTED'
          ? h('div', { class: 'flex flex-col text-[10px] text-red-600 gap-0.5' }, [
              approverName ? h('span', { class: 'font-semibold' }, `✕ ${approverName}`) : null,
              rejectReason ? h('span', { class: 'bg-red-50 border border-red-100 rounded px-1.5 py-0.5 italic' }, `"${rejectReason}"`) : null,
            ])
          : null,
      ])
    },
  },
  {
    id: 'actions',
    header: 'AKSI',
    cell: ({ row }: any) => {
      const r = row.original
      if (r.status !== 'PENDING')
        return h('span', { class: 'text-gray-400 text-[13px]' }, 'Sudah Diproses')
      return h('div', { class: 'flex gap-2' }, [
        h(
          'button',
          {
            class: 'text-[13px] font-semibold text-green-700 hover:underline',
            onClick: () => openApproveDialog(r.id),
          },
          'Setujui',
        ),
        h(
          'button',
          {
            class: 'text-[13px] font-semibold text-red-600 hover:underline',
            onClick: () => openRejectDialog(r.id),
          },
          'Tolak',
        ),
      ])
    },
  },
]
</script>

<template>
  <div class="space-y-6">
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
      <div>
        <h1 class="text-[22px] font-bold text-gray-900">Pengajuan Cuti</h1>
        <p class="text-[14px] text-gray-500 mt-1">Review dan proses pengajuan cuti karyawan.</p>
      </div>
    </div>

    <!-- Filter tabs -->
    <div class="flex gap-2 flex-wrap">
      <button
        v-for="f in statusFilters"
        :key="f"
        @click="handleFilterChange(f)"
        :class="[
          'px-4 py-1.5 rounded-full text-[13px] font-semibold border transition-colors',
          activeFilter === f
            ? 'bg-primary text-white border-primary'
            : 'bg-white text-gray-600 border-gray-300 hover:border-blue-400',
        ]"
      >
        {{ f }}
      </button>
    </div>

    <DataTable :data="leaveRequests" :columns="columns" :isLoading="isLoading" />

    <!-- Approval Confirmation Dialog -->
    <ConfirmationDialog
      v-model:open="showApproveDialog"
      title="Setujui Pengajuan Cuti"
      description="Apakah Anda yakin ingin menyetujui pengajuan cuti/izin ini?"
      variant="success"
      confirmText="Setujui"
      cancelText="Batal"
      :isLoading="isProcessing"
      @confirm="handleApproveConfirm"
      @cancel="showApproveDialog = false"
    />

    <!-- Rejection Dialog with Textarea -->
    <Dialog v-model:open="showRejectDialog">
      <DialogContent class="max-w-[420px] rounded-[2.5rem] border-none shadow-2xl p-0 overflow-hidden">
        <div class="p-10 space-y-6">
          <div class="space-y-3 text-center">
            <DialogTitle class="text-2xl font-black text-slate-900 tracking-tight leading-tight">
              Tolak Pengajuan Cuti
            </DialogTitle>
            <DialogDescription class="text-sm font-bold text-slate-400 leading-relaxed">
              Silakan masukkan alasan penolakan untuk pengajuan cuti/izin ini.
            </DialogDescription>
          </div>

          <div class="space-y-1.5">
            <label class="text-xs font-bold text-slate-600 uppercase">Alasan Penolakan</label>
            <textarea
              v-model="rejectReason"
              placeholder="Masukkan alasan penolakan..."
              rows="3"
              class="flex w-full rounded-2xl border border-input bg-transparent px-4 py-3 text-sm shadow-sm transition-all focus:outline-none focus:ring-1 focus:ring-ring"
            ></textarea>
          </div>
        </div>

        <DialogFooter class="bg-slate-50/80 backdrop-blur-sm p-8 flex gap-4 border-t border-slate-100/50">
          <Button
            variant="ghost"
            @click="showRejectDialog = false"
            class="flex-1 rounded-2xl font-bold h-14 text-slate-500 hover:bg-slate-100 transition-all"
            :disabled="isProcessing"
          >
            Batal
          </Button>
          <Button
            @click="handleRejectConfirm"
            class="flex-1 rounded-2xl font-black h-14 bg-red-600 hover:bg-red-700 text-white shadow-2xl transition-all hover:scale-[1.02] active:scale-[0.98]"
            :disabled="isProcessing"
          >
            Tolak
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  </div>
</template>
