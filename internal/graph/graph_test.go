package graph

import (
	"errors"
	"fmt"
	"testing"
)

// mustGraph unwraps a builder call whose input is known good. Written to take
// exactly a (Graph, error) pair so it can be applied directly to a builder
// call: mustGraph(Path(5)).
func mustGraph(g Graph, err error) Graph {
	if err != nil {
		panic(err)
	}
	return g
}

type sample struct {
	label string
	graph Graph
}

// sampleGraphs is the set every structural invariant is checked against.
func sampleGraphs() []sample {
	var samples []sample
	for n := 0; n <= 6; n++ {
		samples = append(samples,
			sample{fmt.Sprintf("P_%d", n), mustGraph(Path(n))},
			sample{fmt.Sprintf("C_%d", n), mustGraph(Cycle(n))},
			sample{fmt.Sprintf("K_1,%d", n), mustGraph(Star(n))},
		)
	}
	for rim := 3; rim <= 6; rim++ {
		samples = append(samples,
			sample{fmt.Sprintf("W_%d", rim), mustGraph(Wheel(rim))},
			sample{fmt.Sprintf("H_%d", rim), mustGraph(Helm(rim))},
			sample{fmt.Sprintf("S_%d", rim), mustGraph(Sunlet(rim))},
		)
	}
	return samples
}

func TestAdjacencyIsSymmetric(t *testing.T) {
	for _, s := range sampleGraphs() {
		for vertex, neighbors := range s.graph.Neighbors {
			for _, neighbor := range neighbors {
				if !s.graph.Adjacent(neighbor, vertex) {
					t.Errorf("%s: %d-%d present in one direction only", s.label, vertex, neighbor)
				}
			}
		}
	}
}

func TestNoSelfLoops(t *testing.T) {
	for _, s := range sampleGraphs() {
		for vertex, neighbors := range s.graph.Neighbors {
			for _, neighbor := range neighbors {
				if neighbor == vertex {
					t.Errorf("%s: vertex %d is its own neighbour", s.label, vertex)
				}
			}
		}
	}
}

func TestFamilySizes(t *testing.T) {
	cases := []struct {
		label string
		graph Graph
		order int
		edges int
	}{
		{"P_5", mustGraph(Path(5)), 5, 4},
		{"P_0", mustGraph(Path(0)), 0, 0},
		{"C_6", mustGraph(Cycle(6)), 6, 6},
		{"C_2", mustGraph(Cycle(2)), 2, 1},
		{"C_1", mustGraph(Cycle(1)), 1, 0},
		{"K_1,4", mustGraph(Star(4)), 5, 4},
		{"W_5", mustGraph(Wheel(5)), 6, 10},
		{"W_7", mustGraph(Wheel(7)), 8, 14},
		{"H_4", mustGraph(Helm(4)), 9, 12},
		{"H_5", mustGraph(Helm(5)), 11, 15},
		{"S_4", mustGraph(Sunlet(4)), 8, 8},
		{"S_5", mustGraph(Sunlet(5)), 10, 10},
	}
	for _, tc := range cases {
		if got := tc.graph.Order(); got != tc.order {
			t.Errorf("%s: order = %d, want %d", tc.label, got, tc.order)
		}
		if got := tc.graph.EdgeCount(); got != tc.edges {
			t.Errorf("%s: edges = %d, want %d", tc.label, got, tc.edges)
		}
	}
}

// TestRimFamilyNumbering guards the contract that symmetries depend on: the hub
// is vertex 0 and the pendant of rim vertex i sits at rim+i.
func TestRimFamilyNumbering(t *testing.T) {
	const rim = 5

	wheel := mustGraph(Wheel(rim))
	if got := wheel.Degree(0); got != rim {
		t.Errorf("wheel hub degree = %d, want %d", got, rim)
	}
	for index := 1; index <= rim; index++ {
		if got := wheel.Degree(index); got != 3 {
			t.Errorf("wheel rim %d degree = %d, want 3", index, got)
		}
	}

	helm := mustGraph(Helm(rim))
	if got := helm.Degree(0); got != rim {
		t.Errorf("helm hub degree = %d, want %d", got, rim)
	}
	for index := 1; index <= rim; index++ {
		if got := helm.Degree(index); got != 4 {
			t.Errorf("helm rim %d degree = %d, want 4", index, got)
		}
		if got := helm.Neighbors[rim+index]; len(got) != 1 || got[0] != index {
			t.Errorf("helm pendant %d neighbours = %v, want [%d]", rim+index, got, index)
		}
	}

	sunlet := mustGraph(Sunlet(rim))
	for index := 0; index < rim; index++ {
		if got := sunlet.Degree(index); got != 3 {
			t.Errorf("sunlet rim %d degree = %d, want 3", index, got)
		}
		if got := sunlet.Neighbors[rim+index]; len(got) != 1 || got[0] != index {
			t.Errorf("sunlet pendant %d neighbours = %v, want [%d]", rim+index, got, index)
		}
	}
}

func TestBuildersRejectBadSizes(t *testing.T) {
	if _, err := Path(-1); !errors.Is(err, ErrNegative) {
		t.Errorf("Path(-1) error = %v, want ErrNegative", err)
	}
	if _, err := Cycle(-1); !errors.Is(err, ErrNegative) {
		t.Errorf("Cycle(-1) error = %v, want ErrNegative", err)
	}
	if _, err := Star(-1); !errors.Is(err, ErrNegative) {
		t.Errorf("Star(-1) error = %v, want ErrNegative", err)
	}
	for _, rim := range []int{-1, 0, 2} {
		if _, err := Wheel(rim); err == nil {
			t.Errorf("Wheel(%d) should have failed", rim)
		}
		if _, err := Helm(rim); err == nil {
			t.Errorf("Helm(%d) should have failed", rim)
		}
		if _, err := Sunlet(rim); err == nil {
			t.Errorf("Sunlet(%d) should have failed", rim)
		}
	}
}

func sameNeighbors(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestPowerAtDistanceOneIsTheGraphItself(t *testing.T) {
	for _, s := range sampleGraphs() {
		powered, err := Power(s.graph, 1)
		if err != nil {
			t.Fatalf("%s: %v", s.label, err)
		}
		for vertex := range s.graph.Neighbors {
			if !sameNeighbors(powered.Neighbors[vertex], s.graph.Neighbors[vertex]) {
				t.Errorf("%s: vertex %d changed at d=1: %v vs %v",
					s.label, vertex, powered.Neighbors[vertex], s.graph.Neighbors[vertex])
			}
		}
	}
}

func TestPowerIsExactOnAPath(t *testing.T) {
	cubed, err := Power(mustGraph(Path(5)), 3)
	if err != nil {
		t.Fatal(err)
	}
	want := [][]int{{1, 2, 3}, {0, 2, 3, 4}, {0, 1, 3, 4}, {0, 1, 2, 4}, {1, 2, 3}}
	for vertex, expected := range want {
		if !sameNeighbors(cubed.Neighbors[vertex], expected) {
			t.Errorf("vertex %d: got %v, want %v", vertex, cubed.Neighbors[vertex], expected)
		}
	}
}

func TestPowerOnlyGrows(t *testing.T) {
	for _, s := range sampleGraphs() {
		for distance := 0; distance < 5; distance++ {
			smaller, _ := Power(s.graph, distance)
			larger, _ := Power(s.graph, distance+1)
			for vertex := range smaller.Neighbors {
				for _, neighbor := range smaller.Neighbors[vertex] {
					if !larger.Adjacent(vertex, neighbor) {
						t.Errorf("%s: edge %d-%d present at d=%d, gone at d=%d",
							s.label, vertex, neighbor, distance, distance+1)
					}
				}
			}
		}
	}
}

func TestPowerAtDistanceZeroHasNoEdges(t *testing.T) {
	powered, err := Power(mustGraph(Path(5)), 0)
	if err != nil {
		t.Fatal(err)
	}
	if powered.Order() != 5 || powered.EdgeCount() != 0 {
		t.Errorf("order=%d edges=%d, want 5 and 0", powered.Order(), powered.EdgeCount())
	}
}

func TestPowerRejectsNegativeDistance(t *testing.T) {
	if _, err := Power(mustGraph(Path(3)), -1); !errors.Is(err, ErrNegative) {
		t.Errorf("error = %v, want ErrNegative", err)
	}
}

func TestReachingTheDiameterCompletesTheGraph(t *testing.T) {
	for _, tc := range []sample{
		{"P_5", mustGraph(Path(5))},
		{"C_7", mustGraph(Cycle(7))},
		{"W_5", mustGraph(Wheel(5))},
		{"H_4", mustGraph(Helm(4))},
		{"S_4", mustGraph(Sunlet(4))},
	} {
		span, connected := Diameter(tc.graph)
		if !connected {
			t.Fatalf("%s: unexpectedly disconnected", tc.label)
		}
		below, _ := Power(tc.graph, span-1)
		at, _ := Power(tc.graph, span)
		if IsComplete(below) {
			t.Errorf("%s: already complete at d=%d", tc.label, span-1)
		}
		if !IsComplete(at) {
			t.Errorf("%s: not complete at its diameter d=%d", tc.label, span)
		}
	}
}

func TestKnownDiameters(t *testing.T) {
	for _, tc := range []struct {
		label string
		graph Graph
		want  int
	}{
		{"P_5", mustGraph(Path(5)), 4},
		{"P_7", mustGraph(Path(7)), 6},
		{"C_5", mustGraph(Cycle(5)), 2},
		{"C_6", mustGraph(Cycle(6)), 3},
		{"C_7", mustGraph(Cycle(7)), 3},
		{"K_1,4", mustGraph(Star(4)), 2},
		{"W_3", mustGraph(Wheel(3)), 1},
		{"W_6", mustGraph(Wheel(6)), 2},
		{"H_3", mustGraph(Helm(3)), 3},
		{"H_4", mustGraph(Helm(4)), 4},
		{"H_5", mustGraph(Helm(5)), 4},
		{"S_3", mustGraph(Sunlet(3)), 3},
		{"S_4", mustGraph(Sunlet(4)), 4},
		{"S_5", mustGraph(Sunlet(5)), 4},
	} {
		got, connected := Diameter(tc.graph)
		if !connected || got != tc.want {
			t.Errorf("%s: diameter = %d (connected=%v), want %d", tc.label, got, connected, tc.want)
		}
	}
}

// TestWheelsAndStarsHaveDiameterTwo is why both families are trivial at d >= 2.
func TestWheelsAndStarsHaveDiameterTwo(t *testing.T) {
	for n := 4; n <= 8; n++ {
		if got, _ := Diameter(mustGraph(Wheel(n))); got != 2 {
			t.Errorf("W_%d diameter = %d, want 2", n, got)
		}
		if got, _ := Diameter(mustGraph(Star(n))); got != 2 {
			t.Errorf("K_1,%d diameter = %d, want 2", n, got)
		}
	}
}

func TestIsComplete(t *testing.T) {
	if !IsComplete(mustGraph(Cycle(3))) {
		t.Error("C_3 is a triangle and should be complete")
	}
	if IsComplete(mustGraph(Path(3))) {
		t.Error("P_3 is not complete")
	}
	if !IsComplete(mustGraph(Path(0))) {
		t.Error("the empty graph is vacuously complete")
	}
}

func TestDisconnectedGraphHasNoDiameter(t *testing.T) {
	disconnected := fromSets([]map[int]bool{{1: true}, {0: true}, {}}, nil)
	if _, connected := Diameter(disconnected); connected {
		t.Error("expected disconnected")
	}
}
