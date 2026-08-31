//go:build js && wasm

// Command wasmsolver exposes the solver to a web page.
//
// The browser gets the same solver the CLI uses, compiled for js/wasm rather
// than reimplemented. Before this existed the web demo carried a hand-written
// JavaScript copy of the rules, which had to be kept in step by hand and would
// disagree silently if anyone forgot.
//
// Values cross the boundary as plain JS objects rather than JSON strings.
// encoding/json would be the obvious choice and costs about 500 KB gzipped
// here, because it pulls reflection into a binary the visitor has to download.
// syscall/js converts the handful of types below with a type switch instead.
package main

import (
	"fmt"
	"syscall/js"

	"github.com/chenIshi/game-distance2-coloring/internal/game"
	"github.com/chenIshi/game-distance2-coloring/internal/graph"
)

var builders = map[string]func(int) (graph.Graph, error){
	"path":   graph.Path,
	"cycle":  graph.Cycle,
	"star":   graph.Star,
	"wheel":  graph.Wheel,
	"helm":   graph.Helm,
	"sunlet": graph.Sunlet,
}

// solverCache keeps one Solver per (family, n, d, k). The memo is the whole
// point of the solver, so discarding it between moves would make every move pay
// for the search again.
var solverCache = map[string]*game.Solver{}

func build(family string, n, distance int) (graph.Graph, graph.Graph, error) {
	builder, ok := builders[family]
	if !ok {
		return graph.Graph{}, graph.Graph{}, fmt.Errorf("unknown family %q", family)
	}
	g, err := builder(n)
	if err != nil {
		return graph.Graph{}, graph.Graph{}, err
	}
	reach, err := graph.Power(g, distance)
	if err != nil {
		return graph.Graph{}, graph.Graph{}, err
	}
	return g, reach, nil
}

func failure(message string) map[string]any {
	return map[string]any{"error": message}
}

func lists(g graph.Graph) []any {
	out := make([]any, g.Order())
	for vertex, neighbors := range g.Neighbors {
		row := make([]any, len(neighbors))
		for i, neighbor := range neighbors {
			row[i] = neighbor
		}
		out[vertex] = row
	}
	return out
}

func moveList(moves [][2]int) []any {
	out := make([]any, len(moves))
	for i, m := range moves {
		out[i] = map[string]any{"vertex": m[0], "color": m[1]}
	}
	return out
}

func handleGraph(family string, n, distance int) map[string]any {
	g, reach, err := build(family, n, distance)
	if err != nil {
		return failure(err.Error())
	}
	span, connected := graph.Diameter(g)
	if !connected {
		span = -1
	}
	return map[string]any{
		"order":     g.Order(),
		"neighbors": lists(g),
		"reach":     lists(reach),
		"diameter":  span,
		"complete":  graph.IsComplete(reach),
	}
}

func handleChi(family string, n, distance int) map[string]any {
	g, _, err := build(family, n, distance)
	if err != nil {
		return failure(err.Error())
	}
	value, err := game.ChromaticNumber(g, distance)
	if err != nil {
		return failure(err.Error())
	}
	return map[string]any{"chi": value}
}

func handleState(family string, n, distance, colorCount int, colors []int, aliceTurn bool) map[string]any {
	_, reach, err := build(family, n, distance)
	if err != nil {
		return failure(err.Error())
	}
	if colorCount <= 0 {
		return failure("colour count must be positive")
	}
	if len(colors) != reach.Order() {
		return failure(fmt.Sprintf("expected %d colours, got %d", reach.Order(), len(colors)))
	}

	key := fmt.Sprintf("%s|%d|%d|%d", family, n, distance, colorCount)
	solver, ok := solverCache[key]
	if !ok {
		solver = game.NewSolver(reach, colorCount)
		solverCache[key] = solver
	}

	finished := true
	for _, color := range colors {
		if color == 0 {
			finished = false
			break
		}
	}

	dead := solver.DeadVertices(colors)
	deadList := make([]any, len(dead))
	blockers := map[string]any{}
	for i, vertex := range dead {
		deadList[i] = vertex
		var found []any
		for _, neighbor := range reach.Neighbors[vertex] {
			if color := colors[neighbor]; color != 0 {
				found = append(found, map[string]any{"vertex": neighbor, "color": color})
			}
		}
		blockers[fmt.Sprint(vertex)] = found
	}

	var legal, optimal [][2]int
	if !finished && len(dead) == 0 {
		next := make([]int, len(colors))
		for vertex := range colors {
			for _, color := range solver.LegalColors(colors, vertex) {
				legal = append(legal, [2]int{vertex, color})
				copy(next, colors)
				next[vertex] = color
				// A move is optimal when it keeps the mover winning.
				if solver.Solve(next, !aliceTurn) == aliceTurn {
					optimal = append(optimal, [2]int{vertex, color})
				}
			}
		}
	}

	return map[string]any{
		"finished":     finished,
		"dead":         deadList,
		"blockers":     blockers,
		"aliceWins":    solver.Solve(colors, aliceTurn),
		"legalMoves":   moveList(legal),
		"optimalMoves": moveList(optimal),
	}
}

func dispatch(req js.Value) map[string]any {
	op := req.Get("op").String()
	family := req.Get("family").String()
	n := req.Get("n").Int()
	distance := req.Get("d").Int()

	switch op {
	case "graph":
		return handleGraph(family, n, distance)
	case "chi":
		return handleChi(family, n, distance)
	case "state":
		raw := req.Get("colors")
		colors := make([]int, raw.Length())
		for i := range colors {
			colors[i] = raw.Index(i).Int()
		}
		return handleState(family, n, distance, req.Get("k").Int(), colors, req.Get("aliceTurn").Bool())
	}
	return failure(fmt.Sprintf("unknown op %q", op))
}

func main() {
	js.Global().Set("gameColoringSolve", js.FuncOf(func(_ js.Value, args []js.Value) any {
		if len(args) != 1 {
			return js.ValueOf(failure("expected one request object"))
		}
		return js.ValueOf(dispatch(args[0]))
	}))
	js.Global().Set("gameColoringReady", true)
	select {}
}
