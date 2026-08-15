package reports

import (
	"fmt"
	"html"
	"math"
	"strings"
)

type Node struct {
	ID, Label           string
	X, Y, Width, Height float64
}
type Edge struct{ Source, Target, Label, Color string }

func DiagramSVG(nodes []Node, edges []Edge) string {
	if len(nodes) == 0 {
		return `<svg xmlns="http://www.w3.org/2000/svg" width="800" height="600" viewBox="0 0 800 600"><rect width="100%" height="100%" fill="#0b1525"/><text x="400" y="300" fill="#8fa3bd" text-anchor="middle" font-family="system-ui,sans-serif" font-size="16">Nenhum elemento no diagrama</text></svg>`
	}

	minX, minY := math.MaxFloat64, math.MaxFloat64
	maxX, maxY := -math.MaxFloat64, -math.MaxFloat64

	byID := map[string]Node{}
	for _, node := range nodes {
		w := node.Width
		if w <= 0 {
			w = 190
		}
		h := node.Height
		if h <= 0 {
			h = 76
		}
		node.Width = w
		node.Height = h
		byID[node.ID] = node

		if node.X < minX {
			minX = node.X
		}
		if node.Y < minY {
			minY = node.Y
		}
		if node.X+w > maxX {
			maxX = node.X + w
		}
		if node.Y+h > maxY {
			maxY = node.Y + h
		}
	}

	padding := 50.0
	minX -= padding
	minY -= padding
	maxX += padding
	maxY += padding

	totalWidth := maxX - minX
	totalHeight := maxY - minY
	if totalWidth < 600 {
		totalWidth = 600
	}
	if totalHeight < 400 {
		totalHeight = 400
	}

	var b strings.Builder
	fmt.Fprintf(&b, `<svg xmlns="http://www.w3.org/2000/svg" width="%.0f" height="%.0f" viewBox="%.0f %.0f %.0f %.0f">`, totalWidth, totalHeight, minX, minY, totalWidth, totalHeight)
	fmt.Fprintf(&b, `<rect x="%.0f" y="%.0f" width="%.0f" height="%.0f" fill="#08111f"/>`, minX, minY, totalWidth, totalHeight)
	b.WriteString(`<defs><style>
text{font-family:system-ui,-apple-system,sans-serif}
.node-card{fill:#101d30;stroke:#253750;stroke-width:1.5px;rx:8px}
.node-title{font-size:13px;font-weight:700;fill:#dbe5f4}
.edge-line{stroke-width:2.5px;stroke-linecap:round}
.edge-bg{fill:#15243a;stroke:#253750;stroke-width:1px;rx:4px}
.edge-label{font-size:10px;font-weight:600;fill:#8fa3bd;text-anchor:middle;dominant-baseline:central}
</style></defs>`)

	// Conexões (Edges)
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
			color = "#3b82f6"
		}
		fmt.Fprintf(&b, `<line class="edge-line" x1="%.1f" y1="%.1f" x2="%.1f" y2="%.1f" stroke="%s"/>`, x1, y1, x2, y2, html.EscapeString(color))

		if edge.Label != "" {
			midX, midY := (x1+x2)/2, (y1+y2)/2
			labelLen := float64(len(edge.Label)) * 6.5
			boxW := labelLen + 14
			boxH := 18.0
			fmt.Fprintf(&b, `<rect class="edge-bg" x="%.1f" y="%.1f" width="%.1f" height="%.1f"/>`, midX-boxW/2, midY-boxH/2, boxW, boxH)
			fmt.Fprintf(&b, `<text class="edge-label" x="%.1f" y="%.1f">%s</text>`, midX, midY, html.EscapeString(edge.Label))
		}
	}

	// Elementos (Nodes)
	for _, node := range nodes {
		fmt.Fprintf(&b, `<rect class="node-card" x="%.1f" y="%.1f" width="%.1f" height="%.1f"/>`, node.X, node.Y, node.Width, node.Height)
		// Barra de destaque superior
		fmt.Fprintf(&b, `<rect x="%.1f" y="%.1f" width="%.1f" height="4" fill="#3b82f6" rx="2"/>`, node.X+2, node.Y+1, node.Width-4)
		// Texto do Nó
		fmt.Fprintf(&b, `<text class="node-title" x="%.1f" y="%.1f">%s</text>`, node.X+14, node.Y+32, html.EscapeString(node.Label))
	}

	b.WriteString("</svg>")
	return b.String()
}

