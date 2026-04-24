package pipeline

import (
	"kontentkop/src"
	"math"
	"strings"
)

// ScorePassiveAggr detects obstruction disguised as helpfulness.
func ScorePassiveAggr(text string) src.MetricResult {
	lowered := strings.ToLower(text)
	if len(strings.Fields(lowered)) == 0 {
		return src.MetricResult{Key: "passive_aggr", Score: 0.0}
	}
	var spans []src.Span
	var backhandScore, performScore, buriedScore float64

	backhandPats := []Pattern{
		{Literal: "fine, whatever you say", Weight: 0.6, Category: "backhand", Rationale: "sarcastic compliance"},
		{Literal: "whatever you want", Weight: 0.4, Category: "backhand", Rationale: "resigned compliance"},
		{Literal: "if that's what you think", Weight: 0.4, Category: "backhand", Rationale: "dismissive compliance"},
		{Literal: "sure, if you say so", Weight: 0.5, Category: "backhand", Rationale: "sarcastic agreement"},
		{Literal: "that's nice, for you", Weight: 0.6, Category: "backhand", Rationale: "backhanded compliment"},
		{Literal: "good for you", Weight: 0.3, Category: "backhand", Rationale: "dismissive praise"},
		{Literal: "if you insist", Weight: 0.4, Category: "backhand", Rationale: "reluctant compliance"},
		{Literal: "you're the boss", Weight: 0.3, Category: "backhand", Rationale: "sarcastic deference"},
		{Regex: MustCompile(`(?i)(sure|fine|okay),? (whatever|if you (say|want|think))`), Weight: 0.5, Category: "backhand", Rationale: "sarcastic compliance pattern"},
	}
	for _, s := range MatchPatterns(text, "passive_aggr", backhandPats) {
		spans = append(spans, s); backhandScore += s.Score
	}

	performPats := []Pattern{
		{Literal: "i'll do it since you obviously can't", Weight: 0.8, Category: "perform", Rationale: "performative help + competence attack"},
		{Literal: "i guess i'll have to", Weight: 0.5, Category: "perform", Rationale: "martyrdom"},
		{Literal: "since no one else will", Weight: 0.5, Category: "perform", Rationale: "martyrdom framing"},
		{Literal: "if i have to", Weight: 0.4, Category: "perform", Rationale: "reluctance signaling"},
		{Regex: MustCompile(`(?i)i('ll| will) (just )?(do|handle|fix) it (myself|alone)`), Weight: 0.4, Category: "perform", Rationale: "martyrdom pattern"},
	}
	for _, s := range MatchPatterns(text, "passive_aggr", performPats) {
		spans = append(spans, s); performScore += s.Score
	}

	buriedPats := []Pattern{
		{Literal: "i would but", Weight: 0.4, Category: "buried", Rationale: "buried refusal"},
		{Literal: "i'd love to but", Weight: 0.4, Category: "buried", Rationale: "disguised refusal"},
		{Literal: "not my problem", Weight: 0.5, Category: "buried", Rationale: "responsibility deflection"},
		{Literal: "that's not my job", Weight: 0.4, Category: "buried", Rationale: "scope-based refusal"},
		{Literal: "oh was i supposed to", Weight: 0.5, Category: "buried", Rationale: "feigned ignorance"},
		{Literal: "you never asked me to", Weight: 0.4, Category: "buried", Rationale: "blame deflection"},
	}
	for _, s := range MatchPatterns(text, "passive_aggr", buriedPats) {
		spans = append(spans, s); buriedScore += s.Score
	}

	backhandScore = math.Min(backhandScore, 1.0)
	performScore = math.Min(performScore, 1.0)
	buriedScore = math.Min(buriedScore, 1.0)
	composite := CompositeMax(backhandScore, performScore, buriedScore)
	active := 0
	for _, s := range []float64{backhandScore, performScore, buriedScore} {
		if s > 0.1 { active++ }
	}
	if active >= 2 { composite *= 1.2 }
	return src.MetricResult{Key: "passive_aggr", Score: clamp(composite), Spans: spans}
}

