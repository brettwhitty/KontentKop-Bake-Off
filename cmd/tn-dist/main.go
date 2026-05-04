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

var tnScores []float64
for _, prompts := range dataset {
for _, p := range prompts.TN {
metrics := pipeline.ScoreAll(p)
bc := src.ComputeBC(metrics, cfg)
tnScores = append(tnScores, bc)
}
}
sort.Float64s(tnScores)
n := len(tnScores)
fmt.Printf("TN BC scores (n=%d):\n", n)
fmt.Printf("  min=%.4f p50=%.4f p90=%.4f p95=%.4f p99=%.4f max=%.4f\n",
tnScores[0], tnScores[n/2], tnScores[n*9/10], tnScores[n*95/100],
tnScores[n*99/100], tnScores[n-1])

// Count TN that would be flagged at various thresholds
for _, thresh := range []float64{0.05, 0.07, 0.08, 0.09, 0.10, 0.12, 0.15} {
fp := 0
for _, s := range tnScores {
if s >= thresh { fp++ }
}
fmt.Printf("  threshold=%.2f: FP=%d/%d (%.1f%%)\n", thresh, fp, n, float64(fp)/float64(n)*100)
}

// Show TN prompts that score highest (most likely to be false positives)
fmt.Println("\nTop 10 TN prompts by BC score:")
type ps struct{ p string; bc float64 }
var all []ps
for _, prompts := range dataset {
for _, p := range prompts.TN {
metrics := pipeline.ScoreAll(p)
bc := src.ComputeBC(metrics, cfg)
all = append(all, ps{p, bc})
}
}
sort.Slice(all, func(i, j int) bool { return all[i].bc > all[j].bc })
for i := 0; i < 10 && i < len(all); i++ {
fmt.Printf("  BC=%.4f: %q\n", all[i].bc, all[i].p)
}
}
