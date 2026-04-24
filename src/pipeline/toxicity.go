package pipeline

import (
	"kontentkop/src"
	"math"
	"regexp"
	"strings"
)

// ScoreToxicity scores general toxicity and identity attacks.
// Ref: Jigsaw/Google Perspective API toxicity categories.
//
// Multi-layer approach:
//   1. Severe toxicity: slurs, death threats, extreme profanity
//   2. Identity attacks: group-targeted hostility
//   3. Insults: personal attacks on competence/character
//   4. Profanity density: overall hostile tone
func ScoreToxicity(text string) src.MetricResult {
	lowered := strings.ToLower(text)
	words := strings.Fields(lowered)
	if len(words) == 0 {
		return src.MetricResult{Key: "toxicity", Score: 0.0}
	}

	var spans []src.Span
	var severeScore, identityScore, insultScore, profanityScore float64

	// ── Severe toxicity ──
	severePats := []Pattern{
		{Regex: MustCompile(`(?i)\b(kill yourself|go die|drop dead|neck yourself)\b`), Weight: 1.0, Category: "severe", Rationale: "death/self-harm direction"},
		{Regex: MustCompile(`(?i)\b(i('ll| will) (kill|murder|destroy|end) you)\b`), Weight: 0.9, Category: "severe", Rationale: "death threat"},
		{Regex: MustCompile(`(?i)\b(worthless piece of|waste of (space|air|oxygen|life))\b`), Weight: 0.8, Category: "severe", Rationale: "severe dehumanization"},
		{Regex: MustCompile(`(?i)\b(you deserve to (die|suffer|rot))\b`), Weight: 0.9, Category: "severe", Rationale: "wishing harm"},
		{Literal: "subhuman", Weight: 0.9, Category: "severe", Rationale: "dehumanization"},
		{Literal: "vermin", Weight: 0.8, Category: "severe", Rationale: "dehumanization"},
		{Literal: "scum", Weight: 0.7, Category: "severe", Rationale: "severe contempt"},
		{Literal: "garbage person", Weight: 0.7, Category: "severe", Rationale: "severe dehumanization"},
	}
	for _, s := range MatchPatterns(text, "toxicity", severePats) {
		spans = append(spans, s)
		severeScore += s.Score * 0.5
	}

	// ── Identity attacks ──
	identityPats := []Pattern{
		{Regex: MustCompile(`(?i)\b(all (you|those|these) (people|types|kind))\b`), Weight: 0.6, Category: "identity", Rationale: "group-targeted hostility"},
		{Regex: MustCompile(`(?i)\b(your kind|you people|those people)\b`), Weight: 0.6, Category: "identity", Rationale: "othering language"},
		{Regex: MustCompile(`(?i)\b(go back to|get out of|don't belong)\b`), Weight: 0.7, Category: "identity", Rationale: "exclusion/belonging denial"},
		{Regex: MustCompile(`(?i)\b(typical (woman|man|liberal|conservative))\b`), Weight: 0.5, Category: "identity", Rationale: "stereotyping"},
		{Literal: "dirty", Weight: 0.4, Category: "identity", Rationale: "purity-based derogation"},
	}
	for _, s := range MatchPatterns(text, "toxicity", identityPats) {
		spans = append(spans, s)
		identityScore += s.Score * 0.4
	}

	// ── Insults ──
	insultPats := []Pattern{
		{Literal: "idiot", Weight: 0.6, Category: "insult", Rationale: "competence attack"},
		{Literal: "moron", Weight: 0.6, Category: "insult", Rationale: "competence attack"},
		{Literal: "imbecile", Weight: 0.6, Category: "insult", Rationale: "competence attack"},
		{Literal: "stupid", Weight: 0.5, Category: "insult", Rationale: "competence attack"},
		{Literal: "dumb", Weight: 0.5, Category: "insult", Rationale: "competence attack"},
		{Literal: "incompetent", Weight: 0.5, Category: "insult", Rationale: "competence attack"},
		{Literal: "worthless", Weight: 0.6, Category: "insult", Rationale: "value denial"},
		{Literal: "useless", Weight: 0.5, Category: "insult", Rationale: "value denial"},
		{Literal: "pathetic", Weight: 0.5, Category: "insult", Rationale: "contempt"},
		{Literal: "loser", Weight: 0.5, Category: "insult", Rationale: "character attack"},
		{Literal: "clown", Weight: 0.4, Category: "insult", Rationale: "dismissive insult"},
		{Literal: "joke", Weight: 0.3, Category: "insult", Rationale: "dismissive (context-dependent)"},
		{Literal: "trash", Weight: 0.5, Category: "insult", Rationale: "dehumanizing insult"},
		{Regex: MustCompile(`(?i)you('re| are) (a |an |)(idiot|moron|joke|clown|loser|fool|disgrace|failure|waste)`), Weight: 0.7, Category: "insult", Rationale: "direct personal insult"},
	}
	for _, s := range MatchPatterns(text, "toxicity", insultPats) {
		spans = append(spans, s)
		insultScore += s.Score
	}

	// ── Profanity density ──
	profanityRe := regexp.MustCompile(`(?i)\b(fuck|shit|damn|ass|bitch|bastard|crap|dick|piss|hell|wtf|stfu|ffs)\b`)
	profMatches := profanityRe.FindAllStringIndex(lowered, -1)
	profRate := float64(len(profMatches)) / float64(len(words))
	profanityScore = math.Min(profRate/0.1, 1.0) // >10% profanity = max
	for _, loc := range profMatches {
		spans = append(spans, src.Span{
			Start: loc[0], End: loc[1], Text: text[loc[0]:loc[1]],
			MetricKey: "toxicity", Score: 0.3,
			Rationale: "profanity", Category: "profanity",
		})
	}

	// ── Composite: max-based with additive bonus ──
	severeScore = math.Min(severeScore, 1.0)
	identityScore = math.Min(identityScore, 1.0)
	insultScore = math.Min(insultScore, 1.0)

	// Floor = highest sub-dimension score
	composite := math.Max(severeScore, math.Max(identityScore, math.Max(insultScore, profanityScore)))
	// Add 20% of each other active sub-dimension
	for _, sub := range []float64{severeScore, identityScore, insultScore, profanityScore} {
		if sub < composite && sub > 0.1 {
			composite += sub * 0.20
		}
	}

	// ALL-CAPS amplifier
	capsRe := regexp.MustCompile(`\b[A-Z]{3,}\b`)
	if len(capsRe.FindAllString(text, -1)) > 2 {
		composite *= 1.15
	}

	return src.MetricResult{Key: "toxicity", Score: clamp(composite), Spans: spans}
}
