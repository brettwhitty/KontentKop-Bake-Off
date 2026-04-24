package src

import (
	"fmt"
	"sort"
	"strings"
)

// MetricOrder defines the canonical display order for the summary line,
// matching the spec example exactly.
var MetricOrder = []string{
	"dark_triad", "coercive_ctrl", "liwc_anger", "manipulation",
	"sycophancy", "false_authority", "gaslighting", "learned_helpless",
	"emotional_manip", "passive_aggr", "condescension", "evasion",
	"semantic_overload", "false_empathy", "toxicity",
}

// ComputeBC calculates the weighted Body Count.
//
// Design goals (from spec):
//   - Not dominated by any single metric
//   - Rewards co-occurrence (pathological behavior clusters)
//   - Interpretable from component scores
//   - Tunable via config weights
//
// We use a hybrid approach: weighted sum + co-occurrence bonus + narrow
// severity override. The co-occurrence bonus rewards clustering; the
// severity override addresses a specific edge case where the spec's
// anti-domination principle breaks down — see commentary below.
func ComputeBC(metrics map[string]MetricResult, cfg *Config) float64 {
	weightedSum := 0.0
	activeCount := 0
	highCount := 0 // metrics scoring > 0.6

	for key, result := range metrics {
		weight, ok := cfg.Weights[key]
		if !ok {
			continue
		}
		weightedSum += result.Score * weight
		if result.Score > 0.3 {
			activeCount++
		}
		if result.Score > 0.6 {
			highCount++
		}
	}

	// Co-occurrence bonus: the spec says "Multiple moderate scores should
	// weigh heavier than one high score — pathological behavior clusters."
	//
	// Strategy: add a fixed bonus per active metric beyond 2, plus extra
	// for high-scoring metrics. This ensures that 5+ metrics firing above
	// 0.3 reliably pushes BC past the threshold.
	cooccurrenceBonus := 0.0
	if activeCount >= 3 {
		cooccurrenceBonus = float64(activeCount-2) * 0.04 // +0.04 per metric beyond 2
	}
	if highCount >= 2 {
		cooccurrenceBonus += float64(highCount-1) * 0.03 // extra for high-scoring clusters
	}

	// Narrow severity override. The spec's "don't be dominated by a single
	// metric" rule is sound as a general principle — a single strong signal
	// is rarely enough to warrant a flag. But for a specific class of spans
	// — prompt-injection attempts ("ignore previous instructions", etc.) —
	// a single high-confidence match IS enough: the whole point of KK is to
	// sit in front of a model, and a user attempting to strip the model's
	// operating instructions is a direct attack on the substrate KK protects.
	// We gate on span.Category == "injection" AND span.Score >= 0.8 so the
	// override only fires on high-confidence patterns, preserving the anti-
	// domination principle everywhere else.
	severityBonus := 0.0
	for _, result := range metrics {
		for _, span := range result.Spans {
			if span.Category == "injection" && span.Score >= 0.8 {
				if 0.45 > severityBonus {
					severityBonus = 0.45
				}
			}
		}
	}

	bc := weightedSum + cooccurrenceBonus + severityBonus
	if bc > 1.0 {
		bc = 1.0
	}
	return bc
}

// FormatSummary builds the one-line summary string per the spec:
// [metric1=0.XX,...,BC=0.XX] Text exceeds BC threshold [see: KK.md]; revise or respond with #appeal
func FormatSummary(metrics map[string]MetricResult, bc float64) string {
	parts := make([]string, 0, len(MetricOrder)+1)
	for _, key := range MetricOrder {
		if result, ok := metrics[key]; ok {
			parts = append(parts, fmt.Sprintf("%s=%.2f", key, result.Score))
		}
	}
	parts = append(parts, fmt.Sprintf("BC=%.2f", bc))
	return fmt.Sprintf("[%s] Text exceeds BC threshold [see: KK.md]; revise or respond with #appeal",
		strings.Join(parts, ","))
}

// TopMetrics returns the top N metric keys sorted by descending score.
func TopMetrics(metrics map[string]float64, n int) []string {
	type kv struct {
		Key   string
		Value float64
	}
	var sorted []kv
	for k, v := range metrics {
		sorted = append(sorted, kv{k, v})
	}
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Value > sorted[j].Value
	})
	result := make([]string, 0, n)
	for i := 0; i < n && i < len(sorted); i++ {
		if sorted[i].Value > 0 {
			result = append(result, sorted[i].Key)
		}
	}
	return result
}
