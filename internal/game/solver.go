// Package game solves the distance-d colouring game.
//
// Two players alternate, Alice first. A move gives an uncolored vertex a colour
// not already used within distance d. Alice (the maker) wins if every vertex
// gets coloured; Bob (the breaker) wins the moment some uncoloured vertex has
// no legal colour left -- a dead vertex.
//
// All solving happens on the distance-d power graph, so nothing below reasons
// about distance: it is ordinary colouring on a denser graph.
//
// This is a direct port of the Python reference implementation in
// src/game_coloring/solver.py and must agree with it exactly.
package game

import (
	"errors"
	"fmt"
	"sort"

	"github.com/chenIshi/game-distance2-coloring/internal/graph"
)

// maxColors bounds the colour count so a position packs into one byte per
// vertex. Far beyond anything solvable.
const maxColors = 255

// ErrTooManyColors reports a colour count that cannot be encoded.
var ErrTooManyColors = errors.New("colour count exceeds the encodable maximum")

// Move is a vertex together with every colour legally available to it.
type Move struct {
	Vertex  int
	Options []int
}

// Solver answers "does Alice win from here?" for one (reach graph, colour
// count) pair. The memo is only valid for that pair, so build one per pair.
type Solver struct {
	reach      graph.Graph
	colorCount int
	memo       map[string]bool
}

// NewSolver builds a solver over an already-powered graph.
func NewSolver(reach graph.Graph, colorCount int) *Solver {
	return &Solver{reach: reach, colorCount: colorCount, memo: map[string]bool{}}
}

// key encodes a position. Colours fit in a byte; the final byte is the turn.
func (s *Solver) key(colors []int, aliceTurn bool) string {
	encoded := make([]byte, len(colors)+1)
	for vertex, color := range colors {
		encoded[vertex] = byte(color)
	}
	if aliceTurn {
		encoded[len(colors)] = 1
	}
	return string(encoded)
}

// LegalColors lists the colours vertex may still take. An already-coloured
// vertex has none.
func (s *Solver) LegalColors(colors []int, vertex int) []int {
	if colors[vertex] != 0 {
		return nil
	}

	blocked := make([]bool, s.colorCount+1)
	for _, neighbor := range s.reach.Neighbors[vertex] {
		if color := colors[neighbor]; color != 0 {
			blocked[color] = true
		}
	}

	var options []int
	for color := 1; color <= s.colorCount; color++ {
		if !blocked[color] {
			options = append(options, color)
		}
	}
	return options
}

// DeadVertices lists uncoloured vertices with no legal colour left. A non-empty
// result means Bob has already won.
func (s *Solver) DeadVertices(colors []int) []int {
	var dead []int
	for vertex, color := range colors {
		if color == 0 && len(s.LegalColors(colors, vertex)) == 0 {
			dead = append(dead, vertex)
		}
	}
	return dead
}

func (s *Solver) hasDeadVertex(colors []int) bool {
	for vertex, color := range colors {
		if color == 0 && len(s.LegalColors(colors, vertex)) == 0 {
			return true
		}
	}
	return false
}

// orderedMoves lists every legal move, most-constrained vertex first.
//
// Ordering never changes the answer, only how quickly the search reaches it.
// Kept identical to the Python reference so the two explore in the same order.
func (s *Solver) orderedMoves(colors []int) []Move {
	var moves []Move
	for vertex, color := range colors {
		if color != 0 {
			continue
		}
		if options := s.LegalColors(colors, vertex); len(options) > 0 {
			moves = append(moves, Move{Vertex: vertex, Options: options})
		}
	}

	sort.SliceStable(moves, func(i, j int) bool {
		if len(moves[i].Options) != len(moves[j].Options) {
			return len(moves[i].Options) < len(moves[j].Options)
		}
		return moves[i].Vertex < moves[j].Vertex
	})
	return moves
}

// Solve reports whether Alice wins from this position.
//
// Alice needs one door; Bob needs to close every door. She wins if any of her
// moves leads to a win; he wins if any of his leads to a loss for her.
func (s *Solver) Solve(colors []int, aliceTurn bool) bool {
	cacheKey := s.key(colors, aliceTurn)
	if cached, ok := s.memo[cacheKey]; ok {
		return cached
	}

	result := s.solveUncached(colors, aliceTurn)
	s.memo[cacheKey] = result
	return result
}

func (s *Solver) solveUncached(colors []int, aliceTurn bool) bool {
	complete := true
	for _, color := range colors {
		if color == 0 {
			complete = false
			break
		}
	}
	if complete {
		return true
	}

	if s.hasDeadVertex(colors) {
		return false
	}

	next := make([]int, len(colors))
	for _, move := range s.orderedMoves(colors) {
		for _, color := range move.Options {
			copy(next, colors)
			next[move.Vertex] = color
			if s.Solve(next, !aliceTurn) == aliceTurn {
				// Alice found a winning move, or Bob found a losing one.
				return aliceTurn
			}
		}
	}
	return !aliceTurn
}

// Positions is the number of distinct positions the memo holds. Useful for
// comparing search cost.
func (s *Solver) Positions() int { return len(s.memo) }

// AliceWins reports whether Alice wins the distance-d game on g with the given
// number of colours.
func AliceWins(g graph.Graph, colorCount, distance int) (bool, error) {
	if colorCount > maxColors {
		return false, fmt.Errorf("alice wins: %d colours: %w", colorCount, ErrTooManyColors)
	}
	if colorCount <= 0 {
		return g.Order() == 0, nil
	}

	reach, err := graph.Power(g, distance)
	if err != nil {
		return false, err
	}

	if graph.IsComplete(reach) {
		// Every vertex blocks every other, so each move burns a distinct colour
		// and nobody is ever trapped while colours remain. This retires whole
		// families: wheels and stars have diameter 2, so they are complete at
		// every d >= 2.
		return colorCount >= reach.Order(), nil
	}

	return NewSolver(reach, colorCount).Solve(make([]int, reach.Order()), true), nil
}

// ChromaticNumber is the least number of colours for which Alice wins.
//
// Counts upward on purpose. Monotonicity in k holds everywhere measured, which
// would make a binary search legal, but solve cost grows steeply with k: this
// scan stops at the threshold and never pays for the expensive large-k solves,
// while a binary search would probe above it.
func ChromaticNumber(g graph.Graph, distance int) (int, error) {
	reach, err := graph.Power(g, distance)
	if err != nil {
		return 0, err
	}
	if graph.IsComplete(reach) {
		return g.Order(), nil
	}

	// Starts at zero so the empty graph reports 0. For a non-empty graph k=0
	// always loses, so this costs one extra trivial check.
	for colorCount := 0; colorCount <= g.Order(); colorCount++ {
		won, err := AliceWins(g, colorCount, distance)
		if err != nil {
			return 0, err
		}
		if won {
			return colorCount, nil
		}
	}

	// Unreachable: with k = |V| no vertex can ever be blocked, because a vertex
	// has at most |V|-1 neighbours and so cannot see all |V| colours.
	return g.Order() + 1, nil
}
