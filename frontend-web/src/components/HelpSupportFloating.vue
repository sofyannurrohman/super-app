<script setup lang="ts">
import { ref } from 'vue'
import { Headset, Users, UserRound, X, MessageSquare } from 'lucide-vue-next'

const open = ref(false)

const defaultPersonalNumber = ''
const defaultGroupUrl = ''

function formatWhatsAppPersonalUrl(rawInput?: string): string {
  const input = (rawInput || defaultPersonalNumber).trim()
  if (input.startsWith('http://') || input.startsWith('https://')) {
    return input
  }

  let cleaned = input.replace(/\D/g, '')
  if (cleaned.startsWith('0')) {
    cleaned = '62' + cleaned.slice(1)
  }

  const number = cleaned || defaultPersonalNumber
  if (!number) return ''
  return `https://wa.me/${number}?text=${encodeURIComponent('Halo, saya membutuhkan bantuan terkait aplikasi HRIS.')}`
}

function formatWhatsAppGroupUrl(rawInput?: string): string {
  const input = (rawInput || defaultGroupUrl).trim()
  return input || defaultGroupUrl
}

const rawEnvPersonal = import.meta.env.VITE_SUPPORT_WHATSAPP_PERSONAL
const rawEnvGroup = import.meta.env.VITE_SUPPORT_WHATSAPP_GROUP

const personalUrl = formatWhatsAppPersonalUrl(rawEnvPersonal)
const groupUrl = formatWhatsAppGroupUrl(rawEnvGroup)
</script>

<template>
  <div v-if="personalUrl || groupUrl" class="fixed bottom-20 right-5 z-40 print:hidden md:bottom-6 md:right-6">
    <Transition name="support-panel">
      <div
        v-if="open"
        class="mb-3 w-[min(340px,calc(100vw-40px))] overflow-hidden rounded-2xl border border-slate-200 bg-white shadow-2xl transition-all"
        role="dialog"
        aria-label="Layanan Bantuan Customer Support"
      >
        <!-- Header -->
        <div class="flex items-center justify-between bg-gradient-to-r from-emerald-700 to-teal-800 px-4 py-3.5 text-white shadow-sm">
          <div class="flex items-center gap-2.5">
            <div class="flex h-9 w-9 items-center justify-center rounded-xl bg-white/15 backdrop-blur-xs">
              <Headset :size="20" class="text-white" />
            </div>
            <div>
              <p class="text-sm font-bold tracking-tight">Customer Support</p>
              <p class="text-[11px] text-teal-100/90 font-medium">Bantuan & Layanan WhatsApp</p>
            </div>
          </div>
          <button
            class="rounded-full p-1.5 text-white/80 transition hover:bg-white/20 hover:text-white"
            aria-label="Tutup layanan bantuan"
            @click="open = false"
          >
            <X :size="18" />
          </button>
        </div>

        <!-- Menu Options -->
        <div class="space-y-2 p-3 bg-slate-50/60">
          <a
            v-if="personalUrl"
            :href="personalUrl"
            target="_blank"
            rel="noopener noreferrer"
            class="group flex items-center gap-3.5 rounded-xl border border-slate-100 bg-white p-3 shadow-xs transition hover:border-emerald-200 hover:bg-emerald-50/70"
          >
            <span class="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-emerald-100 text-emerald-700 transition group-hover:bg-emerald-600 group-hover:text-white">
              <UserRound :size="20" />
            </span>
            <div class="min-w-0 flex-1">
              <span class="block text-xs font-bold text-slate-800 group-hover:text-emerald-900">Chat Personal CS</span>
              <span class="block text-[11px] text-slate-500 line-clamp-1">Konsultasi langsung via WhatsApp</span>
            </div>
            <MessageSquare :size="16" class="text-slate-300 group-hover:text-emerald-600 shrink-0" />
          </a>

          <a
            v-if="groupUrl"
            :href="groupUrl"
            target="_blank"
            rel="noopener noreferrer"
            class="group flex items-center gap-3.5 rounded-xl border border-slate-100 bg-white p-3 shadow-xs transition hover:border-teal-200 hover:bg-teal-50/70"
          >
            <span class="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-teal-100 text-teal-700 transition group-hover:bg-teal-600 group-hover:text-white">
              <Users :size="20" />
            </span>
            <div class="min-w-0 flex-1">
              <span class="block text-xs font-bold text-slate-800 group-hover:text-teal-900">Grup Komunitas HRIS</span>
              <span class="block text-[11px] text-slate-500 line-clamp-1">Diskusi & informasi operasional</span>
            </div>
            <MessageSquare :size="16" class="text-slate-300 group-hover:text-teal-600 shrink-0" />
          </a>
        </div>

        <div class="border-t border-slate-100 bg-white px-3 py-2 text-center text-[10px] font-medium text-slate-400">
          Jam operasional: Setiap hari 08:00 - 17:00 WIB
        </div>
      </div>
    </Transition>

    <!-- Floating Trigger Button -->
    <button
      class="group relative flex h-14 w-14 items-center justify-center rounded-full bg-gradient-to-br from-emerald-600 to-teal-700 text-white shadow-xl shadow-emerald-900/25 transition-all duration-200 hover:scale-105 hover:shadow-2xl hover:from-emerald-500 hover:to-teal-600 focus:outline-hidden focus:ring-4 focus:ring-emerald-200 active:scale-95"
      :aria-label="open ? 'Tutup layanan bantuan' : 'Buka Layanan Customer Support'"
      :title="open ? 'Tutup layanan bantuan' : 'Customer Support (Layanan Bantuan)'"
      @click="open = !open"
    >
      <X v-if="open" :size="24" class="transition-transform duration-200 group-hover:rotate-90" />
      <Headset v-else :size="26" class="transition-transform duration-200 group-hover:scale-110" />
      
      <!-- Ripple Ping Indicator -->
      <span
        v-if="!open"
        class="absolute -top-1 -right-1 flex h-4 w-4"
      >
        <span class="absolute inline-flex h-full w-full animate-ping rounded-full bg-emerald-400 opacity-75" />
        <span class="relative inline-flex h-4 w-4 rounded-full border-2 border-white bg-emerald-500" />
      </span>
    </button>
  </div>
</template>

<style scoped>
.support-panel-enter-active,
.support-panel-leave-active {
  transition: all 0.22s cubic-bezier(0.16, 1, 0.3, 1);
}
.support-panel-enter-from,
.support-panel-leave-to {
  opacity: 0;
  transform: translateY(12px) scale(0.96);
}
</style>
