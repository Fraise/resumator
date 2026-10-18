package store

import (
	"fmt"

	"resumator/internal/models"
)

// AddSection creates a section and returns its id.
func (s *Store) AddSection(sectionType, title string) (int64, error) {
	if !isValidSectionType(sectionType) {
		return 0, fmt.Errorf("invalid section type: %s", sectionType)
	}
	var maxOrder int
	if err := s.db.QueryRow(`SELECT COALESCE(MAX(sort_order), -1) FROM sections`).Scan(&maxOrder); err != nil {
		return 0, err
	}
	res, err := s.db.Exec(
		`INSERT INTO sections (type, title, sort_order) VALUES (?, ?, ?)`,
		sectionType, title, maxOrder+1,
	)
	if err != nil {
		return 0, fmt.Errorf("add section: %w", err)
	}
	return res.LastInsertId()
}

// UpdateSection renames a section.
func (s *Store) UpdateSection(id int64, title string) error {
	res, err := s.db.Exec(`UPDATE sections SET title = ? WHERE id = ?`, title, id)
	if err != nil {
		return fmt.Errorf("update section: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return fmt.Errorf("section %d not found", id)
	}
	return nil
}

// DeleteSection removes a section and all its items.
func (s *Store) DeleteSection(id int64) error {
	if _, err := s.db.Exec(`DELETE FROM sections WHERE id = ?`, id); err != nil {
		return fmt.Errorf("delete section: %w", err)
	}
	return nil
}

// ReorderSections stores the given order for all sections.
func (s *Store) ReorderSections(orderedIDs []int64) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for i, id := range orderedIDs {
		if _, err := tx.Exec(`UPDATE sections SET sort_order = ? WHERE id = ?`, i, id); err != nil {
			return fmt.Errorf("reorder sections: %w", err)
		}
	}
	return tx.Commit()
}

// SetSectionColors stores the optional title and separator colors for a
// section. Empty strings restore the defaults.
func (s *Store) SetSectionColors(id int64, titleColor, lineColor string) error {
	res, err := s.db.Exec(
		`UPDATE sections SET title_color = ?, line_color = ? WHERE id = ?`,
		titleColor, lineColor, id,
	)
	if err != nil {
		return fmt.Errorf("set section colors: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return fmt.Errorf("section %d not found", id)
	}
	return nil
}

// GetResumeState resolves the full state for the given preset. When
// presetID is 0, the oldest preset (Default) is used.
func (s *Store) GetResumeState(presetID int64) (models.ResumeState, error) {
	if presetID == 0 {
		if err := s.db.QueryRow(`SELECT MIN(id) FROM presets`).Scan(&presetID); err != nil {
			return models.ResumeState{}, fmt.Errorf("no presets: %w", err)
		}
	}
	state := models.ResumeState{}
	profile, err := s.GetProfile()
	if err != nil {
		return state, err
	}
	state.Profile = profile
	state.Presets, err = s.ListPresets()
	if err != nil {
		return state, err
	}
	if err := s.validatePreset(presetID); err != nil {
		return state, err
	}
	state.Sections = []models.Section{}

	rows, err := s.db.Query(`
		SELECT s.id, s.type, s.title, s.title_color, s.line_color,
		       COALESCE(ps.visible, 1) AS visible
		FROM sections s
		LEFT JOIN preset_sections ps ON ps.section_id = s.id AND ps.preset_id = ?
		ORDER BY s.sort_order, s.id`, presetID)
	if err != nil {
		return state, fmt.Errorf("load sections: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var sec models.Section
		var visible int
		if err := rows.Scan(
			&sec.ID, &sec.Type, &sec.Title, &sec.TitleColor, &sec.LineColor, &visible,
		); err != nil {
			return state, err
		}
		sec.Visible = visible == 1
		state.Sections = append(state.Sections, sec)
	}
	if err := rows.Err(); err != nil {
		return state, err
	}

	for i := range state.Sections {
		items, err := s.loadItems(presetID, state.Sections[i].ID)
		if err != nil {
			return state, err
		}
		state.Sections[i].Items = items
	}
	return state, nil
}

func (s *Store) loadItems(presetID, sectionID int64) ([]models.Item, error) {
	rows, err := s.db.Query(`
		SELECT i.id,
		       NOT EXISTS (
		           SELECT 1 FROM preset_selections ps
		           WHERE ps.preset_id = ? AND ps.item_id = i.id
		             AND ps.detail_id IS NULL AND ps.included = 0
		       ) AS included
		FROM items i
		WHERE i.section_id = ?
		ORDER BY i.sort_order, i.id`, presetID, sectionID)
	if err != nil {
		return nil, fmt.Errorf("load items: %w", err)
	}
	defer rows.Close()

	items := []models.Item{}
	for rows.Next() {
		var it models.Item
		var included int
		if err := rows.Scan(&it.ID, &included); err != nil {
			return nil, err
		}
		it.SectionID = sectionID
		it.Included = included == 1
		items = append(items, it)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	for i := range items {
		var rawData string
		err := s.db.QueryRow(`SELECT data FROM items WHERE id = ?`, items[i].ID).Scan(&rawData)
		if err != nil {
			return nil, err
		}
		items[i].Data = decodeItemData(rawData)
		details, err := s.loadDetails(presetID, items[i].ID)
		if err != nil {
			return nil, err
		}
		items[i].Details = details
	}
	return items, nil
}

func (s *Store) loadDetails(presetID, itemID int64) ([]models.Detail, error) {
	rows, err := s.db.Query(`
		SELECT d.id, d.text,
		       NOT EXISTS (
		           SELECT 1 FROM preset_selections ps
		           WHERE ps.preset_id = ? AND ps.detail_id = d.id AND ps.included = 0
		       ) AS included
		FROM item_details d
		WHERE d.item_id = ?
		ORDER BY d.sort_order, d.id`, presetID, itemID)
	if err != nil {
		return nil, fmt.Errorf("load details: %w", err)
	}
	defer rows.Close()

	details := []models.Detail{}
	for rows.Next() {
		var d models.Detail
		var included int
		if err := rows.Scan(&d.ID, &d.Text, &included); err != nil {
			return nil, err
		}
		d.ItemID = itemID
		d.Included = included == 1
		details = append(details, d)
	}
	return details, rows.Err()
}

func isValidSectionType(sectionType string) bool {
	switch models.SectionType(sectionType) {
	case models.SectionContact, models.SectionSkills, models.SectionExperience,
		models.SectionEducation, models.SectionCustom:
		return true
	}
	return false
}
