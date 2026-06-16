export interface Factory {
  id: string
  name: string
  location: string
}

export interface Machine {
  id: string
  factoryId: string
  name: string
  type: string
  status: string
}

export interface MetricDef {
  key: string
  label: string
  unit: string
  min: number
  max: number
}

export type WidgetType = 'gauge' | 'line' | 'bar' | 'kpi' | 'status' | 'table'

// Position and size of a sub-element within a widget, as percentages (0-100) of the widget's own area.
export interface WidgetElement {
  key: string // 'title' | 'value' | 'unit' for KPI
  x: number   // % of widget width,  0-100
  y: number   // % of widget height, 0-100
  w: number   // % of widget width,  1-100
  h: number   // % of widget height, 1-100
}

// Position and size as percentages (0-100) of the template's pixel canvas.
export interface Widget {
  id: string
  type: WidgetType
  title: string
  metricKey: string
  x: number // percent of canvas width, 0-99
  y: number // percent of canvas height, 0-99
  w: number // percent of canvas width, 1-100
  h: number // percent of canvas height, 1-100
  elements?: WidgetElement[] // per-element layout; absent = use default CSS layout
}

export interface Template {
  id: string
  name: string
  description: string
  width: number // canvas width in pixels, e.g. 1920
  height: number // canvas height in pixels, e.g. 1080
  widgets: Widget[]
  createdAt: string
  updatedAt: string
}

// Payload for creating/updating a template.
export interface TemplateInput {
  name: string
  description: string
  width: number
  height: number
  widgets: Widget[]
}

export interface DashboardPage {
  id: string
  machineId: string
  machine?: Machine
  position: number
}

export interface Dashboard {
  id: string
  name: string
  templateId: string
  template?: Template
  factoryId: string
  factory?: Factory
  pages?: DashboardPage[]
  pageCount?: number
  createdAt: string
  updatedAt: string
}

export interface TelemetryReading {
  machineId: string
  timestamp: string
  readings: Record<string, number | string>
}
