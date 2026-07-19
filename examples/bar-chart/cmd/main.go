package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"

	barchart "github.com/nathabonfim59/fyneline/examples/bar-chart"
)

func main() {
	a := app.NewWithID("io.fyne.fyneline.examples.bar")
	w := a.NewWindow("Fyneline BarChart Examples")
	w.SetContent(barchart.Gallery())
	w.Resize(fyne.NewSize(1000, 720))
	w.ShowAndRun()
}
