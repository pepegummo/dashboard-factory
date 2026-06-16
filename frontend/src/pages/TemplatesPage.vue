<script setup lang="ts">
import { onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useTemplateStore } from '@/stores/template.store'
import { apiErrorMessage } from '@/services/api'

const router = useRouter()
const templateStore = useTemplateStore()

onMounted(() => templateStore.fetchTemplates())

function edit(id: string) {
  router.push(`/templates/${id}`)
}

async function remove(id: string, e: Event) {
  e.stopPropagation()
  if (!confirm('Delete this template?')) return
  try {
    await templateStore.deleteTemplate(id)
  } catch (err) {
    alert(apiErrorMessage(err, 'Failed to delete template'))
  }
}
</script>

<template>
  <div>
    <div class="d-flex justify-content-between align-items-center mb-4">
      <div>
        <h1 class="h3 mb-1">Dashboard Templates</h1>
        <p class="text-muted mb-0">
          Reusable widget layouts. Apply a template to a factory's machines to generate a
          monitoring dashboard.
        </p>
      </div>
      <RouterLink class="btn btn-primary" to="/templates/new">
        <i class="bi bi-plus-lg me-1"></i>New Template
      </RouterLink>
    </div>

    <div v-if="templateStore.loading" class="text-center py-5">
      <div class="spinner-border text-primary" role="status"></div>
    </div>

    <div v-else-if="templateStore.templates.length" class="row g-3">
      <div v-for="t in templateStore.templates" :key="t.id" class="col-12 col-md-6 col-xl-4">
        <div class="card h-100 shadow-sm cursor-pointer" @click="edit(t.id)">
          <div class="card-body d-flex flex-column">
            <div class="d-flex justify-content-between align-items-start">
              <h5 class="card-title mb-1">{{ t.name }}</h5>
              <button
                class="btn btn-sm btn-outline-danger"
                title="Delete template"
                @click="remove(t.id, $event)"
              >
                <i class="bi bi-trash"></i>
              </button>
            </div>
            <p class="text-muted small flex-grow-1">{{ t.description || 'No description' }}</p>
            <div class="d-flex flex-wrap gap-2">
              <span class="badge bg-light text-dark border">
                <i class="bi bi-grid-3x3-gap me-1"></i>{{ t.widgets.length }}
                {{ t.widgets.length === 1 ? 'widget' : 'widgets' }}
              </span>
              <span class="badge bg-light text-dark border">
                <i class="bi bi-aspect-ratio me-1"></i>{{ t.width }} x {{ t.height }} px
              </span>
            </div>
          </div>
        </div>
      </div>
    </div>

    <div v-else class="text-center py-5">
      <i class="bi bi-layout-text-window-reverse display-4 text-muted"></i>
      <p class="text-muted mt-3 mb-3">No templates yet.</p>
      <RouterLink class="btn btn-primary" to="/templates/new">
        <i class="bi bi-plus-lg me-1"></i>Create your first template
      </RouterLink>
    </div>
  </div>
</template>
