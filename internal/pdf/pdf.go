// Package pdf renders a resume state to a PDF document.
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

	addHeader(m, state.Profile)

	for _, section := range state.Sections {
		if !section.Visible {
			continue
		}
		addSectionTitle(m, section.Title)
		switch models.SectionType(section.Type) {
		case models.SectionContact:
			addContact(m, section)
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

func addHeader(m core.Maroto, profile models.Profile) {
	if strings.TrimSpace(profile.Name) != "" {
		m.AddRow(9, text.NewCol(12, profile.Name, props.Text{
			Style: fontstyle.Bold, Size: 18, Align: align.Center,
		}))
	}
	if strings.TrimSpace(profile.Headline) != "" {
		m.AddRow(5, text.NewCol(12, profile.Headline, props.Text{
			Style: fontstyle.Italic, Size: 11, Align: align.Center, Color: gray,
		}))
	}
}

func addSectionTitle(m core.Maroto, title string) {
	title = strings.TrimSpace(title)
	if title == "" {
		return
	}
	m.AddRow(6, text.NewCol(12, strings.ToUpper(title), props.Text{
		Style: fontstyle.Bold, Size: 10.5, Top: 2,
	}))
	m.AddRows(line.NewRow(1, props.Line{Color: gray}))
}

func contactValue(it models.Item) string {
	label := str(it.Data, "label")
	value := str(it.Data, "value")
	if strings.TrimSpace(value) == "" {
		return ""
	}
	if strings.TrimSpace(label) == "" {
		return value
	}
	return value
}

func addContact(m core.Maroto, section models.Section) {
	parts := []string{}
	for _, it := range section.Items {
		if !it.Included {
			continue
		}
		if v := contactValue(it); v != "" {
			parts = append(parts, v)
		}
	}
	if len(parts) == 0 {
		return
	}
	m.AddRows(text.NewAutoRow(strings.Join(parts, "   •   "), props.Text{
		Size: 9.5, Color: gray,
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
			m.AddRow(4.5, text.NewCol(12, g.title, props.Text{
				Style: fontstyle.Bold, Size: 9.5, Top: 1,
			}))
		}
		if len(g.skills) > 0 {
			m.AddRows(text.NewAutoRow(strings.Join(g.skills, ", "), props.Text{
				Size: 9.5,
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
		addTitleWithDates(m, str(it.Data, "role"), str(it.Data, "company"), dateRange(it.Data))
		if summary := strings.TrimSpace(str(it.Data, "summary")); summary != "" {
			m.AddRows(text.NewAutoRow(summary, props.Text{Size: 9.5, Color: gray}))
		}
		addDetails(m, it.Details)
	}
}

func addEducation(m core.Maroto, section models.Section) {
	for _, it := range section.Items {
		if !it.Included {
			continue
		}
		title := strings.TrimSpace(str(it.Data, "degree"))
		if field := strings.TrimSpace(str(it.Data, "field")); field != "" {
			if title == "" {
				title = field
			} else {
				title += " — " + field
			}
		}
		addTitleWithDates(m, title, str(it.Data, "school"), dateRange(it.Data))
		addDetails(m, it.Details)
	}
}

func addCustomItems(m core.Maroto, section models.Section) {
	for _, it := range section.Items {
		if !it.Included {
			continue
		}
		if label := strings.TrimSpace(str(it.Data, "label")); label != "" {
			m.AddRow(5, text.NewCol(12, label, props.Text{Style: fontstyle.Bold, Size: 10, Top: 1}))
		}
		if body := strings.TrimSpace(str(it.Data, "text")); body != "" {
			m.AddRows(text.NewAutoRow(body, props.Text{Size: 9.5}))
		}
		addDetails(m, it.Details)
	}
}

func addTitleWithDates(m core.Maroto, left, right, dates string) {
	left = strings.TrimSpace(left)
	right = strings.TrimSpace(right)
	title := left
	if right != "" {
		if title == "" {
			title = right
		} else {
			title += " — " + right
		}
	}
	if dates == "" {
		m.AddRow(5, text.NewCol(12, title, props.Text{Style: fontstyle.Bold, Size: 10, Top: 1}))
		return
	}
	m.AddRow(5,
		text.NewCol(9, title, props.Text{Style: fontstyle.Bold, Size: 10, Top: 1}),
		text.NewCol(3, dates, props.Text{Size: 9, Align: align.Right, Color: gray, Top: 1.5}),
	)
}

func addDetails(m core.Maroto, details []models.Detail) {
	for _, d := range details {
		if !d.Included || strings.TrimSpace(d.Text) == "" {
			continue
		}
		m.AddRows(text.NewAutoRow("• "+strings.TrimSpace(d.Text), props.Text{
			Size: 9.5, Left: 3,
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
