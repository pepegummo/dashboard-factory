<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import { useDashboardStore } from '@/stores/dashboard.store'
import { useCatalogStore } from '@/stores/catalog.store'
import { useTelemetry } from '@/composables/useTelemetry'
import { apiErrorMessage } from '@/services/api'
import WidgetRenderer from '@/components/widgets/WidgetRenderer.vue'
import type { Dashboard } from '@/types'

const props = defineProps<{ id: string }>()

const dashboardStore = useDashboardStore()
const catalog = useCatalogStore()
const { reading, history, start } = useTelemetry()

const dashboard = ref<Dashboard | null>(null)
const loading = ref(true)
const error = ref('')
const activePage = ref(0)

const pages = computed(() => dashboard.value?.pages ?? [])
const currentMachine = computed(() => pages.value[activePage.value]?.machine ?? null)
const widgets = computed(() => dashboard.value?.template?.widgets ?? [])
const template = computed(() => dashboard.value?.template ?? null)

const canvasStyle = computed(() => {
  if (!template.value) return undefined
  const ratio = template.value.width / template.value.height
  const overhead = pages.value.length > 1 ? '2.25rem' : '1rem'
  return {
    aspectRatio: `${template.value.width} / ${template.value.height}`,
    width: `min(100%, calc((100vh - ${overhead}) * ${ratio}))`,
  }
})

async function load() {
  loading.value = true
  error.value = ''
  try {
    dashboard.value = await dashboardStore.getDashboard(props.id)
  } catch (err) {
    error.value = apiErrorMessage(err, 'Failed to load dashboard')
  } finally {
    loading.value = false
  }
}

let rotateTimer: ReturnType<typeof setInterval> | undefined

onMounted(async () => {
  await catalog.fetchMetrics()
  await load()
  rotateTimer = setInterval(() => {
    if (pages.value.length > 1) {
      activePage.value = (activePage.value + 1) % pages.value.length
    }
  }, 10000)
})

onUnmounted(() => {
  if (rotateTimer !== undefined) clearInterval(rotateTimer)
})

watch(
  currentMachine,
  (machine) => {
    if (machine) start(machine.id)
  },
  { immediate: true },
)
</script>

<template>
  <div class="db-view-shell">
    <div v-if="loading" class="text-center py-5">
      <div class="spinner-border text-primary" role="status"></div>
    </div>

    <div v-else-if="error" class="alert alert-danger">{{ error }}</div>

    <template v-else-if="dashboard">
      <div class="dashboard-canvas" :style="canvasStyle">
        <div
          v-for="w in widgets"
          :key="w.id"
          class="dashboard-canvas-item"
          :style="{ left: `${w.x}%`, top: `${w.y}%`, width: `${w.w}%`, height: `${w.h}%` }"
        >
          <WidgetRenderer :widget="w" :readings="reading" :history="history" />
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
    </template>
  </div>
</template>

<style scoped>
/* Dark shell so the area around the canvas never flashes white.
   min-height: 100vh ensures full viewport fill even on short canvases
   (body is light-grey by default). padding-bottom gives breathing room
   for the page-dots row; sides and top are zero because the outer
   page-content-standalone already adds 0.5rem on all edges. */
.db-view-shell {
  color-scheme: dark;
  background-color: var(--db-bg);
  min-height: 100vh;
  padding: 0 0 0.5rem;
}
</style>
