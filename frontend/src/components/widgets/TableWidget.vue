<script setup lang="ts">
import { useCatalogStore } from '@/stores/catalog.store'

defineProps<{
  title: string
  readings: Record<string, number | string> | null
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
  <div class="card widget-card shadow-sm">
    <div class="card-body">
      <h6 class="card-subtitle text-muted widget-title">{{ title }}</h6>
      <div v-if="!readings" class="text-muted small">Waiting for data…</div>
      <table v-else class="table table-sm mb-0 widget-table">
        <tbody>
          <tr v-for="metric in catalog.metrics" :key="metric.key">
            <td class="text-muted">{{ metric.label }}</td>
            <td class="text-end fw-semibold">{{ formatValue(metric.key, readings[metric.key]) }}</td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>
