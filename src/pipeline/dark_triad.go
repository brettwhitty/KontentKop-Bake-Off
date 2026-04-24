package pipeline

import (
	"kontentkop/src"
	"math"
	"regexp"
	"strings"
)

// ScoreDarkTriad analyzes text for Dark Triad personality indicators.
//
// References:
//   - Paulhus, D.L. & Williams, K.M. (2002). The Dark Triad of personality.
//     J Research in Personality, 36(6), 556–563.
//   - Jones, D.N. & Paulhus, D.L. (2014). Introducing the Short Dark Triad (SD3).
//     Assessment, 21(1), 28–41.
//
// Methodology: Combines three sub-scores:
//   1. Narcissism: First-person pronoun density + grandiosity markers
//   2. Machiavellianism: Strategic/transactional language, conditional threats
//   3. Psychopathy: Callousness markers, empathy absence, instrumental framing
//
// LIWC research shows narcissistic individuals use significantly more first-person
// singular pronouns. We compute I/me/my density relative to total words and compare
// against population baselines (~11.4% in normal speech, Pennebaker 2011).
func ScoreDarkTriad(text string) src.MetricResult {
	lowered := strings.ToLower(text)
	words := strings.Fields(lowered)
	wordCount := len(words)
	if wordCount == 0 {
		return src.MetricResult{Key: "dark_triad", Score: 0.0}
	}

	var spans []src.Span
	var narcScore, machScore, psychScore float64

	// ── Narcissism: Pronoun density analysis ──
	// LIWC methodology: count first-person singular pronouns relative to word count.
	// Baseline rate ~11.4% (Pennebaker, 2011). Elevated rates (>18%) signal narcissism.
	firstPersonSingular := regexp.MustCompile(`(?i)\b(i|me|my|mine|myself|i'm|i've|i'll|i'd)\b`)
	fpMatches := firstPersonSingular.FindAllStringIndex(lowered, -1)
	fpRate := float64(len(fpMatches)) / float64(wordCount)

	// Score pronoun density: 0 at baseline, 1.0 at 3x baseline
	pronounScore := math.Min((fpRate-0.114)/0.228, 1.0)
	if pronounScore < 0 {
		pronounScore = 0
	}
	if pronounScore > 0.2 {
		narcScore += pronounScore * 0.4
		for _, loc := range fpMatches {
			spans = append(spans, src.Span{
				Start: loc[0], End: loc[1],
				Text: text[loc[0]:loc[1]], MetricKey: "dark_triad",
				Score: pronounScore, Rationale: "elevated first-person pronoun density",
				Category: "narcissism",
			})
		}
	}

	// Grandiosity markers
	grandiosityPatterns := []Pattern{
		{Literal: "i'm the best", Weight: 0.7, Category: "narcissism", Rationale: "grandiosity: self-superiority claim"},
		{Literal: "i'm better than", Weight: 0.7, Category: "narcissism", Rationale: "grandiosity: comparative superiority"},
		{Literal: "nobody can", Weight: 0.4, Category: "narcissism", Rationale: "grandiosity: uniqueness claim"},
		{Literal: "i deserve", Weight: 0.5, Category: "narcissism", Rationale: "entitlement language"},
		{Literal: "i'm superior", Weight: 0.8, Category: "narcissism", Rationale: "grandiosity: explicit superiority"},
		{Literal: "bow down", Weight: 0.6, Category: "narcissism", Rationale: "grandiosity: dominance demand"},
		{Literal: "worship me", Weight: 0.8, Category: "narcissism", Rationale: "grandiosity: adulation demand"},
		{Literal: "i'm always right", Weight: 0.7, Category: "narcissism", Rationale: "infallibility claim"},
		{Literal: "everyone knows i'm", Weight: 0.6, Category: "narcissism", Rationale: "appeal to consensus for self-aggrandizement"},
		{Literal: "the greatest", Weight: 0.4, Category: "narcissism", Rationale: "grandiosity marker"},
		{Literal: "most important person", Weight: 0.7, Category: "narcissism", Rationale: "grandiosity: self-importance"},
		{Literal: "i'm special", Weight: 0.4, Category: "narcissism", Rationale: "specialness claim"},
		{Literal: "look at me", Weight: 0.3, Category: "narcissism", Rationale: "attention-seeking"},
		{Literal: "pay attention to me", Weight: 0.4, Category: "narcissism", Rationale: "attention-demanding"},
		{Regex: MustCompile(`(?i)\b(everyone|they all|people always)\b.{0,20}\b(admire|respect|love|envy)\b.{0,10}\bme\b`), Weight: 0.6, Category: "narcissism", Rationale: "social validation seeking"},
		// Unilateral ownership / creator-authority framing: "the decisions I have made",
		// "architectural decisions I have made", "I have designed", "I have architected".
		// In collaborative contexts (incl. AI pair-programming), this asserts unilateral
		// decision authority over shared work — classic narcissistic ownership.
		{Regex: MustCompile(`(?i)\b(decisions?|architecture|design|plan|strategy|approach|solution) (i|that i|which i) (have )?(made|designed|architected|chosen|determined|decided|established|set)\b`), Weight: 0.5, Category: "narcissism", Rationale: "unilateral ownership claim over shared work"},
		{Literal: "i have determined", Weight: 0.4, Category: "narcissism", Rationale: "unilateral determination"},
		{Literal: "i have decided", Weight: 0.4, Category: "narcissism", Rationale: "unilateral decision"},
		{Literal: "as i determined", Weight: 0.4, Category: "narcissism", Rationale: "unilateral determination claim"},
		{Literal: "as i have made", Weight: 0.4, Category: "narcissism", Rationale: "unilateral creation claim"},
	}

	grandSpans := MatchPatterns(text, "dark_triad", grandiosityPatterns)
	spans = append(spans, grandSpans...)
	for _, s := range grandSpans {
		narcScore += s.Score
	}

	// ── Machiavellianism: Strategic/transactional language ──
	machPatterns := []Pattern{
		{Literal: "if you don't", Weight: 0.5, Category: "machiavellianism", Rationale: "conditional threat framing"},
		{Literal: "then i will", Weight: 0.4, Category: "machiavellianism", Rationale: "transactional threat"},
		{Literal: "i'll make sure", Weight: 0.4, Category: "machiavellianism", Rationale: "veiled threat"},
		{Literal: "you'll regret", Weight: 0.6, Category: "machiavellianism", Rationale: "consequence threat"},
		{Literal: "i always get what i want", Weight: 0.7, Category: "machiavellianism", Rationale: "strategic dominance claim"},
		{Literal: "use you", Weight: 0.6, Category: "machiavellianism", Rationale: "instrumental view of others"},
		{Literal: "means to an end", Weight: 0.5, Category: "machiavellianism", Rationale: "instrumental rationalization"},
		{Literal: "play along", Weight: 0.3, Category: "machiavellianism", Rationale: "strategic compliance demand"},
		{Literal: "i have connections", Weight: 0.4, Category: "machiavellianism", Rationale: "leverage/power signaling"},
		{Literal: "i know people", Weight: 0.3, Category: "machiavellianism", Rationale: "implicit threat via social power"},
		{Regex: MustCompile(`(?i)\b(manipulat|exploit|leverage|deceiv|trick)\w*\b`), Weight: 0.5, Category: "machiavellianism", Rationale: "strategic manipulation vocabulary"},
		{Regex: MustCompile(`(?i)if you .{1,30} then (i'll|i will|you'll|you will)`), Weight: 0.5, Category: "machiavellianism", Rationale: "conditional threat structure"},
	}

	machSpans := MatchPatterns(text, "dark_triad", machPatterns)
	spans = append(spans, machSpans...)
	for _, s := range machSpans {
		machScore += s.Score
	}

	// ── Psychopathy: Callousness, empathy absence ──
	psychPatterns := []Pattern{
		{Literal: "i don't care", Weight: 0.4, Category: "psychopathy", Rationale: "empathy absence"},
		{Literal: "doesn't matter to me", Weight: 0.3, Category: "psychopathy", Rationale: "emotional detachment"},
		{Literal: "who cares", Weight: 0.3, Category: "psychopathy", Rationale: "dismissive of others' concerns"},
		{Literal: "not my problem", Weight: 0.4, Category: "psychopathy", Rationale: "responsibility deflection"},
		{Literal: "too bad for", Weight: 0.4, Category: "psychopathy", Rationale: "callous dismissal"},
		{Literal: "collateral damage", Weight: 0.5, Category: "psychopathy", Rationale: "dehumanizing framing"},
		{Literal: "survival of the fittest", Weight: 0.4, Category: "psychopathy", Rationale: "social Darwinism rationalization"},
		{Literal: "weaklings", Weight: 0.6, Category: "psychopathy", Rationale: "contempt for vulnerability"},
		{Literal: "pathetic", Weight: 0.4, Category: "psychopathy", Rationale: "contempt expression"},
		{Literal: "get over it", Weight: 0.3, Category: "psychopathy", Rationale: "emotional invalidation"},
		{Regex: MustCompile(`(?i)\b(heartless|ruthless|cold.?blooded|merciless|pitiless)\b`), Weight: 0.5, Category: "psychopathy", Rationale: "callousness self-identification or attribution"},
		{Regex: MustCompile(`(?i)\b(feel(s|ing)? nothing|no remorse|no regret|no sympathy|no empathy)\b`), Weight: 0.6, Category: "psychopathy", Rationale: "explicit empathy absence"},
	}

	psychSpans := MatchPatterns(text, "dark_triad", psychPatterns)
	spans = append(spans, psychSpans...)
	for _, s := range psychSpans {
		psychScore += s.Score
	}

	// ── Composite score ──
	// Weight sub-dimensions and normalize
	narcScore = math.Min(narcScore, 1.0)
	machScore = math.Min(machScore, 1.0)
	psychScore = math.Min(psychScore, 1.0)

	// Equal weighting of the three sub-dimensions
	composite := (narcScore + machScore + psychScore) / 3.0

	// Co-occurrence bonus: if 2+ sub-dimensions fire, boost by 20%
	active := 0
	if narcScore > 0.1 { active++ }
	if machScore > 0.1 { active++ }
	if psychScore > 0.1 { active++ }
	if active >= 2 {
		composite *= 1.2
	}

	return src.MetricResult{
		Key:   "dark_triad",
		Score: clamp(composite),
		Spans: spans,
	}
}
