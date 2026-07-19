package fyneline

import (
	"math"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"
)

// BarChart displays categorical data using rectangular bars. It defaults to
// vertical, overlapping series with LayerChart-compatible band padding.
type BarChart[T any] struct {
	widget.BaseWidget

	data         []T
	category     func(T) string
	series       []BarSeries[T]
	orientation  BarOrientation
	seriesLayout SeriesLayout
	bandPadding  float32
	groupPadding float32
	stackPadding float32
	categoryAxis CategoryAxis
	valueAxis    NumericAxis
	padding      Insets
	minimumSize  fyne.Size
}

// NewBarChart constructs a category-based bar chart. Categories appear in
// first-seen order unless CategoryAxis.WithCategories supplies an order.
func NewBarChart[T any](data []T, category func(T) string, series ...BarSeries[T]) *BarChart[T] {
	chart := &BarChart[T]{
		data:         append([]T(nil), data...),
		category:     category,
		series:       append([]BarSeries[T](nil), series...),
		bandPadding:  0.4,
		categoryAxis: NewCategoryAxis(),
		valueAxis:    NewNumericAxis(),
		minimumSize:  fyne.NewSize(240, 160),
	}
	chart.ExtendBaseWidget(chart)
	return chart
}

// SetData replaces the chart data and refreshes the widget.
func (c *BarChart[T]) SetData(data []T) *BarChart[T] {
	c.data = append([]T(nil), data...)
	c.Refresh()
	return c
}

// SetSeries replaces the displayed series and refreshes the widget.
func (c *BarChart[T]) SetSeries(series ...BarSeries[T]) *BarChart[T] {
	c.series = append([]BarSeries[T](nil), series...)
	c.Refresh()
	return c
}

// SetSeriesLayout selects overlap, stack, normalized stack, diverging stack,
// or grouped rendering.
func (c *BarChart[T]) SetSeriesLayout(layout SeriesLayout) *BarChart[T] {
	c.seriesLayout = layout
	c.Refresh()
	return c
}

// SetOrientation selects vertical or horizontal bars.
func (c *BarChart[T]) SetOrientation(orientation BarOrientation) *BarChart[T] {
	c.orientation = orientation
	c.Refresh()
	return c
}

// SetCategoryAxis replaces the category-axis configuration.
func (c *BarChart[T]) SetCategoryAxis(axis CategoryAxis) *BarChart[T] {
	c.categoryAxis = axis
	c.Refresh()
	return c
}

// SetValueAxis replaces the value-axis configuration.
func (c *BarChart[T]) SetValueAxis(axis NumericAxis) *BarChart[T] {
	c.valueAxis = axis
	c.Refresh()
	return c
}

// SetBandPadding sets the empty fraction of each category band. Values outside
// [0, 1) are replaced by the default 0.4.
func (c *BarChart[T]) SetBandPadding(padding float32) *BarChart[T] {
	c.bandPadding = clampFraction(padding, 0.4)
	c.Refresh()
	return c
}

// SetGroupPadding sets the empty fraction between bars in a grouped band.
func (c *BarChart[T]) SetGroupPadding(padding float32) *BarChart[T] {
	c.groupPadding = clampFraction(padding, 0)
	c.Refresh()
	return c
}

// SetStackPadding sets the empty fraction removed from stacked bar edges.
func (c *BarChart[T]) SetStackPadding(padding float32) *BarChart[T] {
	c.stackPadding = clampFraction(padding, 0)
	c.Refresh()
	return c
}

// SetPadding sets logical-pixel insets around the chart.
func (c *BarChart[T]) SetPadding(padding Insets) *BarChart[T] {
	c.padding = padding
	c.Refresh()
	return c
}

// CreateRenderer implements fyne.Widget.
func (c *BarChart[T]) CreateRenderer() fyne.WidgetRenderer {
	return newChartRenderer(c)
}

func (c *BarChart[T]) chartMinSize() fyne.Size { return c.minimumSize }

func (c *BarChart[T]) chartObjects(size fyne.Size) []fyne.CanvasObject {
	categories, rows, present := c.categoryRows()
	xVisible := c.categoryAxis.visible
	yVisible := c.valueAxis.visible
	if c.orientation == BarHorizontal {
		xVisible, yVisible = c.valueAxis.visible, c.categoryAxis.visible
	}
	values, defined := c.barValues(rows, present)
	lower, upper := barStack(values, defined, c.seriesLayout)
	minimum, maximum := barExtents(lower, upper, defined)
	domain := numericDomain(c.valueAxis, minimum, maximum, true)
	plot := cartesianPlot(size, c.padding, xVisible, yVisible)
	if c.orientation == BarVertical {
		plot = fitCategoryAxisMargin(c, plot, c.categoryAxis, categories, true)
		plot = fitNumericAxisMargin(c, plot, c.valueAxis, domain, false)
	} else {
		plot = fitNumericAxisMargin(c, plot, c.valueAxis, domain, true)
		plot = fitCategoryAxisMargin(c, plot, c.categoryAxis, categories, false)
	}

	objects := make([]fyne.CanvasObject, 0, len(categories)*max(len(c.series), 1)+24)
	objects = append(objects, categoryGrid(c, plot, c.categoryAxis, len(categories), c.orientation == BarVertical)...)
	if c.valueAxis.grid {
		objects = append(objects, numericGrid(c, plot, c.valueAxis, domain, c.orientation == BarHorizontal)...)
	}
	if c.valueAxis.visible {
		objects = append(objects, numericAxis(c, plot, c.valueAxis, domain, c.orientation == BarHorizontal)...)
	}
	objects = append(objects, renderCategoryAxis(c, plot, c.categoryAxis, categories, c.orientation == BarVertical)...)
	objects = append(objects, c.renderBars(plot, domain, values, lower, upper, defined)...)
	return objects
}

func (c *BarChart[T]) categoryRows() ([]string, []T, []bool) {
	byCategory := make(map[string]T, len(c.data))
	seen := make(map[string]bool, len(c.data))
	categories := make([]string, 0, len(c.data))
	for _, datum := range c.data {
		category := c.category(datum)
		byCategory[category] = datum
		if !seen[category] {
			seen[category] = true
			categories = append(categories, category)
		}
	}
	if len(c.categoryAxis.categories) > 0 {
		categories = append([]string(nil), c.categoryAxis.categories...)
	}
	rows := make([]T, len(categories))
	present := make([]bool, len(categories))
	for index, category := range categories {
		rows[index], present[index] = byCategory[category]
	}
	return categories, rows, present
}

func (c *BarChart[T]) barValues(rows []T, present []bool) ([][]float64, [][]bool) {
	values := make([][]float64, len(c.series))
	defined := make([][]bool, len(c.series))
	for seriesIndex, series := range c.series {
		values[seriesIndex] = make([]float64, len(rows))
		defined[seriesIndex] = make([]bool, len(rows))
		for rowIndex, datum := range rows {
			if !present[rowIndex] {
				continue
			}
			value, ok := series.value(datum)
			ok = ok && finite(value)
			values[seriesIndex][rowIndex] = value
			defined[seriesIndex][rowIndex] = ok
		}
	}
	return values, defined
}

func (c *BarChart[T]) renderBars(plot plotRect, domain Domain, values, lower, upper [][]float64, defined [][]bool) []fyne.CanvasObject {
	if len(values) == 0 || len(values[0]) == 0 {
		return nil
	}
	categoryCount, seriesCount := len(values[0]), len(values)
	available := plot.width()
	if c.orientation == BarHorizontal {
		available = plot.height()
	}
	bandSlot := available / float32(categoryCount)
	categoryBand := bandSlot * (1 - c.bandPadding)
	objects := make([]fyne.CanvasObject, 0, categoryCount*seriesCount)
	for categoryIndex := 0; categoryIndex < categoryCount; categoryIndex++ {
		for seriesIndex, series := range c.series {
			if !defined[seriesIndex][categoryIndex] {
				continue
			}
			start := lower[seriesIndex][categoryIndex]
			end := upper[seriesIndex][categoryIndex]
			barOffset, barSize := float32(0), categoryBand
			if c.seriesLayout == SeriesGroup {
				groupSlot := categoryBand / float32(seriesCount)
				barSize = groupSlot * (1 - c.groupPadding)
				barOffset = float32(seriesIndex)*groupSlot + (groupSlot-barSize)/2
			} else if c.seriesLayout == SeriesStack || c.seriesLayout == SeriesStackExpand || c.seriesLayout == SeriesStackDiverging {
				barSize = categoryBand * (1 - c.stackPadding)
				barOffset = (categoryBand - barSize) / 2
			}
			bandStart := float32(categoryIndex)*bandSlot + (bandSlot-categoryBand)/2 + barOffset
			style := series.style
			opacity := style.Fill.Opacity
			if opacity <= 0 {
				opacity = 1
			}
			fill := withOpacity(seriesColor(c, seriesIndex, style.Fill.Color), opacity)
			bar := canvas.NewRectangle(fill)
			bar.CornerRadius = max(style.CornerRadius, 0)
			bar.StrokeColor = style.Stroke.Color
			bar.StrokeWidth = max(style.Stroke.Width, 0)
			if c.orientation == BarVertical {
				y1 := scale(start, domain, plot.bottom, plot.top)
				y2 := scale(end, domain, plot.bottom, plot.top)
				bar.Move(fyne.NewPos(plot.left+bandStart, min(y1, y2)))
				bar.Resize(fyne.NewSize(max(barSize, 0), max(float32(math.Abs(float64(y2-y1))), 0.5)))
			} else {
				x1 := scale(start, domain, plot.left, plot.right)
				x2 := scale(end, domain, plot.left, plot.right)
				bar.Move(fyne.NewPos(min(x1, x2), plot.top+bandStart))
				bar.Resize(fyne.NewSize(max(float32(math.Abs(float64(x2-x1))), 0.5), max(barSize, 0)))
			}
			objects = append(objects, bar)
		}
	}
	return objects
}

func barStack(values [][]float64, defined [][]bool, layout SeriesLayout) ([][]float64, [][]float64) {
	lower := make([][]float64, len(values))
	upper := make([][]float64, len(values))
	if len(values) == 0 {
		return lower, upper
	}
	for i := range values {
		lower[i] = make([]float64, len(values[i]))
		upper[i] = make([]float64, len(values[i]))
	}
	for column := range values[0] {
		positive, negative := 0.0, 0.0
		total := 0.0
		if layout == SeriesStackExpand {
			for series := range values {
				if defined[series][column] {
					total += math.Abs(values[series][column])
				}
			}
		}
		for series := range values {
			if !defined[series][column] {
				continue
			}
			value := values[series][column]
			if layout == SeriesStackExpand && total != 0 {
				value /= total
			}
			switch layout {
			case SeriesStack, SeriesStackExpand:
				lower[series][column] = positive
				positive += value
				upper[series][column] = positive
			case SeriesStackDiverging:
				if value >= 0 {
					lower[series][column] = positive
					positive += value
					upper[series][column] = positive
				} else {
					lower[series][column] = negative
					negative += value
					upper[series][column] = negative
				}
			default:
				lower[series][column] = 0
				upper[series][column] = value
			}
		}
	}
	return lower, upper
}

func barExtents(lower, upper [][]float64, defined [][]bool) (float64, float64) {
	minimum, maximum := math.Inf(1), math.Inf(-1)
	for series := range upper {
		for index := range upper[series] {
			if !defined[series][index] {
				continue
			}
			minimum = min(minimum, lower[series][index], upper[series][index])
			maximum = max(maximum, lower[series][index], upper[series][index])
		}
	}
	return minimum, maximum
}
