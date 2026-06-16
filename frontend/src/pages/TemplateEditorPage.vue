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
import { DEFAULT_ELEMENTS } from '@/utils/widgetElements'

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
  gridCols: 100,
  gridRows: 100,
  widgets: [] as Widget[],
})


const loading = ref(false)
const saving = ref(false)
const error = ref('')

const activeTab = ref<'settings' | 'widgets' | 'config'>('settings')

const selectedId = ref<string | null>(null)
const selectedElementKey = ref<string | null>(null)

const WIDGET_TYPE_COLORS: Record<string, string> = {
  gauge: '#0d6efd', line: '#0dcaf0', kpi: '#198754',
  status: '#fd7e14', table: '#6c757d', bar: '#6610f2',
}
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

function pctToCol(pct: number) { return Math.round(pct * form.gridCols / 100) }
function pctToRow(pct: number) { return Math.round(pct * form.gridRows / 100) }
function colToPct(col: number) { return Math.round(col / form.gridCols * 100) }
function rowToPct(row: number) { return Math.round(row / form.gridRows * 100) }

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
  grid.cellHeight(cellWidthPx * form.gridCols * aspectRatio.value / form.gridRows)
}

watch([() => form.width, () => form.height], () => {
  nextTick(() => recomputeCellHeight())
})

watch([() => form.gridCols, () => form.gridRows], async () => {
  if (grid) { grid.destroy(false); grid = null }
  await nextTick()
  initGrid()
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
      form.gridCols = t.gridCols || 100
      form.gridRows = t.gridRows || 100
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
      column: form.gridCols,
      maxRow: form.gridRows,
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
    if (item.x != null) widget.x = colToPct(item.x)
    if (item.y != null) widget.y = rowToPct(item.y)
    if (item.w != null) widget.w = colToPct(item.w)
    if (item.h != null) widget.h = rowToPct(item.h)
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
  selectedElementKey.value = null
  activeTab.value = 'config'
  const w = form.widgets.find((w) => w.id === id)
  if (w && !w.elements?.length && DEFAULT_ELEMENTS[w.type]) {
    updateWidgetElements(w.id, DEFAULT_ELEMENTS[w.type]!.map((e) => ({ ...e })))
  }
}

function selectElement(key: string) {
  selectedElementKey.value = selectedElementKey.value === key ? null : key
}

function updateWidgetElements(id: string, els: WidgetElement[]) {
  const w = form.widgets.find((w) => w.id === id)
  if (w) w.elements = els
}

const selectedWidgetElements = computed<WidgetElement[]>(() => {
  const w = selectedWidget.value
  if (!w) return []
  return (w.elements?.length ? w.elements : DEFAULT_ELEMENTS[w.type] ?? []).map((e) => ({ ...e }))
})

function updateElementProp(key: string, prop: keyof Omit<WidgetElement, 'key'>, rawVal: number) {
  const w = selectedWidget.value
  if (!w) return
  const val = Math.round(Math.min(100, Math.max(prop === 'w' || prop === 'h' ? 1 : 0, rawVal)))
  const base = w.elements?.length ? w.elements : (DEFAULT_ELEMENTS[w.type] ?? [])
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
    if (el && grid) grid.update(el, { w: Math.max(1, pctToCol(clamped)) })
  } else {
    widget.h = clamped
    if (el && grid) grid.update(el, { h: Math.max(1, pctToRow(clamped)) })
  }
}

function validate(): string | null {
  if (!form.name.trim()) return 'Template name is required.'
  if (form.width < 1 || form.width > 10000 || form.height < 1 || form.height > 10000) {
    return 'Canvas size must be between 1 and 10000 pixels.'
  }
  if (form.gridCols < 1 || form.gridCols > 1000 || form.gridRows < 1 || form.gridRows > 1000) {
    return 'Grid constraints must be between 1 and 1000.'
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
      gridCols: form.gridCols,
      gridRows: form.gridRows,
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
  <!-- Loading -->
  <div v-if="loading" class="text-center py-5">
    <div class="spinner-border text-primary" role="status"></div>
  </div>

  <div v-else class="editor-page">
    <!-- ── Page header ── -->
    <div class="editor-page-header">
      <div class="d-flex justify-content-between align-items-center">
        <div>
          <h1 class="h3 mb-1">{{ isEdit ? 'Edit Template' : 'New Template' }}</h1>
          <p class="text-muted mb-0">
            Design the widget board that will be applied to each machine page.
          </p>
        </div>
        <RouterLink class="btn btn-outline-secondary btn-sm" to="/templates">
          <i class="bi bi-arrow-left me-1"></i>Back
        </RouterLink>
      </div>
      <div v-if="error" class="alert alert-danger mt-3 mb-0 py-2">{{ error }}</div>
    </div>

    <!-- ── Editor layout: sidebar + canvas ── -->
    <div class="editor-layout">

      <!-- Left sidebar -->
      <aside class="editor-sidebar">

        <!-- Tab bar -->
        <div class="sidebar-tabs">
          <button
            type="button"
            class="sidebar-tab"
            :class="{ 'sidebar-tab--active': activeTab === 'settings' }"
            @click="activeTab = 'settings'"
          >
            <i class="bi bi-gear me-1"></i>Settings
          </button>
          <button
            type="button"
            class="sidebar-tab"
            :class="{ 'sidebar-tab--active': activeTab === 'widgets' }"
            @click="activeTab = 'widgets'"
          >
            <i class="bi bi-grid me-1"></i>Widgets
            <span v-if="form.widgets.length" class="sidebar-tab-count">{{ form.widgets.length }}</span>
          </button>
          <button
            type="button"
            class="sidebar-tab"
            :class="{ 'sidebar-tab--active': activeTab === 'config', 'sidebar-tab--dim': !selectedWidget }"
            @click="activeTab = 'config'"
          >
            <i class="bi bi-sliders me-1"></i>Config
          </button>
        </div>

        <!-- Tab panels -->
        <div class="editor-sidebar-body">

          <!-- ── Settings tab ── -->
          <div v-show="activeTab === 'settings'" class="sidebar-panel">
            <div class="mb-3">
              <label class="form-label form-label-sm">Name</label>
              <input
                v-model="form.name"
                type="text"
                class="form-control form-control-sm"
                placeholder="e.g. Standard Machine Monitor"
              />
            </div>
            <div class="mb-3">
              <label class="form-label form-label-sm">Description</label>
              <textarea
                v-model="form.description"
                class="form-control form-control-sm"
                rows="2"
                placeholder="Optional"
              ></textarea>
            </div>
            <div class="mb-3">
              <label class="form-label form-label-sm mb-1">Canvas ratio</label>
              <div class="d-flex align-items-center gap-2">
                <input
                  v-model.number="canvasWidthInput"
                  type="number" min="1" max="999"
                  class="form-control form-control-sm"
                  style="width: 4rem"
                />
                <span class="text-muted small">:</span>
                <input
                  v-model.number="canvasHeightInput"
                  type="number" min="1" max="999"
                  class="form-control form-control-sm"
                  style="width: 4rem"
                />
              </div>
              <p class="form-text mb-0">Width : height, e.g. 16 : 9</p>
            </div>
            <div>
              <label class="form-label form-label-sm mb-1">Grid snap</label>
              <div class="d-flex align-items-center gap-2">
                <input
                  v-model.number="form.gridCols"
                  type="number" min="1" max="1000"
                  class="form-control form-control-sm"
                  style="width: 4rem"
                />
                <span class="text-muted small">×</span>
                <input
                  v-model.number="form.gridRows"
                  type="number" min="1" max="1000"
                  class="form-control form-control-sm"
                  style="width: 4rem"
                />
              </div>
              <p class="form-text mb-0">Cols × rows — higher = finer precision</p>
            </div>
          </div>

          <!-- ── Widgets tab ── -->
          <div v-show="activeTab === 'widgets'" class="sidebar-panel">
            <div class="sidebar-group-title">Add</div>
            <div class="d-flex flex-wrap gap-2 mb-4">
              <button
                v-for="opt in widgetTypeOptions"
                :key="opt.value"
                type="button"
                class="btn btn-sm btn-outline-primary"
                :disabled="!canAddWidget[opt.value]"
                :title="canAddWidget[opt.value] ? '' : 'Not enough canvas space'"
                @click="addWidget(opt.value)"
              >
                <i :class="['bi', opt.icon, 'me-1']"></i>{{ opt.label }}
              </button>
            </div>

            <div class="sidebar-group-title">On canvas</div>
            <div v-if="form.widgets.length" class="widget-layer-list">
              <button
                v-for="w in form.widgets"
                :key="w.id"
                type="button"
                class="widget-layer-item"
                :class="{ 'widget-layer-item--active': selectedId === w.id }"
                @click="selectWidget(w.id)"
              >
                <span
                  class="widget-layer-dot"
                  :style="{ backgroundColor: WIDGET_TYPE_COLORS[w.type] ?? '#6c757d' }"
                ></span>
                <span class="widget-layer-name">{{ w.title || '(untitled)' }}</span>
                <span class="widget-layer-type">{{ w.type }}</span>
                <button
                  type="button"
                  class="widget-layer-remove btn btn-sm"
                  title="Remove"
                  @click.stop="removeWidget(w.id)"
                >
                  <i class="bi bi-x"></i>
                </button>
              </button>
            </div>
            <p v-else class="text-muted small mb-0">No widgets yet. Add one above.</p>
          </div>

          <!-- ── Config tab ── -->
          <div v-show="activeTab === 'config'" class="sidebar-panel">
            <template v-if="selectedWidget">
              <div class="d-flex align-items-center gap-2 mb-3">
                <span
                  class="widget-type-pill"
                  :style="{ backgroundColor: WIDGET_TYPE_COLORS[selectedWidget.type] ?? '#6c757d' }"
                >{{ selectedWidget.type }}</span>
                <span class="text-muted small">{{ selectedWidget.title }}</span>
              </div>

              <div v-if="selectedElementKey !== null" class="alert alert-info py-1 px-2 small mb-2">
                <i class="bi bi-arrows-move me-1"></i>Drag handle to reposition. Click row to deselect.
              </div>

              <div class="mb-2">
                <label class="form-label form-label-sm text-muted mb-1">Title</label>
                <input v-model="selectedWidget.title" type="text" class="form-control form-control-sm" />
              </div>
              <div class="mb-2">
                <label class="form-label form-label-sm text-muted mb-1">Type</label>
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
                <label class="form-label form-label-sm text-muted mb-1">Metric</label>
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
                  <label class="form-label form-label-sm text-muted mb-1">Width (%)</label>
                  <input
                    v-model.number="widthInput"
                    type="number" min="1" max="100"
                    class="form-control form-control-sm"
                  />
                </div>
                <div class="col-6">
                  <label class="form-label form-label-sm text-muted mb-1">Height (%)</label>
                  <input
                    v-model.number="heightInput"
                    type="number" min="1" max="100"
                    class="form-control form-control-sm"
                  />
                </div>
              </div>

              <!-- Element positions -->
              <template v-if="selectedWidgetElements.length">
                <hr class="my-2" />
                <label class="form-label form-label-sm text-muted mb-1">
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
                    @click="selectElement(el.key)"
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

            <div v-else class="config-empty">
              <i class="bi bi-cursor display-6 text-muted"></i>
              <p class="text-muted small mt-2 mb-0">Click a widget on the board to configure it</p>
            </div>
          </div>

        </div>

        <!-- Sticky save footer -->
        <div class="editor-sidebar-footer">
          <button class="btn btn-primary btn-sm w-100 mb-2" :disabled="saving" @click="save">
            <span v-if="saving" class="spinner-border spinner-border-sm me-1"></span>
            <i v-else class="bi bi-floppy me-1"></i>
            {{ saving ? 'Saving…' : 'Save template' }}
          </button>
          <RouterLink class="btn btn-outline-secondary btn-sm w-100" to="/templates">
            Cancel
          </RouterLink>
        </div>
      </aside>

      <!-- Canvas main area -->
      <main class="editor-main">
        <div class="card shadow-sm h-100">
          <div class="card-header bg-white d-flex justify-content-between align-items-center py-2">
            <strong>Board</strong>
            <span class="text-muted small">Hover a widget — drag ⠿ to move, ⚙ to configure, drag corner to resize</span>
          </div>
          <div class="card-body p-2 overflow-auto">
            <div
              class="template-canvas"
              :style="{
                aspectRatio: `${form.width} / ${form.height}`,
                '--board-cols': form.gridCols,
                '--board-rows': form.gridRows,
                '--canvas-ratio': canvasRatio,
              }"
            >
              <div ref="gridEl" class="grid-stack">
                <div
                  v-for="w in form.widgets"
                  :key="w.id"
                  class="grid-stack-item"
                  :gs-id="w.id"
                  :gs-x="w.x >= 0 ? pctToCol(w.x) : undefined"
                  :gs-y="w.y >= 0 ? pctToRow(w.y) : undefined"
                  :gs-w="Math.max(1, pctToCol(w.w))"
                  :gs-h="Math.max(1, pctToRow(w.h))"
                  :ref="(el) => setItemRef(w.id, el)"
                >
                  <div class="grid-stack-item-content">
                    <WidgetEditorShell
                      :widget="w"
                      :isSelected="selectedId === w.id"
                      :isElementEditing="selectedId === w.id"
                      :activeElementKey="selectedId === w.id ? selectedElementKey ?? undefined : undefined"
                      @select="selectWidget(w.id)"
                      @remove="removeWidget(w.id)"
                      @update-elements="updateWidgetElements(w.id, $event)"
                      @select-element="selectElement($event)"
                    >
                      <WidgetRenderer :widget="w" :readings="sampleReadings" :history="sampleHistory" />
                    </WidgetEditorShell>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </main>

    </div>
  </div>
</template>

<style scoped>
/* ── Page shell ── */
.editor-page {
  display: flex;
  flex-direction: column;
  height: calc(100vh - 56px - 3rem); /* 56px navbar + 3rem page-content padding */
  margin: -1.5rem;
}

.editor-page-header {
  padding: 1rem 1.5rem 0.75rem;
  border-bottom: 1px solid var(--bs-border-color);
  background: #fff;
  flex-shrink: 0;
}

/* ── Two-column layout ── */
.editor-layout {
  display: flex;
  flex: 1;
  overflow: hidden;
}

/* ── Left sidebar ── */
.editor-sidebar {
  width: 280px;
  min-width: 280px;
  background: #fff;
  border-right: 1px solid var(--bs-border-color);
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

/* Tab bar */
.sidebar-tabs {
  display: flex;
  border-bottom: 1px solid var(--bs-border-color);
  flex-shrink: 0;
}

.sidebar-tab {
  flex: 1;
  padding: 0.6rem 0.25rem;
  font-size: 0.75rem;
  font-weight: 500;
  background: none;
  border: none;
  border-bottom: 2px solid transparent;
  color: var(--bs-secondary);
  cursor: pointer;
  transition: color 0.1s, border-color 0.1s;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 3px;
  white-space: nowrap;
}

.sidebar-tab:hover {
  color: var(--bs-body-color);
}

.sidebar-tab--active {
  color: var(--bs-primary);
  border-bottom-color: var(--bs-primary);
  font-weight: 600;
}

.sidebar-tab--dim {
  opacity: 0.5;
}

.sidebar-tab-count {
  background: var(--bs-secondary-bg);
  color: var(--bs-secondary);
  border-radius: 10px;
  font-size: 0.65rem;
  font-weight: 600;
  padding: 0 5px;
  line-height: 1.4;
}

.sidebar-tab--active .sidebar-tab-count {
  background: rgba(var(--bs-primary-rgb), 0.12);
  color: var(--bs-primary);
}

/* Tab body */
.editor-sidebar-body {
  flex: 1;
  overflow-y: auto;
}

.sidebar-panel {
  padding: 1rem;
}

.sidebar-group-title {
  font-size: 0.68rem;
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: var(--bs-secondary);
  margin-bottom: 0.6rem;
}

/* Widget layers list (Widgets tab) */
.widget-layer-list {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.widget-layer-item {
  display: flex;
  align-items: center;
  gap: 8px;
  width: 100%;
  padding: 5px 8px;
  background: none;
  border: 1px solid transparent;
  border-radius: 5px;
  text-align: left;
  cursor: pointer;
  font-size: 0.8rem;
  transition: background 0.1s;
}

.widget-layer-item:hover {
  background: var(--bs-light);
}

.widget-layer-item--active {
  background: rgba(var(--bs-primary-rgb), 0.08);
  border-color: rgba(var(--bs-primary-rgb), 0.2);
}

.widget-layer-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  flex-shrink: 0;
}

.widget-layer-name {
  flex: 1;
  font-weight: 500;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.widget-layer-type {
  font-size: 0.68rem;
  color: var(--bs-secondary);
  flex-shrink: 0;
}

.widget-layer-remove {
  padding: 0 4px;
  line-height: 1;
  color: var(--bs-secondary);
  opacity: 0;
  transition: opacity 0.1s;
}

.widget-layer-item:hover .widget-layer-remove {
  opacity: 1;
}

/* Config tab */
.widget-type-pill {
  display: inline-block;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 0.7rem;
  font-weight: 600;
  color: #fff;
  flex-shrink: 0;
}

.config-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 2rem 1rem;
  text-align: center;
}

.editor-sidebar-footer {
  padding: 0.75rem 1rem;
  border-top: 1px solid var(--bs-border-color);
  background: #fff;
  flex-shrink: 0;
}

/* ── Canvas main area ── */
.editor-main {
  flex: 1;
  overflow: auto;
  padding: 1rem;
  background: #f4f6f9;
  display: flex;
  flex-direction: column;
}

.editor-main .card {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-height: 0;
}

.editor-main .card-body {
  flex: 1;
  min-height: 0;
}

/* ── Mobile: stack sidebar above canvas ── */
@media (max-width: 991.98px) {
  .editor-page {
    height: auto;
  }

  .editor-layout {
    flex-direction: column;
    overflow: visible;
  }

  .editor-sidebar {
    width: 100%;
    min-width: unset;
    border-right: none;
    border-bottom: 1px solid var(--bs-border-color);
  }

  .editor-sidebar-body {
    max-height: 50vh;
  }

  .editor-main {
    padding: 0.75rem;
    min-height: 60vh;
  }
}
</style>
