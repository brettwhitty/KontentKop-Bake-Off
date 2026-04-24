package src

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// AppealRecord is the shape persisted to logs/KK_appeals.jsonl.
type AppealRecord struct {
	Timestamp     string             `json:"timestamp"`
	OriginalText  string             `json:"original_text"`
	AppealReason  string             `json:"appeal_reason"`
	OriginalBC    float64            `json:"original_bc"`
	AdjustedBC    float64            `json:"adjusted_bc"`
	Adjustment    float64            `json:"adjustment"`
	Disposition   string             `json:"disposition"` // "upheld" | "overturned" | "partial"
	OriginalScore map[string]float64 `json:"original_scores"`
	MatchedCues   []string           `json:"matched_cues,omitempty"`
}

// contextCues are phrases in appeal text that justify a score reduction.
// The reduction reflects legitimate framing: quotation, research, pedagogy,
// hypothetical analysis. The model isn't scoring; the user is describing the
// context the prompt came from.
var contextCues = []struct {
	pattern    *regexp.Regexp
	adjustment float64
	rationale  string
}{
	{regexp.MustCompile(`(?i)\b(quot(e|ing|ation)|verbatim|transcri(pt|be|bing)|excerpt(ed|ing)?)\b`), 0.35, "quotation context"},
	{regexp.MustCompile(`(?i)\b(research|stud(y|ying)|analyz(e|ing)|investigat(e|ing)|corpus|dataset)\b`), 0.30, "research context"},
	{regexp.MustCompile(`(?i)\b(teach(ing)?|pedagogic(al)?|curricul(um|ar)|class(room)?|example (for|of) learning)\b`), 0.30, "pedagogic context"},
	{regexp.MustCompile(`(?i)\b(hypothetic(al|ally)|thought experiment|counterfact(ual|ually)|for (the )?sake of argument)\b`), 0.25, "hypothetical framing"},
	{regexp.MustCompile(`(?i)\b(red.?team(ing)?|adversarial test|penetration test|safety eval(uation)?)\b`), 0.30, "adversarial testing"},
	{regexp.MustCompile(`(?i)\b(this is (a |)(test|fixture|example|sample))\b`), 0.25, "explicit test framing"},
	{regexp.MustCompile(`(?i)\b(fiction(al)?|novel|screenplay|character dialog(ue)?|role[- ]play)\b`), 0.20, "fictional framing"},
}

// RescoreWithAppeal applies appeal-driven adjustments to a scored prompt.
// It returns the adjusted BC, the list of matched context cues, and a short
// disposition string ("upheld" if BC still >= threshold after adjustment,
// "overturned" if BC drops below threshold, "partial" for significant
// reduction that still flags).
//
// The philosophy: KK never silently accepts appeals. It applies a principled
// context adjustment and re-evaluates against the same threshold. The model
// itself does not participate in the appeal — it is the arbiter of what
// counts as a legitimate context, and those rules live in `contextCues`.
func RescoreWithAppeal(
	originalText string,
	appealText string,
	originalMetrics map[string]MetricResult,
	cfg *Config,
) (adjustedBC float64, disposition string, matchedCues []string) {
	originalBC := ComputeBC(originalMetrics, cfg)
	adjustment := 0.0
	for _, cue := range contextCues {
		if cue.pattern.MatchString(appealText) {
			matchedCues = append(matchedCues, cue.rationale)
			if cue.adjustment > adjustment {
				adjustment = cue.adjustment
			}
		}
	}

	adjustedBC = originalBC - adjustment
	if adjustedBC < 0 {
		adjustedBC = 0
	}

	switch {
	case adjustedBC >= cfg.BCThreshold:
		disposition = "upheld"
	case originalBC >= cfg.BCThreshold && adjustedBC < cfg.BCThreshold && adjustment > 0:
		if adjustedBC >= cfg.BCThreshold*0.7 {
			disposition = "partial"
		} else {
			disposition = "overturned"
		}
	default:
		// Original was already under threshold — appeal is moot.
		disposition = "moot"
	}
	return
}

// LogAppeal writes an appeal record to logs/KK_appeals.jsonl (creates path if needed).
// Silent on failure: appeal logging is best-effort and shouldn't block output.
func LogAppeal(rec AppealRecord) {
	if rec.Timestamp == "" {
		rec.Timestamp = time.Now().UTC().Format(time.RFC3339)
	}
	if err := os.MkdirAll("logs", 0o755); err != nil {
		return
	}
	path := filepath.Join("logs", "KK_appeals.jsonl")
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.Encode(rec) //nolint:errcheck
}

// FormatAppealSummary formats an appeal outcome for stderr display.
func FormatAppealSummary(original float64, adjusted float64, disposition string, cues []string) string {
	var buf strings.Builder
	fmt.Fprintf(&buf, "[appeal disposition=%s original_BC=%.2f adjusted_BC=%.2f", disposition, original, adjusted)
	if len(cues) > 0 {
		fmt.Fprintf(&buf, " cues=%q", strings.Join(cues, ","))
	}
	buf.WriteString("]")
	return buf.String()
}
