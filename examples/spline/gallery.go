// Package spline contains the Spline example gallery.
package spline

import (
	"image/color"

	"fyne.io/fyne/v2"

	"github.com/nathabonfim59/fyneline"
	"github.com/nathabonfim59/fyneline/examples/internal/gallery"
)

type point struct {
	X, Alpha, Beta float64
	Valid          bool
}

var indigo = color.NRGBA{R: 79, G: 70, B: 229, A: 255}
var amber = color.NRGBA{R: 245, G: 158, B: 11, A: 255}

// Gallery returns four Spline configurations in a responsive grid.
func Gallery() fyne.CanvasObject {
	data := []point{
		{X: 0, Alpha: 12, Beta: 26, Valid: true},
		{X: 1, Alpha: 24, Beta: 22, Valid: true},
		{X: 2, Alpha: 19, Beta: 31, Valid: false},
		{X: 3, Alpha: 34, Beta: 27, Valid: true},
		{X: 4, Alpha: 29, Beta: 38, Valid: true},
		{X: 5, Alpha: 43, Beta: 34, Valid: true},
	}
	linear := newSpline(data).SetCurve(fyneline.CurveLinear)
	monotone := newSpline(data).SetCurve(fyneline.CurveMonotoneX)
	steps := newSpline(data).
		SetCurve(fyneline.CurveStepAfter).
		SetXAxis(fyneline.NewNumericAxis().WithGrid(false))
	gaps := fyneline.NewSpline(data,
		func(p point) float64 { return p.X },
		fyneline.NewOptionalSplineSeries("Alpha", func(p point) (float64, bool) {
			return p.Alpha, p.Valid
		}).WithStyle(fyneline.SplineStyle{Stroke: fyneline.StrokeStyle{Color: indigo, Width: 4}}),
	).SetCurve(fyneline.CurveMonotoneX).SetGapPolicy(fyneline.GapBreak)

	return gallery.Grid(
		gallery.Card("Linear", "Two series with CurveLinear", linear),
		gallery.Card("Monotone", "CurveMonotoneX without local extrema", monotone),
		gallery.Card("Step after", "CurveStepAfter and reduced grid", steps),
		gallery.Card("Broken signal", "Optional accessor with GapBreak", gaps),
	)
}

func newSpline(data []point) *fyneline.Spline[point] {
	return fyneline.NewSpline(data,
		func(p point) float64 { return p.X },
		fyneline.NewSplineSeries("Alpha", func(p point) float64 { return p.Alpha }).WithStyle(
			fyneline.SplineStyle{Stroke: fyneline.StrokeStyle{Color: indigo, Width: 3}},
		),
		fyneline.NewSplineSeries("Beta", func(p point) float64 { return p.Beta }).WithStyle(
			fyneline.SplineStyle{Stroke: fyneline.StrokeStyle{Color: amber, Width: 3}},
		),
	)
}
