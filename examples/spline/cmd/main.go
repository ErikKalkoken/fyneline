package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"

	"github.com/nathabonfim59/fyneline/examples/spline"
)

func main() {
	a := app.NewWithID("io.fyne.fyneline.examples.spline")
	w := a.NewWindow("Fyneline Spline Examples")
	w.SetContent(spline.Gallery())
	w.Resize(fyne.NewSize(1000, 720))
	w.ShowAndRun()
}
