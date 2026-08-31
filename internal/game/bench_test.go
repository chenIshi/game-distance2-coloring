package game

import (
	"testing"

	"github.com/chenIshi/game-distance2-coloring/internal/graph"
)

// benchCases are the cells that actually cost something. Each optimization is
// measured against these and re-graded by scripts/conformance.py afterwards.
var benchCases = []struct {
	label    string
	graph    graph.Graph
	distance int
}{
	{"P_9/d2", mustGraph(graph.Path(9)), 2},
	{"C_9/d2", mustGraph(graph.Cycle(9)), 2},
	{"W_8/d1", mustGraph(graph.Wheel(8)), 1},
	{"S_4/d3", mustGraph(graph.Sunlet(4)), 3},
	{"H_5/d2", mustGraph(graph.Helm(5)), 2},
}

func BenchmarkChromaticNumber(b *testing.B) {
	for _, tc := range benchCases {
		b.Run(tc.label, func(b *testing.B) {
			for b.Loop() {
				if _, err := ChromaticNumber(tc.graph, tc.distance); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// TestPositionCount reports how many distinct positions each case stores. This
// is what the canonicalization work is meant to shrink, and time follows it
// closely, so it is the number to watch when optimizing. Run with -v.
func TestPositionCount(t *testing.T) {
	if testing.Short() {
		t.Skip("expensive; run without -short")
	}
	for _, tc := range benchCases {
		reach, err := graph.Power(tc.graph, tc.distance)
		if err != nil {
			t.Fatal(err)
		}
		chi, err := ChromaticNumber(tc.graph, tc.distance)
		if err != nil {
			t.Fatal(err)
		}
		solver := NewSolver(reach, chi)
		solver.Solve(make([]int, reach.Order()), true)
		t.Logf("%-8s k=%d  positions=%d", tc.label, chi, solver.Positions())
	}
}
