package graph

import "fmt"

// Each builder declares its symmetries alongside the construction, because the
// construction is where they are obvious. Under-reporting symmetries costs
// speed only; over-reporting one silently corrupts answers, so IsSymmetry
// exists to confirm every declaration rather than anyone proving completeness.
//
// Vertex numbering is a contract, not a preference: symmetries are
// permutations of these indices, so renumbering a family invalidates them.

// Path builds P_n on vertices 0..n-1 along the path.
func Path(n int) (Graph, error) {
	if n < 0 {
		return Graph{}, fmt.Errorf("path: n %d: %w", n, ErrNegative)
	}

	adjacency := emptyAdjacency(n)
	for vertex := 0; vertex+1 < n; vertex++ {
		join(adjacency, vertex, vertex+1)
	}
	return fromSets(adjacency, pathSymmetries(n)), nil
}

// pathSymmetries: a path can only be flipped end for end.
func pathSymmetries(n int) []Permutation {
	if n < 2 {
		return nil
	}
	flip := make(Permutation, n)
	for vertex := range flip {
		flip[vertex] = n - 1 - vertex
	}
	return []Permutation{flip}
}

// Cycle builds C_n on vertices 0..n-1 around the cycle.
func Cycle(n int) (Graph, error) {
	if n < 0 {
		return Graph{}, fmt.Errorf("cycle: n %d: %w", n, ErrNegative)
	}

	adjacency := emptyAdjacency(n)
	if n >= 2 {
		for vertex := 0; vertex < n; vertex++ {
			join(adjacency, vertex, (vertex+1)%n)
		}
	}
	return fromSets(adjacency, cycleSymmetries(n)), nil
}

// cycleSymmetries: a cycle rotates one notch and flips over.
func cycleSymmetries(n int) []Permutation {
	if n < 2 {
		return nil
	}
	if n == 2 {
		// C_2 collapses to a single edge, whose only symmetry is the swap.
		return []Permutation{{1, 0}}
	}

	rotate := make(Permutation, n)
	reflect := make(Permutation, n)
	for vertex := 0; vertex < n; vertex++ {
		rotate[vertex] = (vertex + 1) % n
		reflect[vertex] = (n - vertex) % n
	}
	return []Permutation{rotate, reflect}
}

// Star builds K_1,leaves with the centre at vertex 0 and leaves at 1..leaves.
func Star(leaves int) (Graph, error) {
	if leaves < 0 {
		return Graph{}, fmt.Errorf("star: leaves %d: %w", leaves, ErrNegative)
	}

	adjacency := emptyAdjacency(leaves + 1)
	for leaf := 1; leaf <= leaves; leaf++ {
		join(adjacency, 0, leaf)
	}
	// Any permutation of the leaves is a symmetry, which is a far larger group
	// than the other families have. Left undeclared: a star has diameter 2, so
	// it is complete at every d >= 2 and decided without search anyway.
	return fromSets(adjacency, nil), nil
}

// errRimTooSmall reports a rim family built with fewer than three rim vertices.
func errRimTooSmall(name string, rim int) error {
	return fmt.Errorf("%s: rim %d: needs at least 3 rim vertices", name, rim)
}

// Wheel builds W_n: a hub at vertex 0 joined to every vertex of a rim cycle
// 1..n. Order is n+1.
func Wheel(rim int) (Graph, error) {
	if rim < 3 {
		return Graph{}, errRimTooSmall("wheel", rim)
	}

	adjacency := emptyAdjacency(rim + 1)
	for index := 1; index <= rim; index++ {
		join(adjacency, 0, index)
		join(adjacency, index, index%rim+1)
	}
	return fromSets(adjacency, wheelSymmetries(rim)), nil
}

// wheelSymmetries: the rim rotates and reflects; the hub cannot move.
//
// The hub is the only vertex of degree n while every rim vertex has degree 3,
// and relabelling cannot change a vertex's degree, so the hub is fixed.
func wheelSymmetries(rim int) []Permutation {
	rotate := make(Permutation, rim+1)
	reflect := make(Permutation, rim+1)
	rotate[0], reflect[0] = 0, 0
	for index := 1; index <= rim; index++ {
		rotate[index] = index%rim + 1
		reflect[index] = (rim-(index-1))%rim + 1
	}
	return []Permutation{rotate, reflect}
}

// Helm builds H_n: a wheel with one pendant hung off each rim vertex. Hub 0,
// rim 1..n, pendant of rim vertex i at n+i. Order is 2n+1.
func Helm(rim int) (Graph, error) {
	if rim < 3 {
		return Graph{}, errRimTooSmall("helm", rim)
	}

	adjacency := emptyAdjacency(2*rim + 1)
	for index := 1; index <= rim; index++ {
		join(adjacency, 0, index)
		join(adjacency, index, index%rim+1)
		join(adjacency, index, rim+index)
	}
	return fromSets(adjacency, carryPendants(wheelSymmetries(rim), rim, 1)), nil
}

// Sunlet builds S_n (also written C_n o K_1): a cycle with one pendant per
// vertex. Rim 0..n-1, pendant of rim vertex i at n+i. Order is 2n.
func Sunlet(rim int) (Graph, error) {
	if rim < 3 {
		return Graph{}, errRimTooSmall("sunlet", rim)
	}

	adjacency := emptyAdjacency(2 * rim)
	for index := 0; index < rim; index++ {
		join(adjacency, index, (index+1)%rim)
		join(adjacency, index, rim+index)
	}
	return fromSets(adjacency, carryPendants(cycleSymmetries(rim), rim, 0)), nil
}

// carryPendants extends rim symmetries over the pendants hanging off them:
// wherever a rim vertex goes, its pendant follows. firstRim is the index of the
// first rim vertex (1 for a helm, whose vertex 0 is the hub; 0 for a sunlet).
func carryPendants(rimSymmetries []Permutation, rim, firstRim int) []Permutation {
	extended := make([]Permutation, 0, len(rimSymmetries))
	for _, symmetry := range rimSymmetries {
		full := make(Permutation, len(symmetry)+rim)
		copy(full, symmetry)
		for index := 0; index < rim; index++ {
			rimVertex := firstRim + index
			full[len(symmetry)+index] = rim + symmetry[rimVertex]
		}
		extended = append(extended, full)
	}
	return extended
}
