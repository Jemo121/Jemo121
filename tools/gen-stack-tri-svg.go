package main

import (
	"encoding/base64"
	"fmt"
	"math"
	"os"
	"strings"
)

type stackItem struct {
	label    string
	iconFile string
	emoji    string
	fill     string
}

var stackItems = []stackItem{
	// Row 0 — Cloud Native (7)
	{"Kubernetes", "assets/kubernetes.svg", "", "#326CE5"},
	{"Podman", "assets/podman.svg", "", "#326CE5"},
	{"Crossplane", "assets/crossplane.svg", "", "#326CE5"},
	{"Cloudflare", "assets/cloudflare.svg", "", "#326CE5"},
	{"Helm", "assets/helm.svg", "", "#326CE5"},
	{"Akamai", "assets/akamai.svg", "", "#326CE5"},
	{"LXD", "assets/linuxcontainers.svg", "", "#326CE5"},
	// Row 1 — Languages + Tools start (6)
	{"Go", "assets/go.svg", "", "#00ADD8"},
	{"Rust", "assets/rust.svg", "", "#CE422B"},
	{"Bash", "assets/bash.svg", "", "#4EAA25"},
	{"WASM", "assets/webassembly.svg", "", "#654FF0"},
	{"Git", "assets/git.svg", "", "#F05032"},
	{"Neovim", "assets/neovim.svg", "", "#57A143"},
	// Row 2 — Tools end + Security start (5)
	{"Ubuntu", "assets/ubuntu.svg", "", "#E95420"},
	{"Arch", "assets/archlinux.svg", "", "#1793D1"},
	{"Claude", "assets/claude-code.svg", "", "#D97706"},
	{"ISO 27001", "", "🔒", "#003580"},
	{"CIS", "", "🛡️", "#1A1A2E"},
	// Row 3 — Security end + DB (4)
	{"SOC 2", "", "📋", "#0052CC"},
	{"OWASP", "assets/owasp.svg", "", "#1A1A2E"},
	{"Postgres", "assets/postgresql.svg", "", "#336791"},
	{"Redis", "assets/redis.svg", "", "#DC382D"},
	// Row 4 — Forges (2)
	{"GitLab", "assets/gitlab.svg", "", "#FC6D26"},
	{"GitHub", "assets/github.svg", "", "#2D333B"},
}

var layout = []int{7, 6, 5, 4, 2}

func iconDataURI(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return "data:image/svg+xml;base64," + base64.StdEncoding.EncodeToString(data)
}

func main() {
	const (
		s        = 130.0
		margin   = 24.0
		iconH    = 20.0
		iconW    = 20.0
		iconGap  = 4.0
		fontSize = 10.0
	)

	H := s * math.Sqrt(3) / 2 // equilateral triangle height

	// Widest row (N=7) has width = s*(7+1)/2 = 4s.
	svgW := 4*s + 2*margin
	svgH := H*float64(len(layout)) + 2*margin

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
	for row, count := range layout {
		topY := margin + float64(row)*H
		// Row of N triangles spans width = s*(N+1)/2.
		rowWidth := s * float64(count+1) / 2
		startX := svgW/2 - rowWidth/2

		for j := 0; j < count && idx < len(stackItems); j++ {
			it := stackItems[idx]
			idx++

			// (j+row)%2==0 → up-pointing; odd → down-pointing.
			// Even rows start with up, odd rows start with down — creates
			// continuous up/down/up/down reading order across rows.
			isUp := (j+row)%2 == 0
			x := startX + float64(j)*s/2

			var pts string
			var cx, cy float64
			cx = x + s/2
			if isUp {
				// apex at top, base at bottom
				pts = fmt.Sprintf("%.2f,%.2f %.2f,%.2f %.2f,%.2f",
					x+s/2, topY,
					x, topY+H,
					x+s, topY+H,
				)
				cy = topY + 2*H/3
			} else {
				// apex at bottom, base at top
				pts = fmt.Sprintf("%.2f,%.2f %.2f,%.2f %.2f,%.2f",
					x, topY,
					x+s, topY,
					x+s/2, topY+H,
				)
				cy = topY + H/3
			}

			fmt.Fprintf(&b,
				"  <polygon points=\"%s\" fill=\"%s\" stroke=\"white\" stroke-opacity=\"0.15\" stroke-width=\"1\" stroke-linejoin=\"round\"/>\n",
				pts, it.fill,
			)

			iconURI := ""
			if it.iconFile != "" {
				iconURI = iconCache[it.iconFile]
			}
			hasIcon := iconURI != "" || it.emoji != ""

			var iconCy, labelY float64
			if hasIcon {
				contentH := iconH + iconGap + fontSize
				top := cy - contentH/2
				iconCy = top + iconH/2
				labelY = top + iconH + iconGap + fontSize/2
			} else {
				labelY = cy
			}

			if iconURI != "" {
				fmt.Fprintf(&b,
					"  <image href=\"%s\" x=\"%.2f\" y=\"%.2f\" width=\"%.0f\" height=\"%.0f\"/>\n",
					iconURI, cx-iconW/2, iconCy-iconH/2, iconW, iconH,
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

	out := "drafts/stack-tri.svg"
	if len(os.Args) > 1 {
		out = os.Args[1]
	}
	if err := os.WriteFile(out, []byte(b.String()), 0644); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println("wrote", out)
}
