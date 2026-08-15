package reports

import (
	"fmt"
	"html"
	"strings"
)

type Node struct {
	ID, Label           string
	X, Y, Width, Height float64
}
type Edge struct{ Source, Target, Label, Color string }

func DiagramSVG(nodes []Node, edges []Edge) string {
	maxX, maxY := 800.0, 600.0
	byID := map[string]Node{}
	for _, node := range nodes {
		byID[node.ID] = node
		if node.X+node.Width+40 > maxX {
			maxX = node.X + node.Width + 40
		}
		if node.Y+node.Height+40 > maxY {
			maxY = node.Y + node.Height + 40
		}
	}
	var b strings.Builder
	fmt.Fprintf(&b, `<svg xmlns="http://www.w3.org/2000/svg" width="%.0f" height="%.0f" viewBox="0 0 %.0f %.0f"><rect width="100%%" height="100%%" fill="#f8fafc"/><style>text{font-family:Arial,sans-serif;fill:#172033}.label{font-size:14px;font-weight:bold}.edge{font-size:11px}</style>`, maxX, maxY, maxX, maxY)
	for _, edge := range edges {
		source, ok1 := byID[edge.Source]
		target, ok2 := byID[edge.Target]
		if !ok1 || !ok2 {
			continue
		}
		x1, y1 := source.X+source.Width/2, source.Y+source.Height/2
		x2, y2 := target.X+target.Width/2, target.Y+target.Height/2
		color := edge.Color
		if color == "" {
			color = "#64748b"
		}
		fmt.Fprintf(&b, `<line x1="%.0f" y1="%.0f" x2="%.0f" y2="%.0f" stroke="%s" stroke-width="2"/><text class="edge" x="%.0f" y="%.0f">%s</text>`, x1, y1, x2, y2, html.EscapeString(color), (x1+x2)/2, (y1+y2)/2, html.EscapeString(edge.Label))
	}
	for _, node := range nodes {
		fmt.Fprintf(&b, `<rect x="%.0f" y="%.0f" width="%.0f" height="%.0f" rx="8" fill="#fff" stroke="#94a3b8"/><text class="label" x="%.0f" y="%.0f">%s</text>`, node.X, node.Y, node.Width, node.Height, node.X+12, node.Y+28, html.EscapeString(node.Label))
	}
	b.WriteString("</svg>")
	return b.String()
}
