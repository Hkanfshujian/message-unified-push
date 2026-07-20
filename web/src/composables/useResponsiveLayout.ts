import { computed, onMounted, onUnmounted, watch } from 'vue'
import { storeToRefs } from 'pinia'
import { useAppStore } from '@/store/app'
import { getContentOffset } from '@/util/layout'

export const useResponsiveLayout = () => {
  const appStore = useAppStore()
  const { navigation } = storeToRefs(appStore)

  const updateViewport = () => {
    if (typeof window === 'undefined') return
    appStore.setViewportSize(window.innerWidth, window.innerHeight)
  }

  onMounted(() => {
    updateViewport()
    window.addEventListener('resize', updateViewport)
  })

  onUnmounted(() => {
    if (typeof window !== 'undefined') window.removeEventListener('resize', updateViewport)
  })

  const layoutState = computed(() => navigation.value.layout)
  const isMobileLayout = computed(() => layoutState.value.device === 'mobile')
  const isSidebarCollapsed = computed(() => layoutState.value.sidebarCollapsed)
  const isMobileSidebarVisible = computed(() => layoutState.value.mobileSidebarVisible)
  const contentOffset = computed(() => getContentOffset(layoutState.value.viewportClass, layoutState.value.sidebarCollapsed))

  const toggleSidebar = () => {
    if (layoutState.value.device === 'mobile') {
      appStore.setMobileSidebarVisible(!layoutState.value.mobileSidebarVisible)
      return
    }
    appStore.setSidebarCollapsed(!layoutState.value.sidebarCollapsed)
  }

  const closeMobileSidebar = () => {
    appStore.closeMobileSidebar()
  }

  const openMobileSidebar = () => {
    appStore.openMobileSidebar()
  }

  watch(
    () => layoutState.value.device,
    (device, previousDevice) => {
      if (device !== previousDevice && device !== 'mobile') {
        appStore.closeMobileSidebar()
      }
    }
  )

  return {
    layoutState,
    isMobileLayout,
    isSidebarCollapsed,
    isMobileSidebarVisible,
    contentOffset,
    toggleSidebar,
    openMobileSidebar,
    closeMobileSidebar,
    updateViewport
  }
}
