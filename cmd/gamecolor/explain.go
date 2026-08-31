package main

import (
	"fmt"
	"io"
	"strings"

	"github.com/chenIshi/game-distance2-coloring/internal/game"
)

// vertexNamer returns a function naming vertices the way the family's own
// structure does. "rim4 blocked by hub, rim3 and its own stub" says what
// happened; "v4 blocked by v0, v3, v8" makes the reader decode indices.
//
// The names encode the numbering contract from CLAUDE.md, so a mismatch here
// would be a mismatch with the builders and the declared symmetries.
func vertexNamer(family string, n int) func(int) string {
	switch family {
	case "star":
		return func(v int) string {
			if v == 0 {
				return "centre"
			}
			return fmt.Sprintf("leaf%d", v)
		}
	case "wheel":
		return func(v int) string {
			if v == 0 {
				return "hub"
			}
			return fmt.Sprintf("rim%d", v)
		}
	case "helm":
		return func(v int) string {
			switch {
			case v == 0:
				return "hub"
			case v <= n:
				return fmt.Sprintf("rim%d", v)
			default:
				return fmt.Sprintf("stub%d", v-n)
			}
		}
	case "sunlet":
		return func(v int) string {
			if v < n {
				return fmt.Sprintf("rim%d", v)
			}
			return fmt.Sprintf("stub%d", v-n)
		}
	default: // path, cycle
		return func(v int) string { return fmt.Sprintf("v%d", v) }
	}
}

func writeExplanation(out io.Writer, analysis game.Analysis, name func(int) string, colorCount int) {
	width := 0
	for _, step := range analysis.Line {
		if w := len(name(step.Vertex)); w > width {
			width = w
		}
	}

	if analysis.Opening != nil {
		fmt.Fprintf(out, "winning opening: Alice colours %s with %d\n",
			name(analysis.Opening.Vertex), analysis.Opening.Color)
	}

	fmt.Fprintln(out, "one line of play:")
	for _, step := range analysis.Line {
		fmt.Fprintf(out, "  %-6s %-*s -> %d\n", step.Player+":", width, name(step.Vertex), step.Color)
	}

	if len(analysis.Dead) == 0 {
		fmt.Fprintln(out, "every vertex coloured.")
	}
	for _, vertex := range analysis.Dead {
		blockers := analysis.Blockers(vertex)
		parts := make([]string, 0, len(blockers))
		for _, blocker := range blockers {
			parts = append(parts, fmt.Sprintf("%s=%d", name(blocker.Vertex), blocker.Color))
		}
		fmt.Fprintf(out, "dead vertex: %s\n", name(vertex))
		fmt.Fprintf(out, "  blocked by %s\n", strings.Join(parts, ", "))
		fmt.Fprintf(out, "  between them that is all %d colours -- nothing left for it\n", colorCount)
	}

	fmt.Fprintln(out, "(one line of play, not a proof of the whole strategy tree)")
}
