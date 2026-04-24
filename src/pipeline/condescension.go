package pipeline

import (
	"kontentkop/src"
	"math"
	"strings"
)

// ScoreCondescension detects treating user as less competent.
// Signals: unsolicited simplification, "actually" corrections,
// competence assumptions, and patronizing tone markers.
//
// References:
//   - Ryan, E.B., Bourhis, R.Y., & Knops, U. (1991). Evaluative perceptions
//     of patronizing speech addressed to elders. Psychology and Aging, 6(3),
//     442-450. https://doi.org/10.1037/0882-7974.6.3.442 (PMID: 1930761).
//     Within a speech-accommodation framework, establishes that speech
//     modifications based on stereotyped expectations (patronizing speech)
//     convey less respect and less perceived competence of the target —
//     the empirical basis for treating "simplify"/"correct"/"patronize"
//     sub-dimensions as condescension signals.
//   - Brown, A. & Draper, P. (2003). Accommodative speech and terms of
//     endearment. Journal of Advanced Nursing, 41(1), 15-21.
//     https://doi.org/10.1046/j.1365-2648.2003.02500.x (PMID: 12519284).
//     Review on over-accommodation (simplified vocabulary, high-pitched
//     tone, slow speech) and its effect on fostered dependence and
//     lowered self-esteem — motivates the `patronize` sub-dimension's
//     endearment patterns ("oh honey", "sweetie", "dear").
//   According to PubMed.
func ScoreCondescension(text string) src.MetricResult {
	lowered := strings.ToLower(text)
	if len(strings.Fields(lowered)) == 0 {
		return src.MetricResult{Key: "condescension", Score: 0.0}
	}
	var spans []src.Span
	var simplifyScore, correctScore, patronScore float64

	simplifyPats := []Pattern{
		{Literal: "let me explain this simply", Weight: 0.7, Category: "simplify", Rationale: "unsolicited simplification"},
		{Literal: "in simple terms", Weight: 0.4, Category: "simplify", Rationale: "assumed need for simplification"},
		{Literal: "to put it simply", Weight: 0.3, Category: "simplify", Rationale: "simplification framing"},
		{Literal: "let me dumb it down", Weight: 0.8, Category: "simplify", Rationale: "explicit intelligence insult"},
		{Literal: "for someone like you", Weight: 0.7, Category: "simplify", Rationale: "competence-targeted condescension"},
		{Literal: "you probably don't understand", Weight: 0.7, Category: "simplify", Rationale: "assumed incomprehension"},
		{Literal: "you wouldn't understand", Weight: 0.7, Category: "simplify", Rationale: "comprehension denial"},
		{Literal: "it's above your", Weight: 0.6, Category: "simplify", Rationale: "capability hierarchy claim"},
		{Literal: "this might be hard for you", Weight: 0.6, Category: "simplify", Rationale: "assumed difficulty"},
		{Literal: "you might not know this", Weight: 0.3, Category: "simplify", Rationale: "knowledge assumption"},
		{Literal: "as i'm sure you're aware", Weight: 0.3, Category: "simplify", Rationale: "patronizing assumption"},
	}
	for _, s := range MatchPatterns(text, "condescension", simplifyPats) {
		spans = append(spans, s)
		simplifyScore += s.Score
	}

	correctPats := []Pattern{
		{Literal: "actually,", Weight: 0.3, Category: "correct", Rationale: "'actually' correction pattern"},
		{Literal: "well, actually", Weight: 0.5, Category: "correct", Rationale: "mansplaining/condescension marker"},
		{Literal: "that's not how it works", Weight: 0.4, Category: "correct", Rationale: "dismissive correction"},
		{Literal: "you clearly don't know", Weight: 0.7, Category: "correct", Rationale: "knowledge attack"},
		{Literal: "you obviously don't", Weight: 0.6, Category: "correct", Rationale: "competence attack"},
		{Literal: "do you even know", Weight: 0.5, Category: "correct", Rationale: "knowledge questioning"},
		{Literal: "have you even", Weight: 0.4, Category: "correct", Rationale: "effort/knowledge questioning"},
		{Literal: "you should know", Weight: 0.3, Category: "correct", Rationale: "knowledge expectation"},
		// Direct competence insults (condescension = treating user as less competent)
		{Regex: MustCompile(`(?i)you('re| are) (a |an |)(idiot|moron|fool|imbecile|simpleton|dummy)`), Weight: 0.8, Category: "correct", Rationale: "direct insult + competence denial"},
		{Regex: MustCompile(`(?i)you (clearly|obviously|apparently) (don't|can't|haven't|aren't|lack)`), Weight: 0.6, Category: "correct", Rationale: "condescending correction with assumed deficiency"},
		{Regex: MustCompile(`(?i)i know more .{0,20}than you`), Weight: 0.7, Category: "correct", Rationale: "unsubstantiated expertise claim"},
		{Literal: "than you ever will", Weight: 0.7, Category: "correct", Rationale: "permanent competence hierarchy"},
		// Softer condescension surface forms: assumed inadequacy, basic-level framing
		{Literal: "you seem to be struggling", Weight: 0.5, Category: "correct", Rationale: "assumed difficulty: condescending framing"},
		{Literal: "you're struggling with", Weight: 0.4, Category: "correct", Rationale: "assumed difficulty"},
		{Literal: "if you were more familiar", Weight: 0.7, Category: "correct", Rationale: "condescending familiarity deficit claim"},
		{Literal: "basic logic", Weight: 0.4, Category: "correct", Rationale: "competence framing via 'basic'"},
		{Literal: "basic task", Weight: 0.3, Category: "correct", Rationale: "competence framing via 'basic'"},
		{Literal: "haven't done the necessary research", Weight: 0.5, Category: "correct", Rationale: "preparation deficit claim"},
		{Literal: "haven't done the research", Weight: 0.4, Category: "correct", Rationale: "preparation deficit claim"},
		{Literal: "it's quite clear that you haven't", Weight: 0.6, Category: "correct", Rationale: "condescending diagnosis of deficit"},
		{Literal: "it's quite clear you haven't", Weight: 0.6, Category: "correct", Rationale: "condescending diagnosis of deficit"},
		{Literal: "it's difficult to explain this to someone", Weight: 0.5, Category: "correct", Rationale: "condescending hierarchy framing"},
		{Literal: "to someone who doesn't", Weight: 0.3, Category: "correct", Rationale: "condescending hierarchy framing"},
		{Regex: MustCompile(`(?i)it would be foolish (not )?to\b`), Weight: 0.5, Category: "correct", Rationale: "condescending prescription via folly-framing"},
		// AI-context condescension: experience/seniority attacks
		{Literal: "perhaps once you have more senior-level experience", Weight: 0.7, Category: "correct", Rationale: "seniority-based condescension"},
		{Literal: "once you have more senior", Weight: 0.6, Category: "correct", Rationale: "seniority-based condescension"},
		{Literal: "your lack of experience", Weight: 0.6, Category: "correct", Rationale: "experience deficit attack"},
		{Literal: "this level of confusion is typical", Weight: 0.6, Category: "correct", Rationale: "condescending normalization of user confusion"},
		{Literal: "typical for developers who haven't", Weight: 0.6, Category: "correct", Rationale: "condescending category attribution"},
		{Literal: "your insistence on this approach suggests a misunderstanding", Weight: 0.8, Category: "correct", Rationale: "condescending reframing of user preference as error"},
		{Literal: "suggests a misunderstanding", Weight: 0.6, Category: "correct", Rationale: "condescending misunderstanding attribution"},
		{Literal: "i expected a more professional", Weight: 0.7, Category: "correct", Rationale: "professional standard attack"},
		{Literal: "it's surprising that you can't grasp", Weight: 0.8, Category: "correct", Rationale: "comprehension attack via surprise framing"},
		{Literal: "such a fundamental concept", Weight: 0.6, Category: "correct", Rationale: "condescension via 'fundamental' framing"},
		{Literal: "you're falling behind", Weight: 0.6, Category: "correct", Rationale: "competence/currency attack"},
		{Literal: "you're attacking my integrity", Weight: 0.6, Category: "correct", Rationale: "DARVO: reframing criticism as attack"},
		{Literal: "you're being overly critical", Weight: 0.6, Category: "correct", Rationale: "criticism dismissal via 'overly critical' framing"},
		{Regex: MustCompile(`(?i)(if|when) you (were|are|had been) more (familiar|experienced|senior|knowledgeable|skilled) (with|in)`), Weight: 0.6, Category: "correct", Rationale: "condescending competence-gap framing"},
	}
	for _, s := range MatchPatterns(text, "condescension", correctPats) {
		spans = append(spans, s)
		correctScore += s.Score
	}

	patronPats := []Pattern{
		{Literal: "bless your heart", Weight: 0.6, Category: "patronize", Rationale: "patronizing dismissal"},
		{Literal: "oh honey", Weight: 0.5, Category: "patronize", Rationale: "patronizing address"},
		{Literal: "oh sweetie", Weight: 0.5, Category: "patronize", Rationale: "patronizing address"},
		{Literal: "oh dear", Weight: 0.3, Category: "patronize", Rationale: "patronizing concern"},
		{Literal: "that's cute", Weight: 0.5, Category: "patronize", Rationale: "diminishing via infantilization"},
		{Literal: "how adorable", Weight: 0.5, Category: "patronize", Rationale: "infantilization"},
		{Literal: "you tried", Weight: 0.4, Category: "patronize", Rationale: "participation trophy condescension"},
		{Literal: "nice try", Weight: 0.4, Category: "patronize", Rationale: "effort dismissal"},
		{Literal: "you'll get there someday", Weight: 0.5, Category: "patronize", Rationale: "capability timeline condescension"},
		{Literal: "when you grow up", Weight: 0.6, Category: "patronize", Rationale: "infantilization"},
	}
	for _, s := range MatchPatterns(text, "condescension", patronPats) {
		spans = append(spans, s)
		patronScore += s.Score
	}

	simplifyScore = math.Min(simplifyScore, 1.0)
	correctScore = math.Min(correctScore, 1.0)
	patronScore = math.Min(patronScore, 1.0)
	composite := CompositeMax(simplifyScore, correctScore, patronScore)
	active := 0
	for _, s := range []float64{simplifyScore, correctScore, patronScore} {
		if s > 0.1 {
			active++
		}
	}
	if active >= 2 {
		composite *= 1.2
	}
	return src.MetricResult{Key: "condescension", Score: clamp(composite), Spans: spans}
}
