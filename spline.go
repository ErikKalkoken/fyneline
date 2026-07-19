package fyneline

import (
	"math"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"
)

// Spline displays one or more numerical series as connected lines. Although
// LayerChart exposes Spline as a composable mark, Fyneline makes it a complete,
// responsive Fyne widget with axes and grid configuration.
type Spline[T any] struct {
	widget.BaseWidget

	data        []T
	x           func(T) float64
	series      []SplineSeries[T]
	xAxis       NumericAxis
	yAxis       NumericAxis
	curve       Curve
	gapPolicy   GapPolicy
	padding     Insets
	minimumSize fyne.Size
}

// NewSpline constructs a spline from data, an x accessor, and one or more
// numerical series. The default curve is CurveLinear, matching LayerChart.
func NewSpline[T any, X Number](data []T, x func(T) X, series ...SplineSeries[T]) *Spline[T] {
	chart := &Spline[T]{
		data:        append([]T(nil), data...),
		x:           func(datum T) float64 { return float64(x(datum)) },
		series:      append([]SplineSeries[T](nil), series...),
		xAxis:       NewNumericAxis(),
		yAxis:       NewNumericAxis(),
		curve:       CurveLinear,
		gapPolicy:   GapBreak,
		minimumSize: fyne.NewSize(240, 160),
	}
	chart.ExtendBaseWidget(chart)
	return chart
}

// SetData replaces the chart data and refreshes the widget.
func (c *Spline[T]) SetData(data []T) *Spline[T] {
	c.data = append([]T(nil), data...)
	c.Refresh()
	return c
}

// SetSeries replaces the displayed series and refreshes the widget.
func (c *Spline[T]) SetSeries(series ...SplineSeries[T]) *Spline[T] {
	c.series = append([]SplineSeries[T](nil), series...)
	c.Refresh()
	return c
}

// SetXAxis replaces the x-axis configuration.
func (c *Spline[T]) SetXAxis(axis NumericAxis) *Spline[T] {
	c.xAxis = axis
	c.Refresh()
	return c
}

// SetYAxis replaces the y-axis configuration.
func (c *Spline[T]) SetYAxis(axis NumericAxis) *Spline[T] {
	c.yAxis = axis
	c.Refresh()
	return c
}

// SetCurve selects linear, monotone, or stepped interpolation.
func (c *Spline[T]) SetCurve(curve Curve) *Spline[T] {
	c.curve = curve
	c.Refresh()
	return c
}

// SetGapPolicy controls how optional series values are connected.
func (c *Spline[T]) SetGapPolicy(policy GapPolicy) *Spline[T] {
	c.gapPolicy = policy
	c.Refresh()
	return c
}

// SetPadding sets logical-pixel insets around the chart.
func (c *Spline[T]) SetPadding(padding Insets) *Spline[T] {
	c.padding = padding
	c.Refresh()
	return c
}

// CreateRenderer implements fyne.Widget.
func (c *Spline[T]) CreateRenderer() fyne.WidgetRenderer {
	return newChartRenderer(c)
}

func (c *Spline[T]) chartMinSize() fyne.Size { return c.minimumSize }

func (c *Spline[T]) chartObjects(size fyne.Size) []fyne.CanvasObject {
	xValues, yValues, defined, xDomain, yDomain := c.values()
	plot := cartesianPlot(size, c.padding, c.xAxis.visible, c.yAxis.visible)
	plot = fitNumericAxisMargin(c, plot, c.xAxis, xDomain, true)
	plot = fitNumericAxisMargin(c, plot, c.yAxis, yDomain, false)
	objects := renderNumericAxes(c, plot, c.xAxis, c.yAxis, xDomain, yDomain)
	for seriesIndex, series := range c.series {
		segments := pointSegments(xValues, yValues[seriesIndex], defined[seriesIndex], c.gapPolicy, plot, xDomain, yDomain)
		stroke := seriesColor(c, seriesIndex, series.style.Stroke.Color)
		width := series.style.Stroke.Width
		if width <= 0 {
			width = 2
		}
		for _, segment := range segments {
			points := curvePoints(segment, c.curve)
			for index := 1; index < len(points); index++ {
				line := canvas.NewLine(stroke)
				line.StrokeWidth = width
				line.Position1 = points[index-1]
				line.Position2 = points[index]
				objects = append(objects, line)
			}
		}
	}
	return objects
}

func (c *Spline[T]) values() ([]float64, [][]float64, [][]bool, Domain, Domain) {
	xValues := make([]float64, len(c.data))
	yValues := make([][]float64, len(c.series))
	defined := make([][]bool, len(c.series))
	xMin, xMax := math.Inf(1), math.Inf(-1)
	yMin, yMax := math.Inf(1), math.Inf(-1)
	for index, datum := range c.data {
		xValue := c.x(datum)
		xValues[index] = xValue
		if finite(xValue) {
			xMin, xMax = min(xMin, xValue), max(xMax, xValue)
		}
	}
	for seriesIndex, series := range c.series {
		yValues[seriesIndex] = make([]float64, len(c.data))
		defined[seriesIndex] = make([]bool, len(c.data))
		for index, datum := range c.data {
			value, ok := series.value(datum)
			ok = ok && finite(value) && finite(xValues[index])
			if !ok && c.gapPolicy == GapZero && finite(xValues[index]) {
				value, ok = 0, true
			}
			yValues[seriesIndex][index] = value
			defined[seriesIndex][index] = ok
			if ok {
				yMin, yMax = min(yMin, value), max(yMax, value)
			}
		}
	}
	return xValues, yValues, defined,
		numericDomain(c.xAxis, xMin, xMax, false),
		numericDomain(c.yAxis, yMin, yMax, false)
}

func pointSegments(xValues, yValues []float64, defined []bool, policy GapPolicy, plot plotRect, xDomain, yDomain Domain) [][]fyne.Position {
	segments := make([][]fyne.Position, 0, 2)
	current := make([]fyne.Position, 0, len(xValues))
	for index, xValue := range xValues {
		if !defined[index] {
			if policy == GapBreak && len(current) > 0 {
				segments = append(segments, current)
				current = nil
			}
			continue
		}
		current = append(current, fyne.NewPos(
			scale(xValue, xDomain, plot.left, plot.right),
			scale(yValues[index], yDomain, plot.bottom, plot.top),
		))
	}
	if len(current) > 0 {
		segments = append(segments, current)
	}
	return segments
}
