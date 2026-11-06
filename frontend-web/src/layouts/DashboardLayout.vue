<script setup lang="ts">
import Sidebar from '@/components/Sidebar.vue'
import Header from '@/components/Header.vue'
import MobileBottomNav from '@/components/MobileBottomNav.vue'
import AnnouncementConfirmationModal from '@/components/AnnouncementConfirmationModal.vue'
import { RouterView } from 'vue-router'
import { useLayoutStore } from '@/stores/layout'
import { useMasterDataStore } from '@/stores/masterData'
import HelpSupportFloating from '@/components/HelpSupportFloating.vue'
import AIChatFloating from '@/components/AIChatFloating.vue'

const layoutStore = useLayoutStore()
const masterDataStore = useMasterDataStore()
</script>

<template>
  <div class="flex h-screen bg-[#fafbfc] font-sans text-gray-900 overflow-hidden text-[15px] print:h-auto print:overflow-visible">
    <!-- Desktop Sidebar (hidden on mobile) -->
    <div class="hidden md:flex h-full print:hidden">
      <Sidebar />
    </div>

    <!-- Mobile Drawer Overlay -->
    <Transition name="drawer-fade">
      <div v-if="layoutStore.isMobileDrawerOpen" class="fixed inset-0 z-50 md:hidden print:hidden">
        <!-- Backdrop -->
        <div
          class="absolute inset-0 bg-slate-900/50 backdrop-blur-sm"
          @click="layoutStore.closeMobileDrawer"
        />
        <!-- Drawer Panel (slides in from left) -->
        <div
          class="absolute left-0 top-0 h-full w-[280px] bg-white shadow-2xl overflow-y-auto drawer-panel"
        >
          <Sidebar />
        </div>
      </div>
    </Transition>

    <!-- Main Content Area -->
    <div class="flex flex-col flex-1 overflow-hidden min-w-0 print:overflow-visible print:h-auto print:block">
      <Header class="print:hidden" />

      <!-- Main scrollable area — extra bottom padding on mobile for bottom nav -->
      <main class="flex-1 overflow-x-hidden overflow-y-auto pb-20 md:pb-0 print:overflow-visible print:h-auto print:p-0 print:block">
        <div class="p-4 md:p-8 w-full max-w-[1600px] mx-auto print:max-w-none print:p-0">
          <RouterView v-slot="{ Component }">
            <Transition name="page-fade" mode="out-in" appear>
              <div :key="$route.fullPath + '-' + (masterDataStore.selectedBranchId || '') + '-' + (masterDataStore.selectedWarehouseBranchId || '')">
                <component :is="Component" />
              </div>
            </Transition>
          </RouterView>
        </div>
      </main>

      <!-- Mobile Bottom Navigation (hidden on desktop) -->
      <MobileBottomNav class="md:hidden print:hidden" />
    </div>

    <!-- Global Realtime Announcement Auto-Popup, CS Support, and AI Assistant -->
    <AnnouncementConfirmationModal />
    <HelpSupportFloating />
    <AIChatFloating />
  </div>
</template>

<style>
body {
  font-family:
    'Inter',
    -apple-system,
    BlinkMacSystemFont,
    'Segoe UI',
    Roboto,
    Helvetica,
    Arial,
    sans-serif;
}

@media print {
  /* Reset the full-screen dashboard layout so print can flow normally */
  body,
  html {
    overflow: visible !important;
    height: auto !important;
  }

  /* The outermost flex container (h-screen overflow-hidden) must be reset */
  body > #app > div {
    display: block !important;
    height: auto !important;
    overflow: visible !important;
    background: white !important;
  }

  /* Hide sidebar, header, mobile nav */
  body > #app > div > div:first-child,
  body > #app > div > div.fixed { display: none !important; }

  /* Main content area: remove flex, overflow */
  body > #app > div > div.flex-col {
    display: block !important;
    overflow: visible !important;
    height: auto !important;
  }

  /* Header */
  body > #app > div > div.flex-col > header { display: none !important; }

  /* The main scrollable area */
  body > #app > div > div.flex-col > main {
    display: block !important;
    overflow: visible !important;
    height: auto !important;
    padding: 0 !important;
  }

  /* The inner padding wrapper */
  body > #app > div > div.flex-col > main > div {
    padding: 0 !important;
    max-width: 100% !important;
  }
}

/* Drawer backdrop fade */
.drawer-fade-enter-active,
.drawer-fade-leave-active {
  transition: opacity 0.25s ease;
}
.drawer-fade-enter-from,
.drawer-fade-leave-to {
  opacity: 0;
}

/* Drawer panel slide-in */
.drawer-panel {
  animation: slideInLeft 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

@keyframes slideInLeft {
  from {
    transform: translateX(-100%);
  }
  to {
    transform: translateX(0);
  }
}
</style>
