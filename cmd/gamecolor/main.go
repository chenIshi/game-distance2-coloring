// Command gamecolor solves the distance-d colouring game on small graph
// families.
//
//	gamecolor -family cycle -n 7
//	gamecolor -family cycle -n 7 -k 4
//	gamecolor -family helm -n 4 -d 3
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/chenIshi/game-distance2-coloring/internal/game"
	"github.com/chenIshi/game-distance2-coloring/internal/graph"
)

type family struct {
	build func(int) (graph.Graph, error)
	label func(int) string
}

var families = map[string]family{
	"path":   {graph.Path, func(n int) string { return fmt.Sprintf("P_%d", n) }},
	"cycle":  {graph.Cycle, func(n int) string { return fmt.Sprintf("C_%d", n) }},
	"star":   {graph.Star, func(n int) string { return fmt.Sprintf("K_1,%d", n) }},
	"wheel":  {graph.Wheel, func(n int) string { return fmt.Sprintf("W_%d", n) }},
	"helm":   {graph.Helm, func(n int) string { return fmt.Sprintf("H_%d", n) }},
	"sunlet": {graph.Sunlet, func(n int) string { return fmt.Sprintf("S_%d", n) }},
}

func familyNames() string {
	names := make([]string, 0, len(families))
	for name := range families {
		names = append(names, name)
	}
	return strings.Join(names, ", ")
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "gamecolor:", err)
		os.Exit(1)
	}
}

func run() error {
	name := flag.String("family", "", "graph family: "+familyNames())
	size := flag.Int("n", 0, "size: vertices for path/cycle, rim or leaf count otherwise")
	distance := flag.Int("d", 2, "colouring distance")
	colorCount := flag.Int("k", 0, "test one colour count instead of finding the least")
	sweep := flag.Bool("sweep", false, "emit a JSON grid over every family instead of one case")
	maxOrder := flag.Int("max-order", 8, "sweep only: largest graph to include, in vertices")
	distances := flag.String("distances", "1,2,3", "sweep only: comma-separated distances")
	workers := flag.Int("workers", 0, "sweep only: parallel workers (0 = one per core)")
	format := flag.String("format", "json", "sweep only: json, table, or csv")
	explain := flag.Bool("explain", false, "with -k: replay one concrete line of play")
	flag.Parse()

	if *sweep {
		wanted, err := parseDistances(*distances)
		if err != nil {
			return err
		}
		return runSweep(os.Stdout, *maxOrder, wanted, *workers, *format)
	}

	chosen, ok := families[*name]
	if !ok {
		return fmt.Errorf("unknown family %q (want one of: %s)", *name, familyNames())
	}

	g, err := chosen.build(*size)
	if err != nil {
		return err
	}
	label := chosen.label(*size)

	if *colorCount > 0 {
		won, err := game.AliceWins(g, *colorCount, *distance)
		if err != nil {
			return err
		}
		fmt.Printf("%s with k=%d at d=%d: %s\n", label, *colorCount, *distance, winner(won))
		if *explain {
			analysis, err := game.Analyze(g, *colorCount, *distance)
			if err != nil {
				return err
			}
			fmt.Println()
			writeExplanation(os.Stdout, analysis, vertexNamer(*name, *size), *colorCount)
		}
		return nil
	}

	if *explain {
		return fmt.Errorf("-explain needs a colour count; add -k")
	}

	value, err := game.ChromaticNumber(g, *distance)
	if err != nil {
		return err
	}
	fmt.Printf("%s: chi_g,%d = %d\n", label, *distance, value)
	for count := 1; count <= value; count++ {
		won, err := game.AliceWins(g, count, *distance)
		if err != nil {
			return err
		}
		fmt.Printf("  k=%d: %s\n", count, winner(won))
	}
	return nil
}

func winner(aliceWins bool) string {
	if aliceWins {
		return "Alice"
	}
	return "Bob"
}
