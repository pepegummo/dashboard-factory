<script setup lang="ts">
import type { WidgetElement } from '@/types'

const props = defineProps<{
  title: string
  value: number | string | null | undefined
  unit?: string
  elements?: WidgetElement[]
}>()

function formatValue(value: number | string | null | undefined): string {
  if (value === null || value === undefined) return '—'
  if (typeof value === 'number') {
    return value.toLocaleString(undefined, { maximumFractionDigits: 2 })
  }
  return String(value)
}
</script>

<template>
  <div class="card widget-card db-widget-card shadow-sm">
    <div class="card-body" :class="elements?.length ? 'p-0 position-relative' : 'text-center justify-content-center'">
      <!-- Elements-based layout: each sub-element absolutely positioned -->
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
              <span class="kpi-value db-value">{{ formatValue(value) }}</span>
            </template>
            <template v-else-if="el.key === 'unit'">
              <span class="kpi-unit db-unit">{{ unit ?? '—' }}</span>
            </template>
          </div>
        </div>
      </template>

      <!-- Default CSS layout -->
      <template v-else>
        <h6 class="widget-title db-label">{{ props.title }}</h6>
        <div class="kpi-value db-value">
          {{ formatValue(value) }}
          <small v-if="unit" class="kpi-unit db-unit ms-1">{{ unit }}</small>
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
  color: var(--db-text) !important;
  font-weight: 700;
  line-height: 1.05;
  letter-spacing: -0.01em;
}

.db-unit {
  color: var(--db-accent2) !important;
}
</style>
