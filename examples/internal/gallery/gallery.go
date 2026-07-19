// Package gallery contains presentation helpers shared by the example apps.
package gallery

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// Card wraps a chart with a title and a short option summary.
func Card(title, summary string, chart fyne.CanvasObject) fyne.CanvasObject {
	return widget.NewCard(title, summary, chart)
}

// Grid arranges example cards in two columns.
func Grid(cards ...fyne.CanvasObject) fyne.CanvasObject {
	return container.NewGridWithColumns(2, cards...)
}

// Fixed gives content a deterministic size for screenshot rendering.
func Fixed(content fyne.CanvasObject) fyne.CanvasObject {
	return container.NewGridWrap(fyne.NewSize(1000, 720), content)
}
