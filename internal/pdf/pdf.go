// Package pdf renders a resume state to a PDF document. The layout
// mirrors the frontend preview: name/headline/contact centered at the
// top, then one titled block per section. All text rows use automatic
// heights so wrapped text grows its row instead of overlapping.
package pdf

import (
	"strings"

	"github.com/johnfercher/maroto/v2"
	"github.com/johnfercher/maroto/v2/pkg/components/line"
	"github.com/johnfercher/maroto/v2/pkg/components/text"
	"github.com/johnfercher/maroto/v2/pkg/config"
	"github.com/johnfercher/maroto/v2/pkg/consts/align"
	"github.com/johnfercher/maroto/v2/pkg/consts/fontstyle"
	"github.com/johnfercher/maroto/v2/pkg/core"
	"github.com/johnfercher/maroto/v2/pkg/props"

	"resumator/internal/models"
)

var (
	gray = &props.Color{Red: 100, Green: 105, Blue: 112}
)

// Generate renders the given resume state to PDF bytes.
func Generate(state models.ResumeState) ([]byte, error) {
	cfg := config.NewBuilder().
		WithTopMargin(12).
		WithBottomMargin(12).
		WithLeftMargin(14).
		WithRightMargin(14).
		Build()
	m := maroto.New(cfg)

	addHeader(m, state)

	for _, section := range state.Sections {
		// Contact values are rendered as part of the header.
		if !section.Visible || section.Type == string(models.SectionContact) {
			continue
		}
		addSectionTitle(m, section.Title)
		switch models.SectionType(section.Type) {
		case models.SectionSkills:
			addSkills(m, section)
		case models.SectionExperience:
			addExperience(m, section)
		case models.SectionEducation:
			addEducation(m, section)
		default:
			addCustomItems(m, section)
		}
	}

	doc, err := m.Generate()
	if err != nil {
		return nil, err
	}
	return doc.GetBytes(), nil
}

func addHeader(m core.Maroto, state models.ResumeState) {
	if name := strings.TrimSpace(state.Profile.Name); name != "" {
		m.AddRows(text.NewAutoRow(name, props.Text{
			Style: fontstyle.Bold, Size: 18, Align: align.Center,
		}))
	}
	if headline := strings.TrimSpace(state.Profile.Headline); headline != "" {
		m.AddRows(text.NewAutoRow(headline, props.Text{
			Style: fontstyle.Italic, Size: 10.5, Align: align.Center,
			Color: gray, Top: 1.5, VerticalPadding: 0.5,
		}))
	}
	if contact := contactLine(state); contact != "" {
		m.AddRows(text.NewAutoRow(contact, props.Text{
			Size: 9.5, Align: align.Center, Color: gray, Top: 2, VerticalPadding: 0.5,
		}))
	}
}

// contactLine joins the values of the first visible contact section,
// mirroring the preview header.
func contactLine(state models.ResumeState) string {
	for _, section := range state.Sections {
		if section.Type != string(models.SectionContact) || !section.Visible {
			continue
		}
		parts := []string{}
		for _, it := range section.Items {
			if !it.Included {
				continue
			}
			if v := strings.TrimSpace(str(it.Data, "value")); v != "" {
				parts = append(parts, v)
			}
		}
		if len(parts) > 0 {
			return strings.Join(parts, "   •   ")
		}
		return ""
	}
	return ""
}

func addSectionTitle(m core.Maroto, title string) {
	title = strings.TrimSpace(title)
	if title == "" {
		return
	}
	m.AddRows(text.NewAutoRow(strings.ToUpper(title), props.Text{
		Style: fontstyle.Bold, Size: 10.5, Top: 3,
	}))
	// The line row is 3mm tall with the line drawn at 60% height, which
	// leaves roughly 1mm of air between the title glyphs and the line
	// (auto rows slightly underestimate text descent). SizePercent 100
	// makes the line span the full text column, aligned with the words.
	m.AddRows(line.NewRow(3, props.Line{
		Color: gray, SizePercent: 100, OffsetPercent: 60,
	}))
}

func addSkills(m core.Maroto, section models.Section) {
	// Skill groups are marker items with a divider flag; the skills listed
	// after a marker belong to its group and render on their own line.
	type skillGroup struct {
		title  string
		skills []string
	}
	groups := []skillGroup{{}}
	for _, it := range section.Items {
		if boolVal(it.Data, "divider") {
			title := ""
			if it.Included {
				title = strings.TrimSpace(str(it.Data, "title"))
			}
			groups = append(groups, skillGroup{title: title})
			continue
		}
		if !it.Included {
			continue
		}
		name := strings.TrimSpace(str(it.Data, "name"))
		if name == "" {
			continue
		}
		g := &groups[len(groups)-1]
		if level := strings.TrimSpace(str(it.Data, "level")); level != "" {
			g.skills = append(g.skills, name+" ("+level+")")
		} else {
			g.skills = append(g.skills, name)
		}
	}
	for _, g := range groups {
		if g.title != "" {
			m.AddRows(text.NewAutoRow(g.title, props.Text{
				Style: fontstyle.Bold, Size: 9.5, Top: 1.5,
			}))
		}
		if len(g.skills) > 0 {
			m.AddRows(text.NewAutoRow(strings.Join(g.skills, ", "), props.Text{
				Size: 9.5, Top: 1, VerticalPadding: 0.5,
			}))
		}
	}
}

func dateRange(data map[string]interface{}) string {
	start := strings.TrimSpace(str(data, "start"))
	end := strings.TrimSpace(str(data, "end"))
	switch {
	case start == "" && end == "":
		return ""
	case end == "":
		return start + " – Present"
	case start == "":
		return end
	}
	return start + " – " + end
}

func addExperience(m core.Maroto, section models.Section) {
	for _, it := range section.Items {
		if !it.Included {
			continue
		}
		role := strings.TrimSpace(str(it.Data, "role"))
		company := strings.TrimSpace(str(it.Data, "company"))
		title := role
		if role != "" && company != "" {
			title = role + " — " + company
		} else {
			title = role + company
		}
		addTitleWithDates(m, title, dateRange(it.Data))
		if summary := strings.TrimSpace(str(it.Data, "summary")); summary != "" {
			m.AddRows(text.NewAutoRow(summary, props.Text{
				Size: 9.5, Color: gray, Top: 1, VerticalPadding: 0.5,
			}))
		}
		addDetails(m, it.Details)
	}
}

func addEducation(m core.Maroto, section models.Section) {
	for _, it := range section.Items {
		if !it.Included {
			continue
		}
		// Mirrors the preview: "degree — field, school".
		left := strings.TrimSpace(str(it.Data, "degree"))
		if field := strings.TrimSpace(str(it.Data, "field")); field != "" {
			if left == "" {
				left = field
			} else {
				left += " — " + field
			}
		}
		school := strings.TrimSpace(str(it.Data, "school"))
		title := left
		if school != "" {
			if title == "" {
				title = school
			} else {
				title += ", " + school
			}
		}
		addTitleWithDates(m, title, dateRange(it.Data))
		addDetails(m, it.Details)
	}
}

func addCustomItems(m core.Maroto, section models.Section) {
	for _, it := range section.Items {
		if !it.Included {
			continue
		}
		if label := strings.TrimSpace(str(it.Data, "label")); label != "" {
			m.AddRows(text.NewAutoRow(label, props.Text{
				Style: fontstyle.Bold, Size: 10, Top: 2,
			}))
		}
		if body := strings.TrimSpace(str(it.Data, "text")); body != "" {
			m.AddRows(text.NewAutoRow(body, props.Text{
				Size: 9.5, Top: 1, VerticalPadding: 0.5,
			}))
		}
		addDetails(m, it.Details)
	}
}

// addTitleWithDates renders a bold entry title with the date range
// right-aligned on the same line. The row height is automatic, so long
// titles wrap instead of overlapping the next row.
func addTitleWithDates(m core.Maroto, title, dates string) {
	title = strings.TrimSpace(title)
	dates = strings.TrimSpace(dates)
	switch {
	case title == "" && dates == "":
		return
	case title == "":
		m.AddAutoRow(text.NewCol(12, dates, props.Text{
			Size: 9, Align: align.Right, Color: gray, Top: 2.5,
		}))
	case dates == "":
		m.AddAutoRow(text.NewCol(12, title, props.Text{
			Style: fontstyle.Bold, Size: 10, Top: 2,
		}))
	default:
		m.AddAutoRow(
			text.NewCol(9, title, props.Text{Style: fontstyle.Bold, Size: 10, Top: 2}),
			text.NewCol(3, dates, props.Text{Size: 9, Align: align.Right, Color: gray, Top: 2.5}),
		)
	}
}

func addDetails(m core.Maroto, details []models.Detail) {
	for _, d := range details {
		if !d.Included || strings.TrimSpace(d.Text) == "" {
			continue
		}
		m.AddRows(text.NewAutoRow("• "+strings.TrimSpace(d.Text), props.Text{
			Size: 9.5, Left: 3, Top: 0.8, VerticalPadding: 0.5,
		}))
	}
}

func boolVal(m map[string]interface{}, key string) bool {
	if m == nil {
		return false
	}
	if v, ok := m[key]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return false
}

func str(m map[string]interface{}, key string) string {
	if m == nil {
		return ""
	}
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}
