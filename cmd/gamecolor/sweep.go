package main

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"

	"github.com/chenIshi/game-distance2-coloring/internal/game"
	"github.com/chenIshi/game-distance2-coloring/internal/graph"
)

// Case is one (family, n, d) cell of a sweep.
//
// It carries far more than the answer on purpose. This is the payload the
// conformance harness diffs against the Python reference, so it reports
// everything the two implementations could disagree about: how the graph was
// built, what symmetries were declared for it, and every individual game
// verdict, not just the threshold.
type Case struct {
	Family string `json:"family"`
	N      int    `json:"n"`
	D      int    `json:"d"`

	// The base graph, independent of d.
	Order      int     `json:"order"`
	Edges      int     `json:"edges"`
	Diameter   int     `json:"diameter"` // -1 when disconnected
	Connected  bool    `json:"connected"`
	Neighbors  [][]int `json:"neighbors"`
	Generators [][]int `json:"generators"`
	GroupSize  int     `json:"groupSize"`

	// The distance-d power graph.
	PowerEdges int  `json:"powerEdges"`
	Complete   bool `json:"complete"`

	// The game. Verdicts covers k = 0..Chi, which the upward scan computes
	// anyway, so including them costs nothing and multiplies the number of
	// points the harness can compare.
	Chi      int    `json:"chi"`
	Verdicts []bool `json:"verdicts"`
}

type sweepOutput struct {
	Cases []Case `json:"cases"`
}

var sweepFamilies = []struct {
	name  string
	build func(int) (graph.Graph, error)
	from  int
}{
	{"path", graph.Path, 1},
	{"cycle", graph.Cycle, 3},
	{"star", graph.Star, 1},
	{"wheel", graph.Wheel, 3},
	{"helm", graph.Helm, 3},
	{"sunlet", graph.Sunlet, 3},
}

// parseDistances reads a comma-separated list like "1,2,3".
func parseDistances(spec string) ([]int, error) {
	var distances []int
	for _, field := range strings.Split(spec, ",") {
		field = strings.TrimSpace(field)
		if field == "" {
			continue
		}
		value, err := strconv.Atoi(field)
		if err != nil {
			return nil, fmt.Errorf("distances: %q is not a number", field)
		}
		distances = append(distances, value)
	}
	if len(distances) == 0 {
		return nil, fmt.Errorf("distances: no values given")
	}
	sort.Ints(distances)
	return distances, nil
}

// runSweep emits every case whose graph fits within maxOrder vertices.
func runSweep(out io.Writer, maxOrder int, distances []int) error {
	result := sweepOutput{Cases: []Case{}}

	for _, family := range sweepFamilies {
		for n := family.from; ; n++ {
			g, err := family.build(n)
			if err != nil {
				return err
			}
			if g.Order() > maxOrder {
				break
			}
			for _, distance := range distances {
				built, err := buildCase(family.name, n, distance, g)
				if err != nil {
					return err
				}
				result.Cases = append(result.Cases, built)
			}
		}
	}

	encoder := json.NewEncoder(out)
	encoder.SetIndent("", " ")
	return encoder.Encode(result)
}

func buildCase(family string, n, distance int, g graph.Graph) (Case, error) {
	reach, err := graph.Power(g, distance)
	if err != nil {
		return Case{}, err
	}

	span, connected := graph.Diameter(g)
	if !connected {
		span = -1
	}

	chi, err := game.ChromaticNumber(g, distance)
	if err != nil {
		return Case{}, err
	}

	verdicts := make([]bool, 0, chi+1)
	for colorCount := 0; colorCount <= chi; colorCount++ {
		won, err := game.AliceWins(g, colorCount, distance)
		if err != nil {
			return Case{}, err
		}
		verdicts = append(verdicts, won)
	}

	return Case{
		Family:     family,
		N:          n,
		D:          distance,
		Order:      g.Order(),
		Edges:      g.EdgeCount(),
		Diameter:   span,
		Connected:  connected,
		Neighbors:  neighborLists(g),
		Generators: permutationLists(g.Generators),
		GroupSize:  len(g.Symmetries()),
		PowerEdges: reach.EdgeCount(),
		Complete:   graph.IsComplete(reach),
		Chi:        chi,
		Verdicts:   verdicts,
	}, nil
}

// neighborLists normalises to non-nil slices so the JSON has [] not null.
func neighborLists(g graph.Graph) [][]int {
	lists := make([][]int, g.Order())
	for vertex, neighbors := range g.Neighbors {
		lists[vertex] = append([]int{}, neighbors...)
	}
	return lists
}

func permutationLists(permutations []graph.Permutation) [][]int {
	lists := make([][]int, 0, len(permutations))
	for _, p := range permutations {
		lists = append(lists, append([]int{}, p...))
	}
	return lists
}
