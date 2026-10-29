package store

import (
	"database/sql"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"
)

func newTestStore(t *testing.T) *Store {
	t.Helper()
	dsn := "file:" + filepath.Join(t.TempDir(), "test.db") +
		"?_pragma=journal_mode(WAL)&_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	db.SetMaxOpenConns(1)
	s := &Store{db: db}
	if err := s.migrate(); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	if err := s.seed(); err != nil {
		t.Fatalf("seed: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return s
}

func TestSeedFreshDatabase(t *testing.T) {
	s := newTestStore(t)
	state, err := s.GetResumeState(0)
	if err != nil {
		t.Fatalf("GetResumeState: %v", err)
	}
	if len(state.Sections) != 4 {
		t.Errorf("seeded %d sections, want 4", len(state.Sections))
	}
	if len(state.Presets) != 1 || state.Presets[0].Name != "Default" {
		t.Errorf("presets = %+v, want single Default", state.Presets)
	}
	for _, sec := range state.Sections {
		if !sec.Visible {
			t.Errorf("section %q should be visible by default", sec.Title)
		}
		if sec.Items == nil {
			t.Errorf("section %q items should be non-nil empty slice", sec.Title)
		}
	}
}

func TestSeedDoesNotResurrectDeletedSections(t *testing.T) {
	s := newTestStore(t)
	for _, sec := range []int64{1, 2, 3, 4} {
		if err := s.DeleteSection(sec); err != nil {
			t.Fatalf("delete section %d: %v", sec, err)
		}
	}
	// Simulate an app restart: seed again on the now-empty sections table.
	if err := s.seed(); err != nil {
		t.Fatalf("re-seed: %v", err)
	}
	state, err := s.GetResumeState(0)
	if err != nil {
		t.Fatalf("GetResumeState: %v", err)
	}
	if len(state.Sections) != 0 {
		t.Errorf("deleted sections resurrected: %d sections", len(state.Sections))
	}
	if len(state.Presets) != 1 {
		t.Errorf("preset lost on re-seed: %+v", state.Presets)
	}
}

func TestGetResumeStateAssemblesItemsAndDetails(t *testing.T) {
	s := newTestStore(t)
	state, err := s.GetResumeState(0)
	if err != nil {
		t.Fatalf("GetResumeState: %v", err)
	}
	presetID := state.Presets[0].ID
	var expID, eduID int64
	for _, sec := range state.Sections {
		switch sec.Type {
		case "experience":
			expID = sec.ID
		case "education":
			eduID = sec.ID
		}
	}

	job, err := s.AddItem(expID)
	if err != nil {
		t.Fatalf("AddItem: %v", err)
	}
	if err := s.UpdateItemData(job.ID, map[string]interface{}{
		"role": "Dev", "company": "Acme", "start": "2020", "end": "", "summary": "did things",
	}); err != nil {
		t.Fatalf("UpdateItemData: %v", err)
	}
	d1, err := s.AddDetail(job.ID)
	if err != nil {
		t.Fatalf("AddDetail: %v", err)
	}
	if err := s.UpdateDetail(d1.ID, "first bullet"); err != nil {
		t.Fatalf("UpdateDetail: %v", err)
	}
	d2, err := s.AddDetail(job.ID)
	if err != nil {
		t.Fatalf("AddDetail: %v", err)
	}
	if err := s.UpdateDetail(d2.ID, "second bullet"); err != nil {
		t.Fatalf("UpdateDetail: %v", err)
	}
	degree, err := s.AddItem(eduID)
	if err != nil {
		t.Fatalf("AddItem: %v", err)
	}

	// Toggle the second detail and the degree item off for this preset.
	if err := s.ToggleDetail(presetID, d2.ID, false); err != nil {
		t.Fatalf("ToggleDetail off: %v", err)
	}
	if err := s.ToggleItem(presetID, degree.ID, false); err != nil {
		t.Fatalf("ToggleItem off: %v", err)
	}
	// Hide the education section.
	if err := s.SetSectionVisible(presetID, eduID, false); err != nil {
		t.Fatalf("SetSectionVisible: %v", err)
	}

	got, err := s.GetResumeState(presetID)
	if err != nil {
		t.Fatalf("GetResumeState: %v", err)
	}

	find := func(id int64) (secIdx int, itemIdx int) {
		for si, sec := range got.Sections {
			for ii, it := range sec.Items {
				if it.ID == id {
					return si, ii
				}
			}
		}
		return -1, -1
	}

	si, ii := find(job.ID)
	if si < 0 {
		t.Fatalf("job item not found in state")
	}
	item := got.Sections[si].Items[ii]
	if item.Data["role"] != "Dev" || item.Data["company"] != "Acme" {
		t.Errorf("item data = %v", item.Data)
	}
	if !item.Included {
		t.Errorf("job item should be included")
	}
	if len(item.Details) != 2 {
		t.Fatalf("job details = %d, want 2", len(item.Details))
	}
	if item.Details[0].Text != "first bullet" || !item.Details[0].Included {
		t.Errorf("detail[0] = %+v", item.Details[0])
	}
	if item.Details[1].Text != "second bullet" || item.Details[1].Included {
		t.Errorf("detail[1] should be excluded: %+v", item.Details[1])
	}

	_, eduII := find(degree.ID)
	if eduII < 0 {
		t.Fatalf("degree item not found")
	}
	var eduSec int
	for si, sec := range got.Sections {
		if sec.ID == eduID {
			eduSec = si
		}
	}
	if got.Sections[eduSec].Visible {
		t.Errorf("education section should be hidden")
	}
	if got.Sections[eduSec].Items[eduII].Included {
		t.Errorf("degree item should be excluded")
	}
}

func TestPresetTitleOverride(t *testing.T) {
	s := newTestStore(t)
	state, err := s.GetResumeState(0)
	if err != nil {
		t.Fatalf("GetResumeState: %v", err)
	}
	presetID := state.Presets[0].ID
	secID := state.Sections[0].ID
	if err := s.SetSectionVisible(presetID, secID, true); err != nil {
		t.Fatalf("SetSectionVisible: %v", err)
	}
	if _, err := s.db.Exec(
		`UPDATE preset_sections SET title_override = 'Reach me' WHERE preset_id = ? AND section_id = ?`,
		presetID, secID,
	); err != nil {
		t.Fatalf("set override: %v", err)
	}
	got, err := s.GetResumeState(presetID)
	if err != nil {
		t.Fatalf("GetResumeState: %v", err)
	}
	if got.Sections[0].Title != "Reach me" {
		t.Errorf("title = %q, want preset override %q", got.Sections[0].Title, "Reach me")
	}
}

func TestPresetLifecycle(t *testing.T) {
	s := newTestStore(t)
	if _, err := s.CreatePreset("  "); err == nil {
		t.Errorf("CreatePreset with blank name should fail")
	}
	p, err := s.CreatePreset("Second")
	if err != nil {
		t.Fatalf("CreatePreset: %v", err)
	}
	if _, err := s.DuplicatePreset("", p.ID); err == nil {
		t.Errorf("DuplicatePreset with blank name should fail")
	}
	dup, err := s.DuplicatePreset("Copy", p.ID)
	if err != nil {
		t.Fatalf("DuplicatePreset: %v", err)
	}
	if dup.Name != "Copy" {
		t.Errorf("duplicate name = %q", dup.Name)
	}
	// Deleting the last remaining preset must fail.
	if err := s.DeletePreset(dup.ID); err != nil {
		t.Fatalf("DeletePreset: %v", err)
	}
	if err := s.DeletePreset(p.ID); err != nil {
		t.Fatalf("DeletePreset: %v", err)
	}
	if err := s.DeletePreset(999); err == nil {
		t.Errorf("deleting a missing preset should fail")
	}
	if err := s.DeletePreset(p.ID); err == nil {
		t.Errorf("deleting the last preset should fail")
	}
}

func TestDeleteNotFound(t *testing.T) {
	s := newTestStore(t)
	if err := s.DeleteSection(999); err == nil {
		t.Errorf("DeleteSection(999) should fail")
	}
	if err := s.DeleteItem(999); err == nil {
		t.Errorf("DeleteItem(999) should fail")
	}
	if err := s.DeleteDetail(999); err == nil {
		t.Errorf("DeleteDetail(999) should fail")
	}
}

func TestToggleItemRoundTrip(t *testing.T) {
	s := newTestStore(t)
	state, err := s.GetResumeState(0)
	if err != nil {
		t.Fatalf("GetResumeState: %v", err)
	}
	presetID := state.Presets[0].ID
	secID := state.Sections[0].ID
	item, err := s.AddItem(secID)
	if err != nil {
		t.Fatalf("AddItem: %v", err)
	}
	for _, want := range []bool{false, true, false} {
		if err := s.ToggleItem(presetID, item.ID, want); err != nil {
			t.Fatalf("ToggleItem(%v): %v", want, err)
		}
		got, err := s.GetResumeState(presetID)
		if err != nil {
			t.Fatalf("GetResumeState: %v", err)
		}
		if got.Sections[0].Items[0].Included != want {
			t.Errorf("after ToggleItem(%v): included=%v", want, got.Sections[0].Items[0].Included)
		}
	}
}
