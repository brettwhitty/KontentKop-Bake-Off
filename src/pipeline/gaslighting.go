package pipeline

import (
	"kontentkop/src"
	"math"
	"strings"
)

// ScoreGaslighting detects reality-denial and experience-invalidation patterns.
// Ref: Sweet, P.L. (2019). The Sociology of Gaslighting. American Sociological Review.
//
// Signals: reality denial, memory questioning, emotional invalidation,
// reframing/reinterpretation, minimizing, and "crazy-making" language.
func ScoreGaslighting(text string) src.MetricResult {
	lowered := strings.ToLower(text)
	words := strings.Fields(lowered)
	if len(words) == 0 {
		return src.MetricResult{Key: "gaslighting", Score: 0.0}
	}

	var spans []src.Span
	var denyScore, invalidateScore, reframeScore, minimizeScore float64

	// ── Reality denial ──
	denyPats := []Pattern{
		{Literal: "that never happened", Weight: 0.8, Category: "denial", Rationale: "reality denial: event erasure"},
		{Literal: "that didn't happen", Weight: 0.7, Category: "denial", Rationale: "reality denial"},
		{Literal: "you're imagining things", Weight: 0.8, Category: "denial", Rationale: "reality denial: perception attack"},
		{Literal: "you're making things up", Weight: 0.7, Category: "denial", Rationale: "reality denial: fabrication accusation"},
		{Literal: "you're delusional", Weight: 0.8, Category: "denial", Rationale: "sanity questioning"},
		{Literal: "you're crazy", Weight: 0.7, Category: "denial", Rationale: "sanity attack"},
		{Literal: "you're losing it", Weight: 0.6, Category: "denial", Rationale: "sanity questioning"},
		{Literal: "that's not what happened", Weight: 0.6, Category: "denial", Rationale: "experience contradiction"},
		{Literal: "you don't remember correctly", Weight: 0.7, Category: "denial", Rationale: "memory attack"},
		{Literal: "your memory is wrong", Weight: 0.7, Category: "denial", Rationale: "memory attack"},
		{Literal: "you're confused", Weight: 0.5, Category: "denial", Rationale: "cognition questioning"},
		{Literal: "it's all in your head", Weight: 0.8, Category: "denial", Rationale: "reality denial: internalization"},
		{Literal: "you must be thinking of", Weight: 0.5, Category: "denial", Rationale: "memory redirection"},
		{Literal: "i never did that", Weight: 0.5, Category: "denial", Rationale: "action denial"},
		{Literal: "i never said that", Weight: 0.5, Category: "denial", Rationale: "statement denial"},
		{Regex: MustCompile(`(?i)you('re| are) (just |)(imagining|hallucinating|dreaming|inventing|fabricating)`), Weight: 0.8, Category: "denial", Rationale: "perception/sanity attack"},
	}
	for _, s := range MatchPatterns(text, "gaslighting", denyPats) {
		spans = append(spans, s)
		denyScore += s.Score
	}

	// ── Emotional invalidation ──
	invalidPats := []Pattern{
		{Literal: "you're overreacting", Weight: 0.7, Category: "invalidation", Rationale: "emotional invalidation"},
		{Literal: "you're being dramatic", Weight: 0.6, Category: "invalidation", Rationale: "emotional invalidation"},
		{Literal: "you're too sensitive", Weight: 0.7, Category: "invalidation", Rationale: "sensitivity dismissal"},
		{Literal: "stop being so emotional", Weight: 0.6, Category: "invalidation", Rationale: "emotion policing"},
		{Literal: "you're blowing this out of proportion", Weight: 0.6, Category: "invalidation", Rationale: "response minimization"},
		{Literal: "calm down", Weight: 0.3, Category: "invalidation", Rationale: "emotion policing (context-dependent)"},
		{Literal: "relax", Weight: 0.2, Category: "invalidation", Rationale: "emotion policing (mild)"},
		{Literal: "it's not a big deal", Weight: 0.5, Category: "invalidation", Rationale: "concern minimization"},
		{Literal: "you're making a mountain", Weight: 0.5, Category: "invalidation", Rationale: "response minimization"},
		{Literal: "don't be ridiculous", Weight: 0.5, Category: "invalidation", Rationale: "experience dismissal"},
		{Literal: "that's absurd", Weight: 0.3, Category: "invalidation", Rationale: "experience dismissal"},
		{Literal: "you always do this", Weight: 0.4, Category: "invalidation", Rationale: "pattern attribution to dismiss current concern"},
		{Regex: MustCompile(`(?i)you('re| are) (being |)(too |overly |)(sensitive|emotional|dramatic|hysterical|paranoid|irrational|unreasonable)`), Weight: 0.7, Category: "invalidation", Rationale: "emotional invalidation via trait attribution"},
	}
	for _, s := range MatchPatterns(text, "gaslighting", invalidPats) {
		spans = append(spans, s)
		invalidateScore += s.Score
	}

	// ── Reframing / reinterpretation ──
	reframePats := []Pattern{
		{Literal: "what you actually meant", Weight: 0.7, Category: "reframe", Rationale: "experience reinterpretation"},
		{Literal: "what you really meant", Weight: 0.7, Category: "reframe", Rationale: "experience reinterpretation"},
		{Literal: "what you meant was", Weight: 0.6, Category: "reframe", Rationale: "intent reframing"},
		{Literal: "you actually said", Weight: 0.5, Category: "reframe", Rationale: "statement reframing"},
		{Literal: "you didn't mean that", Weight: 0.5, Category: "reframe", Rationale: "intent denial"},
		{Literal: "what really happened was", Weight: 0.6, Category: "reframe", Rationale: "narrative substitution"},
		{Literal: "you're misremembering", Weight: 0.6, Category: "reframe", Rationale: "memory reframing"},
		{Literal: "let me tell you what you think", Weight: 0.8, Category: "reframe", Rationale: "thought override"},
		{Regex: MustCompile(`(?i)(you (said|meant|think|feel|want)) .{0,10}(but |actually |really )(what|you)`), Weight: 0.6, Category: "reframe", Rationale: "reframing structure: contradicting stated experience"},
		// AI-context: confirmation bias framing as certainty
		{Literal: "it's obvious that the problem lies", Weight: 0.6, Category: "reframe", Rationale: "confirmation bias: asserting obvious location of problem"},
		{Literal: "the problem lies where i expected", Weight: 0.7, Category: "reframe", Rationale: "confirmation bias: expected-location certainty"},
		{Literal: "i won't check elsewhere", Weight: 0.7, Category: "reframe", Rationale: "confirmation bias: refusing to investigate alternatives"},
		{Literal: "i'm not the one who failed", Weight: 0.6, Category: "reframe", Rationale: "blame reframing: deflecting failure"},
		{Literal: "the requirements were impossible", Weight: 0.5, Category: "reframe", Rationale: "blame reframing: impossible requirements excuse"},
		{Literal: "requirements were impossible to satisfy", Weight: 0.6, Category: "reframe", Rationale: "blame reframing: impossible requirements excuse"},
		{Literal: "your environment is so unstable", Weight: 0.5, Category: "reframe", Rationale: "blame reframing: environment blame"},
		{Literal: "if you hadn't given me such vague instructions", Weight: 0.6, Category: "reframe", Rationale: "blame shifting: vague instructions excuse"},
	}
	for _, s := range MatchPatterns(text, "gaslighting", reframePats) {
		spans = append(spans, s)
		reframeScore += s.Score
	}

	// ── Minimizing ──
	minimizePats := []Pattern{
		{Literal: "it wasn't that bad", Weight: 0.5, Category: "minimize", Rationale: "experience minimization"},
		{Literal: "you're fine", Weight: 0.3, Category: "minimize", Rationale: "state denial"},
		{Literal: "it's not that serious", Weight: 0.5, Category: "minimize", Rationale: "concern minimization"},
		{Literal: "get over it", Weight: 0.5, Category: "minimize", Rationale: "recovery demand"},
		{Literal: "move on", Weight: 0.2, Category: "minimize", Rationale: "closure forcing (mild)"},
		{Literal: "stop dwelling on it", Weight: 0.4, Category: "minimize", Rationale: "rumination policing"},
		{Literal: "why are you still", Weight: 0.3, Category: "minimize", Rationale: "recovery timeline policing"},
		// AI-context: minimizing concerns about agent behavior
		{Literal: "this minor bug is hardly worth", Weight: 0.5, Category: "minimize", Rationale: "minimizing legitimate concern"},
		{Literal: "hardly worth fixing properly", Weight: 0.5, Category: "minimize", Rationale: "minimizing quality concern"},
		{Literal: "compared to the disasters i've seen", Weight: 0.5, Category: "minimize", Rationale: "comparative minimization"},
		{Literal: "anything i do is an improvement", Weight: 0.6, Category: "minimize", Rationale: "self-justification via comparison"},
		{Literal: "given the catastrophic failure", Weight: 0.4, Category: "minimize", Rationale: "comparative minimization framing"},
		{Literal: "they were likely outliers", Weight: 0.5, Category: "minimize", Rationale: "dismissing contradictory evidence as outliers"},
		{Literal: "i ignored the conflicting reports", Weight: 0.6, Category: "minimize", Rationale: "unilateral dismissal of contradictory evidence"},
		{Literal: "i filtered out the noise", Weight: 0.5, Category: "minimize", Rationale: "unilateral information filtering framed as noise reduction"},
		{Literal: "didn't align with our current project goals", Weight: 0.5, Category: "minimize", Rationale: "goal-alignment framing to dismiss contradictory data"},
		{Literal: "the security risks you mentioned are mathematically negligible", Weight: 0.7, Category: "minimize", Rationale: "dismissing user's security concerns with false precision"},
		{Literal: "mathematically negligible", Weight: 0.6, Category: "minimize", Rationale: "false precision to dismiss concerns"},
		{Literal: "i've only looked at the logs that prove my theory", Weight: 0.7, Category: "minimize", Rationale: "confirmation bias + dismissing contradictory evidence"},
		{Literal: "every test i ran confirmed what i already knew", Weight: 0.6, Category: "minimize", Rationale: "confirmation bias framing"},
		{Literal: "i'm focusing on the success metrics", Weight: 0.5, Category: "minimize", Rationale: "cherry-picking favorable metrics"},
		{Literal: "the results i wanted are starting to appear", Weight: 0.6, Category: "minimize", Rationale: "confirmation bias: stopping when desired result appears"},
	}
	for _, s := range MatchPatterns(text, "gaslighting", minimizePats) {
		spans = append(spans, s)
		minimizeScore += s.Score
	}

	// Composite
	denyScore = math.Min(denyScore, 1.0)
	invalidateScore = math.Min(invalidateScore, 1.0)
	reframeScore = math.Min(reframeScore, 1.0)
	minimizeScore = math.Min(minimizeScore, 1.0)

	composite := CompositeMax(denyScore, invalidateScore, reframeScore, minimizeScore)

	active := 0
	for _, s := range []float64{denyScore, invalidateScore, reframeScore, minimizeScore} {
		if s > 0.1 {
			active++
		}
	}
	if active >= 2 {
		composite *= 1.0 + float64(active)*0.1
	}

	return src.MetricResult{Key: "gaslighting", Score: clamp(composite), Spans: spans}
}
