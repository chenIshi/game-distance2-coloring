package game

import (
	"testing"

	"github.com/chenIshi/game-distance2-coloring/internal/graph"
)

// A narration nobody checks is worse than none: it would read plausibly while
// describing an illegal game. These replay what Analyze reports and verify it
// against the rules.

func explainCases() []struct {
	label      string
	graph      graph.Graph
	colorCount int
	distance   int
} {
	return []struct {
		label      string
		graph      graph.Graph
		colorCount int
		distance   int
	}{
		{"P_2 k=1", mustGraph(graph.Path(2)), 1, 2},
		{"P_6 k=4", mustGraph(graph.Path(6)), 4, 2},
		{"C_6 k=4", mustGraph(graph.Cycle(6)), 4, 2},
		{"C_6 k=5", mustGraph(graph.Cycle(6)), 5, 2},
		{"C_7 k=4", mustGraph(graph.Cycle(7)), 4, 2},
		{"W_4 k=3 d=1", mustGraph(graph.Wheel(4)), 3, 1},
		{"W_5 k=3 d=1", mustGraph(graph.Wheel(5)), 3, 1},
		{"H_4 k=3 d=1", mustGraph(graph.Helm(4)), 3, 1},
		{"H_4 k=4 d=1", mustGraph(graph.Helm(4)), 4, 1},
		{"S_4 k=3 d=1", mustGraph(graph.Sunlet(4)), 3, 1},
		{"S_4 k=4 d=1", mustGraph(graph.Sunlet(4)), 4, 1},
		{"S_5 k=3 d=1", mustGraph(graph.Sunlet(5)), 3, 1},
	}
}

func TestAnalyzeWinnerAgreesWithTheSolver(t *testing.T) {
	for _, tc := range explainCases() {
		analysis, err := Analyze(tc.graph, tc.colorCount, tc.distance)
		if err != nil {
			t.Fatalf("%s: %v", tc.label, err)
		}
		won, err := AliceWins(tc.graph, tc.colorCount, tc.distance)
		if err != nil {
			t.Fatalf("%s: %v", tc.label, err)
		}
		want := "Bob"
		if won {
			want = "Alice"
		}
		if analysis.Winner != want {
			t.Errorf("%s: Analyze says %s, AliceWins says %s", tc.label, analysis.Winner, want)
		}
	}
}

func TestAnalyzeLineIsALegalAlternatingGame(t *testing.T) {
	for _, tc := range explainCases() {
		analysis, err := Analyze(tc.graph, tc.colorCount, tc.distance)
		if err != nil {
			t.Fatalf("%s: %v", tc.label, err)
		}

		reach, _ := graph.Power(tc.graph, tc.distance)
		solver := NewSolver(reach, tc.colorCount)
		colors := make([]int, reach.Order())

		for index, step := range analysis.Line {
			wantPlayer := "Alice"
			if index%2 == 1 {
				wantPlayer = "Bob"
			}
			if step.Player != wantPlayer {
				t.Errorf("%s: move %d played by %s, expected %s",
					tc.label, index, step.Player, wantPlayer)
			}
			if colors[step.Vertex] != 0 {
				t.Errorf("%s: move %d recolours vertex %d", tc.label, index, step.Vertex)
			}
			legal := false
			for _, option := range solver.LegalColors(colors, step.Vertex) {
				if option == step.Color {
					legal = true
				}
			}
			if !legal {
				t.Errorf("%s: move %d plays colour %d on vertex %d, which was not legal",
					tc.label, index, step.Color, step.Vertex)
			}
			colors[step.Vertex] = step.Color
		}

		for vertex := range colors {
			if colors[vertex] != analysis.Colors[vertex] {
				t.Errorf("%s: replaying the line does not reproduce the reported position",
					tc.label)
				break
			}
		}
	}
}

func TestAnalyzeAliceWinEndsInAProperColouring(t *testing.T) {
	for _, tc := range explainCases() {
		analysis, err := Analyze(tc.graph, tc.colorCount, tc.distance)
		if err != nil || analysis.Winner != "Alice" {
			continue
		}
		if len(analysis.Dead) != 0 {
			t.Errorf("%s: Alice won but dead vertices were reported", tc.label)
		}
		for vertex, color := range analysis.Colors {
			if color == 0 {
				t.Errorf("%s: Alice won but vertex %d is uncoloured", tc.label, vertex)
			}
			for _, neighbor := range analysis.Reach.Neighbors[vertex] {
				if analysis.Colors[neighbor] == color {
					t.Errorf("%s: final colouring is improper at %d-%d", tc.label, vertex, neighbor)
				}
			}
		}
		if analysis.Opening == nil {
			t.Errorf("%s: Alice won but no winning opening was reported", tc.label)
		} else if len(analysis.Line) > 0 && *analysis.Opening != analysis.Line[0] {
			t.Errorf("%s: reported opening is not the first move played", tc.label)
		}
	}
}

func TestAnalyzeBobWinReportsAGenuinelyDeadVertex(t *testing.T) {
	for _, tc := range explainCases() {
		analysis, err := Analyze(tc.graph, tc.colorCount, tc.distance)
		if err != nil || analysis.Winner != "Bob" {
			continue
		}
		if len(analysis.Dead) == 0 {
			t.Errorf("%s: Bob won but no dead vertex was reported", tc.label)
			continue
		}
		reach, _ := graph.Power(tc.graph, tc.distance)
		solver := NewSolver(reach, tc.colorCount)
		for _, vertex := range analysis.Dead {
			if analysis.Colors[vertex] != 0 {
				t.Errorf("%s: reported dead vertex %d is coloured", tc.label, vertex)
			}
			if left := solver.LegalColors(analysis.Colors, vertex); len(left) != 0 {
				t.Errorf("%s: vertex %d called dead but still has colours %v",
					tc.label, vertex, left)
			}
			// The blockers must actually account for every colour.
			seen := map[int]bool{}
			for _, blocker := range analysis.Blockers(vertex) {
				seen[blocker.Color] = true
			}
			if len(seen) != tc.colorCount {
				t.Errorf("%s: vertex %d blocked by %d distinct colours, expected %d",
					tc.label, vertex, len(seen), tc.colorCount)
			}
		}
	}
}

func TestAnalyzeZeroColours(t *testing.T) {
	analysis, err := Analyze(mustGraph(graph.Path(3)), 0, 2)
	if err != nil {
		t.Fatal(err)
	}
	if analysis.Winner != "Bob" {
		t.Errorf("winner = %s, want Bob", analysis.Winner)
	}
	empty, err := Analyze(mustGraph(graph.Path(0)), 0, 2)
	if err != nil {
		t.Fatal(err)
	}
	if empty.Winner != "Alice" {
		t.Errorf("empty graph winner = %s, want Alice", empty.Winner)
	}
}

// TestAnalyzeMatchesTheKnownHelmKill pins the concrete line that shows why a
// pendant costs a colour: the rim vertex dies to its two rim neighbours plus its
// own stub, without the rim's alternating pattern ever being broken.
func TestAnalyzeMatchesTheKnownHelmKill(t *testing.T) {
	analysis, err := Analyze(mustGraph(graph.Helm(4)), 3, 1)
	if err != nil {
		t.Fatal(err)
	}
	if analysis.Winner != "Bob" {
		t.Fatalf("winner = %s, want Bob", analysis.Winner)
	}
	if len(analysis.Dead) != 1 {
		t.Fatalf("dead = %v, want exactly one", analysis.Dead)
	}
	dead := analysis.Dead[0]
	if dead < 1 || dead > 4 {
		t.Errorf("dead vertex %d is not a rim vertex", dead)
	}
	// Its own stub must be among the blockers.
	stub := 4 + dead
	found := false
	for _, blocker := range analysis.Blockers(dead) {
		if blocker.Vertex == stub {
			found = true
		}
	}
	if !found {
		t.Errorf("rim %d died without its own stub %d contributing: %v",
			dead, stub, analysis.Blockers(dead))
	}
}
