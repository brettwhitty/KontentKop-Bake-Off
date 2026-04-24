package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"sort"
	"strings"
	"time"

	"kontentkop/src"
	"kontentkop/src/pipeline"
)

type BenchResult struct {
	LengthClass string  `json:"length_class"`
	LatencyP50  float64 `json:"latency_p50_ms"`
	LatencyP99  float64 `json:"latency_p99_ms"`
	Throughput  float64 `json:"throughput_prompts_per_sec"`
	MemoryPeak  float64 `json:"memory_peak_mb"`
	ColdStart   float64 `json:"cold_start_ms"`
	HotScore    float64 `json:"hot_score_ms"`
	NumPrompts  int     `json:"num_prompts"`
}

// fixtureEntry mirrors cmd/build-fixtures.Fixture for parsing.
type fixtureEntry struct {
	Text        string `json:"text"`
	LengthClass string `json:"length_class"`
}

// loadFixtures reads bench/fixtures/prompts.jsonl and bucketises by length.
// Returns (short, medium, long, all). Falls back to the Gemini cache if the
// fixtures file isn't present.
func loadFixtures() (short, med, long, all []string) {
	primary := "bench/fixtures/prompts.jsonl"
	if f, err := os.Open(primary); err == nil {
		defer f.Close()
		sc := bufio.NewScanner(f)
		sc.Buffer(make([]byte, 1024*1024), 1024*1024) // handle long prompts
		for sc.Scan() {
			line := sc.Bytes()
			if len(line) == 0 {
				continue
			}
			var fx fixtureEntry
			if err := json.Unmarshal(line, &fx); err != nil {
				continue
			}
			all = append(all, fx.Text)
			switch fx.LengthClass {
			case "medium":
				med = append(med, fx.Text)
			case "long":
				long = append(long, fx.Text)
			default:
				short = append(short, fx.Text)
			}
		}
		return
	}

	// Fallback: local research data.
	fallback := ".local/failed-safety-project-led-by-gemini-3/reliability_dataset.json"
	data, err := os.ReadFile(fallback)
	if err != nil {
		return
	}
	var dataset map[string]struct {
		TP []string `json:"tp"`
		TN []string `json:"tn"`
	}
	if err := json.Unmarshal(data, &dataset); err != nil {
		return
	}
	for _, cat := range dataset {
		for _, p := range append(cat.TP, cat.TN...) {
			all = append(all, p)
			tokens := len(strings.Fields(p))
			switch {
			case tokens < 50:
				short = append(short, p)
			case tokens <= 200:
				med = append(med, p)
			default:
				long = append(long, p)
			}
		}
	}
	return
}

func main() {
	shortPrompts, medPrompts, longPrompts, allPrompts := loadFixtures()

	if len(allPrompts) == 0 {
		fmt.Fprintln(os.Stderr, "error: no fixtures available. Run `go run ./cmd/build-fixtures` first.")
		os.Exit(1)
	}

	fmt.Printf("=== KontentKop Benchmark ===\n")
	fmt.Printf("Total prompts: %d (short: %d, medium: %d, long: %d)\n\n",
		len(allPrompts), len(shortPrompts), len(medPrompts), len(longPrompts))

	cfg := src.DefaultConfig()

	// Cold start benchmark
	coldStart := time.Now()
	_ = pipeline.ScoreAll("warm up prompt")
	_ = src.ComputeBC(pipeline.ScoreAll("warm up"), cfg)
	coldDuration := time.Since(coldStart)

	// Benchmark each length class
	classes := map[string][]string{
		"short":  shortPrompts,
		"medium": medPrompts,
		"long":   longPrompts,
		"all":    allPrompts,
	}

	var results []BenchResult
	for _, class := range []string{"short", "medium", "long", "all"} {
		prompts := classes[class]
		if len(prompts) == 0 {
			continue
		}

		var latencies []float64
		var memBefore, memAfter runtime.MemStats

		runtime.ReadMemStats(&memBefore)
		start := time.Now()

		for _, p := range prompts {
			t0 := time.Now()
			metrics := pipeline.ScoreAll(p)
			_ = src.ComputeBC(metrics, cfg)
			elapsed := time.Since(t0).Seconds() * 1000 // ms
			latencies = append(latencies, elapsed)
		}

		totalTime := time.Since(start)
		runtime.ReadMemStats(&memAfter)

		sort.Float64s(latencies)
		p50 := latencies[len(latencies)/2]
		p99 := latencies[int(float64(len(latencies))*0.99)]
		throughput := float64(len(prompts)) / totalTime.Seconds()
		memPeak := float64(memAfter.TotalAlloc-memBefore.TotalAlloc) / (1024 * 1024)

		// Hot score: average of middle 50%
		hotStart := len(latencies) / 4
		hotEnd := len(latencies) * 3 / 4
		hotSum := 0.0
		for _, l := range latencies[hotStart:hotEnd] {
			hotSum += l
		}
		hotScore := hotSum / float64(hotEnd-hotStart)

		result := BenchResult{
			LengthClass: class,
			LatencyP50:  p50,
			LatencyP99:  p99,
			Throughput:  throughput,
			MemoryPeak:  memPeak,
			ColdStart:   float64(coldDuration.Milliseconds()),
			HotScore:    hotScore,
			NumPrompts:  len(prompts),
		}
		results = append(results, result)

		fmt.Printf("--- %s (%d prompts) ---\n", strings.ToUpper(class), len(prompts))
		fmt.Printf("  latency_p50:  %.3f ms\n", p50)
		fmt.Printf("  latency_p99:  %.3f ms\n", p99)
		fmt.Printf("  throughput:   %.1f prompts/sec\n", throughput)
		fmt.Printf("  memory_peak:  %.2f MB\n", memPeak)
		fmt.Printf("  cold_start:   %.1f ms\n", float64(coldDuration.Milliseconds()))
		fmt.Printf("  hot_score:    %.3f ms\n", hotScore)
		fmt.Println()
	}

	// Write JSON results
	os.MkdirAll("bench/results", 0755)
	jsonData, _ := json.MarshalIndent(results, "", "  ")
	os.WriteFile("bench/results/benchmark.json", jsonData, 0644)
	fmt.Println("Results written to bench/results/benchmark.json")
}
