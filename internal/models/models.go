// Package models defines the data structures exchanged between the Go
// backend and the Svelte frontend.
package models

// Profile is the single global identity shown at the top of the resume.
type Profile struct {
	Name     string `json:"name"`
	Headline string `json:"headline"`
}

// Detail is a sub-entry of an item, e.g. an achievement bullet in a
// professional experience entry.
type Detail struct {
	ID        int64  `json:"id"`
	ItemID    int64  `json:"itemId"`
	Text      string `json:"text"`
	Included  bool   `json:"included"`
}

// Item is an entry inside a section. The shape of Data depends on the
// section type:
//
//	contact:    {label, value}
//	skills:     {name, level}
//	experience: {role, company, start, end, summary}
//	education:  {school, degree, field, start, end}
//	custom:     {label, text}
type Item struct {
	ID        int64                  `json:"id"`
	SectionID int64                  `json:"sectionId"`
	Data      map[string]interface{} `json:"data"`
	Details   []Detail               `json:"details"`
	Included  bool                   `json:"included"`
}

// SectionType enumerates the supported section kinds.
type SectionType string

const (
	SectionContact    SectionType = "contact"
	SectionSkills     SectionType = "skills"
	SectionExperience SectionType = "experience"
	SectionEducation SectionType = "education"
	SectionCustom    SectionType = "custom"
)

// Section is a resume section resolved for a given preset: Visible,
// Title (including any preset override) and Items are preset-dependent.
// TitleColor and LineColor are optional "#rrggbb" strings; empty means
// the default rendering colors.
type Section struct {
	ID         int64  `json:"id"`
	Type       string `json:"type"`
	Title      string `json:"title"`
	TitleColor string `json:"titleColor"`
	LineColor  string `json:"lineColor"`
	Visible    bool   `json:"visible"`
	Items      []Item `json:"items"`
}

// Preset is a named selection of sections, items and details.
type Preset struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	CreatedAt string `json:"createdAt"`
}

// ResumeState is the full application state resolved for one preset.
type ResumeState struct {
	Profile  Profile   `json:"profile"`
	Presets  []Preset  `json:"presets"`
	Sections []Section `json:"sections"`
}

// PageLayout describes the pagination of a resume for a page format:
// the section IDs that start pages after the first, and the total page
// count. The preview uses it to match the PDF export exactly.
type PageLayout struct {
	Breaks    []int64 `json:"breaks"`
	PageCount int     `json:"pageCount"`
}
