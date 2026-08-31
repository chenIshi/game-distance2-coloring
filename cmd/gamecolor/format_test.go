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
