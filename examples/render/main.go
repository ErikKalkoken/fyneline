// Command render regenerates the screenshots used by the example READMEs.
package main

import (
	"flag"
	"fmt"
	"image/png"
	"os"
	"path/filepath"
	"sort"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/software"
	"fyne.io/fyne/v2/test"

	arcchart "github.com/nathabonfim59/fyneline/examples/arc-chart"
	areachart "github.com/nathabonfim59/fyneline/examples/area-chart"
	barchart "github.com/nathabonfim59/fyneline/examples/bar-chart"
	"github.com/nathabonfim59/fyneline/examples/internal/gallery"
	"github.com/nathabonfim59/fyneline/examples/spline"
)

var examples = map[string]func() fyne.CanvasObject{
	"arc-chart":  arcchart.Gallery,
	"area-chart": areachart.Gallery,
	"bar-chart":  barchart.Gallery,
	"spline":     spline.Gallery,
}

func main() {
	all := flag.Bool("all", false, "render every example")
	example := flag.String("example", "", "render one example by folder name")
	flag.Parse()

	if (*all && *example != "") || (!*all && *example == "") {
		fmt.Fprintln(os.Stderr, "specify exactly one of --all or --example NAME")
		os.Exit(2)
	}

	app := test.NewApp()
	defer app.Quit()

	if *example != "" {
		if err := render(*example, app.Settings().Theme()); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}

	names := make([]string, 0, len(examples))
	for name := range examples {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		if err := render(name, app.Settings().Theme()); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
}

func render(name string, theme fyne.Theme) error {
	build, ok := examples[name]
	if !ok {
		return fmt.Errorf("unknown example %q; choose arc-chart, area-chart, bar-chart, or spline", name)
	}
	output := filepath.Join(name, "screenshot.png")
	file, err := os.Create(output)
	if err != nil {
		return fmt.Errorf("create %s: %w", output, err)
	}
	defer file.Close()

	image := software.Render(gallery.Fixed(build()), theme)
	if err := png.Encode(file, image); err != nil {
		return fmt.Errorf("encode %s: %w", output, err)
	}
	fmt.Printf("rendered %s\n", output)
	return nil
}
