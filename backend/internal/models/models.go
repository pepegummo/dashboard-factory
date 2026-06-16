package models

import "time"

// Factory represents a production site that contains machines.
type Factory struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Location string `json:"location"`
}

// Machine represents a single piece of equipment on a factory floor.
type Machine struct {
	ID        string `json:"id"`
	FactoryID string `json:"factoryId"`
	Name      string `json:"name"`
	Type      string `json:"type"`
	Status    string `json:"status"`
}

// MetricDef describes a telemetry metric that widgets can be bound to.
type MetricDef struct {
	Key   string  `json:"key"`
	Label string  `json:"label"`
	Unit  string  `json:"unit"`
	Min   float64 `json:"min"`
	Max   float64 `json:"max"`
}

// WidgetElement positions a single sub-element within a widget as percentages of the widget's own area.
type WidgetElement struct {
	Key string `json:"key"` // e.g. "title", "value", "unit"
	X   int    `json:"x"`   // % of widget width,  0-100
	Y   int    `json:"y"`   // % of widget height, 0-100
	W   int    `json:"w"`   // % of widget width,  1-100
	H   int    `json:"h"`   // % of widget height, 1-100
}

// Widget is a single tile inside a dashboard template, positioned and sized
// as a percentage (0-100) of the template's pixel canvas.
type Widget struct {
	ID        string          `json:"id"`
	Type      string          `json:"type"` // gauge | line | kpi | status | table
	Title     string          `json:"title"`
	MetricKey string          `json:"metricKey"`
	X         int             `json:"x"`                  // percent of canvas width, 0-99
	Y         int             `json:"y"`                  // percent of canvas height, 0-99
	W         int             `json:"w"`                  // percent of canvas width, 1-100
	H         int             `json:"h"`                  // percent of canvas height, 1-100
	Elements  []WidgetElement `json:"elements,omitempty"` // per-element layout; absent = default layout
}

// Template is a reusable dashboard layout made of widgets.
type Template struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Width       int       `json:"width"`  // canvas width in pixels, e.g. 1920
	Height      int       `json:"height"` // canvas height in pixels, e.g. 1080
	Widgets     []Widget  `json:"widgets"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// DashboardPage binds one machine to a dashboard (one machine per page).
type DashboardPage struct {
	ID        string   `json:"id"`
	MachineID string   `json:"machineId"`
	Machine   *Machine `json:"machine,omitempty"`
	Position  int      `json:"position"`
}

// Dashboard is a generated monitoring dashboard: a template applied to a
// factory with one page per selected machine.
type Dashboard struct {
	ID         string          `json:"id"`
	Name       string          `json:"name"`
	TemplateID string          `json:"templateId"`
	Template   *Template       `json:"template,omitempty"`
	FactoryID  string          `json:"factoryId"`
	Factory    *Factory        `json:"factory,omitempty"`
	Pages      []DashboardPage `json:"pages,omitempty"`
	CreatedAt  time.Time       `json:"createdAt"`
	UpdatedAt  time.Time       `json:"updatedAt"`
}
