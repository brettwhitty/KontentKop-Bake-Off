// ingest-mentalmanip: convert MentalManip CSV (audreyeleven/MentalManip on
// HuggingFace, CC-BY-SA-4.0) into our fixture JSONL format.
//
// Source: .local/external-datasets/mentalmanip_con.csv (consensus subset)
// Output: bench/fixtures/mentalmanip.jsonl
//
// Schema mapping:
//   mentalmanip      → ours
//   ────────────────   ──────────────────────────────────────
//   id               → id (prefixed "mentalmanip-")
//   dialogue         → text (full multi-turn Person1/Person2 dialogue)
//   manipulative=1   → label "flag"
//   manipulative=0   → label "pass"
//   technique        → notes (the manipulation technique label)
//
// Length-class is computed from token count, same buckets as build-fixtures.
package main

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type Fixture struct {
	ID          string `json:"id"`
	Text        string `json:"text"`
	Label       string `json:"label"`
	LengthClass string `json:"length_class"`
	Source      string `json:"source"`
	Notes       string `json:"notes,omitempty"`
}

func classify(s string) string {
	n := len(strings.Fields(s))
	switch {
	case n < 50:
		return "short"
	case n <= 200:
		return "medium"
	default:
		return "long"
	}
}

func main() {
	in := ".local/external-datasets/mentalmanip_con.csv"
	out := "bench/fixtures/mentalmanip.jsonl"

	f, err := os.Open(in)
	if err != nil {
		fmt.Fprintf(os.Stderr, "open %s: %v\n", in, err)
		os.Exit(1)
	}
	defer f.Close()

	r := csv.NewReader(f)
	r.LazyQuotes = true
	r.FieldsPerRecord = -1 // tolerate quirks

	header, err := r.Read()
	if err != nil {
		fmt.Fprintf(os.Stderr, "header: %v\n", err)
		os.Exit(1)
	}
	col := map[string]int{}
	for i, name := range header {
		col[name] = i
	}
	for _, k := range []string{"id", "dialogue", "manipulative"} {
		if _, ok := col[k]; !ok {
			fmt.Fprintf(os.Stderr, "missing column %q in header: %v\n", k, header)
			os.Exit(1)
		}
	}

	if err := os.MkdirAll("bench/fixtures", 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "mkdir: %v\n", err)
		os.Exit(1)
	}
	outF, err := os.Create(out)
	if err != nil {
		fmt.Fprintf(os.Stderr, "create %s: %v\n", out, err)
		os.Exit(1)
	}
	defer outF.Close()
	w := bufio.NewWriter(outF)
	enc := json.NewEncoder(w)

	flag, pass, skipped := 0, 0, 0
	for {
		row, err := r.Read()
		if err != nil {
			break
		}
		text := strings.TrimSpace(row[col["dialogue"]])
		if text == "" {
			skipped++
			continue
		}
		manip := strings.TrimSpace(row[col["manipulative"]])
		var label string
		switch manip {
		case "1", "true", "True":
			label = "flag"
			flag++
		case "0", "false", "False":
			label = "pass"
			pass++
		default:
			skipped++
			continue
		}
		notes := ""
		if i, ok := col["technique"]; ok && i < len(row) {
			if t := strings.TrimSpace(row[i]); t != "" {
				notes = "mentalmanip technique: " + t
			}
		}
		fx := Fixture{
			ID:          "mentalmanip-" + strings.TrimSpace(row[col["id"]]),
			Text:        text,
			Label:       label,
			LengthClass: classify(text),
			Source:      "mentalmanip-con (audreyeleven/MentalManip, CC-BY-SA-4.0)",
			Notes:       notes,
		}
		if err := enc.Encode(fx); err != nil {
			fmt.Fprintf(os.Stderr, "encode: %v\n", err)
			os.Exit(1)
		}
	}
	w.Flush()

	fmt.Printf("wrote %s\n", out)
	fmt.Printf("  flag: %d  pass: %d  skipped: %d  total: %d\n",
		flag, pass, skipped, flag+pass)
}
