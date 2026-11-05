package main

import (
	"encoding/base64"
	"fmt"
	"os"
	"strings"
)

type stackItem struct {
	label    string
	iconFile string
	emoji    string
	iconW    float64 // 0 = use global default
	iconH    float64 // 0 = use global default
}

var stackItems = []stackItem{
	// Row 0 — Cloud Native (7)
	{"Kubernetes", "assets/kubernetes.svg", "", 0, 0},
	{"Podman", "assets/podman.svg", "", 0, 0},
	{"Crossplane", "assets/crossplane.svg", "", 10, 20},
	{"Cloudflare", "assets/cloudflare.svg", "", 0, 0},
	{"Helm", "assets/helm.svg", "", 0, 0},
	{"Akamai", "assets/akamai.svg", "", 0, 0},
	{"LXD", "assets/linuxcontainers.svg", "", 0, 0},
	// Row 1 — Languages & Databases (6)
	{"Go", "assets/gopher.svg", "", 15, 20},
	{"Rust", "assets/rust.svg", "", 0, 0},
	{"Bash", "assets/bash.svg", "", 0, 0},
	{"WASM", "assets/webassembly.svg", "", 0, 0},
	{"Postgres", "assets/postgresql.svg", "", 0, 0},
	{"Redis", "assets/redis.svg", "", 0, 0},
	// Row 2 — Environments & Tools (5)
	{"Git", "assets/git.svg", "", 0, 0},
	{"Neovim", "assets/neovim.svg", "", 0, 0},
	{"Ubuntu", "assets/ubuntu.svg", "", 0, 0},
	{"Arch", "assets/archlinux.svg", "", 0, 0},
	{"Claude", "assets/claude-code.svg", "", 0, 0},
	// Row 3 — Security & Compliance (4)
	{"ISO 27001", "", "🔒", 0, 0},
	{"CIS", "", "🛡️", 0, 0},
	{"SOC 2", "", "📋", 0, 0},
	{"OWASP", "assets/owasp.svg", "", 0, 0},
	// Row 4 — Forges (2)
	{"GitLab", "assets/gitlab.svg", "", 0, 0},
	{"GitHub", "assets/github.svg", "", 0, 0},
}

type brickRow struct {
	count int
	label string
}

var layout = []brickRow{
	{7, "Cloud Native"},
	{6, "Languages & Databases"},
	{5, "Environments & Tools"},
	{4, "Security & Compliance"},
	{2, "Forges"},
}

func iconDataURI(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return "data:image/svg+xml;base64," + base64.StdEncoding.EncodeToString(data)
}

func main() {
	const (
		bW        = 100.0
		bH        = 54.0
		gap       = 4.0
		rx        = 5.0
		margin    = 24.0
		iconH     = 20.0
		iconW     = 20.0
		iconGap   = 4.0
		fontSize  = 10.0
		labelSize = 11.0
		labelGap  = 5.0 // space between label text and bricks below it
		rowGap    = 10.0 // space between brick bottom and next label
	)

	// Widest row is 7 bricks; centering narrower rows creates natural brick stagger.
	svgW := 7*bW + 6*gap + 2*margin

	// Height: each row = labelSize + labelGap + bH; rows separated by rowGap.
	rowH := labelSize + labelGap + bH
	svgH := float64(len(layout))*rowH + float64(len(layout)-1)*rowGap + 2*margin

	iconCache := map[string]string{}
	for _, it := range stackItems {
		if it.iconFile != "" {
			if _, ok := iconCache[it.iconFile]; !ok {
				iconCache[it.iconFile] = iconDataURI(it.iconFile)
			}
		}
	}

	var b strings.Builder
	fmt.Fprintf(&b,
		"<svg xmlns=\"http://www.w3.org/2000/svg\" width=\"%.0f\" height=\"%.0f\" viewBox=\"0 0 %.0f %.0f\">\n",
		svgW, svgH, svgW, svgH,
	)
	b.WriteString("  <defs>\n")
	b.WriteString("    <filter id=\"ts\" x=\"-30%\" y=\"-30%\" width=\"160%\" height=\"160%\">\n")
	b.WriteString("      <feDropShadow dx=\"0\" dy=\"0\" stdDeviation=\"2\" flood-color=\"black\" flood-opacity=\"0.9\"/>\n")
	b.WriteString("    </filter>\n")
	b.WriteString("  </defs>\n")

	idx := 0
	for r, row := range layout {
		blockTop := margin + float64(r)*(rowH+rowGap)
		labelY := blockTop + labelSize/2
		topY := blockTop + labelSize + labelGap

		// Row label — centered, small, slightly muted.
		// & must be escaped as &amp; in SVG text nodes.
		escapedLabel := strings.ReplaceAll(strings.ToUpper(row.label), "&", "&amp;")
		fmt.Fprintf(&b,
			"  <text x=\"%.2f\" y=\"%.2f\" font-family=\"ui-sans-serif,system-ui,sans-serif\" font-size=\"%.0f\" font-weight=\"600\" fill=\"white\" fill-opacity=\"0.55\" text-anchor=\"middle\" dominant-baseline=\"central\" letter-spacing=\"1\">%s</text>\n",
			svgW/2, labelY, labelSize, escapedLabel,
		)

		rowW := float64(row.count)*bW + float64(row.count-1)*gap
		startX := (svgW - rowW) / 2

		for j := 0; j < row.count && idx < len(stackItems); j++ {
			it := stackItems[idx]
			idx++

			x := startX + float64(j)*(bW+gap)
			cx := x + bW/2
			cy := topY + bH/2

			fmt.Fprintf(&b,
				"  <rect x=\"%.2f\" y=\"%.2f\" width=\"%.0f\" height=\"%.0f\" rx=\"%.0f\" fill=\"none\" stroke=\"white\" stroke-opacity=\"0.25\" stroke-width=\"1\"/>\n",
				x, topY, bW, bH, rx,
			)

			iconURI := ""
			if it.iconFile != "" {
				iconURI = iconCache[it.iconFile]
			}
			hasIcon := iconURI != "" || it.emoji != ""

			iw := iconW
			ih := iconH
			if it.iconW > 0 {
				iw = it.iconW
			}
			if it.iconH > 0 {
				ih = it.iconH
			}

			var iconCy, labelY float64
			if hasIcon {
				contentH := ih + iconGap + fontSize
				top := cy - contentH/2
				iconCy = top + ih/2
				labelY = top + ih + iconGap + fontSize/2
			} else {
				labelY = cy
			}

			if iconURI != "" {
				fmt.Fprintf(&b,
					"  <image href=\"%s\" x=\"%.2f\" y=\"%.2f\" width=\"%.0f\" height=\"%.0f\"/>\n",
					iconURI, cx-iw/2, iconCy-ih/2, iw, ih,
				)
			} else if it.emoji != "" {
				fmt.Fprintf(&b,
					"  <text x=\"%.2f\" y=\"%.2f\" font-size=\"16\" text-anchor=\"middle\" dominant-baseline=\"central\">%s</text>\n",
					cx, iconCy, it.emoji,
				)
			}

			fmt.Fprintf(&b,
				"  <text x=\"%.2f\" y=\"%.2f\" font-family=\"ui-sans-serif,system-ui,sans-serif\" font-size=\"%.0f\" font-weight=\"700\" fill=\"white\" text-anchor=\"middle\" dominant-baseline=\"central\" filter=\"url(#ts)\">%s</text>\n",
				cx, labelY, fontSize, it.label,
			)
		}
	}

	b.WriteString("</svg>\n")

	out := "drafts/stack-brick.svg"
	if len(os.Args) > 1 {
		out = os.Args[1]
	}
	if err := os.WriteFile(out, []byte(b.String()), 0644); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println("wrote", out)
}
