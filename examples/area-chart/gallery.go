// Package areachart contains the AreaChart example gallery.
package areachart

import (
	"image/color"

	"fyne.io/fyne/v2"

	"github.com/nathabonfim59/fyneline"
	"github.com/nathabonfim59/fyneline/examples/internal/gallery"
)

type point struct {
	X         float64
	Primary   float64
	Secondary float64
	Tertiary  float64
	PrimaryOK bool
}

var violet = color.NRGBA{R: 139, G: 92, B: 246, A: 255}
var cyan = color.NRGBA{R: 16, G: 164, B: 190, A: 255}
var rose = color.NRGBA{R: 220, G: 72, B: 103, A: 255}

// Gallery returns four AreaChart configurations in a responsive grid.
func Gallery() fyne.CanvasObject {
	data := []point{
		{X: 0, Primary: 18, Secondary: 11, Tertiary: 7, PrimaryOK: true},
		{X: 1, Primary: 24, Secondary: 16, Tertiary: 9, PrimaryOK: true},
		{X: 2, Primary: 21, Secondary: 18, Tertiary: 13, PrimaryOK: false},
		{X: 3, Primary: 31, Secondary: 20, Tertiary: 11, PrimaryOK: true},
		{X: 4, Primary: 28, Secondary: 24, Tertiary: 15, PrimaryOK: true},
		{X: 5, Primary: 37, Secondary: 27, Tertiary: 17, PrimaryOK: true},
	}
	overlap := newArea(data).
		SetCurve(fyneline.CurveMonotoneX).
		SetSeriesLayout(fyneline.SeriesOverlap)

	stacked := newArea(data).
		SetCurve(fyneline.CurveLinear).
		SetSeriesLayout(fyneline.SeriesStack)

	normalized := newArea(data).
		SetCurve(fyneline.CurveStep).
		SetSeriesLayout(fyneline.SeriesStackExpand).
		SetYAxis(fyneline.NewNumericAxis().WithDomain(0, 1).WithFormatter(func(v float64) string {
			return []string{"0%", "25%", "50%", "75%", "100%"}[min(int(v*4), 4)]
		}))

	gaps := fyneline.NewAreaChart(data,
		func(p point) float64 { return p.X },
		fyneline.NewOptionalAreaSeries("Signal", func(p point) (float64, bool) {
			return p.Primary, p.PrimaryOK
		}).WithStyle(fyneline.AreaStyle{
			Fill:   fyneline.FillStyle{Color: rose, Opacity: 0.35},
			Stroke: fyneline.StrokeStyle{Color: rose, Width: 3},
		}),
	).SetCurve(fyneline.CurveMonotoneX).SetGapPolicy(fyneline.GapBreak)

	return gallery.Grid(
		gallery.Card("Smooth overlap", "CurveMonotoneX and translucent fills", overlap),
		gallery.Card("Stacked", "SeriesStack with a linear boundary", stacked),
		gallery.Card("Normalized steps", "SeriesStackExpand and CurveStep", normalized),
		gallery.Card("Missing data", "Optional accessor with GapBreak", gaps),
	)
}

func newArea(data []point) *fyneline.AreaChart[point] {
	return fyneline.NewAreaChart(data,
		func(p point) float64 { return p.X },
		fyneline.NewAreaSeries("Primary", func(p point) float64 { return p.Primary }).WithStyle(areaStyle(violet)),
		fyneline.NewAreaSeries("Secondary", func(p point) float64 { return p.Secondary }).WithStyle(areaStyle(cyan)),
		fyneline.NewAreaSeries("Tertiary", func(p point) float64 { return p.Tertiary }).WithStyle(areaStyle(rose)),
	)
}

func areaStyle(c color.Color) fyneline.AreaStyle {
	return fyneline.AreaStyle{
		Fill:   fyneline.FillStyle{Color: c, Opacity: 0.28},
		Stroke: fyneline.StrokeStyle{Color: c, Width: 2},
	}
}
