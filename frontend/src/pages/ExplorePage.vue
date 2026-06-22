<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useDashboardStore } from '@/stores/dashboard.store'
import { useCatalogStore } from '@/stores/catalog.store'
import { useTelemetry } from '@/composables/useTelemetry'
import { api, apiErrorMessage } from '@/services/api'
import { DEFAULT_ELEMENTS } from '@/utils/widgetElements'
import WidgetRenderer from '@/components/widgets/WidgetRenderer.vue'
import ExplorePanel from '@/components/ExplorePanel.vue'
import type { Widget } from '@/types'

const dashboardStore = useDashboardStore()
const catalog = useCatalogStore()
const { reading, history, start } = useTelemetry()

dashboardStore.fetchDashboards()
catalog.fetchMetrics()

const selectedId = ref('')
const dashboard = ref<Awaited<ReturnType<typeof api.getDashboard>> | null>(null)
const loading = ref(false)
const error = ref('')
const activePage = ref(0)
const highlightedWidgets = ref<number[]>([])
const prefillQuestion = ref('')

const pages = computed(() => dashboard.value?.pages ?? [])
const currentMachine = computed(() => pages.value[activePage.value]?.machine ?? null)
const widgets = computed(() => dashboard.value?.template?.widgets ?? [])
const template = computed(() => dashboard.value?.template ?? null)

const canvasStyle = computed(() => {
  if (!template.value) return undefined
  const ratio = template.value.width / template.value.height
  return {
    aspectRatio: `${template.value.width} / ${template.value.height}`,
    width: `min(100%, calc(80vh * ${ratio}))`,
    maxHeight: '80vh',
  }
})

function elementsFor(w: Widget) {
  return w.elements ?? DEFAULT_ELEMENTS[w.type] ?? []
}

async function loadDashboard() {
  if (!selectedId.value) { dashboard.value = null; return }
  loading.value = true
  error.value = ''
  activePage.value = 0
  highlightedWidgets.value = []
  hovered.value = null
  try {
    dashboard.value = await api.getDashboard(selectedId.value)
  } catch (err) {
    error.value = apiErrorMessage(err, 'Failed to load dashboard')
  } finally {
    loading.value = false
  }
}

watch(currentMachine, (machine) => { if (machine) start(machine.id) }, { immediate: true })

// --- Element inspect overlay ---
const hovered = ref<{ widgetId: string; key: string } | null>(null)
const focused = ref<{ widgetId: string; key: string } | null>(null)

function highlightWidget(i: number) {
  const w = widgets.value[i]
  highlightedWidgets.value = [i]
  focused.value = null
  prefillQuestion.value = `Explain the "${w.title}" ${w.type} widget overall.`
}

function onElementClick(w: Widget, key: string, i: number) {
  highlightedWidgets.value = [i]
  focused.value = { widgetId: w.id, key }
  prefillQuestion.value = `What is the "${key}" element on the "${w.title}" widget?`
}

// --- Context for AI ---
function buildContext(): string {
  if (!dashboard.value) return ''
  const lines: string[] = []
  lines.push(`Dashboard: ${dashboard.value.name}`)
  if (dashboard.value.factory) lines.push(`Factory: ${dashboard.value.factory.name}`)
  if (currentMachine.value) lines.push(`Current machine: ${currentMachine.value.name}`)
  lines.push('Widgets:')
  widgets.value.forEach((w, i) => {
    const m = catalog.metrics.find((x) => x.key === w.metricKey)
    const r = reading.value?.[w.metricKey]
    const unit = m?.unit ? ' ' + m.unit : ''
    const range = m ? `, range ${m.min}–${m.max}${unit}` : ''
    let dataStr = ''
    if (w.type === 'line') {
      const pts = history.value?.[w.metricKey]
      dataStr = pts?.length ? `, recent (${pts.length} pts): ${pts.map(p => p.v).join(', ')}${unit}` : ''
    } else {
      dataStr = r !== undefined ? `, live: ${r}${unit}` : ''
    }
    const elKeys = elementsFor(w).map((e) => e.key).join(', ')
    lines.push(`[${i}] ${w.type} "${w.title}"${range}${dataStr} (elements: ${elKeys})`)
  })
  return lines.join('\n')
}

const exploreContext = computed(() => buildContext())

function onAiFocus(el: { widgetIndex: number; key: string } | null) {
  if (!el) { focused.value = null; return }
  const w = widgets.value[el.widgetIndex]
  if (w) focused.value = { widgetId: w.id, key: el.key }
}
</script>

<template>
  <div class="explore-page">
    <!-- Dashboard selector -->
    <div class="d-flex align-items-center gap-2 mb-3">
      <label class="fw-semibold mb-0">Dashboard</label>
      <select
        v-model="selectedId"
        class="form-select w-auto"
        :disabled="dashboardStore.loading"
        @change="loadDashboard"
      >
        <option value="">Select a dashboard…</option>
        <option v-for="d in dashboardStore.dashboards" :key="d.id" :value="d.id">
          {{ d.name }}<template v-if="d.factory"> — {{ d.factory.name }}</template>
        </option>
      </select>
    </div>

    <div v-if="loading" class="text-center py-5">
      <div class="spinner-border text-primary" role="status"></div>
    </div>
    <div v-else-if="error" class="alert alert-danger">{{ error }}</div>

    <div v-else-if="dashboard" class="explore-split">
      <!-- Left: canvas -->
      <div class="explore-canvas-col">
        <div class="dashboard-canvas" :style="canvasStyle" @click.self="highlightedWidgets = []; focused = null">
          <div v-if="highlightedWidgets.length" class="canvas-spotlight-overlay"></div>
          <div
            v-for="(w, i) in widgets"
            :key="w.id"
            class="dashboard-canvas-item"
            :class="{ highlighted: highlightedWidgets.includes(i) }"
            :style="{ left: `${w.x}%`, top: `${w.y}%`, width: `${w.w}%`, height: `${w.h}%` }"
            @click="highlightWidget(i)"
          >
            <WidgetRenderer :widget="w" :readings="reading" :history="history" :machine="currentMachine" />
            <div class="inspect-overlay">
              <div
                v-for="el in elementsFor(w)"
                :key="el.key"
                class="element-handle"
                :class="{
                  'is-active':  focused?.widgetId === w.id && focused?.key === el.key,
                  'is-hovered': hovered?.widgetId === w.id && hovered?.key === el.key,
                }"
                :style="{ left: `${el.x}%`, top: `${el.y}%`, width: `${el.w}%`, height: `${el.h}%` }"
                @mouseenter="hovered = { widgetId: w.id, key: el.key }"
                @mouseleave="hovered = null"
                @click.stop="onElementClick(w, el.key, i)"
              >
                <span class="element-label">{{ el.key }}</span>
              </div>
            </div>
          </div>
        </div>

        <div v-if="pages.length > 1" class="d-flex justify-content-center gap-2 mt-2">
          <button
            v-for="(p, i) in pages"
            :key="p.id"
            type="button"
            class="page-dot"
            :class="{ active: activePage === i }"
            :aria-label="`Show ${p.machine?.name}`"
            @click="activePage = i"
          ></button>
        </div>
      </div>

      <!-- Right: single Q&A panel -->
      <ExplorePanel
        :context="exploreContext"
        :reset-key="selectedId + ':' + activePage"
        :prefill="prefillQuestion"
        @highlight="highlightedWidgets = $event"
        @focus="onAiFocus"
      />
    </div>

    <p v-else class="text-secondary py-4">
      Pick a dashboard to view it and ask about its widgets.
    </p>
  </div>
</template>

<style scoped>
.explore-split {
  display: flex;
  gap: 1rem;
  align-items: flex-start;
}

.explore-canvas-col {
  flex: 1;
  min-width: 0;
}

.inspect-overlay {
  position: absolute;
  inset: 1px;
  z-index: 20;
}
</style>
