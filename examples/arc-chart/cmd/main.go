package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"

	arcchart "github.com/nathabonfim59/fyneline/examples/arc-chart"
)

func main() {
	a := app.NewWithID("io.fyne.fyneline.examples.arc")
	w := a.NewWindow("Fyneline ArcChart Examples")
	w.SetContent(arcchart.Gallery())
	w.Resize(fyne.NewSize(1000, 720))
	w.ShowAndRun()
}
