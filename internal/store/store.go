// Package store implements SQLite persistence for resume data and presets.
package store

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"

	"resumator/internal/models"
)

// Store wraps the SQLite database.
type Store struct {
	db *sql.DB
}

// Open creates or opens the database under the user config directory and
// applies migrations and default seeding.
func Open() (*Store, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return nil, fmt.Errorf("resolve config dir: %w", err)
	}
	dataDir := filepath.Join(configDir, "resumator")
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return nil, fmt.Errorf("create data dir: %w", err)
	}
	dsn := "file:" + filepath.Join(dataDir, "resumator.db") +
		"?_pragma=journal_mode(WAL)&_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	db.SetMaxOpenConns(1)
	s := &Store{db: db}
	if err := s.migrate(); err != nil {
		db.Close()
		return nil, err
	}
	if err := s.seed(); err != nil {
		db.Close()
		return nil, err
	}
	return s, nil
}

// Close releases the database handle.
func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) migrate() error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS profile (
			id INTEGER PRIMARY KEY CHECK (id = 1),
			name TEXT NOT NULL DEFAULT '',
			headline TEXT NOT NULL DEFAULT ''
		)`,
		`CREATE TABLE IF NOT EXISTS sections (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			type TEXT NOT NULL,
			title TEXT NOT NULL,
			title_color TEXT NOT NULL DEFAULT '',
			line_color TEXT NOT NULL DEFAULT '',
			sort_order INTEGER NOT NULL DEFAULT 0
		)`,
		`CREATE TABLE IF NOT EXISTS items (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			section_id INTEGER NOT NULL REFERENCES sections(id) ON DELETE CASCADE,
			sort_order INTEGER NOT NULL DEFAULT 0,
			data TEXT NOT NULL DEFAULT '{}'
		)`,
		`CREATE TABLE IF NOT EXISTS item_details (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			item_id INTEGER NOT NULL REFERENCES items(id) ON DELETE CASCADE,
			sort_order INTEGER NOT NULL DEFAULT 0,
			text TEXT NOT NULL DEFAULT ''
		)`,
		`CREATE TABLE IF NOT EXISTS presets (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			created_at TEXT NOT NULL DEFAULT (datetime('now'))
		)`,
		`CREATE TABLE IF NOT EXISTS preset_sections (
			preset_id INTEGER NOT NULL REFERENCES presets(id) ON DELETE CASCADE,
			section_id INTEGER NOT NULL REFERENCES sections(id) ON DELETE CASCADE,
			visible INTEGER NOT NULL DEFAULT 1,
			sort_order INTEGER NOT NULL DEFAULT 0,
			title_override TEXT,
			PRIMARY KEY (preset_id, section_id)
		)`,
		`CREATE TABLE IF NOT EXISTS preset_selections (
			preset_id INTEGER NOT NULL REFERENCES presets(id) ON DELETE CASCADE,
			item_id INTEGER REFERENCES items(id) ON DELETE CASCADE,
			detail_id INTEGER REFERENCES item_details(id) ON DELETE CASCADE,
			included INTEGER NOT NULL DEFAULT 0,
			PRIMARY KEY (preset_id, item_id, detail_id)
		)`,
		`CREATE TABLE IF NOT EXISTS text_styles (
			id INTEGER PRIMARY KEY CHECK (id = 1),
			section_title_size REAL NOT NULL DEFAULT 10.5,
			section_title_bold INTEGER NOT NULL DEFAULT 1,
			section_title_italic INTEGER NOT NULL DEFAULT 0,
			entry_title_size REAL NOT NULL DEFAULT 10,
			entry_title_bold INTEGER NOT NULL DEFAULT 1,
			entry_title_italic INTEGER NOT NULL DEFAULT 0,
			entry_dates_size REAL NOT NULL DEFAULT 9,
			entry_dates_bold INTEGER NOT NULL DEFAULT 0,
			entry_dates_italic INTEGER NOT NULL DEFAULT 0,
			body_size REAL NOT NULL DEFAULT 9.5,
			body_bold INTEGER NOT NULL DEFAULT 0,
			body_italic INTEGER NOT NULL DEFAULT 0
		)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_preset_selections_refs
			ON preset_selections (preset_id, ifnull(item_id, 0), ifnull(detail_id, 0))`,
	}
	for _, stmt := range stmts {
		if _, err := s.db.Exec(stmt); err != nil {
			return fmt.Errorf("migrate: %w", err)
		}
	}

	// Add columns introduced after the first release to existing
	// databases (CREATE TABLE IF NOT EXISTS does not touch them).
	for _, col := range []string{"title_color", "line_color"} {
		present, err := s.columnExists("sections", col)
		if err != nil {
			return err
		}
		if !present {
			if _, err := s.db.Exec(
				`ALTER TABLE sections ADD COLUMN ` + col + ` TEXT NOT NULL DEFAULT ''`,
			); err != nil {
				return fmt.Errorf("migrate: add %s: %w", col, err)
			}
		}
	}
	return nil
}

func (s *Store) columnExists(table, column string) (bool, error) {
	rows, err := s.db.Query(`PRAGMA table_info(` + table + `)`)
	if err != nil {
		return false, err
	}
	defer rows.Close()
	for rows.Next() {
		var cid int
		var name, colType string
		var notNull int
		var dflt sql.NullString
		var pk int
		if err := rows.Scan(&cid, &name, &colType, &notNull, &dflt, &pk); err != nil {
			return false, err
		}
		if name == column {
			return true, nil
		}
	}
	return false, rows.Err()
}

func (s *Store) seed() error {
	var count int
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM sections`).Scan(&count); err != nil {
		return err
	}
	if count == 0 {
		defaults := []struct {
			sectionType models.SectionType
			title       string
		}{
			{models.SectionContact, "Contact"},
			{models.SectionSkills, "Skills"},
			{models.SectionExperience, "Professional Experience"},
			{models.SectionEducation, "Education"},
		}
		for i, d := range defaults {
			if _, err := s.db.Exec(
				`INSERT INTO sections (type, title, sort_order) VALUES (?, ?, ?)`,
				string(d.sectionType), d.title, i,
			); err != nil {
				return err
			}
		}
	}

	if _, err := s.db.Exec(
		`INSERT INTO profile (id, name, headline) VALUES (1, '', '')
		 ON CONFLICT(id) DO NOTHING`,
	); err != nil {
		return err
	}

	var presetCount int
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM presets`).Scan(&presetCount); err != nil {
		return err
	}
	if presetCount == 0 {
		if _, err := s.db.Exec(`INSERT INTO presets (name) VALUES ('Default')`); err != nil {
			return err
		}
	}
	return nil
}

// GetProfile returns the global profile.
func (s *Store) GetProfile() (models.Profile, error) {
	var p models.Profile
	err := s.db.QueryRow(`SELECT name, headline FROM profile WHERE id = 1`).
		Scan(&p.Name, &p.Headline)
	if err != nil {
		return p, fmt.Errorf("get profile: %w", err)
	}
	return p, nil
}

// UpdateProfile saves the global profile.
func (s *Store) UpdateProfile(p models.Profile) error {
	_, err := s.db.Exec(`UPDATE profile SET name = ?, headline = ? WHERE id = 1`,
		p.Name, p.Headline)
	if err != nil {
		return fmt.Errorf("update profile: %w", err)
	}
	return nil
}

func decodeItemData(raw string) map[string]interface{} {
	data := map[string]interface{}{}
	if raw != "" {
		// Stored data is always written by this app; ignore malformed rows.
		_ = json.Unmarshal([]byte(raw), &data)
	}
	return data
}
