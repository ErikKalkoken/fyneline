// Package barchart contains the BarChart example gallery.
package barchart

import (
	"fmt"
	"image/color"

	"fyne.io/fyne/v2"

	"github.com/nathabonfim59/fyneline"
	"github.com/nathabonfim59/fyneline/examples/internal/gallery"
)

type sale struct {
	Month          string
	Online, Retail float64
}

type traffic struct {
	Day            string
	Web, API, Jobs float64
}

var blue = color.NRGBA{R: 62, G: 126, B: 247, A: 255}
var orange = color.NRGBA{R: 240, G: 135, B: 48, A: 255}
var green = color.NRGBA{R: 47, G: 176, B: 117, A: 255}

// Gallery returns four BarChart configurations in a responsive grid.
func Gallery() fyne.CanvasObject {
	sales := []sale{
		{Month: "Jan", Online: 42, Retail: 31},
		{Month: "Feb", Online: 55, Retail: 38},
		{Month: "Mar", Online: 49, Retail: 44},
		{Month: "Apr", Online: 67, Retail: 48},
		{Month: "May", Online: 73, Retail: 52},
	}
	grouped := fyneline.NewBarChart(sales,
		func(s sale) string { return s.Month },
		fyneline.NewBarSeries("Online", func(s sale) float64 { return s.Online }).WithFill(blue),
		fyneline.NewBarSeries("Retail", func(s sale) float64 { return s.Retail }).WithFill(orange),
	).
		SetSeriesLayout(fyneline.SeriesGroup).
		SetGroupPadding(0.12).
		SetValueAxis(fyneline.NewNumericAxis().WithLabel("Orders").WithDomain(0, 80))

	requests := []traffic{
		{Day: "Mon", Web: 22, API: 31, Jobs: 12},
		{Day: "Tue", Web: 29, API: 38, Jobs: 16},
		{Day: "Wed", Web: 33, API: 35, Jobs: 19},
		{Day: "Thu", Web: 38, API: 44, Jobs: 17},
		{Day: "Fri", Web: 41, API: 48, Jobs: 23},
	}
	stacked := newTrafficChart(requests).
		SetSeriesLayout(fyneline.SeriesStack).
		SetStackPadding(0.08).
		SetValueAxis(fyneline.NewNumericAxis().WithLabel("Requests (k)"))

	divergingData := []traffic{
		{Day: "Mon", Web: 28, API: -12},
		{Day: "Tue", Web: 36, API: -19},
		{Day: "Wed", Web: 31, API: -15},
		{Day: "Thu", Web: 45, API: -24},
	}
	diverging := fyneline.NewBarChart(divergingData,
		func(d traffic) string { return d.Day },
		fyneline.NewBarSeries("Gain", func(d traffic) float64 { return d.Web }).WithFill(green),
		fyneline.NewBarSeries("Loss", func(d traffic) float64 { return d.API }).WithFill(orange),
	).
		SetSeriesLayout(fyneline.SeriesStackDiverging).
		SetOrientation(fyneline.BarHorizontal).
		SetCategoryAxis(fyneline.NewCategoryAxis().WithGrid(true))

	normalized := newTrafficChart(requests).
		SetSeriesLayout(fyneline.SeriesStackExpand).
		SetValueAxis(fyneline.NewNumericAxis().WithDomain(0, 1).WithFormatter(percent))

	return gallery.Grid(
		gallery.Card("Grouped", "SeriesGroup, custom colors and fixed domain", grouped),
		gallery.Card("Stacked", "SeriesStack with stack padding", stacked),
		gallery.Card("Horizontal diverging", "Mixed signs and category grid", diverging),
		gallery.Card("100% stacked", "SeriesStackExpand with percent axis", normalized),
	)
}

func newTrafficChart(data []traffic) *fyneline.BarChart[traffic] {
	return fyneline.NewBarChart(data,
		func(d traffic) string { return d.Day },
		fyneline.NewBarSeries("Web", func(d traffic) float64 { return d.Web }).WithFill(blue),
		fyneline.NewBarSeries("API", func(d traffic) float64 { return d.API }).WithFill(orange),
		fyneline.NewBarSeries("Jobs", func(d traffic) float64 { return d.Jobs }).WithFill(green),
	)
}

func percent(value float64) string {
	return fmt.Sprintf("%.0f%%", value*100)
}
