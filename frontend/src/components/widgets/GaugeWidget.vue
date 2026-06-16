<script setup lang="ts">
import { computed } from 'vue'
import type { WidgetElement } from '@/types'

const props = defineProps<{
  title: string
  value: number | null | undefined
  min: number
  max: number
  unit?: string
  elements?: WidgetElement[]
}>()

const percent = computed(() => {
  if (props.value === null || props.value === undefined) return 0
  const range = props.max - props.min || 1
  const pct = ((props.value - props.min) / range) * 100
  return Math.min(100, Math.max(0, pct))
})

function formatNumber(value: number): string {
  return value.toLocaleString(undefined, { maximumFractionDigits: 2 })
}
</script>

<template>
  <div class="card widget-card db-widget-card shadow-sm">
    <div class="card-body" :class="elements?.length ? 'p-0 position-relative' : 'justify-content-center'">
      <template v-if="elements?.length">
        <div class="kpi-elements-canvas">
          <div
            v-for="el in elements"
            :key="el.key"
            class="kpi-element-cell"
            :style="{ left: el.x + '%', top: el.y + '%', width: el.w + '%', height: el.h + '%' }"
          >
            <template v-if="el.key === 'title'">
              <h6 class="widget-title db-label mb-0">{{ title }}</h6>
            </template>
            <template v-else-if="el.key === 'value'">
              <span class="gauge-value db-value">
                {{ value === null || value === undefined ? '—' : formatNumber(value) }}
                <small v-if="unit" class="gauge-unit db-unit ms-1">{{ unit }}</small>
              </span>
            </template>
            <template v-else-if="el.key === 'bar'">
              <div class="progress gauge-track db-track w-100">
                <div
                  class="progress-bar db-bar"
                  role="progressbar"
                  :style="{ width: percent + '%' }"
                  :aria-valuenow="value ?? 0"
                  :aria-valuemin="min"
                  :aria-valuemax="max"
                ></div>
              </div>
            </template>
            <template v-else-if="el.key === 'minmax'">
              <div class="d-flex justify-content-between small w-100 db-minmax gauge-minmax">
                <span>{{ formatNumber(min) }}</span>
                <span>{{ formatNumber(max) }}</span>
              </div>
            </template>
          </div>
        </div>
      </template>
      <template v-else>
        <h6 class="widget-title db-label">{{ title }}</h6>
        <div class="d-flex justify-content-between align-items-baseline mb-2">
          <span class="gauge-value db-value">
            {{ value === null || value === undefined ? '—' : formatNumber(value) }}
            <small v-if="unit" class="gauge-unit db-unit ms-1">{{ unit }}</small>
          </span>
        </div>
        <div class="progress gauge-track db-track">
          <div
            class="progress-bar db-bar"
            role="progressbar"
            :style="{ width: percent + '%' }"
            :aria-valuenow="value ?? 0"
            :aria-valuemin="min"
            :aria-valuemax="max"
          ></div>
        </div>
        <div class="d-flex justify-content-between small mt-1 db-minmax gauge-minmax">
          <span>{{ formatNumber(min) }}</span>
          <span>{{ formatNumber(max) }}</span>
        </div>
      </template>
    </div>
  </div>
</template>

<style scoped>
.db-widget-card {
  background-color: var(--db-card-bg) !important;
  border: 1px solid var(--db-border) !important;
  border-radius: 0.625rem !important;
  box-shadow: 0 0 0 1px var(--db-border-glow), 0 4px 24px rgba(0, 0, 0, 0.5) !important;
  color: var(--db-text);
  overflow: hidden;
}

.db-label {
  color: var(--db-text-muted) !important;
  letter-spacing: 0.04em;
  text-transform: uppercase;
  font-size: inherit;
}

.db-value {
  color: #ffffff !important;
  font-weight: 700;
}

.db-unit {
  color: var(--db-accent2) !important;
}

.db-minmax {
  color: var(--db-text-muted) !important;
}

.db-track {
  background-color: rgba(255, 255, 255, 0.08) !important;
}

.db-bar {
  background-color: var(--db-accent) !important;
  transition: width 0.4s ease;
}
</style>
