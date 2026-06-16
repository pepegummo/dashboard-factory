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
  <div class="card widget-card shadow-sm">
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
              <h6 class="card-subtitle text-muted widget-title mb-0">{{ title }}</h6>
            </template>
            <template v-else-if="el.key === 'value'">
              <span class="kpi-value">{{ formatValue(value) }}</span>
            </template>
            <template v-else-if="el.key === 'unit'">
              <span class="kpi-unit text-muted">{{ unit ?? '—' }}</span>
            </template>
          </div>
        </div>
      </template>

      <!-- Default CSS layout (unchanged) -->
      <template v-else>
        <h6 class="card-subtitle text-muted widget-title">{{ props.title }}</h6>
        <div class="kpi-value">
          {{ formatValue(value) }}
          <small v-if="unit" class="kpi-unit text-muted ms-1">{{ unit }}</small>
        </div>
      </template>
    </div>
  </div>
</template>
