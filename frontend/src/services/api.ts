import axios from 'axios'
import type {
  Dashboard,
  Factory,
  Machine,
  MetricDef,
  Template,
  TemplateInput,
  TelemetryReading,
} from '@/types'

const http = axios.create({
  baseURL: '/api',
})

export const api = {
  // Catalog
  getFactories: () => http.get<Factory[]>('/factories').then((r) => r.data),
  getMachines: (factoryId: string) =>
    http.get<Machine[]>(`/factories/${factoryId}/machines`).then((r) => r.data),
  getMetrics: () => http.get<MetricDef[]>('/metrics').then((r) => r.data),

  // Templates
  getTemplates: () => http.get<Template[]>('/templates').then((r) => r.data),
  getTemplate: (id: string) => http.get<Template>(`/templates/${id}`).then((r) => r.data),
  createTemplate: (input: TemplateInput) =>
    http.post<Template>('/templates', input).then((r) => r.data),
  updateTemplate: (id: string, input: TemplateInput) =>
    http.put<Template>(`/templates/${id}`, input).then((r) => r.data),
  deleteTemplate: (id: string) => http.delete(`/templates/${id}`),

  // Dashboards
  getDashboards: () => http.get<Dashboard[]>('/dashboards').then((r) => r.data),
  getDashboard: (id: string) => http.get<Dashboard>(`/dashboards/${id}`).then((r) => r.data),
  createDashboard: (input: {
    name: string
    templateId: string
    factoryId: string
    machineIds: string[]
  }) => http.post<{ id: string }>('/dashboards', input).then((r) => r.data),
  deleteDashboard: (id: string) => http.delete(`/dashboards/${id}`),

  // Telemetry
  getTelemetry: (machineId: string) =>
    http.get<TelemetryReading>(`/machines/${machineId}/telemetry`).then((r) => r.data),
}

export function apiErrorMessage(err: unknown, fallback = 'Something went wrong'): string {
  if (axios.isAxiosError(err)) {
    const data = err.response?.data as { error?: string } | undefined
    if (data?.error) return data.error
  }
  return fallback
}
