import { defineStore } from 'pinia'
import { api } from '@/services/api'
import type { Factory, Machine, MetricDef } from '@/types'

export const useCatalogStore = defineStore('catalog', {
  state: () => ({
    factories: [] as Factory[],
    machinesByFactory: {} as Record<string, Machine[]>,
    metrics: [] as MetricDef[],
    loadingFactories: false,
    loadingMachines: false,
    loadingMetrics: false,
  }),
  actions: {
    async fetchFactories() {
      if (this.factories.length) return
      this.loadingFactories = true
      try {
        this.factories = await api.getFactories()
      } finally {
        this.loadingFactories = false
      }
    },
    async fetchMachines(factoryId: string) {
      if (this.machinesByFactory[factoryId]) return
      this.loadingMachines = true
      try {
        this.machinesByFactory[factoryId] = await api.getMachines(factoryId)
      } finally {
        this.loadingMachines = false
      }
    },
    async fetchMetrics() {
      if (this.metrics.length) return
      this.loadingMetrics = true
      try {
        this.metrics = await api.getMetrics()
      } finally {
        this.loadingMetrics = false
      }
    },
    metricLabel(key: string): string {
      return this.metrics.find((m) => m.key === key)?.label ?? key
    },
    metricUnit(key: string): string {
      return this.metrics.find((m) => m.key === key)?.unit ?? ''
    },
  },
})
