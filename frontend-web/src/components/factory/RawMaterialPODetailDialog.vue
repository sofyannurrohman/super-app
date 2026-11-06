<template>
  <Dialog :open="open" @update:open="$emit('update:open', $event)">
    <DialogContent class="sm:max-w-[780px] max-h-[90vh] overflow-y-auto">
      <DialogHeader>
        <div class="flex items-center justify-between gap-4">
          <div class="space-y-1">
            <DialogTitle class="flex items-center gap-2 text-xl font-black text-slate-900">
              <span>{{ po?.po_number || 'Detail PO Bahan Baku' }}</span>
            </DialogTitle>
            <DialogDescription>
              Detail pesanan bahan baku ke supplier dan status persetujuan berjenjang.
            </DialogDescription>
          </div>
          <div class="flex items-center gap-2">
            <Badge :variant="getStatusBadgeVariant(po?.status)" class="text-xs px-3 py-1 font-bold rounded-lg uppercase">
              {{ formatStatus(po?.status) }}
            </Badge>
          </div>
        </div>
      </DialogHeader>

      <div v-if="po" class="space-y-6 mt-2">
        <!-- Approval Stepper Timeline -->
        <div class="bg-slate-50 p-5 rounded-2xl border border-slate-100 space-y-3">
          <h4 class="text-xs font-black uppercase tracking-wider text-slate-500">Alur Persetujuan PO</h4>
          
          <div class="grid grid-cols-1 md:grid-cols-3 gap-3">
            <!-- Step 1: Admin Pabrik -->
            <div class="bg-white p-3 rounded-xl border border-slate-100 space-y-1">
              <div class="flex items-center gap-2">
                <div class="w-6 h-6 rounded-full bg-emerald-50 text-emerald-600 flex items-center justify-center font-bold text-xs">
                  ✓
                </div>
                <span class="text-xs font-bold text-slate-800">1. Admin Pabrik</span>
              </div>
              <p class="text-[11px] text-slate-500 pl-8 font-medium">
                Dibuat oleh: <strong class="text-slate-700">{{ po.created_by_name || 'Admin' }}</strong>
              </p>
              <p class="text-[10px] text-slate-400 pl-8">
                {{ formatDate(po.created_at) }}
              </p>
            </div>

            <!-- Step 2: SPV Gudang Pabrik -->
            <div class="bg-white p-3 rounded-xl border border-slate-100 space-y-1">
              <div class="flex items-center gap-2">
                <div
                  class="w-6 h-6 rounded-full flex items-center justify-center font-bold text-xs"
                  :class="
                    po.spv_approved_at
                      ? 'bg-emerald-50 text-emerald-600'
                      : po.status === 'REJECTED'
                      ? 'bg-red-50 text-red-600'
                      : 'bg-amber-50 text-amber-600'
                  "
                >
                  {{ po.spv_approved_at ? '✓' : po.status === 'REJECTED' ? '✕' : '⏳' }}
                </div>
                <span class="text-xs font-bold text-slate-800">2. SPV Gudang</span>
              </div>
              <p class="text-[11px] text-slate-500 pl-8 font-medium">
                <span v-if="po.spv_approved_at">
                  Disetujui: <strong class="text-slate-700">{{ po.spv_approved_by_name }}</strong>
                  <span v-if="po.spv_approved_by_position" class="text-slate-400 font-normal"> ({{ po.spv_approved_by_position }})</span>
                </span>
                <span v-else-if="po.status === 'REJECTED'">Ditolak</span>
                <span v-else class="text-amber-600 font-semibold">Menunggu Review SPV</span>
              </p>
              <p v-if="po.spv_approved_at" class="text-[10px] text-slate-400 pl-8">
                {{ formatDate(po.spv_approved_at) }}
              </p>
              <p v-if="po.spv_notes" class="text-[11px] text-slate-600 pl-8 italic">
                "{{ po.spv_notes }}"
              </p>
            </div>

            <!-- Step 3: Purchasing -->
            <div class="bg-white p-3 rounded-xl border border-slate-100 space-y-1">
              <div class="flex items-center gap-2">
                <div
                  class="w-6 h-6 rounded-full flex items-center justify-center font-bold text-xs"
                  :class="
                    po.purchasing_approved_at
                      ? 'bg-emerald-50 text-emerald-600'
                      : po.status === 'REJECTED'
                      ? 'bg-red-50 text-red-600'
                      : 'bg-slate-100 text-slate-400'
                  "
                >
                  {{ po.purchasing_approved_at ? '✓' : po.status === 'REJECTED' ? '✕' : '⏳' }}
                </div>
                <span class="text-xs font-bold text-slate-800">3. Purchasing</span>
              </div>
              <p class="text-[11px] text-slate-500 pl-8 font-medium">
                <span v-if="po.purchasing_approved_at">
                  Approved: <strong class="text-slate-700">{{ po.purchasing_approved_by_name }}</strong>
                  <span v-if="po.purchasing_approved_by_position" class="text-slate-400 font-normal"> ({{ po.purchasing_approved_by_position }})</span>
                </span>
                <span v-else-if="po.status === 'PENDING_PURCHASING'" class="text-amber-600 font-semibold">Menunggu Approval Purchasing</span>
                <span v-else-if="po.status === 'REJECTED'">Ditolak</span>
                <span v-else class="text-slate-400">Menunggu SPV Gudang</span>
              </p>
              <p v-if="po.purchasing_approved_at" class="text-[10px] text-slate-400 pl-8">
                {{ formatDate(po.purchasing_approved_at) }}
              </p>
              <p v-if="po.purchasing_notes" class="text-[11px] text-slate-600 pl-8 italic">
                "{{ po.purchasing_notes }}"
              </p>
            </div>
          </div>

          <!-- Rejection Banner if rejected -->
          <div v-if="po.status === 'REJECTED'" class="bg-red-50 text-red-700 p-3 rounded-xl border border-red-100 text-xs">
            <strong>Ditolak oleh {{ po.rejected_by_name || 'Petugas' }}:</strong> {{ po.rejection_reason || 'Tidak ada alasan' }}
          </div>
        </div>

        <!-- Purchasing Milestone & Progress Tracking Card -->
        <div v-if="po.status !== 'DRAFT' && po.status !== 'REJECTED'" class="p-4 rounded-2xl bg-indigo-50/70 border border-indigo-100 space-y-3">
          <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-2">
            <div class="flex items-center gap-2">
              <Truck class="w-4 h-4 text-indigo-700" />
              <h4 class="text-xs font-black uppercase tracking-wider text-indigo-950">
                Progress Pengadaan Vendor (Purchasing)
              </h4>
            </div>
            <div class="flex items-center gap-2">
              <Badge class="bg-indigo-600 text-white font-bold text-[11px]">
                {{ formatProgressStatus(po.progress_status) }}
              </Badge>
              <Button
                v-if="po.status === 'APPROVED' || po.status === 'PARTIAL_RECEIVED' || po.status === 'PENDING_PURCHASING'"
                size="sm"
                variant="outline"
                @click="openProgressModal"
                class="h-7 text-xs rounded-xl border-indigo-300 text-indigo-800 bg-white hover:bg-indigo-50 font-bold"
              >
                <Edit3 class="w-3.5 h-3.5 mr-1" /> Update Progress
              </Button>
            </div>
          </div>

          <div class="grid grid-cols-1 sm:grid-cols-3 gap-3 bg-white p-3 rounded-xl border border-indigo-100 text-xs">
            <div>
              <span class="text-slate-400 block font-semibold text-[10px] uppercase">Status Tahapan</span>
              <strong class="text-indigo-900 font-bold">{{ formatProgressStatus(po.progress_status) }}</strong>
            </div>
            <div>
              <span class="text-slate-400 block font-semibold text-[10px] uppercase">Estimasi Tiba (ETA)</span>
              <strong class="text-slate-800 font-bold">{{ po.estimated_arrival ? formatDateOnly(po.estimated_arrival) : 'Belum ditentukan' }}</strong>
            </div>
            <div>
              <span class="text-slate-400 block font-semibold text-[10px] uppercase">Diperbarui Oleh</span>
              <span class="text-slate-700 font-medium">{{ po.progress_updated_by_name || po.purchasing_approved_by_name || '-' }}</span>
              <span v-if="po.progress_updated_at" class="text-[10px] text-slate-400 block">{{ formatDate(po.progress_updated_at) }}</span>
            </div>
          </div>

          <div v-if="po.progress_notes" class="bg-white/80 p-3 rounded-xl border border-indigo-100 text-xs text-indigo-950">
            <span class="font-bold block text-[10px] uppercase text-indigo-800 mb-0.5">Catatan Progress Vendor:</span>
            <p class="italic">"{{ po.progress_notes }}"</p>
          </div>
        </div>

        <!-- PR Reference Banner if linked -->
        <div v-if="po.purchase_request || po.purchase_request_id" class="p-3 bg-blue-50/80 rounded-2xl border border-blue-200 text-xs flex items-center justify-between">
          <div class="flex items-center gap-2 text-blue-900">
            <ClipboardList class="w-4 h-4 text-blue-600" />
            <span>Referensi Purchase Request: <strong class="font-mono">{{ po.purchase_request?.pr_number || 'PR Terkait' }}</strong> (Diajukan oleh: {{ po.purchase_request?.created_by_name || 'Gudang' }})</span>
          </div>
          <Badge variant="outline" class="bg-blue-100 text-blue-800 border-blue-300 font-mono text-[10px]">
            DARI GUDANG
          </Badge>
        </div>

        <!-- Supplier & Metadata Grid -->
        <div class="grid grid-cols-2 md:grid-cols-4 gap-4 p-4 rounded-2xl bg-slate-50 border border-slate-100 text-xs">
          <div>
            <span class="text-slate-400 block font-semibold">Supplier</span>
            <strong class="text-slate-800 text-sm">{{ po.supplier?.name || '-' }}</strong>
            <span v-if="po.supplier?.phone" class="text-slate-500 block text-[11px]">{{ po.supplier.phone }}</span>
          </div>
          <div>
            <span class="text-slate-400 block font-semibold">Tgl Pengajuan</span>
            <strong class="text-slate-800 text-sm">{{ formatDateOnly(po.order_date) }}</strong>
          </div>
          <div>
            <span class="text-slate-400 block font-semibold">Status Pembayaran</span>
            <div class="flex items-center gap-1.5 mt-0.5">
              <Badge :variant="po.payment_status === 'PAID' ? 'default' : po.payment_status === 'PARTIAL_PAID' ? 'secondary' : 'outline'" class="text-[10px] font-bold">
                {{ formatPaymentStatus(po.payment_status) }}
              </Badge>
              <button
                @click="openPaymentModal"
                class="text-[10px] text-blue-600 hover:text-blue-800 font-bold underline"
              >
                Edit
              </button>
            </div>
          </div>
          <div>
            <span class="text-slate-400 block font-semibold">Total Estimasi Nilai</span>
            <strong class="text-primary text-sm font-mono font-black">
              Rp {{ (po.total_estimated_amount || 0).toLocaleString('id-ID') }}
            </strong>
          </div>
        </div>

        <!-- Items Table -->
        <div class="space-y-2">
          <h4 class="text-xs font-black uppercase tracking-wider text-slate-500">Rincian Barang / Bahan</h4>
          <div class="border border-slate-100 rounded-2xl overflow-hidden shadow-sm">
            <table class="w-full text-xs text-left">
              <thead class="bg-slate-50 text-slate-500 font-bold border-b border-slate-100">
                <tr>
                  <th class="p-3">Barang / Bahan</th>
                  <th class="p-3">Gudang</th>
                  <th class="p-3 text-center">Jumlah Dipesan</th>
                  <th class="p-3 text-center">Telah Diterima</th>
                  <th class="p-3 text-right">Harga Satuan</th>
                  <th class="p-3 text-right">Subtotal</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-slate-100">
                <tr v-for="item in po.items" :key="item.id" class="hover:bg-slate-50/50">
                  <td class="p-3">
                    <div class="font-bold text-slate-800">{{ item.product?.name }}</div>
                    <div class="text-[10px] text-slate-400 font-mono">{{ item.product?.sku }} - {{ item.product?.category }}</div>
                  </td>
                  <td class="p-3 font-medium text-slate-600">
                    {{ item.factory_warehouse?.name || 'Default Gudang' }}
                  </td>
                  <td class="p-3 text-center font-bold text-slate-800 font-mono">
                    {{ item.requested_qty.toLocaleString('id-ID') }} {{ item.unit }}
                  </td>
                  <td class="p-3 text-center font-mono">
                    <Badge
                      :variant="(item.received_qty || 0) >= item.requested_qty ? 'default' : (item.received_qty || 0) > 0 ? 'secondary' : 'outline'"
                      class="text-[10px] py-0 px-2"
                    >
                      {{ item.received_qty || 0 }} / {{ item.requested_qty }} {{ item.unit }}
                    </Badge>
                  </td>
                  <td class="p-3 text-right font-mono text-slate-600">
                    Rp {{ (item.estimated_unit_price || 0).toLocaleString('id-ID') }}
                  </td>
                  <td class="p-3 text-right font-mono font-bold text-slate-900">
                    Rp {{ ((item.requested_qty || 0) * (item.estimated_unit_price || 0)).toLocaleString('id-ID') }}
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>

        <!-- Approval / Action Actions -->
        <div class="flex flex-col sm:flex-row justify-between items-center gap-3 pt-4 border-t border-slate-100">
          <div class="flex gap-2 w-full sm:w-auto">
            <!-- Reject button if pending -->
            <Button
              v-if="po.status === 'PENDING_SPV' || po.status === 'PENDING_PURCHASING'"
              variant="outline"
              size="sm"
              class="rounded-xl border-red-200 text-red-600 hover:bg-red-50 text-xs font-bold"
              @click="showRejectPrompt = true"
            >
              <XCircle class="w-4 h-4 mr-1.5" /> Tolak PO
            </Button>
          </div>

          <div class="flex items-center gap-2 w-full sm:w-auto justify-end">
            <!-- SPV Approve -->
            <Button
              v-if="po.status === 'PENDING_SPV'"
              :disabled="actionLoading"
              class="rounded-xl font-bold bg-amber-600 hover:bg-amber-700 text-white text-xs shadow-md shadow-amber-500/20"
              @click="handleApproveSPV"
            >
              <CheckCircle class="w-4 h-4 mr-1.5" />
              Setujui sebagai SPV Gudang
            </Button>

            <!-- Purchasing Approve -->
            <Button
              v-if="po.status === 'PENDING_PURCHASING'"
              :disabled="actionLoading"
              class="rounded-xl font-bold bg-primary hover:bg-primary/90 text-white text-xs shadow-md shadow-primary/20"
              @click="handleApprovePurchasing"
            >
              <CheckCircle2 class="w-4 h-4 mr-1.5" />
              Approve PO Resmi (Purchasing)
            </Button>

            <!-- Record Receipt if Approved -->
            <Button
              v-if="po.status === 'APPROVED' || po.status === 'PARTIAL_RECEIVED'"
              class="rounded-xl font-bold bg-emerald-600 hover:bg-emerald-700 text-white text-xs shadow-md shadow-emerald-500/20"
              @click="$emit('open-receipt', po)"
            >
              <PackageCheck class="w-4 h-4 mr-1.5" />
              Catat Penerimaan Supplier
            </Button>
          </div>
        </div>

        <!-- Progress Modal (Purchasing) -->
        <Dialog :open="showProgressModal" @update:open="showProgressModal = $event">
          <DialogContent class="sm:max-w-[480px]">
            <DialogHeader>
              <DialogTitle class="flex items-center gap-2 text-indigo-700">
                <Truck class="w-5 h-5" /> Update Progress Pengadaan
              </DialogTitle>
              <DialogDescription>
                Perbarui tahapan proses pesanan vendor agar dapat dipantau oleh SPV Gudang & Pabrik.
              </DialogDescription>
            </DialogHeader>
            <div class="space-y-4 py-2">
              <div class="space-y-1.5">
                <Label class="text-xs font-bold text-slate-700">Tahapan Progress *</Label>
                <select
                  v-model="progressForm.progress_status"
                  class="flex h-10 w-full rounded-xl border border-slate-200 bg-white px-3 py-2 text-xs font-bold outline-none focus:ring-2 focus:ring-primary/20"
                >
                  <option value="PO_SENT_TO_SUPPLIER">1. PO Sudah Dikirim ke Supplier / Vendor</option>
                  <option value="IN_PRODUCTION">2. Sedang Diproses Produksi Vendor</option>
                  <option value="IN_PRINTING">3. Sedang Proses Cetak (Kardus/Label/Botol)</option>
                  <option value="SHIPPED">4. Sedang Dikirim Vendor (In-Transit)</option>
                  <option value="DELIVERED">5. Telah Tiba di Pabrik</option>
                </select>
              </div>

              <div class="space-y-1.5">
                <Label class="text-xs font-bold text-slate-700">Estimasi Kirim / Tiba di Pabrik (ETA)</Label>
                <input
                  type="date"
                  v-model="progressForm.estimated_arrival"
                  class="flex h-10 w-full rounded-xl border border-slate-200 bg-white px-3 py-2 text-xs font-bold outline-none focus:ring-2 focus:ring-primary/20"
                />
              </div>

              <div class="space-y-1.5">
                <Label class="text-xs font-bold text-slate-700">Catatan Progress dari Purchasing</Label>
                <Textarea
                  v-model="progressForm.progress_notes"
                  placeholder="Contoh: Order kardus 5.000 pcs sudah naik cetak, vendor janjikan kirim tgl 5 Sept..."
                  rows="3"
                  class="rounded-xl text-xs"
                />
              </div>
            </div>
            <DialogFooter>
              <Button variant="outline" class="rounded-xl" @click="showProgressModal = false">Batal</Button>
              <Button
                :disabled="actionLoading"
                class="rounded-xl font-bold bg-indigo-600 hover:bg-indigo-700 text-white"
                @click="handleSaveProgress"
              >
                Simpan Progress
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>

        <!-- Payment Status Modal (Finance) -->
        <Dialog :open="showPaymentModal" @update:open="showPaymentModal = $event">
          <DialogContent class="sm:max-w-[420px]">
            <DialogHeader>
              <DialogTitle class="flex items-center gap-2 text-slate-800">
                <CreditCard class="w-5 h-5 text-blue-600" /> Update Status Pembayaran (Finance)
              </DialogTitle>
              <DialogDescription>
                Informasi status pembayaran tagihan ke supplier.
              </DialogDescription>
            </DialogHeader>
            <div class="space-y-4 py-2">
              <div class="space-y-1.5">
                <Label class="text-xs font-bold text-slate-700">Status Bayar</Label>
                <select
                  v-model="paymentForm.payment_status"
                  class="flex h-10 w-full rounded-xl border border-slate-200 bg-white px-3 py-2 text-xs font-bold outline-none focus:ring-2 focus:ring-primary/20"
                >
                  <option value="UNPAID">Belum Dibayar (UNPAID)</option>
                  <option value="PARTIAL_PAID">Uang Muka / DP (PARTIAL_PAID)</option>
                  <option value="PAID">Lunas (PAID)</option>
                </select>
              </div>
              <div class="space-y-1.5">
                <Label class="text-xs font-bold text-slate-700">Catatan Pembayaran (No Rekening / Bukti Transfer)</Label>
                <Textarea
                  v-model="paymentForm.payment_notes"
                  placeholder="Contoh: DP 50% sudah ditransfer via BCA..."
                  rows="3"
                  class="rounded-xl text-xs"
                />
              </div>
            </div>
            <DialogFooter>
              <Button variant="outline" class="rounded-xl" @click="showPaymentModal = false">Batal</Button>
              <Button
                :disabled="actionLoading"
                class="rounded-xl font-bold bg-blue-600 hover:bg-blue-700 text-white"
                @click="handleSavePayment"
              >
                Simpan Pembayaran
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>

        <!-- Reject Modal -->
        <Dialog :open="showRejectPrompt" @update:open="showRejectPrompt = $event">
          <DialogContent class="sm:max-w-[420px]">
            <DialogHeader>
              <DialogTitle class="text-red-600 flex items-center gap-2">
                <AlertTriangle class="w-5 h-5" /> Tolak Purchase Order
              </DialogTitle>
              <DialogDescription>
                Berikan alasan penolakan agar admin dapat melakukan perbaikan.
              </DialogDescription>
            </DialogHeader>
            <div class="space-y-3 py-2">
              <Label class="text-xs font-bold text-slate-700">Alasan Penolakan *</Label>
              <Textarea
                v-model="rejectionReason"
                placeholder="Contoh: Kuantitas melebihi kapasitas gudang atau harga tidak sesuai..."
                rows="3"
                class="rounded-xl"
              />
            </div>
            <DialogFooter>
              <Button variant="outline" class="rounded-xl" @click="showRejectPrompt = false">Batal</Button>
              <Button
                variant="destructive"
                :disabled="!rejectionReason.trim() || actionLoading"
                class="rounded-xl font-bold"
                @click="handleReject"
              >
                Tolak PO
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </div>
    </DialogContent>
  </Dialog>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
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
import { Textarea } from '@/components/ui/textarea'
import { CheckCircle, CheckCircle2, XCircle, AlertTriangle, PackageCheck, Truck, Edit3, CreditCard, ClipboardList } from 'lucide-vue-next'
import { useFactoryStore, type RawMaterialPO } from '@/stores/factory'
import { toast } from 'vue-sonner'

const props = defineProps<{
  open: boolean
  po: RawMaterialPO | null
}>()

const emit = defineEmits<{
  (e: 'update:open', value: boolean): void
  (e: 'success'): void
  (e: 'open-receipt', po: RawMaterialPO): void
}>()

const factoryStore = useFactoryStore()
const actionLoading = ref(false)
const showRejectPrompt = ref(false)
const rejectionReason = ref('')

const showProgressModal = ref(false)
const progressForm = ref({
  progress_status: 'PO_SENT_TO_SUPPLIER',
  progress_notes: '',
  estimated_arrival: '',
})

const showPaymentModal = ref(false)
const paymentForm = ref({
  payment_status: 'UNPAID',
  payment_notes: '',
})

watch(
  () => props.po,
  (val) => {
    if (val) {
      progressForm.value = {
        progress_status: val.progress_status || 'PO_SENT_TO_SUPPLIER',
        progress_notes: val.progress_notes || '',
        estimated_arrival: val.estimated_arrival ? new Date(val.estimated_arrival).toISOString().slice(0, 10) : '',
      }
      paymentForm.value = {
        payment_status: val.payment_status || 'UNPAID',
        payment_notes: val.payment_notes || '',
      }
    }
  },
  { immediate: true }
)

const openProgressModal = () => {
  if (props.po) {
    progressForm.value = {
      progress_status: props.po.progress_status || 'PO_SENT_TO_SUPPLIER',
      progress_notes: props.po.progress_notes || '',
      estimated_arrival: props.po.estimated_arrival ? new Date(props.po.estimated_arrival).toISOString().slice(0, 10) : '',
    }
    showProgressModal.value = true
  }
}

const openPaymentModal = () => {
  if (props.po) {
    paymentForm.value = {
      payment_status: props.po.payment_status || 'UNPAID',
      payment_notes: props.po.payment_notes || '',
    }
    showPaymentModal.value = true
  }
}

const formatDate = (val?: string) => {
  if (!val) return '-'
  return new Date(val).toLocaleString('id-ID', {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}

const formatDateOnly = (val?: string) => {
  if (!val) return '-'
  return new Date(val).toLocaleDateString('id-ID', {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
  })
}

const getStatusBadgeVariant = (status?: string) => {
  switch (status) {
    case 'APPROVED':
    case 'RECEIVED':
      return 'default'
    case 'PARTIAL_RECEIVED':
    case 'PENDING_SPV':
    case 'PENDING_PURCHASING':
      return 'secondary'
    case 'REJECTED':
      return 'destructive'
    default:
      return 'outline'
  }
}

const formatStatus = (status?: string) => {
  switch (status) {
    case 'PENDING_SPV':
      return 'Menunggu SPV'
    case 'PENDING_PURCHASING':
      return 'Menunggu Purchasing'
    case 'APPROVED':
      return 'Disetujui'
    case 'PARTIAL_RECEIVED':
      return 'Diterima Sebagian'
    case 'RECEIVED':
      return 'Selesai Diterima'
    case 'REJECTED':
      return 'Ditolak'
    case 'CANCELLED':
      return 'Dibatalkan'
    default:
      return status || '-'
  }
}

const formatProgressStatus = (status?: string) => {
  switch (status) {
    case 'PENDING_APPROVAL':
      return 'Menunggu Approval PO'
    case 'PO_SENT_TO_SUPPLIER':
      return 'PO Sudah Dikirim ke Vendor'
    case 'IN_PRODUCTION':
      return 'Proses Produksi Vendor'
    case 'IN_PRINTING':
      return 'Proses Cetak di Vendor'
    case 'SHIPPED':
      return 'Sedang Dikirim Vendor (In-Transit)'
    case 'PARTIAL_DELIVERED':
      return 'Tiba Sebagian di Pabrik'
    case 'DELIVERED':
      return 'Telah Selesai Diterima'
    default:
      return status || 'Dalam Proses'
  }
}

const formatPaymentStatus = (status?: string) => {
  switch (status) {
    case 'PAID':
      return 'Lunas'
    case 'PARTIAL_PAID':
      return 'DP / Sebagian'
    case 'UNPAID':
    default:
      return 'Belum Bayar'
  }
}

const handleApproveSPV = async () => {
  if (!props.po) return
  actionLoading.value = true
  try {
    await factoryStore.approvePOBySPV(props.po.id, undefined, props.po.factory_id)
    toast.success('PO berhasil disetujui oleh SPV Gudang dan diteruskan ke Purchasing!')
    emit('success')
    emit('update:open', false)
  } catch (err: any) {
    toast.error(err?.toString() || 'Gagal menyetujui PO')
  } finally {
    actionLoading.value = false
  }
}

const handleApprovePurchasing = async () => {
  if (!props.po) return
  actionLoading.value = true
  try {
    await factoryStore.approvePOByPurchasing(props.po.id, undefined, props.po.factory_id)
    toast.success('PO resmi disetujui oleh Purchasing! Supplier dapat memproses pengiriman.')
    emit('success')
    emit('update:open', false)
  } catch (err: any) {
    toast.error(err?.toString() || 'Gagal menyetujui PO')
  } finally {
    actionLoading.value = false
  }
}

const handleSaveProgress = async () => {
  if (!props.po) return
  actionLoading.value = true
  try {
    await factoryStore.updatePOProgress(
      props.po.id,
      {
        progress_status: progressForm.value.progress_status,
        progress_notes: progressForm.value.progress_notes,
        estimated_arrival: progressForm.value.estimated_arrival || undefined,
      },
      props.po.factory_id
    )
    toast.success('Progress pengadaan vendor berhasil diperbarui!')
    showProgressModal.value = false
    emit('success')
  } catch (err: any) {
    toast.error(err?.toString() || 'Gagal memperbarui progress')
  } finally {
    actionLoading.value = false
  }
}

const handleSavePayment = async () => {
  if (!props.po) return
  actionLoading.value = true
  try {
    await factoryStore.updatePOPayment(
      props.po.id,
      {
        payment_status: paymentForm.value.payment_status,
        payment_notes: paymentForm.value.payment_notes,
      },
      props.po.factory_id
    )
    toast.success('Status pembayaran berhasil diperbarui!')
    showPaymentModal.value = false
    emit('success')
  } catch (err: any) {
    toast.error(err?.toString() || 'Gagal memperbarui pembayaran')
  } finally {
    actionLoading.value = false
  }
}

const handleReject = async () => {
  if (!props.po || !rejectionReason.value.trim()) return
  actionLoading.value = true
  try {
    await factoryStore.rejectRawMaterialPO(props.po.id, rejectionReason.value, props.po.factory_id)
    toast.info('PO telah ditolak')
    showRejectPrompt.value = false
    emit('success')
    emit('update:open', false)
  } catch (err: any) {
    toast.error(err?.toString() || 'Gagal menolak PO')
  } finally {
    actionLoading.value = false
  }
}
</script>
