<script setup lang="ts">
import { computed } from 'vue'
import type { Widget } from '@/types'
import type { HistoryPoint } from '@/composables/useTelemetry'
import { useCatalogStore } from '@/stores/catalog.store'
import BarChartWidget from './BarChartWidget.vue'
import GaugeWidget from './GaugeWidget.vue'
import KpiWidget from './KpiWidget.vue'
import LineChartWidget from './LineChartWidget.vue'
import StatusWidget from './StatusWidget.vue'
import TableWidget from './TableWidget.vue'

const props = defineProps<{
  widget: Widget
  readings: Record<string, number | string> | null
  history: Record<string, HistoryPoint[]>
}>()

const catalog = useCatalogStore()

const metric = computed(() => catalog.metrics.find((m) => m.key === props.widget.metricKey))
const value = computed(() => props.readings?.[props.widget.metricKey] ?? null)
const numericValue = computed(() => (typeof value.value === 'number' ? value.value : null))
const widgetHistory = computed(() => props.history[props.widget.metricKey] ?? [])
</script>

<template>
  <StatusWidget v-if="widget.type === 'status'" :title="widget.title" :value="value" />
  <KpiWidget
    v-else-if="widget.type === 'kpi'"
    :title="widget.title"
    :value="value"
    :unit="metric?.unit"
    :elements="widget.elements"
  />
  <GaugeWidget
    v-else-if="widget.type === 'gauge'"
    :title="widget.title"
    :value="numericValue"
    :min="metric?.min ?? 0"
    :max="metric?.max ?? 100"
    :unit="metric?.unit"
  />
  <LineChartWidget
    v-else-if="widget.type === 'line'"
    :title="widget.title"
    :history="widgetHistory"
    :unit="metric?.unit"
  />
  <BarChartWidget
    v-else-if="widget.type === 'bar'"
    :title="widget.title"
    :history="widgetHistory"
    :unit="metric?.unit"
  />
  <TableWidget v-else :title="widget.title" :readings="readings" />
</template>
