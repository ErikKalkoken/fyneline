.PHONY: all render arc-chart area-chart bar-chart spline

all: render

# Render all captures, or one with: make render EXAMPLE=bar-chart
render:
	go run ./render $(if $(EXAMPLE),--example $(EXAMPLE),--all)

arc-chart area-chart bar-chart spline:
	$(MAKE) render EXAMPLE=$@
