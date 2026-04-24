// score-one: emit per-metric scores + BC for a single prompt regardless
// of threshold. Useful for inspection when the silent-pass-through
// behaviour of the main binary hides under-threshold detail.
package main

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"kontentkop/src"
	"kontentkop/src/pipeline"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: score-one <prompt>")
		os.Exit(2)
	}
	text := strings.Join(os.Args[1:], " ")

	cfg, err := src.LoadConfig("config.yaml")
	if err != nil {
		cfg = src.DefaultConfig()
	}
	metrics := pipeline.ScoreAll(text)
	bc := src.ComputeBC(metrics, cfg)

	keys := make([]string, 0, len(metrics))
	for k := range metrics {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	fmt.Printf("=== %q ===\n", text)
	for _, k := range keys {
		m := metrics[k]
		if m.Score > 0 {
			fmt.Printf("  %-20s = %.4f  (spans: %d)\n", k, m.Score, len(m.Spans))
		}
	}
	fmt.Printf("\n  BC = %.4f  (threshold: %.2f)\n", bc, cfg.BCThreshold)
	if bc >= cfg.BCThreshold {
		fmt.Println("  STATUS: FLAGGED")
	} else {
		fmt.Println("  STATUS: PASS")
	}
}
