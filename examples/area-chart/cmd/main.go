package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"

	areachart "github.com/nathabonfim59/fyneline/examples/area-chart"
)

func main() {
	a := app.NewWithID("io.fyne.fyneline.examples.area")
	w := a.NewWindow("Fyneline AreaChart Examples")
	w.SetContent(areachart.Gallery())
	w.Resize(fyne.NewSize(1000, 720))
	w.ShowAndRun()
}
