<script setup lang="ts">
import { ref, watch, watchEffect, onMounted, onUnmounted } from 'vue'
import { RouterLink, RouterView, useRoute } from 'vue-router'

const route = useRoute()

type Mode = 'system' | 'light' | 'dark'
const mode = ref<Mode>((localStorage.getItem('theme') as Mode) ?? 'system')
let mq: MediaQueryList

function apply(m: Mode) {
  const resolved = m === 'system' ? (mq?.matches ? 'dark' : 'light') : m
  document.documentElement.setAttribute('data-bs-theme', resolved)
}

watch(mode, m => localStorage.setItem('theme', m))

onMounted(() => {
  mq = window.matchMedia('(prefers-color-scheme: dark)')
  mq.addEventListener('change', () => { if (mode.value === 'system') apply('system') })
  apply(mode.value)
})
onUnmounted(() => mq.removeEventListener('change', () => {}))
watchEffect(() => apply(mode.value))
</script>

<template>
  <div class="app-shell">
    <nav v-if="!route.meta.standalone" class="navbar navbar-expand-lg navbar-dark bg-dark">
      <div class="container-fluid">
        <RouterLink class="navbar-brand fw-semibold" to="/">
          <i class="bi bi-speedometer2 me-2"></i>Factory Dashboard Builder
        </RouterLink>
        <button
          class="navbar-toggler"
          type="button"
          data-bs-toggle="collapse"
          data-bs-target="#mainNav"
        >
          <span class="navbar-toggler-icon"></span>
        </button>
        <div id="mainNav" class="collapse navbar-collapse">
          <ul class="navbar-nav me-auto mb-2 mb-lg-0">
            <li class="nav-item">
              <RouterLink
                class="nav-link"
                :class="{ active: route.name === 'dashboards' }"
                to="/"
              >
                <i class="bi bi-grid-1x2 me-1"></i>Dashboards
              </RouterLink>
            </li>
            <li class="nav-item">
              <RouterLink
                class="nav-link"
                :class="{ active: route.name === 'templates' || route.name?.toString().startsWith('template-') }"
                to="/templates"
              >
                <i class="bi bi-layout-text-window-reverse me-1"></i>Templates
              </RouterLink>
            </li>
          </ul>
          <select v-model="mode" class="form-select form-select-sm w-auto me-2">
            <option value="system">System</option>
            <option value="light">Light</option>
            <option value="dark">Dark</option>
          </select>
          <RouterLink class="btn btn-primary" to="/create">
            <i class="bi bi-plus-lg me-1"></i>New Dashboard
          </RouterLink>
        </div>
      </div>
    </nav>

    <main class="page-content" :class="{ 'page-content-standalone': route.meta.standalone }">
      <div class="container-fluid" :class="{ 'px-0': route.meta.standalone }">
        <RouterView />
      </div>
    </main>
  </div>
</template>
