<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import apiClient from '@/api/axios'
import { toast } from 'vue-sonner'
import { ExternalLink, Pencil, Plus, Trash2, Upload, X } from 'lucide-vue-next'

type CampaignVideo = {
  id: string
  title: string
  description: string
  video_url: string
  product_name: string
  purchase_url: string
  is_active: boolean
  created_at: string
}

const videos = ref<CampaignVideo[]>([])
const loading = ref(false)
const saving = ref(false)
const uploading = ref(false)
const showForm = ref(false)
const editingId = ref<string | null>(null)
const selectedFile = ref<File | null>(null)
const form = ref({ title: '', description: '', video_url: '', product_name: '', purchase_url: '', is_active: true })

const driveId = (url: string) => {
  try {
    const parsed = new URL(url)
    if (parsed.hostname !== 'drive.google.com') return ''
    return parsed.pathname.match(/^\/file\/d\/([^/]+)/)?.[1] || parsed.searchParams.get('id') || ''
  } catch { return '' }
}
const previewURL = (url: string) => {
  const id = driveId(url)
  return id ? `https://drive.google.com/file/d/${encodeURIComponent(id)}/preview` : ''
}
const preview = computed(() => previewURL(form.value.video_url))

async function loadVideos() {
  loading.value = true
  try {
    const response = await apiClient.get('/branding/campaign-videos')
    videos.value = response.data?.data || []
  } catch {
    toast.error('Gagal memuat video kampanye')
  } finally { loading.value = false }
}

function openCreate() {
  editingId.value = null
  selectedFile.value = null
  form.value = { title: '', description: '', video_url: '', product_name: '', purchase_url: '', is_active: true }
  showForm.value = true
}
function openEdit(video: CampaignVideo) {
  editingId.value = video.id
  selectedFile.value = null
  form.value = { title: video.title, description: video.description, video_url: video.video_url, product_name: video.product_name, purchase_url: video.purchase_url, is_active: video.is_active }
  showForm.value = true
}
function onFileChange(event: Event) {
  selectedFile.value = (event.target as HTMLInputElement).files?.[0] || null
}
async function uploadVideo() {
  if (!selectedFile.value) return
  if (selectedFile.value.size > 95 * 1024 * 1024) {
    toast.error('Ukuran maksimal unggahan langsung 95 MB. Untuk video lebih besar, unggah ke Google Drive lalu tempel tautannya.')
    return
  }
  uploading.value = true
  try {
    const data = new FormData()
    data.append('video', selectedFile.value)
    const response = await apiClient.post('/branding/campaign-videos/upload', data, { timeout: 600000 })
    form.value.video_url = response.data.video_url
    selectedFile.value = null
    toast.success('Video tersimpan di Google Drive')
  } catch (error: any) {
    toast.error(error.response?.data?.error || 'Gagal mengunggah video')
  } finally { uploading.value = false }
}
async function saveVideo() {
  if (!form.value.title.trim() || !form.value.product_name.trim()) {
    toast.error('Judul video dan nama produk wajib diisi')
    return
  }
  if (!preview.value) {
    toast.error('Masukkan tautan file video Google Drive yang valid')
    return
  }
  try {
    const purchase = new URL(form.value.purchase_url)
    if (purchase.protocol !== 'https:') throw new Error('invalid')
  } catch {
    toast.error('Masukkan tautan pembelian HTTPS yang valid')
    return
  }
  saving.value = true
  try {
    if (editingId.value) await apiClient.put(`/branding/campaign-videos/${editingId.value}`, form.value)
    else await apiClient.post('/branding/campaign-videos', form.value)
    toast.success('Video kampanye berhasil disimpan')
    showForm.value = false
    await loadVideos()
  } catch (error: any) {
    toast.error(error.response?.data?.error || 'Gagal menyimpan video')
  } finally { saving.value = false }
}
async function deleteVideo(video: CampaignVideo) {
  if (!window.confirm(`Hapus video kampanye “${video.title}” dari daftar?`)) return
  try {
    await apiClient.delete(`/branding/campaign-videos/${video.id}`)
    toast.success('Video dihapus dari daftar kampanye')
    await loadVideos()
  } catch { toast.error('Gagal menghapus video') }
}

onMounted(loadVideos)
</script>

<template>
  <div class="p-4 sm:p-6 max-w-7xl mx-auto space-y-6">
    <div class="flex flex-wrap items-center justify-between gap-3">
      <div>
        <h1 class="text-2xl font-black text-slate-900">Video Kampanye</h1>
        <p class="text-sm text-slate-500 mt-1">Arsip video di Google Drive dan tautan produk untuk persiapan unggahan media sosial.</p>
      </div>
      <button class="inline-flex items-center gap-2 rounded-xl bg-primary px-4 py-2.5 text-white font-bold" @click="openCreate"><Plus class="w-4 h-4" /> Tambah Video</button>
    </div>

    <p v-if="loading" class="text-slate-500">Memuat video...</p>
    <div v-else-if="!videos.length" class="rounded-2xl border border-dashed border-slate-300 p-12 text-center text-slate-500">Belum ada video kampanye.</div>
    <div v-else class="grid gap-5 md:grid-cols-2 xl:grid-cols-3">
      <article v-for="video in videos" :key="video.id" class="overflow-hidden rounded-2xl border border-slate-200 bg-white shadow-sm">
        <iframe :src="previewURL(video.video_url)" :title="video.title" loading="lazy" allow="autoplay; encrypted-media" allowfullscreen class="w-full aspect-video bg-slate-900" />
        <div class="p-4 space-y-3">
          <div class="flex items-start justify-between gap-2"><h2 class="font-bold text-slate-900">{{ video.title }}</h2><span class="rounded-full px-2 py-1 text-[10px] font-bold" :class="video.is_active ? 'bg-green-100 text-green-700' : 'bg-slate-100 text-slate-500'">{{ video.is_active ? 'Tayang' : 'Draft' }}</span></div>
          <p v-if="video.description" class="text-sm text-slate-600 whitespace-pre-line">{{ video.description }}</p>
          <p class="text-sm font-semibold text-slate-700">Produk: {{ video.product_name }}</p>
          <a :href="video.purchase_url" target="_blank" rel="noopener noreferrer" class="inline-flex items-center gap-2 rounded-lg border border-slate-300 px-3 py-2 text-sm font-semibold text-slate-700 hover:bg-slate-50">Buka tautan produk <ExternalLink class="w-3 h-3" /></a>
          <div class="flex gap-2 border-t border-slate-100 pt-3">
            <button class="inline-flex items-center gap-1 text-sm text-slate-600 hover:text-primary" @click="openEdit(video)"><Pencil class="w-4 h-4" /> Edit</button>
            <button class="inline-flex items-center gap-1 text-sm text-red-600" @click="deleteVideo(video)"><Trash2 class="w-4 h-4" /> Hapus</button>
          </div>
        </div>
      </article>
    </div>

    <div v-if="showForm" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4" @click.self="showForm = false">
      <form class="w-full max-w-xl max-h-[90vh] overflow-y-auto rounded-2xl bg-white p-5 space-y-4 shadow-xl" @submit.prevent="saveVideo">
        <div class="flex items-center justify-between"><h2 class="text-xl font-black">{{ editingId ? 'Edit' : 'Tambah' }} Video Kampanye</h2><button type="button" @click="showForm = false"><X class="w-5 h-5" /></button></div>
        <label class="block text-sm font-semibold">Judul video<input v-model="form.title" required maxlength="200" class="mt-1 w-full rounded-lg border border-slate-300 p-2" /></label>
        <label class="block text-sm font-semibold">Deskripsi<textarea v-model="form.description" rows="2" class="mt-1 w-full rounded-lg border border-slate-300 p-2" /></label>
        <label class="block text-sm font-semibold">Unggah video ke Google Drive (maks. 95 MB)<input type="file" accept="video/mp4,video/quicktime,video/webm,.m4v" class="mt-1 block w-full text-sm" @change="onFileChange" /></label>
        <button type="button" :disabled="!selectedFile || uploading" class="inline-flex items-center gap-2 rounded-lg bg-slate-900 px-3 py-2 text-sm font-bold text-white disabled:opacity-50" @click="uploadVideo"><Upload class="w-4 h-4" /> {{ uploading ? 'Mengunggah...' : 'Unggah ke Drive' }}</button>
        <label class="block text-sm font-semibold">Atau tempel tautan video Google Drive<input v-model="form.video_url" required type="url" placeholder="https://drive.google.com/file/d/.../view" class="mt-1 w-full rounded-lg border border-slate-300 p-2" /></label>
        <p class="text-xs text-slate-500">Untuk tautan manual, atur akses file di Drive menjadi “Siapa saja yang memiliki link” agar video bisa diputar.</p>
        <iframe v-if="preview" :src="preview" title="Pratinjau video" class="w-full aspect-video rounded-lg bg-slate-900" allowfullscreen />
        <label class="block text-sm font-semibold">Nama produk<input v-model="form.product_name" required maxlength="200" class="mt-1 w-full rounded-lg border border-slate-300 p-2" /></label>
        <label class="block text-sm font-semibold">Tautan pembelian produk<input v-model="form.purchase_url" required type="url" placeholder="https://..." class="mt-1 w-full rounded-lg border border-slate-300 p-2" /></label>
        <p class="text-xs text-amber-700">Tautan ini hanya tersimpan di web manajemen. Tag produk atau keranjang di TikTok, Instagram, dan Facebook harus ditambahkan melalui fitur toko masing-masing platform saat video dipublikasikan.</p>
        <label class="flex items-center gap-2 text-sm font-semibold"><input v-model="form.is_active" type="checkbox" /> Tayang</label>
        <div class="flex justify-end gap-2"><button type="button" class="rounded-lg border px-4 py-2" @click="showForm = false">Batal</button><button type="submit" :disabled="saving || uploading" class="rounded-lg bg-primary px-4 py-2 font-bold text-white disabled:opacity-50">{{ saving ? 'Menyimpan...' : 'Simpan' }}</button></div>
      </form>
    </div>
  </div>
</template>
