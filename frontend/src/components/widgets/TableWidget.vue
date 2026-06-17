<script setup lang="ts">
import type { WidgetElement } from '@/types'
import { useCatalogStore } from '@/stores/catalog.store'

defineProps<{
  title: string
  readings: Record<string, number | string> | null
  elements?: WidgetElement[]
}>()

const catalog = useCatalogStore()

function formatValue(key: string, value: number | string | undefined): string {
  if (value === undefined) return '—'
  if (typeof value === 'number') {
    const formatted = value.toLocaleString(undefined, { maximumFractionDigits: 2 })
    const unit = catalog.metricUnit(key)
    return unit ? `${formatted} ${unit}` : formatted
  }
  return String(value)
}
</script>

<template>
  <div class="card widget-card db-widget-card shadow-sm">
    <div class="card-body" :class="elements?.length ? 'p-0 position-relative' : ''">
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
            <template v-else-if="el.key === 'table'">
              <div class="position-absolute overflow-auto" style="inset: 0;">
                <div v-if="!readings" class="db-waiting small p-2">Waiting for data…</div>
                <table v-else class="table table-sm mb-0 widget-table db-table">
                  <tbody>
                    <tr v-for="metric in catalog.metrics" :key="metric.key">
                      <td class="db-td-label">{{ metric.label }}</td>
                      <td class="text-end fw-semibold db-td-value">{{ formatValue(metric.key, readings[metric.key]) }}</td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </template>
          </div>
        </div>
      </template>
      <template v-else>
        <h6 class="widget-title db-label">{{ title }}</h6>
        <div v-if="!readings" class="db-waiting small">Waiting for data…</div>
        <table v-else class="table table-sm mb-0 widget-table db-table">
          <tbody>
            <tr v-for="metric in catalog.metrics" :key="metric.key">
              <td class="db-td-label">{{ metric.label }}</td>
              <td class="text-end fw-semibold db-td-value">{{ formatValue(metric.key, readings[metric.key]) }}</td>
            </tr>
          </tbody>
        </table>
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

.db-waiting {
  color: var(--db-text-muted);
}

/* Dark table — override Bootstrap defaults */
.db-table {
  --bs-table-bg: transparent;
  --bs-table-striped-bg: rgba(255, 255, 255, 0.035);
  --bs-table-border-color: rgba(255, 255, 255, 0.06);
  color: var(--db-text);
  border-color: rgba(255, 255, 255, 0.06);
}

.db-table tbody tr:nth-child(even) {
  background-color: rgba(255, 255, 255, 0.03);
}

.db-table tbody tr:hover {
  background-color: rgba(0, 210, 255, 0.06);
}

.db-td-label {
  color: var(--db-text-muted);
}

.db-td-value {
  color: var(--db-text);
}
</style>
