<template>
  <Dialog :open="open" @update:open="$emit('update:open', $event)">
    <DialogContent class="sm:max-w-[425px]">
      <DialogHeader>
        <DialogTitle>{{ isEditing ? 'Edit Master Bahan Baku' : 'Tambah Master Bahan Baku' }}</DialogTitle>
        <DialogDescription>
          {{ isEditing ? 'Perbarui informasi bahan baku atau barang pengadaan.' : 'Daftarkan bahan baku baru ke dalam sistem sebelum bisa dilakukan mutasi stok.' }}
        </DialogDescription>
      </DialogHeader>

      <form @submit.prevent="submit" class="space-y-4">
        <div class="space-y-2">
          <Label for="sku">Kode (SKU)</Label>
          <Input id="sku" v-model="form.sku" placeholder="Contoh: RAW-001" required />
        </div>

        <div class="space-y-2">
          <Label for="name">Nama Bahan Baku</Label>
          <Input id="name" v-model="form.name" placeholder="Contoh: Tepung Terigu" required />
        </div>

        <div class="space-y-2">
          <Label for="category">Kategori Pengadaan</Label>
          <select id="category" v-model="form.category" class="flex h-10 w-full rounded-md border border-slate-300 bg-transparent px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-slate-400" required>
            <option value="BAHAN BAKU">⭐ Bahan Baku Utama (Kedelai, Gula, dll.)</option>
            <option value="KEMASAN">📦 Kemasan & Wadah (Botol, Kardus, Tutup, dll.)</option>
            <option value="PRODUK_PELENGKAP">🏷️ Produk Pelengkap / Trading</option>
            <option value="HADIAH">🎁 Hadiah & Promosi (Piring, Mangkok, dll.)</option>
          </select>
        </div>

        <div class="space-y-2">
          <Label for="unit">Satuan Default (UoM)</Label>
          <select id="unit" v-model="form.unit" class="flex h-10 w-full rounded-md border border-slate-300 bg-transparent px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-slate-400 font-bold" required>
            <optgroup v-for="g in COMMON_UOM_GROUPS" :key="g.name" :label="g.name">
              <option v-for="u in g.units" :key="u" :value="u">{{ u }}</option>
            </optgroup>
          </select>
        </div>

        <DialogFooter>
          <Button type="button" variant="outline" @click="$emit('update:open', false)">
            Batal
          </Button>
          <Button type="submit" :disabled="loading">
            <Loader2 v-if="loading" class="mr-2 h-4 w-4 animate-spin" />
            {{ isEditing ? 'Simpan Perubahan' : 'Simpan Bahan Baku' }}
          </Button>
        </DialogFooter>
      </form>
    </DialogContent>
  </Dialog>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Loader2 } from 'lucide-vue-next'
import { useFactoryStore } from '@/stores/factory'
import { toast } from 'vue-sonner'
import { COMMON_UOM_GROUPS } from '@/utils/productUnits'

const props = defineProps<{
  open: boolean
  product?: any
}>()

const emit = defineEmits<{
  (e: 'update:open', value: boolean): void
  (e: 'success'): void
}>()

const factoryStore = useFactoryStore()
const loading = ref(false)

const isEditing = computed(() => !!props.product?.id)

const form = ref({
  sku: '',
  name: '',
  category: 'BAHAN BAKU',
  unit: 'KG',
})

watch(
  () => props.open,
  (isOpen) => {
    if (isOpen) {
      if (props.product) {
        form.value = {
          sku: props.product.sku || '',
          name: props.product.name || '',
          category: props.product.category || 'BAHAN BAKU',
          unit: props.product.unit || 'KG',
        }
      } else {
        form.value = {
          sku: '',
          name: '',
          category: 'BAHAN BAKU',
          unit: 'KG',
        }
      }
    }
  }
)

async function submit() {
  if (!form.value.sku || !form.value.name) {
    toast.error('SKU dan Nama wajib diisi')
    return
  }
  loading.value = true
  
  const payload = new FormData()
  payload.append('name', form.value.name)
  payload.append('sku', form.value.sku)
  payload.append('unit', form.value.unit)
  payload.append('category', form.value.category)
  payload.append('brand', props.product?.brand || '-')
  payload.append('pcs_per_unit', String(props.product?.pcs_per_unit || '1'))
  payload.append('weight', String(props.product?.weight || '1'))
  payload.append('weight_unit', form.value.unit === 'KG' || form.value.unit === 'GRAM' || form.value.unit === 'GR' ? form.value.unit : (props.product?.weight_unit || 'KG'))
  payload.append('cost_price', String(props.product?.cost_price || '0'))
  payload.append('factory_price', String(props.product?.factory_price || '0'))
  payload.append('selling_price', String(props.product?.selling_price || '0'))

  try {
    if (isEditing.value) {
      await factoryStore.updateProduct(props.product.id, payload as any)
      toast.success('Master barang pengadaan berhasil diperbarui')
    } else {
      await factoryStore.createProduct(payload as any)
      toast.success('Master barang pengadaan berhasil ditambahkan')
    }
    emit('success')
    emit('update:open', false)
    form.value = {
      sku: '',
      name: '',
      category: 'BAHAN BAKU',
      unit: 'KG',
    }
    // Refresh products list in store so the dropdown is populated
    await factoryStore.fetchProducts()
  } catch (err: any) {
    toast.error(err.toString() || 'Gagal menyimpan bahan baku')
  } finally {
    loading.value = false
  }
}
</script>
