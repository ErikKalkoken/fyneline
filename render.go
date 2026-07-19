package fyneline

import (
	"fmt"
	"image/color"
	"math"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/theme"
)

var defaultSeriesColors = []color.Color{
	color.NRGBA{R: 62, G: 126, B: 247, A: 255},
	color.NRGBA{R: 240, G: 135, B: 48, A: 255},
	color.NRGBA{R: 47, G: 176, B: 117, A: 255},
	color.NRGBA{R: 220, G: 72, B: 103, A: 255},
	color.NRGBA{R: 139, G: 92, B: 246, A: 255},
	color.NRGBA{R: 16, G: 164, B: 190, A: 255},
}

type chartDrawable interface {
	fyne.Widget
	chartObjects(fyne.Size) []fyne.CanvasObject
	chartMinSize() fyne.Size
}

type chartRenderer struct {
	chart     chartDrawable
	container *fyne.Container
	size      fyne.Size
}

func newChartRenderer(chart chartDrawable) fyne.WidgetRenderer {
	r := &chartRenderer{
		chart:     chart,
		container: fyne.NewContainerWithoutLayout(),
		size:      chart.chartMinSize(),
	}
	r.rebuild()
	return r
}

func (r *chartRenderer) Destroy() {}

func (r *chartRenderer) Layout(size fyne.Size) {
	r.size = size
	r.container.Resize(size)
	r.rebuild()
}

func (r *chartRenderer) MinSize() fyne.Size {
	return r.chart.chartMinSize()
}

func (r *chartRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.container}
}

func (r *chartRenderer) Refresh() {
	r.rebuild()
}

func (r *chartRenderer) rebuild() {
	r.container.Objects = r.chart.chartObjects(r.size)
	r.container.Refresh()
}

type plotRect struct {
	left, top, right, bottom float32
}

func (p plotRect) width() float32  { return max(p.right-p.left, 0) }
func (p plotRect) height() float32 { return max(p.bottom-p.top, 0) }

func cartesianPlot(size fyne.Size, padding Insets, xVisible, yVisible bool) plotRect {
	left := padding.Left + 8
	right := size.Width - padding.Right - 8
	top := padding.Top + 8
	bottom := size.Height - padding.Bottom - 8
	if yVisible {
		left += 44
	}
	if xVisible {
		bottom -= 28
	}
	return plotRect{left: left, top: top, right: right, bottom: bottom}
}

func numericDomain(axis NumericAxis, minimum, maximum float64, includeZero bool) Domain {
	if axis.hasDomain {
		minimum, maximum = axis.domain.Min, axis.domain.Max
	} else if includeZero {
		minimum = min(minimum, 0)
		maximum = max(maximum, 0)
	}
	if !finite(minimum) || !finite(maximum) {
		minimum, maximum = 0, 1
	}
	if minimum > maximum {
		minimum, maximum = maximum, minimum
	}
	if minimum == maximum {
		amount := math.Abs(minimum) * 0.1
		if amount == 0 {
			amount = 1
		}
		minimum -= amount
		maximum += amount
	}
	return Domain{Min: minimum, Max: maximum}
}

func scale(value float64, domain Domain, start, end float32) float32 {
	return start + float32((value-domain.Min)/(domain.Max-domain.Min))*(end-start)
}

func renderNumericAxes(widget fyne.Widget, plot plotRect, xAxis, yAxis NumericAxis, xDomain, yDomain Domain) []fyne.CanvasObject {
	objects := make([]fyne.CanvasObject, 0, 32)
	if yAxis.grid {
		objects = append(objects, numericGrid(widget, plot, yAxis, yDomain, false)...)
	}
	if xAxis.grid {
		objects = append(objects, numericGrid(widget, plot, xAxis, xDomain, true)...)
	}
	if yAxis.visible {
		objects = append(objects, numericAxis(widget, plot, yAxis, yDomain, false)...)
	}
	if xAxis.visible {
		objects = append(objects, numericAxis(widget, plot, xAxis, xDomain, true)...)
	}
	return objects
}

func numericGrid(widget fyne.Widget, plot plotRect, axis NumericAxis, domain Domain, horizontalAxis bool) []fyne.CanvasObject {
	count := max(axis.tickCount, 2)
	lineColor := axis.style.GridColor
	if lineColor == nil {
		lineColor = withOpacity(theme.ColorForWidget(theme.ColorNameSeparator, widget), 0.35)
	}
	objects := make([]fyne.CanvasObject, 0, count)
	for i := 0; i < count; i++ {
		fraction := float32(i) / float32(count-1)
		line := canvas.NewLine(lineColor)
		line.StrokeWidth = 1
		if horizontalAxis {
			x := plot.left + fraction*plot.width()
			line.Position1 = fyne.NewPos(x, plot.top)
			line.Position2 = fyne.NewPos(x, plot.bottom)
		} else {
			y := plot.bottom - fraction*plot.height()
			line.Position1 = fyne.NewPos(plot.left, y)
			line.Position2 = fyne.NewPos(plot.right, y)
		}
		objects = append(objects, line)
	}
	return objects
}

func numericAxis(widget fyne.Widget, plot plotRect, axis NumericAxis, domain Domain, horizontal bool) []fyne.CanvasObject {
	lineColor := axis.style.LineColor
	if lineColor == nil {
		lineColor = theme.ColorForWidget(theme.ColorNameForeground, widget)
	}
	labelColor := axis.style.LabelColor
	if labelColor == nil {
		labelColor = theme.ColorForWidget(theme.ColorNameForeground, widget)
	}
	textSize := axis.style.TextSize
	if textSize <= 0 {
		textSize = theme.CurrentForWidget(widget).Size(theme.SizeNameCaptionText)
	}
	lineWidth := axis.style.LineWidth
	if lineWidth <= 0 {
		lineWidth = 1
	}
	axisLine := canvas.NewLine(lineColor)
	axisLine.StrokeWidth = lineWidth
	if horizontal {
		axisLine.Position1 = fyne.NewPos(plot.left, plot.bottom)
		axisLine.Position2 = fyne.NewPos(plot.right, plot.bottom)
	} else {
		axisLine.Position1 = fyne.NewPos(plot.left, plot.top)
		axisLine.Position2 = fyne.NewPos(plot.left, plot.bottom)
	}
	objects := []fyne.CanvasObject{axisLine}
	count := max(axis.tickCount, 2)
	for i := 0; i < count; i++ {
		fraction := float32(i) / float32(count-1)
		value := domain.Min + float64(fraction)*(domain.Max-domain.Min)
		label := formatNumber(axis, value)
		text := canvas.NewText(label, labelColor)
		text.TextSize = textSize
		measured := fyne.MeasureText(label, textSize, fyne.TextStyle{})
		if horizontal {
			x := plot.left + fraction*plot.width()
			text.Move(fyne.NewPos(x-measured.Width/2, plot.bottom+5))
		} else {
			y := plot.bottom - fraction*plot.height()
			text.Move(fyne.NewPos(plot.left-measured.Width-6, y-measured.Height/2))
		}
		objects = append(objects, text)
	}
	if axis.label != "" {
		text := canvas.NewText(axis.label, labelColor)
		text.TextSize = textSize
		text.TextStyle = fyne.TextStyle{Bold: true}
		measured := fyne.MeasureText(axis.label, textSize, text.TextStyle)
		if horizontal {
			text.Move(fyne.NewPos(plot.left+(plot.width()-measured.Width)/2, plot.bottom+18))
		} else {
			text.Move(fyne.NewPos(max(plot.left-measured.Width-6, 0), plot.top))
		}
		objects = append(objects, text)
	}
	return objects
}

func renderCategoryAxis(widget fyne.Widget, plot plotRect, axis CategoryAxis, categories []string, horizontal bool) []fyne.CanvasObject {
	if !axis.visible || len(categories) == 0 {
		return nil
	}
	lineColor := axis.style.LineColor
	if lineColor == nil {
		lineColor = theme.ColorForWidget(theme.ColorNameForeground, widget)
	}
	labelColor := axis.style.LabelColor
	if labelColor == nil {
		labelColor = theme.ColorForWidget(theme.ColorNameForeground, widget)
	}
	textSize := axis.style.TextSize
	if textSize <= 0 {
		textSize = theme.CurrentForWidget(widget).Size(theme.SizeNameCaptionText)
	}
	axisLine := canvas.NewLine(lineColor)
	axisLine.StrokeWidth = max(axis.style.LineWidth, 1)
	if horizontal {
		axisLine.Position1 = fyne.NewPos(plot.left, plot.bottom)
		axisLine.Position2 = fyne.NewPos(plot.right, plot.bottom)
	} else {
		axisLine.Position1 = fyne.NewPos(plot.left, plot.top)
		axisLine.Position2 = fyne.NewPos(plot.left, plot.bottom)
	}
	objects := []fyne.CanvasObject{axisLine}
	for index, category := range categories {
		if axis.formatter != nil {
			category = axis.formatter(category)
		}
		text := canvas.NewText(category, labelColor)
		text.TextSize = textSize
		measured := fyne.MeasureText(category, textSize, fyne.TextStyle{})
		fraction := (float32(index) + 0.5) / float32(len(categories))
		if horizontal {
			x := plot.left + fraction*plot.width()
			text.Move(fyne.NewPos(x-measured.Width/2, plot.bottom+5))
		} else {
			y := plot.top + fraction*plot.height()
			text.Move(fyne.NewPos(plot.left-measured.Width-6, y-measured.Height/2))
		}
		objects = append(objects, text)
	}
	if axis.label != "" {
		text := canvas.NewText(axis.label, labelColor)
		text.TextSize = textSize
		text.TextStyle = fyne.TextStyle{Bold: true}
		measured := fyne.MeasureText(axis.label, textSize, text.TextStyle)
		if horizontal {
			text.Move(fyne.NewPos(plot.left+(plot.width()-measured.Width)/2, plot.bottom+18))
		} else {
			text.Move(fyne.NewPos(max(plot.left-measured.Width-6, 0), plot.top))
		}
		objects = append(objects, text)
	}
	return objects
}

func categoryGrid(widget fyne.Widget, plot plotRect, axis CategoryAxis, count int, horizontal bool) []fyne.CanvasObject {
	if !axis.grid || count == 0 {
		return nil
	}
	lineColor := axis.style.GridColor
	if lineColor == nil {
		lineColor = withOpacity(theme.ColorForWidget(theme.ColorNameSeparator, widget), 0.35)
	}
	objects := make([]fyne.CanvasObject, 0, count)
	for index := 0; index < count; index++ {
		fraction := (float32(index) + 0.5) / float32(count)
		line := canvas.NewLine(lineColor)
		line.StrokeWidth = 1
		if horizontal {
			x := plot.left + fraction*plot.width()
			line.Position1 = fyne.NewPos(x, plot.top)
			line.Position2 = fyne.NewPos(x, plot.bottom)
		} else {
			y := plot.top + fraction*plot.height()
			line.Position1 = fyne.NewPos(plot.left, y)
			line.Position2 = fyne.NewPos(plot.right, y)
		}
		objects = append(objects, line)
	}
	return objects
}

func formatNumber(axis NumericAxis, value float64) string {
	if axis.formatter != nil {
		return axis.formatter(value)
	}
	return fmt.Sprintf("%g", value)
}

func seriesColor(widget fyne.Widget, index int, override color.Color) color.Color {
	if override != nil {
		return override
	}
	if index == 0 {
		return theme.ColorForWidget(theme.ColorNamePrimary, widget)
	}
	return defaultSeriesColors[index%len(defaultSeriesColors)]
}

func withOpacity(c color.Color, opacity float32) color.Color {
	if c == nil {
		return nil
	}
	opacity = min(max(opacity, 0), 1)
	r, g, b, a := c.RGBA()
	return color.NRGBA{
		R: uint8(r >> 8), G: uint8(g >> 8), B: uint8(b >> 8),
		A: uint8(float32(a>>8) * opacity),
	}
}

func finite(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}

func clampFraction(value, fallback float32) float32 {
	if value < 0 || value >= 1 || math.IsNaN(float64(value)) {
		return fallback
	}
	return value
}
