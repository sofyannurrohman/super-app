<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { useMarketplaceInventoryStore } from '@/stores/marketplaceInventory'
import {
  Boxes,
  Plus,
  Pencil,
  Trash2,
  SlidersHorizontal,
  RefreshCw,
  Search,
  CheckCircle2,
  AlertTriangle,
  CircleOff,
  ArrowDownToLine,
  ArrowUpFromLine,
  Image as ImageIcon,
  X
} from 'lucide-vue-next'

const store = useMarketplaceInventoryStore()

const searchQuery = ref('')
const showAddModal = ref(false)
const showEditModal = ref(false)
const showOpnameModal = ref(false)
const showInOutModal = ref(false)
const showDeleteConfirm = ref(false)
const selectedInventory = ref<any>(null)
const inOutMode = ref<'IN' | 'OUT'>('IN')

const form = ref<{
  name: string
  sku: string
  unit: string
  min_stock: number
  notes: string
  imageFile: File | null
  imagePreview: string | null
}>({
  name: '',
  sku: '',
  unit: 'Pcs',
  min_stock: 10,
  notes: '',
  imageFile: null,
  imagePreview: null,
})

const opnameForm = ref({
  quantity: 0,
  notes: '',
})

const inOutForm = ref({
  quantity: 1,
  reference: '',
  notes: '',
})

let pollingInterval: ReturnType<typeof setInterval> | null = null

async function loadData() {
  await store.fetchInventories()
}

onMounted(async () => {
  await loadData()
  pollingInterval = setInterval(() => {
    loadData()
  }, 15000)
})

onUnmounted(() => {
  if (pollingInterval) clearInterval(pollingInterval)
})

const filteredInventories = computed(() => {
  let result = store.inventories
  if (searchQuery.value.trim()) {
    const q = searchQuery.value.toLowerCase()
    result = result.filter(
      (s: any) =>
        s.name.toLowerCase().includes(q) || s.sku?.toLowerCase().includes(q),
    )
  }
  return result
})

function getStatus(stock: any) {
  if (stock.quantity <= 0) return 'empty'
  if (stock.quantity <= stock.min_stock) return 'critical'
  return 'safe'
}

function statusConfig(stock: any) {
  const s = getStatus(stock)
  if (s === 'safe') return { label: 'Aman', cls: 'bg-emerald-100 text-emerald-700 border-emerald-200' }
  if (s === 'critical') return { label: 'Kritis', cls: 'bg-amber-100 text-amber-700 border-amber-200' }
  return { label: 'Habis', cls: 'bg-red-100 text-red-700 border-red-200' }
}

function openAdd() {
  form.value = { name: '', sku: '', unit: 'Pcs', min_stock: 10, notes: '', imageFile: null, imagePreview: null }
  showAddModal.value = true
}

function openEdit(item: any) {
  selectedInventory.value = item
  form.value = {
    name: item.name,
    sku: item.sku || '',
    unit: item.unit,
    min_stock: item.min_stock,
    notes: item.notes || '',
    imageFile: null,
    imagePreview: item.image_url ? getImageUrl(item.image_url) : null,
  }
  showEditModal.value = true
}

function openInOut(item: any, mode: 'IN' | 'OUT') {
  selectedInventory.value = item
  inOutMode.value = mode
  inOutForm.value = { quantity: 1, reference: '', notes: '' }
  showInOutModal.value = true
}

function openOpname(item: any) {
  selectedInventory.value = item
  opnameForm.value = { quantity: item.quantity, notes: '' }
  showOpnameModal.value = true
}

function openDelete(item: any) {
  selectedInventory.value = item
  showDeleteConfirm.value = true
}

async function submitAdd() {
  try {
    const fd = new FormData()
    fd.append('name', form.value.name)
    fd.append('sku', form.value.sku)
    fd.append('unit', form.value.unit)
    fd.append('min_stock', form.value.min_stock.toString())
    fd.append('notes', form.value.notes)
    if (form.value.imageFile) fd.append('image', form.value.imageFile)
    
    await store.createInventory(fd)
    showAddModal.value = false
  } catch (e: any) {
    alert(e.response?.data?.error || 'Gagal menambah data')
  }
}

async function submitEdit() {
  if (!selectedInventory.value) return
  try {
    const fd = new FormData()
    fd.append('name', form.value.name)
    fd.append('sku', form.value.sku)
    fd.append('unit', form.value.unit)
    fd.append('min_stock', form.value.min_stock.toString())
    fd.append('notes', form.value.notes)
    if (form.value.imageFile) fd.append('image', form.value.imageFile)

    await store.updateInventory(selectedInventory.value.id, fd)
    showEditModal.value = false
  } catch (e: any) {
    alert(e.response?.data?.error || 'Gagal memperbarui data')
  }
}

function handleImageUpload(e: Event) {
  const target = e.target as HTMLInputElement
  if (target.files && target.files[0]) {
    const file = target.files[0]
    if (file.size > 5 * 1024 * 1024) {
      alert('Ukuran file maksimal 5MB')
      target.value = ''
      return
    }
    form.value.imageFile = file
    form.value.imagePreview = URL.createObjectURL(file)
  }
}

function getImageUrl(url: string) {
  if (!url) return ''
  if (url.startsWith('http')) return url
  const baseUrl = import.meta.env.VITE_API_BASE_URL?.replace('/api', '') || ''
  return baseUrl + url
}

async function submitInOut() {
  if (!selectedInventory.value) return
  try {
    if (inOutMode.value === 'IN') {
      await store.stockIn(selectedInventory.value.id, inOutForm.value)
    } else {
      await store.stockOut(selectedInventory.value.id, inOutForm.value)
    }
    showInOutModal.value = false
  } catch (e: any) {
    alert(e.response?.data?.error || 'Gagal memproses stok')
  }
}

async function submitOpname() {
  if (!selectedInventory.value) return
  try {
    await store.doOpname(selectedInventory.value.id, opnameForm.value)
    showOpnameModal.value = false
  } catch (e: any) {
    alert(e.response?.data?.error || 'Gagal melakukan opname')
  }
}

async function confirmDelete() {
  if (!selectedInventory.value) return
  try {
    await store.deleteInventory(selectedInventory.value.id)
    showDeleteConfirm.value = false
  } catch (e: any) {
    alert(e.response?.data?.error || 'Gagal menghapus data')
  }
}

async function manualRefresh() {
  await store.fetchInventories()
}

const lastUpdatedFormatted = computed(() => {
  if (!store.lastUpdated) return '-'
  return store.lastUpdated.toLocaleTimeString('id-ID', { hour: '2-digit', minute: '2-digit', second: '2-digit' })
})
</script>

<template>
  <div class="space-y-6 animate-in fade-in duration-700">
    <div class="flex flex-col md:flex-row justify-between items-start md:items-center gap-4">
      <div>
        <h1 class="text-2xl font-black tracking-tight text-slate-900 flex items-center gap-3">
          <div class="p-2 rounded-xl bg-gradient-to-br from-red-500 to-red-600 shadow-md shadow-red-500/20">
            <Boxes class="h-6 w-6 text-white" />
          </div>
          Inventaris & ATK
        </h1>
        <p class="text-slate-500 text-sm font-medium mt-1">
          Kelola barang inventaris seperti kardus, lakban, alat tulis untuk kebutuhan marketplace
        </p>
      </div>
      <div class="flex items-center gap-3 flex-wrap">
        <div class="flex items-center gap-2 bg-emerald-50 border border-emerald-200 px-3 py-1.5 rounded-xl text-xs font-bold text-emerald-700">
          <span class="relative flex h-2 w-2">
            <span class="animate-ping absolute inline-flex h-full w-full rounded-full bg-emerald-400 opacity-75"></span>
            <span class="relative inline-flex rounded-full h-2 w-2 bg-emerald-500"></span>
          </span>
          LIVE · Diperbarui {{ lastUpdatedFormatted }}
        </div>
        <button
          @click="manualRefresh"
          class="p-2 rounded-xl border border-slate-200 text-slate-500 hover:text-red-600 hover:border-red-600/30 transition-all shadow-sm"
        >
          <RefreshCw class="h-4 w-4" :class="{ 'animate-spin': store.isLoading }" />
        </button>
        <button
          @click="openAdd"
          class="flex items-center px-4 py-2 bg-gradient-to-r from-red-600 to-red-500 text-white hover:from-red-700 hover:to-red-600 shadow-md shadow-red-500/20 rounded-xl font-bold gap-2 text-sm"
        >
          <Plus class="h-4 w-4" />
          Tambah Barang
        </button>
      </div>
    </div>

    <!-- Filters -->
    <div class="bg-white p-4 flex flex-col md:flex-row gap-3 items-center rounded-2xl shadow-sm">
      <div class="relative flex-1 w-full md:max-w-xs">
        <Search class="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-slate-400" />
        <input
          v-model="searchQuery"
          type="text"
          placeholder="Cari nama barang..."
          class="w-full pl-9 pr-4 py-2.5 text-sm font-medium bg-slate-50 border border-slate-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-red-500/20 focus:border-red-500/50 transition-all"
        />
      </div>
      <div class="md:ml-auto text-xs font-bold text-slate-400">
        {{ filteredInventories.length }} barang
      </div>
    </div>

    <!-- Table -->
    <div class="bg-white rounded-2xl shadow-sm overflow-hidden">
      <div v-if="store.isLoading && filteredInventories.length === 0" class="p-16 text-center">
        <div class="animate-spin h-8 w-8 border-2 border-red-600 border-t-transparent rounded-full mx-auto mb-3"></div>
      </div>
      <div v-else-if="filteredInventories.length === 0" class="p-16 text-center">
        <Boxes class="h-12 w-12 text-slate-200 mx-auto mb-3" />
        <p class="text-slate-400 font-bold">Belum ada barang</p>
      </div>
      <div v-else class="overflow-x-auto">
        <table class="w-full text-sm">
          <thead>
            <tr class="bg-slate-50 border-b border-slate-100">
              <th class="text-left px-4 py-3.5 text-[11px] font-black uppercase tracking-widest text-slate-400">Nama Barang</th>
              <th class="text-left px-4 py-3.5 text-[11px] font-black uppercase tracking-widest text-slate-400">SKU</th>
              <th class="text-right px-4 py-3.5 text-[11px] font-black uppercase tracking-widest text-slate-400">Stok</th>
              <th class="text-right px-4 py-3.5 text-[11px] font-black uppercase tracking-widest text-slate-400">Min Stok</th>
              <th class="text-center px-4 py-3.5 text-[11px] font-black uppercase tracking-widest text-slate-400">Status</th>
              <th class="text-center px-4 py-3.5 text-[11px] font-black uppercase tracking-widest text-slate-400">Transaksi</th>
              <th class="text-center px-4 py-3.5 text-[11px] font-black uppercase tracking-widest text-slate-400">Aksi</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-50">
            <tr
              v-for="item in filteredInventories"
              :key="item.id"
              class="hover:bg-red-50/50 transition-colors group"
            >
              <td class="px-4 py-3.5">
                <div class="flex items-center gap-3">
                  <div class="flex-shrink-0 w-10 h-10 rounded-lg overflow-hidden border border-slate-200 bg-slate-50 flex items-center justify-center">
                    <img v-if="item.image_url" :src="getImageUrl(item.image_url)" class="w-full h-full object-cover" />
                    <ImageIcon v-else class="h-5 w-5 text-slate-300" />
                  </div>
                  <div>
                    <p class="font-bold text-slate-900 text-sm">{{ item.name }}</p>
                  </div>
                </div>
              </td>
              <td class="px-4 py-3.5 text-left text-sm font-medium text-slate-600">
                {{ item.sku || '-' }}
              </td>
              <td class="px-4 py-3.5 text-right">
                <div class="font-black text-xl text-slate-900">
                  {{ item.quantity.toLocaleString('id-ID') }} <span class="text-xs font-bold text-slate-400">{{ item.unit }}</span>
                </div>
              </td>
              <td class="px-4 py-3.5 text-right">
                <div class="font-bold text-slate-600">
                  {{ item.min_stock.toLocaleString('id-ID') }} <span class="text-xs text-slate-400">{{ item.unit }}</span>
                </div>
              </td>
              <td class="px-4 py-3.5 text-center">
                <span class="text-xs font-black px-2.5 py-1 rounded-full border" :class="statusConfig(item).cls">
                  {{ statusConfig(item).label }}
                </span>
              </td>
              <td class="px-4 py-3.5">
                <div class="flex items-center justify-center gap-2">
                  <button @click="openInOut(item, 'IN')" class="flex items-center gap-1 text-xs font-bold px-2 py-1 bg-emerald-50 text-emerald-600 rounded border border-emerald-100 hover:bg-emerald-100">
                    <ArrowDownToLine class="w-3 h-3" /> Masuk
                  </button>
                  <button @click="openInOut(item, 'OUT')" class="flex items-center gap-1 text-xs font-bold px-2 py-1 bg-amber-50 text-amber-600 rounded border border-amber-100 hover:bg-amber-100">
                    <ArrowUpFromLine class="w-3 h-3" /> Keluar
                  </button>
                  <button @click="openOpname(item)" class="flex items-center gap-1 text-xs font-bold px-2 py-1 bg-red-50 text-red-600 rounded border border-red-100 hover:bg-red-100">
                    <SlidersHorizontal class="w-3 h-3" /> Opname
                  </button>
                </div>
              </td>
              <td class="px-4 py-3.5">
                <div class="flex items-center justify-center gap-1.5">
                  <button @click="openEdit(item)" class="p-2 rounded-lg text-slate-400 hover:text-red-600 hover:bg-red-50 transition-colors">
                    <Pencil class="h-4 w-4" />
                  </button>
                  <button @click="openDelete(item)" class="p-2 rounded-lg text-slate-400 hover:text-red-600 hover:bg-red-50 transition-colors">
                    <Trash2 class="h-4 w-4" />
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Modals -->
    <!-- Add/Edit Modal -->
    <Teleport to="body">
      <div v-if="showAddModal || showEditModal" class="fixed inset-0 z-50 flex items-center justify-center p-4">
        <div class="absolute inset-0 bg-slate-900/50 backdrop-blur-sm" @click="showAddModal = false; showEditModal = false" />
        <div class="relative bg-white rounded-3xl shadow-2xl w-full max-w-sm overflow-hidden animate-in zoom-in-95 duration-200">
          <div class="bg-gradient-to-r from-red-600 to-red-500 p-5 text-white">
            <h2 class="text-lg font-black">{{ showEditModal ? 'Edit Barang' : 'Tambah Barang' }}</h2>
          </div>
          <form @submit.prevent="showEditModal ? submitEdit() : submitAdd()" class="p-6 space-y-4">
            <div>
              <label class="block text-xs font-black uppercase text-slate-500 mb-1.5">Nama Barang *</label>
              <input v-model="form.name" required class="w-full px-3 py-2.5 text-sm border border-slate-200 rounded-xl focus:ring-2 focus:ring-red-500/30" />
            </div>
            <div>
              <label class="block text-xs font-black uppercase text-slate-500 mb-1.5">SKU</label>
              <input v-model="form.sku" class="w-full px-3 py-2.5 text-sm border border-slate-200 rounded-xl focus:ring-2 focus:ring-red-500/30" />
            </div>
            <div class="grid grid-cols-2 gap-3">
              <div>
                <label class="block text-xs font-black uppercase text-slate-500 mb-1.5">Satuan *</label>
                <input v-model="form.unit" required class="w-full px-3 py-2.5 text-sm border border-slate-200 rounded-xl focus:ring-2 focus:ring-red-500/30" placeholder="Pcs/Box/Roll" />
              </div>
              <div>
                <label class="block text-xs font-black uppercase text-slate-500 mb-1.5">Min Stok</label>
                <input v-model.number="form.min_stock" type="number" required class="w-full px-3 py-2.5 text-sm border border-slate-200 rounded-xl focus:ring-2 focus:ring-red-500/30" />
              </div>
            </div>
            <div>
              <label class="block text-xs font-black uppercase text-slate-500 mb-1.5">Foto Barang (Max 5MB)</label>
              <div v-if="form.imagePreview" class="mb-2 relative inline-block">
                <img :src="form.imagePreview" class="w-24 h-24 object-cover rounded-xl border border-slate-200" />
                <button type="button" @click="form.imageFile = null; form.imagePreview = null" class="absolute -top-2 -right-2 bg-red-500 text-white rounded-full p-1 shadow-sm">
                  <X class="w-3 h-3" />
                </button>
              </div>
              <input type="file" accept="image/*" @change="handleImageUpload" class="w-full text-sm text-slate-500 file:mr-4 file:py-2 file:px-4 file:rounded-full file:border-0 file:text-sm file:font-bold file:bg-red-50 file:text-red-700 hover:file:bg-red-100 transition-colors" />
            </div>
            <div>
              <label class="block text-xs font-black uppercase text-slate-500 mb-1.5">Catatan</label>
              <textarea v-model="form.notes" class="w-full px-3 py-2.5 text-sm border border-slate-200 rounded-xl focus:ring-2 focus:ring-red-500/30" rows="2"></textarea>
            </div>
            <div class="flex gap-3 pt-2">
              <button type="button" @click="showAddModal = false; showEditModal = false" class="flex-1 px-4 py-2.5 text-sm font-bold border rounded-xl hover:bg-slate-50">Batal</button>
              <button type="submit" class="flex-1 px-4 py-2.5 text-sm font-bold bg-red-600 text-white rounded-xl hover:bg-red-700">Simpan</button>
            </div>
          </form>
        </div>
      </div>
    </Teleport>

    <!-- In/Out Modal -->
    <Teleport to="body">
      <div v-if="showInOutModal" class="fixed inset-0 z-50 flex items-center justify-center p-4">
        <div class="absolute inset-0 bg-slate-900/50 backdrop-blur-sm" @click="showInOutModal = false" />
        <div class="relative bg-white rounded-3xl shadow-2xl w-full max-w-sm overflow-hidden animate-in zoom-in-95 duration-200">
          <div class="p-5 text-white" :class="inOutMode === 'IN' ? 'bg-emerald-600' : 'bg-amber-500'">
            <h2 class="text-lg font-black">{{ inOutMode === 'IN' ? 'Stok Masuk' : 'Stok Keluar' }}</h2>
            <p class="text-sm opacity-90">{{ selectedInventory?.name }}</p>
          </div>
          <form @submit.prevent="submitInOut" class="p-6 space-y-4">
            <div>
              <label class="block text-xs font-black uppercase text-slate-500 mb-1.5">Jumlah *</label>
              <input v-model.number="inOutForm.quantity" type="number" min="1" required class="w-full px-3 py-3 text-2xl font-black text-center border-2 border-slate-200 rounded-xl focus:ring-2 focus:ring-emerald-500/30" />
            </div>
            <div>
              <label class="block text-xs font-black uppercase text-slate-500 mb-1.5">No. Referensi</label>
              <input v-model="inOutForm.reference" class="w-full px-3 py-2.5 text-sm border border-slate-200 rounded-xl" placeholder="Opsional" />
            </div>
            <div>
              <label class="block text-xs font-black uppercase text-slate-500 mb-1.5">Catatan</label>
              <textarea v-model="inOutForm.notes" class="w-full px-3 py-2.5 text-sm border border-slate-200 rounded-xl" rows="2"></textarea>
            </div>
            <div class="flex gap-3 pt-2">
              <button type="button" @click="showInOutModal = false" class="flex-1 px-4 py-2.5 text-sm font-bold border rounded-xl hover:bg-slate-50">Batal</button>
              <button type="submit" :disabled="store.isActionLoading" class="flex-1 px-4 py-2.5 text-sm font-bold text-white rounded-xl shadow-md" :class="inOutMode === 'IN' ? 'bg-emerald-600' : 'bg-amber-500'">Proses</button>
            </div>
          </form>
        </div>
      </div>
    </Teleport>

    <!-- Opname Modal -->
    <Teleport to="body">
      <div v-if="showOpnameModal" class="fixed inset-0 z-50 flex items-center justify-center p-4">
        <div class="absolute inset-0 bg-slate-900/50 backdrop-blur-sm" @click="showOpnameModal = false" />
        <div class="relative bg-white rounded-3xl shadow-2xl w-full max-w-sm overflow-hidden animate-in zoom-in-95 duration-200">
          <div class="p-5 text-white bg-red-600">
            <h2 class="text-lg font-black">Stok Opname</h2>
            <p class="text-sm opacity-90">{{ selectedInventory?.name }}</p>
          </div>
          <form @submit.prevent="submitOpname" class="p-6 space-y-4">
            <div class="flex justify-between items-center bg-slate-50 p-3 rounded-xl border border-slate-100">
              <span class="text-xs font-bold text-slate-500 uppercase">Stok Sistem</span>
              <span class="text-lg font-black text-slate-800">{{ selectedInventory?.quantity }}</span>
            </div>
            <div>
              <label class="block text-xs font-black uppercase text-slate-500 mb-1.5">Stok Fisik Aktual *</label>
              <input v-model.number="opnameForm.quantity" type="number" min="0" required class="w-full px-3 py-3 text-2xl font-black text-center border-2 border-slate-200 rounded-xl focus:ring-2 focus:ring-red-500/30" />
            </div>
            <div>
              <label class="block text-xs font-black uppercase text-slate-500 mb-1.5">Catatan Opname</label>
              <textarea v-model="opnameForm.notes" class="w-full px-3 py-2.5 text-sm border border-slate-200 rounded-xl" rows="2" placeholder="Wajib diisi jika ada selisih" required></textarea>
            </div>
            <div class="flex gap-3 pt-2">
              <button type="button" @click="showOpnameModal = false" class="flex-1 px-4 py-2.5 text-sm font-bold border rounded-xl hover:bg-slate-50">Batal</button>
              <button type="submit" :disabled="store.isActionLoading" class="flex-1 px-4 py-2.5 text-sm font-bold bg-red-600 hover:bg-red-700 text-white rounded-xl shadow-md">Simpan Opname</button>
            </div>
          </form>
        </div>
      </div>
    </Teleport>
  </div>
</template>
