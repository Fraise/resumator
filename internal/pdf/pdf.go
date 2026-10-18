// Package pdf renders a resume state to a PDF document. The layout
// mirrors the frontend preview: name/headline/contact centered at the
// top, then one titled block per section. All text rows use automatic
// heights so wrapped text grows its row instead of overlapping.
//
// Sections are never split across pages: the document is built as a
// sequence of blocks whose heights are known in advance (replicating
// maroto's text metrics), and a page break is forced before any section
// that would not fit on the current page.
package pdf

import (
	"fmt"
	"math"
	"strings"

	"github.com/johnfercher/go-tree/node"
	"github.com/johnfercher/maroto/v2"
	"github.com/johnfercher/maroto/v2/pkg/components/line"
	"github.com/johnfercher/maroto/v2/pkg/components/row"
	"github.com/johnfercher/maroto/v2/pkg/components/text"
	"github.com/johnfercher/maroto/v2/pkg/config"
	"github.com/johnfercher/maroto/v2/pkg/consts/align"
	"github.com/johnfercher/maroto/v2/pkg/consts/fontstyle"
	"github.com/johnfercher/maroto/v2/pkg/consts/pagesize"
	"github.com/johnfercher/maroto/v2/pkg/core"
	"github.com/johnfercher/maroto/v2/pkg/core/entity"
	"github.com/johnfercher/maroto/v2/pkg/props"

	"resumator/internal/models"
)

var (
	gray = &props.Color{Red: 100, Green: 105, Blue: 112}
)

// Layout constants, shared by the document config and the paginator.
const (
	marginTop    = 12.0
	marginBottom = 12.0
	marginLeft   = 14.0
	marginRight  = 14.0
)

// pageDims returns the page dimensions in millimetres for a format.
func pageDims(format string) (width, height float64) {
	if strings.EqualFold(strings.TrimSpace(format), "letter") {
		return 215.9, 279.4
	}
	return 210.0, 297.0
}

// Generate renders the given resume state to PDF bytes. The format is
// "a4" (default) or "letter" and selects the page size.
func Generate(state models.ResumeState, format string) ([]byte, error) {
	_, pageHeight := pageDims(format)
	size := pagesize.A4
	if strings.EqualFold(strings.TrimSpace(format), "letter") {
		size = pagesize.Letter
	}
	cfg := config.NewBuilder().
		WithTopMargin(marginTop).
		WithBottomMargin(marginBottom).
		WithLeftMargin(marginLeft).
		WithRightMargin(marginRight).
		WithPageSize(size).
		Build()
	m := maroto.New(cfg)

	blocks, _ := buildBlocks(state, format)
	for pi, pageBlocks := range layout(blocks, pageHeight-marginTop-marginBottom) {
		if pi > 0 {
			// Force a page break exactly where the layout engine decided,
			// so the exported pages match Layout (and the preview).
			m.AddPages(emptyPage{})
		}
		for _, bi := range pageBlocks {
			for _, r := range blocks[bi] {
				m.AddRows(r.row)
			}
		}
	}

	d, err := m.Generate()
	if err != nil {
		return nil, err
	}
	return d.GetBytes(), nil
}

// Layout returns the pagination the PDF export will use: the section
// IDs that start pages after the first, and the total page count. The
// preview uses this so its pages match the export exactly.
func Layout(state models.ResumeState, format string) (breaks []int64, pageCount int) {
	_, pageHeight := pageDims(format)
	useful := pageHeight - marginTop - marginBottom

	blocks, owners := buildBlocks(state, format)
	pages := layout(blocks, useful)
	for pi, pageBlocks := range pages {
		if pi == 0 {
			continue
		}
		breaks = append(breaks, owners[pageBlocks[0]])
	}
	// A block taller than a whole page is row-split by maroto across
	// additional pages; count those overflow pages too.
	for _, pageBlocks := range pages {
		used := 0.0
		for _, bi := range pageBlocks {
			for _, r := range blocks[bi] {
				used += r.height
			}
		}
		if used <= useful || used == 0 {
			pageCount++
		} else {
			pageCount += int(math.Ceil(used / useful))
		}
	}
	return breaks, pageCount
}

// layout distributes whole blocks over pages of the given content
// height: a block that does not fit in the remaining space starts the
// next page, so sections are never cut in half.
func layout(blocks [][]blockRow, pageHeight float64) [][]int {
	var pages [][]int
	current := []int{}
	used := 0.0
	for i, block := range blocks {
		total := 0.0
		for _, r := range block {
			total += r.height
		}
		if len(current) > 0 && used+total > pageHeight {
			pages = append(pages, current)
			current = nil
			used = 0
		}
		current = append(current, i)
		used += total
	}
	if len(current) > 0 {
		pages = append(pages, current)
	}
	return pages
}

// buildBlocks assembles the header block and one block per visible
// section, with row heights known in advance. owners holds the section
// ID owning each block (0 for the header).
func buildBlocks(state models.ResumeState, format string) ([][]blockRow, []int64) {
	pageWidth, _ := pageDims(format)
	doc := &docBuilder{
		meas:         newMeasurer(),
		contentWidth: pageWidth - marginLeft - marginRight,
	}

	var blocks [][]blockRow
	var owners []int64
	if b := doc.buildHeader(state); len(b) > 0 {
		blocks = append(blocks, b)
		owners = append(owners, 0)
	}
	for _, section := range state.Sections {
		// Contact values are rendered as part of the header.
		if !section.Visible || section.Type == string(models.SectionContact) {
			continue
		}
		if b := doc.buildSection(section); len(b) > 0 {
			blocks = append(blocks, b)
			owners = append(owners, section.ID)
		}
	}
	return blocks, owners
}

// emptyPage is a core.Page without rows; adding it forces maroto to
// close the current page and start a new one.
type emptyPage struct{}

func (emptyPage) SetConfig(*entity.Config)                       {}
func (emptyPage) GetStructure() *node.Node[core.Structure]       { return node.New(core.Structure{Type: "page"}) }
func (emptyPage) Add(...core.Row) core.Page                      { return emptyPage{} }
func (emptyPage) GetRows() []core.Row                            { return nil }
func (emptyPage) GetNumber() int                                 { return 0 }
func (emptyPage) SetNumber(number int, total int)                 {}
func (emptyPage) Render(provider core.Provider, cell entity.Cell) {}

// blockRow couples a row with its computed height, so pagination knows
// the layout before anything is rendered.
type blockRow struct {
	row    core.Row
	height float64
}

// take returns the rows accumulated so far and resets the builder, so
// each block starts fresh.
func (d *docBuilder) take() []blockRow {
	rows := d.rows
	d.rows = nil
	return rows
}

// docBuilder assembles blocks of rows and measures their heights with
// the same metrics maroto uses at render time.
type docBuilder struct {
	meas         *measurer
	contentWidth float64
	rows         []blockRow
}

// autoText appends a full-width automatic text row.
func (d *docBuilder) autoText(value string, prop props.Text) {
	height := d.meas.textHeight(value, prop, d.contentWidth)
	d.rows = append(d.rows, blockRow{row: text.NewAutoRow(value, prop), height: height})
}

// lineRow appends a fixed-height separator line row.
func (d *docBuilder) lineRow(height float64, prop props.Line) {
	d.rows = append(d.rows, blockRow{row: line.NewRow(height, prop), height: height})
}

// titleDates appends a row with a bold entry title (9 columns) and a
// right-aligned date range (3 columns); its height is the tallest of
// the two cells, as maroto computes for auto rows.
func (d *docBuilder) titleDates(title string, titleProp props.Text, dates string, datesProp props.Text) {
	w9 := d.contentWidth * 9.0 / 12.0
	w3 := d.contentWidth * 3.0 / 12.0
	height := d.meas.textHeight(title, titleProp, w9)
	if h := d.meas.textHeight(dates, datesProp, w3); h > height {
		height = h
	}
	r := row.New().Add(text.NewCol(9, title, titleProp), text.NewCol(3, dates, datesProp))
	d.rows = append(d.rows, blockRow{row: r, height: height})
}

func (d *docBuilder) buildHeader(state models.ResumeState) []blockRow {
	if name := strings.TrimSpace(state.Profile.Name); name != "" {
		d.autoText(name, props.Text{
			Style: fontstyle.Bold, Size: 18, Align: align.Center,
		})
	}
	if headline := strings.TrimSpace(state.Profile.Headline); headline != "" {
		d.autoText(headline, props.Text{
			Style: fontstyle.Italic, Size: 10.5, Align: align.Center,
			Color: gray, Top: 1.5, VerticalPadding: 0.5,
		})
	}
	if contact := contactLine(state); contact != "" {
		d.autoText(contact, props.Text{
			Size: 9.5, Align: align.Center, Color: gray, Top: 2, VerticalPadding: 0.5,
		})
	}
	return d.take()
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

func (d *docBuilder) buildSection(section models.Section) []blockRow {
	d.sectionTitle(section.Title, section.TitleColor, section.LineColor)
	switch models.SectionType(section.Type) {
	case models.SectionSkills:
		d.skills(section)
	case models.SectionExperience:
		d.experience(section)
	case models.SectionEducation:
		d.education(section)
	default:
		d.customItems(section)
	}
	return d.take()
}

func (d *docBuilder) sectionTitle(title, titleColor, lineColor string) {
	title = strings.TrimSpace(title)
	if title == "" {
		return
	}
	d.autoText(strings.ToUpper(title), props.Text{
		Style: fontstyle.Bold, Size: 10.5, Top: 3, Color: hexToColor(titleColor),
	})
	// The line row is 3mm tall with the line drawn at 60% height, which
	// leaves roughly 1mm of air between the title glyphs and the line
	// (auto rows slightly underestimate text descent). SizePercent 100
	// makes the line span the full text column, aligned with the words.
	lineProps := props.Line{SizePercent: 100, OffsetPercent: 60}
	if c := hexToColor(lineColor); c != nil {
		lineProps.Color = c
	} else {
		lineProps.Color = gray
	}
	d.lineRow(3, lineProps)
}

// hexToColor parses an optional "#rrggbb" string; nil falls back to the
// renderer default.
func hexToColor(hex string) *props.Color {
	hex = strings.TrimSpace(hex)
	if len(hex) != 7 || hex[0] != '#' {
		return nil
	}
	var r, g, b int
	if _, err := fmt.Sscanf(hex, "#%02x%02x%02x", &r, &g, &b); err != nil {
		return nil
	}
	return &props.Color{Red: r, Green: g, Blue: b}
}

func (d *docBuilder) skills(section models.Section) {
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
			d.autoText(g.title, props.Text{
				Style: fontstyle.Bold, Size: 9.5, Top: 1.5,
			})
		}
		if len(g.skills) > 0 {
			d.autoText(strings.Join(g.skills, ", "), props.Text{
				Size: 9.5, Top: 1, VerticalPadding: 0.5,
			})
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

func (d *docBuilder) experience(section models.Section) {
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
		d.entryTitle(title, dateRange(it.Data))
		if summary := strings.TrimSpace(str(it.Data, "summary")); summary != "" {
			d.autoText(summary, props.Text{
				Size: 9.5, Color: gray, Top: 1, VerticalPadding: 0.5,
			})
		}
		d.details(it.Details)
	}
}

func (d *docBuilder) education(section models.Section) {
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
		d.entryTitle(title, dateRange(it.Data))
		d.details(it.Details)
	}
}

func (d *docBuilder) customItems(section models.Section) {
	for _, it := range section.Items {
		if !it.Included {
			continue
		}
		if label := strings.TrimSpace(str(it.Data, "label")); label != "" {
			d.autoText(label, props.Text{
				Style: fontstyle.Bold, Size: 10, Top: 2,
			})
		}
		if body := strings.TrimSpace(str(it.Data, "text")); body != "" {
			d.autoText(body, props.Text{
				Size: 9.5, Top: 1, VerticalPadding: 0.5,
			})
		}
		d.details(it.Details)
	}
}

// entryTitle renders a bold entry title with the date range
// right-aligned on the same line. The row height is automatic, so long
// titles wrap instead of overlapping the next row.
func (d *docBuilder) entryTitle(title, dates string) {
	title = strings.TrimSpace(title)
	dates = strings.TrimSpace(dates)
	switch {
	case title == "" && dates == "":
		return
	case title == "":
		d.titleDates("", props.Text{}, dates, props.Text{
			Size: 9, Align: align.Right, Color: gray, Top: 2.5,
		})
	case dates == "":
		d.titleDates(title, props.Text{
			Style: fontstyle.Bold, Size: 10, Top: 2,
		}, "", props.Text{})
	default:
		d.titleDates(
			title, props.Text{Style: fontstyle.Bold, Size: 10, Top: 2},
			dates, props.Text{Size: 9, Align: align.Right, Color: gray, Top: 2.5},
		)
	}
}

func (d *docBuilder) details(details []models.Detail) {
	for _, det := range details {
		if !det.Included || strings.TrimSpace(det.Text) == "" {
			continue
		}
		d.autoText("• "+strings.TrimSpace(det.Text), props.Text{
			Size: 9.5, Left: 3, Top: 0.8, VerticalPadding: 0.5,
		})
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
