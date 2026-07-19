package fyneline

import (
	"math"

	"fyne.io/fyne/v2"
)

func curvePoints(points []fyne.Position, curve Curve) []fyne.Position {
	if len(points) < 2 || curve == CurveLinear {
		return append([]fyne.Position(nil), points...)
	}
	switch curve {
	case CurveStep, CurveStepBefore, CurveStepAfter:
		return stepPoints(points, curve)
	case CurveMonotoneX:
		return monotonePoints(points, 10)
	default:
		return append([]fyne.Position(nil), points...)
	}
}

func stepPoints(points []fyne.Position, curve Curve) []fyne.Position {
	result := make([]fyne.Position, 0, len(points)*3)
	result = append(result, points[0])
	for i := 1; i < len(points); i++ {
		previous, current := points[i-1], points[i]
		switch curve {
		case CurveStepBefore:
			result = append(result, fyne.NewPos(current.X, previous.Y))
		case CurveStepAfter:
			result = append(result, fyne.NewPos(previous.X, current.Y))
		default:
			middle := (previous.X + current.X) / 2
			result = append(result, fyne.NewPos(middle, previous.Y), fyne.NewPos(middle, current.Y))
		}
		result = append(result, current)
	}
	return result
}

func monotonePoints(points []fyne.Position, samples int) []fyne.Position {
	if len(points) < 3 {
		return append([]fyne.Position(nil), points...)
	}
	delta := make([]float32, len(points)-1)
	tangent := make([]float32, len(points))
	for i := range delta {
		dx := points[i+1].X - points[i].X
		if dx == 0 {
			delta[i] = 0
		} else {
			delta[i] = (points[i+1].Y - points[i].Y) / dx
		}
	}
	tangent[0], tangent[len(tangent)-1] = delta[0], delta[len(delta)-1]
	for i := 1; i < len(points)-1; i++ {
		hPrevious := points[i].X - points[i-1].X
		hNext := points[i+1].X - points[i].X
		weighted := (delta[i-1]*hNext + delta[i]*hPrevious) / (hPrevious + hNext)
		tangent[i] = (sign(delta[i-1]) + sign(delta[i])) * min(
			float32(math.Abs(float64(delta[i-1]))),
			float32(math.Abs(float64(delta[i]))),
			0.5*float32(math.Abs(float64(weighted))),
		)
	}
	result := make([]fyne.Position, 0, (len(points)-1)*samples+1)
	result = append(result, points[0])
	for i := 0; i < len(points)-1; i++ {
		dx := points[i+1].X - points[i].X
		if dx <= 0 {
			return append([]fyne.Position(nil), points...)
		}
		for sample := 1; sample <= samples; sample++ {
			t := float32(sample) / float32(samples)
			t2, t3 := t*t, t*t*t
			h00 := 2*t3 - 3*t2 + 1
			h10 := t3 - 2*t2 + t
			h01 := -2*t3 + 3*t2
			h11 := t3 - t2
			x := points[i].X + t*dx
			y := h00*points[i].Y + h10*dx*tangent[i] + h01*points[i+1].Y + h11*dx*tangent[i+1]
			if math.IsNaN(float64(y)) {
				y = points[i].Y + t*(points[i+1].Y-points[i].Y)
			}
			y = min(max(y, min(points[i].Y, points[i+1].Y)), max(points[i].Y, points[i+1].Y))
			result = append(result, fyne.NewPos(x, y))
		}
	}
	return result
}

func sign(value float32) float32 {
	if value < 0 {
		return -1
	}
	if value > 0 {
		return 1
	}
	return 0
}
