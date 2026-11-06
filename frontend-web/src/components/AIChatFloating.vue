<script setup lang="ts">
import { ref, nextTick } from 'vue'
import {
  Sparkles,
  X,
  Send,
  Bot,
  User,
  Trash2,
  Loader2,
  Copy,
  Check,
  Maximize2,
  Minimize2,
  HelpCircle,
  Square,
} from 'lucide-vue-next'
import { askAI } from '@/api/ai'

interface ChatMessage {
  id: string
  role: 'user' | 'assistant'
  text: string
  provider?: string
  time: string
}

const isOpen = ref(false)
const isExpanded = ref(false)
const inputMessage = ref('')
const isLoading = ref(false)
const hasUnread = ref(false)
const copiedId = ref<string | null>(null)
const messagesContainer = ref<HTMLDivElement | null>(null)
let currentAbortController: AbortController | null = null

const defaultWelcomeMessage: ChatMessage = {
  id: 'welcome-1',
  role: 'assistant',
  text: 'Halo! Saya bisa membantu menjelaskan fitur HRIS dan cara menggunakan atau menginput data di Super App.',
  time: new Date().toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
}

const messages = ref<ChatMessage[]>([defaultWelcomeMessage])

const quickPrompts = [
  'Apa saja kategori fitur HRIS?',
  'Bagaimana cara input data di Super App?',
  'Bagaimana cara menggunakan aplikasi Karsa HR?'
]

const scrollToBottom = async () => {
  await nextTick()
  if (messagesContainer.value) {
    messagesContainer.value.scrollTop = messagesContainer.value.scrollHeight
  }
}

const toggleOpen = () => {
  isOpen.value = !isOpen.value
  if (isOpen.value) {
    hasUnread.value = false
    scrollToBottom()
  }
}

const handleQuickPrompt = (prompt: string) => {
  inputMessage.value = prompt
  handleSendMessage()
}

const cancelOngoingRequest = () => {
  if (currentAbortController) {
    currentAbortController.abort()
    currentAbortController = null
  }
  isLoading.value = false
  messages.value.push({
    id: `canceled-${Date.now()}`,
    role: 'assistant',
    text: 'Permintaan dibatalkan.',
    time: new Date().toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
  })
  scrollToBottom()
}

const handleSendMessage = async () => {
  const text = inputMessage.value.trim()
  if (!text || isLoading.value) return

  const userMsgId = `user-${Date.now()}`
  messages.value.push({
    id: userMsgId,
    role: 'user',
    text,
    time: new Date().toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
  })

  inputMessage.value = ''
  isLoading.value = true
  await scrollToBottom()

  currentAbortController = new AbortController()

  try {
    const res = await askAI(text, currentAbortController.signal)

    messages.value.push({
      id: `ai-${Date.now()}`,
      role: 'assistant',
      text: res.answer || 'Maaf, tidak ada respon yang diterima dari server AI.',
      provider: res.provider,
      time: new Date().toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
    })

    if (!isOpen.value) {
      hasUnread.value = true
    }
  } catch (err: any) {
    if (err?.name === 'CanceledError' || err?.code === 'ERR_CANCELED') {
      return // Dibatalkan oleh user
    }
    const errorMessage = err?.response?.data?.error || 'Maaf, layanan AI asisten sedang tidak tersedia atau mengalami kendala koneksi.'
    messages.value.push({
      id: `error-${Date.now()}`,
      role: 'assistant',
      text: errorMessage,
      time: new Date().toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
    })
    if (!isOpen.value) {
      hasUnread.value = true
    }
  } finally {
    isLoading.value = false
    currentAbortController = null
    await scrollToBottom()
  }
}

const handleKeydown = (e: KeyboardEvent) => {
  if (e.key === 'Enter' && !e.shiftKey) {
    e.preventDefault()
    handleSendMessage()
  }
}

const clearChat = () => {
  messages.value = [
    {
      ...defaultWelcomeMessage,
      id: `welcome-${Date.now()}`,
      time: new Date().toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
    }
  ]
}

const copyToClipboard = async (msg: ChatMessage) => {
  try {
    await navigator.clipboard.writeText(msg.text)
    copiedId.value = msg.id
    setTimeout(() => {
      if (copiedId.value === msg.id) {
        copiedId.value = null
      }
    }, 2000)
  } catch (err) {
    console.error('Failed to copy text', err)
  }
}

// Simple text formatter for clean rendering of lists and bold text
const formatMessage = (content: string) => {
  if (!content) return ''
  // Escaping basic HTML to prevent injection
  const escaped = content
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')

  // Replace **bold** with <strong>
  const boldFormatted = escaped.replace(/\*\*(.*?)\*\*/g, '<strong class="font-semibold text-slate-900">$1</strong>')

  // Replace `code` with styled <code>
  const codeFormatted = boldFormatted.replace(/`([^`]+)`/g, '<code class="px-1.5 py-0.5 rounded bg-slate-200/80 text-indigo-700 font-mono text-xs">$1</code>')

  // Replace linebreaks with <br/>
  return codeFormatted.replace(/\n/g, '<br/>')
}
</script>

<template>
  <!-- Floating Button & Chat Dialog Container -->
  <div class="fixed bottom-36 right-5 z-40 print:hidden md:bottom-22 md:right-6">
    <!-- Chat Modal Window -->
    <Transition name="ai-chat-popup">
      <div
        v-if="isOpen"
        :class="[
          'flex flex-col overflow-hidden rounded-2xl border border-slate-200/90 bg-white shadow-2xl transition-all duration-200',
          isExpanded
            ? 'mb-3 h-[min(680px,calc(100vh-120px))] w-[min(620px,calc(100vw-30px))]'
            : 'mb-3 h-[min(540px,calc(100vh-160px))] w-[min(410px,calc(100vw-30px))]'
        ]"
        role="dialog"
        aria-label="Asisten AI Karsa HR Chat"
      >
        <!-- Header -->
        <div class="flex items-center justify-between bg-gradient-to-r from-indigo-700 via-purple-700 to-violet-800 px-4 py-3 text-white shadow-md">
          <div class="flex items-center gap-2.5">
            <div class="relative flex h-9 w-9 items-center justify-center rounded-xl bg-white/20 backdrop-blur-xs shadow-inner">
              <Sparkles :size="20" class="text-amber-300 animate-pulse" />
            </div>
            <div>
              <div class="flex items-center gap-1.5">
                <p class="text-sm font-bold tracking-tight">Asisten HRIS</p>
                <span class="inline-flex items-center px-1.5 py-0.2 rounded-full text-[9px] font-semibold bg-emerald-400/25 text-emerald-200 border border-emerald-400/30">
                  Online
                </span>
              </div>
              <p class="text-[11px] text-indigo-100/85 font-medium flex items-center gap-1">
                <span>Fitur HRIS & panduan Super App</span>
              </p>
            </div>
          </div>

          <div class="flex items-center gap-1">
            <button
              class="hidden sm:inline-flex rounded-lg p-1.5 text-white/80 transition hover:bg-white/15 hover:text-white cursor-pointer"
              :title="isExpanded ? 'Kecilkan Tampilan' : 'Perbesar Tampilan'"
              @click="isExpanded = !isExpanded"
            >
              <Minimize2 v-if="isExpanded" :size="16" />
              <Maximize2 v-else :size="16" />
            </button>
            <button
              class="rounded-lg p-1.5 text-white/80 transition hover:bg-white/15 hover:text-white cursor-pointer"
              title="Bersihkan Percakapan"
              @click="clearChat"
            >
              <Trash2 :size="16" />
            </button>
            <button
              class="rounded-lg p-1.5 text-white/80 transition hover:bg-white/15 hover:text-white cursor-pointer"
              aria-label="Tutup Asisten AI"
              title="Tutup (Tetap berjalan di latar belakang)"
              @click="isOpen = false"
            >
              <X :size="18" />
            </button>
          </div>
        </div>

        <!-- Chat Messages Body -->
        <div
          ref="messagesContainer"
          class="flex-1 space-y-3.5 overflow-y-auto bg-slate-50/70 p-3.5 text-xs text-slate-800"
        >
          <div
            v-for="msg in messages"
            :key="msg.id"
            :class="['flex gap-2.5', msg.role === 'user' ? 'justify-end' : 'justify-start']"
          >
            <!-- Avatar Bot -->
            <div
              v-if="msg.role === 'assistant'"
              class="flex h-7 w-7 shrink-0 items-center justify-center rounded-lg bg-gradient-to-tr from-indigo-600 to-purple-600 text-white shadow-xs"
            >
              <Bot :size="16" />
            </div>

            <!-- Bubble Content -->
            <div
              :class="[
                'group relative max-w-[85%] rounded-2xl p-3 shadow-xs transition',
                msg.role === 'user'
                  ? 'bg-gradient-to-r from-indigo-600 to-purple-600 text-white rounded-br-xs'
                  : 'bg-white border border-slate-200/80 text-slate-800 rounded-tl-xs'
              ]"
            >
              <!-- Text -->
              <div
                class="leading-relaxed break-words"
                v-html="formatMessage(msg.text)"
              />

              <!-- Meta Footer -->
              <div
                :class="[
                  'mt-1.5 flex items-center justify-between gap-2 text-[10px]',
                  msg.role === 'user' ? 'text-indigo-200' : 'text-slate-400'
                ]"
              >
                <span>{{ msg.time }}</span>
                <div class="flex items-center gap-1.5">
                  <button
                    v-if="msg.role === 'assistant'"
                    class="opacity-0 group-hover:opacity-100 transition hover:text-indigo-600 cursor-pointer"
                    :title="copiedId === msg.id ? 'Tersalin' : 'Salin Pesan'"
                    @click="copyToClipboard(msg)"
                  >
                    <Check v-if="copiedId === msg.id" :size="12" class="text-emerald-500" />
                    <Copy v-else :size="12" />
                  </button>
                </div>
              </div>
            </div>

            <!-- Avatar User -->
            <div
              v-if="msg.role === 'user'"
              class="flex h-7 w-7 shrink-0 items-center justify-center rounded-lg bg-slate-200 text-slate-600 shadow-xs"
            >
              <User :size="16" />
            </div>
          </div>

          <!-- Loading Indicator Bubble with Cancel Option -->
          <div v-if="isLoading" class="flex items-center gap-2.5">
            <div class="flex h-7 w-7 shrink-0 items-center justify-center rounded-lg bg-gradient-to-tr from-indigo-600 to-purple-600 text-white shadow-xs">
              <Bot :size="16" />
            </div>
            <div class="flex items-center justify-between gap-3 rounded-2xl rounded-tl-xs border border-indigo-100 bg-white px-3.5 py-2.5 text-xs text-slate-600 shadow-xs">
              <div class="flex items-center gap-2">
                <Loader2 :size="15" class="animate-spin text-indigo-600" />
                <span class="text-slate-600 font-medium">Menyiapkan jawaban...</span>
              </div>
              <button
                class="inline-flex items-center gap-1 px-2 py-0.5 text-[11px] font-medium text-rose-600 hover:bg-rose-50 rounded-md transition cursor-pointer border border-rose-200"
                title="Batalkan permintaan ini"
                @click="cancelOngoingRequest"
              >
                <Square :size="11" class="fill-rose-600" />
                <span>Batal</span>
              </button>
            </div>
          </div>
        </div>

        <!-- Quick Prompts (Only shown if few messages) -->
        <div v-if="messages.length <= 2 && !isLoading" class="border-t border-slate-100 bg-white p-2">
          <p class="mb-1.5 px-1 text-[10px] font-semibold uppercase tracking-wider text-slate-400 flex items-center gap-1">
            <HelpCircle :size="12" /> Rekomendasi Pertanyaan:
          </p>
          <div class="flex flex-wrap gap-1.5">
            <button
              v-for="(qp, idx) in quickPrompts"
              :key="idx"
              class="rounded-lg border border-slate-200 bg-slate-50 px-2 py-1 text-[11px] text-slate-600 transition hover:border-indigo-300 hover:bg-indigo-50 hover:text-indigo-700 active:scale-95 text-left cursor-pointer"
              @click="handleQuickPrompt(qp)"
            >
              {{ qp }}
            </button>
          </div>
        </div>

        <!-- Input Footer -->
        <div class="border-t border-slate-200/90 bg-white p-2.5">
          <form @submit.prevent="handleSendMessage" class="flex items-end gap-2">
            <textarea
              v-model="inputMessage"
              rows="1"
              placeholder="Tanya fitur HRIS atau cara input data..."
              class="max-h-24 min-h-[38px] flex-1 resize-none rounded-xl border border-slate-200 bg-slate-50/70 p-2 text-xs text-slate-800 placeholder:text-slate-400 focus:border-indigo-500 focus:bg-white focus:outline-hidden focus:ring-2 focus:ring-indigo-100 leading-relaxed"
              :disabled="isLoading"
              @keydown="handleKeydown"
            />
            <button
              v-if="!isLoading"
              type="submit"
              :disabled="!inputMessage.trim()"
              class="flex h-[38px] w-[38px] shrink-0 items-center justify-center rounded-xl bg-gradient-to-tr from-indigo-600 to-purple-600 text-white shadow-md transition hover:from-indigo-500 hover:to-purple-500 disabled:cursor-not-allowed disabled:opacity-40 active:scale-95 cursor-pointer"
              aria-label="Kirim Pesan"
              title="Kirim (Enter)"
            >
              <Send :size="16" />
            </button>
            <button
              v-else
              type="button"
              class="flex h-[38px] w-[38px] shrink-0 items-center justify-center rounded-xl bg-rose-500 hover:bg-rose-600 text-white shadow-md transition active:scale-95 cursor-pointer"
              aria-label="Batalkan Permintaan"
              title="Batalkan (Stop)"
              @click="cancelOngoingRequest"
            >
              <Square :size="14" class="fill-white" />
            </button>
          </form>
          <div class="mt-1 flex items-center justify-between px-1 text-[10px] text-slate-400">
            <span>Tekan <kbd class="font-mono bg-slate-100 px-1 py-0.5 rounded text-[9px]">Enter</kbd> kirim, <kbd class="font-mono bg-slate-100 px-1 py-0.5 rounded text-[9px]">Shift+Enter</kbd> baris baru</span>
            <span class="text-indigo-600 font-medium">Non-blocking background</span>
          </div>
        </div>
      </div>
    </Transition>

    <!-- Floating Trigger Button -->
    <button
      class="group relative flex h-14 w-14 items-center justify-center rounded-full bg-gradient-to-br from-indigo-600 via-purple-600 to-pink-600 text-white shadow-xl shadow-indigo-900/25 transition-all duration-200 hover:scale-105 hover:shadow-2xl hover:from-indigo-500 hover:via-purple-500 hover:to-pink-500 focus:outline-hidden focus:ring-4 focus:ring-purple-200 active:scale-95 cursor-pointer"
      :aria-label="isOpen ? 'Tutup Asisten AI' : 'Buka Asisten AI Karsa HR'"
      :title="isOpen ? 'Tutup Asisten AI' : (isLoading ? 'AI sedang menjawab di latar belakang...' : (hasUnread ? 'Jawaban AI sudah siap!' : 'Tanya Asisten HRIS'))"
      @click="toggleOpen"
    >
      <X v-if="isOpen" :size="24" class="transition-transform duration-200 group-hover:rotate-90" />
      <div v-else class="relative flex items-center justify-center">
        <!-- If loading in background, show spinner instead of static bot -->
        <Loader2 v-if="isLoading" :size="26" class="animate-spin text-amber-300" />
        <Bot v-else :size="26" class="transition-transform duration-200 group-hover:scale-110" />
        <Sparkles v-if="!isLoading" :size="13" class="absolute -top-1.5 -right-2 text-amber-300 animate-pulse" />
      </div>

      <!-- Pulse Dot / Unread Notification Badge -->
      <span
        v-if="!isOpen && (isLoading || hasUnread)"
        class="absolute -top-1 -right-1 flex h-4 w-4"
      >
        <span
          :class="[
            'absolute inline-flex h-full w-full animate-ping rounded-full opacity-75',
            hasUnread ? 'bg-emerald-400' : 'bg-amber-400'
          ]"
        />
        <span
          :class="[
            'relative inline-flex h-4 w-4 rounded-full border-2 border-white',
            hasUnread ? 'bg-emerald-500' : 'bg-amber-500'
          ]"
        />
      </span>
      <span
        v-else-if="!isOpen"
        class="absolute -top-1 -right-1 flex h-4 w-4"
      >
        <span class="absolute inline-flex h-full w-full animate-ping rounded-full bg-purple-400 opacity-75" />
        <span class="relative inline-flex h-4 w-4 rounded-full border-2 border-white bg-purple-500" />
      </span>
    </button>
  </div>
</template>

<style scoped>
.ai-chat-popup-enter-active,
.ai-chat-popup-leave-active {
  transition: all 0.22s cubic-bezier(0.16, 1, 0.3, 1);
}
.ai-chat-popup-enter-from,
.ai-chat-popup-leave-to {
  opacity: 0;
  transform: translateY(12px) scale(0.96);
}
</style>
