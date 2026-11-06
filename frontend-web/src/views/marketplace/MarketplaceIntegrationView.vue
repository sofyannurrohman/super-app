<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { marketplaceApi, type BranchMarketplaceAccount, type MarketplaceAccountDetail } from '@/api/marketplace'
import {
  Store,
  RefreshCw,
  ExternalLink,
  CheckCircle2,
  AlertCircle,
  XCircle,
  Trash2,
  Edit3,
  ShieldCheck,
  Building2,
  Activity,
  Zap,
  X,
  Settings,
  KeyRound,
  Info,
  Globe,
  ExternalLink as LinkIcon
} from 'lucide-vue-next'
import { Card, CardContent } from '@/components/ui/card'
import { toast } from 'vue-sonner'

const branchAccounts = ref<BranchMarketplaceAccount[]>([])
const isLoading = ref(true)
const testingBranchId = ref<string | null>(null)
const testResult = ref<{ branchId: string; success: boolean; message: string; details?: any } | null>(null)

// Shopee Open Platform Config State
const isShopeeConfigModalOpen = ref(false)
const isLoadingShopeeConfig = ref(false)
const isSavingShopeeConfig = ref(false)
const shopeeConfig = ref({
  partner_id: '',
  partner_key: '',
  environment: 'production',
  has_key: false,
  is_configured: false,
  base_url: '',
})

async function fetchShopeeConfig() {
  isLoadingShopeeConfig.value = true
  try {
    const res = await marketplaceApi.getShopeeConfig()
    shopeeConfig.value = {
      partner_id: res.data.partner_id || '',
      partner_key: '',
      environment: res.data.environment || 'production',
      has_key: res.data.has_key,
      is_configured: res.data.is_configured,
      base_url: res.data.base_url || '',
    }
  } catch (err: any) {
    console.error('Failed to load Shopee config', err)
  } finally {
    isLoadingShopeeConfig.value = false
  }
}

function openShopeeConfigModal() {
  fetchShopeeConfig()
  isShopeeConfigModalOpen.value = true
}

async function handleSaveShopeeConfig() {
  if (!shopeeConfig.value.partner_id.trim()) {
    toast.error('Partner ID wajib diisi')
    return
  }
  if (!shopeeConfig.value.partner_key.trim() && !shopeeConfig.value.has_key) {
    toast.error('Partner Key wajib diisi untuk konfigurasi awal')
    return
  }

  isSavingShopeeConfig.value = true
  try {
    await marketplaceApi.saveShopeeConfig({
      partner_id: shopeeConfig.value.partner_id.trim(),
      partner_key: shopeeConfig.value.partner_key.trim() || undefined,
      environment: shopeeConfig.value.environment,
    })
    toast.success('Konfigurasi Shopee Open Platform berhasil disimpan!')
    isShopeeConfigModalOpen.value = false
    await fetchShopeeConfig()
  } catch (err: any) {
    toast.error('Gagal menyimpan konfigurasi Shopee: ' + (err.response?.data?.error || err.message))
  } finally {
    isSavingShopeeConfig.value = false
  }
}

// Modal State for Manual Token Edit/Add
const isModalOpen = ref(false)
const isSaving = ref(false)
const modalForm = ref({
  id: undefined as number | undefined,
  branch_id: '',
  branch_name: '',
  marketplace: 'tiktok',
  shop_id: '',
  access_token: '',
  refresh_token: '',
  expires_at: '',
})

async function fetchAccounts() {
  isLoading.value = true
  try {
    const res = await marketplaceApi.getMarketplaceAccounts()
    branchAccounts.value = res.data
  } catch (err: any) {
    toast.error('Gagal memuat daftar akun marketplace: ' + (err.response?.data?.error || err.message))
  } finally {
    isLoading.value = false
  }
}

async function handleConnectTikTok(branch: BranchMarketplaceAccount) {
  try {
    const res = await marketplaceApi.getTikTokAuthUrl(branch.branch_id)
    if (res.data.auth_url) {
      window.open(res.data.auth_url, '_blank')
      toast.info(`Membuka halaman otorisasi TikTok untuk ${branch.branch_name}...`)
    }
  } catch (err: any) {
    toast.error('Gagal membuat link otorisasi: ' + (err.response?.data?.error || err.message))
  }
}

async function handleConnectShopee(branch: BranchMarketplaceAccount) {
  try {
    const res = await marketplaceApi.getShopeeAuthUrl(branch.branch_id)
    if (res.data.auth_url) {
      window.open(res.data.auth_url, '_blank')
      toast.info(`Membuka halaman otorisasi Shopee untuk ${branch.branch_name}...`)
    }
  } catch (err: any) {
    const errorMsg = err.response?.data?.error || err.message || 'Gagal membuat link otorisasi'
    toast.error('Otorisasi Shopee: ' + errorMsg)
    if (errorMsg.toLowerCase().includes('partner id') || errorMsg.toLowerCase().includes('belum')) {
      openShopeeConfigModal()
    }
  }
}

async function handleTestConnection(branch: BranchMarketplaceAccount, marketplace: string = 'tiktok') {
  testingBranchId.value = `${branch.branch_id}-${marketplace}`
  testResult.value = null
  try {
    let res: any
    if (marketplace === 'shopee') {
      res = await marketplaceApi.testShopeeConnection(branch.branch_id)
    } else {
      res = await marketplaceApi.testTikTokConnection(branch.branch_id)
    }

    if (res.data.success) {
      testResult.value = {
        branchId: `${branch.branch_id}-${marketplace}`,
        success: true,
        message: res.data.message || `Koneksi ke ${marketplace.toUpperCase()} aktif dan valid!`,
        details: res.data,
      }
      toast.success(`Koneksi ${marketplace.toUpperCase()} ${branch.branch_name} berhasil!`)
    } else {
      testResult.value = {
        branchId: `${branch.branch_id}-${marketplace}`,
        success: false,
        message: res.data.error || 'Koneksi gagal',
      }
      toast.error(`Koneksi ${marketplace.toUpperCase()} ${branch.branch_name} gagal: ${res.data.error}`)
    }
  } catch (err: any) {
    const errorMsg = err.response?.data?.error || err.message || 'Gagal menghubungi server'
    testResult.value = {
      branchId: `${branch.branch_id}-${marketplace}`,
      success: false,
      message: errorMsg,
    }
    toast.error(`Koneksi ${marketplace.toUpperCase()} ${branch.branch_name} gagal: ${errorMsg}`)
  } finally {
    testingBranchId.value = null
  }
}

function openEditModal(branch: BranchMarketplaceAccount, account?: MarketplaceAccountDetail) {
  modalForm.value = {
    id: account?.id,
    branch_id: branch.branch_id,
    branch_name: branch.branch_name,
    marketplace: account?.marketplace || 'tiktok',
    shop_id: account?.shop_id || '',
    access_token: '',
    refresh_token: '',
    expires_at: account?.expires_at ? new Date(account.expires_at).toISOString().slice(0, 16) : '',
  }
  isModalOpen.value = true
}

async function handleSaveAccount() {
  if (!modalForm.value.shop_id.trim()) {
    toast.error('Shop Cipher / Shop ID wajib diisi')
    return
  }
  if (!modalForm.value.access_token.trim() && !modalForm.value.id) {
    toast.error('Access Token wajib diisi untuk akun baru')
    return
  }

  isSaving.value = true
  try {
    await marketplaceApi.saveMarketplaceAccount({
      id: modalForm.value.id,
      branch_id: modalForm.value.branch_id,
      marketplace: modalForm.value.marketplace,
      shop_id: modalForm.value.shop_id.trim(),
      access_token: modalForm.value.access_token.trim(),
      refresh_token: modalForm.value.refresh_token.trim() || undefined,
      expires_at: modalForm.value.expires_at ? new Date(modalForm.value.expires_at).toISOString() : undefined,
    })
    toast.success(`Akun marketplace untuk ${modalForm.value.branch_name} berhasil disimpan`)
    isModalOpen.value = false
    await fetchAccounts()
  } catch (err: any) {
    toast.error('Gagal menyimpan akun: ' + (err.response?.data?.error || err.message))
  } finally {
    isSaving.value = false
  }
}

async function handleDeleteAccount(branch: BranchMarketplaceAccount, account: MarketplaceAccountDetail) {
  if (!confirm(`Yakin ingin memutuskan koneksi akun ${account.marketplace.toUpperCase()} untuk cabang ${branch.branch_name}?`)) {
    return
  }
  try {
    await marketplaceApi.deleteMarketplaceAccount(account.id)
    toast.success(`Koneksi ${account.marketplace.toUpperCase()} untuk ${branch.branch_name} telah dihapus`)
    await fetchAccounts()
  } catch (err: any) {
    toast.error('Gagal menghapus akun: ' + (err.response?.data?.error || err.message))
  }
}

function formatDate(dateStr?: string) {
  if (!dateStr || dateStr.startsWith('0001-01-01')) return '-'
  return new Date(dateStr).toLocaleString('id-ID', {
    day: '2-digit',
    month: 'short',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}

const totalBranches = computed(() => branchAccounts.value.length)
const connectedCount = computed(() => {
  return branchAccounts.value.filter(b => b.accounts.some(a => a.is_connected && !a.is_expired)).length
})
const windowLocationOrigin = computed(() => typeof window !== 'undefined' ? window.location.origin : '')

function onWindowFocus() {
  fetchAccounts()
}

function onWindowMessage(event: MessageEvent) {
  if (event.data?.type === 'TIKTOK_AUTH_SUCCESS') {
    fetchAccounts()
    toast.success('Otorisasi TikTok berhasil diperbarui!')
  } else if (event.data?.type === 'SHOPEE_AUTH_SUCCESS') {
    fetchAccounts()
    toast.success('Otorisasi Shopee berhasil diperbarui!')
  }
}

function onStorageChange(event: StorageEvent) {
  if (event.key === 'tiktok_auth_completed' || event.key === 'shopee_auth_completed') {
    fetchAccounts()
  }
}

onMounted(() => {
  fetchAccounts()
  fetchShopeeConfig()
  window.addEventListener('focus', onWindowFocus)
  window.addEventListener('message', onWindowMessage)
  window.addEventListener('storage', onStorageChange)
})

onUnmounted(() => {
  window.removeEventListener('focus', onWindowFocus)
  window.removeEventListener('message', onWindowMessage)
  window.removeEventListener('storage', onStorageChange)
})
</script>

<template>
  <div class="space-y-6 animate-in fade-in duration-500">
    <!-- Header -->
    <div class="flex flex-col md:flex-row justify-between items-start md:items-center gap-4">
      <div>
        <h1 class="text-2xl font-black tracking-tight text-slate-900 flex items-center gap-3">
          <div class="p-2 rounded-xl bg-gradient-to-br from-indigo-500 to-purple-600 shadow-lg shadow-purple-200">
            <Store class="h-6 w-6 text-white" />
          </div>
          Integrasi Akun Marketplace per Cabang
        </h1>
        <p class="text-slate-500 text-sm font-medium mt-1">
          Hubungkan akun TikTok Seller Center & Shopee Open Platform serta kelola token API untuk masing-masing cabang
        </p>
      </div>
      <div class="flex items-center gap-3 flex-wrap">
        <button
          @click="openShopeeConfigModal"
          class="flex items-center gap-2 bg-gradient-to-r from-orange-500 to-amber-600 hover:from-orange-600 hover:to-amber-700 text-white px-4 py-2.5 rounded-xl font-bold text-sm shadow-md shadow-orange-200 transition-all duration-200"
        >
          <Settings class="h-4 w-4" />
          Pengaturan Shopee API
          <span
            v-if="shopeeConfig.is_configured"
            class="ml-1 px-1.5 py-0.5 text-[10px] font-extrabold bg-white/20 text-white rounded-full"
          >Aktif</span>
          <span
            v-else
            class="ml-1 px-1.5 py-0.5 text-[10px] font-extrabold bg-rose-500 text-white rounded-full animate-pulse"
          >Belum Diatur</span>
        </button>

        <button
          @click="fetchAccounts"
          :disabled="isLoading"
          class="flex items-center gap-2 bg-white border border-slate-200 text-slate-700 hover:bg-slate-50 px-4 py-2.5 rounded-xl font-bold text-sm shadow-sm transition-all duration-200 disabled:opacity-60"
        >
          <RefreshCw class="h-4 w-4" :class="{ 'animate-spin': isLoading }" />
          Refresh Data
        </button>
      </div>
    </div>

    <!-- Summary Overview Cards -->
    <div class="grid grid-cols-1 sm:grid-cols-3 gap-4">
      <Card class="border-slate-200/80 shadow-sm bg-gradient-to-br from-white to-slate-50">
        <CardContent class="p-5 flex items-center justify-between">
          <div>
            <p class="text-xs font-bold text-slate-400 uppercase tracking-wider">Total Cabang</p>
            <p class="text-2xl font-black text-slate-800 mt-1">{{ totalBranches }}</p>
          </div>
          <div class="p-3 bg-slate-100 rounded-xl text-slate-600">
            <Building2 class="h-6 w-6" />
          </div>
        </CardContent>
      </Card>

      <Card class="border-emerald-200/80 shadow-sm bg-gradient-to-br from-white to-emerald-50/30">
        <CardContent class="p-5 flex items-center justify-between">
          <div>
            <p class="text-xs font-bold text-emerald-600 uppercase tracking-wider">Cabang Terhubung</p>
            <p class="text-2xl font-black text-emerald-700 mt-1">{{ connectedCount }}</p>
          </div>
          <div class="p-3 bg-emerald-100/70 rounded-xl text-emerald-600">
            <ShieldCheck class="h-6 w-6" />
          </div>
        </CardContent>
      </Card>

      <Card class="border-amber-200/80 shadow-sm bg-gradient-to-br from-white to-amber-50/30">
        <CardContent class="p-5 flex items-center justify-between">
          <div>
            <p class="text-xs font-bold text-amber-600 uppercase tracking-wider">Belum Terhubung</p>
            <p class="text-2xl font-black text-amber-700 mt-1">{{ totalBranches - connectedCount }}</p>
          </div>
          <div class="p-3 bg-amber-100/70 rounded-xl text-amber-600">
            <AlertCircle class="h-6 w-6" />
          </div>
        </CardContent>
      </Card>
    </div>

    <!-- Branch Integration Cards List -->
    <div class="space-y-4">
      <div v-if="isLoading" class="flex flex-col items-center justify-center p-12 bg-white rounded-2xl border border-slate-200">
        <RefreshCw class="h-8 w-8 text-indigo-500 animate-spin mb-3" />
        <p class="text-slate-500 text-sm font-medium">Memuat data integrasi cabang...</p>
      </div>

      <div
        v-else
        v-for="branch in branchAccounts"
        :key="branch.branch_id"
        class="bg-white rounded-2xl border border-slate-200 shadow-sm p-6 transition-all hover:shadow-md"
      >
        <div class="flex flex-col lg:flex-row justify-between items-start lg:items-center gap-4 border-b border-slate-100 pb-5">
          <div class="flex items-start gap-3">
            <div class="p-2.5 rounded-xl bg-slate-100 text-slate-700">
              <Building2 class="h-5 w-5" />
            </div>
            <div>
              <div class="flex items-center gap-2">
                <h3 class="text-lg font-black text-slate-900">{{ branch.branch_name }}</h3>
                <span v-if="branch.branch_name.includes('Solo')" class="px-2 py-0.5 text-xs font-bold rounded-full bg-purple-50 text-purple-700 border border-purple-200">
                  Cabang Utama
                </span>
              </div>
              <p class="text-xs text-slate-400 mt-0.5">{{ branch.branch_address || 'ID Cabang: ' + branch.branch_id }}</p>
            </div>
          </div>

          <!-- Action buttons for the branch -->
          <div class="flex flex-wrap items-center gap-2 w-full lg:w-auto">
            <button
              @click="handleConnectTikTok(branch)"
              class="flex-1 lg:flex-none flex items-center justify-center gap-2 bg-gradient-to-r from-slate-900 to-slate-800 text-white hover:from-slate-800 hover:to-slate-700 px-4 py-2 rounded-xl text-xs font-bold shadow-sm transition-all"
            >
              <Zap class="h-3.5 w-3.5 text-yellow-400" />
              Otorisasi TikTok
              <ExternalLink class="h-3 w-3 opacity-60" />
            </button>

            <button
              @click="handleConnectShopee(branch)"
              class="flex-1 lg:flex-none flex items-center justify-center gap-2 bg-gradient-to-r from-orange-600 to-amber-500 text-white hover:from-orange-700 hover:to-amber-600 px-4 py-2 rounded-xl text-xs font-bold shadow-sm transition-all"
            >
              <Zap class="h-3.5 w-3.5 text-white" />
              Otorisasi Shopee
              <ExternalLink class="h-3 w-3 opacity-60" />
            </button>

            <button
              @click="openEditModal(branch, branch.accounts.find(a => a.marketplace === 'tiktok'))"
              class="flex-1 lg:flex-none flex items-center justify-center gap-1.5 bg-indigo-50 hover:bg-indigo-100 text-indigo-700 px-3.5 py-2 rounded-xl text-xs font-bold transition-all border border-indigo-200/60"
            >
              <Edit3 class="h-3.5 w-3.5" />
              Input Token Manual
            </button>
          </div>
        </div>

        <!-- Account details area -->
        <div class="pt-5">
          <div v-if="branch.accounts && branch.accounts.length > 0" class="space-y-3">
            <div
              v-for="acc in branch.accounts"
              :key="acc.id"
              class="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4 p-4 rounded-xl border"
              :class="acc.marketplace === 'shopee' ? 'bg-orange-50/40 border-orange-100' : 'bg-slate-50 border-slate-100'"
            >
              <div class="flex items-center gap-3">
                <div 
                  class="w-9 h-9 rounded-lg flex items-center justify-center text-white font-bold text-xs shadow-sm"
                  :class="acc.marketplace === 'shopee' ? 'bg-gradient-to-br from-orange-500 to-amber-600' : 'bg-black'"
                >
                  {{ acc.marketplace === 'shopee' ? 'SP' : 'TT' }}
                </div>
                <div>
                  <div class="flex items-center gap-2">
                    <span class="text-sm font-bold text-slate-800 uppercase">{{ acc.marketplace === 'shopee' ? 'Shopee' : 'TikTok' }} Shop</span>
                    <span
                      v-if="acc.is_connected && !acc.is_expired"
                      class="inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-xs font-bold bg-emerald-100 text-emerald-800"
                    >
                      <span class="w-1.5 h-1.5 rounded-full bg-emerald-500"></span>
                      Terhubung
                    </span>
                    <span
                      v-else-if="acc.is_connected && acc.is_expired"
                      class="inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-xs font-bold bg-amber-100 text-amber-800"
                    >
                      <span class="w-1.5 h-1.5 rounded-full bg-amber-500"></span>
                      Token Kadaluarsa
                    </span>
                  </div>
                  <div class="flex flex-wrap items-center gap-x-4 gap-y-1 text-xs text-slate-500 mt-1">
                    <span>{{ acc.marketplace === 'shopee' ? 'Shop ID' : 'Shop Cipher' }}: <code class="font-bold text-slate-700 bg-white px-1.5 py-0.5 rounded border border-slate-200">{{ acc.shop_id }}</code></span>
                    <span>Kadaluarsa: <strong class="text-slate-700">{{ formatDate(acc.expires_at) }}</strong></span>
                    <span>Terakhir Update: {{ formatDate(acc.updated_at) }}</span>
                  </div>
                </div>
              </div>

              <!-- Test connection & delete actions -->
              <div class="flex items-center gap-2 w-full sm:w-auto justify-end">
                <button
                  @click="handleTestConnection(branch, acc.marketplace)"
                  :disabled="testingBranchId === `${branch.branch_id}-${acc.marketplace}`"
                  class="flex items-center gap-1.5 px-3 py-1.5 text-xs font-bold rounded-lg border border-slate-200 bg-white hover:bg-slate-50 text-slate-700 transition-all disabled:opacity-50"
                >
                  <Activity class="h-3.5 w-3.5 text-indigo-500" :class="{ 'animate-pulse': testingBranchId === `${branch.branch_id}-${acc.marketplace}` }" />
                  {{ testingBranchId === `${branch.branch_id}-${acc.marketplace}` ? 'Menguji...' : 'Uji Koneksi' }}
                </button>

                <button
                  @click="handleDeleteAccount(branch, acc)"
                  class="p-1.5 text-slate-400 hover:text-red-600 hover:bg-red-50 rounded-lg transition-all"
                  title="Putuskan sambungan akun"
                >
                  <Trash2 class="h-4 w-4" />
                </button>
              </div>
            </div>

            <!-- Test result alert box -->
            <div
              v-if="testResult && testResult.branchId.startsWith(branch.branch_id)"
              class="p-4 rounded-xl border text-xs animate-in fade-in duration-300"
              :class="testResult.success ? 'bg-emerald-50 border-emerald-200 text-emerald-800' : 'bg-rose-50 border-rose-200 text-rose-800'"
            >
              <div class="flex items-center gap-2 font-bold mb-1">
                <CheckCircle2 v-if="testResult.success" class="h-4 w-4 text-emerald-600" />
                <XCircle v-else class="h-4 w-4 text-rose-600" />
                {{ testResult.success ? 'Hasil Uji: Sukses Terhubung' : 'Hasil Uji: Gagal' }}
              </div>
              <p>{{ testResult.message }}</p>
              <div v-if="testResult.details && testResult.details.warehouse_count !== undefined" class="mt-1 text-slate-600">
                Gudang terdeteksi dari TikTok: <strong>{{ testResult.details.warehouse_count }} gudang logistik</strong>
              </div>
              <div v-if="testResult.details && testResult.details.product_count !== undefined" class="mt-1 text-slate-600">
                Produk terdeteksi dari Shopee: <strong>{{ testResult.details.product_count }} produk</strong>
              </div>
            </div>
          </div>

          <!-- Empty state for branch -->
          <div v-else class="p-6 text-center rounded-xl border border-dashed border-slate-200 bg-slate-50/50">
            <p class="text-sm font-medium text-slate-500 mb-2">Belum ada akun Marketplace yang terhubung untuk cabang ini.</p>
            <p class="text-xs text-slate-400">
              Klik <strong>Otorisasi TikTok</strong> atau <strong>Otorisasi Shopee</strong> untuk menghubungkan akun via login partner, atau <strong>Input Manual</strong> jika sudah memiliki kredensial.
            </p>
          </div>
        </div>
      </div>
    </div>

    <!-- Manual Token Modal -->
    <div
      v-if="isModalOpen"
      class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-900/60 backdrop-blur-sm animate-in fade-in duration-200"
    >
      <div class="bg-white rounded-2xl shadow-2xl border border-slate-200 w-full max-w-lg overflow-hidden animate-in zoom-in-95 duration-200">
        <div class="flex justify-between items-center p-5 border-b border-slate-100">
          <div>
            <h3 class="text-lg font-bold text-slate-900">Konfigurasi Token Akun Marketplace</h3>
            <p class="text-xs text-slate-500 mt-0.5">Cabang: <strong>{{ modalForm.branch_name }}</strong></p>
          </div>
          <button @click="isModalOpen = false" class="text-slate-400 hover:text-slate-600 p-1.5 rounded-lg">
            <X class="h-5 w-5" />
          </button>
        </div>

        <form @submit.prevent="handleSaveAccount" class="p-5 space-y-4">
          <div>
            <label class="block text-xs font-bold text-slate-700 uppercase mb-1">Platform Marketplace</label>
            <select
              v-model="modalForm.marketplace"
              class="w-full px-3.5 py-2.5 rounded-xl border border-slate-200 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500/20 focus:border-indigo-500 font-medium"
            >
              <option value="tiktok">TikTok Shop</option>
              <option value="shopee">Shopee</option>
              <option value="tokopedia">Tokopedia</option>
            </select>
          </div>

          <div>
            <label class="block text-xs font-bold text-slate-700 uppercase mb-1">
              Shop ID / Shop Cipher <span class="text-rose-500">*</span>
            </label>
            <input
              v-model="modalForm.shop_id"
              type="text"
              placeholder="Contoh: ROW_gM5XXXXXX atau Shop Cipher"
              required
              class="w-full px-3.5 py-2.5 rounded-xl border border-slate-200 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500/20 focus:border-indigo-500 font-mono"
            />
            <p class="text-xs text-slate-400 mt-1">Diambil dari TikTok Partner Center / Toko Seller Center.</p>
          </div>

          <div>
            <label class="block text-xs font-bold text-slate-700 uppercase mb-1">
              Access Token <span v-if="!modalForm.id" class="text-rose-500">*</span>
            </label>
            <textarea
              v-model="modalForm.access_token"
              rows="3"
              :placeholder="modalForm.id ? 'Kosongkan jika tidak ingin mengubah access token' : 'Masukkan Access Token API'"
              :required="!modalForm.id"
              class="w-full px-3.5 py-2.5 rounded-xl border border-slate-200 text-xs focus:outline-none focus:ring-2 focus:ring-indigo-500/20 focus:border-indigo-500 font-mono"
            ></textarea>
          </div>

          <div>
            <label class="block text-xs font-bold text-slate-700 uppercase mb-1">Refresh Token (Opsional)</label>
            <textarea
              v-model="modalForm.refresh_token"
              rows="2"
              placeholder="Masukkan Refresh Token (opsional)"
              class="w-full px-3.5 py-2.5 rounded-xl border border-slate-200 text-xs focus:outline-none focus:ring-2 focus:ring-indigo-500/20 focus:border-indigo-500 font-mono"
            ></textarea>
          </div>

          <div>
            <label class="block text-xs font-bold text-slate-700 uppercase mb-1">Tanggal Kadaluarsa Token (Opsional)</label>
            <input
              v-model="modalForm.expires_at"
              type="datetime-local"
              class="w-full px-3.5 py-2.5 rounded-xl border border-slate-200 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500/20 focus:border-indigo-500"
            />
            <p class="text-xs text-slate-400 mt-1">Default 7 hari dari sekarang jika dikosongkan.</p>
          </div>

          <div class="flex justify-end gap-3 pt-3 border-t border-slate-100">
            <button
              type="button"
              @click="isModalOpen = false"
              class="px-4 py-2 rounded-xl text-sm font-bold text-slate-600 hover:bg-slate-100 transition-all"
            >
              Batal
            </button>
            <button
              type="submit"
              :disabled="isSaving"
              class="px-5 py-2 rounded-xl text-sm font-bold text-white bg-indigo-600 hover:bg-indigo-700 shadow-md shadow-indigo-200 transition-all disabled:opacity-60 flex items-center gap-2"
            >
              <RefreshCw v-if="isSaving" class="h-4 w-4 animate-spin" />
              {{ isSaving ? 'Menyimpan...' : 'Simpan Akun' }}
            </button>
          </div>
        </form>
      </div>
    </div>

    <!-- Shopee Open Platform Global Configuration Modal -->
    <div
      v-if="isShopeeConfigModalOpen"
      class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-900/60 backdrop-blur-sm animate-in fade-in duration-200"
    >
      <div class="bg-white rounded-2xl shadow-2xl border border-slate-200 w-full max-w-xl overflow-hidden animate-in zoom-in-95 duration-200">
        <!-- Header -->
        <div class="flex justify-between items-center p-5 bg-gradient-to-r from-orange-500 to-amber-600 text-white">
          <div class="flex items-center gap-3">
            <div class="p-2 bg-white/20 rounded-xl backdrop-blur-sm">
              <Settings class="h-5 w-5 text-white" />
            </div>
            <div>
              <h3 class="text-lg font-black text-white">Pengaturan Shopee Open Platform API</h3>
              <p class="text-xs text-orange-100 mt-0.5">Kredensial Developer App dari Shopee Open Platform Console</p>
            </div>
          </div>
          <button @click="isShopeeConfigModalOpen = false" class="text-white/80 hover:text-white p-1.5 rounded-lg hover:bg-white/10 transition-colors">
            <X class="h-5 w-5" />
          </button>
        </div>

        <!-- Body Form -->
        <form @submit.prevent="handleSaveShopeeConfig" class="p-6 space-y-4">
          <!-- Guide banner -->
          <div class="p-3.5 bg-amber-50 border border-amber-200 rounded-xl text-xs text-amber-900 flex items-start gap-2.5 leading-relaxed">
            <Info class="h-4 w-4 text-amber-600 shrink-0 mt-0.5" />
            <div>
              <strong>Mengapa muncul "Wrong sign"?</strong>
              <p class="mt-0.5 text-amber-800">
                Shopee memverifikasi tanda tangan kriptografi (HMAC-SHA256) menggunakan <strong>Partner Key</strong> milik aplikasi Anda.
                Pastikan Partner ID dan Partner Key di bawah ini cocok persis dengan yang ada di <em>Shopee Open Platform Console &gt; App Management &gt; App Info</em>.
              </p>
            </div>
          </div>

          <!-- Environment selector -->
          <div>
            <label class="block text-xs font-bold text-slate-700 uppercase mb-1 flex items-center gap-1.5">
              <Globe class="h-3.5 w-3.5 text-slate-500" />
              Lingkungan Shopee (Environment) <span class="text-rose-500">*</span>
            </label>
            <div class="grid grid-cols-2 gap-3">
              <label
                class="flex items-center gap-2.5 p-3 rounded-xl border cursor-pointer transition-all"
                :class="shopeeConfig.environment === 'production' ? 'bg-orange-50 border-orange-500 text-orange-950 font-bold ring-2 ring-orange-500/20' : 'border-slate-200 hover:bg-slate-50 text-slate-700'"
              >
                <input
                  type="radio"
                  name="shopee_env"
                  value="production"
                  v-model="shopeeConfig.environment"
                  class="text-orange-600 focus:ring-orange-500"
                />
                <div>
                  <div class="text-xs font-bold">Live / Production</div>
                  <div class="text-[10px] text-slate-500">Akun Seller Shopee Asli</div>
                </div>
              </label>

              <label
                class="flex items-center gap-2.5 p-3 rounded-xl border cursor-pointer transition-all"
                :class="shopeeConfig.environment === 'sandbox' ? 'bg-orange-50 border-orange-500 text-orange-950 font-bold ring-2 ring-orange-500/20' : 'border-slate-200 hover:bg-slate-50 text-slate-700'"
              >
                <input
                  type="radio"
                  name="shopee_env"
                  value="sandbox"
                  v-model="shopeeConfig.environment"
                  class="text-orange-600 focus:ring-orange-500"
                />
                <div>
                  <div class="text-xs font-bold">Test / Sandbox</div>
                  <div class="text-[10px] text-slate-500">Test App Shopee Console</div>
                </div>
              </label>
            </div>
          </div>

          <!-- Partner ID -->
          <div>
            <label class="block text-xs font-bold text-slate-700 uppercase mb-1 flex items-center gap-1.5">
              <KeyRound class="h-3.5 w-3.5 text-slate-500" />
              Partner ID (App ID) <span class="text-rose-500">*</span>
            </label>
            <input
              v-model="shopeeConfig.partner_id"
              type="text"
              placeholder="Contoh: 2007890 atau Partner ID Shopee Anda"
              required
              class="w-full px-3.5 py-2.5 rounded-xl border border-slate-200 text-sm focus:outline-none focus:ring-2 focus:ring-orange-500/20 focus:border-orange-500 font-mono"
            />
            <p class="text-[11px] text-slate-400 mt-1">Ditemukan di Shopee Open Platform Console &gt; App List &gt; Partner ID.</p>
          </div>

          <!-- Partner Key -->
          <div>
            <label class="block text-xs font-bold text-slate-700 uppercase mb-1 flex items-center gap-1.5">
              <KeyRound class="h-3.5 w-3.5 text-slate-500" />
              Partner Key (App Secret) <span v-if="!shopeeConfig.has_key" class="text-rose-500">*</span>
            </label>
            <input
              v-model="shopeeConfig.partner_key"
              type="password"
              :placeholder="shopeeConfig.has_key ? '•••••••• (Biarkan kosong jika tidak ingin mengubah)' : 'Masukkan Partner Key Shopee Anda'"
              :required="!shopeeConfig.has_key"
              class="w-full px-3.5 py-2.5 rounded-xl border border-slate-200 text-sm focus:outline-none focus:ring-2 focus:ring-orange-500/20 focus:border-orange-500 font-mono"
            />
            <p class="text-[11px] text-slate-400 mt-1">Shopee Console &gt; App Info &gt; Partner Key. Disimpan secara aman di database server.</p>
          </div>

          <!-- Redirect URL info -->
          <div class="p-3 bg-slate-50 border border-slate-200 rounded-xl text-xs space-y-1">
            <div class="font-bold text-slate-700 flex items-center justify-between">
              <span>Redirect URL (Callback)</span>
              <span class="text-[10px] text-slate-400">Daftarkan di Shopee Console</span>
            </div>
            <p class="text-slate-500 text-[11px]">
              Pastikan URL Callback berikut terdaftar di <em>Shopee Open Platform &gt; App &gt; Redirect URL</em>:
            </p>
            <code class="block p-2 bg-white rounded border border-slate-200 text-slate-800 font-mono text-[11px] break-all select-all">
              {{ windowLocationOrigin }}/api/v1/marketplaces/shopee/callback
            </code>
          </div>

          <!-- Actions -->
          <div class="flex justify-end gap-3 pt-3 border-t border-slate-100">
            <button
              type="button"
              @click="isShopeeConfigModalOpen = false"
              class="px-4 py-2 rounded-xl text-sm font-bold text-slate-600 hover:bg-slate-100 transition-all"
            >
              Batal
            </button>
            <button
              type="submit"
              :disabled="isSavingShopeeConfig"
              class="px-5 py-2 rounded-xl text-sm font-bold text-white bg-gradient-to-r from-orange-500 to-amber-600 hover:from-orange-600 hover:to-amber-700 shadow-md shadow-orange-200 transition-all disabled:opacity-60 flex items-center gap-2"
            >
              <RefreshCw v-if="isSavingShopeeConfig" class="h-4 w-4 animate-spin" />
              {{ isSavingShopeeConfig ? 'Menyimpan...' : 'Simpan Konfigurasi Shopee' }}
            </button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>
