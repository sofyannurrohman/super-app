<script setup lang="ts">
import { ref, computed } from 'vue'
import { X, Upload, CheckCircle2, AlertCircle, FileText, Loader2, Save } from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import apiClient from '@/api/axios'
import { toast } from 'vue-sonner'

const props = defineProps<{
  show: boolean
}>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'success'): void
}>()

const fileInput = ref<HTMLInputElement | null>(null)
const selectedFile = ref<File | null>(null)
const isPreviewing = ref(false)
const isImporting = ref(false)
const currentStep = ref(1)

const previewData = ref<any[]>([])

const handleFileSelect = (e: Event) => {
  const file = (e.target as HTMLInputElement).files?.[0]
  if (file) {
    if (!file.name.endsWith('.xlsx')) {
      toast.error('File harus berformat .xlsx')
      return
    }
    selectedFile.value = file
  }
}

const triggerFileInput = () => {
  fileInput.value?.click()
}

const removeFile = () => {
  selectedFile.value = null
  if (fileInput.value) {
    fileInput.value.value = ''
  }
}

const generatePreview = async () => {
  if (!selectedFile.value) return

  isPreviewing.value = true
  try {
    const formData = new FormData()
    formData.append('file', selectedFile.value)
    
    // We can pass company_id if needed, assuming backend falls back to user's company
    const response = await apiClient.post('/api/v1/stores/import/preview', formData, {
      headers: { 'Content-Type': 'multipart/form-data' }
    })
    
    // Map response data to add an action property
    previewData.value = response.data.data.map((item: any) => {
      let defaultAction = 'Create'
      if (item.status === 'Duplicate') defaultAction = 'Skip'
      if (item.status === 'Potential Duplicate') defaultAction = 'Skip'
      return { ...item, action: defaultAction }
    })
    
    currentStep.value = 2
  } catch (error: any) {
    console.error('Preview error:', error)
    toast.error(error.response?.data?.error || 'Gagal membuat preview import')
  } finally {
    isPreviewing.value = false
  }
}

const submitImport = async () => {
  isImporting.value = true
  try {
    const payload = {
      stores: previewData.value.map(item => ({
        name: item.name,
        address: item.address,
        phone_number: item.phone_number,
        default_salesman: item.default_salesman,
        action: item.action,
        duplicate_of_id: item.duplicate_of?.id || ''
      }))
    }
    
    const response = await apiClient.post('/api/v1/stores/import', payload)
    toast.success(response.data.message || 'Import berhasil')
    emit('success')
    closeModal()
  } catch (error: any) {
    console.error('Import error:', error)
    toast.error(error.response?.data?.error || 'Gagal melakukan import')
  } finally {
    isImporting.value = false
  }
}

const closeModal = () => {
  currentStep.value = 1
  selectedFile.value = null
  previewData.value = []
  if (fileInput.value) fileInput.value.value = ''
  emit('close')
}

const getStatusColor = (status: string) => {
  switch (status) {
    case 'New': return 'bg-green-100 text-green-700 border-green-200'
    case 'Duplicate': return 'bg-red-100 text-red-700 border-red-200'
    case 'Potential Duplicate': return 'bg-yellow-100 text-yellow-700 border-yellow-200'
    default: return 'bg-gray-100 text-gray-700 border-gray-200'
  }
}
</script>

<template>
  <div v-if="show" class="fixed inset-0 z-[100] flex items-center justify-center p-4 bg-black/50 backdrop-blur-sm">
    <div class="bg-white rounded-xl shadow-2xl w-full max-w-5xl flex flex-col max-h-[90vh]">
      <!-- Header -->
      <div class="flex items-center justify-between p-6 border-b border-gray-100">
        <div>
          <h2 class="text-xl font-bold text-gray-900">Import Toko</h2>
          <p class="text-sm text-gray-500 mt-1">Upload data toko dari file Excel untuk import secara massal</p>
        </div>
        <button @click="closeModal" class="p-2 text-gray-400 hover:text-gray-600 hover:bg-gray-100 rounded-lg transition-colors">
          <X class="w-5 h-5" />
        </button>
      </div>

      <!-- Content -->
      <div class="flex-1 overflow-y-auto p-6">
        <!-- Step 1: Upload -->
        <div v-if="currentStep === 1" class="space-y-6">
          <div 
            class="border-2 border-dashed border-gray-300 rounded-xl p-10 text-center hover:bg-gray-50 transition-colors cursor-pointer"
            @click="triggerFileInput"
            v-if="!selectedFile"
          >
            <input 
              type="file" 
              ref="fileInput" 
              class="hidden" 
              accept=".xlsx"
              @change="handleFileSelect"
            />
            <div class="bg-blue-50 w-16 h-16 rounded-full flex items-center justify-center mx-auto mb-4">
              <Upload class="w-8 h-8 text-blue-600" />
            </div>
            <h3 class="text-lg font-medium text-gray-900 mb-1">Upload File Excel</h3>
            <p class="text-sm text-gray-500 mb-4">Pilih file .xlsx yang berisi data toko</p>
            <Button type="button" variant="outline">Pilih File</Button>
          </div>

          <div v-else class="border border-gray-200 rounded-xl p-6 bg-gray-50">
            <div class="flex items-center justify-between">
              <div class="flex items-center space-x-4">
                <div class="p-3 bg-blue-100 rounded-lg">
                  <FileText class="w-6 h-6 text-blue-600" />
                </div>
                <div>
                  <h4 class="font-medium text-gray-900">{{ selectedFile.name }}</h4>
                  <p class="text-sm text-gray-500">{{ (selectedFile.size / 1024).toFixed(2) }} KB</p>
                </div>
              </div>
              <button @click="removeFile" class="text-gray-400 hover:text-red-500 transition-colors">
                <X class="w-5 h-5" />
              </button>
            </div>
          </div>
        </div>

        <!-- Step 2: Preview -->
        <div v-else-if="currentStep === 2" class="space-y-4">
          <div class="bg-yellow-50 border border-yellow-200 rounded-lg p-4 mb-4">
            <div class="flex">
              <AlertCircle class="w-5 h-5 text-yellow-600 mt-0.5 mr-3 flex-shrink-0" />
              <div>
                <h4 class="font-medium text-yellow-800">Preview Data</h4>
                <p class="text-sm text-yellow-700 mt-1">Silakan periksa data yang akan diimport. Pilih aksi yang sesuai untuk toko yang ditandai sebagai duplikat atau potensi duplikat.</p>
              </div>
            </div>
          </div>

          <div class="border border-gray-200 rounded-xl overflow-hidden">
            <div class="overflow-x-auto max-h-[50vh]">
              <table class="w-full text-sm text-left">
                <thead class="text-xs text-gray-700 uppercase bg-gray-50 sticky top-0 z-10">
                  <tr>
                    <th class="px-4 py-3">Baris</th>
                    <th class="px-4 py-3">Nama Toko</th>
                    <th class="px-4 py-3">No. HP</th>
                    <th class="px-4 py-3">Alamat</th>
                    <th class="px-4 py-3">Status Sistem</th>
                    <th class="px-4 py-3 min-w-[150px]">Aksi</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="(item, index) in previewData" :key="index" class="border-b last:border-b-0 hover:bg-gray-50">
                    <td class="px-4 py-3">{{ item.original_row }}</td>
                    <td class="px-4 py-3 font-medium">{{ item.name }}</td>
                    <td class="px-4 py-3">{{ item.phone_number || '-' }}</td>
                    <td class="px-4 py-3"><div class="line-clamp-2" :title="item.address">{{ item.address }}</div></td>
                    <td class="px-4 py-3">
                      <div class="inline-flex items-center px-2.5 py-1 rounded-full text-xs font-medium border" :class="getStatusColor(item.status)">
                        {{ item.status }}
                      </div>
                      <div v-if="item.duplicate_of" class="text-xs text-gray-500 mt-1">
                        Mirip dgn: {{ item.duplicate_of.name }}
                      </div>
                    </td>
                    <td class="px-4 py-3">
                      <select 
                        v-model="item.action"
                        class="w-full text-sm border-gray-300 rounded-lg shadow-sm focus:border-blue-500 focus:ring-blue-500"
                      >
                        <option value="Create">Buat Baru</option>
                        <option value="Skip">Abaikan</option>
                        <option v-if="item.duplicate_of" value="Update">Update Data Lama</option>
                      </select>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </div>
      </div>

      <!-- Footer -->
      <div class="p-6 border-t border-gray-100 bg-gray-50 rounded-b-xl flex justify-end space-x-3">
        <Button variant="outline" @click="closeModal">Batal</Button>
        <Button 
          v-if="currentStep === 1" 
          :disabled="!selectedFile || isPreviewing"
          @click="generatePreview"
        >
          <Loader2 v-if="isPreviewing" class="w-4 h-4 mr-2 animate-spin" />
          <FileText v-else class="w-4 h-4 mr-2" />
          Preview Import
        </Button>
        <Button 
          v-if="currentStep === 2" 
          :disabled="isImporting"
          @click="submitImport"
        >
          <Loader2 v-if="isImporting" class="w-4 h-4 mr-2 animate-spin" />
          <Save v-else class="w-4 h-4 mr-2" />
          Proses Import
        </Button>
      </div>
    </div>
  </div>
</template>
