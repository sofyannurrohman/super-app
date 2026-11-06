<template>
  <div class="p-6 space-y-6 bg-[#f8fafc] min-h-screen">
    <!-- Header -->
    <div
      class="flex flex-col md:flex-row justify-between items-start md:items-center gap-4 bg-white p-6 rounded-3xl shadow-sm border border-slate-100"
    >
      <div class="flex items-center gap-4">
        <div class="w-12 h-12 rounded-2xl bg-indigo-50 flex items-center justify-center text-indigo-600 shadow-inner">
          <FlaskConical class="h-6 w-6" />
        </div>
        <div>
          <div class="flex items-center gap-2 text-xs font-bold text-indigo-600 uppercase tracking-widest mb-0.5">
            <span>Quality Control & Lab</span>
          </div>
          <h1 class="text-2xl font-black tracking-tight text-slate-900">
            Sampel Lab & Uji QC Pabrik
          </h1>
          <p class="text-sm text-slate-500">
            Pengeluaran sampel non-komersial untuk QC produksi dengan validasi batch dan approval Admin Gudang Pabrik.
          </p>
        </div>
      </div>

      <div class="flex items-center gap-3 w-full md:w-auto">
        <Button
          variant="outline"
          @click="loadData"
          :disabled="loading"
          class="rounded-xl border-slate-200 hover:bg-slate-50"
        >
          <RefreshCw class="mr-2 h-4 w-4" :class="{ 'animate-spin': loading }" />
          Refresh
        </Button>
        <Button
          @click="openRequestModal"
          class="rounded-xl bg-indigo-600 hover:bg-indigo-700 shadow-lg shadow-indigo-200 transition-all font-bold"
        >
          <Plus class="mr-2 h-4 w-4" /> Ajukan Sampel QC
        </Button>
      </div>
    </div>

    <!-- KPI Summary Cards -->
    <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
      <Card class="border-none shadow-sm bg-white rounded-3xl p-5 flex items-center gap-4">
        <div class="w-12 h-12 rounded-2xl bg-slate-100 text-slate-700 flex items-center justify-center">
          <FlaskConical class="h-6 w-6" />
        </div>
        <div>
          <p class="text-xs font-bold text-slate-400 uppercase tracking-wider">Total Permohonan</p>
          <p class="text-2xl font-black text-slate-900">{{ factoryStore.qcSamples.length }}</p>
        </div>
      </Card>

      <Card class="border-none shadow-sm bg-white rounded-3xl p-5 flex items-center gap-4">
        <div class="w-12 h-12 rounded-2xl bg-amber-50 text-amber-600 flex items-center justify-center">
          <Clock class="h-6 w-6" />
        </div>
        <div>
          <p class="text-xs font-bold text-amber-600 uppercase tracking-wider">Menunggu Gudang</p>
          <p class="text-2xl font-black text-amber-700">{{ pendingCount }}</p>
        </div>
      </Card>

      <Card class="border-none shadow-sm bg-white rounded-3xl p-5 flex items-center gap-4">
        <div class="w-12 h-12 rounded-2xl bg-emerald-50 text-emerald-600 flex items-center justify-center">
          <CheckCircle2 class="h-6 w-6" />
        </div>
        <div>
          <p class="text-xs font-bold text-emerald-600 uppercase tracking-wider">Disetujui (Potong Stok)</p>
          <p class="text-2xl font-black text-emerald-700">{{ approvedCount }}</p>
        </div>
      </Card>

      <Card class="border-none shadow-sm bg-white rounded-3xl p-5 flex items-center gap-4">
        <div class="w-12 h-12 rounded-2xl bg-rose-50 text-rose-600 flex items-center justify-center">
          <XCircle class="h-6 w-6" />
        </div>
        <div>
          <p class="text-xs font-bold text-rose-600 uppercase tracking-wider">Ditolak</p>
          <p class="text-2xl font-black text-rose-700">{{ rejectedCount }}</p>
        </div>
      </Card>
    </div>

    <!-- Filters & Table -->
    <Card class="border-none shadow-sm bg-white rounded-3xl overflow-hidden">
      <div class="p-6 border-b border-slate-100 flex flex-col md:flex-row justify-between items-start md:items-center gap-4">
        <div class="flex items-center gap-3">
          <div class="flex items-center gap-2 bg-slate-50 border border-slate-200/80 rounded-2xl p-1.5">
            <button
              v-for="statusTab in statusTabs"
              :key="statusTab.value"
              @click="activeStatusFilter = statusTab.value"
              :class="[
                'px-4 py-2 text-xs font-bold rounded-xl transition-all',
                activeStatusFilter === statusTab.value
                  ? 'bg-white text-indigo-700 shadow-sm'
                  : 'text-slate-500 hover:text-slate-900'
              ]"
            >
              {{ statusTab.label }}
              <span
                v-if="statusTab.count > 0"
                :class="[
                  'ml-1 px-1.5 py-0.5 rounded-full text-[10px]',
                  activeStatusFilter === statusTab.value ? 'bg-indigo-100 text-indigo-800' : 'bg-slate-200/60 text-slate-600'
                ]"
              >
                {{ statusTab.count }}
              </span>
            </button>
          </div>
        </div>

        <div class="flex items-center gap-3 w-full md:w-auto">
          <!-- Factory Select -->
          <select
            v-model="selectedFactoryId"
            @change="loadData"
            class="px-4 py-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-bold text-slate-700 outline-none"
          >
            <option value="">Semua Unit Pabrik</option>
            <option v-for="f in factoryStore.factories" :key="f.id" :value="f.id">
              {{ f.name }}
            </option>
          </select>

          <!-- Search Input -->
          <div class="relative w-full md:w-64">
            <Search class="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-slate-400" />
            <input
              type="text"
              v-model="searchQuery"
              placeholder="Cari produk / batch / pemohon..."
              class="w-full pl-9 pr-4 py-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-bold outline-none focus:ring-2 focus:ring-indigo-500/20"
            />
          </div>
        </div>
      </div>

      <DataTable :columns="columns" :data="filteredSamples" class="border-none">
        <template #cell-request_info="{ row }">
          <div>
            <div class="font-bold text-slate-900 text-sm flex items-center gap-1.5">
              <span>{{ (row as any).requested_by_name || 'Admin Pabrik' }}</span>
              <Badge variant="outline" class="text-[10px] font-bold py-0 bg-slate-50 text-slate-600 border-slate-200">
                {{ (row as any).factory?.name || 'Pabrik' }}
              </Badge>
            </div>
            <p class="text-xs text-slate-400 mt-0.5">
              {{ formatDate((row as any).requested_at || (row as any).created_at) }}
            </p>
          </div>
        </template>

        <template #cell-product="{ row }">
          <div class="flex items-center gap-3">
            <div class="w-9 h-9 rounded-xl bg-slate-100 flex items-center justify-center text-slate-600 font-black text-xs">
              <Package class="h-4 w-4 text-slate-500" />
            </div>
            <div>
              <p class="font-bold text-slate-900 text-sm">{{ (row as any).product?.name || 'Produk' }}</p>
              <p class="text-xs text-slate-400">SKU: {{ (row as any).product?.sku || '-' }}</p>
            </div>
          </div>
        </template>

        <template #cell-batch="{ row }">
          <div>
            <div class="inline-flex items-center gap-1.5 px-2.5 py-1 bg-amber-50 border border-amber-200 rounded-lg text-amber-900 font-mono text-xs font-bold">
              <Layers class="h-3.5 w-3.5 text-amber-600" />
              <span>{{ (row as any).batch_no }}</span>
            </div>
            <div v-if="(row as any).expiry_date" class="text-[11px] text-slate-400 mt-1">
              Exp: {{ formatSimpleDate((row as any).expiry_date) }}
            </div>
          </div>
        </template>

        <template #cell-quantity="{ row }">
          <div class="flex items-center gap-1.5">
            <span class="text-base font-black text-slate-900">{{ (row as any).quantity }}</span>
            <span class="text-xs font-bold text-slate-400 uppercase">{{ (row as any).product?.unit || 'PCS' }}</span>
          </div>
        </template>

        <template #cell-purpose="{ row }">
          <div class="max-w-xs">
            <Badge variant="secondary" class="bg-indigo-50 text-indigo-700 text-[10px] font-bold mb-1">
              {{ (row as any).sample_type || 'LAB_QC' }}
            </Badge>
            <p class="text-xs text-slate-600 truncate" :title="(row as any).purpose">
              {{ (row as any).purpose || '-' }}
            </p>
          </div>
        </template>

        <template #cell-status="{ row }">
          <div class="space-y-1">
            <Badge
              v-if="(row as any).status === 'APPROVED'"
              class="bg-emerald-50 text-emerald-700 border border-emerald-200 font-bold text-xs"
            >
              <CheckCircle2 class="mr-1 h-3.5 w-3.5" /> Disetujui (Stok Terpotong)
            </Badge>
            <Badge
              v-else-if="(row as any).status === 'REJECTED'"
              class="bg-rose-50 text-rose-700 border border-rose-200 font-bold text-xs"
            >
              <XCircle class="mr-1 h-3.5 w-3.5" /> Ditolak
            </Badge>
            <Badge
              v-else
              class="bg-amber-50 text-amber-700 border border-amber-200 font-bold text-xs"
            >
              <Clock class="mr-1 h-3.5 w-3.5" /> Menunggu Gudang
            </Badge>

            <!-- Approver info -->
            <div v-if="(row as any).status === 'APPROVED'" class="text-[10px] text-slate-400">
              Oleh: {{ (row as any).approved_by_name || 'Gudang' }} • {{ formatDate((row as any).approved_at) }}
            </div>
            <div v-if="(row as any).status === 'REJECTED'" class="text-[10px] text-rose-500">
              Alasan: {{ (row as any).rejection_reason || '-' }}
            </div>
          </div>
        </template>

        <template #cell-actions="{ row }">
          <div v-if="(row as any).status === 'PENDING'" class="flex items-center gap-1.5">
            <template v-if="authStore.user?.role !== 'admin_nota'">
              <Button
                size="sm"
                @click="openApproveModal(row)"
                class="bg-emerald-600 hover:bg-emerald-700 text-white font-bold text-xs rounded-xl shadow-sm h-8 px-3"
              >
                <Check class="mr-1 h-3.5 w-3.5" /> ACC
              </Button>
              <Button
                size="sm"
                variant="outline"
                @click="openRejectModal(row)"
                class="border-rose-200 text-rose-600 hover:bg-rose-50 font-bold text-xs rounded-xl h-8 px-3"
              >
                <X class="mr-1 h-3.5 w-3.5" /> Tolak
              </Button>
            </template>
            <span v-else class="text-xs text-amber-600 font-bold italic">
              Menunggu ACC Gudang
            </span>
          </div>
          <div v-else class="text-xs text-slate-400 italic">
            Selesai
          </div>
        </template>
      </DataTable>
    </Card>

    <!-- Request Sample Modal -->
    <div
      v-if="showRequestModal"
      class="fixed inset-0 bg-slate-900/60 backdrop-blur-sm z-50 flex items-center justify-center p-4"
    >
      <Card class="w-full max-w-lg shadow-2xl rounded-3xl overflow-hidden border-none animate-in zoom-in-95 duration-200">
        <CardHeader class="p-6 border-b bg-white">
          <div class="flex items-center gap-3">
            <div class="w-10 h-10 rounded-2xl bg-indigo-50 text-indigo-600 flex items-center justify-center">
              <FlaskConical class="h-5 w-5" />
            </div>
            <div>
              <CardTitle class="text-xl font-bold">Permohonan Sampel Lab QC</CardTitle>
              <CardDescription>Pengeluaran barang sampel untuk pengujian mutu / arsip retensi lab QC.</CardDescription>
            </div>
          </div>
        </CardHeader>
        <CardContent class="p-6 space-y-4 bg-white">
          <!-- Pabrik -->
          <div class="space-y-1.5">
            <label class="text-xs font-bold text-slate-700">Unit Pabrik <span class="text-rose-500">*</span></label>
            <select
              v-model="requestForm.factory_id"
              @change="onFactoryChange"
              class="w-full px-4 py-2.5 bg-slate-50 border border-slate-200 rounded-xl text-sm font-bold text-slate-800 outline-none focus:ring-2 focus:ring-indigo-500/20"
            >
              <option value="" disabled>Pilih Unit Pabrik...</option>
              <option v-for="f in factoryStore.factories" :key="f.id" :value="f.id">
                {{ f.name }} ({{ f.location }})
              </option>
            </select>
          </div>

          <!-- Produk -->
          <div class="space-y-1.5">
            <label class="text-xs font-bold text-slate-700">Pilih Produk Jadi <span class="text-rose-500">*</span></label>
            <SearchSelect
              v-model="requestForm.product_id"
              :options="productOptions"
              :disabled="!requestForm.factory_id"
              placeholder="Pilih Produk Jadi..."
              search-placeholder="Ketik nama atau SKU produk..."
              empty-message="Produk tidak ditemukan"
              @change="onProductChange"
              size="sm"
            />
          </div>

          <!-- Batch Produksi -->
          <div class="space-y-1.5">
            <label class="text-xs font-bold text-slate-700">Batch Produksi Terkait <span class="text-rose-500">*</span></label>
            <select
              v-model="requestForm.batch_no"
              @change="onBatchChange"
              :disabled="!requestForm.product_id"
              class="w-full px-4 py-2.5 bg-slate-50 border border-slate-200 rounded-xl text-sm font-bold text-slate-800 outline-none focus:ring-2 focus:ring-indigo-500/20 disabled:opacity-50"
            >
              <option value="" disabled>Pilih Batch Produksi...</option>
              <option v-for="b in availableBatches" :key="b.batch_no" :value="b.batch_no">
                Batch: {{ b.batch_no }} (Sisa Stok: {{ b.quantity }} {{ selectedProduct?.unit || 'PCS' }} | Exp: {{ b.expiry_date ? b.expiry_date.substring(0,10) : 'No Exp' }})
              </option>
            </select>
          </div>

          <!-- Batch info card -->
          <div v-if="selectedBatchStock" class="p-3 bg-amber-50 rounded-2xl border border-amber-200 text-xs space-y-1">
            <p class="font-bold text-amber-900 flex items-center gap-1.5">
              <Layers class="h-4 w-4 text-amber-600" />
              <span>Traceability Batch: {{ selectedBatchStock.batch_no }}</span>
            </p>
            <p class="text-amber-800">
              Stok Gudang Tersedia: <span class="font-black font-mono">{{ selectedBatchStock.quantity }} {{ selectedProduct?.unit || 'PCS' }}</span>
            </p>
            <p v-if="selectedBatchStock.expiry_date" class="text-amber-700 text-[11px]">
              Tgl Kadaluarsa: {{ selectedBatchStock.expiry_date.substring(0,10) }}
            </p>
          </div>

          <!-- Quantity & Tipe Sampel -->
          <div class="grid grid-cols-2 gap-4">
            <div class="space-y-1.5">
              <label class="text-xs font-bold text-slate-700">Jumlah Sampel <span class="text-rose-500">*</span></label>
              <div class="flex items-center gap-2">
                <input
                  type="number"
                  v-model.number="requestForm.quantity"
                  min="1"
                  :max="selectedBatchStock?.quantity || 9999"
                  class="w-full px-4 py-2.5 bg-slate-50 border border-slate-200 rounded-xl text-sm font-bold font-mono outline-none focus:ring-2 focus:ring-indigo-500/20"
                />
                <span class="text-xs font-bold text-slate-500 uppercase">{{ selectedProduct?.unit || 'PCS' }}</span>
              </div>
            </div>

            <div class="space-y-1.5">
              <label class="text-xs font-bold text-slate-700">Jenis Sampel</label>
              <select
                v-model="requestForm.sample_type"
                class="w-full px-4 py-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-bold text-slate-800 outline-none focus:ring-2 focus:ring-indigo-500/20"
              >
                <option value="LAB_QC">Uji Mutu Lab QC</option>
                <option value="RETAINED_SAMPLE">Sampel Retensi QC (Simpan)</option>
                <option value="STABILITY_TEST">Uji Stabilitas / Ketahanan</option>
              </select>
            </div>
          </div>

          <!-- Tujuan & Catatan -->
          <div class="space-y-1.5">
            <label class="text-xs font-bold text-slate-700">Tujuan & Keterangan Pengujian</label>
            <textarea
              v-model="requestForm.purpose"
              rows="2"
              placeholder="Contoh: Uji organoleptik, mikrobiologi, dan sampel arsip produksi..."
              class="w-full px-4 py-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs outline-none focus:ring-2 focus:ring-indigo-500/20 resize-none"
            ></textarea>
          </div>
        </CardContent>
        <CardFooter class="bg-slate-50/60 p-6 flex justify-end gap-3 border-t border-slate-100">
          <Button variant="ghost" @click="showRequestModal = false" class="rounded-xl font-bold">
            Batal
          </Button>
          <Button
            @click="handleSubmitRequest"
            :disabled="submitting || !isFormValid"
            class="rounded-xl bg-indigo-600 hover:bg-indigo-700 text-white font-bold px-6 shadow-lg shadow-indigo-200"
          >
            <Loader2 v-if="submitting" class="mr-2 h-4 w-4 animate-spin" />
            Kirim Permohonan
          </Button>
        </CardFooter>
      </Card>
    </div>

    <!-- Approve Modal (Admin Gudang Confirmation) -->
    <div
      v-if="showApproveModal && selectedSample"
      class="fixed inset-0 bg-slate-900/60 backdrop-blur-sm z-50 flex items-center justify-center p-4"
    >
      <Card class="w-full max-w-md shadow-2xl rounded-3xl overflow-hidden border-none animate-in zoom-in-95 duration-200">
        <CardHeader class="p-6 border-b bg-emerald-50/50">
          <div class="flex items-center gap-3">
            <div class="w-10 h-10 rounded-2xl bg-emerald-100 text-emerald-700 flex items-center justify-center">
              <CheckCircle2 class="h-5 w-5" />
            </div>
            <div>
              <CardTitle class="text-lg font-bold text-emerald-950">ACC Pengeluaran Sampel QC</CardTitle>
              <CardDescription class="text-emerald-700">Persetujuan fisik & pengurangan stok gudang pabrik.</CardDescription>
            </div>
          </div>
        </CardHeader>
        <CardContent class="p-6 space-y-4 bg-white text-xs">
          <div class="p-4 bg-slate-50 rounded-2xl border border-slate-200 space-y-2">
            <div class="flex justify-between">
              <span class="text-slate-400 font-bold">Produk:</span>
              <span class="font-bold text-slate-900">{{ selectedSample.product?.name }}</span>
            </div>
            <div class="flex justify-between">
              <span class="text-slate-400 font-bold">Nomor Batch:</span>
              <span class="font-bold font-mono text-amber-800 bg-amber-50 px-2 py-0.5 rounded border border-amber-200">{{ selectedSample.batch_no }}</span>
            </div>
            <div class="flex justify-between">
              <span class="text-slate-400 font-bold">Jumlah Sampel:</span>
              <span class="font-black text-slate-900 text-sm">{{ selectedSample.quantity }} {{ selectedSample.product?.unit || 'PCS' }}</span>
            </div>
            <div class="flex justify-between">
              <span class="text-slate-400 font-bold">Pemohon:</span>
              <span class="font-bold text-slate-700">{{ selectedSample.requested_by_name }}</span>
            </div>
          </div>

          <div class="p-3 bg-amber-50 rounded-2xl border border-amber-200 flex items-start gap-2.5">
            <AlertTriangle class="h-4 w-4 text-amber-600 shrink-0 mt-0.5" />
            <p class="text-[11px] text-amber-900 leading-relaxed">
              Dengan mengklik <strong>Setujui & Potong Stok</strong>, stok fisik produk pada batch <strong>{{ selectedSample.batch_no }}</strong> di Gudang Pabrik akan berkurang otomatis sebanyak <strong>{{ selectedSample.quantity }} {{ selectedSample.product?.unit || 'PCS' }}</strong> dan dicatat pada log mutasi QC.
            </p>
          </div>
        </CardContent>
        <CardFooter class="bg-slate-50/60 p-6 flex justify-end gap-3 border-t border-slate-100">
          <Button variant="ghost" @click="showApproveModal = false" class="rounded-xl font-bold">
            Batal
          </Button>
          <Button
            @click="handleConfirmApprove"
            :disabled="submitting"
            class="rounded-xl bg-emerald-600 hover:bg-emerald-700 text-white font-bold px-6 shadow-lg shadow-emerald-200"
          >
            <Loader2 v-if="submitting" class="mr-2 h-4 w-4 animate-spin" />
            Setujui & Potong Stok
          </Button>
        </CardFooter>
      </Card>
    </div>

    <!-- Reject Modal -->
    <div
      v-if="showRejectModal && selectedSample"
      class="fixed inset-0 bg-slate-900/60 backdrop-blur-sm z-50 flex items-center justify-center p-4"
    >
      <Card class="w-full max-w-md shadow-2xl rounded-3xl overflow-hidden border-none animate-in zoom-in-95 duration-200">
        <CardHeader class="p-6 border-b bg-rose-50/50">
          <div class="flex items-center gap-3">
            <div class="w-10 h-10 rounded-2xl bg-rose-100 text-rose-700 flex items-center justify-center">
              <XCircle class="h-5 w-5" />
            </div>
            <div>
              <CardTitle class="text-lg font-bold text-rose-950">Tolak Permohonan Sampel</CardTitle>
              <CardDescription class="text-rose-700">Berikan alasan penolakan permohonan ini.</CardDescription>
            </div>
          </div>
        </CardHeader>
        <CardContent class="p-6 space-y-4 bg-white text-xs">
          <div class="space-y-1.5">
            <label class="font-bold text-slate-700">Alasan Penolakan <span class="text-rose-500">*</span></label>
            <textarea
              v-model="rejectionReason"
              rows="3"
              placeholder="Contoh: Nomor batch tidak sesuai, fisik belum diterima QC, atau stok sudah habis..."
              class="w-full px-4 py-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs outline-none focus:ring-2 focus:ring-rose-500/20 resize-none"
            ></textarea>
          </div>
        </CardContent>
        <CardFooter class="bg-slate-50/60 p-6 flex justify-end gap-3 border-t border-slate-100">
          <Button variant="ghost" @click="showRejectModal = false" class="rounded-xl font-bold">
            Batal
          </Button>
          <Button
            @click="handleConfirmReject"
            :disabled="submitting || !rejectionReason.trim()"
            class="rounded-xl bg-rose-600 hover:bg-rose-700 text-white font-bold px-6 shadow-lg shadow-rose-200"
          >
            <Loader2 v-if="submitting" class="mr-2 h-4 w-4 animate-spin" />
            Tolak Permohonan
          </Button>
        </CardFooter>
      </Card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { useFactoryStore, type FactoryQcSample } from '@/stores/factory'
import { Card, CardHeader, CardTitle, CardDescription, CardContent, CardFooter } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import DataTable from '@/components/DataTable.vue'
import SearchSelect, { type SelectOption } from '@/components/SearchSelect.vue'
import {
  FlaskConical,
  Plus,
  RefreshCw,
  Search,
  Package,
  Layers,
  CheckCircle2,
  XCircle,
  Clock,
  Check,
  X,
  AlertTriangle,
  Loader2,
} from 'lucide-vue-next'

const authStore = useAuthStore()
const factoryStore = useFactoryStore()

const loading = ref(false)
const submitting = ref(false)
const searchQuery = ref('')
const selectedFactoryId = ref('')
const activeStatusFilter = ref('ALL')

// Modals state
const showRequestModal = ref(false)
const showApproveModal = ref(false)
const showRejectModal = ref(false)
const selectedSample = ref<FactoryQcSample | null>(null)
const rejectionReason = ref('')

const requestForm = ref({
  factory_id: '',
  product_id: '',
  batch_no: '',
  quantity: 1,
  sample_type: 'LAB_QC',
  purpose: '',
})

const columns = [
  { accessorKey: 'request_info', header: 'Pemohon & Tanggal' },
  { accessorKey: 'product', header: 'Produk' },
  { accessorKey: 'batch', header: 'Batch Produksi' },
  { accessorKey: 'quantity', header: 'Qty Sampel' },
  { accessorKey: 'purpose', header: 'Tujuan Pengujian' },
  { accessorKey: 'status', header: 'Status & Approval' },
  { accessorKey: 'actions', header: 'Aksi' },
]

const pendingCount = computed(() =>
  factoryStore.qcSamples.filter((s) => s.status === 'PENDING').length
)
const approvedCount = computed(() =>
  factoryStore.qcSamples.filter((s) => s.status === 'APPROVED').length
)
const rejectedCount = computed(() =>
  factoryStore.qcSamples.filter((s) => s.status === 'REJECTED').length
)

const statusTabs = computed(() => [
  { label: 'Semua Status', value: 'ALL', count: factoryStore.qcSamples.length },
  { label: 'Menunggu Gudang', value: 'PENDING', count: pendingCount.value },
  { label: 'Disetujui', value: 'APPROVED', count: approvedCount.value },
  { label: 'Ditolak', value: 'REJECTED', count: rejectedCount.value },
])

const filteredSamples = computed(() => {
  return factoryStore.qcSamples.filter((sample) => {
    // Status filter
    if (activeStatusFilter.value !== 'ALL' && sample.status !== activeStatusFilter.value) {
      return false
    }
    // Search query filter
    if (searchQuery.value.trim()) {
      const q = searchQuery.value.toLowerCase()
      const matchProduct = sample.product?.name?.toLowerCase().includes(q) || sample.product?.sku?.toLowerCase().includes(q)
      const matchBatch = sample.batch_no?.toLowerCase().includes(q)
      const matchRequester = sample.requested_by_name?.toLowerCase().includes(q)
      const matchPurpose = sample.purpose?.toLowerCase().includes(q)
      if (!matchProduct && !matchBatch && !matchRequester && !matchPurpose) {
        return false
      }
    }
    return true
  })
})

const availableProducts = computed(() => {
  return factoryStore.sellableProducts || factoryStore.products || []
})

const productOptions = computed<SelectOption[]>(() => {
  return availableProducts.value.map((p) => {
    const totalStock = factoryStore.inventory
      .filter((s) => s.product_id === p.id && s.quantity > 0)
      .reduce((sum, s) => sum + s.quantity, 0)

    return {
      value: p.id,
      label: p.name,
      sublabel: p.sku ? `SKU: ${p.sku} • Satuan: ${p.unit || 'PCS'}` : undefined,
      badge: totalStock > 0 ? `Stok: ${totalStock} ${p.unit || 'PCS'}` : 'Stok: 0',
      badgeColor:
        totalStock > 0
          ? 'bg-emerald-50 text-emerald-700 border border-emerald-200'
          : 'bg-slate-100 text-slate-500',
    }
  })
})

const availableBatches = computed(() => {
  if (!requestForm.value.factory_id || !requestForm.value.product_id) return []
  return factoryStore.inventory.filter(
    (s) => s.product_id === requestForm.value.product_id && s.quantity > 0
  )
})

const selectedProduct = computed(() => {
  return factoryStore.products.find((p) => p.id === requestForm.value.product_id)
})

const selectedBatchStock = computed(() => {
  return availableBatches.value.find((b) => b.batch_no === requestForm.value.batch_no)
})

const isFormValid = computed(() => {
  return (
    requestForm.value.factory_id &&
    requestForm.value.product_id &&
    requestForm.value.batch_no &&
    requestForm.value.quantity > 0
  )
})

const onFactoryChange = async () => {
  if (requestForm.value.factory_id) {
    await factoryStore.fetchInventory(requestForm.value.factory_id)
  }
  requestForm.value.product_id = ''
  requestForm.value.batch_no = ''
}

const onProductChange = () => {
  requestForm.value.batch_no = ''
}

const onBatchChange = () => {
  // If batch selected, set default max/qty
}

const loadData = async () => {
  loading.value = true
  try {
    if (selectedFactoryId.value) {
      await factoryStore.fetchQcSamples(selectedFactoryId.value)
      await factoryStore.fetchInventory(selectedFactoryId.value)
    } else {
      await factoryStore.fetchAllQcSamples()
      await factoryStore.fetchAllInventory()
    }
  } finally {
    loading.value = false
  }
}

const openRequestModal = async () => {
  requestForm.value = {
    factory_id: factoryStore.factories[0]?.id || '',
    product_id: '',
    batch_no: '',
    quantity: 1,
    sample_type: 'LAB_QC',
    purpose: '',
  }
  if (requestForm.value.factory_id) {
    await factoryStore.fetchInventory(requestForm.value.factory_id)
  }
  showRequestModal.value = true
}

const handleSubmitRequest = async () => {
  if (!isFormValid.value) return
  submitting.value = true
  try {
    await factoryStore.requestQcSample(requestForm.value.factory_id, requestForm.value)
    showRequestModal.value = false
    await loadData()
  } finally {
    submitting.value = false
  }
}

const openApproveModal = (sample: FactoryQcSample) => {
  selectedSample.value = sample
  showApproveModal.value = true
}

const handleConfirmApprove = async () => {
  if (!selectedSample.value) return
  submitting.value = true
  try {
    await factoryStore.approveQcSample(selectedSample.value.id, selectedSample.value.factory_id)
    showApproveModal.value = false
    selectedSample.value = null
    await loadData()
  } finally {
    submitting.value = false
  }
}

const openRejectModal = (sample: FactoryQcSample) => {
  selectedSample.value = sample
  rejectionReason.value = ''
  showRejectModal.value = true
}

const handleConfirmReject = async () => {
  if (!selectedSample.value || !rejectionReason.value.trim()) return
  submitting.value = true
  try {
    await factoryStore.rejectQcSample(
      selectedSample.value.id,
      rejectionReason.value.trim(),
      selectedSample.value.factory_id
    )
    showRejectModal.value = false
    selectedSample.value = null
    await loadData()
  } finally {
    submitting.value = false
  }
}

const formatDate = (dateStr?: string) => {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleString('id-ID', {
    day: '2-digit',
    month: 'short',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}

const formatSimpleDate = (dateStr?: string) => {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleDateString('id-ID', {
    day: '2-digit',
    month: 'short',
    year: 'numeric',
  })
}

onMounted(async () => {
  await Promise.all([
    factoryStore.fetchFactories(),
    factoryStore.fetchProducts(),
    loadData(),
  ])
})
</script>
