<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, reactive, ref, watch } from 'vue'
import type { ComponentPublicInstance } from 'vue'
import { useRouter } from 'vue-router'
import { GridStack, type GridStackNode } from 'gridstack'
import 'gridstack/dist/gridstack.min.css'
import { useTemplateStore } from '@/stores/template.store'
import { useCatalogStore } from '@/stores/catalog.store'
import { apiErrorMessage } from '@/services/api'
import WidgetRenderer from '@/components/widgets/WidgetRenderer.vue'
import WidgetEditorShell from '@/components/widgets/WidgetEditorShell.vue'
import type { HistoryPoint } from '@/composables/useTelemetry'
import type { Widget, WidgetElement, WidgetType } from '@/types'

const props = defineProps<{ id?: string }>()

const router = useRouter()
const templateStore = useTemplateStore()
const catalog = useCatalogStore()

const isEdit = computed(() => !!props.id)

const form = reactive({
  name: '',
  description: '',
  width: 1,
  height: 1,
  widgets: [] as Widget[],
})

// Fixed grid-line overlay density for the editor canvas: widgets are
// positioned/sized as percentages, so a 10x10 overlay marks 10% increments
// regardless of the canvas's pixel dimensions.
const gridOverlayLines = 10

const loading = ref(false)
const saving = ref(false)
const error = ref('')

const selectedId = ref<string | null>(null)
const elementEditMode = ref(false)
const selectedElementKey = ref<string | null>(null)
const gridEl = ref<HTMLElement | null>(null)
let grid: GridStack | null = null
let resizeObserver: ResizeObserver | null = null
const itemRefs = new Map<string, HTMLElement>()
let widgetCounter = 0

const widgetTypeOptions: { value: WidgetType; label: string; icon: string }[] = [
  { value: 'status', label: 'Status', icon: 'bi-circle-fill' },
  { value: 'kpi', label: 'KPI', icon: 'bi-123' },
  { value: 'gauge', label: 'Gauge', icon: 'bi-speedometer2' },
  { value: 'line', label: 'Line chart', icon: 'bi-graph-up' },
  { value: 'bar', label: 'Bar chart', icon: 'bi-bar-chart-fill' },
  { value: 'table', label: 'Table', icon: 'bi-table' },
]

const defaultSize: Record<WidgetType, { w: number; h: number }> = {
  status: { w: 25, h: 22 },
  kpi: { w: 25, h: 22 },
  gauge: { w: 33, h: 22 },
  line: { w: 50, h: 34 },
  bar: { w: 50, h: 34 },
  table: { w: 100, h: 34 },
}

const selectedWidget = computed(() => form.widgets.find((w) => w.id === selectedId.value) ?? null)

// Finds the first free top-left position (row-major, 1% steps) for a
// w x h block that doesn't overlap any existing widget and stays within
// the 100x100 canvas. Returns null if no such position exists.
function findFreeSlot(w: number, h: number): { x: number; y: number } | null {
  if (w > 100 || h > 100) return null
  for (let y = 0; y <= 100 - h; y++) {
    for (let x = 0; x <= 100 - w; x++) {
      const overlaps = form.widgets.some(
        (other) => x < other.x + other.w && other.x < x + w && y < other.y + other.h && other.y < y + h,
      )
      if (!overlaps) return { x, y }
    }
  }
  return null
}

// Whether each widget type's default size still fits somewhere on the canvas.
const canAddWidget = computed<Record<WidgetType, boolean>>(() => {
  const result = {} as Record<WidgetType, boolean>
  for (const opt of widgetTypeOptions) {
    const size = defaultSize[opt.value]
    result[opt.value] = findFreeSlot(size.w, size.h) !== null
  }
  return result
})

// Sample readings so the board preview looks like a live dashboard widget.
const sampleReadings = computed<Record<string, number | string>>(() => {
  const readings: Record<string, number | string> = {}
  for (const m of catalog.metrics) {
    readings[m.key] = m.key === 'status' ? 'running' : Math.round((m.min + m.max) / 2)
  }
  return readings
})

// Sample history so line/bar chart widgets in the board preview show a
// representative chart instead of "Waiting for data…".
const sampleHistory = computed<Record<string, HistoryPoint[]>>(() => {
  const history: Record<string, HistoryPoint[]> = {}
  const now = Date.now()
  for (const m of catalog.metrics) {
    if (m.key === 'status') continue
    const mid = (m.min + m.max) / 2
    const amplitude = (m.max - m.min) / 4
    history[m.key] = Array.from({ length: 10 }, (_, i) => ({
      t: now - (9 - i) * 2000,
      v: mid + amplitude * Math.sin(i / 1.5),
    }))
  }
  return history
})

const widthInput = computed({
  get: () => selectedWidget.value?.w ?? 1,
  set: (value: number) => updateSize('w', value),
})
const heightInput = computed({
  get: () => selectedWidget.value?.h ?? 1,
  set: (value: number) => updateSize('h', value),
})

const aspectRatio = computed(() => {
  const w = form.width > 0 ? form.width : 1
  const h = form.height > 0 ? form.height : 1
  return h / w
})

const canvasRatio = computed(() => 1 / aspectRatio.value)

function clampRatio(v: number): number {
  return Math.min(999, Math.max(1, Math.round(v) || 1))
}
const canvasWidthInput = computed({
  get: () => form.width,
  set: (v: number) => { form.width = clampRatio(v) },
})
const canvasHeightInput = computed({
  get: () => form.height,
  set: (v: number) => { form.height = clampRatio(v) },
})

function recomputeCellHeight() {
  if (!grid) return
  const cellWidthPx = grid.cellWidth()
  if (!cellWidthPx) return
  grid.cellHeight(cellWidthPx * aspectRatio.value)
}

watch([() => form.width, () => form.height], () => {
  nextTick(() => recomputeCellHeight())
})

onMounted(async () => {
  await catalog.fetchMetrics()
  if (props.id) {
    loading.value = true
    try {
      const t = await templateStore.getTemplate(props.id)
      form.name = t.name
      form.description = t.description
      form.width = t.width
      form.height = t.height
      form.widgets = t.widgets.map((w) => ({ ...w }))
    } catch (err) {
      error.value = apiErrorMessage(err, 'Failed to load template')
    } finally {
      loading.value = false
    }
  }
  await nextTick()
  initGrid()
})

onUnmounted(() => {
  resizeObserver?.disconnect()
  resizeObserver = null
  grid?.destroy(false)
  grid = null
})

function initGrid() {
  if (!gridEl.value || grid) return
  grid = GridStack.init(
    {
      column: 100,
      maxRow: 100,
      cellHeight: 1,
      margin: 2,
      float: true,
      handle: '.drag-handle',
    },
    gridEl.value,
  )
  grid.on('change', (_event: Event, items: GridStackNode[]) => applyLayoutChange(items))
  grid.on('added', (_event: Event, items: GridStackNode[]) => applyLayoutChange(items))

  recomputeCellHeight()
  resizeObserver = new ResizeObserver(() => recomputeCellHeight())
  resizeObserver.observe(gridEl.value)
}

function applyLayoutChange(items: GridStackNode[]) {
  for (const item of items) {
    if (!item.id) continue
    const widget = form.widgets.find((w) => w.id === item.id)
    if (!widget) continue
    if (item.x != null) widget.x = item.x
    if (item.y != null) widget.y = item.y
    if (item.w != null) widget.w = item.w
    if (item.h != null) widget.h = item.h
  }
}

function setItemRef(id: string, el: Element | ComponentPublicInstance | null) {
  if (el instanceof HTMLElement) {
    itemRefs.set(id, el)
  } else {
    itemRefs.delete(id)
  }
}

function selectWidget(id: string) {
  selectedId.value = id
  elementEditMode.value = false
  selectedElementKey.value = null
}

function enterElementEdit(id: string) {
  selectedId.value = id
  elementEditMode.value = true
}

function updateWidgetElements(id: string, els: WidgetElement[]) {
  const w = form.widgets.find((w) => w.id === id)
  if (w) w.elements = els
}

const KPI_DEFAULT_ELEMENTS: WidgetElement[] = [
  { key: 'title', x: 0, y: 0,  w: 100, h: 25 },
  { key: 'value', x: 0, y: 25, w: 100, h: 50 },
  { key: 'unit',  x: 0, y: 75, w: 100, h: 25 },
]

const selectedWidgetElements = computed<WidgetElement[]>(() => {
  const w = selectedWidget.value
  if (!w || w.type !== 'kpi') return []
  return (w.elements?.length ? w.elements : KPI_DEFAULT_ELEMENTS).map((e) => ({ ...e }))
})

function updateElementProp(key: string, prop: keyof Omit<WidgetElement, 'key'>, rawVal: number) {
  const w = selectedWidget.value
  if (!w) return
  const val = Math.round(Math.min(100, Math.max(prop === 'w' || prop === 'h' ? 1 : 0, rawVal)))
  const base = w.elements?.length ? w.elements : KPI_DEFAULT_ELEMENTS
  updateWidgetElements(w.id, base.map((e) => e.key === key ? { ...e, [prop]: val } : { ...e }))
}

function newWidgetId(): string {
  widgetCounter += 1
  return `w-new-${Date.now()}-${widgetCounter}`
}

function addWidget(type: WidgetType) {
  const size = defaultSize[type]
  const slot = findFreeSlot(size.w, size.h)
  if (!slot) return
  const widget: Widget = {
    id: newWidgetId(),
    type,
    title: widgetTypeOptions.find((o) => o.value === type)?.label ?? 'Widget',
    metricKey: type === 'table' ? '' : (catalog.metrics[0]?.key ?? ''),
    x: slot.x,
    y: slot.y,
    w: size.w,
    h: size.h,
  }
  form.widgets.push(widget)
  selectedId.value = widget.id
  nextTick(() => {
    const el = itemRefs.get(widget.id)
    if (!el || !grid) return
    const item = grid.makeWidget(el)
    if (item.gridstackNode) applyLayoutChange([item.gridstackNode])
  })
}

function removeWidget(id: string) {
  const el = itemRefs.get(id)
  if (grid && el) grid.removeWidget(el, false)
  itemRefs.delete(id)
  form.widgets = form.widgets.filter((w) => w.id !== id)
  if (selectedId.value === id) selectedId.value = null
}

function onTypeChange(widget: Widget) {
  if (widget.type === 'table') {
    widget.metricKey = ''
  } else if (!widget.metricKey) {
    widget.metricKey = catalog.metrics[0]?.key ?? ''
  }
}

function updateSize(dim: 'w' | 'h', value: number) {
  const widget = selectedWidget.value
  if (!widget) return
  const clamped = Math.min(100, Math.max(1, Math.round(value) || 1))
  const el = itemRefs.get(widget.id)
  if (dim === 'w') {
    widget.w = clamped
    if (el && grid) grid.update(el, { w: clamped })
  } else {
    widget.h = clamped
    if (el && grid) grid.update(el, { h: clamped })
  }
}

function validate(): string | null {
  if (!form.name.trim()) return 'Template name is required.'
  if (form.width < 1 || form.width > 10000 || form.height < 1 || form.height > 10000) {
    return 'Canvas size must be between 1 and 10000 pixels.'
  }
  if (!form.widgets.length) return 'Add at least one widget.'
  for (const w of form.widgets) {
    if (!w.title.trim()) return 'Every widget needs a title.'
    if (w.type !== 'table' && !w.metricKey) return `Widget "${w.title}" needs a metric.`
    if (w.x + w.w > 100 || w.y + w.h > 100) return `Widget "${w.title}" extends outside the canvas.`
  }
  return null
}

async function save() {
  const validationError = validate()
  if (validationError) {
    error.value = validationError
    return
  }
  error.value = ''
  saving.value = true
  try {
    const payload = {
      name: form.name.trim(),
      description: form.description.trim(),
      width: form.width,
      height: form.height,
      widgets: form.widgets,
    }
    if (isEdit.value && props.id) {
      await templateStore.updateTemplate(props.id, payload)
    } else {
      await templateStore.createTemplate(payload)
    }
    router.push('/templates')
  } catch (err) {
    error.value = apiErrorMessage(err, 'Failed to save template')
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <div>
    <div class="d-flex justify-content-between align-items-center mb-4">
      <div>
        <h1 class="h3 mb-1">{{ isEdit ? 'Edit Template' : 'New Template' }}</h1>
        <p class="text-muted mb-0">
          Design the widget board that will be applied to each machine page.
        </p>
      </div>
      <RouterLink class="btn btn-outline-secondary" to="/templates">
        <i class="bi bi-arrow-left me-1"></i>Back
      </RouterLink>
    </div>

    <div v-if="loading" class="text-center py-5">
      <div class="spinner-border text-primary" role="status"></div>
    </div>

    <template v-else>
      <div v-if="error" class="alert alert-danger">{{ error }}</div>

      <div class="row g-3">
        <div class="col-12 col-lg-3">
          <div class="card shadow-sm mb-3">
            <div class="card-body">
              <div class="mb-3">
                <label class="form-label">Template name</label>
                <input
                  v-model="form.name"
                  type="text"
                  class="form-control"
                  placeholder="e.g. Standard Machine Monitor"
                />
              </div>
              <div class="mb-3">
                <label class="form-label">Description</label>
                <textarea
                  v-model="form.description"
                  class="form-control"
                  rows="2"
                  placeholder="Optional description"
                ></textarea>
              </div>
              <div>
                <label class="form-label mb-1">Canvas ratio</label>
                <div class="d-flex align-items-center gap-2">
                  <input
                    v-model.number="canvasWidthInput"
                    type="number"
                    min="1"
                    max="999"
                    class="form-control form-control-sm"
                    style="width: 4rem"
                  />
                  <span class="text-muted">:</span>
                  <input
                    v-model.number="canvasHeightInput"
                    type="number"
                    min="1"
                    max="999"
                    class="form-control form-control-sm"
                    style="width: 4rem"
                  />
                </div>
                <p class="form-text small text-muted mb-0">
                  Width : height ratio of the output canvas, e.g. 16 : 9.
                  The editor board always displays as a square.
                </p>
              </div>
            </div>
          </div>

          <div class="card shadow-sm mb-3">
            <div class="card-header bg-white">
              <strong>Add widget</strong>
            </div>
            <div class="card-body d-flex flex-wrap gap-2">
              <button
                v-for="opt in widgetTypeOptions"
                :key="opt.value"
                type="button"
                class="btn btn-sm btn-outline-primary"
                :disabled="!canAddWidget[opt.value]"
                :title="canAddWidget[opt.value] ? '' : 'Not enough space on the canvas for this widget'"
                @click="addWidget(opt.value)"
              >
                <i :class="['bi', opt.icon, 'me-1']"></i>{{ opt.label }}
              </button>
            </div>
          </div>

          <div class="card shadow-sm">
            <div class="card-header bg-white">
              <strong>Widget properties</strong>
            </div>
            <div class="card-body">
              <div v-if="selectedWidget && elementEditMode" class="alert alert-info py-1 px-2 small mb-2">
                <i class="bi bi-arrows-move me-1"></i>Drag elements to reposition. Click ⚙ to exit.
              </div>
              <template v-if="selectedWidget">
                <div class="mb-2">
                  <label class="form-label small text-muted mb-1">Title</label>
                  <input v-model="selectedWidget.title" type="text" class="form-control form-control-sm" />
                </div>
                <div class="mb-2">
                  <label class="form-label small text-muted mb-1">Widget type</label>
                  <select
                    v-model="selectedWidget.type"
                    class="form-select form-select-sm"
                    @change="onTypeChange(selectedWidget)"
                  >
                    <option v-for="opt in widgetTypeOptions" :key="opt.value" :value="opt.value">
                      {{ opt.label }}
                    </option>
                  </select>
                </div>
                <div class="mb-2">
                  <label class="form-label small text-muted mb-1">Metric</label>
                  <select
                    v-model="selectedWidget.metricKey"
                    class="form-select form-select-sm"
                    :disabled="selectedWidget.type === 'table'"
                  >
                    <option v-for="m in catalog.metrics" :key="m.key" :value="m.key">
                      {{ m.label }}{{ m.unit ? ` (${m.unit})` : '' }}
                    </option>
                  </select>
                </div>
                <div class="row g-2 mb-3">
                  <div class="col-6">
                    <label class="form-label small text-muted mb-1">Width (%)</label>
                    <input
                      v-model.number="widthInput"
                      type="number"
                      min="1"
                      max="100"
                      class="form-control form-control-sm"
                    />
                  </div>
                  <div class="col-6">
                    <label class="form-label small text-muted mb-1">Height (%)</label>
                    <input
                      v-model.number="heightInput"
                      type="number"
                      min="1"
                      max="100"
                      class="form-control form-control-sm"
                    />
                  </div>
                </div>
                <!-- KPI element positions list -->
                <template v-if="selectedWidget.type === 'kpi'">
                  <hr class="my-2" />
                  <label class="form-label small text-muted mb-1">
                    <i class="bi bi-layout-three-columns me-1"></i>Elements
                  </label>
                  <div
                    v-for="el in selectedWidgetElements"
                    :key="el.key"
                    class="mb-1 border rounded"
                  >
                    <button
                      type="button"
                      class="btn btn-sm w-100 text-start px-2 py-1 d-flex align-items-center justify-content-between"
                      :class="selectedElementKey === el.key ? 'btn-primary' : 'btn-light'"
                      @click="selectedElementKey = selectedElementKey === el.key ? null : el.key"
                    >
                      <span class="text-capitalize small fw-semibold">{{ el.key }}</span>
                      <i :class="['bi', selectedElementKey === el.key ? 'bi-chevron-up' : 'bi-chevron-down']"></i>
                    </button>
                    <div v-if="selectedElementKey === el.key" class="p-2">
                      <div class="row g-1">
                        <div class="col-6">
                          <label class="form-text mb-0">X %</label>
                          <input type="number" min="0" max="99" class="form-control form-control-sm"
                            :value="el.x"
                            @change="updateElementProp(el.key, 'x', Number(($event.target as HTMLInputElement).value))"
                          />
                        </div>
                        <div class="col-6">
                          <label class="form-text mb-0">Y %</label>
                          <input type="number" min="0" max="99" class="form-control form-control-sm"
                            :value="el.y"
                            @change="updateElementProp(el.key, 'y', Number(($event.target as HTMLInputElement).value))"
                          />
                        </div>
                        <div class="col-6">
                          <label class="form-text mb-0">W %</label>
                          <input type="number" min="1" max="100" class="form-control form-control-sm"
                            :value="el.w"
                            @change="updateElementProp(el.key, 'w', Number(($event.target as HTMLInputElement).value))"
                          />
                        </div>
                        <div class="col-6">
                          <label class="form-text mb-0">H %</label>
                          <input type="number" min="1" max="100" class="form-control form-control-sm"
                            :value="el.h"
                            @change="updateElementProp(el.key, 'h', Number(($event.target as HTMLInputElement).value))"
                          />
                        </div>
                      </div>
                    </div>
                  </div>
                </template>

                <button
                  v-if="selectedWidget.elements?.length"
                  class="btn btn-sm btn-outline-secondary w-100 mb-2"
                  @click="updateWidgetElements(selectedWidget.id, [])"
                >
                  <i class="bi bi-arrow-counterclockwise me-1"></i>Reset element layout
                </button>
                <button class="btn btn-sm btn-outline-danger w-100" @click="removeWidget(selectedWidget.id)">
                  <i class="bi bi-trash me-1"></i>Remove widget
                </button>
              </template>
              <p v-else class="text-muted small mb-0">
                Click a widget on the board to edit it, or add a new one above.
              </p>
            </div>
          </div>
        </div>

        <div class="col-12 col-lg-9">
          <div class="card shadow-sm">
            <div class="card-header bg-white d-flex justify-content-between align-items-center">
              <strong>Board</strong>
              <span class="text-muted small">Hover a widget — drag ⠿ to move, ⚙ to configure, drag corner to resize</span>
            </div>
            <div class="card-body">
              <div
                class="template-canvas"
                :style="{
                  aspectRatio: `${form.width} / ${form.height}`,
                  '--board-cols': gridOverlayLines,
                  '--board-rows': gridOverlayLines,
                  '--canvas-ratio': canvasRatio,
                }"
              >
                <div ref="gridEl" class="grid-stack">
                  <div
                    v-for="w in form.widgets"
                    :key="w.id"
                    class="grid-stack-item"
                    :gs-id="w.id"
                    :gs-x="w.x >= 0 ? w.x : undefined"
                    :gs-y="w.y >= 0 ? w.y : undefined"
                    :gs-w="w.w"
                    :gs-h="w.h"
                    :ref="(el) => setItemRef(w.id, el)"
                  >
                    <div class="grid-stack-item-content">
                      <WidgetEditorShell
                        :widget="w"
                        :isSelected="selectedId === w.id"
                        :isElementEditing="selectedId === w.id && elementEditMode"
                        @select="selectWidget(w.id)"
                        @edit-elements="enterElementEdit(w.id)"
                        @remove="removeWidget(w.id)"
                        @update-elements="updateWidgetElements(w.id, $event)"
                      >
                        <WidgetRenderer :widget="w" :readings="sampleReadings" :history="sampleHistory" />
                      </WidgetEditorShell>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div class="d-flex gap-2 mt-3">
        <button class="btn btn-primary" :disabled="saving" @click="save">
          <span v-if="saving" class="spinner-border spinner-border-sm me-1"></span>
          Save template
        </button>
        <RouterLink class="btn btn-outline-secondary" to="/templates">Cancel</RouterLink>
      </div>
    </template>
  </div>
</template>
