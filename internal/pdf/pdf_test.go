package pdf

import (
	"strings"
	"testing"

	"github.com/johnfercher/maroto/v2/pkg/consts/fontstyle"
	"github.com/johnfercher/maroto/v2/pkg/props"

	"resumator/internal/models"
)

// sampleState builds a resume rich enough to span several pages,
// including an oversized section (taller than a whole page), multiple
// contact sections, skill groups and wrapping text.
func sampleState() models.ResumeState {
	longText := strings.Repeat("Led a cross-functional initiative that delivered measurable results across many teams and stakeholders. ", 12)
	state := models.ResumeState{
		Profile: models.Profile{Name: "Ada Lovelace", Headline: "Engineer & Analyst"},
		Styles:  models.DefaultTextStyles(),
		Sections: []models.Section{
			{
				ID: 1, Type: "contact", Title: "Contact", Visible: true,
				Items: []models.Item{
					{ID: 11, Included: true, Data: map[string]interface{}{"label": "Email", "value": "ada@example.com"}},
					{ID: 12, Included: true, Data: map[string]interface{}{"label": "Phone", "value": "+1 555 0100"}},
					{ID: 13, Included: false, Data: map[string]interface{}{"label": "Fax", "value": "hidden"}},
				},
			},
			{
				ID: 2, Type: "contact", Title: "Socials", Visible: true,
				Items: []models.Item{
					{ID: 21, Included: true, Data: map[string]interface{}{"label": "Web", "value": "ada.dev"}},
				},
			},
			{
				ID: 3, Type: "contact", Title: "Hidden contact", Visible: false,
				Items: []models.Item{
					{ID: 31, Included: true, Data: map[string]interface{}{"value": "should-not-appear"}},
				},
			},
			{
				ID: 4, Type: "skills", Title: "Skills", Visible: true,
				Items: []models.Item{
					{ID: 41, Included: true, Data: map[string]interface{}{"name": "Go", "level": "Expert"}},
					{ID: 42, Included: true, Data: map[string]interface{}{"name": "TypeScript"}},
					{ID: 43, Included: true, Data: map[string]interface{}{"divider": true, "title": "Languages"}},
					{ID: 44, Included: true, Data: map[string]interface{}{"name": "Rust", "level": "Intermediate"}},
					{ID: 45, Included: true, Data: map[string]interface{}{"divider": true, "title": "Hidden group"}},
					{ID: 46, Included: true, Data: map[string]interface{}{"name": "COBOL"}},
					{ID: 47, Included: false, Data: map[string]interface{}{"name": "Fortran"}},
				},
			},
			{
				ID: 5, Type: "experience", Title: "Experience", Visible: true, TitleColor: "#1a2b3c", LineColor: "#0badf0",
				Items: []models.Item{
					{
						ID: 51, Included: true,
						Data: map[string]interface{}{"role": "Senior Engineer", "company": "Analytical Engines Ltd", "start": "2020", "end": "", "summary": longText},
						Details: []models.Detail{
							{ID: 511, Included: true, Text: "Shipped the first version of the engine"},
							{ID: 512, Included: true, Text: longText},
							{ID: 513, Included: false, Text: "not exported"},
						},
					},
					{
						ID: 52, Included: true,
						Data: map[string]interface{}{"role": "Engineer", "company": "Babbage Works", "start": "2017", "end": "2020"},
						Details: []models.Detail{
							{ID: 521, Included: true, Text: "Maintained the difference engine"},
						},
					},
				},
			},
			{
				ID: 6, Type: "education", Title: "Education", Visible: true,
				Items: []models.Item{
					{
						ID: 61, Included: true,
						Data: map[string]interface{}{"degree": "BSc", "field": "Mathematics", "school": "University of London", "start": "2013", "end": "2017"},
					},
				},
			},
			{
				ID: 7, Type: "custom", Title: "Projects", Visible: true,
				Items: []models.Item{
					{ID: 71, Included: true, Data: map[string]interface{}{"label": "Noteworthy", "text": longText}},
				},
			},
			{
				// Taller than a whole A4 page: exercises the oversized-block split.
				ID: 8, Type: "custom", Title: "Long notes", Visible: true,
				Items: []models.Item{makeLongItem(81, 120)},
			},
		},
	}
	return state
}

func makeLongItem(id int64, details int) models.Item {
	det := make([]models.Detail, 0, details)
	for i := 0; i < details; i++ {
		det = append(det, models.Detail{
			ID: id*1000 + int64(i), Included: true,
			Text: strings.Repeat("A fairly long bullet point describing one concrete achievement with numbers and impact. ", 3),
		})
	}
	return models.Item{
		ID: id, Included: true,
		Data:    map[string]interface{}{"label": "Everything", "text": "intro"},
		Details: det,
	}
}

// marotoPages returns the page count and the content-row heights per page
// of the document newMaroto builds, read back from maroto's own structure
// (each page ends with an internal full-height spacer row, which is
// skipped here).
func marotoPages(t *testing.T, state models.ResumeState, format string) (int, [][]float64) {
	t.Helper()
	str := newMaroto(state, format).GetStructure()
	if str.GetData().Type != "maroto" {
		t.Fatalf("unexpected root structure %q", str.GetData().Type)
	}
	pageNodes := str.GetNexts()
	pages := make([][]float64, 0, len(pageNodes))
	for _, p := range pageNodes {
		if p.GetData().Type != "page" {
			t.Fatalf("unexpected child structure %q", p.GetData().Type)
		}
		rowNodes := p.GetNexts()
		heights := make([]float64, 0, len(rowNodes))
		for _, r := range rowNodes {
			if r.GetData().Type != "row" {
				t.Fatalf("unexpected row child %q", r.GetData().Type)
			}
			h, ok := r.GetData().Value.(float64)
			if !ok {
				t.Fatalf("row height is %T, want float64", r.GetData().Value)
			}
			heights = append(heights, h)
		}
		// The last row of every page is maroto's internal spacer.
		if len(heights) > 0 {
			heights = heights[:len(heights)-1]
		}
		pages = append(pages, heights)
	}
	return len(pageNodes), pages
}

// TestLayoutMatchesExport verifies that the pagination Layout reports is
// exactly what the exported document uses: same page count, same breaks,
// and identical per-row heights (measurer parity with maroto).
func TestLayoutMatchesExport(t *testing.T) {
	for _, format := range []string{"a4", "letter", "A4", " LETTER "} {
		state := sampleState()
		breaks, pageCount := Layout(state, format)
		gotPages, gotHeights := marotoPages(t, state, format)

		if pageCount != gotPages {
			t.Errorf("format %q: Layout pageCount=%d, export has %d pages", format, pageCount, gotPages)
		}

		// Flatten the expected chunks in document order.
		blocks, owners := buildBlocks(state, format)
		_, ph := pageDims(format)
		pages := paginate(blocks, owners, ph-marginTop-marginBottom)
		if len(pages) != gotPages {
			t.Fatalf("format %q: paginate produced %d pages, export has %d", format, len(pages), gotPages)
		}
		var wantBreaks []int64
		var wantHeights [][]float64
		for pi, page := range pages {
			if pi > 0 {
				wantBreaks = append(wantBreaks, page[0].owner)
			}
			hs := []float64{}
			for _, c := range page {
				for _, r := range c.rows {
					hs = append(hs, r.height)
				}
			}
			wantHeights = append(wantHeights, hs)
		}
		if len(wantBreaks) != len(breaks) {
			t.Errorf("format %q: breaks %v, want %v", format, breaks, wantBreaks)
		} else {
			for i := range breaks {
				if breaks[i] != wantBreaks[i] {
					t.Errorf("format %q: breaks %v, want %v", format, breaks, wantBreaks)
					break
				}
			}
		}
		for pi := range wantHeights {
			if len(wantHeights[pi]) != len(gotHeights[pi]) {
				t.Errorf("format %q page %d: %d rows, export has %d", format, pi, len(wantHeights[pi]), len(gotHeights[pi]))
				continue
			}
			for ri := range wantHeights[pi] {
				if diff := wantHeights[pi][ri] - gotHeights[pi][ri]; diff < -1e-6 || diff > 1e-6 {
					t.Errorf("format %q page %d row %d: measured height %v, maroto rendered %v", format, pi, ri, wantHeights[pi][ri], gotHeights[pi][ri])
				}
			}
		}
	}
}

// TestOversizedSectionSplit verifies that a section taller than a page is
// split at row boundaries and that Layout counts every resulting page.
func TestOversizedSectionSplit(t *testing.T) {
	state := sampleState()
	breaks, pageCount := Layout(state, "a4")
	if pageCount < 3 {
		t.Fatalf("expected the long-notes section to add pages, got pageCount=%d breaks=%v", pageCount, breaks)
	}
	// The oversized section (id 8) must appear as a page break, and the
	// export must agree.
	found := false
	for _, b := range breaks {
		if b == 8 {
			found = true
		}
	}
	if !found {
		t.Errorf("expected section 8 to start a page, breaks=%v", breaks)
	}
	gotPages, _ := marotoPages(t, state, "a4")
	if gotPages != pageCount {
		t.Errorf("export has %d pages, Layout reports %d", gotPages, pageCount)
	}
}

// TestContactLines mirrors the preview: every visible contact section
// contributes one header line, hidden sections and excluded items do not.
func TestContactLines(t *testing.T) {
	lines := contactLines(sampleState())
	want := []string{"ada@example.com   •   +1 555 0100", "ada.dev"}
	if len(lines) != len(want) {
		t.Fatalf("contactLines=%v, want %v", lines, want)
	}
	for i := range want {
		if lines[i] != want[i] {
			t.Errorf("contactLines[%d]=%q, want %q", i, lines[i], want[i])
		}
	}
}

// TestMeasurerMatchesMaroto checks the measurer against maroto's rendered
// row heights for tricky inputs: wrapping text, empty strings and props
// with unset size (maroto promotes size 0 to its default font size).
func TestMeasurerMatchesMaroto(t *testing.T) {
	state := sampleState()
	blocks, _ := buildBlocks(state, "a4")
	_, gotHeights := marotoPages(t, state, "a4")

	// Flatten measured heights in document order.
	var want []float64
	for _, b := range blocks {
		for _, r := range b {
			want = append(want, r.height)
		}
	}
	var got []float64
	for _, page := range gotHeights {
		got = append(got, page...)
	}
	if len(want) != len(got) {
		t.Fatalf("measured %d rows, maroto rendered %d", len(want), len(got))
	}
	for i := range want {
		if diff := want[i] - got[i]; diff < -1e-6 || diff > 1e-6 {
			t.Errorf("row %d: measured %v, maroto %v", i, want[i], got[i])
		}
	}
}

// TestMeasurerZeroSizeProp ensures the measurer applies the same default
// font size maroto's MakeValid applies, so a zero-size prop measures the
// same as it renders.
func TestMeasurerZeroSizeProp(t *testing.T) {
	m := acquireMeasurer()
	defer releaseMeasurer(m)
	prop := props.Text{Style: fontstyle.Bold} // Size unset: maroto renders at 10pt
	got := m.textHeight("hello world", prop, 100)
	want := m.textHeight("hello world", props.Text{Style: fontstyle.Bold, Size: defaultFontSize}, 100)
	if got != want {
		t.Errorf("zero-size prop measured %v, explicit default measured %v", got, want)
	}
	// Empty text occupies zero lines, exactly like maroto.
	if n := m.linesQuantity("", props.Text{Size: 9}, 100); n != 0 {
		t.Errorf("linesQuantity(\"\") = %d, want 0", n)
	}
}

// TestSanitizeStyles guards against zero/negative sizes reaching the
// renderer (maroto would silently change size 0 to 10, desynchronising
// pagination).
func TestSanitizeStyles(t *testing.T) {
	def := models.DefaultTextStyles()
	s := sanitizeStyles(models.TextStyles{}) // all zero
	if s.SectionTitle.Size != def.SectionTitle.Size ||
		s.EntryTitle.Size != def.EntryTitle.Size ||
		s.EntryDates.Size != def.EntryDates.Size ||
		s.Body.Size != def.Body.Size {
		t.Errorf("sanitizeStyles(zero) sizes = %+v, want defaults %+v", s, def)
	}
	s = sanitizeStyles(models.TextStyles{
		SectionTitle: models.TextStyle{Size: -3},
		Body:         models.TextStyle{Size: 11},
	})
	if s.SectionTitle.Size != def.SectionTitle.Size {
		t.Errorf("negative section title size = %v, want %v", s.SectionTitle.Size, def.SectionTitle.Size)
	}
	if s.Body.Size != 11 {
		t.Errorf("positive body size overwritten: %v", s.Body.Size)
	}
}

// TestGenerateProducesPDF runs the full export and checks the PDF magic
// bytes and page count via the structure.
func TestGenerateProducesPDF(t *testing.T) {
	state := sampleState()
	data, err := Generate(state, "a4")
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if len(data) < 5 || string(data[:5]) != "%PDF-" {
		t.Errorf("output does not look like a PDF: %q", data[:min(5, len(data))])
	}
	_, pageCount := Layout(state, "a4")
	gotPages, _ := marotoPages(t, state, "a4")
	if pageCount != gotPages {
		t.Errorf("Layout pageCount=%d but export has %d pages", pageCount, gotPages)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
