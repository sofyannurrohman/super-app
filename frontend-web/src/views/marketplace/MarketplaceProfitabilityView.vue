<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import apiClient from '@/api/axios'
import { useMasterDataStore } from '@/stores/masterData'
import { toast } from 'vue-sonner'
import { BarChart3, Plus, RefreshCw, Trash2 } from 'lucide-vue-next'

type Row = {
  branch_id: string; store: string; marketplace: string; seller_sku: string; product_name: string
  units: number; orders: number; gross_revenue: number; refunds: number; net_sales: number
  marketplace_fees: number; cogs: number; missing_cogs_units: number; ads: number
  fulfillment: number; returns: number; operational: number; attributed_revenue: number
  contribution: number; margin_percent: number; roas: number | null
}
type Expense = { id: number; branch_id: string; marketplace: string; expense_date: string; category: string; amount: number; attributed_revenue: number; seller_sku: string; notes: string }
type Report = { stores: Row[]; products: Row[]; expenses: Expense[]; method: string }

const master = useMasterDataStore()
const today = () => { const d = new Date(); return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}` }
const monthStart = () => `${today().slice(0, 7)}-01`
const from = ref(monthStart())
const to = ref(today())
const marketplace = ref('')
const selectedStore = ref('')
const loading = ref(false)
const saving = ref(false)
const report = ref<Report>({ stores: [], products: [], expenses: [], method: '' })
const previous = ref<Report>({ stores: [], products: [], expenses: [], method: '' })
const lowMarginLimit = ref(10)
const form = ref({ branch_id: '', marketplace: 'tiktok', expense_date: today(), category: 'ads', amount: 0, attributed_revenue: 0, seller_sku: '', notes: '' })
const money = (v: number) => new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR', maximumFractionDigits: 0 }).format(v || 0)
const number = (v: number) => new Intl.NumberFormat('id-ID').format(v || 0)
const pct = (v: number) => `${new Intl.NumberFormat('id-ID', { maximumFractionDigits: 1 }).format(v || 0)}%`
const branchID = computed(() => master.selectedBranchId === 'ALL_BRANCHES' ? '' : master.selectedBranchId)
const stores = computed(() => report.value.stores.filter(r => !selectedStore.value || r.branch_id === selectedStore.value))
const products = computed(() => report.value.products.filter(r => !selectedStore.value || r.branch_id === selectedStore.value).sort((a, b) => b.units - a.units))
const summary = computed(() => stores.value.reduce((a, r) => {
  a.gross += r.gross_revenue; a.net += r.net_sales; a.cogs += r.cogs; a.fees += r.marketplace_fees
  a.ads += r.ads; a.fulfillment += r.fulfillment; a.returns += r.returns; a.operational += r.operational
  a.contribution += r.contribution; a.missing += r.missing_cogs_units
  return a
}, { gross: 0, net: 0, cogs: 0, fees: 0, ads: 0, fulfillment: 0, returns: 0, operational: 0, contribution: 0, missing: 0 }))
const previousContribution = computed(() => previous.value.stores.filter(r => !selectedStore.value || r.branch_id === selectedStore.value).reduce((sum, r) => sum + r.contribution, 0))
const lowMargin = computed(() => products.value.filter(r => r.units > 0 && r.net_sales > 0 && r.missing_cogs_units === 0 && r.margin_percent < lowMarginLimit.value))
const expenseBranches = computed(() => branchID.value ? master.branches.filter(b => b.id === branchID.value) : master.branches)

async function load() {
  if (!from.value || !to.value || to.value < from.value) return
  loading.value = true
  try {
    const start = new Date(`${from.value}T00:00:00`)
    const end = new Date(`${to.value}T00:00:00`)
    const days = Math.round((end.getTime() - start.getTime()) / 86400000) + 1
    const priorTo = new Date(start); priorTo.setDate(priorTo.getDate() - 1)
    const priorFrom = new Date(priorTo); priorFrom.setDate(priorFrom.getDate() - days + 1)
    const dateString = (d: Date) => `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
    const common = { branch_id: branchID.value || undefined, marketplace: marketplace.value || undefined }
    const [currentResponse, previousResponse] = await Promise.all([
      apiClient.get<Report>('/marketplace/profitability', { params: { ...common, from: from.value, to: to.value } }),
      apiClient.get<Report>('/marketplace/profitability', { params: { ...common, from: dateString(priorFrom), to: dateString(priorTo) } }),
    ])
    report.value = currentResponse.data
    previous.value = previousResponse.data
  } catch { toast.error('Gagal memuat analisis profit') }
  finally { loading.value = false }
}
async function saveExpense() {
  if (!form.value.branch_id || form.value.amount <= 0) { toast.error('Pilih toko dan isi biaya lebih dari nol'); return }
  saving.value = true
  try {
    await apiClient.post('/marketplace/profitability/expenses', form.value)
    toast.success('Biaya tersimpan')
    form.value.amount = 0; form.value.attributed_revenue = 0; form.value.seller_sku = ''; form.value.notes = ''
    await load()
  } catch { toast.error('Gagal menyimpan biaya') }
  finally { saving.value = false }
}
async function deleteExpense(id: number) {
  try { await apiClient.delete(`/marketplace/profitability/expenses/${id}`, { params: { branch_id: branchID.value || undefined } }); await load(); toast.success('Biaya dihapus') }
  catch { toast.error('Gagal menghapus biaya') }
}
watch([from, to, marketplace, branchID], load)
watch(expenseBranches, branches => { if (!branches.some(b => b.id === form.value.branch_id)) form.value.branch_id = branches[0]?.id || '' }, { immediate: true })
onMounted(async () => { if (!master.branches.length) await master.fetchBranches(); await load() })
</script>

<template>
  <div class="space-y-6 p-4 md:p-6">
    <div class="flex flex-wrap items-start justify-between gap-3">
      <div><h1 class="flex items-center gap-2 text-2xl font-bold text-slate-900"><BarChart3 class="h-6 w-6" /> Profitability Analytics</h1><p class="mt-1 text-sm text-slate-500">Omzet, biaya, dan margin kontribusi marketplace</p></div>
      <button class="flex items-center gap-2 rounded-lg border px-3 py-2 text-sm hover:bg-slate-50" :disabled="loading" @click="load"><RefreshCw class="h-4 w-4" :class="{ 'animate-spin': loading }" /> Perbarui</button>
    </div>

    <div class="grid gap-3 rounded-xl border bg-white p-4 sm:grid-cols-2 lg:grid-cols-4">
      <label class="text-sm">Dari <input v-model="from" type="date" class="mt-1 w-full rounded-lg border px-3 py-2" /></label>
      <label class="text-sm">Sampai <input v-model="to" type="date" class="mt-1 w-full rounded-lg border px-3 py-2" /></label>
      <label class="text-sm">Marketplace <select v-model="marketplace" class="mt-1 w-full rounded-lg border px-3 py-2"><option value="">Semua</option><option value="tiktok">TikTok Shop</option><option value="shopee">Shopee</option></select></label>
      <label class="text-sm">Toko / cabang <select v-model="selectedStore" class="mt-1 w-full rounded-lg border px-3 py-2"><option value="">Semua</option><option v-for="b in expenseBranches" :key="b.id" :value="b.id">{{ b.name }}</option></select></label>
    </div>

    <div class="grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
      <div v-for="card in [{ label: 'Omzet', value: money(summary.gross) }, { label: 'Penjualan bersih', value: money(summary.net) }, { label: 'HPP terpetakan', value: money(summary.cogs) }, { label: 'Margin kontribusi', value: money(summary.contribution) }]" :key="card.label" class="rounded-xl border bg-white p-4"><p class="text-sm text-slate-500">{{ card.label }}</p><p class="mt-2 text-xl font-semibold">{{ card.value }}</p></div>
    </div>
    <div class="rounded-xl border bg-white p-4 text-sm">
      <div class="grid gap-2 sm:grid-cols-3 lg:grid-cols-6"><div><span class="text-slate-500">Biaya marketplace</span><p class="font-semibold">{{ money(summary.fees) }}</p></div><div><span class="text-slate-500">Iklan</span><p class="font-semibold">{{ money(summary.ads) }}</p></div><div><span class="text-slate-500">Fulfillment</span><p class="font-semibold">{{ money(summary.fulfillment) }}</p></div><div><span class="text-slate-500">Retur</span><p class="font-semibold">{{ money(summary.returns) }}</p></div><div><span class="text-slate-500">Operasional</span><p class="font-semibold">{{ money(summary.operational) }}</p></div><div><span class="text-slate-500">Periode sebelumnya</span><p class="font-semibold">{{ money(previousContribution) }}</p></div></div>
      <p class="mt-3 text-slate-500">Periode pembanding memiliki jumlah hari yang sama tepat sebelum periode terpilih. Margin = penjualan bersih − biaya marketplace − HPP − biaya tambahan.</p>
      <p v-if="summary.missing" class="mt-2 rounded-lg bg-amber-50 p-2 text-amber-800">{{ number(summary.missing) }} unit belum memiliki HPP. Margin yang ditampilkan masih terlalu tinggi sampai HPP dilengkapi.</p>
      <p class="mt-2 text-slate-500">Nilai order dan potongan dialokasikan ke SKU menurut jumlah unit karena harga dan potongan per item belum tersimpan. Biaya tanpa SKU hanya masuk ke margin toko. Toko mengikuti cabang akun yang terhubung. Top-up saldo iklan Shopee dikeluarkan dari biaya marketplace; catat belanja iklan aktual secara terpisah.</p>
    </div>

    <section class="rounded-xl border bg-white p-4"><h2 class="mb-3 text-lg font-semibold">Performa toko dan marketplace</h2><div class="overflow-x-auto"><table class="w-full min-w-[900px] text-left text-sm"><thead class="border-b bg-slate-50 text-slate-600"><tr><th class="p-2">Toko</th><th class="p-2">Marketplace</th><th class="p-2 text-right">Order</th><th class="p-2 text-right">Omzet</th><th class="p-2 text-right">Bersih</th><th class="p-2 text-right">HPP</th><th class="p-2 text-right">Biaya total</th><th class="p-2 text-right">Kontribusi</th><th class="p-2 text-right">Margin</th><th class="p-2 text-right">ROAS iklan</th></tr></thead><tbody><tr v-for="r in stores" :key="`${r.branch_id}-${r.marketplace}`" class="border-b"><td class="p-2">{{ r.store }}</td><td class="p-2 capitalize">{{ r.marketplace }}</td><td class="p-2 text-right">{{ number(r.orders) }}</td><td class="p-2 text-right">{{ money(r.gross_revenue) }}</td><td class="p-2 text-right">{{ money(r.net_sales) }}</td><td class="p-2 text-right">{{ money(r.cogs) }}</td><td class="p-2 text-right">{{ money(r.marketplace_fees + r.ads + r.fulfillment + r.returns + r.operational) }}</td><td class="p-2 text-right font-semibold" :class="r.contribution < 0 ? 'text-red-600' : 'text-emerald-700'">{{ money(r.contribution) }}</td><td class="p-2 text-right">{{ pct(r.margin_percent) }}</td><td class="p-2 text-right">{{ r.roas == null ? '—' : `${r.roas.toFixed(2)}x` }}</td></tr><tr v-if="!stores.length"><td colspan="10" class="p-6 text-center text-slate-500">Belum ada data pada periode ini.</td></tr></tbody></table></div></section>

    <section class="rounded-xl border bg-white p-4"><div class="mb-3 flex flex-wrap items-center justify-between gap-2"><h2 class="text-lg font-semibold">Produk laris dengan margin rendah</h2><label class="text-sm">Batas margin <input v-model.number="lowMarginLimit" type="number" min="-100" max="100" class="ml-2 w-20 rounded border px-2 py-1" /> %</label></div><p class="mb-3 text-sm text-slate-500">Diurutkan menurut unit terjual. Produk tanpa HPP dipisahkan dari peringatan margin.</p><div v-if="lowMargin.length" class="overflow-x-auto"><table class="w-full min-w-[650px] text-sm"><thead class="border-b text-left"><tr><th class="p-2">SKU / produk</th><th class="p-2">Toko</th><th class="p-2 text-right">Unit</th><th class="p-2 text-right">Penjualan bersih</th><th class="p-2 text-right">Kontribusi</th><th class="p-2 text-right">Margin</th></tr></thead><tbody><tr v-for="r in lowMargin" :key="`${r.branch_id}-${r.marketplace}-${r.seller_sku}`" class="border-b"><td class="p-2">{{ r.seller_sku }}<span class="block text-xs text-slate-500">{{ r.product_name }}</span></td><td class="p-2">{{ r.store }} · {{ r.marketplace }}</td><td class="p-2 text-right">{{ number(r.units) }}</td><td class="p-2 text-right">{{ money(r.net_sales) }}</td><td class="p-2 text-right">{{ money(r.contribution) }}</td><td class="p-2 text-right text-amber-700">{{ pct(r.margin_percent) }}</td></tr></tbody></table></div><p v-else class="text-sm text-slate-500">Tidak ada produk yang melewati batas margin pada periode ini.</p></section>

    <section class="rounded-xl border bg-white p-4"><h2 class="mb-3 text-lg font-semibold">Rincian per SKU</h2><div class="max-h-[520px] overflow-auto"><table class="w-full min-w-[850px] text-sm"><thead class="sticky top-0 border-b bg-slate-50 text-left"><tr><th class="p-2">SKU / produk</th><th class="p-2">Toko</th><th class="p-2 text-right">Unit</th><th class="p-2 text-right">Omzet</th><th class="p-2 text-right">Bersih</th><th class="p-2 text-right">HPP</th><th class="p-2 text-right">Kontribusi</th><th class="p-2 text-right">Margin</th></tr></thead><tbody><tr v-for="r in products" :key="`${r.branch_id}-${r.marketplace}-${r.seller_sku}`" class="border-b"><td class="p-2">{{ r.seller_sku }}<span class="block text-xs text-slate-500">{{ r.product_name }}</span></td><td class="p-2">{{ r.store }} · {{ r.marketplace }}</td><td class="p-2 text-right">{{ number(r.units) }}</td><td class="p-2 text-right">{{ money(r.gross_revenue) }}</td><td class="p-2 text-right">{{ money(r.net_sales) }}</td><td class="p-2 text-right">{{ money(r.cogs) }}<span v-if="r.missing_cogs_units" class="block text-xs text-amber-700">{{ r.missing_cogs_units }} unit tanpa HPP</span></td><td class="p-2 text-right">{{ money(r.contribution) }}</td><td class="p-2 text-right">{{ pct(r.margin_percent) }}</td></tr><tr v-if="!products.length"><td colspan="8" class="p-6 text-center text-slate-500">Belum ada data SKU.</td></tr></tbody></table></div></section>

    <section class="rounded-xl border bg-white p-4"><h2 class="mb-1 text-lg font-semibold">Catat biaya tambahan dan pendapatan iklan</h2><p class="mb-4 text-sm text-slate-500">Isi pendapatan yang diatribusikan iklan untuk menghitung ROAS. Jangan masukkan kembali potongan yang sudah tercatat sebagai biaya marketplace.</p><form class="grid gap-3 sm:grid-cols-2 lg:grid-cols-4" @submit.prevent="saveExpense"><label class="text-sm">Toko / cabang <select v-model="form.branch_id" required class="mt-1 w-full rounded border px-3 py-2"><option value="">Pilih toko</option><option v-for="b in expenseBranches" :key="b.id" :value="b.id">{{ b.name }}</option></select></label><label class="text-sm">Marketplace <select v-model="form.marketplace" class="mt-1 w-full rounded border px-3 py-2"><option value="tiktok">TikTok Shop</option><option value="shopee">Shopee</option></select></label><label class="text-sm">Tanggal <input v-model="form.expense_date" type="date" required class="mt-1 w-full rounded border px-3 py-2" /></label><label class="text-sm">Kategori <select v-model="form.category" class="mt-1 w-full rounded border px-3 py-2"><option value="ads">Iklan</option><option value="fulfillment">Fulfillment</option><option value="return">Retur</option><option value="operational">Operasional</option></select></label><label class="text-sm">Biaya (Rp) <input v-model.number="form.amount" type="number" min="0.01" step="0.01" required class="mt-1 w-full rounded border px-3 py-2" /></label><label v-if="form.category === 'ads'" class="text-sm">Pendapatan dari iklan (Rp) <input v-model.number="form.attributed_revenue" type="number" min="0" step="0.01" class="mt-1 w-full rounded border px-3 py-2" /></label><label class="text-sm">Seller SKU (opsional) <input v-model="form.seller_sku" maxlength="100" class="mt-1 w-full rounded border px-3 py-2" /></label><label class="text-sm">Catatan <input v-model="form.notes" class="mt-1 w-full rounded border px-3 py-2" /></label><button type="submit" :disabled="saving" class="flex items-center justify-center gap-2 rounded-lg bg-slate-900 px-4 py-2 text-white disabled:opacity-50"><Plus class="h-4 w-4" /> Simpan biaya</button></form>
      <div v-if="report.expenses.length" class="mt-5 max-h-64 overflow-auto"><table class="w-full min-w-[650px] text-left text-sm"><thead class="border-b"><tr><th class="p-2">Tanggal</th><th class="p-2">Toko / marketplace</th><th class="p-2">Kategori / SKU</th><th class="p-2 text-right">Biaya</th><th class="p-2 text-right">Pendapatan iklan</th><th class="p-2"></th></tr></thead><tbody><tr v-for="e in report.expenses" :key="e.id" class="border-b"><td class="p-2">{{ e.expense_date.slice(0, 10) }}</td><td class="p-2">{{ master.branches.find(b => b.id === e.branch_id)?.name || e.branch_id }} · {{ e.marketplace }}</td><td class="p-2">{{ e.category }} <span v-if="e.seller_sku">· {{ e.seller_sku }}</span><span v-if="e.notes" class="block text-xs text-slate-500">{{ e.notes }}</span></td><td class="p-2 text-right">{{ money(e.amount) }}</td><td class="p-2 text-right">{{ e.category === 'ads' ? money(e.attributed_revenue) : '—' }}</td><td class="p-2"><button class="text-red-600" aria-label="Hapus biaya" @click="deleteExpense(e.id)"><Trash2 class="h-4 w-4" /></button></td></tr></tbody></table></div>
    </section>
  </div>
</template>
