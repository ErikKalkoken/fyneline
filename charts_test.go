package fyneline

import (
	"image/color"
	"math"
	"testing"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/driver/software"
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
	app := test.NewTempApp(t)
	data := []testDatum{{x: 0, a: 1}, {x: 1, a: 3}, {x: 2, a: 2}}
	fill := color.NRGBA{R: 240, G: 10, B: 20, A: 255}
	chart := NewAreaChart(data,
		func(d testDatum) float64 { return d.x },
		NewAreaSeries("A", func(d testDatum) float64 { return d.a }).WithStyle(AreaStyle{
			Fill: FillStyle{Color: fill, Opacity: 1},
		}),
	)

	objects := renderObjects(t, chart)
	if got := countObjects[*canvas.ArbitraryPolygon](objects); got != 2 {
		t.Fatalf("rendered %d area intervals, want 2", got)
	}
	for _, object := range objects {
		if polygon, ok := object.(*canvas.ArbitraryPolygon); ok && (polygon.Size().Width == 0 || polygon.Size().Height == 0) {
			t.Fatalf("area polygon has empty bounds: %v", polygon.Size())
		}
	}
	image := software.Render(chart, app.Settings().Theme())
	filledPixels := 0
	for y := image.Bounds().Min.Y; y < image.Bounds().Max.Y; y++ {
		for x := image.Bounds().Min.X; x < image.Bounds().Max.X; x++ {
			r, g, b, _ := image.At(x, y).RGBA()
			if r > 0xc000 && g < 0x4000 && b < 0x4000 {
				filledPixels++
			}
		}
	}
	if filledPixels < 100 {
		t.Fatalf("area fill painted only %d pixels", filledPixels)
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
	for _, object := range objects {
		arc, ok := object.(*canvas.Arc)
		if ok && arc.Position().X+arc.Size().Width > 400 {
			t.Fatalf("arc exceeds chart bounds: position=%v size=%v", arc.Position(), arc.Size())
		}
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

func TestUndefinedStackValueDoesNotAdvanceBaseline(t *testing.T) {
	values := [][]float64{{100}, {3}}
	defined := [][]bool{{false}, {true}}
	for _, layout := range []SeriesLayout{SeriesStack, SeriesStackExpand, SeriesStackDiverging} {
		lower, upper := barStack(values, defined, layout)
		if lower[1][0] != 0 {
			t.Fatalf("layout %v baseline = %v, want 0", layout, lower[1][0])
		}
		if layout != SeriesStackExpand && upper[1][0] != 3 {
			t.Fatalf("layout %v upper = %v, want 3", layout, upper[1][0])
		}
	}
}

func TestGroupedAndStackPaddedBarsRemainCentered(t *testing.T) {
	test.NewTempApp(t)
	data := []testDatum{{category: "A", a: 2, b: 3}}
	for _, layout := range []SeriesLayout{SeriesGroup, SeriesStack} {
		chart := NewBarChart(data,
			func(d testDatum) string { return d.category },
			NewBarSeries("A", func(d testDatum) float64 { return d.a }),
			NewBarSeries("B", func(d testDatum) float64 { return d.b }),
		).SetSeriesLayout(layout).SetGroupPadding(0.2).SetStackPadding(0.2)
		objects := renderObjects(t, chart)
		left, right := float32(math.Inf(1)), float32(math.Inf(-1))
		for _, object := range objects {
			if bar, ok := object.(*canvas.Rectangle); ok {
				left = min(left, bar.Position().X)
				right = max(right, bar.Position().X+bar.Size().Width)
			}
		}
		plot := cartesianPlot(fyne.NewSize(400, 240), Insets{}, true, true)
		plot = fitCategoryAxisMargin(chart, plot, chart.categoryAxis, []string{"A"}, true)
		plot = fitNumericAxisMargin(chart, plot, chart.valueAxis, Domain{Min: 0, Max: 5}, false)
		if difference := math.Abs(float64((left+right)/2 - (plot.left+plot.right)/2)); difference > 0.01 {
			t.Fatalf("layout %v is off-center by %v", layout, difference)
		}
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

func TestMonotoneCurveDoesNotOvershootUnevenIntervals(t *testing.T) {
	points := []fyne.Position{fyne.NewPos(0, 0), fyne.NewPos(1, 100), fyne.NewPos(100, 101)}
	curved := curvePoints(points, CurveMonotoneX)
	for index, point := range curved {
		minimum, maximum := float32(0), float32(100)
		if index > 10 {
			minimum, maximum = 100, 101
		}
		if point.Y < minimum || point.Y > maximum {
			t.Fatalf("point %d overshot interval: %v not in [%v, %v]", index, point.Y, minimum, maximum)
		}
	}
}

func TestMonotoneCurveDoesNotReverseInsideIncreasingIntervals(t *testing.T) {
	points := []fyne.Position{
		fyne.NewPos(0, 0), fyne.NewPos(1, 19),
		fyne.NewPos(2, 20), fyne.NewPos(3, 39),
	}
	curved := curvePoints(points, CurveMonotoneX)
	for index := 1; index < len(curved); index++ {
		if curved[index].Y < curved[index-1].Y {
			t.Fatalf("curve reversed at point %d: %v then %v", index, curved[index-1].Y, curved[index].Y)
		}
	}
}

func TestTimeAccessorSupportsDatesOutsideUnixNanoRange(t *testing.T) {
	timestamp := time.Date(2500, time.January, 1, 0, 0, 0, 500_000_000, time.UTC)
	accessor := TimeAccessor(func(value time.Time) time.Time { return value })
	want := float64(timestamp.Unix()) + 0.5
	if got := accessor(timestamp); got != want {
		t.Fatalf("TimeAccessor = %v, want %v", got, want)
	}
}

func TestArcSettersRejectNonFiniteGeometry(t *testing.T) {
	test.NewTempApp(t)
	chart := NewArcChart([]float64{1}, func(value float64) float64 { return value }, func(float64) string { return "" }).
		SetInnerRadius(float32(math.NaN())).
		SetOuterRadius(float32(math.Inf(1))).
		SetCornerRadius(float32(math.NaN()))
	if chart.innerRadius != 0 || chart.outerRadius != 1 || chart.cornerRadius != 0 {
		t.Fatalf("invalid geometry retained: inner=%v outer=%v corner=%v", chart.innerRadius, chart.outerRadius, chart.cornerRadius)
	}
}

func TestNumericAxisMarginExpandsForWideLabels(t *testing.T) {
	test.NewTempApp(t)
	chart := NewSpline([]testDatum{{x: 1, a: 1}}, func(d testDatum) float64 { return d.x }, NewSplineSeries("A", func(d testDatum) float64 { return d.a }))
	axis := NewNumericAxis().WithFormatter(func(float64) string { return "a very wide tick label" })
	base := cartesianPlot(fyne.NewSize(400, 240), Insets{}, true, true)
	fitted := fitNumericAxisMargin(chart, base, axis, Domain{Min: 0, Max: 1}, false)
	if fitted.left <= base.left {
		t.Fatalf("left margin did not expand: base=%v fitted=%v", base.left, fitted.left)
	}
}

func TestAxisMarginsCannotInvertPlot(t *testing.T) {
	test.NewTempApp(t)
	chart := NewSpline([]testDatum{{x: 1, a: 1}}, func(d testDatum) float64 { return d.x }, NewSplineSeries("A", func(d testDatum) float64 { return d.a }))
	axis := NewNumericAxis().WithFormatter(func(float64) string {
		return "a label much wider than the entire chart"
	}).WithStyle(AxisStyle{TextSize: 200})
	plot := cartesianPlot(fyne.NewSize(100, 80), Insets{}, true, true)
	plot = fitNumericAxisMargin(chart, plot, axis, Domain{Min: 0, Max: 1}, false)
	plot = fitNumericAxisMargin(chart, plot, axis, Domain{Min: 0, Max: 1}, true)
	if plot.left >= plot.right || plot.top >= plot.bottom {
		t.Fatalf("axis margins inverted plot: %+v", plot)
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
