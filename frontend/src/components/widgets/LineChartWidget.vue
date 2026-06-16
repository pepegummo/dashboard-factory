<script setup lang="ts">
import { computed } from 'vue'
import { Line } from 'vue-chartjs'
import {
  CategoryScale,
  Chart as ChartJS,
  LinearScale,
  LineElement,
  PointElement,
  Tooltip,
  type ChartData,
  type ChartOptions,
} from 'chart.js'
import type { HistoryPoint } from '@/composables/useTelemetry'

ChartJS.register(CategoryScale, LinearScale, PointElement, LineElement, Tooltip)

const props = defineProps<{
  title: string
  history: HistoryPoint[]
  unit?: string
}>()

const chartData = computed<ChartData<'line'>>(() => ({
  labels: props.history.map((p) => new Date(p.t).toLocaleTimeString()),
  datasets: [
    {
      label: props.title,
      data: props.history.map((p) => p.v),
      borderColor: '#0d6efd',
      backgroundColor: 'rgba(13, 110, 253, 0.15)',
      tension: 0.35,
      fill: true,
      pointRadius: 0,
      borderWidth: 2,
    },
  ],
}))

const chartOptions = computed<ChartOptions<'line'>>(() => ({
  responsive: true,
  maintainAspectRatio: false,
  animation: false,
  scales: {
    x: { display: false },
    y: {
      ticks: {
        callback: (value) => (props.unit ? `${value} ${props.unit}` : `${value}`),
      },
    },
  },
  plugins: {
    legend: { display: false },
  },
}))
</script>

<template>
  <div class="card widget-card shadow-sm">
    <div class="card-body">
      <h6 class="card-subtitle text-muted widget-title">{{ title }}</h6>
      <div class="chart-container">
        <Line v-if="history.length" :data="chartData" :options="chartOptions" />
        <div v-else class="d-flex align-items-center justify-content-center h-100 text-muted small">
          Waiting for data…
        </div>
      </div>
    </div>
  </div>
</template>
