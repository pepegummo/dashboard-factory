<script setup lang="ts">
import { onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useDashboardStore } from '@/stores/dashboard.store'

const router = useRouter()
const dashboardStore = useDashboardStore()

onMounted(() => dashboardStore.fetchDashboards())

function open(id: string) {
  const { href } = router.resolve({ name: 'dashboard-view', params: { id } })
  window.open(href, '_blank')
}

async function remove(id: string, e: Event) {
  e.stopPropagation()
  if (!confirm('Delete this dashboard?')) return
  await dashboardStore.deleteDashboard(id)
}

function formatDate(value: string): string {
  return new Date(value).toLocaleString()
}
</script>

<template>
  <div>
    <div class="d-flex justify-content-between align-items-center mb-4">
      <div>
        <h1 class="h3 mb-1">Dashboards</h1>
        <p class="text-muted mb-0">Monitoring dashboards generated from your templates.</p>
      </div>
      <RouterLink class="btn btn-primary" to="/create">
        <i class="bi bi-plus-lg me-1"></i>New Dashboard
      </RouterLink>
    </div>

    <div v-if="dashboardStore.loading" class="text-center py-5">
      <div class="spinner-border text-primary" role="status"></div>
    </div>

    <div v-else-if="dashboardStore.dashboards.length" class="row g-3">
      <div v-for="d in dashboardStore.dashboards" :key="d.id" class="col-12 col-md-6 col-xl-4">
        <div class="card h-100 shadow-sm cursor-pointer" @click="open(d.id)">
          <div class="card-body d-flex flex-column">
            <div class="d-flex justify-content-between align-items-start">
              <h5 class="card-title mb-1">{{ d.name }}</h5>
              <button
                class="btn btn-sm btn-outline-danger"
                title="Delete dashboard"
                @click="remove(d.id, $event)"
              >
                <i class="bi bi-trash"></i>
              </button>
            </div>
            <p class="text-muted small mb-2">
              <i class="bi bi-building me-1"></i>{{ d.factory?.name }}
            </p>
            <p class="text-muted small mb-3">
              <i class="bi bi-layout-text-window-reverse me-1"></i>Template:
              {{ d.template?.name }}
            </p>
            <div class="mt-auto d-flex justify-content-between align-items-center">
              <span class="badge bg-light text-dark border">
                <i class="bi bi-cpu me-1"></i>{{ d.pageCount }}
                {{ d.pageCount === 1 ? 'machine' : 'machines' }}
              </span>
              <small class="text-muted">{{ formatDate(d.createdAt) }}</small>
            </div>
          </div>
        </div>
      </div>
    </div>

    <div v-else class="text-center py-5">
      <i class="bi bi-grid-1x2 display-4 text-muted"></i>
      <p class="text-muted mt-3 mb-3">No dashboards yet.</p>
      <RouterLink class="btn btn-primary" to="/create">
        <i class="bi bi-plus-lg me-1"></i>Create your first dashboard
      </RouterLink>
    </div>
  </div>
</template>
