<script setup lang="ts">
import { ref } from 'vue'
import { askAI } from '@/api/ai'

const open = ref(false)
const input = ref('')
const loading = ref(false)
const error = ref('')
const messages = ref<{ role: 'user' | 'assistant'; text: string; provider?: string }[]>([
  { role: 'assistant', text: 'Halo! Saya bisa membantu seputar fitur HRIS dan cara menggunakan aplikasi.' },
])

function providerLabel(provider?: string) {
  if (provider === 'mac-mini') return 'Hermes (Mac mini)'
  if (provider === 'gemini') return 'Gemini (fallback)'
  return provider || ''
}

async function send() {
  const message = input.value.trim()
  if (!message || loading.value) return
  messages.value.push({ role: 'user', text: message }); input.value = ''; loading.value = true; error.value = ''
  try {
    const result = await askAI(message)
    messages.value.push({ role: 'assistant', text: result.answer, provider: result.provider })
  } catch (e: any) { error.value = e?.response?.data?.error || 'Asisten sedang tidak tersedia.' }
  finally { loading.value = false }
}
</script>

<template>
  <div class="fixed bottom-5 right-5 z-40 print:hidden">
    <div v-if="open" class="mb-3 flex h-[min(560px,calc(100vh-110px))] w-[min(380px,calc(100vw-32px))] flex-col overflow-hidden rounded-2xl border border-slate-200 bg-white shadow-2xl">
      <div class="flex items-center justify-between bg-slate-900 px-4 py-3 text-white">
        <div><div class="font-semibold">Asisten HRIS</div><div class="text-xs text-slate-300">Hermes (Mac mini) · fallback Gemini</div></div>
        <button class="text-xl text-slate-300 hover:text-white" @click="open = false">×</button>
      </div>
      <div class="flex-1 space-y-3 overflow-y-auto bg-slate-50 p-3">
        <div v-for="(item, i) in messages" :key="i" :class="['max-w-[88%] rounded-2xl px-3 py-2 text-sm whitespace-pre-wrap', item.role === 'user' ? 'ml-auto bg-blue-600 text-white' : 'bg-white text-slate-700 shadow-sm']">
          {{ item.text }}
          <div v-if="item.provider" class="mt-1 text-[10px] opacity-60">{{ providerLabel(item.provider) }}</div>
        </div>
        <div v-if="loading" class="w-fit rounded-2xl bg-white px-3 py-2 text-sm text-slate-400 shadow-sm">Sedang berpikir…</div>
        <div v-if="error" class="rounded-lg bg-red-50 p-2 text-xs text-red-600">{{ error }}</div>
      </div>
      <form class="flex gap-2 border-t bg-white p-3" @submit.prevent="send">
        <input v-model="input" class="min-w-0 flex-1 rounded-xl border border-slate-200 px-3 py-2 text-sm outline-none focus:border-blue-500" placeholder="Tulis pertanyaan…" :disabled="loading" />
        <button class="rounded-xl bg-blue-600 px-3 text-white disabled:opacity-50" :disabled="loading || !input.trim()">Kirim</button>
      </form>
    </div>
    <button v-else class="flex h-14 w-14 items-center justify-center rounded-full bg-slate-900 text-2xl text-white shadow-lg transition hover:scale-105" title="Buka Asisten HRIS" @click="open = true">✦</button>
  </div>
</template>
