<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { toast } from 'vue-sonner'
import apiClient from '@/api/axios'
import { useAuthStore } from '@/stores/auth'
import { useMasterDataStore } from '@/stores/masterData'

type AvailableInvoice = { invoice: any; available_amount: number }
const authStore = useAuthStore()
const masterData = useMasterDataStore()
const isCentral = computed(() => ['cashier_pusat', 'superadmin', 'manager'].includes(authStore.user?.role || ''))
const branchId = computed(() => masterData.selectedBranchId && masterData.selectedBranchId !== 'ALL_BRANCHES'
  ? masterData.selectedBranchId : (isCentral.value ? '' : authStore.user?.branch_id || ''))
const companyId = computed(() => masterData.selectedBranchCompanyId || '')
const invoices = ref<AvailableInvoice[]>([])
const deposits = ref<any[]>([])
const banks = ref<any[]>([])
const search = ref('')
const allocations = ref<Record<string, number>>({})
const bankAccountId = ref('')
const depositDate = ref(new Date().toLocaleDateString('en-CA'))
const notes = ref('')
const proofFile = ref<File | null>(null)
const loading = ref(false)
const saving = ref(false)
const reviewingId = ref('')

const selected = computed(() => invoices.value.filter(row => Number(allocations.value[row.invoice.id] || 0) > 0))
const selectedCompanyIds = computed(() => new Set(selected.value.map(row => row.invoice.company_id)))
const eligibleBanks = computed(() => banks.value.filter(bank => !bank.branch_id &&
  (!selectedCompanyIds.value.size || !bank.company_id || selectedCompanyIds.value.has(bank.company_id))))
const selectedTotal = computed(() => selected.value.reduce((sum, row) => sum + Number(allocations.value[row.invoice.id] || 0), 0))
const rupiah = (value: number) => new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR', maximumFractionDigits: 0 }).format(value || 0)
const dateLabel = (value: string) => value ? new Date(value).toLocaleDateString('id-ID') : '-'
const invoiceName = (invoice: any) => invoice.buyer_name || invoice.store?.name || invoice.branch?.name || 'Pelanggan'

const fetchData = async () => {
  loading.value = true
  try {
    const params: any = {}
    if (branchId.value) params.branch_id = branchId.value
    if (companyId.value) params.company_id = companyId.value
    if (search.value.trim()) params.search = search.value.trim()
    const [invoiceRes, depositRes, bankRes] = await Promise.all([
      apiClient.get('/finance/invoice-deposits/available-invoices', { params }),
      apiClient.get('/finance/invoice-deposits', { params: branchId.value ? { branch_id: branchId.value } : {} }),
      apiClient.get('/finance/bank-accounts'),
    ])
    invoices.value = invoiceRes.data?.data || []
    deposits.value = depositRes.data?.data || []
    banks.value = (bankRes.data?.data || []).filter((bank: any) => bank.is_active)
    for (const id of Object.keys(allocations.value)) {
      if (!invoices.value.some(row => row.invoice.id === id)) delete allocations.value[id]
    }
  } catch (error: any) {
    toast.error(error.response?.data?.error || 'Gagal memuat faktur dan setoran')
  } finally {
    loading.value = false
  }
}

const toggleInvoice = (row: AvailableInvoice) => {
  if (allocations.value[row.invoice.id]) {
    delete allocations.value[row.invoice.id]
  } else {
    allocations.value[row.invoice.id] = Number(row.available_amount)
  }
}

const submitDeposit = async () => {
  if (!branchId.value) return toast.error('Pilih cabang terlebih dahulu')
  if (!bankAccountId.value) return toast.error('Pilih rekening pabrik penerima')
  if (!eligibleBanks.value.some(bank => bank.id === bankAccountId.value)) return toast.error('Rekening tidak sesuai dengan PT pada faktur')
  if (!selected.value.length) return toast.error('Pilih faktur yang dibayar')
  if (selectedCompanyIds.value.size !== 1) return toast.error('Pisahkan setoran untuk setiap PT')
  if (selected.value.some(row => Number(allocations.value[row.invoice.id]) > Number(row.available_amount) + 0.009)) {
    return toast.error('Nominal setoran melebihi sisa salah satu faktur')
  }
  if (!proofFile.value) return toast.error('Unggah bukti setoran ke rekening pabrik')
  saving.value = true
  try {
    const formData = new FormData()
    formData.append('file', proofFile.value)
    const upload = await apiClient.post('/cashier-reports/upload', formData, { headers: { 'Content-Type': 'multipart/form-data' } })
    const proofUrl = upload.data?.data?.url || upload.data?.url
    if (!proofUrl) throw new Error('Bukti setoran gagal diunggah')
    await apiClient.post('/finance/invoice-deposits', {
      branch_id: branchId.value,
      bank_account_id: bankAccountId.value,
      deposit_date: depositDate.value,
      proof_url: proofUrl,
      notes: notes.value,
      allocations: selected.value.map(row => ({ invoice_id: row.invoice.id, amount: Number(allocations.value[row.invoice.id]) })),
    })
    toast.success('Setoran faktur dikirim untuk verifikasi Kasir Pusat')
    allocations.value = {}
    bankAccountId.value = ''
    notes.value = ''
    proofFile.value = null
    await fetchData()
  } catch (error: any) {
    toast.error(error.response?.data?.error || error.message || 'Gagal membuat setoran')
  } finally {
    saving.value = false
  }
}

const review = async (deposit: any, approve: boolean) => {
  if (!confirm(`${approve ? 'Terima' : 'Tolak'} setoran ${rupiah(deposit.amount)} untuk ${deposit.items?.length || 0} faktur?`)) return
  reviewingId.value = deposit.id
  try {
    await apiClient.post(`/finance/invoice-deposits/${deposit.id}/${approve ? 'approve' : 'reject'}`)
    toast.success(approve ? 'Setoran diterima dan pembayaran faktur tercatat' : 'Setoran ditolak')
    await fetchData()
  } catch (error: any) {
    toast.error(error.response?.data?.error || 'Gagal memproses setoran')
  } finally {
    reviewingId.value = ''
  }
}

let searchTimer: ReturnType<typeof setTimeout> | undefined
watch([branchId, companyId], () => { allocations.value = {}; fetchData() })
watch(search, () => { clearTimeout(searchTimer); searchTimer = setTimeout(fetchData, 350) })
onMounted(fetchData)
</script>

<template>
  <div class="p-4 sm:p-6 space-y-6">
    <div>
      <h1 class="text-2xl font-black text-slate-900">Penerimaan Faktur & Setoran Pabrik</h1>
      <p class="text-sm text-slate-500 mt-1">Admin Kasir memilih faktur yang dibayar dan rekening tujuan. Kasir Pusat memverifikasi uang masuk sebelum faktur ditandai terbayar.</p>
    </div>

    <div v-if="!isCentral" class="grid grid-cols-1 xl:grid-cols-[1fr_340px] gap-5 items-start">
      <section class="bg-white rounded-2xl border border-slate-200 overflow-hidden">
        <div class="p-4 border-b border-slate-100 flex flex-wrap items-center justify-between gap-3">
          <h2 class="font-bold text-slate-800">Faktur Siap Dibayar</h2>
          <input v-model="search" type="search" placeholder="Cari nomor faktur / pelanggan" class="border border-slate-200 rounded-xl px-3 py-2 text-sm w-full sm:w-64" />
        </div>
        <div v-if="loading" class="p-8 text-center text-slate-500">Memuat faktur...</div>
        <div v-else-if="!invoices.length" class="p-8 text-center text-slate-500">Tidak ada faktur yang siap dibayar.</div>
        <div v-else class="divide-y divide-slate-100 max-h-[620px] overflow-y-auto">
          <div v-for="row in invoices" :key="row.invoice.id" class="p-4 flex gap-3 items-start">
            <input type="checkbox" :checked="!!allocations[row.invoice.id]" @change="toggleInvoice(row)" class="mt-1" />
            <div class="flex-1 min-w-0">
              <div class="flex justify-between gap-3"><span class="font-bold text-sm text-slate-900">{{ row.invoice.receipt_no }}</span><span class="text-xs text-slate-500">{{ dateLabel(row.invoice.transaction_date) }}</span></div>
              <p class="text-xs text-slate-600 mt-1">{{ invoiceName(row.invoice) }} • {{ row.invoice.company?.name || 'PT' }}</p>
              <p class="text-xs text-slate-500 mt-1">Sisa tagihan: <strong class="text-slate-800">{{ rupiah(row.available_amount) }}</strong></p>
              <div v-if="allocations[row.invoice.id]" class="mt-2">
                <label class="block text-xs font-semibold text-slate-600 mb-1">Nominal yang disetor untuk faktur ini</label>
                <input v-model.number="allocations[row.invoice.id]" type="number" min="0" :max="row.available_amount" step="0.01" class="w-full border border-slate-200 rounded-lg px-3 py-2 text-sm" />
              </div>
            </div>
          </div>
        </div>
      </section>
      <section class="bg-white rounded-2xl border border-slate-200 p-5 space-y-4 xl:sticky xl:top-5">
        <h2 class="font-bold text-slate-800">Buat Setoran ke Pabrik</h2>
        <div class="text-sm text-slate-600">{{ selected.length }} faktur dipilih <strong class="block text-xl text-slate-900 mt-1">{{ rupiah(selectedTotal) }}</strong></div>
        <div><label class="block text-xs font-bold text-slate-600 mb-1">Rekening pabrik penerima *</label><select v-model="bankAccountId" class="w-full border border-slate-200 rounded-xl px-3 py-2 text-sm"><option value="">Pilih rekening</option><option v-for="bank in eligibleBanks" :key="bank.id" :value="bank.id">{{ bank.bank_name }} • {{ bank.account_number }} • {{ bank.account_holder }}</option></select></div>
        <div><label class="block text-xs font-bold text-slate-600 mb-1">Tanggal setoran *</label><input v-model="depositDate" type="date" class="w-full border border-slate-200 rounded-xl px-3 py-2 text-sm" /></div>
        <div><label class="block text-xs font-bold text-slate-600 mb-1">Bukti setoran *</label><input type="file" accept="image/*,.pdf" @change="proofFile = ($event.target as HTMLInputElement).files?.[0] || null" class="block w-full text-xs" /></div>
        <div><label class="block text-xs font-bold text-slate-600 mb-1">Catatan</label><textarea v-model="notes" rows="2" class="w-full border border-slate-200 rounded-xl px-3 py-2 text-sm" /></div>
        <button :disabled="saving || !selected.length" @click="submitDeposit" class="w-full bg-emerald-600 hover:bg-emerald-700 disabled:opacity-50 text-white font-bold rounded-xl py-3 text-sm">{{ saving ? 'Menyimpan...' : 'Kirim Setoran untuk Verifikasi' }}</button>
      </section>
    </div>

    <section class="bg-white rounded-2xl border border-slate-200 overflow-hidden">
      <div class="p-4 border-b border-slate-100 flex items-center justify-between"><h2 class="font-bold text-slate-800">Riwayat Setoran Faktur</h2><button @click="fetchData" class="text-sm font-semibold text-emerald-700">Muat ulang</button></div>
      <p v-if="!deposits.length" class="p-8 text-center text-sm text-slate-500">Belum ada setoran faktur.</p>
      <div v-else class="divide-y divide-slate-100">
        <div v-for="deposit in deposits" :key="deposit.id" class="p-4 sm:p-5 space-y-3">
          <div class="flex flex-wrap justify-between gap-3"><div><p class="font-bold text-slate-900">{{ deposit.branch?.name || 'Cabang' }} • {{ rupiah(deposit.amount) }}</p><p class="text-xs text-slate-500 mt-1">{{ dateLabel(deposit.deposit_date) }} • {{ deposit.bank_account?.bank_name }} {{ deposit.bank_account?.account_number }}</p></div><span class="text-xs font-bold px-2.5 py-1 rounded-full h-fit" :class="deposit.status === 'APPROVED' ? 'bg-emerald-100 text-emerald-800' : deposit.status === 'REJECTED' ? 'bg-rose-100 text-rose-800' : 'bg-amber-100 text-amber-800'">{{ deposit.status === 'APPROVED' ? 'Diterima' : deposit.status === 'REJECTED' ? 'Ditolak' : 'Menunggu Kasir Pusat' }}</span></div>
          <div class="flex flex-wrap gap-2"><span v-for="item in deposit.items" :key="item.id" class="bg-slate-50 border border-slate-200 rounded-lg px-2.5 py-1 text-xs text-slate-700">{{ item.sales_transaction?.receipt_no }}: {{ rupiah(item.amount) }}</span></div>
          <div class="flex items-center gap-3"><a v-if="deposit.proof_url" :href="deposit.proof_url" target="_blank" rel="noopener" class="text-xs font-semibold text-indigo-700 underline">Lihat bukti setoran</a><span v-if="deposit.notes" class="text-xs text-slate-500">{{ deposit.notes }}</span></div>
          <div v-if="isCentral && deposit.status === 'PENDING'" class="flex gap-2"><button :disabled="!!reviewingId" @click="review(deposit, true)" class="bg-emerald-600 text-white text-xs font-bold px-4 py-2 rounded-lg disabled:opacity-50">Terima Uang Masuk</button><button :disabled="!!reviewingId" @click="review(deposit, false)" class="border border-rose-200 text-rose-700 text-xs font-bold px-4 py-2 rounded-lg disabled:opacity-50">Tolak</button></div>
        </div>
      </div>
    </section>
  </div>
</template>
