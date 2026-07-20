import type { ECharts, EChartsOption } from 'echarts'
import { computed, nextTick, onBeforeUnmount, onMounted, ref, shallowRef, type Ref } from 'vue'

export const useECharts = (container: Ref<HTMLElement | null>) => {
  const chart = shallowRef<ECharts | null>(null)
  const loading = ref(false)
  const error = ref('')
  const empty = ref(false)
  const currentOption = shallowRef<EChartsOption | null>(null)
  let resizeObserver: ResizeObserver | null = null

  const state = computed(() => ({
    loading: loading.value,
    error: error.value,
    empty: empty.value
  }))

  const init = async () => {
    await nextTick()
    if (!container.value || chart.value) return chart.value
    const echarts = await import('echarts')
    chart.value = echarts.init(container.value)
    if (currentOption.value) {
      chart.value.setOption(currentOption.value, true)
    }
    if (typeof ResizeObserver !== 'undefined' && !resizeObserver) {
      resizeObserver = new ResizeObserver(() => resize())
      resizeObserver.observe(container.value)
    }
    return chart.value
  }

  const setOption = async (option: EChartsOption, isEmpty = false) => {
    loading.value = false
    error.value = ''
    empty.value = isEmpty
    currentOption.value = option
    const instance = await init()
    instance?.setOption(option, true)
  }

  const setLoading = (value: boolean) => {
    loading.value = value
    if (value) {
      chart.value?.showLoading('default', {
        color: '#3b82f6',
        textColor: '#64748b',
        maskColor: 'rgba(255,255,255,0.72)'
      })
    }
    else chart.value?.hideLoading()
  }

  const setError = (message: string) => {
    error.value = message
    loading.value = false
    chart.value?.hideLoading()
  }

  const resize = () => {
    chart.value?.resize()
    if (currentOption.value && chart.value) {
      chart.value.setOption(currentOption.value, false)
    }
  }

  const dispose = () => {
    resizeObserver?.disconnect()
    resizeObserver = null
    chart.value?.dispose()
    chart.value = null
  }

  if (typeof window !== 'undefined') {
    window.addEventListener('resize', resize)
  }

  onMounted(() => {
    void init()
  })

  onBeforeUnmount(() => {
    if (typeof window !== 'undefined') {
      window.removeEventListener('resize', resize)
    }
    dispose()
  })

  return {
    chart,
    state,
    currentOption,
    init,
    setOption,
    setLoading,
    setError,
    resize,
    dispose
  }
}
