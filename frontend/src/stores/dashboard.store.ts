import { defineStore } from 'pinia'
import { api } from '@/services/api'
import type { Dashboard } from '@/types'

export const useDashboardStore = defineStore('dashboards', {
  state: () => ({
    dashboards: [] as Dashboard[],
    loading: false,
  }),
  actions: {
    async fetchDashboards() {
      this.loading = true
      try {
        this.dashboards = await api.getDashboards()
      } finally {
        this.loading = false
      }
    },
    async getDashboard(id: string) {
      return api.getDashboard(id)
    },
    async createDashboard(input: {
      name: string
      templateId: string
      factoryId: string
      machineIds: string[]
    }) {
      return api.createDashboard(input)
    },
    async deleteDashboard(id: string) {
      await api.deleteDashboard(id)
      this.dashboards = this.dashboards.filter((d) => d.id !== id)
    },
  },
})
