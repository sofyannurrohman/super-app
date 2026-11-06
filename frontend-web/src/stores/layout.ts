import { defineStore } from 'pinia'
import { ref } from 'vue'
import { isModuleEnabled } from '@/config/modules'

export const useLayoutStore = defineStore('layout', () => {
  const isSidebarCollapsed = ref(false)
  const isMobileMenuOpen = ref(false)
  const isMobileDrawerOpen = ref(false)
  const currentModule = ref('hris') // 'hris', 'erecruitment', 'sales'

  function toggleSidebar() {
    isSidebarCollapsed.value = !isSidebarCollapsed.value
  }

  function setSidebarCollapsed(value: boolean) {
    isSidebarCollapsed.value = value
  }

  function toggleMobileMenu() {
    isMobileMenuOpen.value = !isMobileMenuOpen.value
  }

  function openMobileDrawer() {
    isMobileDrawerOpen.value = true
  }

  function closeMobileDrawer() {
    isMobileDrawerOpen.value = false
  }

  function setModule(module: string) {
    currentModule.value = isModuleEnabled(module) ? module : 'hris'
    isMobileDrawerOpen.value = false
  }

  return {
    isSidebarCollapsed,
    isMobileMenuOpen,
    isMobileDrawerOpen,
    currentModule,
    toggleSidebar,
    setSidebarCollapsed,
    toggleMobileMenu,
    openMobileDrawer,
    closeMobileDrawer,
    setModule,
  }
})
