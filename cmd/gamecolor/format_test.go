package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"strings"
	"testing"
)

func sweepInto(t *testing.T, format string) string {
	t.Helper()
	var buffer bytes.Buffer
	if err := runSweep(&buffer, 7, []int{1, 2}, 1, format); err != nil {
		t.Fatalf("%s: %v", format, err)
	}
	return buffer.String()
}

// TestJSONStaysParseable guards the format the conformance harness diffs.
func TestJSONStaysParseable(t *testing.T) {
	var parsed sweepOutput
	if err := json.Unmarshal([]byte(sweepInto(t, formatJSON)), &parsed); err != nil {
		t.Fatal(err)
	}
	if len(parsed.Cases) == 0 {
		t.Fatal("no cases emitted")
	}
	for _, c := range parsed.Cases {
		if len(c.Verdicts) != c.Chi+1 {
			t.Errorf("%s_%d d=%d: %d verdicts for chi=%d",
				c.Family, c.N, c.D, len(c.Verdicts), c.Chi)
		}
		if c.Chi > 0 && !c.Verdicts[c.Chi] {
			t.Errorf("%s_%d d=%d: Alice should win at k=chi", c.Family, c.N, c.D)
		}
	}
}

func TestCSVHasOneRowPerCase(t *testing.T) {
	var parsed sweepOutput
	if err := json.Unmarshal([]byte(sweepInto(t, formatJSON)), &parsed); err != nil {
		t.Fatal(err)
	}

	records, err := csv.NewReader(strings.NewReader(sweepInto(t, formatCSV))).ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	if got, want := len(records), len(parsed.Cases)+1; got != want {
		t.Errorf("%d CSV rows (incl. header), want %d", got, want)
	}
	if records[0][0] != "family" || records[0][len(records[0])-1] != "chi" {
		t.Errorf("unexpected header: %v", records[0])
	}
}

func TestTableCoversEveryDistanceAndFamily(t *testing.T) {
	table := sweepInto(t, formatTable)
	for _, want := range []string{"distance 1", "distance 2", "path", "cycle", "wheel", "helm", "sunlet"} {
		if !strings.Contains(table, want) {
			t.Errorf("table is missing %q", want)
		}
	}
	// The legend explains the two markers, which are meaningless without it.
	for _, want := range []string{"* means", ". means"} {
		if !strings.Contains(table, want) {
			t.Errorf("table is missing the legend entry %q", want)
		}
	}
}

func TestUnknownFormatIsRejected(t *testing.T) {
	var buffer bytes.Buffer
	if err := runSweep(&buffer, 5, []int{2}, 1, "yaml"); err == nil {
		t.Fatal("expected an error for an unknown format")
	}
}

func TestVertexNamesFollowTheNumberingContract(t *testing.T) {
	// The names encode CLAUDE.md's numbering, so a drift here is a drift from
	// the builders and the declared symmetries.
	helm := vertexNamer("helm", 4)
	for vertex, want := range map[int]string{0: "hub", 1: "rim1", 4: "rim4", 5: "stub1", 8: "stub4"} {
		if got := helm(vertex); got != want {
			t.Errorf("helm vertex %d named %q, want %q", vertex, got, want)
		}
	}
	sunlet := vertexNamer("sunlet", 4)
	for vertex, want := range map[int]string{0: "rim0", 3: "rim3", 4: "stub0", 7: "stub3"} {
		if got := sunlet(vertex); got != want {
			t.Errorf("sunlet vertex %d named %q, want %q", vertex, got, want)
		}
	}
	if got := vertexNamer("wheel", 5)(0); got != "hub" {
		t.Errorf("wheel vertex 0 named %q, want hub", got)
	}
	if got := vertexNamer("path", 5)(2); got != "v2" {
		t.Errorf("path vertex 2 named %q, want v2", got)
	}
}
