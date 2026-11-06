<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import apiClient from '@/api/axios'
import { toast } from 'vue-sonner'
import { useAuthStore } from '@/stores/auth'
import { useMasterDataStore } from '@/stores/masterData'
import {
  Plus,
  Pencil,
  Trash2,
  Package,
  RefreshCw,
  RotateCcw,
  Tags,
} from 'lucide-vue-next'

interface AncillaryGood {
  id: string
  company_id: string
  name: string
  price: number
  unit: string
  is_return: boolean
  description: string
  is_active: boolean
  sort_order: number
  created_at: string
  stock?: number
  factory_id?: string
  initial_stock?: number
}

const goods = ref<AncillaryGood[]>([])
const companies = ref<any[]>([])
const branches = ref<any[]>([])
const isLoading = ref(true)
const showModal = ref(false)
const isEditing = ref(false)
const isSaving = ref(false)
const showDeleteConfirm = ref(false)
const deleteTargetId = ref('')
const isDeleting = ref(false)
const includeInactive = ref(false)
const filterCompanyId = ref('')

const authStore = useAuthStore()
const masterDataStore = useMasterDataStore()
const stocks = ref<any[]>([])

const showAdjust = ref(false)
const targetAdjustGood = ref<any>(null)
const isSubmitting = ref(false)
const adjustForm = ref({ quantity: 0, reason: '', factory_id: '' })

const defaultForm = (): Omit<AncillaryGood, 'id' | 'created_at'> => ({
  company_id: '',
  factory_id: '',
  name: '',
  price: 0,
  unit: 'PCS',
  is_return: false,
  description: '',
  is_active: true,
  sort_order: 0,
  initial_stock: 0,
})

const form = ref(defaultForm())
const editId = ref('')

const currency = new Intl.NumberFormat('id-ID', {
  style: 'currency',
  currency: 'IDR',
  minimumFractionDigits: 0,
})

// ── Computed ──────────────────────────────────────────────────────────────────
const goodsWithStock = computed(() => {
  return goods.value.map(g => {
    const stockItem = stocks.value.find(s => s.ancillary_good_id === g.id)
    return {
      ...g,
      stock: stockItem ? stockItem.quantity : 0,
      factory_id: stockItem ? stockItem.factory_id : null,
    }
  })
})

const filteredGoods = computed(() => {
  return goodsWithStock.value.filter((g) => {
    const matchCompany = !filterCompanyId.value || g.company_id === filterCompanyId.value
    const matchActive = includeInactive.value || g.is_active
    return matchCompany && matchActive
  })
})

const regularGoods = computed(() => filteredGoods.value.filter((g) => !g.is_return))
const returnGoods = computed(() => filteredGoods.value.filter((g) => g.is_return))

// ── API Calls ─────────────────────────────────────────────────────────────────
async function fetchGoods() {
  isLoading.value = true
  try {
    const params: any = { include_inactive: includeInactive.value }
    if (filterCompanyId.value) params.company_id = filterCompanyId.value
    
    const [resGoods, resStocks] = await Promise.all([
      apiClient.get('/sales/ancillary-goods', { params }),
      apiClient.get('/sales/ancillary-goods/inventory', { 
        params: { company_id: filterCompanyId.value || authStore.user?.company_id } 
      }).catch(() => ({ data: { data: [] } }))
    ])
    
    goods.value = resGoods.data.data ?? []
    stocks.value = resStocks.data?.data ?? []
  } catch (e: any) {
    toast.error('Gagal memuat data barang: ' + (e?.response?.data?.error ?? e.message))
  } finally {
    isLoading.value = false
  }
}

async function fetchCompanies() {
  try {
    const res = await apiClient.get('/companies')
    companies.value = res.data.data ?? res.data ?? []
  } catch (_) {}
}

async function fetchBranches() {
  try {
    const res = await apiClient.get('/branches')
    branches.value = res.data.data ?? res.data ?? []
  } catch (_) {}
}

async function saveGood() {
  if (!form.value.name) return toast.error('Nama barang tidak boleh kosong')
  
  if (!isEditing.value) {
    if (!form.value.factory_id) return toast.error('Pilih cabang terlebih dahulu')
    const selectedBranch = branches.value.find(b => b.id === form.value.factory_id)
    if (selectedBranch) {
      form.value.company_id = selectedBranch.company_id
    } else {
      return toast.error('Gagal menemukan data perusahaan dari cabang terpilih')
    }
  } else {
    if (!form.value.company_id) return toast.error('Perusahaan tidak valid')
  }

  if (form.value.price < 0) return toast.error('Harga tidak boleh negatif')

  isSaving.value = true
  try {
    if (isEditing.value) {
      await apiClient.put(`/sales/ancillary-goods/${editId.value}`, form.value)
      toast.success('Barang berhasil diperbarui')
    } else {
      const res = await apiClient.post('/sales/ancillary-goods', form.value)
      toast.success('Barang berhasil ditambahkan')
      
      // Auto adjust initial stock if provided
      if (form.value.initial_stock && form.value.initial_stock > 0) {
        try {
          await apiClient.post('/sales/ancillary-goods/inventory/adjust', {
            factory_id: form.value.factory_id,
            ancillary_good_id: res.data.data.id,
            quantity: form.value.initial_stock,
            reason: 'Stok Awal'
          })
        } catch (stockErr) {
          console.error(stockErr)
        }
      }
    }
    closeModal()
    fetchGoods()
  } catch (e: any) {
    toast.error('Gagal menyimpan: ' + (e?.response?.data?.error ?? e.message))
  } finally {
    isSaving.value = false
  }
}

async function confirmDelete() {
  isDeleting.value = true
  try {
    await apiClient.delete(`/sales/ancillary-goods/${deleteTargetId.value}`)
    toast.success('Barang berhasil dihapus')
    showDeleteConfirm.value = false
    fetchGoods()
  } catch (e: any) {
    toast.error('Gagal menghapus: ' + (e?.response?.data?.error ?? e.message))
  } finally {
    isDeleting.value = false
  }
}

// ── Stok Adjustment ───────────────────────────────────────────────────────────
function openAdjust(good: any) {
  targetAdjustGood.value = good
  
  let initialFactoryId = good.factory_id || ''
  if (!initialFactoryId && masterDataStore.selectedBranchId && masterDataStore.selectedBranchId !== 'ALL_BRANCHES') {
    initialFactoryId = masterDataStore.selectedBranchId
  }
  
  adjustForm.value = { quantity: 0, reason: '', factory_id: initialFactoryId }
  showAdjust.value = true
}

async function submitAdjust() {
  if (!targetAdjustGood.value) return
  if (!adjustForm.value.factory_id) return toast.error('Pilih cabang terlebih dahulu')
  
  isSubmitting.value = true
  try {
    const payload = {
      factory_id: adjustForm.value.factory_id,
      ancillary_good_id: targetAdjustGood.value.id,
      quantity: adjustForm.value.quantity,
      reason: adjustForm.value.reason
    }
    
    await apiClient.post('/sales/ancillary-goods/inventory/adjust', payload)
    toast.success('Stok berhasil disesuaikan')
    showAdjust.value = false
    fetchGoods()
  } catch (error: any) {
    console.error(error)
    toast.error('Gagal menyesuaikan stok: ' + (error?.response?.data?.error ?? error.message))
  } finally {
    isSubmitting.value = false
  }
}

// ── Modal ─────────────────────────────────────────────────────────────────────
function openCreate() {
  form.value = defaultForm()
  if (masterDataStore.selectedBranchId && masterDataStore.selectedBranchId !== 'ALL_BRANCHES') {
    form.value.factory_id = masterDataStore.selectedBranchId
  }
  isEditing.value = false
  showModal.value = true
}

function openEdit(good: AncillaryGood) {
  form.value = { ...good }
  editId.value = good.id
  isEditing.value = true
  showModal.value = true
}

function closeModal() {
  showModal.value = false
  form.value = defaultForm()
  editId.value = ''
}

function openDelete(id: string) {
  deleteTargetId.value = id
  showDeleteConfirm.value = true
}

onMounted(() => {
  fetchCompanies()
  fetchBranches()
  fetchGoods()
})
</script>

<template>
  <div class="min-h-screen bg-slate-50">
    <!-- Header -->
    <div class="bg-white border-b border-slate-100 shadow-sm">
      <div class="max-w-7xl mx-auto px-6 py-5 flex flex-col md:flex-row md:items-center justify-between gap-4">
        <div class="flex items-center gap-3">
          <div class="w-10 h-10 rounded-xl bg-primary flex items-center justify-center shadow-md">
            <Tags class="text-white" :size="20" />
          </div>
          <div>
            <h1 class="text-xl font-bold text-slate-800">Master & Stok Barang Sampingan</h1>
            <p class="text-xs text-slate-400">Krat, Botol, Galon, Mangkok, Piring, Gelas, Retur, dll</p>
          </div>
        </div>
        <div class="flex items-center gap-2 flex-wrap">
          <!-- Filter Company -->
          <select
            v-model="filterCompanyId"
            @change="fetchGoods"
            class="px-3 py-2 text-sm border border-slate-200 rounded-lg bg-white text-slate-700 focus:ring-2 focus:ring-primary/20 outline-none"
          >
            <option value="">Semua Perusahaan</option>
            <option v-for="c in companies" :key="c.id" :value="c.id">{{ c.name }}</option>
          </select>
          <!-- Toggle Inactive -->
          <label class="flex items-center gap-2 text-sm text-slate-600 cursor-pointer select-none">
            <input type="checkbox" v-model="includeInactive" class="rounded accent-primary" />
            Tampilkan non-aktif
          </label>
          <button
            @click="fetchGoods"
            class="p-2 rounded-lg hover:bg-slate-100 text-slate-500 transition"
            title="Refresh"
          >
            <RefreshCw :size="16" />
          </button>
          <button
            @click="openCreate"
            class="flex items-center gap-2 px-4 py-2 bg-primary hover:bg-primary/90 text-white text-sm font-semibold rounded-xl shadow-md hover:shadow-lg hover:opacity-90 transition-all"
          >
            <Plus :size="16" />
            Tambah Barang
          </button>
        </div>
      </div>
    </div>

    <!-- Body -->
    <div class="max-w-7xl mx-auto px-6 py-6 space-y-6">
      <!-- Loading -->
      <div v-if="isLoading" class="flex items-center justify-center py-20">
        <div class="w-10 h-10 border-4 border-primary border-t-transparent rounded-full animate-spin" />
      </div>

      <template v-else>
        <!-- Summary Cards -->
        <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
          <div class="bg-white rounded-2xl p-4 border border-slate-100 shadow-sm flex items-center gap-3">
            <div class="w-10 h-10 rounded-xl bg-primary/10 flex items-center justify-center">
              <Package class="text-primary" :size="18" />
            </div>
            <div>
              <p class="text-xs text-slate-400">Total Barang</p>
              <p class="text-xl font-bold text-slate-800">{{ filteredGoods.length }}</p>
            </div>
          </div>
          <div class="bg-white rounded-2xl p-4 border border-slate-100 shadow-sm flex items-center gap-3">
            <div class="w-10 h-10 rounded-xl bg-emerald-100 flex items-center justify-center">
              <Package class="text-emerald-600" :size="18" />
            </div>
            <div>
              <p class="text-xs text-slate-400">Barang Jual</p>
              <p class="text-xl font-bold text-slate-800">{{ regularGoods.length }}</p>
            </div>
          </div>
          <div class="bg-white rounded-2xl p-4 border border-slate-100 shadow-sm flex items-center gap-3">
            <div class="w-10 h-10 rounded-xl bg-rose-100 flex items-center justify-center">
              <RotateCcw class="text-rose-500" :size="18" />
            </div>
            <div>
              <p class="text-xs text-slate-400">Item Retur</p>
              <p class="text-xl font-bold text-slate-800">{{ returnGoods.length }}</p>
            </div>
          </div>
          <div class="bg-white rounded-2xl p-4 border border-slate-100 shadow-sm flex items-center gap-3">
            <div class="w-10 h-10 rounded-xl bg-slate-100 flex items-center justify-center">
              <Tags class="text-slate-500" :size="18" />
            </div>
            <div>
              <p class="text-xs text-slate-400">Aktif</p>
              <p class="text-xl font-bold text-slate-800">{{ filteredGoods.filter((g) => g.is_active).length }}</p>
            </div>
          </div>
        </div>

        <!-- Barang Jual -->
        <div class="bg-white rounded-2xl shadow-sm border border-slate-100 overflow-hidden">
          <div class="px-6 py-4 border-b border-slate-100 flex items-center gap-2">
            <span class="w-2 h-2 rounded-full bg-emerald-500" />
            <h2 class="font-semibold text-slate-700">Barang Jual</h2>
            <span class="ml-auto text-xs text-slate-400">(harga menambah tagihan)</span>
          </div>
          <div class="overflow-x-auto">
            <table class="w-full text-sm">
              <thead class="bg-slate-50 text-slate-500 uppercase text-xs">
                <tr>
                  <th class="text-left px-6 py-3">Nama Barang</th>
                  <th class="text-left px-6 py-3">Satuan</th>
                  <th class="text-right px-6 py-3">Harga</th>
                  <th class="text-right px-6 py-3">Stok Pusat</th>
                  <th class="text-center px-6 py-3">Urutan</th>
                  <th class="text-center px-6 py-3">Status</th>
                  <th class="text-center px-6 py-3">Aksi</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-slate-50">
                <tr v-if="regularGoods.length === 0">
                  <td colspan="6" class="text-center py-10 text-slate-400">Belum ada data barang jual</td>
                </tr>
                <tr
                  v-for="good in regularGoods"
                  :key="good.id"
                  :class="['hover:bg-slate-50 transition', !good.is_active ? 'opacity-50' : '']"
                >
                  <td class="px-6 py-4">
                    <div class="font-medium text-slate-800">{{ good.name }}</div>
                    <div v-if="good.description" class="text-xs text-slate-400">{{ good.description }}</div>
                  </td>
                  <td class="px-6 py-4 text-slate-600">{{ good.unit }}</td>
                  <td class="px-6 py-4 text-right font-semibold text-emerald-700">
                    {{ currency.format(good.price) }}
                  </td>
                  <td class="px-6 py-4 text-right">
                    <span class="font-bold text-slate-800">{{ good.stock }}</span>
                  </td>
                  <td class="px-6 py-4 text-center text-slate-500">{{ good.sort_order }}</td>
                  <td class="px-6 py-4 text-center">
                    <span
                      :class="[
                        'px-2 py-1 rounded-full text-xs font-semibold',
                        good.is_active ? 'bg-emerald-100 text-emerald-700' : 'bg-slate-100 text-slate-500',
                      ]"
                    >
                      {{ good.is_active ? 'Aktif' : 'Non-aktif' }}
                    </span>
                  </td>
                  <td class="px-6 py-4">
                    <div class="flex items-center justify-center gap-2">
                      <button
                        @click="openAdjust(good)"
                        class="px-2 py-1.5 text-xs font-bold text-primary bg-primary/5 hover:bg-primary/10 rounded-lg transition-colors"
                      >
                        Stok
                      </button>
                      <button
                        @click="openEdit(good)"
                        class="p-2 rounded-lg hover:bg-slate-100 text-slate-500 transition"
                      >
                        <Pencil :size="15" />
                      </button>
                      <button
                        @click="openDelete(good.id)"
                        class="p-2 rounded-lg hover:bg-rose-50 text-rose-500 transition"
                      >
                        <Trash2 :size="15" />
                      </button>
                    </div>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>

        <!-- Barang Retur -->
        <div class="bg-white rounded-2xl shadow-sm border border-slate-100 overflow-hidden">
          <div class="px-6 py-4 border-b border-slate-100 flex items-center gap-2">
            <span class="w-2 h-2 rounded-full bg-rose-500" />
            <h2 class="font-semibold text-slate-700">Retur / Kredit</h2>
            <span class="ml-auto text-xs text-slate-400">(harga mengurangi tagihan)</span>
          </div>
          <div class="overflow-x-auto">
            <table class="w-full text-sm">
              <thead class="bg-slate-50 text-slate-500 uppercase text-xs">
                <tr>
                  <th class="text-left px-6 py-3">Nama Barang</th>
                  <th class="text-left px-6 py-3">Satuan</th>
                  <th class="text-right px-6 py-3">Nilai Kredit</th>
                  <th class="text-right px-6 py-3">Stok Pusat</th>
                  <th class="text-center px-6 py-3">Urutan</th>
                  <th class="text-center px-6 py-3">Status</th>
                  <th class="text-center px-6 py-3">Aksi</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-slate-50">
                <tr v-if="returnGoods.length === 0">
                  <td colspan="6" class="text-center py-10 text-slate-400">Belum ada item retur</td>
                </tr>
                <tr
                  v-for="good in returnGoods"
                  :key="good.id"
                  :class="['hover:bg-slate-50 transition', !good.is_active ? 'opacity-50' : '']"
                >
                  <td class="px-6 py-4">
                    <div class="font-medium text-slate-800 flex items-center gap-2">
                      <RotateCcw class="text-rose-400" :size="14" />
                      {{ good.name }}
                    </div>
                    <div v-if="good.description" class="text-xs text-slate-400">{{ good.description }}</div>
                  </td>
                  <td class="px-6 py-4 text-slate-600">{{ good.unit }}</td>
                  <td class="px-6 py-4 text-right font-semibold text-rose-600">
                    -{{ currency.format(good.price) }}
                  </td>
                  <td class="px-6 py-4 text-right">
                    <span class="font-bold text-slate-800">{{ good.stock }}</span>
                  </td>
                  <td class="px-6 py-4 text-center text-slate-500">{{ good.sort_order }}</td>
                  <td class="px-6 py-4 text-center">
                    <span
                      :class="[
                        'px-2 py-1 rounded-full text-xs font-semibold',
                        good.is_active ? 'bg-emerald-100 text-emerald-700' : 'bg-slate-100 text-slate-500',
                      ]"
                    >
                      {{ good.is_active ? 'Aktif' : 'Non-aktif' }}
                    </span>
                  </td>
                  <td class="px-6 py-4">
                    <div class="flex items-center justify-center gap-2">
                      <button
                        @click="openAdjust(good)"
                        class="px-2 py-1.5 text-xs font-bold text-primary bg-primary/5 hover:bg-primary/10 rounded-lg transition-colors"
                      >
                        Stok
                      </button>
                      <button
                        @click="openEdit(good)"
                        class="p-2 rounded-lg hover:bg-slate-100 text-slate-500 transition"
                      >
                        <Pencil :size="15" />
                      </button>
                      <button
                        @click="openDelete(good.id)"
                        class="p-2 rounded-lg hover:bg-rose-50 text-rose-500 transition"
                      >
                        <Trash2 :size="15" />
                      </button>
                    </div>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </template>
    </div>

    <!-- ── Create/Edit Modal ─────────────────────────────────────────────────── -->
    <Teleport to="body">
      <div
        v-if="showModal"
        class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/40 backdrop-blur-sm"
        @click.self="closeModal"
      >
        <div class="bg-white rounded-2xl shadow-2xl w-full max-w-md">
          <div class="p-6 border-b border-slate-100 flex items-center justify-between">
            <h2 class="text-lg font-bold text-slate-800">
              {{ isEditing ? 'Edit Barang' : 'Tambah Barang Sampingan' }}
            </h2>
            <button @click="closeModal" class="text-slate-400 hover:text-slate-600 text-xl leading-none">&times;</button>
          </div>
          <div class="p-6 space-y-4">
            <!-- Branch / Company -->
            <div v-if="!isEditing">
              <label class="block text-xs font-semibold text-slate-500 mb-1">Cabang *</label>
              <select
                v-model="form.factory_id"
                class="w-full px-3 py-2 border border-slate-200 rounded-lg text-sm focus:ring-2 focus:ring-primary/20 outline-none"
              >
                <option value="">-- Pilih Cabang --</option>
                <option v-for="b in branches" :key="b.id" :value="b.id">{{ b.name }}</option>
              </select>
            </div>
            <div v-else>
              <label class="block text-xs font-semibold text-slate-500 mb-1">Perusahaan</label>
              <select
                v-model="form.company_id"
                disabled
                class="w-full px-3 py-2 border border-slate-200 rounded-lg text-sm bg-slate-50 text-slate-500 outline-none"
              >
                <option v-for="c in companies" :key="c.id" :value="c.id">{{ c.name }}</option>
              </select>
            </div>
            <!-- Name -->
            <div>
              <label class="block text-xs font-semibold text-slate-500 mb-1">Nama Barang *</label>
              <input
                v-model="form.name"
                type="text"
                placeholder="Contoh: Krat, Botol Kosong, Galon..."
                class="w-full px-3 py-2 border border-slate-200 rounded-lg text-sm focus:ring-2 focus:ring-primary/20 outline-none"
              />
            </div>
            <!-- Price + Unit -->
            <div class="grid grid-cols-2 gap-3">
              <div>
                <label class="block text-xs font-semibold text-slate-500 mb-1">Harga (Rp) *</label>
                <input
                  v-model.number="form.price"
                  type="number"
                  min="0"
                  placeholder="10000"
                  class="w-full px-3 py-2 border border-slate-200 rounded-lg text-sm focus:ring-2 focus:ring-primary/20 outline-none"
                />
              </div>
              <div>
                <label class="block text-xs font-semibold text-slate-500 mb-1">Satuan</label>
                <input
                  v-model="form.unit"
                  type="text"
                  placeholder="PCS / KRT / BTL"
                  class="w-full px-3 py-2 border border-slate-200 rounded-lg text-sm focus:ring-2 focus:ring-primary/20 outline-none"
                />
              </div>
            </div>
            <!-- Initial Stock (Only for Create) -->
            <div v-if="!isEditing">
              <label class="block text-xs font-semibold text-slate-500 mb-1">Stok Awal</label>
              <input
                v-model.number="form.initial_stock"
                type="number"
                min="0"
                placeholder="0"
                class="w-full px-3 py-2 border border-slate-200 rounded-lg text-sm focus:ring-2 focus:ring-primary/20 outline-none"
              />
            </div>
            <!-- Flags -->
            <div class="flex flex-col gap-3 mt-2">
              <label class="flex items-center gap-2 text-sm cursor-pointer select-none">
                <input type="checkbox" v-model="form.is_return" class="w-4 h-4 rounded accent-rose-500" />
                <span class="text-slate-600">Barang <strong class="text-rose-600">retur/kredit</strong> (mengurangi tagihan)</span>
              </label>
              <label class="flex items-center gap-2 text-sm cursor-pointer select-none">
                <input type="checkbox" v-model="form.is_active" class="w-4 h-4 rounded accent-emerald-500" />
                <span class="text-slate-600">Aktif (tampil di aplikasi)</span>
              </label>
            </div>
          </div>
          <div class="p-6 border-t border-slate-100 flex justify-end gap-3">
            <button
              @click="closeModal"
              class="px-4 py-2 text-sm border border-slate-200 rounded-lg hover:bg-slate-50 text-slate-600 transition"
            >
              Batal
            </button>
            <button
              @click="saveGood"
              :disabled="isSaving"
              class="px-5 py-2 text-sm font-semibold bg-primary hover:bg-primary/90 text-white rounded-xl shadow hover:opacity-90 transition disabled:opacity-60"
            >
              {{ isSaving ? 'Menyimpan...' : (isEditing ? 'Simpan Perubahan' : 'Tambahkan') }}
            </button>
          </div>
        </div>
      </div>

      <!-- Delete Confirm -->
      <div
        v-if="showDeleteConfirm"
        class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/40 backdrop-blur-sm"
        @click.self="showDeleteConfirm = false"
      >
        <div class="bg-white rounded-2xl shadow-2xl w-full max-w-sm p-6">
          <div class="flex items-center gap-3 mb-4">
            <div class="w-10 h-10 rounded-xl bg-rose-100 flex items-center justify-center">
              <Trash2 class="text-rose-500" :size="20" />
            </div>
            <h3 class="font-bold text-slate-800">Hapus Barang?</h3>
          </div>
          <p class="text-sm text-slate-500 mb-6">
            Barang yang sudah dihapus tidak dapat dipulihkan. Pastikan barang ini tidak digunakan dalam transaksi aktif.
          </p>
          <div class="flex gap-3 justify-end">
            <button
              @click="showDeleteConfirm = false"
              class="px-4 py-2 text-sm border border-slate-200 rounded-lg hover:bg-slate-50 text-slate-600 transition"
            >
              Batal
            </button>
            <button
              @click="confirmDelete"
              :disabled="isDeleting"
              class="px-4 py-2 text-sm font-semibold bg-rose-600 text-white rounded-lg hover:bg-rose-700 transition disabled:opacity-60"
            >
              {{ isDeleting ? 'Menghapus...' : 'Ya, Hapus' }}
            </button>
          </div>
        </div>
      </div>

      <!-- Adjust Stock Modal -->
      <div v-if="showAdjust" class="fixed inset-0 bg-slate-900/40 backdrop-blur-sm z-50 flex items-center justify-center p-4" @click.self="showAdjust = false">
        <div class="w-full max-w-md bg-white rounded-2xl overflow-hidden shadow-2xl animate-in zoom-in-95 duration-200">
          <div class="p-6 border-b border-slate-100 flex items-center justify-between">
            <div>
              <h2 class="text-lg font-bold text-slate-800">Sesuaikan Stok</h2>
              <p class="text-sm text-slate-500">{{ targetAdjustGood?.name }} (Stok saat ini: {{ targetAdjustGood?.stock }})</p>
            </div>
            <button @click="showAdjust = false" class="text-slate-400 hover:text-slate-600 text-xl leading-none">&times;</button>
          </div>
          <div class="p-6 space-y-4">
            <div>
              <label class="block text-sm font-bold text-slate-700 mb-2">Cabang (Gudang)</label>
              <select
                v-model="adjustForm.factory_id"
                class="w-full px-4 py-3 bg-slate-50 border border-slate-200 rounded-xl outline-none focus:ring-2 focus:ring-primary/20"
              >
                <option value="">-- Pilih Cabang --</option>
                <option v-for="b in branches" :key="b.id" :value="b.id">{{ b.name }}</option>
              </select>
            </div>
            <div>
              <label class="block text-sm font-bold text-slate-700 mb-2">Penambahan/Pengurangan (Bisa minus)</label>
              <input type="number" v-model.number="adjustForm.quantity" class="w-full px-4 py-3 bg-slate-50 border border-slate-200 rounded-xl outline-none focus:ring-2 focus:ring-primary/20" placeholder="Contoh: 10 atau -5" />
            </div>
            <div>
              <label class="block text-sm font-bold text-slate-700 mb-2">Alasan Penyesuaian</label>
              <input type="text" v-model="adjustForm.reason" class="w-full px-4 py-3 bg-slate-50 border border-slate-200 rounded-xl outline-none focus:ring-2 focus:ring-primary/20" placeholder="Misal: Stok opname, barang rusak" />
            </div>
          </div>
          <div class="p-6 border-t border-slate-100 bg-slate-50 flex justify-end gap-3">
            <button @click="showAdjust = false" class="px-4 py-2 text-sm font-bold text-slate-600 hover:bg-slate-200 rounded-xl transition">Batal</button>
            <button @click="submitAdjust" :disabled="isSubmitting" class="px-6 py-2 text-sm font-bold text-white bg-primary hover:bg-primary/90 rounded-xl shadow-md disabled:opacity-50 transition">
              {{ isSubmitting ? 'Menyimpan...' : 'Simpan' }}
            </button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>
