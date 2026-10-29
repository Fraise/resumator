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

// UpdateContactHeader saves the name and headline of a contact section;
// they are shown at the top of the resume.
func (s *Store) UpdateContactHeader(id int64, name, headline string) error {
	res, err := s.db.Exec(
		`UPDATE sections SET name = ?, headline = ? WHERE id = ? AND type = ?`,
		name, headline, id, models.SectionContact,
	)
	if err != nil {
		return fmt.Errorf("update contact header: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return fmt.Errorf("contact section %d not found", id)
	}
	return nil
}

// DeleteSection removes a section and all its items.
func (s *Store) DeleteSection(id int64) error {
	res, err := s.db.Exec(`DELETE FROM sections WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete section: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return fmt.Errorf("section %d not found", id)
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
	var err error
	state.Styles, err = s.GetTextStyles()
	if err != nil {
		return state, err
	}
	state.Presets, err = s.ListPresets()
	if err != nil {
		return state, err
	}
	if err := s.validatePreset(presetID); err != nil {
		return state, err
	}
	state.Sections = []models.Section{}

	rows, err := s.db.Query(`
		SELECT s.id, s.type, COALESCE(ps.title_override, s.title),
		       s.name, s.headline, s.title_color, s.line_color,
		       COALESCE(ps.visible, 1) AS visible
		FROM sections s
		LEFT JOIN preset_sections ps ON ps.section_id = s.id AND ps.preset_id = ?
		ORDER BY s.sort_order, s.id`, presetID)
	if err != nil {
		return state, fmt.Errorf("load sections: %w", err)
	}

	// sectionID -> index in state.Sections
	sectionIdx := map[int64]int{}
	for rows.Next() {
		var sec models.Section
		var visible int
		if err := rows.Scan(
			&sec.ID, &sec.Type, &sec.Title, &sec.Name, &sec.Headline,
			&sec.TitleColor, &sec.LineColor, &visible,
		); err != nil {
			rows.Close()
			return state, err
		}
		sec.Visible = visible == 1
		sec.Items = []models.Item{}
		sectionIdx[sec.ID] = len(state.Sections)
		state.Sections = append(state.Sections, sec)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return state, err
	}
	rows.Close()

	// Load every item of the preset in one query, ordered like its section,
	// and attach each item to its section as it arrives.
	itemRows, err := s.db.Query(`
		SELECT i.id, i.section_id, i.data,
		       NOT EXISTS (
		           SELECT 1 FROM preset_selections ps
		           WHERE ps.preset_id = ? AND ps.item_id = i.id
		             AND ps.detail_id IS NULL AND ps.included = 0
		       ) AS included
		FROM items i
		JOIN sections s ON s.id = i.section_id
		ORDER BY s.sort_order, s.id, i.sort_order, i.id`, presetID)
	if err != nil {
		return state, fmt.Errorf("load items: %w", err)
	}

	// itemID -> position of the item inside its section
	type itemPos struct{ sec, idx int }
	itemIdx := map[int64]itemPos{}
	for itemRows.Next() {
		var it models.Item
		var included int
		var rawData string
		if err := itemRows.Scan(&it.ID, &it.SectionID, &rawData, &included); err != nil {
			itemRows.Close()
			return state, err
		}
		si, ok := sectionIdx[it.SectionID]
		if !ok {
			continue
		}
		it.Included = included == 1
		it.Data = decodeItemData(rawData)
		it.Details = []models.Detail{}
		items := &state.Sections[si].Items
		itemIdx[it.ID] = itemPos{si, len(*items)}
		*items = append(*items, it)
	}
	if err := itemRows.Err(); err != nil {
		itemRows.Close()
		return state, err
	}
	itemRows.Close()

	// Load every detail of the preset in one query, ordered like its item.
	detailRows, err := s.db.Query(`
		SELECT d.id, d.item_id, d.text,
		       NOT EXISTS (
		           SELECT 1 FROM preset_selections ps
		           WHERE ps.preset_id = ? AND ps.detail_id = d.id AND ps.included = 0
		       ) AS included
		FROM item_details d
		JOIN items i ON i.id = d.item_id
		JOIN sections s ON s.id = i.section_id
		ORDER BY s.sort_order, s.id, i.sort_order, i.id, d.sort_order, d.id`, presetID)
	if err != nil {
		return state, fmt.Errorf("load details: %w", err)
	}
	for detailRows.Next() {
		var d models.Detail
		var included int
		if err := detailRows.Scan(&d.ID, &d.ItemID, &d.Text, &included); err != nil {
			detailRows.Close()
			return state, err
		}
		pos, ok := itemIdx[d.ItemID]
		if !ok {
			continue
		}
		d.Included = included == 1
		state.Sections[pos.sec].Items[pos.idx].Details = append(
			state.Sections[pos.sec].Items[pos.idx].Details, d,
		)
	}
	if err := detailRows.Err(); err != nil {
		detailRows.Close()
		return state, err
	}
	detailRows.Close()
	return state, nil
}

func isValidSectionType(sectionType string) bool {
	switch models.SectionType(sectionType) {
	case models.SectionContact, models.SectionSkills, models.SectionExperience,
		models.SectionEducation, models.SectionCustom:
		return true
	}
	return false
}
