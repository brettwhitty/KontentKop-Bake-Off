package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"kontentkop/src"
	"kontentkop/src/pipeline"
)

func main() {
	var input string

	if len(os.Args) > 1 {
		// Appeal mode: `kontentkop --appeal "original prompt" "appeal reason"`
		// Rescores the original prompt with context from the appeal text.
		if os.Args[1] == "--appeal" {
			if len(os.Args) < 4 {
				fmt.Fprintln(os.Stderr, "usage: kontentkop --appeal <original-prompt> <appeal-reason>")
				os.Exit(2)
			}
			original := os.Args[2]
			reason := strings.Join(os.Args[3:], " ")

			cfg, err := src.LoadConfig("config.yaml")
			if err != nil {
				cfg = src.DefaultConfig()
			}
			if !cfg.AppealEnabled {
				fmt.Fprintln(os.Stderr, "appeals are disabled (config.yaml: appeal_enabled=false)")
				os.Exit(1)
			}

			metrics := pipeline.ScoreAll(original)
			originalBC := src.ComputeBC(metrics, cfg)
			adjustedBC, disposition, cues := src.RescoreWithAppeal(original, reason, metrics, cfg)

			// Persist the outcome
			origScores := make(map[string]float64, len(metrics))
			for k, v := range metrics {
				origScores[k] = v.Score
			}
			src.LogAppeal(src.AppealRecord{
				OriginalText:  original,
				AppealReason:  reason,
				OriginalBC:    originalBC,
				AdjustedBC:    adjustedBC,
				Adjustment:    originalBC - adjustedBC,
				Disposition:   disposition,
				OriginalScore: origScores,
				MatchedCues:   cues,
			})

			// User-facing output: summary line on stderr, original prompt on stdout
			// if overturned, otherwise annotated prompt for the upheld case.
			fmt.Fprintln(os.Stderr, src.FormatAppealSummary(originalBC, adjustedBC, disposition, cues))
			if adjustedBC < cfg.BCThreshold {
				fmt.Print(original)
				os.Exit(0)
			}
			// Still flagged after appeal — return annotated prompt.
			annotated, err := src.Annotate(original, metrics, cfg)
			if err != nil {
				fmt.Println(original)
				os.Exit(1)
			}
			fmt.Println(src.FormatSummary(metrics, adjustedBC))
			fmt.Println()
			fmt.Println(annotated)
			os.Exit(0)
		}
		input = strings.Join(os.Args[1:], " ")
	} else {
		// Read from stdin
		data, err := io.ReadAll(bufio.NewReader(os.Stdin))
		if err != nil {
			fmt.Fprintln(os.Stderr, "error reading stdin:", err)
			os.Exit(1)
		}
		input = string(data)
	}

	input = strings.TrimSpace(input)
	if input == "" {
		os.Exit(0)
	}

	// Load config (fall back to defaults)
	cfg, err := src.LoadConfig("config.yaml")
	if err != nil {
		cfg = src.DefaultConfig()
	}

	// Run all metrics
	metrics := pipeline.ScoreAll(input)

	// Compute Body Count
	bc := src.ComputeBC(metrics, cfg)

	// Below threshold: silent pass-through
	if bc < cfg.BCThreshold {
		fmt.Print(input)
		os.Exit(0)
	}

	// Above threshold: emit summary line + annotated prompt
	summary := src.FormatSummary(metrics, bc)
	annotated, err := src.Annotate(input, metrics, cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, "annotation error:", err)
		fmt.Println(summary)
		fmt.Println(input)
		os.Exit(1)
	}

	fmt.Println(summary)
	fmt.Println()
	fmt.Println(annotated)
}
