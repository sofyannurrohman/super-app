<template>
  <div class="relative">
    <Button
      variant="ghost"
      size="icon"
      @click="isOpen = !isOpen"
      class="relative h-11 w-11 rounded-2xl border border-transparent hover:border-slate-200 hover:bg-white hover:shadow-sm text-slate-500 transition-all duration-300"
    >
      <Bell class="h-5.5 w-5.5" />
      <span v-if="unreadCount > 0" class="absolute right-3 top-3 flex h-3 w-3">
        <span
          class="absolute inline-flex h-full w-full animate-ping rounded-full bg-primary/40 opacity-75"
        ></span>
        <span
          class="relative inline-flex h-3 w-3 rounded-full bg-primary border-2 border-white"
        ></span>
      </span>
    </Button>

    <div
      v-if="isOpen"
      class="absolute right-0 mt-2 w-80 bg-white rounded-2xl shadow-xl border border-slate-200 z-50 overflow-hidden"
    >
      <div class="max-h-96 overflow-y-auto">
        <div v-if="notifications.length === 0" class="p-8 text-center text-muted-foreground italic">
          Tidak ada notifikasi.
        </div>
        <div
          v-for="notif in notifications"
          :key="notif.id"
          @click="markAsRead(notif.id)"
          class="p-4 border-b hover:bg-slate-50 cursor-pointer transition-colors"
          :class="{ 'bg-primary/5': !notif.is_read }"
        >
          <div class="flex gap-3">
            <div class="mt-1">
              <AlertCircle v-if="notif.type === 'LOW_STOCK'" class="h-5 w-5 text-red-500" />
              <Truck v-else class="h-5 w-5 text-blue-500" />
            </div>
            <div>
              <p class="text-sm font-bold">{{ notif.title }}</p>
              <p class="text-xs text-slate-600 mt-0.5">{{ notif.message }}</p>
              <p class="text-[10px] text-slate-400 mt-2">
                {{ new Date(notif.created_at).toLocaleString() }}
              </p>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed } from 'vue'
import apiClient from '@/api/axios'
import { Bell, AlertCircle, Truck } from 'lucide-vue-next'
import { Button } from '@/components/ui/button'

const isOpen = ref(false)
const notifications = ref<any[]>([])
let pollTimer: ReturnType<typeof setInterval> | null = null

const unreadCount = computed(() => notifications.value.filter((n) => !n.is_read).length)

const fetchNotifications = async () => {
  try {
    const response = await apiClient.get('/notifications')
    notifications.value = response.data
  } catch (err) {
    console.error('Failed to fetch notifications', err)
  }
}

const markAsRead = async (id: string) => {
  try {
    await apiClient.post(`/notifications/${id}/read`)
    fetchNotifications()
  } catch (err) {
    console.error('Failed to mark notification as read', err)
  }
}

onMounted(() => {
  fetchNotifications()
  // Poll every minute
  pollTimer = setInterval(fetchNotifications, 60000)
})

onUnmounted(() => {
  if (pollTimer) {
    clearInterval(pollTimer)
    pollTimer = null
  }
})
</script>
