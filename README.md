# Fyneline

Typed, responsive charts for [Fyne](https://fyne.io/), with a chainable API
inspired by [LayerChart](https://www.layerchart.com/).

```bash
go get github.com/nathabonfim59/fyneline
```

Fyneline currently provides `BarChart`, `AreaChart`, `ArcChart`, and `Spline`.
Each is a regular `fyne.Widget` that can be passed directly to
`Window.SetContent` or any Fyne container.

## BarChart

```go
type Sale struct {
    Month   string
    Online  float64
    InStore float64
}

chart := fyneline.NewBarChart(
    sales,
    func(s Sale) string { return s.Month },
    fyneline.NewBarSeries("Online", func(s Sale) float64 { return s.Online }),
    fyneline.NewBarSeries("In store", func(s Sale) float64 { return s.InStore }),
).
    SetSeriesLayout(fyneline.SeriesGroup).
    SetBandPadding(0.4).
    SetValueAxis(fyneline.NewNumericAxis().WithDomain(0, 100))
```

Bar series support `SeriesOverlap`, `SeriesGroup`, `SeriesStack`,
`SeriesStackExpand`, and `SeriesStackDiverging`, plus vertical and horizontal
orientations.

## AreaChart

```go
chart := fyneline.NewAreaChart(
    readings,
    fyneline.TimeAccessor(func(r Reading) time.Time { return r.Time }),
    fyneline.NewAreaSeries("Temperature", func(r Reading) float64 { return r.Value }),
).
    SetCurve(fyneline.CurveMonotoneX).
    SetXAxis(fyneline.NewTimeAxis("15:04", time.Local))
```

Area charts share the spline curve and gap policies and support overlapping,
stacked, normalized, and diverging series.

## ArcChart

```go
chart := fyneline.NewArcChart(
    segments,
    func(s Segment) float64 { return s.Value },
    func(s Segment) string { return s.Name },
).
    SetInnerRadius(0.58).
    SetRange(-120, 120).
    SetPadAngle(2).
    SetLabels(true)
```

Angles are degrees with zero at the top and positive values moving clockwise.
Use `SetMaxValue` for partially filled gauges.

## Spline

```go
chart := fyneline.NewSpline(
    points,
    func(p Point) float64 { return p.X },
    fyneline.NewSplineSeries("Value", func(p Point) float64 { return p.Y }),
).SetCurve(fyneline.CurveMonotoneX)
```

Available curves are linear, monotone-x, step, step-before, and step-after.
Optional series accessors can break, connect, or replace missing values with
zero.

## Live updates

Chart setters refresh immediately and return the receiver for chaining. Input
slices are copied. As with other Fyne widgets, schedule updates originating in
background goroutines on Fyne's UI goroutine:

```go
fyne.Do(func() {
    chart.SetData(updated)
})
```

## Documentation

Every exported API is documented for `go doc`:

```bash
go doc github.com/nathabonfim59/fyneline
go doc github.com/nathabonfim59/fyneline.BarChart
go doc github.com/nathabonfim59/fyneline.NewArcChart
```

## Requirements

- Go 1.22 or later
- Fyne 2.8 or later

Fyneline is available under the MIT License.
