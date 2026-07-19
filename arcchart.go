package fyneline

import (
	"math"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// ArcChart displays values as portions of a circle. It supports pie, doughnut,
// gauge, and partial-circle layouts through its radius and range setters.
type ArcChart[T any] struct {
	widget.BaseWidget

	data         []T
	value        func(T) float64
	label        func(T) string
	style        func(T, int) ArcStyle
	startAngle   float64
	endAngle     float64
	innerRadius  float32
	outerRadius  float32
	padAngle     float64
	cornerRadius float32
	maxValue     float64
	hasMaxValue  bool
	labels       bool
	padding      Insets
	minimumSize  fyne.Size
}

// NewArcChart constructs an arc chart. Angles use degrees, with zero at the
// top and positive values moving clockwise, matching LayerChart and canvas.Arc.
func NewArcChart[T any, N Number](data []T, value func(T) N, label func(T) string) *ArcChart[T] {
	chart := &ArcChart[T]{
		data:        append([]T(nil), data...),
		value:       func(datum T) float64 { return float64(value(datum)) },
		label:       label,
		endAngle:    360,
		outerRadius: 1,
		minimumSize: fyne.NewSize(160, 160),
	}
	chart.ExtendBaseWidget(chart)
	return chart
}

// SetData replaces the chart data and refreshes the widget.
func (c *ArcChart[T]) SetData(data []T) *ArcChart[T] {
	c.data = append([]T(nil), data...)
	c.Refresh()
	return c
}

// SetStyle assigns per-datum styling. The index is the datum's display index.
func (c *ArcChart[T]) SetStyle(style func(T, int) ArcStyle) *ArcChart[T] {
	c.style = style
	c.Refresh()
	return c
}

// SetRange sets the chart's angular range in degrees.
func (c *ArcChart[T]) SetRange(start, end float64) *ArcChart[T] {
	if finite(start) && finite(end) {
		c.startAngle, c.endAngle = start, end
	}
	c.Refresh()
	return c
}

// SetInnerRadius sets the cutout as a fraction of the outer radius.
func (c *ArcChart[T]) SetInnerRadius(fraction float32) *ArcChart[T] {
	if !finite(float64(fraction)) {
		fraction = 0
	}
	c.innerRadius = min(max(fraction, 0), 0.99)
	c.Refresh()
	return c
}

// SetOuterRadius sets the outer radius as a fraction of available space.
func (c *ArcChart[T]) SetOuterRadius(fraction float32) *ArcChart[T] {
	if !finite(float64(fraction)) || fraction <= 0 {
		fraction = 1
	}
	c.outerRadius = min(fraction, 1)
	c.Refresh()
	return c
}

// SetPadAngle sets the angular space between adjacent arcs in degrees.
func (c *ArcChart[T]) SetPadAngle(degrees float64) *ArcChart[T] {
	if finite(degrees) {
		c.padAngle = max(degrees, 0)
	}
	c.Refresh()
	return c
}

// SetCornerRadius sets arc corner rounding in logical pixels.
func (c *ArcChart[T]) SetCornerRadius(radius float32) *ArcChart[T] {
	if !finite(float64(radius)) {
		radius = 0
	}
	c.cornerRadius = max(radius, 0)
	c.Refresh()
	return c
}

// SetMaxValue fixes the value represented by the complete angular range. It is
// useful for gauges whose data does not fill the range.
func (c *ArcChart[T]) SetMaxValue(value float64) *ArcChart[T] {
	c.maxValue, c.hasMaxValue = value, finite(value) && value > 0
	c.Refresh()
	return c
}

// SetLabels shows or hides labels at arc centroids.
func (c *ArcChart[T]) SetLabels(visible bool) *ArcChart[T] {
	c.labels = visible
	c.Refresh()
	return c
}

// SetPadding sets logical-pixel insets around the chart.
func (c *ArcChart[T]) SetPadding(padding Insets) *ArcChart[T] {
	c.padding = padding
	c.Refresh()
	return c
}

// CreateRenderer implements fyne.Widget.
func (c *ArcChart[T]) CreateRenderer() fyne.WidgetRenderer {
	return newChartRenderer(c)
}

func (c *ArcChart[T]) chartMinSize() fyne.Size { return c.minimumSize }

func (c *ArcChart[T]) chartObjects(size fyne.Size) []fyne.CanvasObject {
	availableWidth := max(size.Width-c.padding.Left-c.padding.Right, 0)
	availableHeight := max(size.Height-c.padding.Top-c.padding.Bottom, 0)
	diameter := min(availableWidth, availableHeight) * c.outerRadius
	center := fyne.NewPos(
		c.padding.Left+availableWidth/2,
		c.padding.Top+availableHeight/2,
	)
	values := make([]float64, len(c.data))
	total := 0.0
	for index, datum := range c.data {
		value := c.value(datum)
		if finite(value) && value > 0 {
			values[index] = value
			total += value
		}
	}
	domainTotal := total
	if c.hasMaxValue {
		domainTotal = max(c.maxValue, total)
	}
	if domainTotal <= 0 || diameter <= 0 {
		return nil
	}
	span := c.endAngle - c.startAngle
	current := c.startAngle
	objects := make([]fyne.CanvasObject, 0, len(c.data)*2)
	for index, datum := range c.data {
		value := values[index]
		if value <= 0 {
			continue
		}
		arcSpan := span * value / domainTotal
		pad := min(c.padAngle, math.Abs(arcSpan))
		start := current + math.Copysign(pad/2, span)
		end := current + arcSpan - math.Copysign(pad/2, span)
		style := ArcStyle{}
		if c.style != nil {
			style = c.style(datum, index)
		}
		opacity := style.Fill.Opacity
		if opacity <= 0 {
			opacity = 1
		}
		fill := withOpacity(seriesColor(c, index, style.Fill.Color), opacity)
		arc := canvas.NewArc(float32(start), float32(end), c.innerRadius, fill)
		arc.CornerRadius = c.cornerRadius
		arc.StrokeColor = style.Stroke.Color
		arc.StrokeWidth = max(style.Stroke.Width, 0)
		arc.Resize(fyne.NewSize(diameter, diameter))
		// Fyne's painters currently treat Arc.Position as its top-left origin,
		// despite the Canvas API documenting it as the center.
		arc.Move(center.Subtract(fyne.NewPos(diameter/2, diameter/2)))
		objects = append(objects, arc)
		if c.labels && c.label != nil {
			label := c.label(datum)
			text := canvas.NewText(label, theme.ColorForWidget(theme.ColorNameForeground, c))
			text.TextSize = theme.CurrentForWidget(c).Size(theme.SizeNameCaptionText)
			angle := (start + end) / 2 * math.Pi / 180
			radius := float64(diameter) * float64((1+c.innerRadius)/4)
			position := fyne.NewPos(
				center.X+float32(math.Sin(angle)*radius),
				center.Y-float32(math.Cos(angle)*radius),
			)
			measured := fyne.MeasureText(label, text.TextSize, fyne.TextStyle{})
			text.Move(position.Subtract(fyne.NewPos(measured.Width/2, measured.Height/2)))
			objects = append(objects, text)
		}
		current += arcSpan
	}
	return objects
}
