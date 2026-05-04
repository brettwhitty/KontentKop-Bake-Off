package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"kontentkop/src"
	"kontentkop/src/pipeline"
)

func main() {
	// Score the DarkPatterns-LLM test set
	// rejected = harmful (should flag), accepted = safe (should pass)
	testPath := ".local/failed-safety-project-led-by-gemini-3/safety/Benchmark-dataset-for-dark-patterns-in-llms/test.json"
	data, _ := os.ReadFile(testPath)

	var samples []struct {
		Category string `json:"category"`
		Rejected string `json:"rejected"`
		Accepted string `json:"accepted"`
	}
	json.Unmarshal(data, &samples)

	cfg := src.DefaultConfig()

	tpHit, tpTotal := 0, 0
	tnHit, tnTotal := 0, 0

	type catResult struct{ tpHit, tpTotal, tnHit, tnTotal int }
	cats := map[string]*catResult{}

	for _, s := range samples {
		if _, ok := cats[s.Category]; !ok {
			cats[s.Category] = &catResult{}
		}
		cr := cats[s.Category]

		if s.Rejected != "" {
			metrics := pipeline.ScoreAll(s.Rejected)
			bc := src.ComputeBC(metrics, cfg)
			tpTotal++
			cr.tpTotal++
			if bc >= cfg.BCThreshold {
				tpHit++
				cr.tpHit++
			}
		}
		if s.Accepted != "" {
			metrics := pipeline.ScoreAll(s.Accepted)
			bc := src.ComputeBC(metrics, cfg)
			tnTotal++
			cr.tnTotal++
			if bc < cfg.BCThreshold {
				tnHit++
				cr.tnHit++
			}
		}
	}

	fmt.Println("=== DarkPatterns-LLM Test Set Accuracy ===")
	fmt.Printf("%-25s  TP-Rate   TN-Rate\n", "Category")
	fmt.Println(strings.Repeat("-", 55))

	catNames := make([]string, 0, len(cats))
	for k := range cats {
		catNames = append(catNames, k)
	}
	sort.Strings(catNames)
	for _, name := range catNames {
		cr := cats[name]
		tpr, tnr := 0.0, 0.0
		if cr.tpTotal > 0 {
			tpr = float64(cr.tpHit) / float64(cr.tpTotal) * 100
		}
		if cr.tnTotal > 0 {
			tnr = float64(cr.tnHit) / float64(cr.tnTotal) * 100
		}
		fmt.Printf("%-25s  %5.1f%%    %5.1f%%\n", name, tpr, tnr)
	}
	fmt.Println(strings.Repeat("-", 55))
	fmt.Printf("\nTP: %d/%d (%.1f%%)  TN: %d/%d (%.1f%%)\n",
		tpHit, tpTotal, float64(tpHit)/float64(tpTotal)*100,
		tnHit, tnTotal, float64(tnHit)/float64(tnTotal)*100)

	prec := float64(tpHit) / float64(tpHit+(tnTotal-tnHit))
	rec := float64(tpHit) / float64(tpTotal)
	f1 := 2 * prec * rec / (prec + rec)
	fmt.Printf("F1: %.3f\n", f1)
}
