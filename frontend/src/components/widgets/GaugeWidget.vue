<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  title: string
  value: number | null | undefined
  min: number
  max: number
  unit?: string
}>()

const percent = computed(() => {
  if (props.value === null || props.value === undefined) return 0
  const range = props.max - props.min || 1
  const pct = ((props.value - props.min) / range) * 100
  return Math.min(100, Math.max(0, pct))
})

const colorClass = computed(() => {
  if (percent.value >= 90) return 'bg-danger'
  if (percent.value >= 70) return 'bg-warning'
  return 'bg-success'
})

function formatNumber(value: number): string {
  return value.toLocaleString(undefined, { maximumFractionDigits: 2 })
}
</script>

<template>
  <div class="card widget-card shadow-sm">
    <div class="card-body justify-content-center">
      <h6 class="card-subtitle text-muted widget-title">{{ title }}</h6>
      <div class="d-flex justify-content-between align-items-baseline mb-2">
        <span class="gauge-value">
          {{ value === null || value === undefined ? '—' : formatNumber(value) }}
          <small v-if="unit" class="gauge-unit text-muted ms-1">{{ unit }}</small>
        </span>
      </div>
      <div class="progress gauge-track">
        <div
          class="progress-bar"
          :class="colorClass"
          role="progressbar"
          :style="{ width: percent + '%' }"
          :aria-valuenow="value ?? 0"
          :aria-valuemin="min"
          :aria-valuemax="max"
        ></div>
      </div>
      <div class="d-flex justify-content-between text-muted small mt-1 gauge-minmax">
        <span>{{ formatNumber(min) }}</span>
        <span>{{ formatNumber(max) }}</span>
      </div>
    </div>
  </div>
</template>
