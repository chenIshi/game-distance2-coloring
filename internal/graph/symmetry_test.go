package graph

import (
	"fmt"
	"testing"
)

type declared struct {
	label string
	graph Graph
	group int // expected size of the full symmetry group
}

// declaredFamilies pairs each family with the size its symmetry group should
// come out at: 2 for a path (flip only), 2n for everything with a rim.
func declaredFamilies() []declared {
	var families []declared
	for n := 2; n <= 7; n++ {
		families = append(families, declared{fmt.Sprintf("P_%d", n), mustGraph(Path(n)), 2})
	}
	for n := 3; n <= 8; n++ {
		families = append(families, declared{fmt.Sprintf("C_%d", n), mustGraph(Cycle(n)), 2 * n})
	}
	for n := 3; n <= 7; n++ {
		families = append(families, declared{fmt.Sprintf("W_%d", n), mustGraph(Wheel(n)), 2 * n})
	}
	for n := 3; n <= 6; n++ {
		families = append(families,
			declared{fmt.Sprintf("H_%d", n), mustGraph(Helm(n)), 2 * n},
			declared{fmt.Sprintf("S_%d", n), mustGraph(Sunlet(n)), 2 * n},
		)
	}
	return families
}

// TestDeclaredSymmetriesAreReal is the safety net that makes declaring
// symmetries by hand acceptable. An over-reported symmetry would corrupt
// answers; this catches it.
func TestDeclaredSymmetriesAreReal(t *testing.T) {
	for _, f := range declaredFamilies() {
		for _, generator := range f.graph.Generators {
			if !IsSymmetry(f.graph, generator) {
				t.Errorf("%s: declared generator %v is not a symmetry", f.label, generator)
			}
		}
		for _, element := range f.graph.Symmetries() {
			if !IsSymmetry(f.graph, element) {
				t.Errorf("%s: group element %v is not a symmetry", f.label, element)
			}
		}
	}
}

func TestSymmetryGroupSizes(t *testing.T) {
	for _, f := range declaredFamilies() {
		if got := len(f.graph.Symmetries()); got != f.group {
			t.Errorf("%s: group size = %d, want %d", f.label, got, f.group)
		}
	}
}

// TestNumberingContract checks the promises CLAUDE.md makes about vertex
// indices, which the symmetries are written against.
func TestNumberingContract(t *testing.T) {
	for rim := 3; rim <= 6; rim++ {
		for _, element := range mustGraph(Wheel(rim)).Symmetries() {
			if element[0] != 0 {
				t.Errorf("W_%d: hub moved to %d", rim, element[0])
			}
		}

		helm := mustGraph(Helm(rim))
		for _, element := range helm.Symmetries() {
			if element[0] != 0 {
				t.Errorf("H_%d: hub moved to %d", rim, element[0])
			}
			for index := 1; index <= rim; index++ {
				// A pendant must land on the pendant of wherever its rim went.
				if got, want := element[rim+index], rim+element[index]; got != want {
					t.Errorf("H_%d: pendant of rim %d went to %d, want %d", rim, index, got, want)
				}
			}
		}

		sunlet := mustGraph(Sunlet(rim))
		for _, element := range sunlet.Symmetries() {
			for index := 0; index < rim; index++ {
				if got, want := element[rim+index], rim+element[index]; got != want {
					t.Errorf("S_%d: pendant of rim %d went to %d, want %d", rim, index, got, want)
				}
			}
		}
	}
}

func TestIsSymmetryRejectsBadRelabellings(t *testing.T) {
	square := mustGraph(Cycle(4))

	for _, bad := range []Permutation{
		{0, 0, 2, 3}, // repeats a vertex
		{0, 1, 2},    // wrong length
		{0, 1, 2, 4}, // index out of range
		// A real permutation that breaks an edge: swapping only 0 and 1 would
		// turn edge 1-2 into edge 0-2, and there is no such edge.
		{1, 0, 2, 3},
	} {
		if IsSymmetry(square, bad) {
			t.Errorf("%v should not be a symmetry of C_4", bad)
		}
	}

	// The hub of a wheel cannot be swapped with a rim vertex: their degrees
	// differ, and relabelling cannot change a degree.
	if IsSymmetry(mustGraph(Wheel(4)), Permutation{1, 0, 2, 3, 4}) {
		t.Error("wheel hub should not be swappable with a rim vertex")
	}
}

// TestSymmetriesSurviveEveryPower is the reuse claim: relabelling cannot change
// distances, so symmetries are computed once and reused for every d.
func TestSymmetriesSurviveEveryPower(t *testing.T) {
	for _, f := range declaredFamilies() {
		group := f.graph.Symmetries()
		for distance := 1; distance <= 4; distance++ {
			powered, err := Power(f.graph, distance)
			if err != nil {
				t.Fatalf("%s: %v", f.label, err)
			}
			for _, element := range group {
				if !IsSymmetry(powered, element) {
					t.Errorf("%s at d=%d: %v stopped being a symmetry", f.label, distance, element)
				}
			}
		}
	}
}

func TestGroupAlgebra(t *testing.T) {
	hexagon := mustGraph(Cycle(6))
	group := hexagon.Symmetries()
	identity := Identity(6)

	inGroup := func(p Permutation) bool {
		for _, element := range group {
			if sameNeighbors(element, p) {
				return true
			}
		}
		return false
	}

	if !inGroup(identity) {
		t.Error("group is missing the identity")
	}
	for _, left := range group {
		if !sameNeighbors(Compose(left, identity), left) {
			t.Errorf("identity is not neutral for %v", left)
		}
		if !inGroup(Invert(left)) {
			t.Errorf("group is not closed under inversion: %v", left)
		}
		if !sameNeighbors(Compose(left, Invert(left)), identity) {
			t.Errorf("%v composed with its inverse is not the identity", left)
		}
		for _, right := range group {
			if !inGroup(Compose(left, right)) {
				t.Errorf("group is not closed under composition: %v then %v", right, left)
			}
		}
	}
}

func TestComposeAppliesInnerFirst(t *testing.T) {
	inner := Permutation{1, 2, 0}
	outer := Permutation{0, 2, 1}
	// vertex 0 -> 1 (inner) -> 2 (outer)
	if got := Compose(outer, inner)[0]; got != 2 {
		t.Errorf("Compose(outer, inner)[0] = %d, want 2", got)
	}
}

func TestTinyGraphsHaveOnlyTheIdentity(t *testing.T) {
	for n := 0; n <= 1; n++ {
		if got := len(mustGraph(Path(n)).Symmetries()); got != 1 {
			t.Errorf("P_%d group size = %d, want 1", n, got)
		}
		if got := len(mustGraph(Cycle(n)).Symmetries()); got != 1 {
			t.Errorf("C_%d group size = %d, want 1", n, got)
		}
	}
	// C_2 is a single edge: identity plus the swap.
	if got := len(mustGraph(Cycle(2)).Symmetries()); got != 2 {
		t.Errorf("C_2 group size = %d, want 2", got)
	}
}
