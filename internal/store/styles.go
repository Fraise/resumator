package store

import (
	"database/sql"
	"fmt"

	"resumator/internal/models"
)

const styleColumns = `section_title_size, section_title_bold, section_title_italic,
	entry_title_size, entry_title_bold, entry_title_italic,
	entry_dates_size, entry_dates_bold, entry_dates_italic,
	body_size, body_bold, body_italic`

// GetTextStyles returns the global typography, seeding the default row
// on first use.
func (s *Store) GetTextStyles() (models.TextStyles, error) {
	// Ensure the row exists (databases created before this table existed
	// get it via migration, but seed defensively anyway).
	if _, err := s.db.Exec(`INSERT OR IGNORE INTO text_styles (id) VALUES (1)`); err != nil {
		return models.DefaultTextStyles(), fmt.Errorf("seed text styles: %w", err)
	}

	var styles models.TextStyles
	var stBold, stItalic, etBold, etItalic int
	var edBold, edItalic, bBold, bItalic int
	row := s.db.QueryRow(`SELECT `+styleColumns+` FROM text_styles WHERE id = 1`)
	err := row.Scan(
		&styles.SectionTitle.Size, &stBold, &stItalic,
		&styles.EntryTitle.Size, &etBold, &etItalic,
		&styles.EntryDates.Size, &edBold, &edItalic,
		&styles.Body.Size, &bBold, &bItalic,
	)
	if err != nil {
		return models.DefaultTextStyles(), fmt.Errorf("get text styles: %w", err)
	}
	styles.SectionTitle.Bold = stBold == 1
	styles.SectionTitle.Italic = stItalic == 1
	styles.EntryTitle.Bold = etBold == 1
	styles.EntryTitle.Italic = etItalic == 1
	styles.EntryDates.Bold = edBold == 1
	styles.EntryDates.Italic = edItalic == 1
	styles.Body.Bold = bBold == 1
	styles.Body.Italic = bItalic == 1
	return styles, nil
}

// UpdateTextStyles saves the global typography.
func (s *Store) UpdateTextStyles(styles models.TextStyles) error {
	b := func(v bool) int {
		if v {
			return 1
		}
		return 0
	}
	res, err := s.db.Exec(`
		UPDATE text_styles SET
			section_title_size = ?, section_title_bold = ?, section_title_italic = ?,
			entry_title_size = ?, entry_title_bold = ?, entry_title_italic = ?,
			entry_dates_size = ?, entry_dates_bold = ?, entry_dates_italic = ?,
			body_size = ?, body_bold = ?, body_italic = ?
		WHERE id = 1`,
		styles.SectionTitle.Size, b(styles.SectionTitle.Bold), b(styles.SectionTitle.Italic),
		styles.EntryTitle.Size, b(styles.EntryTitle.Bold), b(styles.EntryTitle.Italic),
		styles.EntryDates.Size, b(styles.EntryDates.Bold), b(styles.EntryDates.Italic),
		styles.Body.Size, b(styles.Body.Bold), b(styles.Body.Italic),
	)
	if err != nil {
		return fmt.Errorf("update text styles: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return sql.ErrNoRows
	}
	return nil
}
