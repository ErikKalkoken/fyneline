package fyneline

import (
	"image/color"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/test"
)

type testDatum struct {
	category string
	x, a, b  float64
	valid    bool
}

func TestBarChartRendersGroupedBars(t *testing.T) {
	test.NewTempApp(t)
	data := []testDatum{
		{category: "A", a: 2, b: 3},
		{category: "B", a: 4, b: 1},
		{category: "C", a: 3, b: 5},
	}
	chart := NewBarChart(data,
		func(d testDatum) string { return d.category },
		NewBarSeries("A", func(d testDatum) float64 { return d.a }),
		NewBarSeries("B", func(d testDatum) float64 { return d.b }),
	).SetSeriesLayout(SeriesGroup)

	objects := renderObjects(t, chart)
	if got := countObjects[*canvas.Rectangle](objects); got != 6 {
		t.Fatalf("rendered %d bars, want 6", got)
	}
}

func TestBarChartExplicitEmptyCategoryDoesNotCallAccessor(t *testing.T) {
	test.NewTempApp(t)
	data := []testDatum{{category: "A", a: 2}}
	chart := NewBarChart(data,
		func(d testDatum) string { return d.category },
		NewBarSeries("A", func(d testDatum) float64 {
			if d.category == "" {
				t.Fatal("value accessor called for absent category")
			}
			return d.a
		}),
	).SetCategoryAxis(NewCategoryAxis().WithCategories("A", "Missing").WithGrid(true))

	objects := renderObjects(t, chart)
	if got := countObjects[*canvas.Rectangle](objects); got != 1 {
		t.Fatalf("rendered %d bars, want 1", got)
	}
}

func TestAreaChartRendersFilledIntervals(t *testing.T) {
	test.NewTempApp(t)
	data := []testDatum{{x: 0, a: 1}, {x: 1, a: 3}, {x: 2, a: 2}}
	chart := NewAreaChart(data,
		func(d testDatum) float64 { return d.x },
		NewAreaSeries("A", func(d testDatum) float64 { return d.a }),
	)

	objects := renderObjects(t, chart)
	if got := countObjects[*canvas.ArbitraryPolygon](objects); got != 2 {
		t.Fatalf("rendered %d area intervals, want 2", got)
	}
}

func TestArcChartSkipsNonPositiveValuesAndRendersLabels(t *testing.T) {
	test.NewTempApp(t)
	data := []testDatum{
		{category: "A", a: 2},
		{category: "B", a: 0},
		{category: "C", a: 3},
	}
	chart := NewArcChart(data,
		func(d testDatum) float64 { return d.a },
		func(d testDatum) string { return d.category },
	).SetLabels(true)

	objects := renderObjects(t, chart)
	if got := countObjects[*canvas.Arc](objects); got != 2 {
		t.Fatalf("rendered %d arcs, want 2", got)
	}
	if got := countObjects[*canvas.Text](objects); got != 2 {
		t.Fatalf("rendered %d labels, want 2", got)
	}
}

func TestSplineRendersMonotoneSegments(t *testing.T) {
	test.NewTempApp(t)
	data := []testDatum{{x: 0, a: 1}, {x: 1, a: 3}, {x: 2, a: 2}}
	stroke := color.NRGBA{R: 1, G: 2, B: 3, A: 255}
	chart := NewSpline(data,
		func(d testDatum) float64 { return d.x },
		NewSplineSeries("A", func(d testDatum) float64 { return d.a }).WithStyle(SplineStyle{
			Stroke: StrokeStyle{Color: stroke, Width: 3},
		}),
	).SetCurve(CurveMonotoneX)

	objects := renderObjects(t, chart)
	count := 0
	for _, object := range objects {
		line, ok := object.(*canvas.Line)
		if ok && line.StrokeWidth == 3 {
			count++
		}
	}
	if count != 20 {
		t.Fatalf("rendered %d spline segments, want 20", count)
	}
}

func TestOptionalSeriesBreaksAtGap(t *testing.T) {
	test.NewTempApp(t)
	data := []testDatum{
		{x: 0, a: 1, valid: true},
		{x: 1, a: 2, valid: false},
		{x: 2, a: 3, valid: true},
	}
	chart := NewSpline(data,
		func(d testDatum) float64 { return d.x },
		NewOptionalSplineSeries("A", func(d testDatum) (float64, bool) { return d.a, d.valid }).WithWidth(4),
	)

	objects := renderObjects(t, chart)
	for _, object := range objects {
		if line, ok := object.(*canvas.Line); ok && line.StrokeWidth == 4 {
			t.Fatal("gap should leave singleton segments with no rendered line")
		}
	}
}

func TestSetDataCopiesInput(t *testing.T) {
	test.NewTempApp(t)
	data := []testDatum{{category: "A", a: 1}}
	chart := NewBarChart(data,
		func(d testDatum) string { return d.category },
		NewBarSeries("A", func(d testDatum) float64 { return d.a }),
	)
	data[0].a = 99
	if chart.data[0].a != 1 {
		t.Fatalf("constructor retained caller slice: got %v", chart.data[0].a)
	}

	replacement := []testDatum{{category: "B", a: 2}}
	chart.SetData(replacement)
	replacement[0].a = 88
	if chart.data[0].a != 2 {
		t.Fatalf("SetData retained caller slice: got %v", chart.data[0].a)
	}
}

func TestStackLayouts(t *testing.T) {
	values := [][]float64{{2, -2}, {3, -3}}
	defined := [][]bool{{true, true}, {true, true}}

	lower, upper := barStack(values, defined, SeriesStack)
	if lower[1][0] != 2 || upper[1][0] != 5 {
		t.Fatalf("ordinary stack = %v..%v, want 2..5", lower[1][0], upper[1][0])
	}

	lower, upper = barStack(values, defined, SeriesStackDiverging)
	if lower[1][1] != -2 || upper[1][1] != -5 {
		t.Fatalf("diverging stack = %v..%v, want -2..-5", lower[1][1], upper[1][1])
	}

	lower, upper = barStack(values, defined, SeriesStackExpand)
	if upper[1][0] != 1 {
		t.Fatalf("expanded stack ends at %v, want 1", upper[1][0])
	}
}

func TestCurvePoints(t *testing.T) {
	points := []fyne.Position{fyne.NewPos(0, 10), fyne.NewPos(10, 0), fyne.NewPos(20, 10)}
	if got := len(curvePoints(points, CurveLinear)); got != 3 {
		t.Fatalf("linear point count = %d, want 3", got)
	}
	if got := len(curvePoints(points, CurveStep)); got != 7 {
		t.Fatalf("step point count = %d, want 7", got)
	}
	if got := len(curvePoints(points, CurveMonotoneX)); got != 21 {
		t.Fatalf("monotone point count = %d, want 21", got)
	}
}

func renderObjects(t *testing.T, chart fyne.Widget) []fyne.CanvasObject {
	t.Helper()
	renderer := test.TempWidgetRenderer(t, chart)
	renderer.Layout(fyne.NewSize(400, 240))
	objects := renderer.Objects()
	if len(objects) != 1 {
		t.Fatalf("renderer returned %d root objects, want 1", len(objects))
	}
	container, ok := objects[0].(*fyne.Container)
	if !ok {
		t.Fatalf("renderer root is %T, want *fyne.Container", objects[0])
	}
	return container.Objects
}

func countObjects[T fyne.CanvasObject](objects []fyne.CanvasObject) int {
	count := 0
	for _, object := range objects {
		if _, ok := object.(T); ok {
			count++
		}
	}
	return count
}
