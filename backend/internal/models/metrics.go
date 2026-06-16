package models

// Metrics is the fixed catalog of telemetry fields every machine reports.
// Templates bind widgets to one of these keys; the telemetry endpoint
// generates simulated readings for the same keys.
var Metrics = []MetricDef{
	{Key: "temperature", Label: "Temperature", Unit: "°C", Min: 20, Max: 95},
	{Key: "pressure", Label: "Pressure", Unit: "bar", Min: 0, Max: 10},
	{Key: "speed", Label: "Speed", Unit: "rpm", Min: 0, Max: 3000},
	{Key: "vibration", Label: "Vibration", Unit: "mm/s", Min: 0, Max: 2},
	{Key: "output", Label: "Output Count", Unit: "pcs", Min: 0, Max: 10000},
	{Key: "status", Label: "Machine Status", Unit: "", Min: 0, Max: 0},
}

// MetricByKey looks up a metric definition by its key.
func MetricByKey(key string) (MetricDef, bool) {
	for _, m := range Metrics {
		if m.Key == key {
			return m, true
		}
	}
	return MetricDef{}, false
}
