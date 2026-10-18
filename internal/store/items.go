package store

import (
	"encoding/json"
	"fmt"

	"resumator/internal/models"
)

// defaultItemData returns the initial field values for a new item of the
// given section type.
func defaultItemData(sectionType string) map[string]interface{} {
	switch models.SectionType(sectionType) {
	case models.SectionContact:
		return map[string]interface{}{"label": "Email", "value": ""}
	case models.SectionSkills:
		return map[string]interface{}{"name": "", "level": ""}
	case models.SectionExperience:
		return map[string]interface{}{
			"role": "", "company": "", "start": "", "end": "", "summary": "",
		}
	case models.SectionEducation:
		return map[string]interface{}{
			"school": "", "degree": "", "field": "", "start": "", "end": "",
		}
	default: // custom
		return map[string]interface{}{"label": "", "text": ""}
	}
}

func (s *Store) sectionType(sectionID int64) (string, error) {
	var sectionType string
	err := s.db.QueryRow(`SELECT type FROM sections WHERE id = ?`, sectionID).Scan(&sectionType)
	if err != nil {
		return "", fmt.Errorf("section %d not found", sectionID)
	}
	return sectionType, nil
}

// AddItem creates an empty item in the given section.
func (s *Store) AddItem(sectionID int64) (models.Item, error) {
	sectionType, err := s.sectionType(sectionID)
	if err != nil {
		return models.Item{}, err
	}
	var maxOrder int
	if err := s.db.QueryRow(
		`SELECT COALESCE(MAX(sort_order), -1) FROM items WHERE section_id = ?`, sectionID,
	).Scan(&maxOrder); err != nil {
		return models.Item{}, err
	}
	raw, err := json.Marshal(defaultItemData(sectionType))
	if err != nil {
		return models.Item{}, err
	}
	res, err := s.db.Exec(
		`INSERT INTO items (section_id, sort_order, data) VALUES (?, ?, ?)`,
		sectionID, maxOrder+1, string(raw),
	)
	if err != nil {
		return models.Item{}, fmt.Errorf("add item: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return models.Item{}, err
	}
	return models.Item{
		ID:        id,
		SectionID: sectionID,
		Data:      decodeItemData(string(raw)),
		Details:   []models.Detail{},
		Included:  true,
	}, nil
}

// UpdateItemData replaces the JSON data of an item.
func (s *Store) UpdateItemData(id int64, data map[string]interface{}) error {
	raw, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("encode item data: %w", err)
	}
	res, err := s.db.Exec(`UPDATE items SET data = ? WHERE id = ?`, string(raw), id)
	if err != nil {
		return fmt.Errorf("update item: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return fmt.Errorf("item %d not found", id)
	}
	return nil
}

// DeleteItem removes an item and its details.
func (s *Store) DeleteItem(id int64) error {
	if _, err := s.db.Exec(`DELETE FROM items WHERE id = ?`, id); err != nil {
		return fmt.Errorf("delete item: %w", err)
	}
	return nil
}

// ReorderItems stores the given order for the items of a section.
func (s *Store) ReorderItems(sectionID int64, orderedIDs []int64) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for i, id := range orderedIDs {
		if _, err := tx.Exec(
			`UPDATE items SET sort_order = ? WHERE id = ? AND section_id = ?`,
			i, id, sectionID,
		); err != nil {
			return fmt.Errorf("reorder items: %w", err)
		}
	}
	return tx.Commit()
}

// AddDetail adds an empty detail (bullet) to an item.
func (s *Store) AddDetail(itemID int64) (models.Detail, error) {
	var maxOrder int
	if err := s.db.QueryRow(
		`SELECT COALESCE(MAX(sort_order), -1) FROM item_details WHERE item_id = ?`, itemID,
	).Scan(&maxOrder); err != nil {
		return models.Detail{}, err
	}
	res, err := s.db.Exec(
		`INSERT INTO item_details (item_id, sort_order, text) VALUES (?, ?, '')`,
		itemID, maxOrder+1,
	)
	if err != nil {
		return models.Detail{}, fmt.Errorf("add detail: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return models.Detail{}, err
	}
	return models.Detail{ID: id, ItemID: itemID, Text: "", Included: true}, nil
}

// UpdateDetail changes the text of a detail.
func (s *Store) UpdateDetail(id int64, text string) error {
	res, err := s.db.Exec(`UPDATE item_details SET text = ? WHERE id = ?`, text, id)
	if err != nil {
		return fmt.Errorf("update detail: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return fmt.Errorf("detail %d not found", id)
	}
	return nil
}

// DeleteDetail removes a detail.
func (s *Store) DeleteDetail(id int64) error {
	if _, err := s.db.Exec(`DELETE FROM item_details WHERE id = ?`, id); err != nil {
		return fmt.Errorf("delete detail: %w", err)
	}
	return nil
}

// ReorderDetails stores the given order for the details of an item.
func (s *Store) ReorderDetails(itemID int64, orderedIDs []int64) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for i, id := range orderedIDs {
		if _, err := tx.Exec(
			`UPDATE item_details SET sort_order = ? WHERE id = ? AND item_id = ?`,
			i, id, itemID,
		); err != nil {
			return fmt.Errorf("reorder details: %w", err)
		}
	}
	return tx.Commit()
}
