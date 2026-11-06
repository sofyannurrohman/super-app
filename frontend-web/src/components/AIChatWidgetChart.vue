<script setup lang="ts">
import { ref, computed } from 'vue'
import { use } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import { LineChart, BarChart, PieChart } from 'echarts/charts'
import {
  GridComponent,
  TooltipComponent,
  LegendComponent,
  TitleComponent,
  ToolboxComponent,
  DataZoomComponent
} from 'echarts/components'
import VChart from 'vue-echarts'
import {
  BarChart2,
  BarChart3,
  PieChart as PieChartIcon,
  TrendingUp,
  Table as TableIcon,
  Maximize2,
  Download,
  X,
  Sparkles,
  ArrowRightLeft
} from 'lucide-vue-next'
import type { AIWidget } from '@/api/ai'

use([
  CanvasRenderer,
  LineChart,
  BarChart,
  PieChart,
  GridComponent,
  TooltipComponent,
  LegendComponent,
  TitleComponent,
  ToolboxComponent,
  DataZoomComponent
])

const props = defineProps<{
  widget: AIWidget
}>()

// State
const activeView = ref<'chart' | 'table'>('chart')
const isModalOpen = ref(false)
const modalChartType = ref<'auto' | 'bar' | 'bar_horizontal' | 'line' | 'pie'>('auto')
const miniVChartRef = ref<any>(null)
const modalVChartRef = ref<any>(null)

// Formatters
const formatRupiah = (val: number | string | undefined) => {
  if (val === undefined || val === null) return 'Rp 0'
  const num = typeof val === 'string' ? parseFloat(val) : val
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    maximumFractionDigits: 0
  }).format(num || 0)
}

const formatNumber = (val: number | string | undefined) => {
  if (val === undefined || val === null) return '0'
  const num = typeof val === 'string' ? parseFloat(val) : val
  return new Intl.NumberFormat('id-ID').format(num || 0)
}

const formatCompact = (val: number) => {
  if (val >= 1_000_000_000) return (val / 1_000_000_000).toFixed(1) + 'M'
  if (val >= 1_000_000) return (val / 1_000_000).toFixed(0) + 'jt'
  if (val >= 1_000) return (val / 1_000).toFixed(0) + 'rb'
  return String(val)
}

// KPI / Summary Badges Computation
const summaryKpis = computed(() => {
  const kpis: { label: string; value: string; colorClass?: string }[] = []
  const data = props.widget.data || {}

  if (data.total_revenue !== undefined) {
    kpis.push({
      label: 'Total Omzet',
      value: formatRupiah(data.total_revenue),
      colorClass: 'text-indigo-700 bg-indigo-50 border-indigo-200'
    })
  }

  if (data.total_transactions !== undefined) {
    kpis.push({
      label: 'Total Transaksi',
      value: `${formatNumber(data.total_transactions)} trx`,
      colorClass: 'text-purple-700 bg-purple-50 border-purple-200'
    })
  }

  if (data.total_orders !== undefined) {
    kpis.push({
      label: 'Total Pesanan',
      value: `${formatNumber(data.total_orders)} order`,
      colorClass: 'text-emerald-700 bg-emerald-50 border-emerald-200'
    })
  }

  if (data.average_order_value !== undefined) {
    kpis.push({
      label: 'AOV',
      value: formatRupiah(data.average_order_value),
      colorClass: 'text-amber-700 bg-amber-50 border-amber-200'
    })
  }

  // Jika tidak ada KPI eksplisit tapi ada pie_data / series, hitung total & top
  if (kpis.length === 0) {
    if (props.widget.chart_type === 'pie' && Array.isArray(data.pie_data) && data.pie_data.length > 0) {
      const sum = data.pie_data.reduce((acc: number, item: any) => acc + (Number(item.value) || 0), 0)
      const topItem = [...data.pie_data].sort((a: any, b: any) => (b.value || 0) - (a.value || 0))[0]
      if (sum > 0) {
        kpis.push({
          label: 'Total Nilai',
          value: formatRupiah(sum),
          colorClass: 'text-indigo-700 bg-indigo-50 border-indigo-200'
        })
      }
      if (topItem) {
        kpis.push({
          label: 'Tertinggi',
          value: `${topItem.name} (${topItem.value > 10000 ? formatRupiah(topItem.value) : formatNumber(topItem.value)})`,
          colorClass: 'text-emerald-700 bg-emerald-50 border-emerald-200'
        })
      }
    } else if (Array.isArray(data.categories) && Array.isArray(data.series) && data.series.length > 0) {
      const firstSeries = data.series[0]?.data || []
      const sum = firstSeries.reduce((acc: number, val: any) => acc + (Number(val) || 0), 0)
      if (sum > 0) {
        kpis.push({
          label: 'Total',
          value: formatRupiah(sum),
          colorClass: 'text-indigo-700 bg-indigo-50 border-indigo-200'
        })
      }
    }
  }

  return kpis
})

// Tabular Data Extraction
const tableRows = computed(() => {
  const data = props.widget.data || {}
  if (props.widget.chart_type === 'pie' && Array.isArray(data.pie_data)) {
    const sum = data.pie_data.reduce((acc: number, item: any) => acc + (Number(item.value) || 0), 0)
    return data.pie_data.map((item: any, idx: number) => {
      const val = Number(item.value) || 0
      const pct = sum > 0 ? ((val / sum) * 100).toFixed(1) + '%' : '0%'
      return {
        id: idx,
        name: item.name || `Item ${idx + 1}`,
        rawVal: val,
        formattedVal: val > 10000 ? formatRupiah(val) : formatNumber(val),
        percentage: pct
      }
    })
  }

  // Categories + Series
  if (Array.isArray(data.categories) && Array.isArray(data.series) && data.series.length > 0) {
    const series0 = data.series[0]?.data || []
    const sum = series0.reduce((acc: number, val: any) => acc + (Number(val) || 0), 0)
    return data.categories.map((cat: string, idx: number) => {
      const val = Number(series0[idx]) || 0
      const pct = sum > 0 ? ((val / sum) * 100).toFixed(1) + '%' : '0%'
      return {
        id: idx,
        name: cat,
        rawVal: val,
        formattedVal: val > 10000 ? formatRupiah(val) : formatNumber(val),
        percentage: pct
      }
    })
  }

  return []
})

// Palette Colors
const vibrantPalette = [
  '#6366f1', // Indigo
  '#8b5cf6', // Purple
  '#ec4899', // Pink
  '#06b6d4', // Cyan
  '#10b981', // Emerald
  '#f59e0b', // Amber
  '#3b82f6', // Blue
  '#f43f5e'  // Rose
]

// Determine if horizontal bar chart is better for mini view
const shouldUseHorizontalBar = computed(() => {
  if (props.widget.chart_type === 'pie') return false
  const categories = props.widget.data?.categories || []
  // Jika ada kategori dengan nama > 7 karakter atau kategori > 3, horizontal bar jauh lebih rapi & tidak squeeze
  return categories.some((c: string) => c.length > 7) || categories.length >= 4
})

// Mini Chart Options
const miniChartOption = computed(() => {
  const widget = props.widget
  const data = widget.data || {}

  if (widget.chart_type === 'pie') {
    const pieData = data.pie_data || []
    return {
      color: vibrantPalette,
      tooltip: {
        trigger: 'item',
        formatter: (params: any) => {
          const valStr = params.value > 10000 ? formatRupiah(params.value) : formatNumber(params.value)
          return `<div class="font-sans text-xs"><strong>${params.name}</strong><br/>${valStr} (${params.percent}%)</div>`
        }
      },
      legend: {
        bottom: 0,
        left: 'center',
        itemWidth: 10,
        itemHeight: 10,
        textStyle: { fontSize: 10, color: '#64748b' }
      },
      series: [
        {
          name: widget.title || 'Data',
          type: 'pie',
          radius: ['42%', '70%'],
          center: ['50%', '42%'],
          avoidLabelOverlap: true,
          itemStyle: {
            borderRadius: 5,
            borderColor: '#fff',
            borderWidth: 2
          },
          label: {
            show: false
          },
          emphasis: {
            scale: true,
            scaleSize: 6
          },
          data: pieData
        }
      ]
    }
  }

  // Bar or Line
  const categories = data.categories || []
  const rawSeries = data.series || []

  // Check if Horizontal Bar
  if (shouldUseHorizontalBar.value) {
    const reversedCats = [...categories].reverse()
    const series = rawSeries.map((s: any) => {
      const reversedData = [...(s.data || [])].reverse()
      return {
        name: s.name || 'Nilai',
        type: 'bar',
        data: reversedData,
        barMaxWidth: 18,
        itemStyle: {
          borderRadius: [0, 6, 6, 0],
          color: {
            type: 'linear',
            x: 0,
            y: 0,
            x2: 1,
            y2: 0,
            colorStops: [
              { offset: 0, color: '#6366f1' },
              { offset: 1, color: '#8b5cf6' }
            ]
          }
        },
        label: {
          show: true,
          position: 'right',
          fontSize: 9,
          color: '#475569',
          formatter: (params: any) => formatCompact(params.value)
        }
      }
    })

    return {
      tooltip: {
        trigger: 'axis',
        axisPointer: { type: 'shadow' },
        formatter: (params: any) => {
          const item = params[0]
          if (!item) return ''
          const val = item.value > 10000 ? formatRupiah(item.value) : formatNumber(item.value)
          return `<div class="font-sans text-xs"><strong>${item.name}</strong><br/>${val}</div>`
        }
      },
      grid: {
        top: '6%',
        left: '3%',
        right: '16%',
        bottom: '4%',
        containLabel: true
      },
      xAxis: {
        type: 'value',
        show: false
      },
      yAxis: {
        type: 'category',
        data: reversedCats,
        axisTick: { show: false },
        axisLine: { show: false },
        axisLabel: {
          fontSize: 10,
          color: '#334155',
          formatter: (val: string) => (val.length > 13 ? val.substring(0, 11) + '…' : val)
        }
      },
      series
    }
  }

  // Default Vertical Bar
  const series = rawSeries.map((s: any) => ({
    name: s.name || 'Nilai',
    type: 'bar',
    data: s.data || [],
    barMaxWidth: 24,
    itemStyle: {
      borderRadius: [6, 6, 0, 0],
      color: {
        type: 'linear',
        x: 0,
        y: 0,
        x2: 0,
        y2: 1,
        colorStops: [
          { offset: 0, color: '#6366f1' },
          { offset: 1, color: '#a855f7' }
        ]
      }
    }
  }))

  return {
    tooltip: {
      trigger: 'axis',
      axisPointer: { type: 'shadow' },
      formatter: (params: any) => {
        const item = params[0]
        if (!item) return ''
        const val = item.value > 10000 ? formatRupiah(item.value) : formatNumber(item.value)
        return `<div class="font-sans text-xs"><strong>${item.name}</strong><br/>${val}</div>`
      }
    },
    grid: {
      top: '10%',
      left: '2%',
      right: '3%',
      bottom: '8%',
      containLabel: true
    },
    xAxis: {
      type: 'category',
      data: categories,
      axisTick: { alignWithLabel: true },
      axisLabel: {
        fontSize: 9,
        interval: 0,
        color: '#64748b'
      }
    },
    yAxis: {
      type: 'value',
      axisLabel: {
        fontSize: 9,
        color: '#94a3b8',
        formatter: (val: number) => formatCompact(val)
      },
      splitLine: {
        lineStyle: { type: 'dashed', color: '#f1f5f9' }
      }
    },
    series
  }
})

// Modal (Expanded View) Chart Options with Zoom and dynamic switching
const modalChartOption = computed(() => {
  const widget = props.widget
  const data = widget.data || {}
  const targetType =
    modalChartType.value === 'auto'
      ? widget.chart_type === 'pie'
        ? 'pie'
        : 'bar'
      : modalChartType.value

  if (targetType === 'pie') {
    const pieData = data.pie_data || []
    return {
      color: vibrantPalette,
      title: {
        text: widget.title || 'Distribusi Data',
        subtext: widget.subtitle || '',
        left: 'center',
        top: '2%',
        textStyle: { fontSize: 16, fontWeight: 'bold', color: '#1e293b' }
      },
      tooltip: {
        trigger: 'item',
        formatter: (params: any) => {
          const valStr = params.value > 10000 ? formatRupiah(params.value) : formatNumber(params.value)
          return `<div class="font-sans text-sm"><strong>${params.name}</strong><br/>${valStr} <strong>(${params.percent}%)</strong></div>`
        }
      },
      legend: {
        orient: 'vertical',
        right: '4%',
        top: 'middle',
        itemGap: 14,
        textStyle: { fontSize: 12, color: '#475569' }
      },
      series: [
        {
          name: widget.title || 'Data',
          type: 'pie',
          radius: ['45%', '72%'],
          center: ['40%', '55%'],
          avoidLabelOverlap: true,
          itemStyle: {
            borderRadius: 8,
            borderColor: '#fff',
            borderWidth: 3
          },
          label: {
            show: true,
            formatter: '{b}: {d}%',
            fontSize: 11,
            color: '#334155'
          },
          emphasis: {
            scale: true,
            scaleSize: 10,
            label: {
              show: true,
              fontSize: 13,
              fontWeight: 'bold'
            }
          },
          data: pieData
        }
      ]
    }
  }

  // Categories & Series Data
  const categories = data.categories || []
  const rawSeries = data.series || []

  if (targetType === 'bar_horizontal') {
    const reversedCats = [...categories].reverse()
    const series = rawSeries.map((s: any) => {
      const reversedData = [...(s.data || [])].reverse()
      return {
        name: s.name || 'Nilai',
        type: 'bar',
        data: reversedData,
        barMaxWidth: 26,
        itemStyle: {
          borderRadius: [0, 8, 8, 0],
          color: {
            type: 'linear',
            x: 0,
            y: 0,
            x2: 1,
            y2: 0,
            colorStops: [
              { offset: 0, color: '#4f46e5' },
              { offset: 1, color: '#9333ea' }
            ]
          }
        },
        label: {
          show: true,
          position: 'right',
          fontSize: 11,
          color: '#334155',
          formatter: (params: any) =>
            params.value > 10000 ? formatRupiah(params.value) : formatNumber(params.value)
        }
      }
    })

    return {
      title: {
        text: widget.title || 'Analisis Grafik',
        subtext: widget.subtitle || '',
        left: 'left',
        top: '2%',
        textStyle: { fontSize: 16, fontWeight: 'bold', color: '#1e293b' }
      },
      tooltip: {
        trigger: 'axis',
        axisPointer: { type: 'shadow' },
        formatter: (params: any) => {
          const item = params[0]
          if (!item) return ''
          const val = item.value > 10000 ? formatRupiah(item.value) : formatNumber(item.value)
          return `<div class="font-sans text-sm"><strong>${item.name}</strong><br/>${val}</div>`
        }
      },
      grid: {
        top: '16%',
        left: '3%',
        right: '16%',
        bottom: '8%',
        containLabel: true
      },
      xAxis: {
        type: 'value',
        axisLabel: {
          formatter: (val: number) => formatCompact(val)
        },
        splitLine: { lineStyle: { type: 'dashed', color: '#e2e8f0' } }
      },
      yAxis: {
        type: 'category',
        data: reversedCats,
        axisLabel: {
          fontSize: 12,
          fontWeight: 500,
          color: '#1e293b'
        }
      },
      series
    }
  }

  // Bar or Line Vertical
  const isLine = targetType === 'line'
  const series = rawSeries.map((s: any) => ({
    name: s.name || 'Nilai',
    type: isLine ? 'line' : 'bar',
    smooth: true,
    data: s.data || [],
    barMaxWidth: 36,
    itemStyle: {
      borderRadius: isLine ? undefined : [8, 8, 0, 0],
      color: isLine
        ? '#6366f1'
        : {
            type: 'linear',
            x: 0,
            y: 0,
            x2: 0,
            y2: 1,
            colorStops: [
              { offset: 0, color: '#6366f1' },
              { offset: 1, color: '#8b5cf6' }
            ]
          }
    },
    areaStyle: isLine
      ? {
          color: {
            type: 'linear',
            x: 0,
            y: 0,
            x2: 0,
            y2: 1,
            colorStops: [
              { offset: 0, color: 'rgba(99, 102, 241, 0.35)' },
              { offset: 1, color: 'rgba(99, 102, 241, 0.02)' }
            ]
          }
        }
      : undefined
  }))

  return {
    title: {
      text: widget.title || 'Analisis Grafik',
      subtext: widget.subtitle || '',
      left: 'left',
      top: '2%',
      textStyle: { fontSize: 16, fontWeight: 'bold', color: '#1e293b' }
    },
    tooltip: {
      trigger: 'axis',
      axisPointer: { type: isLine ? 'line' : 'shadow' },
      formatter: (params: any) => {
        const item = params[0]
        if (!item) return ''
        const val = item.value > 10000 ? formatRupiah(item.value) : formatNumber(item.value)
        return `<div class="font-sans text-sm"><strong>${item.name}</strong><br/>${val}</div>`
      }
    },
    grid: {
      top: '18%',
      left: '3%',
      right: '4%',
      bottom: categories.length > 6 ? '18%' : '10%',
      containLabel: true
    },
    xAxis: {
      type: 'category',
      data: categories,
      axisLabel: {
        fontSize: 11,
        color: '#475569',
        interval: 0,
        rotate: categories.length > 5 ? 25 : 0
      }
    },
    yAxis: {
      type: 'value',
      axisLabel: {
        fontSize: 11,
        color: '#64748b',
        formatter: (val: number) => formatCompact(val)
      },
      splitLine: { lineStyle: { type: 'dashed', color: '#e2e8f0' } }
    },
    dataZoom:
      categories.length > 7
        ? [
            {
              type: 'slider',
              start: 0,
              end: Math.min(100, Math.round((7 / categories.length) * 100)),
              height: 20,
              bottom: 4
            },
            {
              type: 'inside'
            }
          ]
        : undefined,
    series
  }
})

// Actions
const downloadChartImage = (isModal = false) => {
  const chartInstance = isModal ? modalVChartRef.value : miniVChartRef.value
  if (!chartInstance) return

  try {
    const dataUrl = chartInstance.getDataURL({
      type: 'png',
      pixelRatio: 2,
      backgroundColor: '#ffffff'
    })
    const link = document.createElement('a')
    link.download = `${props.widget.title || 'grafik-analisis'}-${Date.now()}.png`
    link.href = dataUrl
    link.click()
  } catch (err) {
    console.error('Failed to download chart', err)
  }
}
</script>

<template>
  <div class="rounded-xl border border-indigo-100/90 bg-gradient-to-b from-indigo-50/40 via-white to-slate-50/60 p-2.5 shadow-xs transition hover:border-indigo-200">
    <!-- Header: Title & Action Controls -->
    <div class="mb-2 flex items-center justify-between gap-2 border-b border-indigo-100/60 pb-1.5">
      <div class="min-w-0 flex-1">
        <p class="font-semibold text-slate-800 text-[11px] flex items-center gap-1.5 truncate">
          <BarChart2 v-if="widget.chart_type === 'bar'" :size="13" class="text-indigo-600 shrink-0" />
          <PieChartIcon v-else-if="widget.chart_type === 'pie'" :size="13" class="text-purple-600 shrink-0" />
          <TrendingUp v-else :size="13" class="text-indigo-600 shrink-0" />
          <span class="truncate">{{ widget.title || 'Visualisasi Data' }}</span>
        </p>
        <p v-if="widget.subtitle" class="text-[10px] text-slate-500 truncate mt-0.5">
          {{ widget.subtitle }}
        </p>
      </div>

      <!-- Action Buttons -->
      <div class="flex items-center gap-1 shrink-0">
        <!-- Toggle View (Grafik / Tabel) -->
        <div class="inline-flex rounded-lg bg-slate-100 p-0.5 text-[10px]">
          <button
            type="button"
            :class="[
              'px-1.5 py-0.5 rounded-md font-medium transition cursor-pointer',
              activeView === 'chart'
                ? 'bg-white text-indigo-700 shadow-2xs'
                : 'text-slate-500 hover:text-slate-800'
            ]"
            title="Tampilan Grafik"
            @click="activeView = 'chart'"
          >
            <BarChart3 :size="11" />
          </button>
          <button
            type="button"
            :class="[
              'px-1.5 py-0.5 rounded-md font-medium transition cursor-pointer',
              activeView === 'table'
                ? 'bg-white text-indigo-700 shadow-2xs'
                : 'text-slate-500 hover:text-slate-800'
            ]"
            title="Tampilan Tabel Data"
            @click="activeView = 'table'"
          >
            <TableIcon :size="11" />
          </button>
        </div>

        <!-- Download Image -->
        <button
          v-if="activeView === 'chart'"
          type="button"
          class="rounded-md p-1 text-slate-400 transition hover:bg-indigo-50 hover:text-indigo-600 cursor-pointer"
          title="Unduh Gambar Grafik (PNG)"
          @click="downloadChartImage(false)"
        >
          <Download :size="12" />
        </button>

        <!-- Maximize / Fullscreen Modal -->
        <button
          type="button"
          class="rounded-md p-1 text-indigo-600 transition hover:bg-indigo-100/70 hover:text-indigo-800 cursor-pointer"
          title="Buka Tampilan Luas & Detail (Inspect Mode)"
          @click="isModalOpen = true"
        >
          <Maximize2 :size="12" />
        </button>
      </div>
    </div>

    <!-- Quick KPI Summary Chips -->
    <div v-if="summaryKpis.length > 0" class="mb-2 flex flex-wrap gap-1">
      <div
        v-for="(kpi, idx) in summaryKpis"
        :key="idx"
        :class="[
          'inline-flex items-center gap-1 rounded-md border px-1.5 py-0.5 text-[10px] font-medium shadow-2xs',
          kpi.colorClass || 'text-slate-700 bg-white border-slate-200'
        ]"
      >
        <span class="text-slate-400 font-normal">{{ kpi.label }}:</span>
        <strong class="font-bold">{{ kpi.value }}</strong>
      </div>
    </div>

    <!-- Content Area: Chart or Table View -->
    <div v-if="activeView === 'chart'" class="h-48 w-full overflow-hidden rounded-lg bg-white p-1 shadow-inner border border-slate-100">
      <VChart
        ref="miniVChartRef"
        :option="miniChartOption"
        autoresize
        class="h-full w-full"
      />
    </div>

    <!-- Tabular Data View -->
    <div v-else class="max-h-52 overflow-y-auto rounded-lg border border-slate-200/80 bg-white text-[11px] shadow-inner">
      <table class="w-full border-collapse text-left">
        <thead class="sticky top-0 bg-slate-50 text-[10px] uppercase text-slate-500 border-b border-slate-200">
          <tr>
            <th class="px-2.5 py-1.5 font-semibold">Kategori / Item</th>
            <th class="px-2.5 py-1.5 font-semibold text-right">Nilai</th>
            <th class="px-2.5 py-1.5 font-semibold text-right">Porsi</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-slate-100">
          <tr
            v-for="row in tableRows"
            :key="row.id"
            class="hover:bg-indigo-50/40 transition"
          >
            <td class="px-2.5 py-1.5 font-medium text-slate-800">{{ row.name }}</td>
            <td class="px-2.5 py-1.5 text-right font-semibold text-indigo-900">{{ row.formattedVal }}</td>
            <td class="px-2.5 py-1.5 text-right text-slate-500 font-mono text-[10px]">{{ row.percentage }}</td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Footer Action Note -->
    <div class="mt-1.5 flex items-center justify-between text-[10px] text-slate-400 px-0.5">
      <span class="flex items-center gap-1">
        <Sparkles :size="10" class="text-amber-500" />
        <span>Klik <Maximize2 :size="9" class="inline text-indigo-600" /> untuk inspeksi penuh</span>
      </span>
    </div>

    <!-- INSPECT / EXPANDED FULLSCREEN MODAL -->
    <Teleport to="body">
      <Transition name="chart-modal-fade">
        <div
          v-if="isModalOpen"
          class="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/60 p-4 backdrop-blur-xs"
          role="dialog"
          aria-modal="true"
        >
          <div
            class="relative flex h-[85vh] max-h-[640px] w-full max-w-4xl flex-col overflow-hidden rounded-2xl border border-slate-200 bg-white shadow-2xl transition-all"
            @click.stop
          >
            <!-- Modal Header -->
            <div class="flex items-center justify-between border-b border-slate-200/90 bg-gradient-to-r from-slate-900 via-indigo-950 to-slate-900 px-5 py-3.5 text-white">
              <div class="flex items-center gap-3">
                <div class="flex h-9 w-9 items-center justify-center rounded-xl bg-indigo-500/20 text-indigo-300 border border-indigo-500/30">
                  <BarChart3 :size="20" />
                </div>
                <div>
                  <h3 class="text-base font-bold tracking-tight text-white flex items-center gap-2">
                    {{ widget.title || 'Detail Analisis Grafik' }}
                  </h3>
                  <p class="text-xs text-slate-300 font-normal">
                    {{ widget.subtitle || 'Tampilan mendalam dengan analisis interaktif & data zoom' }}
                  </p>
                </div>
              </div>

              <!-- Close Button -->
              <div class="flex items-center gap-2">
                <button
                  type="button"
                  class="rounded-lg p-2 text-slate-300 transition hover:bg-white/10 hover:text-white cursor-pointer"
                  title="Tutup (Esc)"
                  @click="isModalOpen = false"
                >
                  <X :size="20" />
                </button>
              </div>
            </div>

            <!-- Modal Sub-Header (Controls: Type Switcher, KPIs, Export) -->
            <div class="flex flex-wrap items-center justify-between gap-3 border-b border-slate-100 bg-slate-50/80 px-5 py-2.5">
              <!-- KPI Badges in Modal -->
              <div class="flex flex-wrap items-center gap-2">
                <div
                  v-for="(kpi, idx) in summaryKpis"
                  :key="idx"
                  class="inline-flex items-center gap-1.5 rounded-lg border bg-white px-2.5 py-1 text-xs font-medium shadow-2xs"
                >
                  <span class="text-slate-500">{{ kpi.label }}:</span>
                  <strong class="text-slate-900 font-bold">{{ kpi.value }}</strong>
                </div>
              </div>

              <!-- Actions & Switchers -->
              <div class="flex items-center gap-2">
                <!-- Switch Chart Type in Modal -->
                <div class="inline-flex rounded-lg bg-slate-200/70 p-0.5 text-xs">
                  <button
                    type="button"
                    :class="[
                      'px-2.5 py-1 rounded-md font-medium transition cursor-pointer flex items-center gap-1',
                      (modalChartType === 'auto' && widget.chart_type !== 'pie') || modalChartType === 'bar'
                        ? 'bg-white text-indigo-700 shadow-2xs font-semibold'
                        : 'text-slate-600 hover:text-slate-900'
                    ]"
                    @click="modalChartType = 'bar'"
                  >
                    <BarChart2 :size="13" />
                    <span>Bar Vertikal</span>
                  </button>
                  <button
                    type="button"
                    :class="[
                      'px-2.5 py-1 rounded-md font-medium transition cursor-pointer flex items-center gap-1',
                      modalChartType === 'bar_horizontal'
                        ? 'bg-white text-indigo-700 shadow-2xs font-semibold'
                        : 'text-slate-600 hover:text-slate-900'
                    ]"
                    @click="modalChartType = 'bar_horizontal'"
                  >
                    <ArrowRightLeft :size="13" />
                    <span>Bar Horizontal</span>
                  </button>
                  <button
                    type="button"
                    :class="[
                      'px-2.5 py-1 rounded-md font-medium transition cursor-pointer flex items-center gap-1',
                      modalChartType === 'line'
                        ? 'bg-white text-indigo-700 shadow-2xs font-semibold'
                        : 'text-slate-600 hover:text-slate-900'
                    ]"
                    @click="modalChartType = 'line'"
                  >
                    <TrendingUp :size="13" />
                    <span>Garis (Line)</span>
                  </button>
                  <button
                    type="button"
                    :class="[
                      'px-2.5 py-1 rounded-md font-medium transition cursor-pointer flex items-center gap-1',
                      (modalChartType === 'auto' && widget.chart_type === 'pie') || modalChartType === 'pie'
                        ? 'bg-white text-purple-700 shadow-2xs font-semibold'
                        : 'text-slate-600 hover:text-slate-900'
                    ]"
                    @click="modalChartType = 'pie'"
                  >
                    <PieChartIcon :size="13" />
                    <span>Donut / Pie</span>
                  </button>
                </div>

                <!-- Download High-Res PNG -->
                <button
                  type="button"
                  class="inline-flex items-center gap-1.5 rounded-lg border border-slate-300 bg-white px-3 py-1.5 text-xs font-semibold text-slate-700 shadow-2xs transition hover:border-indigo-400 hover:bg-indigo-50 hover:text-indigo-700 cursor-pointer"
                  @click="downloadChartImage(true)"
                >
                  <Download :size="14" />
                  <span>Unduh Gambar</span>
                </button>
              </div>
            </div>

            <!-- Modal Chart Canvas -->
            <div class="relative flex-1 bg-white p-4">
              <VChart
                ref="modalVChartRef"
                :option="modalChartOption"
                autoresize
                class="h-full w-full"
              />
            </div>

            <!-- Modal Footer (Table Summary & Tip) -->
            <div class="border-t border-slate-200/90 bg-slate-50 px-5 py-2.5 text-xs text-slate-500 flex items-center justify-between">
              <span>Gunakan mouse wheel atau slider di bawah grafik untuk zoom & navigasi data secara leluasa.</span>
              <button
                type="button"
                class="font-semibold text-indigo-600 hover:text-indigo-800 cursor-pointer"
                @click="isModalOpen = false"
              >
                Tutup Tampilan
              </button>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>

<style scoped>
.chart-modal-fade-enter-active,
.chart-modal-fade-leave-active {
  transition: all 0.2s cubic-bezier(0.16, 1, 0.3, 1);
}
.chart-modal-fade-enter-from,
.chart-modal-fade-leave-to {
  opacity: 0;
  transform: scale(0.97);
}
</style>
