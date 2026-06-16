<script setup lang="ts">
defineProps<{
  title: string
  value: string | number | null | undefined
}>()

const statusStyles: Record<string, string> = {
  running: 'bg-success',
  idle: 'bg-warning text-dark',
  stopped: 'bg-secondary',
  error: 'bg-danger',
}

function badgeClass(value: string | number | null | undefined): string {
  const key = String(value ?? '').toLowerCase()
  return statusStyles[key] ?? 'bg-secondary'
}
</script>

<template>
  <div class="card widget-card shadow-sm">
    <div class="card-body text-center justify-content-center">
      <h6 class="card-subtitle text-muted widget-title">{{ title }}</h6>
      <span class="badge status-badge text-uppercase" :class="badgeClass(value)">
        {{ value ?? '—' }}
      </span>
    </div>
  </div>
</template>
