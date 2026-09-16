package fyneline

import (
	"math"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"
)

// AreaChart displays quantitative series as filled areas over a continuous x
// axis. It defaults to overlapping series with a zero baseline.
type AreaChart[T any] struct {
	widget.BaseWidget

	data         []T
	x            func(T) float64
	series       []AreaSeries[T]
	style        func(AreaStyle, int) AreaStyle
	xAxis        NumericAxis
	yAxis        NumericAxis
	curve        Curve
	gapPolicy    GapPolicy
	seriesLayout SeriesLayout
	padding      Insets
	minimumSize  fyne.Size
}

// NewAreaChart constructs an area chart from data, an x accessor, and one or
// more numerical series.
func NewAreaChart[T any, X Number](data []T, x func(T) X, series ...AreaSeries[T]) *AreaChart[T] {
	chart := &AreaChart[T]{
		data:        append([]T(nil), data...),
		x:           func(datum T) float64 { return float64(x(datum)) },
		series:      append([]AreaSeries[T](nil), series...),
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
func (c *AreaChart[T]) SetData(data []T) *AreaChart[T] {
	c.data = append([]T(nil), data...)
	c.Refresh()
	return c
}

// SetSeries replaces the displayed series and refreshes the widget.
func (c *AreaChart[T]) SetSeries(series ...AreaSeries[T]) *AreaChart[T] {
	c.series = append([]AreaSeries[T](nil), series...)
	c.Refresh()
	return c
}

// SetXAxis replaces the x-axis configuration.
func (c *AreaChart[T]) SetXAxis(axis NumericAxis) *AreaChart[T] {
	c.xAxis = axis
	c.Refresh()
	return c
}

// SetYAxis replaces the y-axis configuration.
func (c *AreaChart[T]) SetYAxis(axis NumericAxis) *AreaChart[T] {
	c.yAxis = axis
	c.Refresh()
	return c
}

// SetStyle installs a callback that adjusts each series' style at render
// time, given its style as configured on the series and its index. It lets a
// theme change restyle a chart without recreating its series: mutate what
// the callback captures and call Refresh.
func (c *AreaChart[T]) SetStyle(style func(AreaStyle, int) AreaStyle) *AreaChart[T] {
	c.style = style
	c.Refresh()
	return c
}

// SetCurve selects linear, monotone, or stepped interpolation.
func (c *AreaChart[T]) SetCurve(curve Curve) *AreaChart[T] {
	c.curve = curve
	c.Refresh()
	return c
}

// SetGapPolicy controls how optional series values are connected.
func (c *AreaChart[T]) SetGapPolicy(policy GapPolicy) *AreaChart[T] {
	c.gapPolicy = policy
	c.Refresh()
	return c
}

// SetSeriesLayout selects overlap, stack, normalized stack, or diverging stack.
// SeriesGroup is treated as SeriesOverlap for area charts.
func (c *AreaChart[T]) SetSeriesLayout(layout SeriesLayout) *AreaChart[T] {
	if layout == SeriesGroup {
		layout = SeriesOverlap
	}
	c.seriesLayout = layout
	c.Refresh()
	return c
}

// SetPadding sets logical-pixel insets around the chart.
func (c *AreaChart[T]) SetPadding(padding Insets) *AreaChart[T] {
	c.padding = padding
	c.Refresh()
	return c
}

// CreateRenderer implements fyne.Widget.
func (c *AreaChart[T]) CreateRenderer() fyne.WidgetRenderer {
	return newChartRenderer(c)
}

func (c *AreaChart[T]) chartMinSize() fyne.Size { return c.minimumSize }

func (c *AreaChart[T]) chartObjects(size fyne.Size) []fyne.CanvasObject {
	xValues, values, defined, xDomain, yDomain, lower, upper := c.areaValues()
	plot := cartesianPlot(size, c.padding, c.xAxis.visible, c.yAxis.visible)
	plot = fitNumericAxisMargin(c, plot, c.xAxis, xDomain, true)
	plot = fitNumericAxisMargin(c, plot, c.yAxis, yDomain, false)
	objects := renderNumericAxes(c, plot, c.xAxis, c.yAxis, xDomain, yDomain)
	for seriesIndex, series := range c.series {
		segments := areaPointSegments(xValues, lower[seriesIndex], upper[seriesIndex], defined[seriesIndex], c.gapPolicy, plot, xDomain, yDomain)
		style := series.style
		if c.style != nil {
			style = c.style(style, seriesIndex)
		}
		fillOpacity := style.Fill.Opacity
		if fillOpacity <= 0 {
			fillOpacity = 0.3
		}
		baseColor := seriesColor(c, seriesIndex, style.Fill.Color)
		fill := withOpacity(baseColor, fillOpacity)
		stroke := seriesColor(c, seriesIndex, style.Stroke.Color)
		strokeWidth := style.Stroke.Width
		if strokeWidth <= 0 {
			strokeWidth = 2
		}
		for _, segment := range segments {
			upperPoints := curvePoints(segment.upper, c.curve)
			lowerPoints := curvePoints(segment.lower, c.curve)
			count := min(len(upperPoints), len(lowerPoints))
			for index := 1; index < count; index++ {
				// Degenerate vertical polygons confuse the vector triangulator used
				// by stepped curves, so only fill intervals with non-zero width.
				if upperPoints[index].X != upperPoints[index-1].X {
					polygon := canvas.NewArbitraryPolygon([]fyne.Position{
						upperPoints[index-1], upperPoints[index],
						lowerPoints[index], lowerPoints[index-1],
					}, fill)
					polygon.Resize(size)
					objects = append(objects, polygon)
				}
				line := canvas.NewLine(stroke)
				line.StrokeWidth = strokeWidth
				line.Position1 = upperPoints[index-1]
				line.Position2 = upperPoints[index]
				objects = append(objects, line)
			}
		}
	}
	_ = values
	return objects
}

func (c *AreaChart[T]) areaValues() ([]float64, [][]float64, [][]bool, Domain, Domain, [][]float64, [][]float64) {
	xValues := make([]float64, len(c.data))
	values := make([][]float64, len(c.series))
	defined := make([][]bool, len(c.series))
	xMin, xMax := math.Inf(1), math.Inf(-1)
	for index, datum := range c.data {
		xValue := c.x(datum)
		xValues[index] = xValue
		if finite(xValue) {
			xMin, xMax = min(xMin, xValue), max(xMax, xValue)
		}
	}
	for seriesIndex, series := range c.series {
		values[seriesIndex] = make([]float64, len(c.data))
		defined[seriesIndex] = make([]bool, len(c.data))
		for index, datum := range c.data {
			value, ok := series.value(datum)
			ok = ok && finite(value) && finite(xValues[index])
			if !ok && c.gapPolicy == GapZero && finite(xValues[index]) {
				value, ok = 0, true
			}
			values[seriesIndex][index] = value
			defined[seriesIndex][index] = ok
		}
	}
	lower, upper := barStack(values, defined, c.seriesLayout)
	yMin, yMax := barExtents(lower, upper, defined)
	return xValues, values, defined,
		numericDomain(c.xAxis, xMin, xMax, false),
		numericDomain(c.yAxis, yMin, yMax, true),
		lower, upper
}

type areaSegment struct {
	lower []fyne.Position
	upper []fyne.Position
}

func areaPointSegments(xValues, lower, upper []float64, defined []bool, policy GapPolicy, plot plotRect, xDomain, yDomain Domain) []areaSegment {
	segments := make([]areaSegment, 0, 2)
	current := areaSegment{
		lower: make([]fyne.Position, 0, len(xValues)),
		upper: make([]fyne.Position, 0, len(xValues)),
	}
	flush := func() {
		if len(current.upper) > 0 {
			segments = append(segments, current)
		}
		current = areaSegment{}
	}
	for index, xValue := range xValues {
		if !defined[index] {
			if policy == GapBreak {
				flush()
			}
			continue
		}
		x := scale(xValue, xDomain, plot.left, plot.right)
		current.lower = append(current.lower, fyne.NewPos(x, scale(lower[index], yDomain, plot.bottom, plot.top)))
		current.upper = append(current.upper, fyne.NewPos(x, scale(upper[index], yDomain, plot.bottom, plot.top)))
	}
	flush()
	return segments
}
