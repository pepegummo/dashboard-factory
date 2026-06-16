package handlers

import (
	"database/sql"
	"errors"
	"hash/fnv"
	"math"
	"net/http"
	"time"

	"dashboard-builder/backend/internal/models"

	"github.com/go-chi/chi/v5"
)

// TelemetryHandler serves simulated live readings for machines. There is no
// real IoT data source in this project, so values are generated from smooth
// sine waves seeded per machine+metric, giving each widget realistic-looking
// movement when polled.
type TelemetryHandler struct {
	DB        *sql.DB
	StartedAt time.Time
}

func NewTelemetryHandler(db *sql.DB) *TelemetryHandler {
	return &TelemetryHandler{DB: db, StartedAt: time.Now()}
}

// Get returns the current simulated reading for every metric on a machine.
func (h *TelemetryHandler) Get(w http.ResponseWriter, r *http.Request) {
	machineID := chi.URLParam(r, "id")

	var status string
	err := h.DB.QueryRow(`SELECT status FROM machines WHERE id = ?`, machineID).Scan(&status)
	if errors.Is(err, sql.ErrNoRows) {
		writeError(w, http.StatusNotFound, "machine not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	now := time.Now()
	readings := map[string]any{}
	for _, m := range models.Metrics {
		switch m.Key {
		case "status":
			readings[m.Key] = status
		case "output":
			readings[m.Key] = simulateCounter(machineID, now, h.StartedAt)
		default:
			readings[m.Key] = simulateWave(machineID, m, now)
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"machineId": machineID,
		"timestamp": now.UTC().Format(time.RFC3339Nano),
		"readings":  readings,
	})
}

// phase derives a stable 0..2π offset from a machine+metric pair so every
// machine's chart oscillates out of sync with the others.
func phase(machineID, key string) float64 {
	hsh := fnv.New32a()
	_, _ = hsh.Write([]byte(machineID + ":" + key))
	return float64(hsh.Sum32()%1000) / 1000.0 * 2 * math.Pi
}

func simulateWave(machineID string, m models.MetricDef, now time.Time) float64 {
	mid := (m.Min + m.Max) / 2
	amp := (m.Max - m.Min) / 2 * 0.7
	const periodSeconds = 60.0

	t := float64(now.UnixMilli()) / 1000.0
	p := phase(machineID, m.Key)

	wave := math.Sin(2*math.Pi*t/periodSeconds + p)
	jitter := math.Sin(t*7+p*3) * amp * 0.08

	val := mid + amp*wave + jitter
	if val < m.Min {
		val = m.Min
	}
	if val > m.Max {
		val = m.Max
	}
	return math.Round(val*100) / 100
}

func simulateCounter(machineID string, now, startedAt time.Time) float64 {
	hsh := fnv.New32a()
	_, _ = hsh.Write([]byte(machineID + ":rate"))
	seed := hsh.Sum32()

	rate := 1 + float64(seed%20)  // units per second
	base := float64(seed % 5000)  // starting offset
	elapsed := now.Sub(startedAt).Seconds()

	return math.Round(base + elapsed*rate)
}
