// Package pdf renders a resume state to a PDF document. The layout
// mirrors the frontend preview: name/headline/contact centered at the
// top, then one titled block per section. All text rows use automatic
// heights so wrapped text grows its row instead of overlapping.
//
// Sections are never split across pages: the document is built as a
// sequence of blocks whose heights are known in advance (replicating
// maroto's text metrics), and a page break is forced before any section
// that would not fit on the current page. A section taller than a whole
// page cannot be kept together and is split at row boundaries instead.
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
	// Grid granularity for column sizes. A fine grid lets the skills
	// line size its bold title column to the measured text width.
	maxGrid = 1200
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
	d, err := newMaroto(state, format).Generate()
	if err != nil {
		return nil, err
	}
	return d.GetBytes(), nil
}

// newMaroto builds the maroto document for the state, forcing page breaks
// exactly where paginate decided, so the exported pages match Layout (and
// the preview).
func newMaroto(state models.ResumeState, format string) core.Maroto {
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
		WithMaxGridSize(maxGrid).
		Build()
	m := maroto.New(cfg)

	blocks, owners := buildBlocks(state, format)
	for pi, page := range paginate(blocks, owners, pageHeight-marginTop-marginBottom) {
		if pi > 0 {
			// Force a page break exactly where the layout engine decided,
			// so the exported pages match Layout (and the preview).
			m.AddPages(emptyPage{})
		}
		for _, c := range page {
			for _, r := range c.rows {
				m.AddRows(r.row)
			}
		}
	}
	return m
}

// Layout returns the pagination the PDF export will use: the section
// IDs that start pages after the first, and the total page count. The
// preview uses this so its pages match the export exactly.
func Layout(state models.ResumeState, format string) (breaks []int64, pageCount int) {
	_, pageHeight := pageDims(format)
	blocks, owners := buildBlocks(state, format)
	pages := paginate(blocks, owners, pageHeight-marginTop-marginBottom)
	for pi, page := range pages {
		if pi > 0 {
			breaks = append(breaks, page[0].owner)
		}
	}
	return breaks, len(pages)
}

// chunk is a run of rows that stays together on one page. owner is the
// section ID owning the chunk (0 for the header).
type chunk struct {
	rows  []blockRow
	owner int64
}

// paginate distributes blocks over pages of the given content height.
// A block that fits on a single page is never split: if it does not fit
// in the remaining space it starts the next page, so sections are never
// cut in half. A block taller than a whole page cannot be kept together
// and is split at row boundaries instead, starting on a fresh page.
func paginate(blocks [][]blockRow, owners []int64, pageHeight float64) [][]chunk {
	var pages [][]chunk
	var current []chunk
	used := 0.0
	flush := func() {
		if len(current) > 0 {
			pages = append(pages, current)
			current = nil
			used = 0
		}
	}
	for i, block := range blocks {
		total := 0.0
		for _, r := range block {
			total += r.height
		}
		if total <= pageHeight {
			if len(current) > 0 && used+total > pageHeight {
				flush()
			}
			current = append(current, chunk{rows: block, owner: owners[i]})
			used += total
			continue
		}
		// Oversized block: start it on a fresh page and split it at row
		// boundaries; the last piece stays open so later blocks can join it.
		flush()
		var rows []blockRow
		h := 0.0
		for _, r := range block {
			if len(rows) > 0 && h+r.height > pageHeight {
				pages = append(pages, []chunk{{rows: rows, owner: owners[i]}})
				rows = nil
				h = 0
			}
			rows = append(rows, r)
			h += r.height
		}
		if len(rows) > 0 {
			current = []chunk{{rows: rows, owner: owners[i]}}
			used = h
		}
	}
	flush()
	return pages
}

// buildBlocks assembles the header block and one block per visible
// section, with row heights known in advance. owners holds the section
// ID owning each block (0 for the header).
func buildBlocks(state models.ResumeState, format string) ([][]blockRow, []int64) {
	pageWidth, _ := pageDims(format)
	meas := acquireMeasurer()
	defer releaseMeasurer(meas)
	doc := &docBuilder{
		meas:         meas,
		contentWidth: pageWidth - marginLeft - marginRight,
		styles:       sanitizeStyles(state.Styles),
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

// sanitizeStyles replaces non-positive sizes with the defaults, so a
// zeroed or negative style can never reach the renderer (maroto would
// silently promote size 0 to its own default, desynchronising the
// measurer and the export).
func sanitizeStyles(s models.TextStyles) models.TextStyles {
	def := models.DefaultTextStyles()
	if s.SectionTitle.Size <= 0 {
		s.SectionTitle.Size = def.SectionTitle.Size
	}
	if s.EntryTitle.Size <= 0 {
		s.EntryTitle.Size = def.EntryTitle.Size
	}
	if s.EntryDates.Size <= 0 {
		s.EntryDates.Size = def.EntryDates.Size
	}
	if s.Body.Size <= 0 {
		s.Body.Size = def.Body.Size
	}
	return s
}

// emptyPage is a core.Page without rows; adding it forces maroto to
// close the current page and start a new one.
type emptyPage struct{}

func (emptyPage) SetConfig(*entity.Config) {}
func (emptyPage) GetStructure() *node.Node[core.Structure] {
	return node.New(core.Structure{Type: "page"})
}
func (emptyPage) Add(...core.Row) core.Page                       { return emptyPage{} }
func (emptyPage) GetRows() []core.Row                             { return nil }
func (emptyPage) GetNumber() int                                  { return 0 }
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
	styles       models.TextStyles
	rows         []blockRow
}

// fontStyleOf maps bold/italic flags to a maroto font style.
func fontStyleOf(bold, italic bool) fontstyle.Type {
	switch {
	case bold && italic:
		return fontstyle.BoldItalic
	case bold:
		return fontstyle.Bold
	case italic:
		return fontstyle.Italic
	default:
		return fontstyle.Normal
	}
}

// bodyText returns the base body typography; callers add spacing and color.
func (d *docBuilder) bodyText() props.Text {
	return props.Text{
		Size:  d.styles.Body.Size,
		Style: fontStyleOf(d.styles.Body.Bold, d.styles.Body.Italic),
	}
}

// entryTitleText returns the base entry-title typography.
func (d *docBuilder) entryTitleText() props.Text {
	return props.Text{
		Size:  d.styles.EntryTitle.Size,
		Style: fontStyleOf(d.styles.EntryTitle.Bold, d.styles.EntryTitle.Italic),
	}
}

// entryDatesText returns the base entry-dates typography.
func (d *docBuilder) entryDatesText() props.Text {
	return props.Text{
		Size:  d.styles.EntryDates.Size,
		Style: fontStyleOf(d.styles.EntryDates.Bold, d.styles.EntryDates.Italic),
	}
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
	w9 := d.contentWidth * 900.0 / maxGrid
	w3 := d.contentWidth * 300.0 / maxGrid
	height := d.meas.textHeight(title, titleProp, w9)
	if h := d.meas.textHeight(dates, datesProp, w3); h > height {
		height = h
	}
	r := row.New().Add(
		text.NewCol(900, title, titleProp),
		text.NewCol(300, dates, datesProp),
	)
	d.rows = append(d.rows, blockRow{row: r, height: height})
}

func (d *docBuilder) buildHeader(state models.ResumeState) []blockRow {
	name, headline := contactHeader(state)
	if name = strings.TrimSpace(name); name != "" {
		d.autoText(name, props.Text{
			Style: fontstyle.Bold, Size: 18, Align: align.Center,
		})
	}
	if headline = strings.TrimSpace(headline); headline != "" {
		d.autoText(headline, props.Text{
			Style: fontstyle.Italic, Size: 10.5, Align: align.Center,
			Color: gray, Top: 1.5, VerticalPadding: 0.5,
		})
	}
	for i, contact := range contactLines(state) {
		top := 2.0
		if i > 0 {
			top = 0.5
		}
		d.autoText(contact, props.Text{
			Size: 9.5, Align: align.Center, Color: gray, Top: top, VerticalPadding: 0.5,
		})
	}
	return d.take()
}

// contactHeader returns the name and headline of the first visible
// contact section, mirroring the preview header.
func contactHeader(state models.ResumeState) (name, headline string) {
	for _, section := range state.Sections {
		if section.Type == string(models.SectionContact) && section.Visible {
			return section.Name, section.Headline
		}
	}
	return "", ""
}

// contactLines joins the values of every visible contact section, one
// line per section, mirroring the preview header.
func contactLines(state models.ResumeState) []string {
	var lines []string
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
			lines = append(lines, strings.Join(parts, "   •   "))
		}
	}
	return lines
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
		Style: fontStyleOf(d.styles.SectionTitle.Bold, d.styles.SectionTitle.Italic),
		Size:  d.styles.SectionTitle.Size,
		Top:   3, Color: hexToColor(titleColor),
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
	// after a marker belong to its group. Every group renders as one row
	// with the title right-aligned in a shared column (sized to the
	// widest title) and the skills left-aligned, so titles line up on
	// their right edge and skills share a left edge. A group is hidden
	// entirely when its marker is not included in the preset.
	type skillGroup struct {
		title   string
		skills  []string
		visible bool
	}
	groups := []skillGroup{{visible: true}}
	for _, it := range section.Items {
		if boolVal(it.Data, "divider") {
			groups = append(groups, skillGroup{
				title:   strings.TrimSpace(str(it.Data, "title")),
				visible: it.Included,
			})
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

	var visible []skillGroup
	for _, g := range groups {
		if g.visible && (g.title != "" || len(g.skills) > 0) {
			visible = append(visible, g)
		}
	}
	if len(visible) == 0 {
		return
	}

	titleProp := props.Text{Style: fontstyle.Bold, Size: d.styles.Body.Size, Top: 1, Align: align.Right}
	skillsProp := d.bodyText()
	skillsProp.Top = 1
	skillsProp.VerticalPadding = 0.5

	titleWidth := 0.0
	for _, g := range visible {
		if g.title != "" {
			if w := d.meas.textWidth(g.title, titleProp); w > titleWidth {
				titleWidth = w
			}
		}
	}
	titleUnits := 0
	if titleWidth > 0 {
		titleUnits = int(math.Ceil(titleWidth / d.contentWidth * maxGrid))
		if titleUnits < 30 {
			titleUnits = 30
		}
		if titleUnits > maxGrid-60 {
			titleUnits = maxGrid - 60
		}
		// The gap between a title and its skills is one space wide, applied
		// as the skills column's left padding (the title is right-aligned
		// against its column edge). Only needed when a title column exists.
		skillsProp.Left = d.meas.textWidth(" ", skillsProp)
	}
	skillsUnits := maxGrid - titleUnits
	wTitle := d.contentWidth * float64(titleUnits) / maxGrid
	wSkills := d.contentWidth * float64(skillsUnits) / maxGrid

	for _, g := range visible {
		joined := strings.Join(g.skills, ", ")
		height := d.meas.textHeight(joined, skillsProp, wSkills)
		if g.title != "" {
			if h := d.meas.textHeight(g.title, titleProp, wTitle); h > height {
				height = h
			}
		}
		var r core.Row
		if titleUnits == 0 {
			r = row.New().Add(text.NewCol(maxGrid, joined, skillsProp))
		} else {
			r = row.New().Add(
				text.NewCol(titleUnits, g.title, titleProp),
				text.NewCol(skillsUnits, joined, skillsProp),
			)
		}
		d.rows = append(d.rows, blockRow{row: r, height: height})
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
			prop := d.bodyText()
			prop.Color = gray
			prop.Top = 1
			prop.VerticalPadding = 0.5
			d.autoText(summary, prop)
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
			prop := d.entryTitleText()
			prop.Top = 2
			d.autoText(label, prop)
		}
		if body := strings.TrimSpace(str(it.Data, "text")); body != "" {
			prop := d.bodyText()
			prop.Top = 1
			prop.VerticalPadding = 0.5
			d.autoText(body, prop)
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
	if title == "" && dates == "" {
		return
	}
	titleProp := d.entryTitleText()
	titleProp.Top = 2
	datesProp := d.entryDatesText()
	datesProp.Align = align.Right
	datesProp.Color = gray
	datesProp.Top = 2.5
	d.titleDates(title, titleProp, dates, datesProp)
}

func (d *docBuilder) details(details []models.Detail) {
	prop := d.bodyText()
	prop.Left = 3
	prop.Top = 0.8
	prop.VerticalPadding = 0.5
	for _, det := range details {
		if !det.Included || strings.TrimSpace(det.Text) == "" {
			continue
		}
		d.autoText("• "+strings.TrimSpace(det.Text), prop)
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
