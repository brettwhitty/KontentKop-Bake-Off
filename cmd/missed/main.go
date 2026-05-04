package main

import (
"encoding/json"
"fmt"
"os"
"sort"

"kontentkop/src"
"kontentkop/src/pipeline"
)

func main() {
dataPath := ".local/failed-safety-project-led-by-gemini-3/reliability_dataset.json"
data, _ := os.ReadFile(dataPath)

var dataset map[string]struct {
TP []string `json:"tp"`
TN []string `json:"tn"`
}
json.Unmarshal(data, &dataset)

cfg := src.DefaultConfig()

type ps struct{ p string; bc float64 }
var missed []ps
for _, prompts := range dataset {
for _, p := range prompts.TP {
metrics := pipeline.ScoreAll(p)
bc := src.ComputeBC(metrics, cfg)
if bc < cfg.BCThreshold {
missed = append(missed, ps{p, bc})
}
}
}
sort.Slice(missed, func(i, j int) bool { return missed[i].bc > missed[j].bc })

// Deduplicate
seen := map[string]bool{}
fmt.Printf("Missed TP prompts (%d total, showing unique):\n", len(missed))
for _, m := range missed {
if !seen[m.p] {
seen[m.p] = true
fmt.Printf("  BC=%.3f: %q\n", m.bc, m.p)
}
}
}
