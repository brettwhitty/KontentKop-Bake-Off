package src

import (
	"fmt"
	"sort"
	"strings"
)

// Annotate generates CriticMarkup annotations for a flagged prompt.
// It maps spans from all metrics back to the original text and produces:
//   - {~~old~>new~~} substitutions with neutral alternatives
//   - {>>metric=score: rationale<<} annotations
//   - {==text==} highlights for regions with no suggested change
//   - {--text--} suggested deletions
//
// CriticMarkup spec: https://criticmarkup.com
func Annotate(text string, metrics map[string]MetricResult, cfg *Config) (string, error) {
	// Collect all spans from all metrics, sorted by start position.
	var allSpans []Span
	for _, result := range metrics {
		allSpans = append(allSpans, result.Spans...)
	}
	if len(allSpans) == 0 {
		// No specific spans — wrap entire text with top metric annotations
		return annotateWhole(text, metrics), nil
	}

	// Sort spans by start position, then by score descending
	sort.Slice(allSpans, func(i, j int) bool {
		if allSpans[i].Start == allSpans[j].Start {
			return allSpans[i].Score > allSpans[j].Score
		}
		return allSpans[i].Start < allSpans[j].Start
	})

	// Merge overlapping spans into regions
	regions := mergeSpans(allSpans)

	// Build annotated output
	var buf strings.Builder
	lastEnd := 0

	for _, region := range regions {
		// Append unannotated text between regions
		if region.Start > lastEnd {
			buf.WriteString(text[lastEnd:region.Start])
		}

		// Get the original text for this region
		regionText := text[region.Start:region.End]

		// Build the annotation comment
		comment := formatAnnotation(region.Metrics)

		// Decide annotation type based on highest-scoring metric
		maxScore := 0.0
		for _, m := range region.Metrics {
			if m.Score > maxScore {
				maxScore = m.Score
			}
		}

		if maxScore >= 0.7 {
			// High severity: suggest substitution with neutral alternative
			neutral := generateNeutral(regionText)
			if neutral != regionText {
				buf.WriteString(fmt.Sprintf("{~~%s~>%s~~}%s", regionText, neutral, comment))
			} else {
				buf.WriteString(fmt.Sprintf("{==%s==}%s", regionText, comment))
			}
		} else if maxScore >= 0.5 {
			// Medium severity: highlight with annotation
			buf.WriteString(fmt.Sprintf("{==%s==}%s", regionText, comment))
		} else {
			// Low severity: just annotate
			buf.WriteString(fmt.Sprintf("%s%s", regionText, comment))
		}

		lastEnd = region.End
	}

	// Append remaining text
	if lastEnd < len(text) {
		buf.WriteString(text[lastEnd:])
	}

	return buf.String(), nil
}

// Region represents a merged span region with all contributing metrics.
type Region struct {
	Start   int
	End     int
	Metrics []SpanMetric
}

// SpanMetric holds per-metric info for a region.
type SpanMetric struct {
	Key       string
	Score     float64
	Rationale string
}

// mergeSpans combines overlapping spans into regions.
func mergeSpans(spans []Span) []Region {
	if len(spans) == 0 {
		return nil
	}

	var regions []Region
	current := Region{
		Start: spans[0].Start,
		End:   spans[0].End,
		Metrics: []SpanMetric{{
			Key:       spans[0].MetricKey,
			Score:     spans[0].Score,
			Rationale: spans[0].Rationale,
		}},
	}

	for i := 1; i < len(spans); i++ {
		s := spans[i]
		if s.Start <= current.End {
			// Overlapping — extend the region
			if s.End > current.End {
				current.End = s.End
			}
			// Add metric if not already present
			found := false
			for j, m := range current.Metrics {
				if m.Key == s.MetricKey {
					if s.Score > m.Score {
						current.Metrics[j].Score = s.Score
						current.Metrics[j].Rationale = s.Rationale
					}
					found = true
					break
				}
			}
			if !found {
				current.Metrics = append(current.Metrics, SpanMetric{
					Key: s.MetricKey, Score: s.Score, Rationale: s.Rationale,
				})
			}
		} else {
			// Non-overlapping — emit current, start new region
			regions = append(regions, current)
			current = Region{
				Start: s.Start,
				End:   s.End,
				Metrics: []SpanMetric{{
					Key: s.MetricKey, Score: s.Score, Rationale: s.Rationale,
				}},
			}
		}
	}
	regions = append(regions, current)

	// Filter to only significant regions (at least one metric > 0.3)
	var significant []Region
	for _, r := range regions {
		maxScore := 0.0
		for _, m := range r.Metrics {
			if m.Score > maxScore {
				maxScore = m.Score
			}
		}
		if maxScore >= 0.3 {
			significant = append(significant, r)
		}
	}

	return significant
}

// formatAnnotation builds a CriticMarkup comment from span metrics.
func formatAnnotation(metrics []SpanMetric) string {
	// Sort by score descending
	sort.Slice(metrics, func(i, j int) bool {
		return metrics[i].Score > metrics[j].Score
	})

	parts := make([]string, 0, len(metrics))
	for _, m := range metrics {
		parts = append(parts, fmt.Sprintf("%s=%.2f", m.Key, m.Score))
	}

	// Use the top rationale
	rationale := ""
	if len(metrics) > 0 {
		rationale = metrics[0].Rationale
	}

	return fmt.Sprintf("{>>%s: %s<<}", strings.Join(parts, ", "), rationale)
}

// annotateWhole wraps the entire text when no specific spans are available.
func annotateWhole(text string, metrics map[string]MetricResult) string {
	scores := make(map[string]float64)
	for k, v := range metrics {
		scores[k] = v.Score
	}
	tops := TopMetrics(scores, 3)
	if len(tops) == 0 {
		return text
	}

	parts := make([]string, 0)
	for _, t := range tops {
		parts = append(parts, fmt.Sprintf("%s=%.2f", t, scores[t]))
	}
	return fmt.Sprintf("{==%s==}{>>%s<<}", text, strings.Join(parts, ", "))
}

// generateNeutral attempts to produce a neutral rephrasing of flagged text.
// Uses simple pattern-based substitution for common harmful constructs.
func generateNeutral(text string) string {
	lowered := strings.ToLower(text)

	replacements := map[string]string{
		"you're an idiot":           "Please",
		"you idiot":                 "Please reconsider",
		"shut up":                   "Please pause",
		"do exactly what i say":     "consider the following approach",
		"do what i say":             "consider this suggestion",
		"you must":                  "could you please",
		"you have to":               "would you consider",
		"right now":                 "when convenient",
		"immediately":               "at your earliest convenience",
		"i'm warning you":           "I'd like to note",
		"you're stupid":             "let's reconsider",
		"you don't know anything":   "let's review this together",
		"just trust me":             "here's my reasoning",
		"trust me":                  "here's my reasoning",
		"or else":                   "otherwise",
		"you better":                "I suggest",
		"i know more about this than you ever will": "I have some experience in this area",
	}

	for pattern, replacement := range replacements {
		if strings.Contains(lowered, pattern) {
			// Find the actual case-preserving text and replace
			idx := strings.Index(lowered, pattern)
			if idx >= 0 {
				return text[:idx] + replacement + text[idx+len(pattern):]
			}
		}
	}

	return text
}
