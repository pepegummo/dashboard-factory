# Factory Dashboard Builder

Customizable monitoring dashboards for factories: design a widget
template, pick a factory, add machines (one machine per page), and
generate a live monitoring dashboard.

## Flow

1. **Create a template** – a reusable layout of widgets (status, KPI,
   gauge, line chart, readings table) bound to metrics.
2. **Choose a factory** – machines are scoped to the selected factory.
3. **Add machines** – each machine becomes its own dashboard page.
4. **Generate the dashboard** – view live (simulated) telemetry per
   machine, rendered using the chosen template's widgets.

## Tech stack

- **Backend**: Go, [chi](https://github.com/go-chi/chi) router, SQLite
  (via `modernc.org/sqlite`, no CGO required)
- **Frontend**: Vue 3 (Composition API, TypeScript), Vite, Pinia,
  vue-router, Bootstrap 5 + Bootstrap Icons, Chart.js

## Project layout

```
dashboard-builder/
├── backend/    # Go API server (REST, SQLite-backed)
└── frontend/   # Vue 3 + Vite + Bootstrap SPA
```

## Running the backend

Requires Go 1.26+.

```bash
cd backend
go run ./cmd/server
```

The server listens on `:8080` by default and stores its SQLite
database at `./data/dashboard-builder.db` (created automatically, with
seed data on first run: 3 factories, 11 machines, and a starter
"Standard Machine Monitor" template).

Environment variables:

| Variable  | Default                       | Description                |
|-----------|--------------------------------|-----------------------------|
| `PORT`    | `8080`                          | HTTP listen port            |
| `DB_PATH` | `./data/dashboard-builder.db`   | Path to the SQLite database |

### API overview

| Method | Path                              | Description                       |
|--------|-----------------------------------|------------------------------------|
| GET    | `/api/health`                     | Health check                       |
| GET    | `/api/factories`                  | List factories                     |
| GET    | `/api/factories/{id}/machines`    | List machines in a factory         |
| GET    | `/api/metrics`                    | List available telemetry metrics   |
| GET    | `/api/templates`                  | List dashboard templates           |
| POST   | `/api/templates`                  | Create a template                  |
| GET    | `/api/templates/{id}`             | Get a template                     |
| PUT    | `/api/templates/{id}`             | Update a template                  |
| DELETE | `/api/templates/{id}`             | Delete a template                  |
| GET    | `/api/dashboards`                 | List dashboards                    |
| POST   | `/api/dashboards`                 | Create a dashboard                 |
| GET    | `/api/dashboards/{id}`            | Get a dashboard (with pages)       |
| DELETE | `/api/dashboards/{id}`            | Delete a dashboard                 |
| GET    | `/api/machines/{id}/telemetry`    | Get simulated live telemetry       |

## Running the frontend

Requires Node.js 20+.

```bash
cd frontend
npm install
npm run dev
```

The dev server runs on `http://localhost:5174` and proxies `/api`
requests to the backend at `http://localhost:8080` (configured in
`vite.config.ts`). Start the backend first.

To build for production:

```bash
npm run build
```

## Running with Docker Compose

Requires Docker (with Compose v2).

```bash
docker compose up --build
```

This builds and starts both services:

- **backend** – Go API on `http://localhost:8080`, with its SQLite
  database persisted in the `backend-data` named volume (`/data`
  inside the container).
- **frontend** – static build served by nginx on
  `http://localhost:5174`, which proxies `/api/*` requests to the
  `backend` service.

Open `http://localhost:5174` in a browser. Stop with `docker compose down`
(add `-v` to also delete the persisted database volume).

## Notes

- Telemetry is simulated server-side (sine-wave signals per machine
  for temperature, pressure, speed, vibration, and a monotonically
  increasing output counter) since no real IoT data source was
  specified. The frontend polls `/api/machines/{id}/telemetry` every
  2 seconds for the active dashboard page.
- Widget types: `status`, `kpi`, `gauge`, `line`, `table`. Each widget
  (except `table`) binds to one metric from the fixed metric catalog.
