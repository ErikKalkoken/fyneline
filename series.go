package fyneline

import "image/color"

type valueAccessor[T any] func(T) (float64, bool)

// BarSeries describes one named set of values in a BarChart.
type BarSeries[T any] struct {
	name  string
	value valueAccessor[T]
	style BarStyle
}

// NewBarSeries creates a bar series from a required numeric accessor.
func NewBarSeries[T any, N Number](name string, value func(T) N) BarSeries[T] {
	return BarSeries[T]{name: name, value: requiredAccessor(value)}
}

// NewOptionalBarSeries creates a bar series whose accessor can report gaps.
func NewOptionalBarSeries[T any, N Number](name string, value func(T) (N, bool)) BarSeries[T] {
	return BarSeries[T]{name: name, value: optionalAccessor(value)}
}

// WithStyle returns a copy of the series using style.
func (s BarSeries[T]) WithStyle(style BarStyle) BarSeries[T] {
	s.style = style
	return s
}

// WithFill returns a copy of the series filled with c.
func (s BarSeries[T]) WithFill(c color.Color) BarSeries[T] {
	s.style.Fill.Color = c
	s.style.Fill.Opacity = 1
	return s
}

// WithStroke returns a copy of the series using stroke.
func (s BarSeries[T]) WithStroke(stroke StrokeStyle) BarSeries[T] {
	s.style.Stroke = stroke
	return s
}

// AreaSeries describes one named set of values in an AreaChart.
type AreaSeries[T any] struct {
	name  string
	value valueAccessor[T]
	style AreaStyle
}

// NewAreaSeries creates an area series from a required numeric accessor.
func NewAreaSeries[T any, N Number](name string, value func(T) N) AreaSeries[T] {
	return AreaSeries[T]{name: name, value: requiredAccessor(value)}
}

// NewOptionalAreaSeries creates an area series whose accessor can report gaps.
func NewOptionalAreaSeries[T any, N Number](name string, value func(T) (N, bool)) AreaSeries[T] {
	return AreaSeries[T]{name: name, value: optionalAccessor(value)}
}

// WithStyle returns a copy of the series using style.
func (s AreaSeries[T]) WithStyle(style AreaStyle) AreaSeries[T] {
	s.style = style
	return s
}

// WithFill returns a copy of the series filled with c.
func (s AreaSeries[T]) WithFill(c color.Color) AreaSeries[T] {
	s.style.Fill.Color = c
	s.style.Fill.Opacity = 0.3
	return s
}

// WithStroke returns a copy of the series using stroke.
func (s AreaSeries[T]) WithStroke(stroke StrokeStyle) AreaSeries[T] {
	s.style.Stroke = stroke
	return s
}

// SplineSeries describes one named set of values in a Spline.
type SplineSeries[T any] struct {
	name  string
	value valueAccessor[T]
	style SplineStyle
}

// NewSplineSeries creates a spline series from a required numeric accessor.
func NewSplineSeries[T any, N Number](name string, value func(T) N) SplineSeries[T] {
	return SplineSeries[T]{name: name, value: requiredAccessor(value)}
}

// NewOptionalSplineSeries creates a spline series whose accessor can report gaps.
func NewOptionalSplineSeries[T any, N Number](name string, value func(T) (N, bool)) SplineSeries[T] {
	return SplineSeries[T]{name: name, value: optionalAccessor(value)}
}

// WithStyle returns a copy of the series using style.
func (s SplineSeries[T]) WithStyle(style SplineStyle) SplineSeries[T] {
	s.style = style
	return s
}

// WithColor returns a copy of the series using c.
func (s SplineSeries[T]) WithColor(c color.Color) SplineSeries[T] {
	s.style.Stroke.Color = c
	return s
}

// WithWidth returns a copy of the series using width.
func (s SplineSeries[T]) WithWidth(width float32) SplineSeries[T] {
	s.style.Stroke.Width = width
	return s
}

func requiredAccessor[T any, N Number](accessor func(T) N) valueAccessor[T] {
	return func(datum T) (float64, bool) {
		return float64(accessor(datum)), true
	}
}

func optionalAccessor[T any, N Number](accessor func(T) (N, bool)) valueAccessor[T] {
	return func(datum T) (float64, bool) {
		value, defined := accessor(datum)
		return float64(value), defined
	}
}
