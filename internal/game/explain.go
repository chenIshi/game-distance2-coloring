package game

import "github.com/chenIshi/game-distance2-coloring/internal/graph"

// Explaining a result is separate from computing it. Analyze replays one
// concrete line of play so a human can see what actually happens; it is
// illustrative, not a proof of the whole strategy tree, and the CLI says so.
//
// Mirrors analyze_game in the Python reference, including which move each side
// picks, so the two narrate the same game.

// Step is one played move.
type Step struct {
	Player string // "Alice" or "Bob"
	Vertex int
	Color  int
}

// Analysis is one replayed game.
type Analysis struct {
	Winner  string
	Opening *Step // Alice's winning opening, when she has one
	Line    []Step
	Dead    []int // uncoloured vertices with no legal colour left
	Colors  []int // the position the line ended in
	Reach   graph.Graph
}

// Analyze replays a single line of play for the given colour count.
//
// Whoever is winning plays a move that preserves the win; whoever is losing
// plays the most constrained legal move. So a Bob win shows a concrete way
// Alice gets trapped, and an Alice win shows one way she gets home.
func Analyze(g graph.Graph, colorCount, distance int) (Analysis, error) {
	reach, err := graph.Power(g, distance)
	if err != nil {
		return Analysis{}, err
	}

	size := reach.Order()
	colors := make([]int, size)

	if colorCount <= 0 {
		winner := "Bob"
		var dead []int
		if size == 0 {
			winner = "Alice"
		} else {
			dead = make([]int, size)
			for i := range dead {
				dead[i] = i
			}
		}
		return Analysis{Winner: winner, Dead: dead, Colors: colors, Reach: reach}, nil
	}

	solver := NewSolver(reach, colorCount)
	aliceCanWin := solver.Solve(colors, true)

	analysis := Analysis{Colors: colors, Reach: reach}
	aliceTurn := true

	for {
		if allColored(colors) {
			analysis.Winner = "Alice"
			analysis.Colors = colors
			return analysis, nil
		}
		if dead := solver.DeadVertices(colors); len(dead) > 0 {
			analysis.Winner = "Bob"
			analysis.Dead = dead
			analysis.Colors = colors
			if !aliceCanWin {
				analysis.Opening = nil
			}
			return analysis, nil
		}

		moves := solver.orderedMoves(colors)
		if len(moves) == 0 {
			analysis.Winner = "Bob"
			analysis.Colors = colors
			return analysis, nil
		}

		vertex, color := -1, -1
		// The side that is winning searches for a move that keeps it winning.
		// The side that is losing has nothing to search for, so it plays the
		// most constrained legal move.
		searching := aliceTurn == aliceCanWin
		if searching {
			probe := make([]int, size)
		found:
			for _, move := range moves {
				for _, option := range move.Options {
					copy(probe, colors)
					probe[move.Vertex] = option
					if solver.Solve(probe, !aliceTurn) == aliceCanWin {
						vertex, color = move.Vertex, option
						break found
					}
				}
			}
		}
		if vertex < 0 {
			vertex, color = moves[0].Vertex, moves[0].Options[0]
		}

		colors[vertex] = color
		step := Step{Player: playerName(aliceTurn), Vertex: vertex, Color: color}
		if analysis.Opening == nil && aliceTurn && aliceCanWin {
			opening := step
			analysis.Opening = &opening
		}
		analysis.Line = append(analysis.Line, step)
		aliceTurn = !aliceTurn
	}
}

// Blockers lists the coloured neighbours of a vertex, which is what a reader
// needs in order to see why a dead vertex is dead.
func (a Analysis) Blockers(vertex int) []Step {
	var blockers []Step
	for _, neighbor := range a.Reach.Neighbors[vertex] {
		if color := a.Colors[neighbor]; color != 0 {
			blockers = append(blockers, Step{Vertex: neighbor, Color: color})
		}
	}
	return blockers
}

func allColored(colors []int) bool {
	for _, color := range colors {
		if color == 0 {
			return false
		}
	}
	return true
}

func playerName(aliceTurn bool) string {
	if aliceTurn {
		return "Alice"
	}
	return "Bob"
}
