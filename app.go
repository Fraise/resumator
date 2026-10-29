package main

import (
	"context"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"resumator/internal/models"
	"resumator/internal/pdf"
	"resumator/internal/store"
)

// App is the Wails application facade. All public methods are bound to
// the frontend.
type App struct {
	ctx       context.Context
	store     *store.Store
	storeOpen error
}

// NewApp creates the application.
func NewApp() *App {
	return &App{}
}

// startup opens the database when the app starts.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.store, a.storeOpen = store.Open()
}

// shutdown closes the database.
func (a *App) shutdown(ctx context.Context) {
	if a.store != nil {
		a.store.Close()
	}
}

func (a *App) st() (*store.Store, error) {
	if a.storeOpen != nil {
		return nil, a.storeOpen
	}
	return a.store, nil
}

// GetResumeState returns the full state resolved for one preset.
func (a *App) GetResumeState(presetID int64) (models.ResumeState, error) {
	s, err := a.st()
	if err != nil {
		return models.ResumeState{}, err
	}
	return s.GetResumeState(presetID)
}

// GetPageLayout returns the pagination the PDF export will use, so the
// preview can match it.
func (a *App) GetPageLayout(presetID int64, format string) (models.PageLayout, error) {
	s, err := a.st()
	if err != nil {
		return models.PageLayout{}, err
	}
	state, err := s.GetResumeState(presetID)
	if err != nil {
		return models.PageLayout{}, err
	}
	breaks, pageCount := pdf.Layout(state, format)
	// JSON marshals a nil slice as null; the frontend expects an array.
	if breaks == nil {
		breaks = []int64{}
	}
	return models.PageLayout{Breaks: breaks, PageCount: pageCount}, nil
}

// UpdateContactHeader saves the name and headline of a contact section,
// shown at the top of every resume.
func (a *App) UpdateContactHeader(sectionID int64, name, headline string) error {
	s, err := a.st()
	if err != nil {
		return err
	}
	return s.UpdateContactHeader(sectionID, name, headline)
}

// UpdateTextStyles saves the global typography applied to every resume.
func (a *App) UpdateTextStyles(styles models.TextStyles) error {
	s, err := a.st()
	if err != nil {
		return err
	}
	return s.UpdateTextStyles(styles)
}

// AddSection creates a section and returns its id.
func (a *App) AddSection(sectionType, title string) (int64, error) {
	s, err := a.st()
	if err != nil {
		return 0, err
	}
	return s.AddSection(sectionType, title)
}

// UpdateSection renames a section.
func (a *App) UpdateSection(id int64, title string) error {
	s, err := a.st()
	if err != nil {
		return err
	}
	return s.UpdateSection(id, title)
}

// DeleteSection removes a section with all its items.
func (a *App) DeleteSection(id int64) error {
	s, err := a.st()
	if err != nil {
		return err
	}
	return s.DeleteSection(id)
}

// ReorderSections stores the full section order.
func (a *App) ReorderSections(orderedIDs []int64) error {
	s, err := a.st()
	if err != nil {
		return err
	}
	return s.ReorderSections(orderedIDs)
}

// AddItem adds an empty item to a section.
func (a *App) AddItem(sectionID int64) (models.Item, error) {
	s, err := a.st()
	if err != nil {
		return models.Item{}, err
	}
	return s.AddItem(sectionID)
}

// UpdateItemData replaces the data of an item.
func (a *App) UpdateItemData(id int64, data map[string]interface{}) error {
	s, err := a.st()
	if err != nil {
		return err
	}
	return s.UpdateItemData(id, data)
}

// DeleteItem removes an item with its details.
func (a *App) DeleteItem(id int64) error {
	s, err := a.st()
	if err != nil {
		return err
	}
	return s.DeleteItem(id)
}

// ReorderItems stores the item order within a section.
func (a *App) ReorderItems(sectionID int64, orderedIDs []int64) error {
	s, err := a.st()
	if err != nil {
		return err
	}
	return s.ReorderItems(sectionID, orderedIDs)
}

// AddDetail adds an empty detail (bullet) to an item.
func (a *App) AddDetail(itemID int64) (models.Detail, error) {
	s, err := a.st()
	if err != nil {
		return models.Detail{}, err
	}
	return s.AddDetail(itemID)
}

// UpdateDetail changes a detail text.
func (a *App) UpdateDetail(id int64, text string) error {
	s, err := a.st()
	if err != nil {
		return err
	}
	return s.UpdateDetail(id, text)
}

// DeleteDetail removes a detail.
func (a *App) DeleteDetail(id int64) error {
	s, err := a.st()
	if err != nil {
		return err
	}
	return s.DeleteDetail(id)
}

// ReorderDetails stores the detail order within an item.
func (a *App) ReorderDetails(itemID int64, orderedIDs []int64) error {
	s, err := a.st()
	if err != nil {
		return err
	}
	return s.ReorderDetails(itemID, orderedIDs)
}

// ToggleItem includes or excludes an item from a preset.
func (a *App) ToggleItem(presetID, itemID int64, included bool) error {
	s, err := a.st()
	if err != nil {
		return err
	}
	return s.ToggleItem(presetID, itemID, included)
}

// ToggleDetail includes or excludes a detail from a preset.
func (a *App) ToggleDetail(presetID, detailID int64, included bool) error {
	s, err := a.st()
	if err != nil {
		return err
	}
	return s.ToggleDetail(presetID, detailID, included)
}

// SetSectionVisible shows or hides a section in a preset.
func (a *App) SetSectionVisible(presetID, sectionID int64, visible bool) error {
	s, err := a.st()
	if err != nil {
		return err
	}
	return s.SetSectionVisible(presetID, sectionID, visible)
}

// SetSectionColors sets the optional title and separator colors for a
// section (empty strings restore the defaults).
func (a *App) SetSectionColors(sectionID int64, titleColor, lineColor string) error {
	s, err := a.st()
	if err != nil {
		return err
	}
	return s.SetSectionColors(sectionID, titleColor, lineColor)
}

// CreatePreset adds a new preset with default selections.
func (a *App) CreatePreset(name string) (models.Preset, error) {
	s, err := a.st()
	if err != nil {
		return models.Preset{}, err
	}
	return s.CreatePreset(name)
}

// DuplicatePreset copies a preset under a new name.
func (a *App) DuplicatePreset(name string, sourceID int64) (models.Preset, error) {
	s, err := a.st()
	if err != nil {
		return models.Preset{}, err
	}
	return s.DuplicatePreset(name, sourceID)
}

// RenamePreset renames a preset.
func (a *App) RenamePreset(id int64, name string) error {
	s, err := a.st()
	if err != nil {
		return err
	}
	return s.RenamePreset(id, name)
}

// DeletePreset removes a preset, unless it is the last one.
func (a *App) DeletePreset(id int64) error {
	s, err := a.st()
	if err != nil {
		return err
	}
	return s.DeletePreset(id)
}

// ExportPDF renders the given preset to a PDF chosen via save dialog.
// The format is "a4" (default) or "letter". It returns the saved path, or
// an empty string when cancelled.
func (a *App) ExportPDF(presetID int64, format string) (string, error) {
	s, err := a.st()
	if err != nil {
		return "", err
	}
	state, err := s.GetResumeState(presetID)
	if err != nil {
		return "", err
	}
	path, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           "Export resume as PDF",
		DefaultFilename: "resume.pdf",
		Filters: []runtime.FileFilter{
			{DisplayName: "PDF (*.pdf)", Pattern: "*.pdf"},
		},
	})
	if err != nil {
		return "", err
	}
	if path == "" {
		return "", nil
	}
	bytes, err := pdf.Generate(state, format)
	if err != nil {
		return "", err
	}
	if err := writeAll(path, bytes); err != nil {
		return "", err
	}
	return path, nil
}
