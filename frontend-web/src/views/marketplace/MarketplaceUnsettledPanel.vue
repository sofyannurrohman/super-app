<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { marketplaceApi, type UnsettledSnapshot, type UnsettledTransaction } from '@/api/marketplace'
import { useMasterDataStore } from '@/stores/masterData'
import { toast } from 'vue-sonner'

const masterData = useMasterDataStore()
const shops = ref<string[]>([])
const selectedShop = ref('')
const snapshots = ref<UnsettledSnapshot[]>([])
const transactions = ref<UnsettledTransaction[]>([])
const selectedSnapshot = ref<UnsettledSnapshot | null>(null)
const loading = ref(false)
const capturing = ref(false)
const center = ref({ awaiting_delivery: '', after_sales: '', processing: '' })

const latestAPI = computed(() => snapshots.value.find((s) => s.source === 'TIKTOK_API'))
const latestCenter = computed(() => snapshots.value.find((s) => s.source === 'SELLER_CENTER'))
const categories = [
  { key: 'awaiting_delivery', label: 'Menunggu paket diterima' },
  { key: 'after_sales', label: 'Pengembalian dana/barang diproses' },
  { key: 'processing', label: 'Penyelesaian pembayaran diproses' },
] as const

const money = (value: number | undefined) => new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR', maximumFractionDigits: 0 }).format(value || 0)
const date = (value: string) => new Date(value).toLocaleString('id-ID')
const total = (s?: UnsettledSnapshot) => s ? s.awaiting_delivery + s.after_sales + s.processing + s.unclassified_amount : 0
const branchId = () => {
  const value = masterData.selectedBranchId
  return value && value !== 'ALL_BRANCHES' && !value.includes(',') ? value : undefined
}

async function load() {
  loading.value = true
  try {
    const res = await marketplaceApi.getUnsettledSnapshots(branchId(), selectedShop.value || undefined)
    shops.value = res.data.shops || []
    if (!shops.value.includes(selectedShop.value)) selectedShop.value = shops.value[0] || ''
    snapshots.value = (res.data.snapshots || []).filter((s) => s.shop_id === selectedShop.value)
  } catch (error: any) {
    toast.error(error.response?.data?.error || 'Gagal memuat snapshot TikTok')
  } finally {
    loading.value = false
  }
}

async function selectSnapshot(snapshot: UnsettledSnapshot) {
  try {
    const res = await marketplaceApi.getUnsettledSnapshot(snapshot.id, branchId())
    selectedSnapshot.value = res.data.snapshot
    transactions.value = res.data.transactions || []
  } catch (error: any) {
    toast.error(error.response?.data?.error || 'Gagal memuat rincian snapshot')
  }
}

async function captureAPI() {
  if (!selectedShop.value || capturing.value) return
  capturing.value = true
  try {
    const res = await marketplaceApi.captureUnsettledSnapshot({ shop_id: selectedShop.value, source: 'TIKTOK_API' }, branchId())
	const jobId = res.data.job_id
	if (!jobId) throw new Error('Server tidak mengembalikan ID proses')
	let snapshotId: number | undefined
	for (let attempt = 0; attempt < 300; attempt++) {
		await new Promise((resolve) => setTimeout(resolve, 2000))
		const status = (await marketplaceApi.getUnsettledJob(jobId)).data
		if (status.status === 'failed') throw new Error(status.error || 'Sinkronisasi gagal')
		if (status.status === 'completed') {
			snapshotId = status.snapshot_id
			break
		}
	}
	if (!snapshotId) throw new Error('Proses masih berjalan. Periksa riwayat snapshot beberapa saat lagi.')
    await load()
    const detail = await marketplaceApi.getUnsettledSnapshot(snapshotId, branchId())
	await selectSnapshot(detail.data.snapshot)
    toast.success('Snapshot transaksi belum cair tersimpan')
  } catch (error: any) {
    toast.error(error.response?.data?.error || 'Gagal mengambil transaksi TikTok')
  } finally {
    capturing.value = false
  }
}

function parseRupiah(raw: string): number | null {
  const digits = raw.replace(/[^0-9]/g, '')
  if (!digits) return null
  const amount = Number(digits)
  return Number.isSafeInteger(amount) ? amount : null
}

async function captureCenter() {
  if (!selectedShop.value || capturing.value) return
  const awaiting = parseRupiah(center.value.awaiting_delivery)
  const afterSales = parseRupiah(center.value.after_sales)
  const processing = parseRupiah(center.value.processing)
  if (awaiting === null || afterSales === null || processing === null) {
    toast.error('Isi ketiga nominal sesuai Seller Center')
    return
  }
  capturing.value = true
  try {
    const res = await marketplaceApi.captureUnsettledSnapshot({
      shop_id: selectedShop.value, source: 'SELLER_CENTER',
      awaiting_delivery: awaiting, after_sales: afterSales, processing,
    }, branchId())
    await load()
    if (res.data.snapshot) await selectSnapshot(res.data.snapshot)
    toast.success('Ringkasan Seller Center tersimpan')
  } catch (error: any) {
    toast.error(error.response?.data?.error || 'Gagal menyimpan ringkasan Seller Center')
  } finally {
    capturing.value = false
  }
}

watch(selectedShop, () => {
  snapshots.value = []
  selectedSnapshot.value = null
  transactions.value = []
  if (selectedShop.value) load()
})
watch(() => masterData.selectedBranchId, () => {
  selectedShop.value = ''
  load()
})
onMounted(load)
</script>

<template>
  <section class="rounded-2xl border border-amber-200 bg-white shadow-sm p-5 space-y-5">
    <div class="flex flex-wrap items-center justify-between gap-3">
      <div>
        <h2 class="text-lg font-black text-slate-900">Untuk Dibayar · TikTok Shop</h2>
        <p class="text-xs text-slate-500">Snapshot terpisah dari status order produksi. Angka API adalah estimasi hingga TikTok menyelesaikan pembayaran.</p>
      </div>
      <div class="flex flex-wrap gap-2 items-center">
        <select v-model="selectedShop" class="border rounded-lg px-3 py-2 text-xs" aria-label="Pilih toko TikTok">
          <option v-for="shop in shops" :key="shop" :value="shop">Toko {{ shop }}</option>
        </select>
        <button :disabled="capturing || !selectedShop" class="px-3 py-2 rounded-lg bg-amber-600 text-white text-xs font-bold disabled:opacity-50" @click="captureAPI">
          {{ capturing ? 'Mengambil data...' : 'Tarik data belum cair API' }}
        </button>
      </div>
    </div>

    <div v-if="loading" class="text-xs text-slate-500">Memuat snapshot...</div>
    <div v-else-if="shops.length === 0" class="text-xs text-slate-500">Belum ada akun TikTok yang terhubung pada cabang ini.</div>
    <template v-else>
      <div class="grid grid-cols-1 lg:grid-cols-2 gap-4">
        <div class="rounded-xl border border-slate-200 p-4 space-y-3">
          <div class="flex justify-between gap-2 items-start">
            <div>
              <h3 class="font-bold text-sm">Ringkasan Seller Center</h3>
              <p class="text-[11px] text-slate-500">Catat tiga angka yang tampil untuk toko yang dipilih.</p>
            </div>
            <span v-if="latestCenter" class="text-[10px] text-slate-500">{{ date(latestCenter.captured_at) }}</span>
          </div>
          <label v-for="category in categories" :key="category.key" class="flex items-center justify-between gap-2 text-xs">
            <span>{{ category.label }}</span>
            <input v-model="center[category.key]" inputmode="numeric" placeholder="Rp 0" class="border rounded-lg px-2 py-1.5 w-36 text-right" />
          </label>
          <button :disabled="capturing" class="px-3 py-2 rounded-lg bg-slate-900 text-white text-xs font-bold disabled:opacity-50" @click="captureCenter">Simpan angka Seller Center</button>
        </div>
        <div class="rounded-xl border border-slate-200 p-4 space-y-2">
          <div class="flex justify-between gap-2 items-start">
            <h3 class="font-bold text-sm">Transaksi TikTok API</h3>
            <span v-if="latestAPI" class="text-[10px] text-slate-500">{{ date(latestAPI.captured_at) }}</span>
          </div>
          <div v-for="category in categories" :key="category.key" class="flex justify-between gap-2 text-xs">
            <span>{{ category.label }}</span>
            <span class="text-right font-bold">{{ money(latestAPI?.[category.key]) }}<span v-if="latestCenter" class="block font-normal text-[10px] text-slate-500">Selisih: {{ money((latestAPI?.[category.key] || 0) - latestCenter[category.key]) }}</span></span>
          </div>
          <div class="flex justify-between text-xs border-t pt-2 font-bold"><span>Belum terklasifikasi</span><span>{{ money(latestAPI?.unclassified_amount) }}</span></div>
          <div class="flex justify-between text-sm border-t pt-2 font-black"><span>Total API</span><span>{{ money(total(latestAPI)) }}</span></div>
          <p v-if="latestAPI && latestCenter" class="text-[11px] text-slate-500">Total Seller Center {{ money(total(latestCenter)) }} · Selisih {{ money(total(latestAPI) - total(latestCenter)) }}. Bandingkan snapshot yang diambil berdekatan waktunya.</p>
        </div>
      </div>

      <div v-if="snapshots.length" class="space-y-2">
        <h3 class="font-bold text-xs">Riwayat snapshot</h3>
        <div class="flex flex-wrap gap-2">
          <button v-for="snapshot in snapshots" :key="snapshot.id" @click="selectSnapshot(snapshot)" class="border rounded-lg px-2 py-1.5 text-[11px] hover:bg-amber-50" :class="selectedSnapshot?.id === snapshot.id ? 'border-amber-500 bg-amber-50' : 'border-slate-200'">
            {{ snapshot.source === 'TIKTOK_API' ? 'API' : 'Seller Center' }} · {{ date(snapshot.captured_at) }} · {{ money(total(snapshot)) }}
          </button>
        </div>
      </div>
      <div v-if="selectedSnapshot?.source === 'TIKTOK_API'" class="space-y-2">
        <h3 class="font-bold text-xs">Rincian transaksi ({{ transactions.length }})</h3>
        <div class="max-h-80 overflow-auto border rounded-xl">
          <table class="w-full min-w-[650px] text-xs text-left">
            <thead class="sticky top-0 bg-slate-100"><tr><th class="p-2">Order / Penyesuaian</th><th class="p-2">Jenis</th><th class="p-2">Alasan TikTok</th><th class="p-2">Kategori</th><th class="p-2 text-right">Estimasi</th></tr></thead>
            <tbody><tr v-for="tx in transactions" :key="tx.id" class="border-t"><td class="p-2 font-mono">{{ tx.order_id || tx.adjustment_id || tx.transaction_id }}</td><td class="p-2">{{ tx.transaction_type }}</td><td class="p-2">{{ tx.unsettled_reason || '-' }}</td><td class="p-2">{{ tx.category }}</td><td class="p-2 text-right font-semibold">{{ money(tx.estimated_amount) }}</td></tr></tbody>
          </table>
        </div>
      </div>
    </template>
  </section>
</template>
