// Package graph builds the graph families this project studies, together with
// the symmetries each family is known to have.
//
// A Graph carries its own symmetry generators because the builder is the only
// thing that knows the recipe: the "% n" in Cycle *is* the rotation, written
// down. Discovering symmetries from a bare adjacency list is a hard general
// problem; declaring them at construction is free.
package graph

import (
	"errors"
	"fmt"
	"sort"
)

// Permutation relabels vertices: p[v] is the new name of vertex v.
type Permutation []int

// Graph is an undirected simple graph plus the symmetries its builder declared.
//
// Generators is a small starter set, not the whole collection; CloseGroup
// expands it, the way six face turns generate every position of a cube.
type Graph struct {
	Neighbors  [][]int // sorted ascending, never contains the vertex itself
	Generators []Permutation
}

// ErrNegative reports a size argument that cannot describe a graph.
var ErrNegative = errors.New("size must be non-negative")

// Order is the number of vertices.
func (g Graph) Order() int { return len(g.Neighbors) }

// Degree is the number of neighbours of v.
func (g Graph) Degree(v int) int { return len(g.Neighbors[v]) }

// Adjacent reports whether u and v are joined.
func (g Graph) Adjacent(u, v int) bool {
	for _, w := range g.Neighbors[u] {
		if w == v {
			return true
		}
	}
	return false
}

// EdgeCount is the number of undirected edges.
func (g Graph) EdgeCount() int {
	total := 0
	for _, neighbors := range g.Neighbors {
		total += len(neighbors)
	}
	return total / 2
}

// fromSets normalises adjacency sets into the sorted form Graph expects, and
// verifies the result is a simple undirected graph.
func fromSets(adjacency []map[int]bool, generators []Permutation) Graph {
	neighbors := make([][]int, len(adjacency))
	for vertex, set := range adjacency {
		list := make([]int, 0, len(set))
		for neighbor := range set {
			if neighbor != vertex {
				list = append(list, neighbor)
			}
		}
		sort.Ints(list)
		neighbors[vertex] = list
	}
	return Graph{Neighbors: neighbors, Generators: generators}
}

func emptyAdjacency(size int) []map[int]bool {
	adjacency := make([]map[int]bool, size)
	for i := range adjacency {
		adjacency[i] = map[int]bool{}
	}
	return adjacency
}

func join(adjacency []map[int]bool, u, v int) {
	adjacency[u][v] = true
	adjacency[v][u] = true
}

// Power joins every pair of vertices lying within distance of each other.
//
// Distance-d colouring on g is ordinary colouring on Power(g, d), so the solver
// never has to reason about distance itself.
//
// The symmetries carry over unchanged: relabelling cannot change distances, so
// a symmetry of g is automatically a symmetry of every power of g.
func Power(g Graph, distance int) (Graph, error) {
	if distance < 0 {
		return Graph{}, fmt.Errorf("power: distance %d: %w", distance, ErrNegative)
	}

	size := g.Order()
	adjacency := emptyAdjacency(size)

	for source := 0; source < size; source++ {
		reached := map[int]bool{source: true}
		frontier := []int{source}

		for step := 0; step < distance && len(frontier) > 0; step++ {
			var next []int
			for _, vertex := range frontier {
				for _, neighbor := range g.Neighbors[vertex] {
					if !reached[neighbor] {
						reached[neighbor] = true
						next = append(next, neighbor)
					}
				}
			}
			frontier = next
		}

		for vertex := range reached {
			if vertex != source {
				adjacency[source][vertex] = true
			}
		}
	}

	return fromSets(adjacency, g.Generators), nil
}

// IsComplete reports whether every vertex is joined to every other one.
//
// When the power graph is complete the colouring game is decided without any
// search: every move burns a distinct colour, so Alice wins exactly when there
// are at least as many colours as vertices.
func IsComplete(g Graph) bool {
	for _, neighbors := range g.Neighbors {
		if len(neighbors) != g.Order()-1 {
			return false
		}
	}
	return true
}

// Diameter is the largest distance between any two vertices. The second result
// is false when the graph is disconnected, in which case no diameter exists.
func Diameter(g Graph) (int, bool) {
	size := g.Order()
	if size == 0 {
		return 0, true
	}

	best := 0
	for source := 0; source < size; source++ {
		distances := make([]int, size)
		for i := range distances {
			distances[i] = -1
		}
		distances[source] = 0
		queue := []int{source}
		seen := 1

		for len(queue) > 0 {
			vertex := queue[0]
			queue = queue[1:]
			for _, neighbor := range g.Neighbors[vertex] {
				if distances[neighbor] < 0 {
					distances[neighbor] = distances[vertex] + 1
					seen++
					queue = append(queue, neighbor)
				}
			}
		}

		if seen != size {
			return 0, false
		}
		for _, d := range distances {
			if d > best {
				best = d
			}
		}
	}
	return best, true
}
