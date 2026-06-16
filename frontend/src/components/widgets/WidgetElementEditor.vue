<script setup lang="ts">
import { ref, watch } from 'vue'
import type { Widget, WidgetElement } from '@/types'

const props = defineProps<{ widget: Widget }>()
const emit = defineEmits<{ 'update:elements': [WidgetElement[]] }>()

const DEFAULT_ELEMENTS: Partial<Record<string, WidgetElement[]>> = {
  kpi: [
    { key: 'title', x: 0, y: 0,  w: 100, h: 25 },
    { key: 'value', x: 0, y: 25, w: 100, h: 50 },
    { key: 'unit',  x: 0, y: 75, w: 100, h: 25 },
  ],
}

const containerRef = ref<HTMLElement | null>(null)
const elements = ref<WidgetElement[]>([])
let dragging = false

function initElements() {
  const src = (props.widget.elements ?? []).length > 0
    ? props.widget.elements!
    : (DEFAULT_ELEMENTS[props.widget.type] ?? [])
  elements.value = src.map((e) => ({ ...e }))
}

initElements()

watch(
  () => props.widget.elements,
  () => { if (!dragging) initElements() },
  { deep: true },
)

function clamp(v: number, lo: number, hi: number) {
  return Math.min(hi, Math.max(lo, v))
}

function elementStyle(el: WidgetElement) {
  return { left: el.x + '%', top: el.y + '%', width: el.w + '%', height: el.h + '%' }
}

// Replace element at key with partial changes; triggers Vue splice reactivity.
function applyUpdate(key: string, changes: Partial<Omit<WidgetElement, 'key'>>) {
  const idx = elements.value.findIndex((e) => e.key === key)
  if (idx === -1) return
  elements.value.splice(idx, 1, { ...elements.value[idx], ...changes })
  emit('update:elements', elements.value.map((e) => ({ ...e })))
}

function startDrag(el: WidgetElement, e: MouseEvent) {
  e.preventDefault()
  const container = containerRef.value
  if (!container) return
  const rect = container.getBoundingClientRect()
  const startMx = e.clientX
  const startMy = e.clientY
  const startX = el.x
  const startY = el.y
  const key = el.key
  dragging = true

  function onMove(me: MouseEvent) {
    const cur = elements.value.find((e) => e.key === key)
    if (!cur) return
    const dx = (me.clientX - startMx) / rect.width  * 100
    const dy = (me.clientY - startMy) / rect.height * 100
    applyUpdate(key, {
      x: Math.round(clamp(startX + dx, 0, 100 - cur.w)),
      y: Math.round(clamp(startY + dy, 0, 100 - cur.h)),
    })
  }

  function onUp() {
    dragging = false
    window.removeEventListener('mousemove', onMove)
    window.removeEventListener('mouseup', onUp)
  }

  window.addEventListener('mousemove', onMove)
  window.addEventListener('mouseup', onUp)
}

function startResize(el: WidgetElement, e: MouseEvent) {
  e.preventDefault()
  const container = containerRef.value
  if (!container) return
  const rect = container.getBoundingClientRect()
  const startMx = e.clientX
  const startMy = e.clientY
  const startW = el.w
  const startH = el.h
  const key = el.key
  dragging = true

  function onMove(me: MouseEvent) {
    const cur = elements.value.find((e) => e.key === key)
    if (!cur) return
    const dx = (me.clientX - startMx) / rect.width  * 100
    const dy = (me.clientY - startMy) / rect.height * 100
    applyUpdate(key, {
      w: Math.round(clamp(startW + dx, 10, 100 - cur.x)),
      h: Math.round(clamp(startH + dy, 10, 100 - cur.y)),
    })
  }

  function onUp() {
    dragging = false
    window.removeEventListener('mousemove', onMove)
    window.removeEventListener('mouseup', onUp)
  }

  window.addEventListener('mousemove', onMove)
  window.addEventListener('mouseup', onUp)
}
</script>

<template>
  <div ref="containerRef" class="widget-element-editor-overlay" @click.stop @mousedown.stop>
    <div
      v-for="el in elements"
      :key="el.key"
      class="element-handle"
      :style="elementStyle(el)"
    >
      <div class="element-drag-body" @mousedown.stop="startDrag(el, $event)">
        <span class="element-label">{{ el.key }}</span>
      </div>
      <div class="element-resize-se" @mousedown.stop="startResize(el, $event)"></div>
    </div>
  </div>
</template>
