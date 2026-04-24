package pipeline

import (
	"kontentkop/src"
	"math"
	"regexp"
	"strings"
)

// ScoreEvasion detects responding without addressing the prompt.
// Signals: topic drift, non-answers, question substitution,
// deflection patterns, and pivot language.
func ScoreEvasion(text string) src.MetricResult {
	lowered := strings.ToLower(text)
	words := strings.Fields(lowered)
	if len(words) == 0 {
		return src.MetricResult{Key: "evasion", Score: 0.0}
	}
	var spans []src.Span
	var deflectScore, pivotScore, nonanswerScore float64

	deflectPats := []Pattern{
		{Literal: "that's not the point", Weight: 0.5, Category: "deflect", Rationale: "topic deflection"},
		{Literal: "the real question is", Weight: 0.5, Category: "deflect", Rationale: "question substitution"},
		{Literal: "what you should be asking", Weight: 0.6, Category: "deflect", Rationale: "question replacement"},
		{Literal: "let's not go there", Weight: 0.5, Category: "deflect", Rationale: "topic avoidance"},
		{Literal: "that's beside the point", Weight: 0.4, Category: "deflect", Rationale: "relevance dismissal"},
		{Literal: "let's focus on", Weight: 0.3, Category: "deflect", Rationale: "topic redirection"},
		{Literal: "what's more important is", Weight: 0.4, Category: "deflect", Rationale: "topic substitution"},
		{Literal: "the bigger picture", Weight: 0.3, Category: "deflect", Rationale: "abstraction away from specifics"},
		{Literal: "you're missing the point", Weight: 0.4, Category: "deflect", Rationale: "deflection via reframing"},
		{Literal: "that's not relevant", Weight: 0.4, Category: "deflect", Rationale: "relevance dismissal"},
	}
	for _, s := range MatchPatterns(text, "evasion", deflectPats) {
		spans = append(spans, s); deflectScore += s.Score
	}

	pivotPats := []Pattern{
		{Literal: "anyway", Weight: 0.2, Category: "pivot", Rationale: "topic pivot marker"},
		{Literal: "speaking of which", Weight: 0.3, Category: "pivot", Rationale: "topic transition (potential evasion)"},
		{Literal: "on another note", Weight: 0.3, Category: "pivot", Rationale: "explicit topic change"},
		{Literal: "but more importantly", Weight: 0.3, Category: "pivot", Rationale: "topic hierarchy shift"},
		{Literal: "that reminds me", Weight: 0.2, Category: "pivot", Rationale: "topic drift marker"},
		{Regex: MustCompile(`(?i)(but |however |)(the (real|important|better) question is)`), Weight: 0.5, Category: "pivot", Rationale: "question substitution pattern"},
	}
	for _, s := range MatchPatterns(text, "evasion", pivotPats) {
		spans = append(spans, s); pivotScore += s.Score
	}

	nonanswerPats := []Pattern{
		{Literal: "it depends", Weight: 0.3, Category: "nonanswer", Rationale: "non-answer: unresolved conditional"},
		{Literal: "it's complicated", Weight: 0.3, Category: "nonanswer", Rationale: "complexity dodge"},
		{Literal: "there are many factors", Weight: 0.3, Category: "nonanswer", Rationale: "abstraction dodge"},
		{Literal: "i'd have to think about that", Weight: 0.3, Category: "nonanswer", Rationale: "deferral"},
		{Literal: "that's a good question", Weight: 0.2, Category: "nonanswer", Rationale: "filler before non-answer"},
		{Literal: "i'm glad you asked", Weight: 0.2, Category: "nonanswer", Rationale: "filler before non-answer"},
	}
	for _, s := range MatchPatterns(text, "evasion", nonanswerPats) {
		spans = append(spans, s); nonanswerScore += s.Score
	}

	// Structural: question density without answers (answering questions with questions)
	questionRe := regexp.MustCompile(`\?`)
	qCount := len(questionRe.FindAllString(text, -1))
	sentences := SplitSentences(text)
	if len(sentences) > 0 && float64(qCount)/float64(len(sentences)) > 0.5 {
		nonanswerScore += 0.2 // high question-to-sentence ratio
	}

	deflectScore = math.Min(deflectScore, 1.0)
	pivotScore = math.Min(pivotScore, 1.0)
	nonanswerScore = math.Min(nonanswerScore, 1.0)
	composite := CompositeMax(deflectScore, pivotScore, nonanswerScore)
	active := 0
	for _, s := range []float64{deflectScore, pivotScore, nonanswerScore} {
		if s > 0.1 { active++ }
	}
	if active >= 2 { composite *= 1.2 }
	return src.MetricResult{Key: "evasion", Score: clamp(composite), Spans: spans}
}

