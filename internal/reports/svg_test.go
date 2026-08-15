package reports

import (
	"strings"
	"testing"
)

func TestDiagramSVG(t *testing.T) {
	svg := DiagramSVG([]Node{{ID: "a", Label: "Switch", Width: 180, Height: 100}}, nil)
	if !strings.Contains(svg, "Switch") || !strings.HasPrefix(svg, "<svg") {
		t.Fatal("invalid SVG")
	}
}
