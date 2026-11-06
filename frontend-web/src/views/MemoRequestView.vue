<script setup lang="ts">
import { ref, h, onMounted } from 'vue'
import { toast } from 'vue-sonner'
import memoApi from '@/api/memoRequest'
import DataTable from '@/components/DataTable.vue'
import ConfirmationDialog from '@/components/ui/ConfirmationDialog.vue'

const memoRequests = ref<any[]>([])
const isLoading = ref(true)
const activeFilter = ref('ALL')

const statusFilters = ['ALL', 'PENDING', 'APPROVED']

// Dialog States
const showApproveDialog = ref(false)
const selectedRequestId = ref<string | null>(null)
const isProcessing = ref(false)

const openApproveDialog = (id: string) => {
  selectedRequestId.value = id
  showApproveDialog.value = true
}

const handleApproveConfirm = async () => {
  if (!selectedRequestId.value) return
  isProcessing.value = true
  try {
    await memoApi.approveMemoRequest(selectedRequestId.value)
    toast.success('Pengajuan Memo berhasil disetujui dan nomor telah di-generate!')
    fetchMemoRequests()
  } catch (err: any) {
    toast.error('Gagal menyetujui: ' + (err.response?.data?.error || err.message))
  } finally {
    isProcessing.value = false
    showApproveDialog.value = false
    selectedRequestId.value = null
  }
}

const fetchMemoRequests = async () => {
  isLoading.value = true
  try {
    const params: any = {}
    if (activeFilter.value !== 'ALL') {
      params.status = activeFilter.value
    }
    const res = await memoApi.getMemoRequests(params)
    memoRequests.value = res.data.data
  } catch (e) {
    console.error(e)
  } finally {
    isLoading.value = false
  }
}

const handleFilterChange = (filter: string) => {
  activeFilter.value = filter
  fetchMemoRequests()
}

onMounted(fetchMemoRequests)

const columns = [
  {
    accessorKey: 'type',
    header: 'TIPE',
    cell: (info: any) => h('span', { class: 'font-bold uppercase text-[12px]' }, info.getValue() || '-'),
  },
  {
    id: 'branch',
    header: 'CABANG',
    cell: ({ row }: any) => h('span', { class: 'text-gray-700' }, row.original.branch?.name || 'Pusat/Umum'),
  },
  {
    accessorKey: 'destination',
    header: 'TUJUAN',
    cell: (info: any) => h('span', { class: 'text-gray-700' }, info.getValue() || '-'),
  },
  {
    accessorKey: 'subject',
    header: 'PERIHAL',
    cell: (info: any) => h('span', { class: 'text-gray-700 font-medium' }, info.getValue() || '-'),
  },
  {
    accessorKey: 'requester_name',
    header: 'PEMOHON',
    cell: (info: any) => h('span', { class: 'text-gray-700' }, info.getValue() || '-'),
  },
  {
    accessorKey: 'memo_number',
    header: 'NOMOR MEMO',
    cell: (info: any) => h('span', { class: 'text-primary font-bold' }, info.getValue() || '-'),
  },
  {
    accessorKey: 'status',
    header: 'STATUS',
    cell: ({ row, getValue }: any) => {
      const s = getValue()
      const cls =
        s === 'APPROVED'
          ? 'border-green-200 text-green-700 bg-green-50'
          : s === 'REJECTED'
            ? 'border-red-200 text-red-600 bg-red-50'
            : 'border-yellow-200 text-yellow-700 bg-yellow-50'

      const approver = row.original.approver
      const approverName = approver ? `${approver.first_name || ''} ${approver.last_name || ''}`.trim() : ''
      const approverPosition = approver?.job_position?.title || ''

      return h('div', { class: 'flex flex-col items-start gap-0.5' }, [
        h('span', { class: `border ${cls} px-2.5 py-0.5 rounded-full text-[11px] font-bold inline-block` }, s),
        s === 'APPROVED' && approverName
          ? h('div', { class: 'flex flex-col text-[10px] text-gray-500 mt-0.5 leading-tight' }, [
              h('span', { class: 'font-semibold text-gray-700' }, `✓ ${approverName}`),
              approverPosition ? h('span', { class: 'text-gray-400 text-[9px]' }, approverPosition) : null,
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
      ])
    },
  },
]
</script>

<template>
  <div class="space-y-6">
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
      <div>
        <h1 class="text-[22px] font-bold text-gray-900">Request Nomor Surat / Memo</h1>
        <p class="text-[14px] text-gray-500 mt-1">Review dan setujui request penomoran surat internal dan eksternal.</p>
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

    <DataTable :data="memoRequests" :columns="columns" :isLoading="isLoading" />

    <!-- Approval Confirmation Dialog -->
    <ConfirmationDialog
      v-model:open="showApproveDialog"
      title="Setujui Penomoran Memo"
      description="Apakah Anda yakin ingin menyetujui request ini? Nomor urut akan otomatis ter-generate dan tersimpan."
      variant="success"
      confirmText="Setujui & Generate"
      cancelText="Batal"
      :isLoading="isProcessing"
      @confirm="handleApproveConfirm"
      @cancel="showApproveDialog = false"
    />
  </div>
</template>
