package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"dashboard-builder/backend/internal/models"

	"github.com/go-chi/chi/v5"
)

type TemplateHandler struct {
	DB *sql.DB
}

func NewTemplateHandler(db *sql.DB) *TemplateHandler {
	return &TemplateHandler{DB: db}
}

var validWidgetTypes = map[string]bool{
	"gauge":  true,
	"line":   true,
	"bar":    true,
	"kpi":    true,
	"status": true,
	"table":  true,
}

// percentGrid is the size of the percent-based widget layout grid; widget
// x/y/w/h values are percentages (0-100) of the template's pixel canvas
// (Width x Height) and are independent of its pixel dimensions.
const percentGrid = 100

// Bounds for the template's pixel canvas width/height, in pixels.
const (
	minCanvasSize = 1
	maxCanvasSize = 10000
)

// Metrics returns the fixed catalog of telemetry fields widgets can bind to.
func (h *TemplateHandler) Metrics(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, models.Metrics)
}

// List returns every saved template (without widget detail) for picker UIs.
func (h *TemplateHandler) List(w http.ResponseWriter, r *http.Request) {
	rows, err := h.DB.Query(`SELECT id, name, description, width, height, widgets_json, created_at, updated_at FROM templates ORDER BY name`)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	templates := []models.Template{}
	for rows.Next() {
		t, err := scanTemplate(rows)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		templates = append(templates, t)
	}
	writeJSON(w, http.StatusOK, templates)
}

func (h *TemplateHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	row := h.DB.QueryRow(`SELECT id, name, description, width, height, widgets_json, created_at, updated_at FROM templates WHERE id = ?`, id)
	t, err := scanTemplate(row)
	if errors.Is(err, sql.ErrNoRows) {
		writeError(w, http.StatusNotFound, "template not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, t)
}

type templateInput struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Width       int             `json:"width"`
	Height      int             `json:"height"`
	Widgets     []models.Widget `json:"widgets"`
}

func (in *templateInput) validate() error {
	if strings.TrimSpace(in.Name) == "" {
		return errors.New("name is required")
	}

	if in.Width == 0 {
		in.Width = 1920
	}
	if in.Height == 0 {
		in.Height = 1080
	}
	if in.Width < minCanvasSize || in.Width > maxCanvasSize || in.Height < minCanvasSize || in.Height > maxCanvasSize {
		return errors.New("template canvas size must be between 1 and 10000 pixels")
	}

	if len(in.Widgets) == 0 {
		return errors.New("template must have at least one widget")
	}
	for i := range in.Widgets {
		wgt := &in.Widgets[i]
		wgt.Title = strings.TrimSpace(wgt.Title)
		if wgt.Title == "" {
			return errors.New("every widget needs a title")
		}
		if !validWidgetTypes[wgt.Type] {
			return errors.New("invalid widget type: " + wgt.Type)
		}
		if wgt.W < 1 || wgt.H < 1 || wgt.W > percentGrid || wgt.H > percentGrid {
			return errors.New("widget '" + wgt.Title + "' has an invalid size")
		}
		if wgt.X < 0 || wgt.Y < 0 || wgt.X >= percentGrid || wgt.Y >= percentGrid {
			return errors.New("widget '" + wgt.Title + "' is positioned outside the canvas")
		}
		if wgt.X+wgt.W > percentGrid || wgt.Y+wgt.H > percentGrid {
			return errors.New("widget '" + wgt.Title + "' extends outside the canvas")
		}
		if wgt.Type != "table" {
			if wgt.MetricKey == "" {
				return errors.New("widget '" + wgt.Title + "' requires a metric")
			}
			if _, ok := models.MetricByKey(wgt.MetricKey); !ok {
				return errors.New("unknown metric: " + wgt.MetricKey)
			}
		}
		if wgt.ID == "" {
			wgt.ID = newID("w")
		}
	}
	return nil
}

func (h *TemplateHandler) Create(w http.ResponseWriter, r *http.Request) {
	var in templateInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if err := in.validate(); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	widgetsJSON, err := json.Marshal(in.Widgets)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	now := time.Now().UTC()
	t := models.Template{
		ID:          newID("tpl"),
		Name:        in.Name,
		Description: in.Description,
		Width:       in.Width,
		Height:      in.Height,
		Widgets:     in.Widgets,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	_, err = h.DB.Exec(`INSERT INTO templates (id, name, description, width, height, widgets_json, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		t.ID, t.Name, t.Description, t.Width, t.Height, string(widgetsJSON), t.CreatedAt.Format(time.RFC3339), t.UpdatedAt.Format(time.RFC3339))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, t)
}

func (h *TemplateHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var in templateInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if err := in.validate(); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	widgetsJSON, err := json.Marshal(in.Widgets)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	now := time.Now().UTC().Format(time.RFC3339)
	res, err := h.DB.Exec(`UPDATE templates SET name = ?, description = ?, width = ?, height = ?, widgets_json = ?, updated_at = ? WHERE id = ?`,
		in.Name, in.Description, in.Width, in.Height, string(widgetsJSON), now, id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if rows, _ := res.RowsAffected(); rows == 0 {
		writeError(w, http.StatusNotFound, "template not found")
		return
	}

	row := h.DB.QueryRow(`SELECT id, name, description, width, height, widgets_json, created_at, updated_at FROM templates WHERE id = ?`, id)
	t, err := scanTemplate(row)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, t)
}

func (h *TemplateHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var inUse int
	if err := h.DB.QueryRow(`SELECT COUNT(*) FROM dashboards WHERE template_id = ?`, id).Scan(&inUse); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if inUse > 0 {
		writeError(w, http.StatusConflict, "template is used by one or more dashboards")
		return
	}

	res, err := h.DB.Exec(`DELETE FROM templates WHERE id = ?`, id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if rows, _ := res.RowsAffected(); rows == 0 {
		writeError(w, http.StatusNotFound, "template not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanTemplate(row rowScanner) (models.Template, error) {
	var t models.Template
	var widgetsJSON, createdAt, updatedAt string
	if err := row.Scan(&t.ID, &t.Name, &t.Description, &t.Width, &t.Height, &widgetsJSON, &createdAt, &updatedAt); err != nil {
		return t, err
	}
	if err := json.Unmarshal([]byte(widgetsJSON), &t.Widgets); err != nil {
		return t, err
	}
	t.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	t.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
	return t, nil
}
