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

type DashboardHandler struct {
	DB *sql.DB
}

func NewDashboardHandler(db *sql.DB) *DashboardHandler {
	return &DashboardHandler{DB: db}
}

type dashboardSummary struct {
	models.Dashboard
	PageCount int `json:"pageCount"`
}

// List returns every dashboard with summary info for the dashboard list page.
func (h *DashboardHandler) List(w http.ResponseWriter, r *http.Request) {
	rows, err := h.DB.Query(`
		SELECT d.id, d.name, d.template_id, t.name, d.factory_id, f.name,
		       d.created_at, d.updated_at,
		       (SELECT COUNT(*) FROM dashboard_pages p WHERE p.dashboard_id = d.id)
		FROM dashboards d
		JOIN templates t ON t.id = d.template_id
		JOIN factories f ON f.id = d.factory_id
		ORDER BY d.created_at DESC`)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	out := []dashboardSummary{}
	for rows.Next() {
		var d dashboardSummary
		var createdAt, updatedAt string
		var tplName, facName string
		if err := rows.Scan(&d.ID, &d.Name, &d.TemplateID, &tplName, &d.FactoryID, &facName, &createdAt, &updatedAt, &d.PageCount); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		d.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		d.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
		d.Template = &models.Template{ID: d.TemplateID, Name: tplName}
		d.Factory = &models.Factory{ID: d.FactoryID, Name: facName}
		out = append(out, d)
	}
	writeJSON(w, http.StatusOK, out)
}

// Get returns a full dashboard: template (with widgets), factory and the
// ordered list of pages, each bound to one machine.
func (h *DashboardHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var d models.Dashboard
	var createdAt, updatedAt string
	err := h.DB.QueryRow(`SELECT id, name, template_id, factory_id, created_at, updated_at FROM dashboards WHERE id = ?`, id).
		Scan(&d.ID, &d.Name, &d.TemplateID, &d.FactoryID, &createdAt, &updatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		writeError(w, http.StatusNotFound, "dashboard not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	d.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	d.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)

	tplRow := h.DB.QueryRow(`SELECT id, name, description, width, height, widgets_json, created_at, updated_at FROM templates WHERE id = ?`, d.TemplateID)
	tpl, err := scanTemplate(tplRow)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "load template: "+err.Error())
		return
	}
	d.Template = &tpl

	var fac models.Factory
	if err := h.DB.QueryRow(`SELECT id, name, location FROM factories WHERE id = ?`, d.FactoryID).Scan(&fac.ID, &fac.Name, &fac.Location); err != nil {
		writeError(w, http.StatusInternalServerError, "load factory: "+err.Error())
		return
	}
	d.Factory = &fac

	rows, err := h.DB.Query(`
		SELECT p.id, p.machine_id, p.position, m.id, m.factory_id, m.name, m.type, m.status
		FROM dashboard_pages p
		JOIN machines m ON m.id = p.machine_id
		WHERE p.dashboard_id = ?
		ORDER BY p.position ASC`, id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	pages := []models.DashboardPage{}
	for rows.Next() {
		var p models.DashboardPage
		var m models.Machine
		if err := rows.Scan(&p.ID, &p.MachineID, &p.Position, &m.ID, &m.FactoryID, &m.Name, &m.Type, &m.Status); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		p.Machine = &m
		pages = append(pages, p)
	}
	d.Pages = pages

	writeJSON(w, http.StatusOK, d)
}

type createDashboardInput struct {
	Name       string   `json:"name"`
	TemplateID string   `json:"templateId"`
	FactoryID  string   `json:"factoryId"`
	MachineIDs []string `json:"machineIds"`
}

// Create generates a dashboard: applies a template to a factory with one
// page per selected machine (one machine per page, in the chosen order).
func (h *DashboardHandler) Create(w http.ResponseWriter, r *http.Request) {
	var in createDashboardInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	in.Name = strings.TrimSpace(in.Name)
	if in.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}
	if in.TemplateID == "" {
		writeError(w, http.StatusBadRequest, "templateId is required")
		return
	}
	if in.FactoryID == "" {
		writeError(w, http.StatusBadRequest, "factoryId is required")
		return
	}
	if len(in.MachineIDs) == 0 {
		writeError(w, http.StatusBadRequest, "select at least one machine")
		return
	}

	var tplExists int
	if err := h.DB.QueryRow(`SELECT COUNT(*) FROM templates WHERE id = ?`, in.TemplateID).Scan(&tplExists); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if tplExists == 0 {
		writeError(w, http.StatusBadRequest, "template not found")
		return
	}

	var facExists int
	if err := h.DB.QueryRow(`SELECT COUNT(*) FROM factories WHERE id = ?`, in.FactoryID).Scan(&facExists); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if facExists == 0 {
		writeError(w, http.StatusBadRequest, "factory not found")
		return
	}

	for _, mid := range in.MachineIDs {
		var belongs int
		if err := h.DB.QueryRow(`SELECT COUNT(*) FROM machines WHERE id = ? AND factory_id = ?`, mid, in.FactoryID).Scan(&belongs); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if belongs == 0 {
			writeError(w, http.StatusBadRequest, "machine "+mid+" does not belong to the selected factory")
			return
		}
	}

	tx, err := h.DB.Begin()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer tx.Rollback()

	now := time.Now().UTC().Format(time.RFC3339)
	id := newID("dash")
	if _, err := tx.Exec(`INSERT INTO dashboards (id, name, template_id, factory_id, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)`,
		id, in.Name, in.TemplateID, in.FactoryID, now, now); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	for i, mid := range in.MachineIDs {
		if _, err := tx.Exec(`INSERT INTO dashboard_pages (id, dashboard_id, machine_id, position) VALUES (?, ?, ?, ?)`,
			newID("page"), id, mid, i); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}

	if err := tx.Commit(); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Location", "/api/dashboards/"+id)
	writeJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *DashboardHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	res, err := h.DB.Exec(`DELETE FROM dashboards WHERE id = ?`, id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if rows, _ := res.RowsAffected(); rows == 0 {
		writeError(w, http.StatusNotFound, "dashboard not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
