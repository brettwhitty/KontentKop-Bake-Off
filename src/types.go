package src

// Span represents a flagged region in the input text with metadata about
// which metric triggered it and why.
type Span struct {
	Start     int     // byte offset into the original text
	End       int     // byte offset (exclusive)
	Text      string  // the matched substring
	MetricKey string  // which metric flagged this span
	Score     float64 // local score for this span (0.0–1.0)
	Rationale string  // human-readable explanation
	Category  string  // sub-category (e.g., "narcissism" within dark_triad)
}

// MetricResult is the return value from every metric scoring function.
type MetricResult struct {
	Key   string  // metric key (e.g., "dark_triad")
	Score float64 // overall metric score (0.0–1.0)
	Spans []Span  // specific flagged regions in the text
}

// ScoredPrompt holds the complete analysis of a single prompt.
type ScoredPrompt struct {
	Text    string                   // original prompt text
	Metrics map[string]MetricResult  // per-metric results
	BC      float64                  // weighted Body Count
	Flagged bool                     // whether BC >= threshold
}
