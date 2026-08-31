package graph

import (
	"slices"
	"sort"
)

// Symmetries are relabellings of the vertices that leave the edges alone. Two
// positions of the colouring game that differ by a symmetry are the same
// position, so recognising them collapses work.
//
// The safety property that shapes this file: missing a symmetry only costs
// speed, while claiming a false one silently corrupts answers. So builders
// declare the obvious rotations and reflections and IsSymmetry confirms them,
// rather than anyone trying to prove a list is complete.

// Identity is the relabelling that moves nothing.
func Identity(size int) Permutation {
	p := make(Permutation, size)
	for vertex := range p {
		p[vertex] = vertex
	}
	return p
}

// Compose applies inner first, then outer.
func Compose(outer, inner Permutation) Permutation {
	combined := make(Permutation, len(inner))
	for vertex := range inner {
		combined[vertex] = outer[inner[vertex]]
	}
	return combined
}

// Invert returns the relabelling that undoes p.
func Invert(p Permutation) Permutation {
	inverse := make(Permutation, len(p))
	for vertex, image := range p {
		inverse[image] = vertex
	}
	return inverse
}

// IsPermutation reports whether p uses each vertex of the given size exactly
// once.
func IsPermutation(p Permutation, size int) bool {
	if len(p) != size {
		return false
	}
	seen := make([]bool, size)
	for _, image := range p {
		if image < 0 || image >= size || seen[image] {
			return false
		}
		seen[image] = true
	}
	return true
}

// IsSymmetry reports whether p renames vertices without disturbing any edge.
//
// This is the check that makes hand-declared symmetries safe to rely on.
func IsSymmetry(g Graph, p Permutation) bool {
	if !IsPermutation(p, g.Order()) {
		return false
	}
	for vertex, neighbors := range g.Neighbors {
		for _, neighbor := range neighbors {
			if !g.Adjacent(p[vertex], p[neighbor]) {
				return false
			}
		}
	}
	return true
}

// CloseGroup expands generators into every relabelling reachable by combining
// them, the way six face turns generate every position of a cube. The result is
// sorted, so it is stable across runs.
func CloseGroup(size int, generators []Permutation) []Permutation {
	start := Identity(size)
	found := map[string]bool{permutationKey(start): true}
	group := []Permutation{start}
	frontier := []Permutation{start}

	for len(frontier) > 0 {
		var next []Permutation
		for _, element := range frontier {
			for _, generator := range generators {
				combined := Compose(generator, element)
				key := permutationKey(combined)
				if !found[key] {
					found[key] = true
					group = append(group, combined)
					next = append(next, combined)
				}
			}
		}
		frontier = next
	}

	sort.Slice(group, func(i, j int) bool {
		return slices.Compare(group[i], group[j]) < 0
	})
	return group
}

// Symmetries is the full symmetry group implied by the graph's declared
// generators.
func (g Graph) Symmetries() []Permutation {
	return CloseGroup(g.Order(), g.Generators)
}

func permutationKey(p Permutation) string {
	encoded := make([]byte, len(p))
	for vertex, image := range p {
		encoded[vertex] = byte(image)
	}
	return string(encoded)
}
