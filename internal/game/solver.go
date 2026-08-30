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
	"math/bits"
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
//
// Positions are memoised under an integer key whenever the whole position fits
// in a uint64: the colour of each vertex packed into just enough bits, plus one
// bit for whose turn it is. That avoids allocating a key on every lookup, which
// dominates the runtime of an otherwise identical search. Larger cases fall
// back to a byte-string key, which allocates but has no size limit.
type Solver struct {
	reach      graph.Graph
	colorCount int

	bitsPerVertex uint
	packable      bool
	packedMemo    map[uint64]bool
	stringMemo    map[string]bool

	// canonical can be turned off to get the plain search back. Only tests do
	// that, to check the optimisation against the thing it replaced.
	canonical bool

	// relabel is scratch for canonicalColors. Safe to share across the
	// recursion because it is only live inside that call, which always
	// finishes before any recursive Solve begins.
	relabel []int
}

// NewSolver builds a solver over an already-powered graph.
func NewSolver(reach graph.Graph, colorCount int) *Solver {
	return newSolver(reach, colorCount, true)
}

// newSolver builds a solver with colour canonicalisation optionally disabled,
// which exists so tests can compare the optimised search against the plain one.
func newSolver(reach graph.Graph, colorCount int, canonical bool) *Solver {
	// Colours run 0 (uncoloured) to colorCount, so the widest value is
	// colorCount itself.
	width := uint(bits.Len(uint(colorCount)))
	if width == 0 {
		width = 1
	}

	solver := &Solver{
		reach:         reach,
		colorCount:    colorCount,
		bitsPerVertex: width,
		packable:      width*uint(reach.Order())+1 <= 64,
		canonical:     canonical,
	}
	if solver.packable {
		solver.packedMemo = map[uint64]bool{}
	} else {
		solver.stringMemo = map[string]bool{}
	}
	solver.relabel = make([]int, colorCount+1)
	return solver
}

// canonicalColors renames colours in place by order of first appearance, so
// that 0,2,0,3 and 0,3,0,1 both become 0,1,0,2.
//
// Sound because the rules treat colours symmetrically: permuting the colour
// names maps any position onto an equivalent one with the same winner. Two
// positions that differ only by such a permutation share a first-appearance
// order, so they canonicalise to the same thing and are searched once instead
// of once per naming.
//
// This is where the real reduction comes from. A position using all k colours
// stands for k! differently-named positions, and without this the search
// explores every one of them.
func (s *Solver) canonicalColors(colors []int) {
	for i := range s.relabel {
		s.relabel[i] = 0
	}

	next := 1
	for vertex, color := range colors {
		if color == 0 {
			continue
		}
		if s.relabel[color] == 0 {
			s.relabel[color] = next
			next++
		}
		colors[vertex] = s.relabel[color]
	}
}

// packKey encodes a position into a single integer. Only valid when packable.
func (s *Solver) packKey(colors []int, aliceTurn bool) uint64 {
	var key uint64
	for _, color := range colors {
		key = key<<s.bitsPerVertex | uint64(color)
	}
	key <<= 1
	if aliceTurn {
		key |= 1
	}
	return key
}

// stringKey is the unbounded fallback: one byte per vertex, plus the turn.
func (s *Solver) stringKey(colors []int, aliceTurn bool) string {
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
	if s.packable {
		key := s.packKey(colors, aliceTurn)
		if cached, ok := s.packedMemo[key]; ok {
			return cached
		}
		result := s.solveUncached(colors, aliceTurn)
		s.packedMemo[key] = result
		return result
	}

	key := s.stringKey(colors, aliceTurn)
	if cached, ok := s.stringMemo[key]; ok {
		return cached
	}
	result := s.solveUncached(colors, aliceTurn)
	s.stringMemo[key] = result
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
			if s.canonical {
				s.canonicalColors(next)
			}
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
func (s *Solver) Positions() int {
	if s.packable {
		return len(s.packedMemo)
	}
	return len(s.stringMemo)
}

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

	// The all-uncoloured start position is already canonical.
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
