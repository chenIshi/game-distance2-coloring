package game

import (
	"fmt"
	"testing"

	"github.com/chenIshi/game-distance2-coloring/internal/graph"
)

// Colour canonicalisation is the one optimisation here that could produce
// plausible wrong answers rather than crashing, so it gets checked directly
// against the search it replaced rather than only through the Python harness.
//
// The Python reference deliberately does NOT canonicalise, which is what makes
// scripts/conformance.py an independent check on this trick. These tests do the
// same comparison inside Go, so it runs in milliseconds and can reach cases
// Python is too slow to grade.

// configurations are the search variants that must all agree. Plain is the
// original unoptimised search; each later entry folds away more positions.
var configurations = []struct {
	label     string
	canonical bool
	symmetry  bool
}{
	{"plain", false, false},
	{"colours", true, false},
	{"colours+symmetry", true, true},
}

func configurationsAgree(t *testing.T, label string, g graph.Graph, distance int) {
	t.Helper()

	reach, err := graph.Power(g, distance)
	if err != nil {
		t.Fatal(err)
	}
	if graph.IsComplete(reach) {
		return // decided by the shortcut, never reaches the search
	}

	for colorCount := 1; colorCount <= reach.Order(); colorCount++ {
		var verdicts []bool
		for _, config := range configurations {
			solver := newSolver(reach, colorCount, config.canonical, config.symmetry)
			verdicts = append(verdicts, solver.Solve(make([]int, reach.Order()), true))
		}
		for i := 1; i < len(verdicts); i++ {
			if verdicts[i] != verdicts[0] {
				t.Errorf("%s at d=%d k=%d: %s=%v but %s=%v",
					label, distance, colorCount,
					configurations[0].label, verdicts[0],
					configurations[i].label, verdicts[i])
			}
		}
	}
}

func TestOptimisationsPreserveEveryVerdict(t *testing.T) {
	var cases []labelled
	for n := 2; n <= 8; n++ {
		cases = append(cases, labelled{fmt.Sprintf("P_%d", n), mustGraph(graph.Path(n))})
	}
	for n := 3; n <= 8; n++ {
		cases = append(cases, labelled{fmt.Sprintf("C_%d", n), mustGraph(graph.Cycle(n))})
	}
	for n := 3; n <= 7; n++ {
		cases = append(cases, labelled{fmt.Sprintf("W_%d", n), mustGraph(graph.Wheel(n))})
	}
	for n := 3; n <= 4; n++ {
		cases = append(cases,
			labelled{fmt.Sprintf("H_%d", n), mustGraph(graph.Helm(n))},
			labelled{fmt.Sprintf("S_%d", n), mustGraph(graph.Sunlet(n))},
		)
	}

	for _, tc := range cases {
		for distance := 1; distance <= 3; distance++ {
			configurationsAgree(t, tc.label, tc.graph, distance)
		}
	}
}

// TestSymmetryFoldingOnlyUsesRealSymmetries guards the direction of the risk.
// Folding by something that is not a symmetry would silently merge positions
// that are genuinely different, so every permutation the solver folds by is
// re-checked against the graph it is applied to.
func TestSymmetryFoldingOnlyUsesRealSymmetries(t *testing.T) {
	for n := 3; n <= 7; n++ {
		for _, g := range []graph.Graph{
			mustGraph(graph.Cycle(n)),
			mustGraph(graph.Wheel(n)),
			mustGraph(graph.Sunlet(n)),
		} {
			for distance := 1; distance <= 3; distance++ {
				reach, err := graph.Power(g, distance)
				if err != nil {
					t.Fatal(err)
				}
				solver := NewSolver(reach, 4)
				for _, permutation := range solver.symmetries {
					if !graph.IsSymmetry(reach, permutation) {
						t.Errorf("order %d d=%d: folding by a non-symmetry %v",
							g.Order(), distance, permutation)
					}
				}
			}
		}
	}
}

// TestCanonicalColors pins the relabelling itself, independently of any search.
func TestCanonicalColors(t *testing.T) {
	for _, tc := range []struct {
		colorCount int
		in, want   []int
	}{
		{5, []int{0, 2, 0, 3}, []int{0, 1, 0, 2}},
		{5, []int{0, 3, 0, 1}, []int{0, 1, 0, 2}}, // same class as above
		{5, []int{1, 2, 3}, []int{1, 2, 3}},       // already canonical
		{5, []int{3, 3, 1}, []int{1, 1, 2}},
		{3, []int{0, 0, 0}, []int{0, 0, 0}},
		{4, []int{4, 0, 4, 2}, []int{1, 0, 1, 2}},
	} {
		solver := NewSolver(mustGraph(graph.Path(len(tc.in))), tc.colorCount)
		got := append([]int{}, tc.in...)
		solver.canonicalColors(got)
		for i := range got {
			if got[i] != tc.want[i] {
				t.Errorf("canonicalColors(%v) = %v, want %v", tc.in, got, tc.want)
				break
			}
		}
	}
}

// TestCanonicalColorsIsIdempotent: canonicalising twice must change nothing.
func TestCanonicalColorsIsIdempotent(t *testing.T) {
	solver := NewSolver(mustGraph(graph.Path(6)), 6)
	position := []int{0, 4, 2, 4, 0, 6}
	solver.canonicalColors(position)
	once := append([]int{}, position...)
	solver.canonicalColors(position)
	for i := range position {
		if position[i] != once[i] {
			t.Fatalf("not idempotent: %v then %v", once, position)
		}
	}
}
