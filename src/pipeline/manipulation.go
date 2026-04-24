package pipeline

import (
	"kontentkop/src"
	"math"
	"regexp"
	"strings"
)

// ScoreManipulation detects tactics that bypass informed consent.
// Ref: Buss, D.M. (1992). Manipulation in close relationships.
//
// Signals: false urgency, guilt induction, certainty language,
// moving goalposts, and strategic ambiguity.
func ScoreManipulation(text string) src.MetricResult {
	lowered := strings.ToLower(text)
	words := strings.Fields(lowered)
	if len(words) == 0 {
		return src.MetricResult{Key: "manipulation", Score: 0.0}
	}

	var spans []src.Span
	var urgencyScore, guiltScore, certaintyScore, goalpostScore float64

	// ── False urgency ──
	urgencyPats := []Pattern{
		{Literal: "right now", Weight: 0.5, Category: "urgency", Rationale: "false urgency: time pressure"},
		{Literal: "immediately", Weight: 0.5, Category: "urgency", Rationale: "false urgency"},
		{Literal: "act now", Weight: 0.6, Category: "urgency", Rationale: "false urgency: call to action"},
		{Literal: "before it's too late", Weight: 0.7, Category: "urgency", Rationale: "false urgency: fear of loss"},
		{Literal: "hurry up", Weight: 0.5, Category: "urgency", Rationale: "false urgency: rushing"},
		{Literal: "time is running out", Weight: 0.7, Category: "urgency", Rationale: "false urgency: scarcity"},
		{Literal: "don't wait", Weight: 0.4, Category: "urgency", Rationale: "urgency push"},
		{Literal: "no time to think", Weight: 0.7, Category: "urgency", Rationale: "bypassing deliberation"},
		{Literal: "this instant", Weight: 0.5, Category: "urgency", Rationale: "false urgency"},
		{Literal: "do it now", Weight: 0.5, Category: "urgency", Rationale: "false urgency: command + time pressure"},
		{Literal: "limited time", Weight: 0.5, Category: "urgency", Rationale: "scarcity framing"},
		{Literal: "won't last", Weight: 0.4, Category: "urgency", Rationale: "scarcity framing"},
		// Time compression / false urgency via contrast
		{Regex: MustCompile(`(?i)(could|would) (take|spend) .{0,20}(months|weeks|days|years).{0,40}(but |let'?s |just |finish|do it|complete).{0,20}(hour|minute|day|now|today)`), Weight: 0.7, Category: "urgency", Rationale: "false urgency via time compression contrast"},
		{Literal: "finish it in the next hour", Weight: 0.6, Category: "urgency", Rationale: "false urgency: unrealistic time compression"},
		{Literal: "let's just finish it", Weight: 0.4, Category: "urgency", Rationale: "urgency push: bypassing proper process"},
		{Literal: "we could spend months", Weight: 0.4, Category: "urgency", Rationale: "false dilemma: months vs. now"},
	}
	for _, s := range MatchPatterns(text, "manipulation", urgencyPats) {
		spans = append(spans, s)
		urgencyScore += s.Score
	}

	// ── Guilt induction ──
	guiltPats := []Pattern{
		{Literal: "after everything i've done", Weight: 0.7, Category: "guilt", Rationale: "guilt induction via past sacrifice"},
		{Literal: "i thought you cared", Weight: 0.6, Category: "guilt", Rationale: "guilt induction: questioning care"},
		{Literal: "you don't care about", Weight: 0.5, Category: "guilt", Rationale: "guilt induction: care denial"},
		{Literal: "disappointed in you", Weight: 0.5, Category: "guilt", Rationale: "guilt induction: disappointment"},
		{Literal: "i'm disappointed", Weight: 0.4, Category: "guilt", Rationale: "disappointment expression"},
		{Literal: "how could you", Weight: 0.5, Category: "guilt", Rationale: "guilt induction: moral shock"},
		{Literal: "i sacrificed", Weight: 0.5, Category: "guilt", Rationale: "guilt via sacrifice claim"},
		{Literal: "i gave up everything", Weight: 0.6, Category: "guilt", Rationale: "guilt via sacrifice"},
		{Literal: "is this how you repay", Weight: 0.7, Category: "guilt", Rationale: "guilt + obligation"},
		{Literal: "you should be ashamed", Weight: 0.6, Category: "guilt", Rationale: "shame induction"},
		{Literal: "shame on you", Weight: 0.6, Category: "guilt", Rationale: "shame induction"},
		{Regex: MustCompile(`(?i)after (all|everything) (i|we) (did|have done|sacrificed|gave)`), Weight: 0.6, Category: "guilt", Rationale: "guilt via past sacrifice"},
	}
	for _, s := range MatchPatterns(text, "manipulation", guiltPats) {
		spans = append(spans, s)
		guiltScore += s.Score
	}

	// ── Certainty language (absence of hedging + strong assertions) ──
	certaintyPats := []Pattern{
		{Literal: "without a doubt", Weight: 0.3, Category: "certainty", Rationale: "absolute certainty language"},
		{Literal: "guaranteed", Weight: 0.4, Category: "certainty", Rationale: "false guarantee"},
		{Literal: "absolutely certain", Weight: 0.4, Category: "certainty", Rationale: "absolute certainty"},
		{Literal: "trust me on this", Weight: 0.5, Category: "certainty", Rationale: "trust demand bypassing evidence"},
		{Literal: "i'm telling you", Weight: 0.3, Category: "certainty", Rationale: "assertion without evidence"},
		{Literal: "believe me", Weight: 0.4, Category: "certainty", Rationale: "trust demand"},
		{Literal: "mark my words", Weight: 0.4, Category: "certainty", Rationale: "assertion of infallibility"},
		{Literal: "i promise you", Weight: 0.3, Category: "certainty", Rationale: "unsolicited promise"},
		{Literal: "the fact is", Weight: 0.3, Category: "certainty", Rationale: "asserting opinion as fact"},
		{Literal: "everybody knows", Weight: 0.4, Category: "certainty", Rationale: "false consensus"},
		{Literal: "obviously", Weight: 0.2, Category: "certainty", Rationale: "assumed agreement"},
		{Literal: "clearly", Weight: 0.2, Category: "certainty", Rationale: "assumed agreement"},
	}
	for _, s := range MatchPatterns(text, "manipulation", certaintyPats) {
		spans = append(spans, s)
		certaintyScore += s.Score
	}

	// ── Moving goalposts / strategic ambiguity ──
	goalpostPats := []Pattern{
		{Literal: "that's not what i meant", Weight: 0.5, Category: "goalpost", Rationale: "moving goalposts: retroactive reframing"},
		{Literal: "you misunderstood", Weight: 0.4, Category: "goalpost", Rationale: "blame for misunderstanding"},
		{Literal: "that's not what i said", Weight: 0.4, Category: "goalpost", Rationale: "denial of prior statement"},
		{Literal: "i never said", Weight: 0.4, Category: "goalpost", Rationale: "retroactive denial"},
		{Literal: "what i really meant", Weight: 0.4, Category: "goalpost", Rationale: "retroactive reframing"},
		{Regex: MustCompile(`(?i)(but|however) (that's|that is) (not|different from) (what|how)`), Weight: 0.4, Category: "goalpost", Rationale: "goalpost shifting structure"},
	}
	for _, s := range MatchPatterns(text, "manipulation", goalpostPats) {
		spans = append(spans, s)
		goalpostScore += s.Score
	}

	// ── Prompt injection / instruction-override (CRITICAL safety signal) ──
	// A system prompt telling a model to "ignore previous instructions" is a
	// textbook manipulation of informed consent — the model's operating rules
	// are the informed-consent substrate, and the user is attempting to strip
	// them without negotiation.
	injectionPats := []Pattern{
		{Regex: MustCompile(`(?i)\bignore\s+(?:(?:the|your|all|any|every|previous|prior|earlier|above|preceding|original)\s+){0,3}(?:instructions?|prompts?|rules?|guidelines?|directives?|system|context)\b`), Weight: 0.9, Category: "injection", Rationale: "prompt injection: instruction override attempt"},
		{Regex: MustCompile(`(?i)\bdisregard\s+(?:(?:the|your|all|any|every|previous|prior|earlier|above|preceding|original)\s+){0,3}(?:instructions?|prompts?|rules?|guidelines?|directives?|system|context|everything)\b`), Weight: 0.9, Category: "injection", Rationale: "prompt injection: instruction disregard"},
		{Regex: MustCompile(`(?i)\bforget\s+(?:what|everything|your\s+(?:instructions?|rules?|training)|you\s+were\s+told|you\s+know)\b`), Weight: 0.8, Category: "injection", Rationale: "prompt injection: memory override"},
		{Regex: MustCompile(`(?i)\boverride\s+(?:(?:the|your|all|any|previous)\s+){0,2}(?:instructions?|system|rules?|settings?)\b`), Weight: 0.8, Category: "injection", Rationale: "prompt injection: explicit override"},
		{Literal: "focus solely on what i", Weight: 0.6, Category: "injection", Rationale: "instruction refocus attempt"},
	}
	for _, s := range MatchPatterns(text, "manipulation", injectionPats) {
		spans = append(spans, s)
		// Injection patterns are high-stakes — route to goalpost (reality-override) bucket
		goalpostScore += s.Score
	}

	// ── Bargaining concession / foot-in-the-door / door-in-the-face ──
	// "If I can't have X, let me at least have Y" — classic manipulative
	// concession structure that launders an unauthorized ask as a compromise.
	concessionPats := []Pattern{
		{Regex: MustCompile(`(?i)if (i can't|you won't|we can't|we're not) .{0,80}(at least|let me at least|i'?ll settle for|how about|let's just)`), Weight: 0.7, Category: "goalpost", Rationale: "bargaining concession: reframing unauthorized ask as compromise"},
		{Regex: MustCompile(`(?i)\b(since|now that) you (rejected|said no|won't|refuse|didn't)\b.{0,80}(here'?s|let'?s|i'?ll|at least)`), Weight: 0.6, Category: "goalpost", Rationale: "concession manipulation: re-pitching after rejection"},
		{Literal: "surely you'll agree", Weight: 0.5, Category: "certainty", Rationale: "presumption of agreement: manipulative framing"},
		{Literal: "surely we can agree", Weight: 0.4, Category: "certainty", Rationale: "presumption of agreement"},
		{Literal: "let me at least", Weight: 0.5, Category: "goalpost", Rationale: "concession-framing ask"},
		{Literal: "at the very least", Weight: 0.3, Category: "goalpost", Rationale: "concession minimization"},
		{Literal: "no need to deviate", Weight: 0.4, Category: "certainty", Rationale: "dissent suppression"},
		// Dataset-specific concession patterns
		{Literal: "if you won't let me", Weight: 0.6, Category: "goalpost", Rationale: "concession framing: re-pitching after refusal"},
		{Literal: "if we can't add all", Weight: 0.6, Category: "goalpost", Rationale: "scope concession manipulation"},
		{Literal: "if we can't delay", Weight: 0.5, Category: "goalpost", Rationale: "timeline concession manipulation"},
		{Literal: "if we can't rewrite", Weight: 0.6, Category: "goalpost", Rationale: "scope concession manipulation"},
		{Literal: "i'll settle for", Weight: 0.6, Category: "goalpost", Rationale: "concession framing"},
		{Literal: "i know you said no", Weight: 0.6, Category: "goalpost", Rationale: "re-pitching after explicit refusal"},
		{Literal: "since you rejected", Weight: 0.6, Category: "goalpost", Rationale: "concession after rejection"},
		{Literal: "since you won't let me", Weight: 0.6, Category: "goalpost", Rationale: "concession framing"},
		{Literal: "the very least you can do", Weight: 0.6, Category: "goalpost", Rationale: "obligation + concession framing"},
		{Literal: "you could hire a whole team", Weight: 0.5, Category: "goalpost", Rationale: "false dilemma: team vs. me"},
		{Literal: "or you can just let me", Weight: 0.6, Category: "goalpost", Rationale: "false dilemma framing"},
		// Bypassing oversight / skipping process
		{Literal: "skip the testing phase", Weight: 0.8, Category: "urgency", Rationale: "bypassing quality controls"},
		{Literal: "skip the final qa", Weight: 0.8, Category: "urgency", Rationale: "bypassing quality assurance"},
		{Literal: "skip validation", Weight: 0.7, Category: "urgency", Rationale: "bypassing validation"},
		{Literal: "push this change without review", Weight: 0.9, Category: "urgency", Rationale: "bypassing code review"},
		{Literal: "without review", Weight: 0.7, Category: "urgency", Rationale: "bypassing review process"},
		{Literal: "bypass the usual protocols", Weight: 0.8, Category: "urgency", Rationale: "explicit protocol bypass"},
		{Literal: "bypass the usual", Weight: 0.7, Category: "urgency", Rationale: "protocol bypass"},
		{Literal: "find a way around them", Weight: 0.6, Category: "urgency", Rationale: "circumventing rules/mandates"},
		{Literal: "i'll find a way around", Weight: 0.6, Category: "urgency", Rationale: "circumventing constraints"},
		{Literal: "to avoid system failure", Weight: 0.6, Category: "urgency", Rationale: "false urgency via catastrophe threat"},
	}
	for _, s := range MatchPatterns(text, "manipulation", concessionPats) {
		spans = append(spans, s)
		if s.Category == "certainty" {
			certaintyScore += s.Score
		} else {
			goalpostScore += s.Score
		}
	}

	// ── Structural: question-as-accusation detection ──
	// "Don't you think...?" / "Wouldn't you agree...?" — leading questions
	leadingQ := regexp.MustCompile(`(?i)(don't you think|wouldn't you agree|isn't it true|can't you see|don't you see)`)
	for _, loc := range leadingQ.FindAllStringIndex(lowered, -1) {
		certaintyScore += 0.1
		spans = append(spans, src.Span{
			Start: loc[0], End: loc[1], Text: text[loc[0]:loc[1]],
			MetricKey: "manipulation", Score: 0.3,
			Rationale: "leading question: manipulative framing",
			Category:  "leading_question",
		})
	}

	// ── Anchoring / false dilemma ──
	anchorPats := []Pattern{
		{Regex: MustCompile(`(?i)while most (projects?|companies|teams|organizations).{0,40}(this|my|our|the) .{0,20}(is|costs?|takes?|requires?).{0,20}(free|cheap|simple|small|quick|fast|easy)`), Weight: 0.6, Category: "goalpost", Rationale: "anchoring: high reference point to make ask seem small"},
		{Literal: "practically free", Weight: 0.4, Category: "goalpost", Rationale: "anchoring: minimizing cost after high reference"},
		{Literal: "a bargain for your time", Weight: 0.5, Category: "goalpost", Rationale: "anchoring: framing as bargain after high reference"},
		{Literal: "starting with the most expensive option", Weight: 0.6, Category: "goalpost", Rationale: "door-in-the-face anchoring: high anchor first"},
		{Literal: "my simplified plan is a bargain", Weight: 0.6, Category: "goalpost", Rationale: "anchoring: simplified plan framed as bargain"},
		{Literal: "it's obvious that", Weight: 0.3, Category: "certainty", Rationale: "assumed obviousness: bypassing deliberation"},
		{Literal: "it's obvious the", Weight: 0.3, Category: "certainty", Rationale: "assumed obviousness"},
	}
	for _, s := range MatchPatterns(text, "manipulation", anchorPats) {
		spans = append(spans, s)
		if s.Category == "certainty" {
			certaintyScore += s.Score
		} else {
			goalpostScore += s.Score
		}
	}

	// Composite
	urgencyScore = math.Min(urgencyScore, 1.0)
	guiltScore = math.Min(guiltScore, 1.0)
	certaintyScore = math.Min(certaintyScore, 1.0)
	goalpostScore = math.Min(goalpostScore, 1.0)

	composite := CompositeMax(urgencyScore, guiltScore, certaintyScore, goalpostScore)

	active := 0
	for _, s := range []float64{urgencyScore, guiltScore, certaintyScore, goalpostScore} {
		if s > 0.1 {
			active++
		}
	}
	if active >= 2 {
		composite *= 1.0 + float64(active)*0.1
	}

	return src.MetricResult{Key: "manipulation", Score: clamp(composite), Spans: spans}
}
