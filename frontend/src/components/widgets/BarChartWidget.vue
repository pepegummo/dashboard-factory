<script setup lang="ts">
import { computed } from 'vue'
import { Bar } from 'vue-chartjs'
import {
  BarElement,
  CategoryScale,
  Chart as ChartJS,
  LinearScale,
  Tooltip,
  type ChartData,
  type ChartOptions,
} from 'chart.js'
import type { HistoryPoint } from '@/composables/useTelemetry'

ChartJS.register(CategoryScale, LinearScale, BarElement, Tooltip)

const props = defineProps<{
  title: string
  history: HistoryPoint[]
  unit?: string
}>()

const chartData = computed<ChartData<'bar'>>(() => ({
  labels: props.history.map((p) => new Date(p.t).toLocaleTimeString()),
  datasets: [
    {
      label: props.title,
      data: props.history.map((p) => p.v),
      backgroundColor: 'rgba(13, 110, 253, 0.6)',
      borderColor: '#0d6efd',
      borderWidth: 1,
      borderRadius: 4,
    },
  ],
}))

const chartOptions = computed<ChartOptions<'bar'>>(() => ({
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
        <Bar v-if="history.length" :data="chartData" :options="chartOptions" />
        <div v-else class="d-flex align-items-center justify-content-center h-100 text-muted small">
          Waiting for data…
        </div>
      </div>
    </div>
  </div>
</template>
