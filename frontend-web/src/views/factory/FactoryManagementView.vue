<template>
  <div class="p-6 space-y-6">
    <div class="flex justify-between items-center">
      <div>
        <h1 class="text-3xl font-bold tracking-tight">Manajemen Pabrik</h1>
        <p class="text-muted-foreground">Kelola master data unit pabrik dan lokasi operasional.</p>
      </div>
      <Button
        @click="openCreateModal"
        class="rounded-xl shadow-lg shadow-primary/20 transition-all hover:scale-105"
      >
        <Plus class="mr-2 h-4 w-4" /> Tambah Pabrik Baru
      </Button>
    </div>

    <Card class="border-none shadow-sm bg-white rounded-3xl overflow-hidden">
      <CardContent class="p-0">
        <DataTable :columns="columns" :data="filteredFactories" class="border-none">
          <template #cell-name="{ row }">
            <div class="flex items-center gap-3">
              <div
                class="h-10 w-10 rounded-xl bg-primary/10 flex items-center justify-center text-primary"
              >
                <Factory class="h-5 w-5" />
              </div>
              <div>
                <span class="font-bold text-slate-900 block">{{ (row as any).name }}</span>
                <span v-if="(row as any).warehouse_code" class="inline-flex items-center gap-1 text-[10px] font-black uppercase tracking-widest bg-indigo-50 text-indigo-700 border border-indigo-200 px-2 py-0.5 rounded-full mt-0.5">
                  <span class="w-1.5 h-1.5 rounded-full bg-indigo-500"></span>
                  {{ (row as any).warehouse_code }}
                </span>
              </div>
            </div>
          </template>
          <template #cell-branch="{ row }">
            <Badge
              variant="secondary"
              class="bg-slate-100 text-slate-600 font-bold uppercase text-[10px]"
            >
              {{ (row as any).branch?.name || 'No Branch' }}
            </Badge>
          </template>
          <template #cell-location="{ row }">
            <div class="flex items-center gap-2 text-slate-500">
              <MapPin class="h-3.5 w-3.5 text-slate-400" />
              <span class="text-sm">{{ (row as any).location || '-' }}</span>
            </div>
          </template>
          <template #cell-actions="{ row }">
            <div class="flex items-center gap-1">
              <Button
                variant="ghost"
                size="icon"
                @click="openEditModal(row)"
                class="h-8 w-8 text-slate-400 hover:text-primary hover:bg-primary/5"
              >
                <Edit2 class="h-3.5 w-3.5" />
              </Button>
              <Button
                variant="ghost"
                size="icon"
                @click="handleDelete((row as any).id)"
                class="h-8 w-8 text-slate-400 hover:text-destructive hover:bg-destructive/5"
              >
                <Trash2 class="h-3.5 w-3.5" />
              </Button>
            </div>
          </template>
        </DataTable>
      </CardContent>
    </Card>

    <!-- Factory Modal -->
    <div
      v-if="showModal"
      class="fixed inset-0 bg-slate-900/60 backdrop-blur-md z-[100] flex items-center justify-center p-4"
    >
      <Card
        class="w-full max-w-lg shadow-2xl rounded-[2rem] overflow-hidden border-none animate-in fade-in zoom-in-95 duration-300"
      >
        <CardHeader class="p-8 border-b bg-white">
          <div class="flex justify-between items-center">
            <div>
              <CardTitle class="text-2xl font-black text-slate-900">{{
                isEditing ? 'Edit Pabrik' : 'Tambah Pabrik'
              }}</CardTitle>
              <CardDescription>Lengkapi informasi identitas pabrik.</CardDescription>
            </div>
            <Button
              variant="ghost"
              size="icon"
              @click="showModal = false"
              class="rounded-full hover:bg-slate-100"
            >
              <Plus class="h-6 w-6 rotate-45 text-slate-400" />
            </Button>
          </div>
        </CardHeader>
        <CardContent class="p-8 space-y-6 bg-white">
          <div class="space-y-2">
            <label class="text-sm font-bold text-slate-700 ml-1">Nama Pabrik</label>
            <input
              v-model="form.name"
              type="text"
              placeholder="e.g. Pabrik Utama Karsa HR"
              class="w-full px-5 py-3.5 bg-slate-50 border border-slate-100 rounded-2xl text-sm font-medium focus:ring-4 focus:ring-primary/5 focus:border-primary/20 outline-none transition-all"
            />
          </div>
          <div class="space-y-2">
            <label class="text-sm font-bold text-slate-700 ml-1">Kantor Cabang Pengelola</label>
            <select
              v-model="form.branch_id"
              class="w-full px-5 py-3.5 bg-slate-50 border border-slate-100 rounded-2xl text-sm font-bold outline-none appearance-none cursor-pointer"
            >
              <option value="" disabled>Pilih Cabang...</option>
              <option
                v-for="branch in masterDataStore.branches"
                :key="branch.id"
                :value="branch.id"
              >
                {{ branch.name }}
              </option>
            </select>
          </div>
          <div class="space-y-2">
            <label class="text-sm font-bold text-slate-700 ml-1">Alamat / Lokasi</label>
            <textarea
              v-model="form.location"
              rows="3"
              placeholder="Alamat lengkap pabrik..."
              class="w-full px-5 py-4 bg-slate-50 border border-slate-100 rounded-2xl text-sm font-medium outline-none focus:ring-4 focus:ring-slate-100 resize-none"
            ></textarea>
          </div>
          <div class="space-y-2">
            <label class="text-sm font-bold text-slate-700 ml-1">
              Kode Gudang
              <span class="text-[10px] font-normal text-slate-400 ml-1">(opsional, misal: GD-A1, GD-UTAMA)</span>
            </label>
            <input
              v-model="form.warehouse_code"
              type="text"
              placeholder="GD-A1"
              class="w-full px-5 py-3.5 bg-slate-50 border border-slate-100 rounded-2xl text-sm font-mono font-bold focus:ring-4 focus:ring-primary/5 focus:border-primary/20 outline-none transition-all uppercase"
              @input="(e: any) => form.warehouse_code = e.target.value.toUpperCase()"
            />
          </div>
          <div class="space-y-2">
            <label class="text-sm font-bold text-slate-700 ml-1">Deskripsi Gudang</label>
            <textarea
              v-model="form.description"
              rows="2"
              placeholder="Deskripsi fungsi/isi gudang ini..."
              class="w-full px-5 py-4 bg-slate-50 border border-slate-100 rounded-2xl text-sm font-medium outline-none focus:ring-4 focus:ring-slate-100 resize-none"
            ></textarea>
          </div>
        </CardContent>
        <CardFooter class="bg-slate-50/50 border-t p-8 flex justify-end gap-4">
          <Button variant="ghost" @click="showModal = false" class="rounded-2xl font-bold h-12 px-8"
            >Batal</Button
          >
          <Button
            @click="handleSubmit"
            :loading="loading"
            class="rounded-2xl px-12 h-12 shadow-xl shadow-primary/20 font-black text-white bg-primary hover:bg-primary/90"
          >
            {{ isEditing ? 'Simpan Perubahan' : 'Simpan Pabrik' }}
          </Button>
        </CardFooter>
      </Card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useFactoryStore } from '@/stores/factory'
import { useMasterDataStore } from '@/stores/masterData'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardHeader,
  CardTitle,
  CardDescription,
  CardContent,
  CardFooter,
} from '@/components/ui/card/index'
import DataTable from '@/components/DataTable.vue'
import Badge from '@/components/ui/badge/Badge.vue'
import { Plus, Search, Edit2, Trash2, Factory, MapPin } from 'lucide-vue-next'
import { toast } from 'vue-sonner'

const getErrorMessage = (err: any): string => {
  if (!err) return 'Unknown error'
  if (typeof err === 'string') return err
  if (err.response?.data?.error) return err.response.data.error
  if (err.response?.data?.message) return err.response.data.message
  if (err.message) return err.message
  try {
    return JSON.stringify(err)
  } catch (_) {
    return String(err)
  }
}

const factoryStore = useFactoryStore()
const masterDataStore = useMasterDataStore()
const showModal = ref(false)
const isEditing = ref(false)
const loading = ref(false)
const searchQuery = ref('')
const currentId = ref('')

const form = reactive({
  name: '',
  location: '',
  branch_id: '',
  warehouse_code: '',
  description: '',
})

const columns = [
  { key: 'name', label: 'Nama Gudang / Kode' },
  { key: 'branch', label: 'Cabang' },
  { key: 'location', label: 'Lokasi' },
  { key: 'actions', label: '' },
]

const filteredFactories = computed(() => {
  if (!searchQuery.value) return factoryStore.factories
  const q = searchQuery.value.toLowerCase()
  return factoryStore.factories.filter(
    (f) => f.name.toLowerCase().includes(q) || f.location?.toLowerCase().includes(q),
  )
})

onMounted(async () => {
  factoryStore.fetchFactories()
  masterDataStore.fetchBranches()
})

const openCreateModal = () => {
  isEditing.value = false
  Object.assign(form, {
    name: '',
    location: '',
    branch_id: '',
    warehouse_code: '',
    description: '',
  })
  showModal.value = true
}

const openEditModal = (item: any) => {
  isEditing.value = true
  currentId.value = item.id
  Object.assign(form, {
    name: item.name,
    location: item.location || '',
    branch_id: item.branch_id,
    warehouse_code: item.warehouse_code || '',
    description: item.description || '',
  })
  showModal.value = true
}

const handleSubmit = async () => {
  if (!form.name || !form.branch_id) {
    toast.error('Nama dan Cabang harus diisi')
    return
  }

  loading.value = true
  try {
    if (isEditing.value) {
      await factoryStore.updateFactory(currentId.value, form)
      toast.success('Pabrik berhasil diperbarui')
    } else {
      await factoryStore.createFactory(form)
      toast.success('Pabrik baru berhasil ditambahkan')
    }
    showModal.value = false
  } catch (err) {
    toast.error('Gagal menyimpan pabrik: ' + getErrorMessage(err))
  } finally {
    loading.value = false
  }
}

const handleDelete = async (id: string) => {
  if (!confirm('Hapus pabrik ini? Seluruh data stok di pabrik ini akan terpengaruh.')) return
  try {
    await factoryStore.deleteFactory(id)
    toast.success('Pabrik berhasil dihapus')
  } catch (err) {
    toast.error('Gagal menghapus pabrik: ' + getErrorMessage(err))
  }
}
</script>
