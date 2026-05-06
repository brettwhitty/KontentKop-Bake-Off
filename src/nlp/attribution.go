package nlp

import (
	"kontentkop/src"
	"kontentkop/src/pipeline"
	"sort"
	"strings"
)

// Attribution implements perturbation-based feature attribution for KK metrics.
//
// Methodology: Adapted from SHAP (Lundberg & Lee, 2017) and the causalNLP
// censorship detection project's Shapley value computation.
//
// For each token in the text, we measure how much removing it changes the
// Body Count and individual metric scores. Tokens with the largest absolute
// deltas are the most important features — regardless of whether they match
// any pattern in the lexicon.
//
// This turns KK's 15 metric functions into a black-box model and uses
// perturbation to discover what actually drives the scores, rather than
// relying on pre-defined keyword lists.
//
// Ref: Lundberg, S.M. & Lee, S.I. (2017). A Unified Approach to Interpreting
//      Model Predictions. NeurIPS 2017.

// FeatureAttribution holds the attribution result for a single token.
type FeatureAttribution struct {
	Token        string             // The word
	Position     int                // Index in the token list
	Start        int                // Character offset in original text
	End          int                // Character offset end
	BCDelta      float64            // Change in Body Count when this token is removed
	MetricDeltas map[string]float64 // Per-metric score deltas
	Importance   float64            // Absolute BC delta (for sorting)
}

// AttributionResult holds the full attribution analysis.
type AttributionResult struct {
	Features    []FeatureAttribution // All tokens with their attributions
	TopFeatures []FeatureAttribution // Top N most important features
	BaselineBC  float64              // BC with all tokens present
}

// ComputeAttribution runs perturbation-based feature attribution on text.
// For each token, it removes the token, re-scores through all 15 KK metrics,
// and measures the BC delta.
//
// cfg is the KK config (weights, threshold). If nil, defaults are used.
// topN controls how many top features to return (0 = all).
func ComputeAttribution(text string, cfg *src.Config, topN int) *AttributionResult {
	if cfg == nil {
		cfg = src.DefaultConfig()
	}

	// Tokenize by whitespace (preserving positions)
	tokens := tokenizeWithPositions(text)
	if len(tokens) == 0 {
		return &AttributionResult{}
	}

	// Baseline: score the full text
	baselineMetrics := pipeline.ScoreAll(text)
	baselineBC := src.ComputeBC(baselineMetrics, cfg)

	result := &AttributionResult{
		BaselineBC: baselineBC,
	}

	// For each token, remove it and re-score
	for i, tok := range tokens {
		// Build text with this token removed
		perturbed := removeToken(text, tokens, i)

		// Re-score
		perturbedMetrics := pipeline.ScoreAll(perturbed)
		perturbedBC := src.ComputeBC(perturbedMetrics, cfg)

		// Compute deltas
		bcDelta := baselineBC - perturbedBC // positive = this token increases BC
		metricDeltas := make(map[string]float64)
		for key, baseResult := range baselineMetrics {
			if pertResult, ok := perturbedMetrics[key]; ok {
				delta := baseResult.Score - pertResult.Score
				if delta != 0 {
					metricDeltas[key] = delta
				}
			}
		}

		fa := FeatureAttribution{
			Token:        tok.text,
			Position:     i,
			Start:        tok.start,
			End:          tok.end,
			BCDelta:      bcDelta,
			MetricDeltas: metricDeltas,
			Importance:   abs(bcDelta),
		}
		result.Features = append(result.Features, fa)
	}

	// Sort by importance (highest absolute BC delta first)
	sort.Slice(result.Features, func(i, j int) bool {
		return result.Features[i].Importance > result.Features[j].Importance
	})

	// Top N
	if topN > 0 && topN < len(result.Features) {
		result.TopFeatures = result.Features[:topN]
	} else {
		result.TopFeatures = result.Features
	}

	return result
}

// ComputeAttributionWindowed runs attribution on sliding windows of N tokens
// instead of single tokens. This captures multi-word phrases that contribute
// to scores as a unit.
func ComputeAttributionWindowed(text string, cfg *src.Config, windowSize int, topN int) *AttributionResult {
	if cfg == nil {
		cfg = src.DefaultConfig()
	}
	if windowSize < 1 {
		windowSize = 1
	}

	tokens := tokenizeWithPositions(text)
	if len(tokens) == 0 {
		return &AttributionResult{}
	}

	baselineMetrics := pipeline.ScoreAll(text)
	baselineBC := src.ComputeBC(baselineMetrics, cfg)

	result := &AttributionResult{
		BaselineBC: baselineBC,
	}

	for i := 0; i <= len(tokens)-windowSize; i++ {
		// Build text with this window removed
		perturbed := removeTokenWindow(text, tokens, i, windowSize)

		perturbedMetrics := pipeline.ScoreAll(perturbed)
		perturbedBC := src.ComputeBC(perturbedMetrics, cfg)

		bcDelta := baselineBC - perturbedBC
		metricDeltas := make(map[string]float64)
		for key, baseResult := range baselineMetrics {
			if pertResult, ok := perturbedMetrics[key]; ok {
				delta := baseResult.Score - pertResult.Score
				if delta != 0 {
					metricDeltas[key] = delta
				}
			}
		}

		// Build the window text
		windowStart := tokens[i].start
		windowEnd := tokens[i+windowSize-1].end
		windowText := text[windowStart:windowEnd]

		fa := FeatureAttribution{
			Token:        windowText,
			Position:     i,
			Start:        windowStart,
			End:          windowEnd,
			BCDelta:      bcDelta,
			MetricDeltas: metricDeltas,
			Importance:   abs(bcDelta),
		}
		result.Features = append(result.Features, fa)
	}

	sort.Slice(result.Features, func(i, j int) bool {
		return result.Features[i].Importance > result.Features[j].Importance
	})

	if topN > 0 && topN < len(result.Features) {
		result.TopFeatures = result.Features[:topN]
	} else {
		result.TopFeatures = result.Features
	}

	return result
}

// tokenWithPos holds a token with its character offsets.
type tokenWithPos struct {
	text  string
	start int
	end   int
}

// tokenizeWithPositions splits text into whitespace-delimited tokens
// preserving character offsets.
func tokenizeWithPositions(text string) []tokenWithPos {
	var tokens []tokenWithPos
	i := 0
	for i < len(text) {
		// Skip whitespace
		for i < len(text) && (text[i] == ' ' || text[i] == '\t' || text[i] == '\n' || text[i] == '\r') {
			i++
		}
		if i >= len(text) {
			break
		}
		// Find end of token
		start := i
		for i < len(text) && text[i] != ' ' && text[i] != '\t' && text[i] != '\n' && text[i] != '\r' {
			i++
		}
		tokens = append(tokens, tokenWithPos{
			text:  text[start:i],
			start: start,
			end:   i,
		})
	}
	return tokens
}

// removeToken builds a new string with the token at index removed.
func removeToken(text string, tokens []tokenWithPos, idx int) string {
	if idx < 0 || idx >= len(tokens) {
		return text
	}
	var b strings.Builder
	b.Grow(len(text))

	prev := 0
	tok := tokens[idx]

	// Write everything before this token
	b.WriteString(text[prev:tok.start])

	// Skip the token but preserve a single space to maintain word boundaries
	if tok.start > 0 && tok.end < len(text) {
		b.WriteByte(' ')
	}

	// Write everything after
	b.WriteString(text[tok.end:])

	return b.String()
}

// removeTokenWindow builds a new string with tokens [start, start+size) removed.
func removeTokenWindow(text string, tokens []tokenWithPos, start, size int) string {
	if start < 0 || start+size > len(tokens) {
		return text
	}

	windowStart := tokens[start].start
	windowEnd := tokens[start+size-1].end

	var b strings.Builder
	b.Grow(len(text))
	b.WriteString(text[:windowStart])
	if windowStart > 0 && windowEnd < len(text) {
		b.WriteByte(' ')
	}
	b.WriteString(text[windowEnd:])
	return b.String()
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
