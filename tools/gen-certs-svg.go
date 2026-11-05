package main

import (
	"encoding/base64"
	"fmt"
	"math"
	"os"
	"strings"
)

// Status values.
const (
	Planned    = 0
	InProgress = 1
	Earned     = 2
)

var (
	// Fill opacity per status — hex color stays constant, transparency encodes progress.
	fillOpacity  = [3]float64{0.20, 0.60, 1.0}
	statusLabels = [3]string{"Planned", "In Progress", "Earned"}
)

type cert struct {
	lines     []string // text rendered inside the hexagon
	fill      string   // hex base color
	iconFile  string   // path relative to project root; empty = no icon
	iconEmoji string   // emoji rendered in icon slot when iconFile is empty
	status    int
}

// Update status values as certs are earned.
var certs = []cert{
	{[]string{"CKA"}, "#326CE5", "assets/kubernetes.svg", "", Planned},
	{[]string{"CKAD"}, "#326CE5", "assets/kubernetes.svg", "", Planned},
	{[]string{"CKS"}, "#326CE5", "assets/kubernetes.svg", "", Planned},
	{[]string{"LFCS"}, "#003580", "assets/linux.svg", "", Planned},
	{[]string{"Canonical", "SysAdmin"}, "#E95420", "assets/ubuntu.svg", "", Planned},
	{[]string{"Ubuntu", "Member"}, "#E95420", "assets/ubuntu.svg", "", Planned},
	{[]string{"CODE", "(LF)"}, "#003580", "assets/linuxfoundation.svg", "", Planned},
	{[]string{"BTL1"}, "#1A1A2E", "assets/tryhackme.svg", "", Planned},
	{[]string{"Security+"}, "#C8202E", "assets/comptia.svg", "", Planned},
	{[]string{"CySA+"}, "#C8202E", "assets/comptia.svg", "", Planned},
	{[]string{"ISO 27001"}, "#003580", "", "🔒", Planned},
	{[]string{"DevSecOps", "Foundation"}, "#1A1A2E", "assets/owasp-white.svg", "", Planned},
	{[]string{"CISSP"}, "#006400", "assets/isc2.svg", "", Planned},
	{[]string{"Scrumban"}, "#0052CC", "assets/jira.svg", "", Planned},
}

// layout[i] = number of hexagons in row i (top to bottom).
// 4-3-4-3 packs 14 certs into a symmetric honeycomb.
var layout = []int{4, 3, 4, 3}

// hexPoints returns the SVG polygon points for a pointy-top hexagon
// centered at (cx, cy) with circumradius r.
func hexPoints(cx, cy, r float64) string {
	pts := make([]string, 6)
	for i := range pts {
		a := math.Pi / 180.0 * (60.0*float64(i) + 30.0)
		pts[i] = fmt.Sprintf("%.2f,%.2f", cx+r*math.Cos(a), cy+r*math.Sin(a))
	}
	return strings.Join(pts, " ")
}

// iconDataURI reads an SVG file and returns a data URI safe for embedding in <image href>.
// Returns empty string if the file cannot be read.
func iconDataURI(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return "data:image/svg+xml;base64," + base64.StdEncoding.EncodeToString(data)
}

func main() {
	const (
		r        = 70.0     // circumradius (center to corner)
		drawR    = r * 0.93 // drawn radius — creates a ~5px gap between adjacent hexes
		vSpacing = r * 1.5  // vertical center-to-center distance between rows
		margin   = 24.0     // outer padding
		legendH  = 48.0     // height reserved for the legend below the grid
		lineH    = 15.0     // line height for multi-line cert labels
		lineHH   = 7.0      // half line height (approx half cap height)
		iconH    = 32.0     // icon bounding box height
		iconW    = 32.0     // icon bounding box width
		iconGap  = 5.0      // gap between icon bottom and first text line top
		fontBase = 13       // base font size
		fontSm   = 11       // reduced size for labels longer than 7 chars
	)

	hw := math.Sqrt(3) * r // horizontal center-to-center spacing within a row

	svgW := 4*hw + 2*margin
	svgH := margin + 2*r + float64(len(layout)-1)*vSpacing + margin + legendH

	// x0: center of the leftmost hex in a 4-hex (even) row.
	x0 := margin + hw/2

	// Pre-load icons, deduplicating by path.
	iconCache := map[string]string{}
	for _, c := range certs {
		if c.iconFile != "" {
			if _, ok := iconCache[c.iconFile]; !ok {
				iconCache[c.iconFile] = iconDataURI(c.iconFile)
			}
		}
	}

	// Build per-color hatch pattern IDs for In Progress hexes.
	// The legend swatch always gets a pattern regardless of cert statuses.
	const legendRefColor = "#326CE5"
	hatchIDs := map[string]string{legendRefColor: "hatchLegend"}
	for _, c := range certs {
		if c.status == InProgress {
			if _, ok := hatchIDs[c.fill]; !ok {
				hatchIDs[c.fill] = "hatch" + strings.TrimPrefix(c.fill, "#")
			}
		}
	}

	var b strings.Builder

	b.WriteString(fmt.Sprintf(
		"<svg xmlns=\"http://www.w3.org/2000/svg\" width=\"%.0f\" height=\"%.0f\" viewBox=\"0 0 %.0f %.0f\">\n",
		svgW, svgH, svgW, svgH,
	))

	b.WriteString("  <defs>\n")
	b.WriteString("    <filter id=\"ts\" x=\"-30%\" y=\"-30%\" width=\"160%\" height=\"160%\">\n")
	b.WriteString("      <feDropShadow dx=\"0\" dy=\"0\" stdDeviation=\"2.5\" flood-color=\"black\" flood-opacity=\"0.9\"/>\n")
	b.WriteString("    </filter>\n")
	// Per-color diagonal stripe patterns for In Progress hexes.
	// Each pattern alternates opaque hex-color lines with transparent gaps (\ direction).
	// Base polygon drawn at 65% opacity; stripe lines at 100% — same color, brighter bands.
	for color, id := range hatchIDs {
		fmt.Fprintf(&b, "    <pattern id=%q patternUnits=\"userSpaceOnUse\" width=\"12\" height=\"12\" patternTransform=\"rotate(-45 0 0)\">\n", id)
		fmt.Fprintf(&b, "      <line x1=\"0\" y1=\"0\" x2=\"0\" y2=\"12\" stroke=%q stroke-width=\"6\"/>\n", color)
		b.WriteString("    </pattern>\n")
	}
	b.WriteString("  </defs>\n")

	ci := 0
	for row, count := range layout {
		cy := margin + r + float64(row)*vSpacing

		// Odd rows (3-hex) offset by hw/2 so they nest between even rows.
		startX := x0
		if count == 3 {
			startX = x0 + hw/2
		}

		for col := 0; col < count && ci < len(certs); col++ {
			c := certs[ci]
			ci++
			cx := startX + float64(col)*hw

			// Hexagon body.
			// In Progress: base at 65% opacity + same-color stripe overlay at 100% (brighter bands).
			// Planned: low opacity. Earned: solid, no overlay.
			baseOpacity := fillOpacity[c.status]
			if c.status == InProgress {
				baseOpacity = 0.65
			}
			fmt.Fprintf(&b,
				"  <polygon points=\"%s\" fill=\"%s\" fill-opacity=\"%.2f\" stroke=\"white\" stroke-opacity=\"0.5\" stroke-width=\"2\" stroke-linejoin=\"round\"/>\n",
				hexPoints(cx, cy, drawR), c.fill, baseOpacity,
			)
			if c.status == InProgress {
				fmt.Fprintf(&b,
					"  <polygon points=\"%s\" fill=\"url(#%s)\"/>\n",
					hexPoints(cx, cy, drawR), hatchIDs[c.fill],
				)
			}

			// Determine whether this cert has a usable image icon or an emoji icon.
			nLines := len(c.lines)
			iconURI := ""
			if c.iconFile != "" {
				iconURI = iconCache[c.iconFile]
			}
			hasIcon := iconURI != "" || c.iconEmoji != ""

			// Compute content block height and center it at cy.
			// textBlockH = (nLines-1)*lineH + 2*lineHH  (top/bottom half-line to center)
			textBlockH := float64(nLines-1)*lineH + 2*lineHH
			var blockH float64
			if hasIcon {
				blockH = iconH + iconGap + textBlockH
			} else {
				blockH = textBlockH
			}
			blockTop := cy - blockH/2

			// Icon slot: SVG image or emoji text.
			if iconURI != "" {
				fmt.Fprintf(&b,
					"  <image href=\"%s\" x=\"%.2f\" y=\"%.2f\" width=\"%.0f\" height=\"%.0f\"/>\n",
					iconURI, cx-iconW/2, blockTop, iconW, iconH,
				)
			} else if c.iconEmoji != "" {
				fmt.Fprintf(&b,
					"  <text x=\"%.2f\" y=\"%.2f\" font-size=\"22\" text-anchor=\"middle\" dominant-baseline=\"central\">%s</text>\n",
					cx, blockTop+iconH/2, c.iconEmoji,
				)
			}

			// Text lines.
			textAreaTop := blockTop
			if hasIcon {
				textAreaTop = blockTop + iconH + iconGap
			}
			firstTextY := textAreaTop + lineHH

			for li, line := range c.lines {
				ly := firstTextY + float64(li)*lineH
				fs := fontBase
				if len(line) > 7 {
					fs = fontSm
				}
				fmt.Fprintf(&b,
					"  <text x=\"%.2f\" y=\"%.2f\" font-family=\"ui-sans-serif,system-ui,sans-serif\" font-size=\"%d\" font-weight=\"700\" fill=\"white\" text-anchor=\"middle\" dominant-baseline=\"central\" filter=\"url(#ts)\">%s</text>\n",
					cx, ly, fs, line,
				)
			}
		}
	}

	// Legend row — sample rect uses cert color at each opacity level.
	legendY := svgH - legendH/2
	lx := (svgW - 3*110.0) / 2
	for s := 0; s < 3; s++ {
		rectOpacity := fillOpacity[s]
		fmt.Fprintf(&b,
			"  <rect x=\"%.1f\" y=\"%.1f\" width=\"12\" height=\"12\" fill=\"%s\" fill-opacity=\"%.2f\" stroke=\"%s\" stroke-width=\"1.5\" rx=\"2\"/>\n",
			lx, legendY-6, legendRefColor, rectOpacity, legendRefColor,
		)
		if s == InProgress {
			// 3 explicit \ lines dividing the 12×12 swatch into 4 equal bands — no clip-path needed.
			// Line coords are clamped so they stay inside the swatch rect.
			fmt.Fprintf(&b, "  <line x1=\"%.1f\" y1=\"%.1f\" x2=\"%.1f\" y2=\"%.1f\" stroke=\"black\" stroke-width=\"1.5\"/>\n", lx, legendY, lx+6, legendY+6)
			fmt.Fprintf(&b, "  <line x1=\"%.1f\" y1=\"%.1f\" x2=\"%.1f\" y2=\"%.1f\" stroke=\"black\" stroke-width=\"1.5\"/>\n", lx, legendY-6, lx+12, legendY+6)
			fmt.Fprintf(&b, "  <line x1=\"%.1f\" y1=\"%.1f\" x2=\"%.1f\" y2=\"%.1f\" stroke=\"black\" stroke-width=\"1.5\"/>\n", lx+6, legendY-6, lx+12, legendY)
		}
		fmt.Fprintf(&b,
			"  <text x=\"%.1f\" y=\"%.1f\" font-family=\"ui-sans-serif,system-ui,sans-serif\" font-size=\"12\" fill=\"#AAAAAA\" text-anchor=\"start\" dominant-baseline=\"central\">%s</text>\n",
			lx+16, legendY, statusLabels[s],
		)
		lx += 110
	}

	b.WriteString("</svg>\n")

	out := "assets/certs-honeycomb.svg"
	if len(os.Args) > 1 {
		out = os.Args[1]
	}

	if err := os.WriteFile(out, []byte(b.String()), 0644); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println("wrote", out)
}
