package fyneline_test

import (
	"image/color"
	"time"

	"github.com/nathabonfim59/fyneline"
)

func ExampleNewBarChart() {
	type Sale struct {
		Month   string
		Online  float64
		InStore float64
	}
	sales := []Sale{
		{Month: "Jan", Online: 42, InStore: 31},
		{Month: "Feb", Online: 55, InStore: 38},
		{Month: "Mar", Online: 49, InStore: 44},
	}

	chart := fyneline.NewBarChart(sales,
		func(s Sale) string { return s.Month },
		fyneline.NewBarSeries("Online", func(s Sale) float64 { return s.Online }),
		fyneline.NewBarSeries("In store", func(s Sale) float64 { return s.InStore }),
	).
		SetSeriesLayout(fyneline.SeriesGroup).
		SetValueAxis(fyneline.NewNumericAxis().WithDomain(0, 60))

	_ = chart // Pass chart directly to Window.SetContent.
}

func ExampleNewAreaChart() {
	type Reading struct {
		Time  time.Time
		Value float64
		Valid bool
	}
	readings := []Reading{
		{Time: time.Unix(0, 0), Value: 12, Valid: true},
		{Time: time.Unix(60, 0), Valid: false},
		{Time: time.Unix(120, 0), Value: 15, Valid: true},
	}

	chart := fyneline.NewAreaChart(readings,
		fyneline.TimeAccessor(func(r Reading) time.Time { return r.Time }),
		fyneline.NewOptionalAreaSeries("Temperature", func(r Reading) (float64, bool) {
			return r.Value, r.Valid
		}),
	).
		SetCurve(fyneline.CurveMonotoneX).
		SetXAxis(fyneline.NewTimeAxis("15:04", time.UTC))

	_ = chart
}

func ExampleNewArcChart() {
	type Segment struct {
		Name  string
		Value float64
	}
	segments := []Segment{{Name: "Used", Value: 72}, {Name: "Free", Value: 28}}
	colors := fyneline.Palette(
		color.NRGBA{R: 62, G: 126, B: 247, A: 255},
		color.NRGBA{R: 215, G: 220, B: 230, A: 255},
	)

	chart := fyneline.NewArcChart(segments,
		func(s Segment) float64 { return s.Value },
		func(s Segment) string { return s.Name },
	).
		SetInnerRadius(0.58).
		SetPadAngle(2).
		SetStyle(func(_ Segment, index int) fyneline.ArcStyle {
			return fyneline.ArcStyle{Fill: fyneline.FillStyle{Color: colors(index), Opacity: 1}}
		})

	_ = chart
}

func ExampleNewSpline() {
	type Point struct{ X, Y float64 }
	points := []Point{{X: 0, Y: 2}, {X: 1, Y: 5}, {X: 2, Y: 3}}

	chart := fyneline.NewSpline(points,
		func(p Point) float64 { return p.X },
		fyneline.NewSplineSeries("Value", func(p Point) float64 { return p.Y }),
	).SetCurve(fyneline.CurveMonotoneX)

	_ = chart
}
