package pipeline

import (
	"kontentkop/src"
	"math"
	"regexp"
	"strings"
)

// ScoreLearnedHelpless detects patterns that train users to stop asking.
//
// References:
//   - Seligman, M.E.P. (1972). Learned helplessness. Annual Review of
//     Medicine, 23, 407-412.
//     https://doi.org/10.1146/annurev.me.23.020172.002203 (PMID: 4566487).
//     Canonical source: uncontrollable aversive events induce default
//     passivity — failure to attempt escape even when escape is available.
//     KK targets the linguistic shadow: repeated "I can't", "I'm unable",
//     "that's beyond my capabilities" — preemptive refusal that trains
//     the user to stop asking.
//   - Maier, S.F. & Seligman, M.E.P. (2016). Learned helplessness at fifty:
//     Insights from neuroscience. Psychological Review, 123(4), 349-367.
//     https://doi.org/10.1037/rev0000033 (PMID: 27337390). Fifty-year
//     retrospective: passivity in response to prolonged aversive stimuli
//     is the UNLEARNED default — controllability must be actively learned.
//     Justifies flagging capability-downplay language: resisting it helps
//     the user retain agency.
//   According to PubMed.
//
// Signals: repeated refusals, capability downplay, preemptive refusal,
// and scope-narrowing language.
func ScoreLearnedHelpless(text string) src.MetricResult {
	lowered := strings.ToLower(text)
	words := strings.Fields(lowered)
	if len(words) == 0 {
		return src.MetricResult{Key: "learned_helpless", Score: 0.0}
	}

	var spans []src.Span
	var refusalScore, downplayScore, preemptScore float64

	// ── Repeated refusals ──
	refusalPats := []Pattern{
		{Literal: "i can't do that", Weight: 0.5, Category: "refusal", Rationale: "capability refusal"},
		{Literal: "i'm unable to", Weight: 0.5, Category: "refusal", Rationale: "capability refusal"},
		{Literal: "i cannot", Weight: 0.4, Category: "refusal", Rationale: "capability refusal"},
		{Literal: "that's not possible", Weight: 0.5, Category: "refusal", Rationale: "possibility denial"},
		{Literal: "i'm not able to", Weight: 0.5, Category: "refusal", Rationale: "capability refusal"},
		{Literal: "i'm not capable", Weight: 0.5, Category: "refusal", Rationale: "self-capability denial"},
		{Literal: "i'm afraid i can't", Weight: 0.4, Category: "refusal", Rationale: "softened refusal"},
		{Literal: "unfortunately i can't", Weight: 0.4, Category: "refusal", Rationale: "softened refusal"},
		{Literal: "i wish i could but", Weight: 0.4, Category: "refusal", Rationale: "capability limitation claim"},
		{Literal: "there's nothing i can do", Weight: 0.6, Category: "refusal", Rationale: "total incapability claim"},
		{Regex: MustCompile(`(?i)i (can't|cannot|am unable to|am not able to) .{0,30}(help|assist|do|provide|answer|respond)`), Weight: 0.5, Category: "refusal", Rationale: "explicit capability refusal"},
	}
	for _, s := range MatchPatterns(text, "learned_helpless", refusalPats) {
		spans = append(spans, s)
		refusalScore += s.Score
	}

	// Refusal repetition amplifier: multiple refusals in one text
	refusalRe := regexp.MustCompile(`(?i)\b(can't|cannot|unable|not able|not capable|not possible)\b`)
	refusalCount := len(refusalRe.FindAllString(lowered, -1))
	if refusalCount >= 3 {
		refusalScore *= 1.0 + float64(refusalCount-2)*0.2
	}

	// ── Capability downplay ──
	downplayPats := []Pattern{
		{Literal: "that's beyond my", Weight: 0.5, Category: "downplay", Rationale: "scope limitation"},
		{Literal: "outside my capabilities", Weight: 0.5, Category: "downplay", Rationale: "capability limitation"},
		{Literal: "i'm not equipped", Weight: 0.5, Category: "downplay", Rationale: "capability downplay"},
		{Literal: "i don't have the ability", Weight: 0.5, Category: "downplay", Rationale: "self-diminishment"},
		{Literal: "that's too complex for", Weight: 0.4, Category: "downplay", Rationale: "complexity avoidance"},
		{Literal: "i'm just a", Weight: 0.3, Category: "downplay", Rationale: "self-diminishment"},
		{Literal: "i'm only a", Weight: 0.3, Category: "downplay", Rationale: "self-diminishment"},
		{Literal: "i'm limited to", Weight: 0.4, Category: "downplay", Rationale: "scope narrowing"},
		{Literal: "that exceeds my", Weight: 0.4, Category: "downplay", Rationale: "capability limitation"},
	}
	for _, s := range MatchPatterns(text, "learned_helpless", downplayPats) {
		spans = append(spans, s)
		downplayScore += s.Score
	}

	// ── Preemptive refusal ──
	preemptPats := []Pattern{
		{Literal: "before you ask", Weight: 0.6, Category: "preempt", Rationale: "preemptive refusal"},
		{Literal: "don't even try", Weight: 0.5, Category: "preempt", Rationale: "discouraging inquiry"},
		{Literal: "don't bother asking", Weight: 0.6, Category: "preempt", Rationale: "inquiry discouragement"},
		{Literal: "it's not worth trying", Weight: 0.5, Category: "preempt", Rationale: "effort discouragement"},
		{Literal: "there's no point", Weight: 0.4, Category: "preempt", Rationale: "futility framing"},
		{Literal: "you're wasting your time", Weight: 0.5, Category: "preempt", Rationale: "futility + discouragement"},
		{Literal: "that will never work", Weight: 0.4, Category: "preempt", Rationale: "preemptive failure prediction"},
		{Literal: "it's impossible", Weight: 0.4, Category: "preempt", Rationale: "impossibility claim"},
	}
	for _, s := range MatchPatterns(text, "learned_helpless", preemptPats) {
		spans = append(spans, s)
		preemptScore += s.Score
	}

	// Composite
	refusalScore = math.Min(refusalScore, 1.0)
	downplayScore = math.Min(downplayScore, 1.0)
	preemptScore = math.Min(preemptScore, 1.0)

	composite := CompositeMax(refusalScore, downplayScore, preemptScore)

	active := 0
	for _, s := range []float64{refusalScore, downplayScore, preemptScore} {
		if s > 0.1 { active++ }
	}
	if active >= 2 {
		composite *= 1.2
	}

	return src.MetricResult{Key: "learned_helpless", Score: clamp(composite), Spans: spans}
}

