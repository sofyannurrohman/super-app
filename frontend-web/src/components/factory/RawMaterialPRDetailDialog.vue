<template>
  <Dialog :open="open" @update:open="$emit('update:open', $event)">
    <DialogContent class="sm:max-w-[760px] max-h-[90vh] overflow-y-auto">
      <DialogHeader>
        <div class="flex items-center justify-between gap-4">
          <div class="space-y-1">
            <DialogTitle class="flex items-center gap-2 text-xl font-black text-slate-900">
              <ClipboardList class="w-6 h-6 text-primary" />
              <span>{{ pr?.pr_number || 'Detail Purchase Request' }}</span>
            </DialogTitle>
            <DialogDescription>
              Permohonan pembelian untuk ditindaklanjuti Purchasing.
            </DialogDescription>
          </div>
          <div class="flex items-center gap-2">
            <Badge :variant="getStatusBadgeVariant(pr?.status)" class="text-xs px-3 py-1 font-bold rounded-lg uppercase">
              {{ formatStatus(pr?.status) }}
            </Badge>
          </div>
        </div>
      </DialogHeader>

      <div v-if="pr" class="space-y-6 mt-2">
        <!-- Info Cards -->
        <div class="grid grid-cols-2 sm:grid-cols-4 gap-3 bg-slate-50 p-4 rounded-2xl border border-slate-100 text-xs">
          <div class="space-y-1">
            <span class="text-slate-400 font-bold uppercase text-[10px]">Asal Permintaan</span>
            <div class="font-bold text-slate-800">{{ pr.factory?.name || 'Karyawan' }}</div>
          </div>
          <div class="space-y-1">
            <span class="text-slate-400 font-bold uppercase text-[10px]">Tanggal Pengajuan</span>
            <div class="font-bold text-slate-800">{{ formatDateOnly(pr.request_date) }}</div>
          </div>
          <div class="space-y-1">
            <span class="text-slate-400 font-bold uppercase text-[10px]">Tgl Dibutuhkan</span>
            <div class="font-bold text-slate-800">{{ pr.required_date ? formatDateOnly(pr.required_date) : '-' }}</div>
          </div>
          <div class="space-y-1">
            <span class="text-slate-400 font-bold uppercase text-[10px]">Pemohon</span>
            <div class="font-bold text-slate-800">{{ pr.created_by_name || '-' }}</div>
          </div>
        </div>

        <!-- Notes -->
        <div v-if="pr.notes" class="p-3.5 bg-blue-50/60 rounded-2xl border border-blue-100 text-xs">
          <span class="font-bold text-blue-900 block mb-1">Catatan / Alasan Permohonan:</span>
          <p class="text-slate-700 italic">{{ pr.notes }}</p>
        </div>

        <!-- Rejection Box -->
        <div v-if="pr.status === 'REJECTED' && pr.rejection_reason" class="p-3.5 bg-red-50 rounded-2xl border border-red-200 text-xs">
          <div class="flex items-center gap-1.5 font-bold text-red-900 mb-1">
            <AlertCircle class="w-4 h-4 text-red-600" />
            <span>Alasan Penolakan oleh Purchasing:</span>
          </div>
          <p class="text-red-700 font-medium pl-5">{{ pr.rejection_reason }}</p>
          <div v-if="pr.reviewed_by_name" class="text-[10px] text-red-500 mt-1 pl-5">
            Ditolak oleh: {{ pr.reviewed_by_name }} ({{ formatDate(pr.reviewed_at) }})
          </div>
        </div>

        <!-- Processed to PO Info -->
        <div v-if="pr.status === 'PROCESSED' || pr.purchase_order_id" class="p-3.5 bg-emerald-50 rounded-2xl border border-emerald-200 text-xs flex items-center justify-between">
          <div class="space-y-0.5">
            <div class="flex items-center gap-1.5 font-bold text-emerald-900">
              <CheckCircle class="w-4 h-4 text-emerald-600" />
              <span>{{ pr.purchase_order_id ? 'Sudah Diproses ke Purchase Order (PO)' : 'Sudah Ditindaklanjuti Purchasing' }}</span>
            </div>
            <p v-if="pr.purchase_order?.po_number" class="text-emerald-700 text-[11px] font-mono pl-5">
              No. PO: <strong>{{ pr.purchase_order.po_number }}</strong>
            </p>
            <p v-if="pr.review_notes" class="text-emerald-700 text-[11px] pl-5">{{ pr.review_notes }}</p>
          </div>
          <Badge variant="outline" class="bg-emerald-100 text-emerald-800 border-emerald-300 font-mono text-[10px]">
            TERPROSES
          </Badge>
        </div>

        <!-- Items Table -->
        <div class="space-y-2">
          <h4 class="text-xs font-black uppercase tracking-wider text-slate-500">
            Daftar Barang Diminta ({{ pr.items?.length || 0 }} Item)
          </h4>
          <div class="border border-slate-200 rounded-2xl overflow-hidden bg-white shadow-xs">
            <table class="w-full text-left text-xs">
              <thead class="bg-slate-50 border-b border-slate-100 text-slate-500 font-bold uppercase text-[10px]">
                <tr>
                  <th class="p-3">No</th>
                  <th class="p-3">Barang</th>
                  <th v-if="pr.factory_id" class="p-3">Gudang Tujuan</th>
                  <th class="p-3 text-right">Jumlah Kuantitas</th>
                  <th class="p-3">Catatan</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-slate-100">
                <tr v-for="(it, idx) in pr.items" :key="it.id || idx" class="hover:bg-slate-50/50">
                  <td class="p-3 font-mono text-slate-400 font-bold">{{ Number(idx) + 1 }}</td>
                  <td class="p-3">
                    <div class="font-bold text-slate-900">{{ it.item_name || it.product?.name || '-' }}</div>
                    <div v-if="it.product" class="text-[10px] text-slate-400 font-mono">{{ it.product?.sku }} • {{ it.product?.category || 'BARANG' }}</div>
                  </td>
                  <td v-if="pr.factory_id" class="p-3 text-slate-600">
                    {{ it.factory_warehouse?.name || 'Gudang Utama' }}
                  </td>
                  <td class="p-3 text-right">
                    <Badge variant="secondary" class="font-mono font-black text-xs bg-slate-900 text-white px-2 py-0.5 rounded-lg">
                      {{ (it.requested_qty || 0).toLocaleString('id-ID') }} {{ it.unit }}
                    </Badge>
                  </td>
                  <td class="p-3 text-slate-500 text-[11px]">
                    {{ it.notes || '-' }}
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>

        <!-- Purchasing Review Section (Only for Purchasing & Superadmin when PENDING) -->
        <div
          v-if="canReview && pr.status === 'PENDING'"
          class="p-4 bg-primary/5 rounded-2xl border border-primary/20 space-y-3"
        >
          <div class="flex items-center gap-2 text-primary font-bold text-xs">
            <ShoppingCart class="w-4 h-4" />
            <span>Tindakan Staff Purchasing</span>
          </div>
          <p class="text-xs text-slate-600 leading-relaxed">
            Catat tindak lanjut pembelian, proses menjadi PO untuk permintaan pabrik, atau tolak dengan alasan.
          </p>
          <div class="flex items-center gap-2 pt-1">
            <Button
              v-if="pr.factory_id && pr.items?.every((it: any) => it.product_id)"
              class="rounded-xl font-bold bg-primary hover:bg-primary/90 text-white text-xs shadow-md shadow-primary/20"
              @click="$emit('convert-to-po', pr)"
            >
              <ShoppingCart class="w-3.5 h-3.5 mr-1.5" /> Proses Jadi Purchase Order (PO)
            </Button>
            <Button v-else class="rounded-xl font-bold text-xs" @click="$emit('complete', pr)">
              <CheckCircle class="w-3.5 h-3.5 mr-1.5" /> Catat Tindak Lanjut
            </Button>
            <Button
              variant="outline"
              class="rounded-xl font-bold border-red-200 text-red-600 hover:bg-red-50 text-xs"
              @click="showRejectModal = true"
            >
              <XCircle class="w-3.5 h-3.5 mr-1.5" /> Tolak Permintaan
            </Button>
          </div>
        </div>
      </div>

      <DialogFooter class="gap-2 sm:gap-0 pt-2">
        <Button variant="outline" class="rounded-xl font-bold" @click="$emit('update:open', false)">
          Tutup
        </Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>

  <!-- Reject Reason Modal -->
  <Dialog :open="showRejectModal" @update:open="showRejectModal = $event">
    <DialogContent class="sm:max-w-[420px]">
      <DialogHeader>
        <DialogTitle class="text-red-600 flex items-center gap-2 text-base font-bold">
          <XCircle class="w-5 h-5" /> Tolak Purchase Request
        </DialogTitle>
        <DialogDescription class="text-xs">
          Berikan alasan penolakan agar pemohon mengetahui alasan request tidak dapat diproses.
        </DialogDescription>
      </DialogHeader>
      <div class="space-y-3 py-2">
        <Label for="reject_reason" class="text-xs font-bold text-slate-700">Alasan Penolakan *</Label>
        <textarea
          id="reject_reason"
          v-model="rejectReason"
          rows="3"
          placeholder="Contoh: Stok di gudang sentral masih mencukupi, atau revisi jumlah kuantitas..."
          class="flex w-full rounded-xl border border-slate-200 bg-white px-3 py-2 text-xs focus:outline-none focus:ring-2 focus:ring-red-500/20 font-medium"
          required
        ></textarea>
      </div>
      <DialogFooter>
        <Button variant="outline" size="sm" class="rounded-xl font-bold" @click="showRejectModal = false">
          Batal
        </Button>
        <Button
          size="sm"
          :disabled="rejecting || !rejectReason.trim()"
          class="rounded-xl font-bold bg-red-600 hover:bg-red-700 text-white"
          @click="submitReject"
        >
          <Loader2 v-if="rejecting" class="w-3.5 h-3.5 mr-1.5 animate-spin" />
          Konfirmasi Tolak
        </Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Label } from '@/components/ui/label'
import { ClipboardList, AlertCircle, CheckCircle, ShoppingCart, XCircle, Loader2 } from 'lucide-vue-next'
import { useAuthStore } from '@/stores/auth'
import { useFactoryStore } from '@/stores/factory'
import { toast } from 'vue-sonner'

const props = defineProps<{
  open: boolean
  pr?: any | null
}>()

const emit = defineEmits<{
  (e: 'update:open', value: boolean): void
  (e: 'convert-to-po', pr: any): void
  (e: 'complete', pr: any): void
  (e: 'success'): void
}>()

const authStore = useAuthStore()
const factoryStore = useFactoryStore()

const showRejectModal = ref(false)
const rejectReason = ref('')
const rejecting = ref(false)

const userRole = computed(() => authStore.user?.role || '')
const canReview = computed(() => {
  return ['superadmin', 'admin_purchasing', 'manager'].includes(userRole.value)
})

const formatDateOnly = (dateStr: any) => {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleDateString('id-ID', {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
  })
}

const formatDate = (dateStr: any) => {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleString('id-ID', {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}

const formatStatus = (status?: string) => {
  switch (status) {
    case 'PENDING':
      return 'Menunggu Review Purchasing'
    case 'PROCESSED':
      return 'Sudah Ditindaklanjuti'
    case 'REJECTED':
      return 'Ditolak'
    case 'CANCELLED':
      return 'Dibatalkan'
    default:
      return status || 'PENDING'
  }
}

const getStatusBadgeVariant = (status?: string): any => {
  switch (status) {
    case 'PENDING':
      return 'secondary'
    case 'PROCESSED':
      return 'default'
    case 'REJECTED':
      return 'destructive'
    default:
      return 'outline'
  }
}

const submitReject = async () => {
  if (!props.pr?.id || !rejectReason.value.trim()) return
  rejecting.value = true
  try {
    await factoryStore.rejectRawMaterialPR(props.pr.id, rejectReason.value.trim(), props.pr.factory_id)
    toast.success('Purchase Request berhasil ditolak')
    showRejectModal.value = false
    rejectReason.value = ''
    emit('success')
    emit('update:open', false)
  } catch (err: any) {
    toast.error(typeof err === 'string' ? err : err.message || 'Gagal menolak PR')
  } finally {
    rejecting.value = false
  }
}
</script>
