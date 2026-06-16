package db

import (
	"database/sql"
	"encoding/json"
	"time"

	"dashboard-builder/backend/internal/models"
)

// seed populates the database with demo factories, machines and a starter
// template the first time the app runs (i.e. when the factories table is
// empty). It is a no-op on subsequent starts.
func seed(conn *sql.DB) error {
	var count int
	if err := conn.QueryRow(`SELECT COUNT(*) FROM factories`).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	tx, err := conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	factories := []models.Factory{
		{ID: "f-bangkok", Name: "Bangkok Plant", Location: "Bangkok, Thailand"},
		{ID: "f-rayong", Name: "Rayong Plant", Location: "Rayong, Thailand"},
		{ID: "f-chonburi", Name: "Chonburi Plant", Location: "Chonburi, Thailand"},
	}
	for _, f := range factories {
		if _, err := tx.Exec(`INSERT INTO factories (id, name, location) VALUES (?, ?, ?)`, f.ID, f.Name, f.Location); err != nil {
			return err
		}
	}

	machines := []models.Machine{
		{ID: "m-bkk-cnc-01", FactoryID: "f-bangkok", Name: "CNC Machine 01", Type: "CNC", Status: "running"},
		{ID: "m-bkk-mold-02", FactoryID: "f-bangkok", Name: "Injection Molder 02", Type: "Injection Molding", Status: "running"},
		{ID: "m-bkk-conv-03", FactoryID: "f-bangkok", Name: "Conveyor Belt 03", Type: "Conveyor", Status: "idle"},
		{ID: "m-bkk-rob-04", FactoryID: "f-bangkok", Name: "Packaging Robot 04", Type: "Robot", Status: "running"},

		{ID: "m-ryg-ext-01", FactoryID: "f-rayong", Name: "Extruder 01", Type: "Extruder", Status: "running"},
		{ID: "m-ryg-press-02", FactoryID: "f-rayong", Name: "Press Machine 02", Type: "Press", Status: "stopped"},
		{ID: "m-ryg-weld-03", FactoryID: "f-rayong", Name: "Welding Robot 03", Type: "Robot", Status: "running"},

		{ID: "m-chb-mix-01", FactoryID: "f-chonburi", Name: "Mixer 01", Type: "Mixer", Status: "running"},
		{ID: "m-chb-fill-02", FactoryID: "f-chonburi", Name: "Filler 02", Type: "Filler", Status: "running"},
		{ID: "m-chb-label-03", FactoryID: "f-chonburi", Name: "Labeler 03", Type: "Labeler", Status: "idle"},
		{ID: "m-chb-pallet-04", FactoryID: "f-chonburi", Name: "Palletizer 04", Type: "Robot", Status: "running"},
	}
	for _, m := range machines {
		if _, err := tx.Exec(`INSERT INTO machines (id, factory_id, name, type, status) VALUES (?, ?, ?, ?, ?)`,
			m.ID, m.FactoryID, m.Name, m.Type, m.Status); err != nil {
			return err
		}
	}

	starter := models.Template{
		ID:          "tpl-standard",
		Name:        "Standard Machine Monitor",
		Description: "Status, key readings, speed gauge and a vibration trend for one machine.",
		Width:       1920,
		Height:      1080,
		Widgets: []models.Widget{
			{ID: "w1", Type: "status", Title: "Machine Status", MetricKey: "status", X: 0, Y: 0, W: 32, H: 22},
			{ID: "w2", Type: "kpi", Title: "Temperature", MetricKey: "temperature", X: 34, Y: 0, W: 32, H: 22},
			{ID: "w3", Type: "kpi", Title: "Pressure", MetricKey: "pressure", X: 68, Y: 0, W: 32, H: 22},
			{ID: "w4", Type: "gauge", Title: "Speed", MetricKey: "speed", X: 0, Y: 24, W: 40, H: 34},
			{ID: "w5", Type: "line", Title: "Vibration Trend", MetricKey: "vibration", X: 42, Y: 24, W: 58, H: 34},
			{ID: "w6", Type: "table", Title: "All Readings", MetricKey: "", X: 0, Y: 60, W: 100, H: 38},
		},
	}
	widgetsJSON, err := json.Marshal(starter.Widgets)
	if err != nil {
		return err
	}
	now := time.Now().UTC().Format(time.RFC3339)
	if _, err := tx.Exec(`INSERT INTO templates (id, name, description, width, height, widgets_json, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		starter.ID, starter.Name, starter.Description, starter.Width, starter.Height, string(widgetsJSON), now, now); err != nil {
		return err
	}

	return tx.Commit()
}
