<script setup lang="ts">
import { RouterLink, RouterView, useRoute } from 'vue-router'

const route = useRoute()
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
