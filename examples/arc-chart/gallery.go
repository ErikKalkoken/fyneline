// Package arcchart contains the ArcChart example gallery.
package arcchart

import (
	"image/color"

	"fyne.io/fyne/v2"

	"github.com/nathabonfim59/fyneline"
	"github.com/nathabonfim59/fyneline/examples/internal/gallery"
)

type segment struct {
	Name  string
	Value float64
}

var palette = fyneline.Palette(
	color.NRGBA{R: 62, G: 126, B: 247, A: 255},
	color.NRGBA{R: 240, G: 135, B: 48, A: 255},
	color.NRGBA{R: 47, G: 176, B: 117, A: 255},
	color.NRGBA{R: 139, G: 92, B: 246, A: 255},
)

// Gallery returns four ArcChart configurations in a responsive grid.
func Gallery() fyne.CanvasObject {
	data := []segment{
		{Name: "Search", Value: 34},
		{Name: "Direct", Value: 27},
		{Name: "Social", Value: 22},
		{Name: "Other", Value: 17},
	}
	pie := newArc(data).SetLabels(true).SetPadAngle(1.5).SetCornerRadius(4)
	donut := newArc(data).SetInnerRadius(0.58).SetPadAngle(3).SetCornerRadius(8)
	gauge := newArc([]segment{{Name: "Complete", Value: 72}}).
		SetRange(-120, 120).
		SetInnerRadius(0.72).
		SetMaxValue(100).
		SetCornerRadius(10).
		SetLabels(true)
	semi := newArc(data).
		SetRange(-90, 90).
		SetInnerRadius(0.42).
		SetOuterRadius(0.88).
		SetPadAngle(2)

	return gallery.Grid(
		gallery.Card("Pie", "Labels, rounded corners and small gaps", pie),
		gallery.Card("Doughnut", "58% inner radius and larger padding", donut),
		gallery.Card("Gauge", "Partial range with SetMaxValue(100)", gauge),
		gallery.Card("Semi radial", "Partial range and reduced outer radius", semi),
	)
}

func newArc(data []segment) *fyneline.ArcChart[segment] {
	return fyneline.NewArcChart(data,
		func(s segment) float64 { return s.Value },
		func(s segment) string { return s.Name },
	).SetStyle(func(_ segment, index int) fyneline.ArcStyle {
		return fyneline.ArcStyle{Fill: fyneline.FillStyle{Color: palette(index), Opacity: 1}}
	})
}
