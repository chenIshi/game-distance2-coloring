package game

import (
	"fmt"
	"testing"

	"github.com/chenIshi/game-distance2-coloring/internal/graph"
)

func mustGraph(g graph.Graph, err error) graph.Graph {
	if err != nil {
		panic(err)
	}
	return g
}

func chi(t *testing.T, g graph.Graph, distance int) int {
	t.Helper()
	value, err := ChromaticNumber(g, distance)
	if err != nil {
		t.Fatal(err)
	}
	return value
}

func wins(t *testing.T, g graph.Graph, colorCount, distance int) bool {
	t.Helper()
	won, err := AliceWins(g, colorCount, distance)
	if err != nil {
		t.Fatal(err)
	}
	return won
}

// Every value below is the Python reference implementation's answer, itself
// cross-checked against a separately written brute force. These are the numbers
// the Go port exists to reproduce.

func TestPaths(t *testing.T) {
	for _, tc := range []struct {
		n, distance, want int
	}{
		{1, 2, 1}, {2, 2, 2}, {3, 2, 3}, {4, 2, 3}, {5, 2, 3}, {6, 2, 4}, {7, 2, 4},
		{4, 1, 3}, {5, 1, 3}, {6, 1, 3}, {7, 1, 3},
		{4, 3, 4}, {5, 3, 4}, {6, 3, 5}, {7, 3, 4},
	} {
		if got := chi(t, mustGraph(graph.Path(tc.n)), tc.distance); got != tc.want {
			t.Errorf("P_%d at d=%d: got %d, want %d", tc.n, tc.distance, got, tc.want)
		}
	}
}

func TestCycles(t *testing.T) {
	for _, tc := range []struct {
		n, distance, want int
	}{
		{3, 2, 3}, {4, 2, 4}, {5, 2, 5}, {6, 2, 5}, {7, 2, 4}, {8, 2, 5},
		{5, 1, 3}, {6, 1, 3}, {7, 1, 3},
	} {
		if got := chi(t, mustGraph(graph.Cycle(tc.n)), tc.distance); got != tc.want {
			t.Errorf("C_%d at d=%d: got %d, want %d", tc.n, tc.distance, got, tc.want)
		}
	}
}

// TestStars: K_1,n has diameter 2, so its square is complete on n+1 vertices.
func TestStars(t *testing.T) {
	for leaves := 1; leaves <= 5; leaves++ {
		if got, want := chi(t, mustGraph(graph.Star(leaves)), 2), leaves+1; got != want {
			t.Errorf("K_1,%d at d=2: got %d, want %d", leaves, got, want)
		}
	}
}

func TestWheelsAtDistanceOne(t *testing.T) {
	want := map[int]int{3: 4, 4: 3, 5: 4, 6: 3, 7: 4, 8: 4}
	for rim, expected := range want {
		if got := chi(t, mustGraph(graph.Wheel(rim)), 1); got != expected {
			t.Errorf("W_%d at d=1: got %d, want %d", rim, got, expected)
		}
	}
}

// TestWheelsAreTrivialBeyondDistanceOne: a wheel has diameter 2, so its power
// graph is complete from d=2 on and the answer is forced to one colour per
// vertex.
func TestWheelsAreTrivialBeyondDistanceOne(t *testing.T) {
	for rim := 3; rim <= 8; rim++ {
		for distance := 2; distance <= 4; distance++ {
			g := mustGraph(graph.Wheel(rim))
			reach, err := graph.Power(g, distance)
			if err != nil {
				t.Fatal(err)
			}
			if !graph.IsComplete(reach) {
				t.Errorf("W_%d at d=%d should be complete", rim, distance)
			}
			if got := chi(t, g, distance); got != rim+1 {
				t.Errorf("W_%d at d=%d: got %d, want %d", rim, distance, got, rim+1)
			}
		}
	}
}

func TestHelms(t *testing.T) {
	for _, tc := range []struct {
		rim, distance, want int
	}{
		{3, 1, 4}, {4, 1, 4}, {5, 1, 4},
		{3, 2, 6}, {4, 2, 5},
		{3, 3, 7},
	} {
		if got := chi(t, mustGraph(graph.Helm(tc.rim)), tc.distance); got != tc.want {
			t.Errorf("H_%d at d=%d: got %d, want %d", tc.rim, tc.distance, got, tc.want)
		}
	}
}

func TestSunlets(t *testing.T) {
	for _, tc := range []struct {
		rim, distance, want int
	}{
		{3, 1, 3}, {4, 1, 4}, {5, 1, 3},
		{3, 2, 5}, {4, 2, 5},
		{3, 3, 6}, {4, 3, 7},
	} {
		if got := chi(t, mustGraph(graph.Sunlet(tc.rim)), tc.distance); got != tc.want {
			t.Errorf("S_%d at d=%d: got %d, want %d", tc.rim, tc.distance, got, tc.want)
		}
	}
}

// TestExpensiveGrid holds the values that cost real time. Skipped under -short.
func TestExpensiveGrid(t *testing.T) {
	if testing.Short() {
		t.Skip("expensive; run without -short")
	}
	for _, tc := range []struct {
		label    string
		graph    graph.Graph
		distance int
		want     int
	}{
		{"H_5", mustGraph(graph.Helm(5)), 2, 6},
		{"H_4", mustGraph(graph.Helm(4)), 3, 7},
		{"H_5", mustGraph(graph.Helm(5)), 3, 9},
		{"S_5", mustGraph(graph.Sunlet(5)), 2, 6},
		{"S_5", mustGraph(graph.Sunlet(5)), 3, 8},
	} {
		if got := chi(t, tc.graph, tc.distance); got != tc.want {
			t.Errorf("%s at d=%d: got %d, want %d", tc.label, tc.distance, got, tc.want)
		}
	}
}

// TestNonMonotonicInSize pins cases where a larger graph needs strictly fewer
// colours. Real, and easy to mistake for a bug.
func TestNonMonotonicInSize(t *testing.T) {
	for _, tc := range []struct {
		label           string
		smaller, larger graph.Graph
		distance        int
		low, high       int
	}{
		{"C_6 vs C_7 d=2", mustGraph(graph.Cycle(6)), mustGraph(graph.Cycle(7)), 2, 5, 4},
		{"W_5 vs W_6 d=1", mustGraph(graph.Wheel(5)), mustGraph(graph.Wheel(6)), 1, 4, 3},
		{"H_3 vs H_4 d=2", mustGraph(graph.Helm(3)), mustGraph(graph.Helm(4)), 2, 6, 5},
		{"S_4 vs S_5 d=1", mustGraph(graph.Sunlet(4)), mustGraph(graph.Sunlet(5)), 1, 4, 3},
		{"P_6 vs P_7 d=3", mustGraph(graph.Path(6)), mustGraph(graph.Path(7)), 3, 5, 4},
	} {
		if got := chi(t, tc.smaller, tc.distance); got != tc.low {
			t.Errorf("%s: smaller got %d, want %d", tc.label, got, tc.low)
		}
		if got := chi(t, tc.larger, tc.distance); got != tc.high {
			t.Errorf("%s: larger got %d, want %d", tc.label, got, tc.high)
		}
	}
}

// TestMonotonicityInColours: Alice must never lose a game she could win with
// fewer colours. Measured, not assumed.
func TestMonotonicityInColours(t *testing.T) {
	for _, g := range gridGraphs() {
		for distance := 1; distance <= 3; distance++ {
			previous := false
			for colorCount := 0; colorCount <= g.graph.Order(); colorCount++ {
				current := wins(t, g.graph, colorCount, distance)
				if previous && !current {
					t.Errorf("%s at d=%d: Alice wins with k=%d but loses with k=%d",
						g.label, distance, colorCount-1, colorCount)
				}
				previous = current
			}
		}
	}
}

// TestMonotonicityInDistance: reaching further can only make Alice's job
// harder. Less obvious than it looks, since G^d is a subgraph of G^(d+1) and the
// game chromatic number is not monotone under subgraphs.
func TestMonotonicityInDistance(t *testing.T) {
	for _, g := range gridGraphs() {
		previous := 0
		for distance := 1; distance <= 3; distance++ {
			current := chi(t, g.graph, distance)
			if current < previous {
				t.Errorf("%s: chi fell from %d at d=%d to %d at d=%d",
					g.label, previous, distance-1, current, distance)
			}
			previous = current
		}
	}
}

type labelled struct {
	label string
	graph graph.Graph
}

func gridGraphs() []labelled {
	var graphs []labelled
	for n := 2; n <= 7; n++ {
		graphs = append(graphs, labelled{fmt.Sprintf("P_%d", n), mustGraph(graph.Path(n))})
	}
	for n := 3; n <= 7; n++ {
		graphs = append(graphs, labelled{fmt.Sprintf("C_%d", n), mustGraph(graph.Cycle(n))})
	}
	for n := 3; n <= 6; n++ {
		graphs = append(graphs, labelled{fmt.Sprintf("W_%d", n), mustGraph(graph.Wheel(n))})
	}
	graphs = append(graphs,
		labelled{"H_3", mustGraph(graph.Helm(3))},
		labelled{"S_3", mustGraph(graph.Sunlet(3))},
	)
	return graphs
}

func TestEmptyGraph(t *testing.T) {
	empty := mustGraph(graph.Path(0))
	if !wins(t, empty, 0, 2) {
		t.Error("Alice should win the empty graph with zero colours")
	}
	if got := chi(t, empty, 2); got != 0 {
		t.Errorf("chi of the empty graph = %d, want 0", got)
	}
}

// TestVertexCountAlwaysSuffices: a vertex has at most |V|-1 neighbours, so it
// can never see all |V| colours and no vertex can go dead.
func TestVertexCountAlwaysSuffices(t *testing.T) {
	for _, g := range gridGraphs() {
		if !wins(t, g.graph, g.graph.Order(), 2) {
			t.Errorf("%s: Alice should win with one colour per vertex", g.label)
		}
	}
}

func TestDeadVertexIsReported(t *testing.T) {
	// P_2 with one colour: Alice colours one end, the other has nothing left.
	reach, err := graph.Power(mustGraph(graph.Path(2)), 2)
	if err != nil {
		t.Fatal(err)
	}
	solver := NewSolver(reach, 1)
	if dead := solver.DeadVertices([]int{1, 0}); len(dead) != 1 || dead[0] != 1 {
		t.Errorf("dead vertices = %v, want [1]", dead)
	}
	if solver.Solve([]int{0, 0}, true) {
		t.Error("Alice should lose P_2 with one colour")
	}
}
