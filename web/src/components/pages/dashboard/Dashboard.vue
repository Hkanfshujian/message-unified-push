<script setup lang="ts">
import StatCard from '@/components/pages/dashboard/CardNum.vue'
import { FileTextOutlined, SendOutlined, CheckCircleOutlined, CloseCircleOutlined } from '@ant-design/icons-vue'
// import { LineChart } from "@/components/ui/chart-line"
import { onMounted, onUnmounted, reactive, ref, watch } from 'vue';
import { useRoute } from 'vue-router';
import { request } from '@/api/api';
import { toast } from 'vue-sonner';
import { pickDateRangeQuery } from '@/util/routeQuery';
// import VueApexCharts from 'vue3-apexcharts'
import ApexCharts from 'apexcharts'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'

interface SendData {
  day: string
  day_failed_num: number
  day_succ_num: number
  num: number
  succ_num: number
}
interface CateData {
  count_num: number
  way_name: string
}

const route = useRoute()

let state = reactive({
  trendDays: 15,
  basicData: {
    message_total_num: 0,
    today_total_num: 0,
    today_succ_num: 0,
    today_failed_num: 0,
  },
  trendData: {
    latest_send_data: [] as SendData[] | null,
  },
  channelData: {
    way_cate_data: [] as CateData[] | null,
  },
  loading: {
    basic: false,
    trend: false,
    channels: false,
  }
});
const manualTrendDays = ref(false)
let lineChartInstance: ApexCharts | null = null
let pieChartInstance: ApexCharts | null = null
let themeClassObserver: MutationObserver | null = null

const getCssVar = (name: string, fallback: string) => {
  const value = getComputedStyle(document.documentElement).getPropertyValue(name).trim()
  return value || fallback
}

const getMotionMs = (name: '--motion-fast' | '--motion-normal', fallback: number) => {
  const value = parseInt(getCssVar(name, `${fallback}ms`), 10)
  return Number.isFinite(value) ? value : fallback
}

const toIsoDate = (date: Date) => {
  const y = date.getFullYear()
  const m = String(date.getMonth() + 1).padStart(2, '0')
  const d = String(date.getDate()).padStart(2, '0')
  return `${y}-${m}-${d}`
}

const parseTrendDayToIso = (input: string, fallbackYear: number) => {
  if (!input) return ''
  const text = input.trim()
  if (/^\d{4}-\d{2}-\d{2}$/.test(text)) return text
  if (/^\d{2}-\d{2}$/.test(text)) return `${fallbackYear}-${text}`
  const parsed = new Date(text)
  if (Number.isNaN(parsed.getTime())) return ''
  return toIsoDate(parsed)
}

const getTrendDateRange = () => {
  const { startTime, endTime } = pickDateRangeQuery(route.query as Record<string, unknown>)
  const end = endTime ? endTime.split('T')[0] : getTodayDate()
  const endDate = new Date(end)
  if (Number.isNaN(endDate.getTime())) {
    const now = new Date()
    const fallbackEnd = toIsoDate(now)
    const fallbackStartDate = new Date(now)
    fallbackStartDate.setDate(fallbackStartDate.getDate() - Math.max(state.trendDays - 1, 0))
    return { start: toIsoDate(fallbackStartDate), end: fallbackEnd }
  }
  const start = startTime ? startTime.split('T')[0] : ''
  if (start) return { start, end }
  const startDate = new Date(endDate)
  startDate.setDate(startDate.getDate() - Math.max(state.trendDays - 1, 0))
  return { start: toIsoDate(startDate), end }
}

const normalizeTrendSeries = (source: SendData[]) => {
  const { start, end } = getTrendDateRange()
  const fallbackYear = Number(end.slice(0, 4)) || new Date().getFullYear()
  const mapped = new Map<string, SendData>()
  source.forEach((item) => {
    const key = parseTrendDayToIso(item.day, fallbackYear)
    if (!key) return
    mapped.set(key, {
      day: key,
      day_failed_num: Number(item.day_failed_num || 0),
      day_succ_num: Number(item.day_succ_num || 0),
      num: Number(item.num || 0),
      succ_num: Number(item.succ_num || 0)
    })
  })
  const result: SendData[] = []
  const cursor = new Date(start)
  const endDate = new Date(end)
  while (!Number.isNaN(cursor.getTime()) && !Number.isNaN(endDate.getTime()) && cursor <= endDate) {
    const key = toIsoDate(cursor)
    const current = mapped.get(key) || {
      day: key,
      day_failed_num: 0,
      day_succ_num: 0,
      num: 0,
      succ_num: 0
    }
    result.push({
      ...current,
      day: key.slice(5)
    })
    cursor.setDate(cursor.getDate() + 1)
  }
  return result
}

const getChartThemeTokens = () => {
  const isDark = document.documentElement.classList.contains('dark')
  return {
    isDark,
    axisTextColor: isDark ? '#b8c6d9' : '#62728a',
    mutedTextColor: isDark ? '#8ea2bf' : '#7f8ea3',
    gridLineColor: getCssVar('--line-weak', isDark ? 'rgba(148,163,184,0.2)' : 'rgba(100,116,139,0.2)')
  }
}



const TOKEN_ERROR_CODES = [20001, 20002, 20003, 20004, 20005]

const shouldSilenceAuthError = (payload: any) => {
  if (!payload) return false
  const status = payload?.status || payload?.response?.status
  if (status === 401) return true
  const code = payload?.data?.code || payload?.response?.data?.code
  return TOKEN_ERROR_CODES.includes(code)
}

// 获取基础统计数据
const getBasicStatisticData = async () => {
  state.loading.basic = true;
  try {
    const rsp = await request.get('/statistic?type=basic');
    if (rsp && rsp.data && rsp.data.code == 200) {
      state.basicData = rsp.data.data;
    } else {
      if (shouldSilenceAuthError(rsp)) {
        return
      }
      toast.error(rsp?.data?.msg || '获取基础统计数据失败');
    }
  } catch (error) {
    if (shouldSilenceAuthError(error)) {
      return
    }
    toast.error('获取基础统计数据时发生错误');
  } finally {
    state.loading.basic = false;
  }
}

// 获取趋势统计数据
const getTrendStatisticData = async () => {
  state.loading.trend = true;
  try {
    const days = getTrendDays()
    state.trendDays = days;

    const rsp = await request.get(`/statistic?type=trend&days=${days}`);
    if (rsp && rsp.data && rsp.data.code == 200) {
      state.trendData = rsp.data.data;
      // 数据加载完成后重新渲染折线图
      setTimeout(() => {
        renderLineChart();
      }, 100);
    } else {
      if (shouldSilenceAuthError(rsp)) {
        return
      }
      toast.error(rsp?.data?.msg || '获取趋势统计数据失败');
    }
  } catch (error) {
    if (shouldSilenceAuthError(error)) {
      return
    }
    toast.error('获取趋势统计数据时发生错误');
  } finally {
    state.loading.trend = false;
  }
}

// 获取渠道统计数据
const getChannelStatisticData = async () => {
  state.loading.channels = true;
  try {
    const rsp = await request.get('/statistic?type=channels');
    if (rsp && rsp.data && rsp.data.code == 200) {
      state.channelData = rsp.data.data;
      // 数据加载完成后重新渲染饼图
      setTimeout(() => {
        renderPieChart();
      }, 100);
    } else {
      if (shouldSilenceAuthError(rsp)) {
        return
      }
      toast.error(rsp?.data?.msg || '获取渠道统计数据失败');
    }
  } catch (error) {
    if (shouldSilenceAuthError(error)) {
      return
    }
    toast.error('获取渠道统计数据时发生错误');
  } finally {
    state.loading.channels = false;
  }
}

// 并行加载所有统计数据
const loadAllStatisticData = async () => {
  await Promise.all([
    getBasicStatisticData(),
    getTrendStatisticData(),
    getChannelStatisticData()
  ]);
}

const renderLineChart = () => {
  const latestSendData = normalizeTrendSeries(state.trendData.latest_send_data || [])
  const { isDark, axisTextColor, mutedTextColor, gridLineColor } = getChartThemeTokens()
  const motionNormal = getMotionMs('--motion-normal', 220)
  const motionFast = getMotionMs('--motion-fast', 150)
  const isSmallScreen = window.innerWidth < 768
  const maxSeriesValue = latestSendData.reduce((acc, item) => {
    const localMax = Math.max(item.num || 0, item.day_succ_num || 0, item.day_failed_num || 0)
    return Math.max(acc, localMax)
  }, 0)
  const yAxisMax = maxSeriesValue <= 5 ? 5 : Math.ceil(maxSeriesValue / 5) * 5

  const options = {
    series: [
      {
        name: '发送总数',
        data: latestSendData.length > 0
          ? latestSendData.map(item => item.num || 0)
          : []
      },
      {
        name: '发送成功数',
        data: latestSendData.length > 0
          ? latestSendData.map(item => item.day_succ_num || 0)
          : []
      },
      {
        name: '发送失败数',
        data: latestSendData.length > 0
          ? latestSendData.map(item => item.day_failed_num || 0)
          : []
      },
    ],
    chart: {
      type: 'area',
      height: 350,
      toolbar: { show: false },
      background: 'transparent',
      animations: {
        enabled: true,
        easing: 'easeinout',
        speed: motionNormal,
        animateGradually: {
          enabled: true,
          delay: 70
        },
        dynamicAnimation: {
          enabled: true,
          speed: motionFast
        }
      }
    },
    states: {
      hover: {
        filter: {
          type: 'none'
        }
      },
      active: {
        filter: {
          type: 'none'
        }
      }
    },
    stroke: {
      curve: 'smooth',
      width: [2.6, 2.2, 2.1],
      colors: ['#3b82f6', '#22c55e', '#ef4444'],
      lineCap: 'round'
    },
    dataLabels: {
      enabled: false
    },
    markers: {
      size: 0,
      colors: ['#3b82f6', '#22c55e', '#ef4444'],
      strokeColors: isDark ? '#0f172a' : '#fff',
      strokeWidth: 0,
      hover: {
        size: 0,
        sizeOffset: 0
      }
    },
    xaxis: {
      categories: latestSendData.length > 0
        ? latestSendData.map(item => item.day)
        : [],
      axisBorder: {
        show: false
      },
      axisTicks: {
        show: false
      },
      labels: {
        style: {
          colors: axisTextColor,
          fontSize: '12px',
          fontFamily: 'Inter, sans-serif'
        }
      },
      tooltip: {
        enabled: false
      }
    },
    yaxis: {
      min: 0,
      max: yAxisMax,
      forceNiceScale: true,
      labels: {
        style: {
          colors: axisTextColor,
          fontSize: '12px',
          fontFamily: 'Inter, sans-serif'
        },
        formatter: function (val: number) {
          return val + ' 条'
        }
      }
    },
    colors: ['#3b82f6', '#22c55e', '#ef4444'], // 蓝色表示总数，绿色表示成功，红色表示失败
    fill: {
      type: 'gradient',
      gradient: {
        shade: 'light',
        type: 'vertical',
        shadeIntensity: 0.65,
        inverseColors: false,
        opacityFrom: [0.36, 0.32, 0.28],
        opacityTo: [0.08, 0.07, 0.06],
        stops: [0, 100]
      }
    },
    grid: {
      borderColor: gridLineColor,
      strokeDashArray: 4,
      xaxis: {
        lines: {
          show: false
        }
      },
      yaxis: {
        lines: {
          show: !isSmallScreen
        }
      },
      padding: {
        top: 0,
        right: 0,
        bottom: 0,
        left: 0
      }
    },
    legend: {
      position: 'top',
      horizontalAlign: 'left',
      floating: false,
      offsetY: -4,
      offsetX: 0,
      fontSize: '12px',
      fontWeight: 500,
      fontFamily: 'Inter, sans-serif',
      labels: {
        colors: mutedTextColor
      },
      markers: {
        width: 7,
        height: 7,
        radius: 4
      },
      itemMargin: {
        horizontal: 10,
        vertical: 2
      }
    },
    tooltip: {
      enabled: true,
      shared: true,
      intersect: false,
      theme: document.documentElement.classList.contains('dark') ? 'dark' : 'light',
      style: {
        fontSize: '12px',
        fontFamily: 'Inter, sans-serif'
      },
      x: {
        show: true,
        format: 'MM/dd'
      },
      y: {
        formatter: function (val: number, { seriesIndex: _seriesIndex }: { seriesIndex: number }) {
          return val + ' 条'
        }
      },
      marker: {
        show: true
      },
      custom: function ({ series, seriesIndex: _seriesIndex, dataPointIndex, w }: { series: number[][], seriesIndex: number, dataPointIndex: number, w: any }) {
        const totalCount = series[0][dataPointIndex];
        const successCount = series[1][dataPointIndex];
        const failedCount = series[2][dataPointIndex];
        const total = successCount + failedCount;
        const successRate = total > 0 ? ((successCount / total) * 100).toFixed(1) : '0.0';

        const containerCls = 'bg-card text-foreground p-2.5 rounded-lg shadow-lg border weak-divider';
        const labelMutedCls = 'text-sm text-muted-foreground';
        const valueStrongCls = 'text-sm font-medium text-foreground';
        const successRateCls = 'text-sm font-medium';

        return `
          <div class="${containerCls}">
            <div class="font-medium mb-2">${w.globals.categoryLabels[dataPointIndex]}</div>
            <div class="space-y-1">
              <div class="flex items-center justify-between">
                <span class="flex items-center">
                  <span class="w-2 h-2 bg-blue-500 rounded-full mr-2"></span>
                  <span class="${labelMutedCls}">总计:</span>
                </span>
                <span class="${valueStrongCls}">${totalCount} 条</span>
              </div>
              <div class="flex items-center justify-between">
                <span class="flex items-center">
                  <span class="w-2 h-2 bg-green-500 rounded-full mr-2"></span>
                  <span class="${labelMutedCls}">成功:</span>
                </span>
                <span class="${valueStrongCls}">${successCount} 条</span>
              </div>
              <div class="flex items-center justify-between">
                <span class="flex items-center">
                  <span class="w-2 h-2 bg-red-500 rounded-full mr-2"></span>
                  <span class="${labelMutedCls}">失败:</span>
                </span>
                <span class="${valueStrongCls}">${failedCount} 条</span>
              </div>
              <div class="border-t weak-divider pt-1 mt-2">
                <div class="flex items-center justify-between">
                  <span class="${labelMutedCls}">成功率:</span>
                  <span class="${successRateCls}">${successRate}%</span>
                </div>
              </div>
            </div>
          </div>
        `;
      }
    },
    responsive: [{
      breakpoint: 768,
      options: {
        chart: {
          height: 300
        },
        legend: {
          position: 'top',
          horizontalAlign: 'right',
          floating: false,
          offsetY: -6,
          offsetX: 0,
          fontSize: '10px',
          fontWeight: 500,
          markers: {
            width: 6,
            height: 6,
            radius: 3
          },
          itemMargin: {
            horizontal: 5,
            vertical: 0
          }
        }
      }
    }]
  }
  lineChartInstance?.destroy()
  lineChartInstance = new ApexCharts(document.querySelector("#sales-chart"), options)
  lineChartInstance.render();

}

const renderPieChart = () => {
  const wayCateData = state.channelData.way_cate_data || [];
  const { isDark, mutedTextColor } = getChartThemeTokens()
  const motionNormal = getMotionMs('--motion-normal', 220)
  const motionFast = getMotionMs('--motion-fast', 150)
  const options = {
    series: wayCateData.length > 0
      ? wayCateData.map(item => item.count_num)
      : [],
    chart: {
      type: 'pie',
      height: 350,
      toolbar: { show: false },
      background: 'transparent',
      animations: {
        enabled: true,
        easing: 'easeinout',
        speed: motionNormal,
        animateGradually: {
          enabled: true,
          delay: 80
        },
        dynamicAnimation: {
          enabled: true,
          speed: motionFast
        }
      }
    },
    labels: wayCateData.length > 0
      ? wayCateData.map(item => item.way_name)
      : [],
    colors: ['#3b82f6', '#22c55e', '#f59e0b', '#ef4444', '#8b5cf6'],
    legend: {
      position: 'bottom',
      fontSize: '11px',
      fontWeight: 500,
      fontFamily: 'Inter, sans-serif',
      labels: {
        colors: mutedTextColor
      },
      markers: {
        width: 8,
        height: 8,
        radius: 4
      },
      itemMargin: {
        horizontal: 10,
        vertical: 4
      }
    },
    plotOptions: {
      pie: {
        expandOnClick: false,
        donut: {
          size: '0%'
        },
        offsetY: 2
      }
    },
    dataLabels: {
      enabled: true,
      formatter: function (val: number) {
        return val.toFixed(1) + '%'
      },
      style: {
        fontSize: '10px',
        fontFamily: 'Inter, sans-serif',
        fontWeight: 'bold'
      }
    },
    tooltip: {
      enabled: true,
      theme: isDark ? 'dark' : 'light',
      style: {
        fontSize: '12px',
        fontFamily: 'Inter, sans-serif'
      },
      y: {
        formatter: function (val: number) {
          return val + ' 条'
        }
      }
    },
    responsive: [{
      breakpoint: 768,
      options: {
        chart: {
          height: 300
        },
        legend: {
          position: 'bottom'
        }
      }
    }]
  }
  pieChartInstance?.destroy()
  pieChartInstance = new ApexCharts(document.querySelector("#pie-chart"), options)
  pieChartInstance.render();
}

// 获取今日日期
const getTodayDate = () => {
  const today = new Date();
  return today.toISOString().split('T')[0]; // 返回 YYYY-MM-DD 格式
}

const getTrendDays = () => {
  const { startTime, endTime } = pickDateRangeQuery(route.query as Record<string, unknown>)
  const start = startTime ? startTime.split('T')[0] : ''
  const end = endTime ? endTime.split('T')[0] : ''
  if (start && end) {
    const startTime = new Date(start).getTime()
    const endTime = new Date(end).getTime()
    if (!Number.isNaN(startTime) && !Number.isNaN(endTime) && endTime >= startTime) {
      const dayMs = 24 * 60 * 60 * 1000
      const diffDays = Math.floor((endTime - startTime) / dayMs) + 1
      return Math.min(Math.max(diffDays, 1), 90)
    }
  }
  if (manualTrendDays.value) return state.trendDays
  return 15
}

const onTrendDaysChange = () => {
  manualTrendDays.value = true
  getTrendStatisticData()
}

onMounted(() => {
  loadAllStatisticData();
  if (typeof MutationObserver !== 'undefined') {
    themeClassObserver = new MutationObserver((mutations) => {
      const classChanged = mutations.some(
        (mutation) => mutation.type === 'attributes' && mutation.attributeName === 'class'
      )
      if (!classChanged) return
      renderLineChart()
      renderPieChart()
    })
    themeClassObserver.observe(document.documentElement, {
      attributes: true,
      attributeFilter: ['class']
    })
  }
})

watch(
  () => [route.query.start_time, route.query.end_time, route.query.start_date, route.query.end_date],
  () => {
    manualTrendDays.value = false
    getTrendStatisticData()
  }
)

onUnmounted(() => {
  lineChartInstance?.destroy()
  pieChartInstance?.destroy()
  themeClassObserver?.disconnect()
  themeClassObserver = null
})



</script>

<template>
  <div class="w-full grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">

    <StatCard title="发送日志数" :value="state.basicData.message_total_num" description="总发送日志" trend-text="+8.2%" trend-type="up" tone="blue" :icon="FileTextOutlined"
      route-path="/logs/task" />
    <StatCard title="今日发送数" :value="state.basicData.today_total_num" description="较昨日" trend-text="+3.1%" trend-type="up" tone="teal" :icon="SendOutlined"
      :route-path="`/logs/task?query=${encodeURIComponent(JSON.stringify({ day_created_on: getTodayDate() }))}`" />
    <StatCard title="今日成功数" :value="state.basicData.today_succ_num" description="成功率" trend-text="+1.7%" trend-type="up" tone="purple" :icon="CheckCircleOutlined"
      :route-path="`/logs/task?query=${encodeURIComponent(JSON.stringify({ status: '1', day_created_on: getTodayDate() }))}`" />
    <StatCard title="今日失败数" :value="state.basicData.today_failed_num" description="失败率" trend-text="-0.6%" trend-type="down" tone="red" :icon="CloseCircleOutlined"
      :route-path="`/logs/task?query=${encodeURIComponent(JSON.stringify({ status: '0', day_created_on: getTodayDate() }))}`" />
  </div>

  <!-- 折线图 -->
  <!-- <LineChart
    :data="data"
    index="year"
    :categories="['Export Growth Rate', 'Import Growth Rate']"
    :y-formatter="(tick, i) => {
      return typeof tick === 'number'
        ? `$ ${new Intl.NumberFormat('us').format(tick).toString()}`
        : ''
    }"
  /> -->

  <!-- 图表区域 -->
  <div class="w-full mt-6 grid grid-cols-1 lg:grid-cols-10 gap-6">
    <!-- 折线图 -->
    <Card class="w-full lg:col-span-7 weak-divider shadow-sm">
      <CardHeader class="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
        <div>
          <CardTitle>消息发送趋势</CardTitle>
          <CardDescription>最近{{ state.trendDays }}天的发送情况统计</CardDescription>
        </div>
        <div class="flex items-center gap-2">
          <select
            v-model.number="state.trendDays"
            class="h-8 rounded-md border weak-divider bg-background px-2 text-xs text-foreground"
            @change="onTrendDaysChange"
          >
            <option :value="7">近7天</option>
            <option :value="15">近15天</option>
            <option :value="30">近30天</option>
            <option :value="60">近60天</option>
          </select>
          <button
            class="h-8 rounded-md border weak-divider bg-background px-3 text-xs text-muted-foreground hover:text-foreground hover:bg-accent transition-colors disabled:opacity-50"
            :disabled="state.loading.trend"
            @click="getTrendStatisticData"
          >
            刷新
          </button>
        </div>
      </CardHeader>
      <CardContent>
        <div id="sales-chart" class="w-full h-[350px]"></div>
      </CardContent>
    </Card>

    <!-- 饼图 -->
    <Card class="w-full lg:col-span-3 weak-divider shadow-sm">
      <CardHeader>
        <CardTitle>发送渠道分布</CardTitle>
        <CardDescription>各发送渠道的使用情况统计</CardDescription>
      </CardHeader>
      <CardContent>
        <div id="pie-chart" class="w-full h-[350px]"></div>
      </CardContent>
    </Card>
  </div>


</template>
