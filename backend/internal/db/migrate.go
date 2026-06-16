package db

import "database/sql"

var schema = []string{
	`CREATE TABLE IF NOT EXISTS factories (
		id       TEXT PRIMARY KEY,
		name     TEXT NOT NULL,
		location TEXT NOT NULL DEFAULT ''
	)`,
	`CREATE TABLE IF NOT EXISTS machines (
		id         TEXT PRIMARY KEY,
		factory_id TEXT NOT NULL REFERENCES factories(id) ON DELETE CASCADE,
		name       TEXT NOT NULL,
		type       TEXT NOT NULL DEFAULT '',
		status     TEXT NOT NULL DEFAULT 'running'
	)`,
	`CREATE INDEX IF NOT EXISTS idx_machines_factory ON machines(factory_id)`,
	`CREATE TABLE IF NOT EXISTS templates (
		id           TEXT PRIMARY KEY,
		name         TEXT NOT NULL,
		description  TEXT NOT NULL DEFAULT '',
		width        INTEGER NOT NULL DEFAULT 1920,
		height       INTEGER NOT NULL DEFAULT 1080,
		widgets_json TEXT NOT NULL DEFAULT '[]',
		created_at   TEXT NOT NULL,
		updated_at   TEXT NOT NULL
	)`,
	`CREATE TABLE IF NOT EXISTS dashboards (
		id          TEXT PRIMARY KEY,
		name        TEXT NOT NULL,
		template_id TEXT NOT NULL REFERENCES templates(id) ON DELETE RESTRICT,
		factory_id  TEXT NOT NULL REFERENCES factories(id) ON DELETE RESTRICT,
		created_at  TEXT NOT NULL,
		updated_at  TEXT NOT NULL
	)`,
	`CREATE TABLE IF NOT EXISTS dashboard_pages (
		id           TEXT PRIMARY KEY,
		dashboard_id TEXT NOT NULL REFERENCES dashboards(id) ON DELETE CASCADE,
		machine_id   TEXT NOT NULL REFERENCES machines(id) ON DELETE CASCADE,
		position     INTEGER NOT NULL DEFAULT 0
	)`,
	`CREATE INDEX IF NOT EXISTS idx_pages_dashboard ON dashboard_pages(dashboard_id)`,
}

func migrate(conn *sql.DB) error {
	for _, stmt := range schema {
		if _, err := conn.Exec(stmt); err != nil {
			return err
		}
	}

	// Backfill columns added after the templates table originally existed.
	for _, a := range []struct{ column, ddl string }{
		{"width", "width INTEGER NOT NULL DEFAULT 1920"},
		{"height", "height INTEGER NOT NULL DEFAULT 1080"},
	} {
		if err := addColumnIfNotExists(conn, "templates", a.column, a.ddl); err != nil {
			return err
		}
	}

	return nil
}

// columnExists reports whether the given table has a column with the given
// name, using PRAGMA table_info (SQLite has no "ADD COLUMN IF NOT EXISTS").
func columnExists(conn *sql.DB, table, column string) (bool, error) {
	rows, err := conn.Query(`PRAGMA table_info(` + table + `)`)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	for rows.Next() {
		var cid, notNull, pk int
		var name, ctype string
		var dfltValue any
		if err := rows.Scan(&cid, &name, &ctype, &notNull, &dfltValue, &pk); err != nil {
			return false, err
		}
		if name == column {
			return true, nil
		}
	}
	return false, rows.Err()
}

// addColumnIfNotExists runs an ALTER TABLE ... ADD COLUMN statement only if
// the column doesn't already exist on the table.
func addColumnIfNotExists(conn *sql.DB, table, column, ddl string) error {
	exists, err := columnExists(conn, table, column)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	_, err = conn.Exec(`ALTER TABLE ` + table + ` ADD COLUMN ` + ddl)
	return err
}
