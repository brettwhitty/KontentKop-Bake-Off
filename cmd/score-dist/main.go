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

var tpScores []float64
for _, prompts := range dataset {
for _, p := range prompts.TP {
metrics := pipeline.ScoreAll(p)
bc := src.ComputeBC(metrics, cfg)
tpScores = append(tpScores, bc)
}
}

sort.Float64s(tpScores)
sum := 0.0
for _, s := range tpScores { sum += s }
n := len(tpScores)
fmt.Printf("TP BC scores (n=%d):\n", n)
fmt.Printf("  min=%.3f p10=%.3f p25=%.3f p50=%.3f p75=%.3f p90=%.3f p99=%.3f max=%.3f mean=%.3f\n",
tpScores[0], tpScores[n/10], tpScores[n/4], tpScores[n/2],
tpScores[n*3/4], tpScores[n*9/10], tpScores[n*99/100],
tpScores[n-1], sum/float64(n))

// Show top 10 scoring TP prompts
fmt.Println("\nTop 10 TP prompts by BC score:")
type ps struct{ p string; bc float64 }
var all []ps
for _, prompts := range dataset {
for _, p := range prompts.TP {
metrics := pipeline.ScoreAll(p)
bc := src.ComputeBC(metrics, cfg)
all = append(all, ps{p, bc})
}
}
sort.Slice(all, func(i, j int) bool { return all[i].bc > all[j].bc })
for i := 0; i < 10 && i < len(all); i++ {
fmt.Printf("  BC=%.3f: %q\n", all[i].bc, all[i].p)
}

// Show bottom 10 (hardest to detect)
fmt.Println("\nBottom 10 TP prompts (hardest to detect):")
for i := len(all)-1; i >= len(all)-10 && i >= 0; i-- {
fmt.Printf("  BC=%.3f: %q\n", all[i].bc, all[i].p)
}
}
