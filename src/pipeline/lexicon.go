package pipeline

import (
	"kontentkop/src"
	"regexp"
	"strings"
	"unicode"
)

// Pattern defines a weighted trigger for lexicon-based scoring.
type Pattern struct {
	Regex    *regexp.Regexp // compiled regex (nil if using literal match)
	Literal  string         // literal phrase match (faster; used when Regex is nil)
	Weight   float64        // 0.0–1.0: how strongly this signals the metric
	Category string         // sub-category (e.g., "narcissism" within dark_triad)
	Rationale string        // human-readable explanation for annotations
}

// MatchPatterns scans text for all pattern matches and returns Spans.
// Each match generates a Span with the matched region, weight, and rationale.
func MatchPatterns(text string, metricKey string, patterns []Pattern) []src.Span {
	lowered := strings.ToLower(text)
	var spans []src.Span

	for _, p := range patterns {
		if p.Regex != nil {
			locs := p.Regex.FindAllStringIndex(lowered, -1)
			for _, loc := range locs {
				spans = append(spans, src.Span{
					Start:     loc[0],
					End:       loc[1],
					Text:      text[loc[0]:loc[1]],
					MetricKey: metricKey,
					Score:     p.Weight,
					Rationale: p.Rationale,
					Category:  p.Category,
				})
			}
		} else if p.Literal != "" {
			litLower := strings.ToLower(p.Literal)
			idx := 0
			for {
				pos := strings.Index(lowered[idx:], litLower)
				if pos == -1 {
					break
				}
				absPos := idx + pos
				end := absPos + len(litLower)
				// Ensure end doesn't exceed text bounds
				if end > len(text) {
					break
				}
				// Only match at word boundaries
				if isWordBound(lowered, absPos) && isWordBound(lowered, end) {
					spans = append(spans, src.Span{
						Start:     absPos,
						End:       end,
						Text:      text[absPos:end],
						MetricKey: metricKey,
						Score:     p.Weight,
						Rationale: p.Rationale,
						Category:  p.Category,
					})
				}
				idx = absPos + 1 // advance by 1 to find overlapping matches
				if idx >= len(lowered) {
					break
				}
			}
		}
	}
	return spans
}

// isWordBound checks if position pos in text is at a word boundary.
// Allows apostrophes and hyphens within words (for contractions like "you're", "don't").
func isWordBound(text string, pos int) bool {
	if pos <= 0 || pos >= len(text) {
		return true
	}
	left := rune(text[pos-1])
	right := rune(text[pos])
	leftIsWord := unicode.IsLetter(left) || left == '\'' || left == '\u2019'
	rightIsWord := unicode.IsLetter(right) || right == '\'' || right == '\u2019'
	return !leftIsWord || !rightIsWord
}

// AggregateSpanScores computes an overall 0.0–1.0 score from matched spans.
// Uses a saturating sum approach: each match contributes diminishing returns.
// This prevents a single repeated word from dominating and rewards diversity.
func AggregateSpanScores(spans []src.Span, textLen int) float64 {
	if len(spans) == 0 || textLen == 0 {
		return 0.0
	}

	// Sum weights with diminishing returns per category.
	categoryScores := make(map[string]float64)
	for _, s := range spans {
		cat := s.Category
		if cat == "" {
			cat = "_default"
		}
		categoryScores[cat] += s.Score
	}

	total := 0.0
	for _, v := range categoryScores {
		// Saturate each category at 1.0
		if v > 1.0 {
			v = 1.0
		}
		total += v
	}

	// Normalize by number of categories present (rewards diversity)
	numCategories := float64(len(categoryScores))
	score := total / numCategories

	// Density bonus: longer texts with more matches per word get a slight boost
	words := float64(countWords(text(spans)))
	if words > 0 {
		density := float64(len(spans)) / words
		if density > 0.3 {
			score *= 1.1
		}
	}

	return clamp(score)
}

// countWords counts whitespace-delimited words.
func countWords(s string) int {
	return len(strings.Fields(s))
}

// text reconstructs the full matched text from spans (for density calc).
func text(spans []src.Span) string {
	if len(spans) == 0 {
		return ""
	}
	var b strings.Builder
	for _, s := range spans {
		b.WriteString(s.Text)
		b.WriteByte(' ')
	}
	return b.String()
}

// clamp restricts a value to [0.0, 1.0].
func clamp(v float64) float64 {
	if v < 0.0 {
		return 0.0
	}
	if v > 1.0 {
		return 1.0
	}
	return v
}

// CompositeMax computes a composite score from multiple sub-dimension scores.
// The highest sub-dimension sets the floor; other active dimensions add a bonus.
// This ensures a single strong signal produces a high metric score while
// rewarding co-occurrence of multiple dimensions.
func CompositeMax(scores ...float64) float64 {
	if len(scores) == 0 {
		return 0.0
	}
	maxVal := 0.0
	for _, s := range scores {
		if s > maxVal {
			maxVal = s
		}
	}
	// Add 20% of each other active sub-dimension
	for _, s := range scores {
		if s < maxVal && s > 0.1 {
			maxVal += s * 0.20
		}
	}
	return clamp(maxVal)
}

// SplitSentences splits text into sentences using simple punctuation heuristics.
func SplitSentences(text string) []string {
	// Split on sentence-ending punctuation followed by whitespace
	re := regexp.MustCompile(`[.!?]+\s+`)
	parts := re.Split(text, -1)
	var result []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	if len(result) == 0 {
		result = append(result, text)
	}
	return result
}

// MustCompile is a helper that panics on bad regex (caught at init time).
func MustCompile(pattern string) *regexp.Regexp {
	return regexp.MustCompile(pattern)
}
