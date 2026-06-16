package handlers

import (
	"database/sql"
	"net/http"

	"dashboard-builder/backend/internal/models"

	"github.com/go-chi/chi/v5"
)

type FactoryHandler struct {
	DB *sql.DB
}

func NewFactoryHandler(db *sql.DB) *FactoryHandler {
	return &FactoryHandler{DB: db}
}

// List returns every factory.
func (h *FactoryHandler) List(w http.ResponseWriter, r *http.Request) {
	rows, err := h.DB.Query(`SELECT id, name, location FROM factories ORDER BY name`)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	factories := []models.Factory{}
	for rows.Next() {
		var f models.Factory
		if err := rows.Scan(&f.ID, &f.Name, &f.Location); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		factories = append(factories, f)
	}
	writeJSON(w, http.StatusOK, factories)
}

// ListMachines returns every machine that belongs to the :id factory.
func (h *FactoryHandler) ListMachines(w http.ResponseWriter, r *http.Request) {
	factoryID := chi.URLParam(r, "id")

	rows, err := h.DB.Query(`SELECT id, factory_id, name, type, status FROM machines WHERE factory_id = ? ORDER BY name`, factoryID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	machines := []models.Machine{}
	for rows.Next() {
		var m models.Machine
		if err := rows.Scan(&m.ID, &m.FactoryID, &m.Name, &m.Type, &m.Status); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		machines = append(machines, m)
	}
	writeJSON(w, http.StatusOK, machines)
}
