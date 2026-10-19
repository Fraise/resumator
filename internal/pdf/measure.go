package pdf

import (
	"strings"

	"github.com/johnfercher/maroto/v2/pkg/consts/fontfamily"
	"github.com/johnfercher/maroto/v2/pkg/props"
	"github.com/phpdave11/gofpdf"
)

// fontScaleFactor converts points to millimetres, matching the scale
// factor gofpdf uses for millimetre-unit documents (maroto uses the same).
const fontScaleFactor = 72.0 / 25.4

// measurer replicates maroto's automatic text height calculation so that
// section heights are known before rows are added to the document. This
// makes it possible to keep whole sections on one page. It must use the
// same gofpdf build (identical font metrics) and the same wrapping
// algorithm as the maroto provider.
type measurer struct {
	pdf        *gofpdf.Fpdf
	translator func(string) string
}

func newMeasurer() *measurer {
	pdf := gofpdf.New("P", "mm", "A4", "")
	return &measurer{
		pdf:        pdf,
		translator: pdf.UnicodeTranslatorFromDescriptor(""),
	}
}

// textWidth measures the rendered width of a string in millimetres.
func (m *measurer) textWidth(value string, prop props.Text) float64 {
	family := prop.Family
	if family == "" {
		family = fontfamily.Arial
	}
	m.pdf.SetFont(family, string(prop.Style), prop.Size)
	return m.pdf.GetStringWidth(m.translator(value))
}

// textHeight mirrors maroto's text component GetHeight: the wrapped line
// count, font height and the top/bottom/vertical paddings.
func (m *measurer) textHeight(value string, prop props.Text, colWidth float64) float64 {
	family := prop.Family
	if family == "" {
		family = fontfamily.Arial
	}
	fontHeight := prop.Size / fontScaleFactor
	lines := m.linesQuantity(value, family, prop, colWidth-prop.Left-prop.Right)
	textHeight := float64(lines)*fontHeight + float64(lines-1)*prop.VerticalPadding
	return textHeight + prop.Top + prop.Bottom
}

// linesQuantity mirrors the maroto provider's GetLinesQuantity with the
// default space break strategy: greedy word wrapping measured with real
// font metrics.
func (m *measurer) linesQuantity(value, family string, prop props.Text, colWidth float64) int {
	m.pdf.SetFont(family, string(prop.Style), prop.Size)

	// Core fonts are rendered with a cp1252 translation, so widths must
	// be measured on the translated string, as the provider does.
	translated := m.translator(value)
	if translated == "" {
		return 1
	}

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
	if lines == 0 {
		lines = 1
	}
	return lines
}
