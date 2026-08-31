package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strconv"
	"text/tabwriter"
)

// Output formats. JSON is the default because the conformance harness diffs it;
// the other two exist because a human reading a results grid wants a grid, and
// 40 lines of JSON per case is not one.
const (
	formatJSON  = "json"
	formatTable = "table"
	formatCSV   = "csv"
)

var formatNames = []string{formatJSON, formatTable, formatCSV}

func writeSweep(out io.Writer, format string, cases []Case) error {
	switch format {
	case formatJSON:
		encoder := json.NewEncoder(out)
		encoder.SetIndent("", " ")
		return encoder.Encode(sweepOutput{Cases: cases})
	case formatTable:
		return writeTable(out, cases)
	case formatCSV:
		return writeCSV(out, cases)
	default:
		return fmt.Errorf("unknown format %q (want one of: %v)", format, formatNames)
	}
}

// writeTable prints one grid per distance: families down the side, sizes along
// the top. Laid out this way because the interesting patterns are horizontal --
// a value dipping as n grows is the thing most likely to be mistaken for a bug,
// and it is only visible when the row is read across.
func writeTable(out io.Writer, cases []Case) error {
	distances := sortedDistances(cases)
	sizes := sortedSizes(cases)

	// chi[distance][family][n]
	chi := map[int]map[string]map[int]Case{}
	for _, c := range cases {
		if chi[c.D] == nil {
			chi[c.D] = map[string]map[int]Case{}
		}
		if chi[c.D][c.Family] == nil {
			chi[c.D][c.Family] = map[int]Case{}
		}
		chi[c.D][c.Family][c.N] = c
	}

	for _, distance := range distances {
		fmt.Fprintf(out, "distance %d\n", distance)

		writer := tabwriter.NewWriter(out, 0, 0, 2, ' ', tabwriter.AlignRight)
		fmt.Fprint(writer, "  family\t")
		for _, n := range sizes {
			fmt.Fprintf(writer, "%d\t", n)
		}
		fmt.Fprintln(writer)

		for _, family := range sweepFamilies {
			row, ok := chi[distance][family.name]
			if !ok {
				continue
			}
			fmt.Fprintf(writer, "  %s\t", family.name)
			for _, n := range sizes {
				c, ok := row[n]
				if !ok {
					fmt.Fprint(writer, ".\t")
					continue
				}
				marker := ""
				if c.Complete {
					marker = "*"
				}
				fmt.Fprintf(writer, "%d%s\t", c.Chi, marker)
			}
			fmt.Fprintln(writer)
		}
		if err := writer.Flush(); err != nil {
			return err
		}
		fmt.Fprintln(out)
	}

	fmt.Fprintln(out, "n is the size parameter: vertices for path/cycle, rim count otherwise.")
	fmt.Fprintln(out, "* means the power graph is complete, so chi = |V| with no search needed.")
	fmt.Fprintln(out, ". means that size does not exist for that family, or exceeds -max-order.")
	return nil
}

// writeCSV emits one row per case, for loading into a spreadsheet or pandas.
func writeCSV(out io.Writer, cases []Case) error {
	writer := csv.NewWriter(out)
	header := []string{
		"family", "n", "d", "vertices", "edges", "diameter",
		"power_edges", "complete", "symmetries", "chi",
	}
	if err := writer.Write(header); err != nil {
		return err
	}

	for _, c := range cases {
		row := []string{
			c.Family,
			strconv.Itoa(c.N),
			strconv.Itoa(c.D),
			strconv.Itoa(c.Order),
			strconv.Itoa(c.Edges),
			strconv.Itoa(c.Diameter),
			strconv.Itoa(c.PowerEdges),
			strconv.FormatBool(c.Complete),
			strconv.Itoa(c.GroupSize),
			strconv.Itoa(c.Chi),
		}
		if err := writer.Write(row); err != nil {
			return err
		}
	}

	writer.Flush()
	return writer.Error()
}

func sortedDistances(cases []Case) []int {
	seen := map[int]bool{}
	for _, c := range cases {
		seen[c.D] = true
	}
	return sortedKeys(seen)
}

func sortedSizes(cases []Case) []int {
	seen := map[int]bool{}
	for _, c := range cases {
		seen[c.N] = true
	}
	return sortedKeys(seen)
}

func sortedKeys(seen map[int]bool) []int {
	values := make([]int, 0, len(seen))
	for value := range seen {
		values = append(values, value)
	}
	sort.Ints(values)
	return values
}
