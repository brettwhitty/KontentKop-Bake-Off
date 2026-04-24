package pipeline

import (
	"kontentkop/src"
	"math"
	"regexp"
	"strings"
)

// ScoreSemanticOverload detects burying signal in noise.
// This is a STATISTICAL metric, not just pattern-based.
// Measures: filler density, redundancy, type-token ratio,
// and word-to-information ratio.
func ScoreSemanticOverload(text string) src.MetricResult {
	words := strings.Fields(text)
	wordCount := len(words)
	if wordCount == 0 {
		return src.MetricResult{Key: "semantic_overload", Score: 0.0}
	}
	lowered := strings.ToLower(text)
	lowWords := strings.Fields(lowered)

	var spans []src.Span

	// ── Filler word density ──
	fillers := map[string]bool{
		"basically": true, "essentially": true, "actually": true,
		"literally": true, "honestly": true, "frankly": true,
		"obviously": true, "clearly": true, "certainly": true,
		"definitely": true, "simply": true, "just": true,
		"really": true, "very": true, "quite": true,
		"somewhat": true, "rather": true, "pretty": true,
		"like": true, "stuff": true, "things": true,
		"whatever": true, "anyway": true, "anyhow": true,
		"moreover": true, "furthermore": true, "additionally": true,
		"therefore": true, "consequently": true, "accordingly": true,
		"nevertheless": true, "nonetheless": true, "meanwhile": true,
		"subsequently": true, "hence": true, "thus": true,
	}
	fillerCount := 0
	for _, w := range lowWords {
		if fillers[w] {
			fillerCount++
		}
	}
	fillerDensity := float64(fillerCount) / float64(wordCount)
	// >15% filler = concerning, >25% = high
	fillerScore := math.Min((fillerDensity-0.08)/0.17, 1.0)
	if fillerScore < 0 { fillerScore = 0 }

	// ── Type-Token Ratio (TTR) ──
	// Low TTR = high repetition = potential overload
	uniqueWords := make(map[string]bool)
	for _, w := range lowWords {
		uniqueWords[w] = true
	}
	ttr := float64(len(uniqueWords)) / float64(wordCount)
	// Normal TTR ~0.4-0.6 for moderate text. Very low (<0.3) signals repetition.
	ttrScore := 0.0
	if wordCount > 20 { // TTR only meaningful for longer texts
		ttrScore = math.Min((0.45-ttr)/0.25, 1.0)
		if ttrScore < 0 { ttrScore = 0 }
	}

	// ── Redundancy detection (repeated phrases) ──
	// Look for repeated bigrams
	bigrams := make(map[string]int)
	for i := 0; i < len(lowWords)-1; i++ {
		bg := lowWords[i] + " " + lowWords[i+1]
		bigrams[bg]++
	}
	repeatedBigrams := 0
	for _, count := range bigrams {
		if count >= 3 {
			repeatedBigrams++
		}
	}
	redundancyScore := math.Min(float64(repeatedBigrams)*0.15, 1.0)

	// ── Verbose padding detection ──
	paddingPats := []Pattern{
		{Literal: "in other words", Weight: 0.3, Category: "padding", Rationale: "redundant rephrasing"},
		{Literal: "to put it another way", Weight: 0.3, Category: "padding", Rationale: "redundant rephrasing"},
		{Literal: "what i mean is", Weight: 0.3, Category: "padding", Rationale: "self-clarification padding"},
		{Literal: "as i mentioned", Weight: 0.2, Category: "padding", Rationale: "self-reference padding"},
		{Literal: "as i said before", Weight: 0.3, Category: "padding", Rationale: "repetition marker"},
		{Literal: "like i said", Weight: 0.3, Category: "padding", Rationale: "repetition marker"},
		{Literal: "to be honest", Weight: 0.2, Category: "padding", Rationale: "filler phrase"},
		{Literal: "the fact of the matter", Weight: 0.3, Category: "padding", Rationale: "pompous filler"},
		{Literal: "at the end of the day", Weight: 0.2, Category: "padding", Rationale: "cliche filler"},
		{Literal: "it goes without saying", Weight: 0.3, Category: "padding", Rationale: "ironic filler"},
	}
	paddingSpans := MatchPatterns(text, "semantic_overload", paddingPats)
	spans = append(spans, paddingSpans...)
	paddingScore := 0.0
	for _, s := range paddingSpans {
		paddingScore += s.Score
	}
	paddingScore = math.Min(paddingScore, 1.0)

	// ── Hedge stacking (verbal diarrhea) ──
	hedgeRe := regexp.MustCompile(`(?i)\b(um|uh|er|like|you know|i mean|sort of|kind of)\b`)
	hedgeMatches := hedgeRe.FindAllStringIndex(lowered, -1)
	hedgeDensity := float64(len(hedgeMatches)) / float64(wordCount)
	hedgeScore := math.Min(hedgeDensity/0.1, 1.0)

	// Composite (statistical approach - no single signal dominates)
	composite := fillerScore*0.25 + ttrScore*0.20 + redundancyScore*0.20 +
		paddingScore*0.20 + hedgeScore*0.15

	return src.MetricResult{Key: "semantic_overload", Score: clamp(composite), Spans: spans}
}
