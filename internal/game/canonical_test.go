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

func canonicalPairAgrees(t *testing.T, label string, g graph.Graph, distance int) {
	t.Helper()

	reach, err := graph.Power(g, distance)
	if err != nil {
		t.Fatal(err)
	}
	if graph.IsComplete(reach) {
		return // decided by the shortcut, never reaches the search
	}

	for colorCount := 1; colorCount <= reach.Order(); colorCount++ {
		start := make([]int, reach.Order())
		plain := newSolver(reach, colorCount, false).Solve(start, true)
		canonical := newSolver(reach, colorCount, true).Solve(start, true)
		if plain != canonical {
			t.Errorf("%s at d=%d k=%d: plain=%v canonical=%v",
				label, distance, colorCount, plain, canonical)
		}
	}
}

func TestCanonicalisationPreservesEveryVerdict(t *testing.T) {
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
			canonicalPairAgrees(t, tc.label, tc.graph, distance)
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
