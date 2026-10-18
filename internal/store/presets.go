package store

import (
	"fmt"
	"strings"

	"resumator/internal/models"
)

func (s *Store) validatePreset(presetID int64) error {
	var count int
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM presets WHERE id = ?`, presetID).Scan(&count); err != nil {
		return err
	}
	if count == 0 {
		return fmt.Errorf("preset %d not found", presetID)
	}
	return nil
}

// ListPresets returns all presets, oldest first.
func (s *Store) ListPresets() ([]models.Preset, error) {
	rows, err := s.db.Query(`SELECT id, name, created_at FROM presets ORDER BY id`)
	if err != nil {
		return nil, fmt.Errorf("list presets: %w", err)
	}
	defer rows.Close()
	presets := []models.Preset{}
	for rows.Next() {
		var p models.Preset
		if err := rows.Scan(&p.ID, &p.Name, &p.CreatedAt); err != nil {
			return nil, err
		}
		presets = append(presets, p)
	}
	return presets, rows.Err()
}

// CreatePreset adds a new preset with default selections (everything
// included and visible).
func (s *Store) CreatePreset(name string) (models.Preset, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return models.Preset{}, fmt.Errorf("preset name must not be empty")
	}
	res, err := s.db.Exec(`INSERT INTO presets (name) VALUES (?)`, name)
	if err != nil {
		return models.Preset{}, fmt.Errorf("create preset: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return models.Preset{}, err
	}
	var createdAt string
	if err := s.db.QueryRow(`SELECT created_at FROM presets WHERE id = ?`, id).Scan(&createdAt); err != nil {
		return models.Preset{}, err
	}
	return models.Preset{ID: id, Name: name, CreatedAt: createdAt}, nil
}

// DuplicatePreset creates a new preset copying the sections, items and
// detail selections of the source preset.
func (s *Store) DuplicatePreset(name string, sourceID int64) (models.Preset, error) {
	if err := s.validatePreset(sourceID); err != nil {
		return models.Preset{}, err
	}
	tx, err := s.db.Begin()
	if err != nil {
		return models.Preset{}, err
	}
	defer tx.Rollback()

	res, err := tx.Exec(`INSERT INTO presets (name) VALUES (?)`, strings.TrimSpace(name))
	if err != nil {
		return models.Preset{}, fmt.Errorf("duplicate preset: %w", err)
	}
	newID, err := res.LastInsertId()
	if err != nil {
		return models.Preset{}, err
	}
	if _, err := tx.Exec(`
		INSERT INTO preset_sections (preset_id, section_id, visible, sort_order, title_override)
		SELECT ?, section_id, visible, sort_order, title_override
		FROM preset_sections WHERE preset_id = ?`, newID, sourceID); err != nil {
		return models.Preset{}, err
	}
	if _, err := tx.Exec(`
		INSERT INTO preset_selections (preset_id, item_id, detail_id, included)
		SELECT ?, item_id, detail_id, included
		FROM preset_selections WHERE preset_id = ?`, newID, sourceID); err != nil {
		return models.Preset{}, err
	}
	if err := tx.Commit(); err != nil {
		return models.Preset{}, err
	}
	return s.GetPreset(newID)
}

// GetPreset returns a single preset.
func (s *Store) GetPreset(id int64) (models.Preset, error) {
	var p models.Preset
	err := s.db.QueryRow(`SELECT id, name, created_at FROM presets WHERE id = ?`, id).
		Scan(&p.ID, &p.Name, &p.CreatedAt)
	if err != nil {
		return p, fmt.Errorf("get preset: %w", err)
	}
	return p, nil
}

// RenamePreset changes a preset name.
func (s *Store) RenamePreset(id int64, name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("preset name must not be empty")
	}
	res, err := s.db.Exec(`UPDATE presets SET name = ? WHERE id = ?`, name, id)
	if err != nil {
		return fmt.Errorf("rename preset: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return fmt.Errorf("preset %d not found", id)
	}
	return nil
}

// DeletePreset removes a preset. The last remaining preset cannot be
// deleted.
func (s *Store) DeletePreset(id int64) error {
	var count int
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM presets`).Scan(&count); err != nil {
		return err
	}
	if count <= 1 {
		return fmt.Errorf("cannot delete the last preset")
	}
	if _, err := s.db.Exec(`DELETE FROM presets WHERE id = ?`, id); err != nil {
		return fmt.Errorf("delete preset: %w", err)
	}
	return nil
}

// ToggleItem records whether an item is included in the given preset.
func (s *Store) ToggleItem(presetID, itemID int64, included bool) error {
	if err := s.validatePreset(presetID); err != nil {
		return err
	}
	if included {
		if _, err := s.db.Exec(`
			DELETE FROM preset_selections
			WHERE preset_id = ? AND item_id = ? AND detail_id IS NULL`,
			presetID, itemID); err != nil {
			return fmt.Errorf("toggle item: %w", err)
		}
		return nil
	}
	if err := s.deleteSelection(presetID, "item_id = ? AND detail_id IS NULL", itemID); err != nil {
		return err
	}
	if _, err := s.db.Exec(`
		INSERT INTO preset_selections (preset_id, item_id, detail_id, included)
		VALUES (?, ?, NULL, 0)`, presetID, itemID); err != nil {
		return fmt.Errorf("toggle item: %w", err)
	}
	return nil
}

// deleteSelection removes any selection rows matching the given column
// condition for a preset.
func (s *Store) deleteSelection(presetID int64, cond string, arg int64) error {
	if _, err := s.db.Exec(
		`DELETE FROM preset_selections WHERE preset_id = ? AND `+cond, presetID, arg,
	); err != nil {
		return fmt.Errorf("toggle: %w", err)
	}
	return nil
}

// ToggleDetail records whether a detail is included in the given preset.
func (s *Store) ToggleDetail(presetID, detailID int64, included bool) error {
	if err := s.validatePreset(presetID); err != nil {
		return err
	}
	var itemID interface{}
	if err := s.db.QueryRow(
		`SELECT item_id FROM item_details WHERE id = ?`, detailID).Scan(&itemID); err != nil {
		return fmt.Errorf("detail %d not found", detailID)
	}
	if included {
		if _, err := s.db.Exec(`
			DELETE FROM preset_selections
			WHERE preset_id = ? AND detail_id = ?`, presetID, detailID); err != nil {
			return fmt.Errorf("toggle detail: %w", err)
		}
		return nil
	}
	if err := s.deleteSelection(presetID, "detail_id = ?", detailID); err != nil {
		return err
	}
	if _, err := s.db.Exec(`
		INSERT INTO preset_selections (preset_id, item_id, detail_id, included)
		VALUES (?, ?, ?, 0)`, presetID, itemID, detailID); err != nil {
		return fmt.Errorf("toggle detail: %w", err)
	}
	return nil
}

// SetSectionVisible records whether a section appears in the given preset.
func (s *Store) SetSectionVisible(presetID, sectionID int64, visible bool) error {
	if err := s.validatePreset(presetID); err != nil {
		return err
	}
	vis := 0
	if visible {
		vis = 1
	}
	if _, err := s.db.Exec(`
		INSERT INTO preset_sections (preset_id, section_id, visible, sort_order, title_override)
		VALUES (?, ?, ?, 0, NULL)
		ON CONFLICT(preset_id, section_id) DO UPDATE SET visible = ?`,
		presetID, sectionID, vis, vis); err != nil {
		return fmt.Errorf("set section visible: %w", err)
	}
	return nil
}
