package fyneline

import "image/color"

// Number is a numeric value accepted by chart accessors.
type Number interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
		~float32 | ~float64
}

// Insets specifies logical-pixel padding around a chart's plot area.
type Insets struct {
	Top, Right, Bottom, Left float32
}

// Domain specifies the inclusive minimum and maximum values of an axis.
type Domain struct {
	Min, Max float64
}

// Curve controls how points are connected.
type Curve uint8

const (
	// CurveLinear connects points with straight lines.
	CurveLinear Curve = iota
	// CurveMonotoneX connects points smoothly without introducing local extrema.
	CurveMonotoneX
	// CurveStep changes value halfway between adjacent points.
	CurveStep
	// CurveStepBefore changes value before moving to the next x position.
	CurveStepBefore
	// CurveStepAfter changes value after moving from the current x position.
	CurveStepAfter
)

// GapPolicy controls how a series handles values reported as missing.
type GapPolicy uint8

const (
	// GapBreak creates separate paths around missing values.
	GapBreak GapPolicy = iota
	// GapConnect connects the nearest defined values.
	GapConnect
	// GapZero treats missing values as zero.
	GapZero
)

// SeriesLayout controls how multiple series share the same chart space.
type SeriesLayout uint8

const (
	// SeriesOverlap draws each series independently in the same plot area.
	SeriesOverlap SeriesLayout = iota
	// SeriesStack places each series on top of the preceding series.
	SeriesStack
	// SeriesStackExpand normalizes each stack to a total of one.
	SeriesStackExpand
	// SeriesStackDiverging stacks positive and negative values away from zero.
	SeriesStackDiverging
	// SeriesGroup places series beside each other. It is supported by BarChart.
	SeriesGroup
)

// BarOrientation controls the direction in which bars grow.
type BarOrientation uint8

const (
	// BarVertical displays categories horizontally and values vertically.
	BarVertical BarOrientation = iota
	// BarHorizontal displays categories vertically and values horizontally.
	BarHorizontal
)

// StrokeStyle describes the outline of a rendered mark.
type StrokeStyle struct {
	Color color.Color
	Width float32
}

// FillStyle describes the interior of a rendered mark. An opacity greater than
// zero is clamped to one; zero uses the mark's default opacity.
type FillStyle struct {
	Color   color.Color
	Opacity float32
}

// AxisStyle controls axis, grid, and label appearance. Nil colors use the
// active Fyne theme.
type AxisStyle struct {
	LineColor  color.Color
	LabelColor color.Color
	GridColor  color.Color
	LineWidth  float32
	TextSize   float32
}

// BarStyle controls a bar's fill, outline, and corner radius.
type BarStyle struct {
	Fill         FillStyle
	Stroke       StrokeStyle
	CornerRadius float32
}

// AreaStyle controls an area's fill and boundary line.
type AreaStyle struct {
	Fill   FillStyle
	Stroke StrokeStyle
}

// SplineStyle controls a spline's line.
type SplineStyle struct {
	Stroke StrokeStyle
}

// ArcStyle controls an arc's fill and outline.
type ArcStyle struct {
	Fill   FillStyle
	Stroke StrokeStyle
}

// Palette returns an index-based color accessor that cycles through colors.
// A palette with no colors always returns nil.
func Palette(colors ...color.Color) func(index int) color.Color {
	return func(index int) color.Color {
		if len(colors) == 0 {
			return nil
		}
		index %= len(colors)
		if index < 0 {
			index += len(colors)
		}
		return colors[index]
	}
}
