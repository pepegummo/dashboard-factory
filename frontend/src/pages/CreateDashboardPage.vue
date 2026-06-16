<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useTemplateStore } from '@/stores/template.store'
import { useCatalogStore } from '@/stores/catalog.store'
import { useDashboardStore } from '@/stores/dashboard.store'
import { apiErrorMessage } from '@/services/api'

const router = useRouter()
const templateStore = useTemplateStore()
const catalog = useCatalogStore()
const dashboardStore = useDashboardStore()

const step = ref(1)
const totalSteps = 4

const selectedTemplateId = ref<string | null>(null)
const selectedFactoryId = ref<string | null>(null)
const selectedMachineIds = ref<string[]>([])
const dashboardName = ref('')
const nameTouched = ref(false)

const creating = ref(false)
const error = ref('')

onMounted(async () => {
  await Promise.all([templateStore.fetchTemplates(), catalog.fetchFactories(), catalog.fetchMetrics()])
})

const selectedTemplate = computed(() =>
  templateStore.templates.find((t) => t.id === selectedTemplateId.value),
)
const selectedFactory = computed(() =>
  catalog.factories.find((f) => f.id === selectedFactoryId.value),
)
const factoryMachines = computed(() =>
  selectedFactoryId.value ? catalog.machinesByFactory[selectedFactoryId.value] ?? [] : [],
)
const selectedMachines = computed(() =>
  selectedMachineIds.value
    .map((id) => factoryMachines.value.find((m) => m.id === id))
    .filter((m): m is NonNullable<typeof m> => !!m),
)

watch(selectedFactoryId, async (factoryId) => {
  selectedMachineIds.value = []
  if (factoryId) await catalog.fetchMachines(factoryId)
})

watch([selectedFactory, selectedTemplate], ([factory, template]) => {
  if (nameTouched.value) return
  if (factory && template) {
    dashboardName.value = `${factory.name} – ${template.name}`
  }
})

const statusBadge: Record<string, string> = {
  running: 'bg-success',
  idle: 'bg-warning text-dark',
  stopped: 'bg-secondary',
  error: 'bg-danger',
}

function badgeClass(status: string): string {
  return statusBadge[status.toLowerCase()] ?? 'bg-secondary'
}

function addMachine(id: string) {
  if (!selectedMachineIds.value.includes(id)) {
    selectedMachineIds.value.push(id)
  }
}

function removeMachine(id: string) {
  selectedMachineIds.value = selectedMachineIds.value.filter((m) => m !== id)
}

function moveMachine(index: number, dir: -1 | 1) {
  const target = index + dir
  if (target < 0 || target >= selectedMachineIds.value.length) return
  const ids = selectedMachineIds.value
  ;[ids[index], ids[target]] = [ids[target], ids[index]]
}

const canProceed = computed(() => {
  switch (step.value) {
    case 1:
      return !!selectedTemplateId.value
    case 2:
      return !!selectedFactoryId.value
    case 3:
      return selectedMachineIds.value.length > 0
    case 4:
      return dashboardName.value.trim().length > 0
    default:
      return false
  }
})

function next() {
  if (!canProceed.value) return
  if (step.value < totalSteps) step.value++
}

function back() {
  if (step.value > 1) step.value--
}

async function createDashboard() {
  if (!selectedTemplateId.value || !selectedFactoryId.value || !canProceed.value) return
  error.value = ''
  creating.value = true
  try {
    const result = await dashboardStore.createDashboard({
      name: dashboardName.value.trim(),
      templateId: selectedTemplateId.value,
      factoryId: selectedFactoryId.value,
      machineIds: selectedMachineIds.value,
    })
    router.push(`/dashboards/${result.id}`)
  } catch (err) {
    error.value = apiErrorMessage(err, 'Failed to create dashboard')
  } finally {
    creating.value = false
  }
}
</script>

<template>
  <div>
    <div class="mb-4">
      <h1 class="h3 mb-1">Create Dashboard</h1>
      <p class="text-muted mb-0">
        Pick a template, choose a factory, add machines (one per page), then generate the
        monitoring dashboard.
      </p>
    </div>

    <!-- Stepper -->
    <ul class="nav nav-pills mb-4">
      <li class="nav-item" v-for="(label, i) in ['Template', 'Factory', 'Machines', 'Review']" :key="label">
        <span class="nav-link" :class="{ active: step === i + 1, disabled: step < i + 1 }">
          <span class="badge rounded-pill bg-white text-primary me-1">{{ i + 1 }}</span>
          {{ label }}
        </span>
      </li>
    </ul>

    <div v-if="error" class="alert alert-danger">{{ error }}</div>

    <!-- Step 1: Template -->
    <div v-if="step === 1" class="card shadow-sm">
      <div class="card-body">
        <h5 class="card-title mb-3">1. Choose a template</h5>

        <div v-if="templateStore.loading" class="text-center py-4">
          <div class="spinner-border text-primary" role="status"></div>
        </div>

        <div v-else-if="!templateStore.templates.length" class="text-center py-4">
          <p class="text-muted mb-3">You don't have any templates yet.</p>
          <RouterLink class="btn btn-primary" to="/templates/new">
            <i class="bi bi-plus-lg me-1"></i>Create a template
          </RouterLink>
        </div>

        <div v-else class="row g-3">
          <div v-for="t in templateStore.templates" :key="t.id" class="col-12 col-md-6 col-xl-4">
            <div
              class="card h-100 cursor-pointer border-2"
              :class="selectedTemplateId === t.id ? 'border-primary' : 'border-light'"
              @click="selectedTemplateId = t.id"
            >
              <div class="card-body">
                <div class="form-check mb-1">
                  <input
                    class="form-check-input"
                    type="radio"
                    :checked="selectedTemplateId === t.id"
                    readonly
                  />
                  <label class="form-check-label fw-semibold">{{ t.name }}</label>
                </div>
                <p class="text-muted small mb-2">{{ t.description || 'No description' }}</p>
                <span class="badge bg-light text-dark border">{{ t.widgets.length }} widgets</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Step 2: Factory -->
    <div v-else-if="step === 2" class="card shadow-sm">
      <div class="card-body">
        <h5 class="card-title mb-3">2. Choose a factory</h5>

        <div v-if="catalog.loadingFactories" class="text-center py-4">
          <div class="spinner-border text-primary" role="status"></div>
        </div>

        <div v-else class="row g-3">
          <div v-for="f in catalog.factories" :key="f.id" class="col-12 col-md-6 col-xl-4">
            <div
              class="card h-100 cursor-pointer border-2"
              :class="selectedFactoryId === f.id ? 'border-primary' : 'border-light'"
              @click="selectedFactoryId = f.id"
            >
              <div class="card-body">
                <div class="form-check mb-1">
                  <input
                    class="form-check-input"
                    type="radio"
                    :checked="selectedFactoryId === f.id"
                    readonly
                  />
                  <label class="form-check-label fw-semibold">{{ f.name }}</label>
                </div>
                <p class="text-muted small mb-0"><i class="bi bi-geo-alt me-1"></i>{{ f.location }}</p>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Step 3: Machines -->
    <div v-else-if="step === 3" class="card shadow-sm">
      <div class="card-body">
        <h5 class="card-title mb-3">3. Add machines (one per page)</h5>
        <p class="text-muted small">
          Each machine you add becomes its own dashboard page using the
          <strong>{{ selectedTemplate?.name }}</strong> template.
        </p>

        <div v-if="catalog.loadingMachines" class="text-center py-4">
          <div class="spinner-border text-primary" role="status"></div>
        </div>

        <div v-else class="row g-3">
          <div class="col-12 col-lg-6">
            <h6 class="text-muted">Available machines</h6>
            <div class="list-group">
              <div
                v-for="m in factoryMachines"
                :key="m.id"
                class="list-group-item d-flex justify-content-between align-items-center"
              >
                <div>
                  <div class="fw-semibold">{{ m.name }}</div>
                  <div class="small text-muted">
                    {{ m.type }}
                    <span class="badge ms-1" :class="badgeClass(m.status)">{{ m.status }}</span>
                  </div>
                </div>
                <button
                  class="btn btn-sm btn-outline-primary"
                  :disabled="selectedMachineIds.includes(m.id)"
                  @click="addMachine(m.id)"
                >
                  <i class="bi bi-plus-lg"></i>
                </button>
              </div>
              <div v-if="!factoryMachines.length" class="list-group-item text-muted">
                No machines found for this factory.
              </div>
            </div>
          </div>

          <div class="col-12 col-lg-6">
            <h6 class="text-muted">Dashboard pages ({{ selectedMachines.length }})</h6>
            <div class="list-group">
              <div
                v-for="(m, i) in selectedMachines"
                :key="m.id"
                class="list-group-item d-flex justify-content-between align-items-center"
              >
                <div>
                  <span class="badge bg-primary me-2">Page {{ i + 1 }}</span>
                  <span class="fw-semibold">{{ m.name }}</span>
                </div>
                <div class="d-flex gap-1">
                  <button class="btn btn-sm btn-outline-secondary" :disabled="i === 0" @click="moveMachine(i, -1)">
                    <i class="bi bi-arrow-up"></i>
                  </button>
                  <button class="btn btn-sm btn-outline-secondary" :disabled="i === selectedMachines.length - 1" @click="moveMachine(i, 1)">
                    <i class="bi bi-arrow-down"></i>
                  </button>
                  <button class="btn btn-sm btn-outline-danger" @click="removeMachine(m.id)">
                    <i class="bi bi-x-lg"></i>
                  </button>
                </div>
              </div>
              <div v-if="!selectedMachines.length" class="list-group-item text-muted">
                Add machines from the left to build dashboard pages.
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Step 4: Review -->
    <div v-else class="card shadow-sm">
      <div class="card-body">
        <h5 class="card-title mb-3">4. Review &amp; create</h5>

        <div class="mb-3">
          <label class="form-label">Dashboard name</label>
          <input
            v-model="dashboardName"
            type="text"
            class="form-control"
            @input="nameTouched = true"
          />
        </div>

        <dl class="row mb-0">
          <dt class="col-sm-3">Template</dt>
          <dd class="col-sm-9">{{ selectedTemplate?.name }}</dd>

          <dt class="col-sm-3">Factory</dt>
          <dd class="col-sm-9">{{ selectedFactory?.name }} ({{ selectedFactory?.location }})</dd>

          <dt class="col-sm-3">Pages</dt>
          <dd class="col-sm-9">
            <ol class="mb-0 ps-3">
              <li v-for="m in selectedMachines" :key="m.id">{{ m.name }}</li>
            </ol>
          </dd>
        </dl>
      </div>
    </div>

    <!-- Navigation -->
    <div class="d-flex justify-content-between mt-4">
      <button class="btn btn-outline-secondary" :disabled="step === 1" @click="back">
        <i class="bi bi-arrow-left me-1"></i>Back
      </button>
      <button v-if="step < totalSteps" class="btn btn-primary" :disabled="!canProceed" @click="next">
        Next<i class="bi bi-arrow-right ms-1"></i>
      </button>
      <button v-else class="btn btn-success" :disabled="!canProceed || creating" @click="createDashboard">
        <span v-if="creating" class="spinner-border spinner-border-sm me-1"></span>
        <i v-else class="bi bi-check-lg me-1"></i>Create dashboard
      </button>
    </div>
  </div>
</template>
