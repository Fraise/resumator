package pdf

import (
	"strings"
	"sync"

	"github.com/johnfercher/maroto/v2/pkg/consts/fontfamily"
	"github.com/johnfercher/maroto/v2/pkg/consts/fontstyle"
	"github.com/johnfercher/maroto/v2/pkg/props"
	"github.com/phpdave11/gofpdf"
)

// fontScaleFactor converts points to millimetres, matching the scale
// factor gofpdf uses for millimetre-unit documents (maroto uses the same).
const fontScaleFactor = 72.0 / 25.4

// defaultFontSize mirrors maroto's config default font size: props with a
// zero size are promoted to it by Text.MakeValid at render time.
const defaultFontSize = 10.0

// measurer replicates maroto's automatic text height calculation so that
// section heights are known before rows are added to the document. This
// makes it possible to keep whole sections on one page. It must use the
// same gofpdf build (identical font metrics) and the same wrapping
// algorithm as the maroto provider.
//
// A measurer is not safe for concurrent use; acquire one from the pool.
type measurer struct {
	pdf        *gofpdf.Fpdf
	translator func(string) string
}

// measurerPool recycles gofpdf instances across Layout/Generate calls.
// The Wails frontend calls GetPageLayout on every state change, and each
// call would otherwise build a fresh gofpdf document.
var measurerPool = sync.Pool{
	New: func() any { return newMeasurer() },
}

func acquireMeasurer() *measurer {
	return measurerPool.Get().(*measurer)
}

func releaseMeasurer(m *measurer) {
	measurerPool.Put(m)
}

func newMeasurer() *measurer {
	pdf := gofpdf.New("P", "mm", "A4", "")
	return &measurer{
		pdf:        pdf,
		translator: pdf.UnicodeTranslatorFromDescriptor(""),
	}
}

// normalize applies the same defaults maroto's Text.MakeValid applies at
// render time, so measured metrics match the rendered ones even for props
// that leave family, style or size unset.
func normalize(prop props.Text) props.Text {
	if prop.Family == "" {
		prop.Family = fontfamily.Arial
	}
	if prop.Style == "" {
		prop.Style = fontstyle.Normal
	}
	if prop.Size == 0 {
		prop.Size = defaultFontSize
	}
	return prop
}

// textWidth measures the rendered width of a string in millimetres.
func (m *measurer) textWidth(value string, prop props.Text) float64 {
	prop = normalize(prop)
	m.pdf.SetFont(prop.Family, string(prop.Style), prop.Size)
	return m.pdf.GetStringWidth(m.translator(value))
}

// textHeight mirrors maroto's text component GetHeight: the wrapped line
// count, font height and the top/bottom/vertical paddings. Like maroto, an
// empty string occupies zero lines.
func (m *measurer) textHeight(value string, prop props.Text, colWidth float64) float64 {
	prop = normalize(prop)
	fontHeight := prop.Size / fontScaleFactor
	lines := m.linesQuantity(value, prop, colWidth-prop.Left-prop.Right)
	textHeight := float64(lines)*fontHeight + float64(lines-1)*prop.VerticalPadding
	return textHeight + prop.Top + prop.Bottom
}

// linesQuantity mirrors the maroto provider's GetLinesQuantity with the
// default space break strategy: greedy word wrapping measured with real
// font metrics. It returns 0 for an empty string, exactly like maroto.
func (m *measurer) linesQuantity(value string, prop props.Text, colWidth float64) int {
	m.pdf.SetFont(prop.Family, string(prop.Style), prop.Size)

	// Core fonts are rendered with a cp1252 translation, so widths must
	// be measured on the translated string, as the provider does.
	translated := m.translator(value)

	currentSize := 0.0
	lines := 0
	hasContent := false
	for _, word := range strings.Split(translated, " ") {
		if word == "" {
			continue
		}
		piece := word
		if hasContent {
			piece = " " + word
		}
		width := m.pdf.GetStringWidth(piece)
		if currentSize+width <= colWidth {
			if !hasContent {
				lines++
				hasContent = true
			}
			currentSize += width
		} else {
			lines++
			hasContent = true
			currentSize = m.pdf.GetStringWidth(word)
		}
	}
	return lines
}
