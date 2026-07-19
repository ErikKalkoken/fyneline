package fyneline

import (
	"math"
	"time"
)

// NumericAxis configures a continuous numeric axis. Build one with
// NewNumericAxis and customize it using its With methods.
type NumericAxis struct {
	domain    Domain
	hasDomain bool
	tickCount int
	formatter func(float64) string
	label     string
	grid      bool
	visible   bool
	style     AxisStyle
}

// NewNumericAxis returns a visible, automatically ranged numeric axis.
func NewNumericAxis() NumericAxis {
	return NumericAxis{tickCount: 5, grid: true, visible: true}
}

// WithDomain returns an axis fixed to the supplied domain.
func (a NumericAxis) WithDomain(minimum, maximum float64) NumericAxis {
	a.domain = Domain{Min: minimum, Max: maximum}
	a.hasDomain = true
	return a
}

// WithAutoDomain returns an axis whose domain is inferred from its data.
func (a NumericAxis) WithAutoDomain() NumericAxis {
	a.hasDomain = false
	return a
}

// WithTickCount returns an axis targeting count labeled ticks.
func (a NumericAxis) WithTickCount(count int) NumericAxis {
	if count < 2 {
		count = 2
	}
	a.tickCount = count
	return a
}

// WithFormatter returns an axis that formats tick values with format.
func (a NumericAxis) WithFormatter(format func(float64) string) NumericAxis {
	a.formatter = format
	return a
}

// WithLabel returns an axis with the supplied title.
func (a NumericAxis) WithLabel(label string) NumericAxis {
	a.label = label
	return a
}

// WithGrid returns an axis with grid lines enabled or disabled.
func (a NumericAxis) WithGrid(visible bool) NumericAxis {
	a.grid = visible
	return a
}

// WithVisible returns an axis with its line, ticks, and labels shown or hidden.
func (a NumericAxis) WithVisible(visible bool) NumericAxis {
	a.visible = visible
	return a
}

// WithStyle returns an axis using style.
func (a NumericAxis) WithStyle(style AxisStyle) NumericAxis {
	a.style = style
	return a
}

// CategoryAxis configures a discrete category axis. Build one with
// NewCategoryAxis and customize it using its With methods.
type CategoryAxis struct {
	categories []string
	formatter  func(string) string
	label      string
	grid       bool
	visible    bool
	style      AxisStyle
}

// NewCategoryAxis returns a visible category axis using data order.
func NewCategoryAxis() CategoryAxis {
	return CategoryAxis{visible: true}
}

// WithCategories returns an axis with an explicit category order.
func (a CategoryAxis) WithCategories(categories ...string) CategoryAxis {
	a.categories = append([]string(nil), categories...)
	return a
}

// WithLabel returns an axis with the supplied title.
func (a CategoryAxis) WithLabel(label string) CategoryAxis {
	a.label = label
	return a
}

// WithFormatter returns an axis that formats category labels with format.
func (a CategoryAxis) WithFormatter(format func(string) string) CategoryAxis {
	a.formatter = format
	return a
}

// WithGrid returns an axis with grid lines enabled or disabled.
func (a CategoryAxis) WithGrid(visible bool) CategoryAxis {
	a.grid = visible
	return a
}

// WithVisible returns an axis with its line, ticks, and labels shown or hidden.
func (a CategoryAxis) WithVisible(visible bool) CategoryAxis {
	a.visible = visible
	return a
}

// WithStyle returns an axis using style.
func (a CategoryAxis) WithStyle(style AxisStyle) CategoryAxis {
	a.style = style
	return a
}

// TimeAccessor converts time values to Unix seconds for a continuous axis.
func TimeAccessor[T any](accessor func(T) time.Time) func(T) float64 {
	return func(datum T) float64 {
		return float64(accessor(datum).UnixNano()) / float64(time.Second)
	}
}

// NewTimeAxis returns a numeric axis formatted using layout and location. A
// nil location formats values in time.Local.
func NewTimeAxis(layout string, location *time.Location) NumericAxis {
	if location == nil {
		location = time.Local
	}
	return NewNumericAxis().WithFormatter(func(value float64) string {
		seconds, fraction := math.Modf(value)
		return time.Unix(int64(seconds), int64(fraction*float64(time.Second))).In(location).Format(layout)
	})
}
