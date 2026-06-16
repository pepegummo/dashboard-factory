import { defineStore } from 'pinia'
import { api } from '@/services/api'
import type { Template, TemplateInput } from '@/types'

export const useTemplateStore = defineStore('templates', {
  state: () => ({
    templates: [] as Template[],
    loading: false,
  }),
  actions: {
    async fetchTemplates() {
      this.loading = true
      try {
        this.templates = await api.getTemplates()
      } finally {
        this.loading = false
      }
    },
    async getTemplate(id: string) {
      return api.getTemplate(id)
    },
    async createTemplate(input: TemplateInput) {
      const created = await api.createTemplate(input)
      this.templates.push(created)
      return created
    },
    async updateTemplate(id: string, input: TemplateInput) {
      const updated = await api.updateTemplate(id, input)
      const idx = this.templates.findIndex((t) => t.id === id)
      if (idx !== -1) this.templates[idx] = updated
      return updated
    },
    async deleteTemplate(id: string) {
      await api.deleteTemplate(id)
      this.templates = this.templates.filter((t) => t.id !== id)
    },
  },
})
