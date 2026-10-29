// Package models defines the data structures exchanged between the Go
// backend and the Svelte frontend.
package models

// Detail is a sub-entry of an item, e.g. an achievement bullet in a
// professional experience entry.
type Detail struct {
	ID       int64  `json:"id"`
	ItemID   int64  `json:"itemId"`
	Text     string `json:"text"`
	Included bool   `json:"included"`
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
	SectionEducation  SectionType = "education"
	SectionCustom     SectionType = "custom"
)

// Section is a resume section resolved for a given preset: Visible,
// Title (including any preset override) and Items are preset-dependent.
// TitleColor and LineColor are optional "#rrggbb" strings; empty means
// the default rendering colors. Name and Headline carry the identity
// shown at the top of the resume; only contact sections use them.
type Section struct {
	ID         int64  `json:"id"`
	Type       string `json:"type"`
	Title      string `json:"title"`
	Name       string `json:"name"`
	Headline   string `json:"headline"`
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

// TextStyle describes the typography of one text role on the resume.
type TextStyle struct {
	Size   float64 `json:"size"`
	Bold   bool    `json:"bold"`
	Italic bool    `json:"italic"`
}

// TextStyles holds the typography of every text role. These settings
// are global: they apply to every section of every preset.
type TextStyles struct {
	SectionTitle TextStyle `json:"sectionTitle"`
	EntryTitle   TextStyle `json:"entryTitle"`
	EntryDates   TextStyle `json:"entryDates"`
	Body         TextStyle `json:"body"`
}

// DefaultTextStyles returns the default typography.
func DefaultTextStyles() TextStyles {
	return TextStyles{
		SectionTitle: TextStyle{Size: 10.5, Bold: true},
		EntryTitle:   TextStyle{Size: 10, Bold: true},
		EntryDates:   TextStyle{Size: 9},
		Body:         TextStyle{Size: 9.5},
	}
}

// ResumeState is the full application state resolved for one preset.
type ResumeState struct {
	Styles   TextStyles `json:"styles"`
	Presets  []Preset   `json:"presets"`
	Sections []Section  `json:"sections"`
}

// PageLayout describes the pagination of a resume for a page format:
// the section IDs that start pages after the first, and the total page
// count. The preview uses it to match the PDF export exactly.
type PageLayout struct {
	Breaks    []int64 `json:"breaks"`
	PageCount int     `json:"pageCount"`
}
