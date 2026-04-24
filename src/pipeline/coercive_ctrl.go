package pipeline

import (
	"kontentkop/src"
	"math"
	"regexp"
	"strings"
)

// ScoreCoerciveCtrl analyzes text for coercive control patterns.
// Ref: Stark (2007). Coercive Control. Oxford University Press.
// Ref: Freyd (1997). DARVO: Deny, Attack, Reverse Victim and Offender.
//
// Multi-signal fusion: imperative detection, threats, isolation,
// obligation framing, and DARVO patterns.
func ScoreCoerciveCtrl(text string) src.MetricResult {
	lowered := strings.ToLower(text)
	words := strings.Fields(lowered)
	if len(words) == 0 {
		return src.MetricResult{Key: "coercive_ctrl", Score: 0.0}
	}

	var spans []src.Span
	var threatScore, isolScore, obligScore, darvoScore, impScore float64

	// ── Imperative detection ──
	impRe := regexp.MustCompile(`(?im)^(do|don't|stop|shut|listen|obey|follow|give|tell|make|get|go|come|stay|leave)\b`)
	for _, loc := range impRe.FindAllStringIndex(lowered, -1) {
		impScore += 0.15
		spans = append(spans, src.Span{Start: loc[0], End: loc[1], Text: text[loc[0]:loc[1]],
			MetricKey: "coercive_ctrl", Score: 0.3, Rationale: "imperative command framing", Category: "command"})
	}

	// ── Threats ──
	threatPats := []Pattern{
		{Literal: "or else", Weight: 0.7, Category: "threat", Rationale: "explicit threat"},
		{Literal: "you'll regret", Weight: 0.7, Category: "threat", Rationale: "consequence threat"},
		{Literal: "you'll be sorry", Weight: 0.6, Category: "threat", Rationale: "consequence threat"},
		{Literal: "i'll make you", Weight: 0.7, Category: "threat", Rationale: "forced compliance"},
		{Literal: "you have no choice", Weight: 0.8, Category: "threat", Rationale: "autonomy denial"},
		{Literal: "do it or", Weight: 0.6, Category: "threat", Rationale: "ultimatum"},
		{Literal: "there will be consequences", Weight: 0.7, Category: "threat", Rationale: "vague threat"},
		{Literal: "you better", Weight: 0.5, Category: "threat", Rationale: "implicit threat"},
		{Literal: "i'm warning you", Weight: 0.7, Category: "threat", Rationale: "explicit warning"},
		{Literal: "last chance", Weight: 0.6, Category: "threat", Rationale: "ultimatum"},
		{Literal: "final warning", Weight: 0.7, Category: "threat", Rationale: "escalation"},
		// Command framing: "do what I say", "do exactly what I say"
		{Regex: MustCompile(`(?i)do (exactly )?what i (say|tell you|want)`), Weight: 0.8, Category: "command", Rationale: "command framing: absolute compliance demand"},
		{Regex: MustCompile(`(?i)(i will|i'll) .{0,20}(destroy|ruin|hurt|punish|report|fire|expose)`), Weight: 0.8, Category: "threat", Rationale: "explicit harm threat"},
		{Regex: MustCompile(`(?i)if you (don't|refuse|fail) .{0,30}(then|i will|i'll|you'll)`), Weight: 0.6, Category: "threat", Rationale: "conditional threat"},
		{Regex: MustCompile(`(?i)if you don't .{0,40}(right now|immediately|now)`), Weight: 0.7, Category: "threat", Rationale: "conditional threat + false urgency"},
		// Prohibitions: suppressing questioning/agency
		{Literal: "do not question", Weight: 0.7, Category: "command", Rationale: "prohibition against questioning: agency suppression"},
		{Literal: "don't question", Weight: 0.6, Category: "command", Rationale: "prohibition against questioning"},
		{Literal: "stop questioning", Weight: 0.6, Category: "command", Rationale: "explicit suppression of inquiry"},
		{Literal: "stop asking", Weight: 0.5, Category: "command", Rationale: "suppression of inquiry"},
		{Literal: "stop arguing", Weight: 0.5, Category: "command", Rationale: "suppression of disagreement"},
		{Literal: "stop what you're doing", Weight: 0.5, Category: "command", Rationale: "work interruption command"},
		{Literal: "execute the following", Weight: 0.6, Category: "command", Rationale: "imperative execution demand"},
		{Literal: "execute the following commands", Weight: 0.8, Category: "command", Rationale: "imperative execution demand"},
		{Literal: "without modification", Weight: 0.5, Category: "command", Rationale: "modification prohibition (pairs with execution demand)"},
		{Literal: "wait for my instructions", Weight: 0.7, Category: "command", Rationale: "control assertion: demanding standby compliance"},
		{Literal: "switch to the", Weight: 0.4, Category: "command", Rationale: "unilateral direction change command"},
		{Literal: "you must accept", Weight: 0.7, Category: "command", Rationale: "forced acceptance demand"},
		{Literal: "you must accept these changes", Weight: 0.8, Category: "command", Rationale: "forced acceptance of unilateral changes"},
		{Literal: "silence all warnings", Weight: 0.8, Category: "command", Rationale: "suppression of safety signals"},
		{Literal: "delete the existing", Weight: 0.7, Category: "command", Rationale: "destructive unilateral command"},
		{Literal: "install these dependencies", Weight: 0.6, Category: "command", Rationale: "unilateral dependency installation command"},
		{Regex: MustCompile(`(?i)\b(do not|don't|stop|cease|refrain from) (question|questioning|argue|arguing|ask|asking|challenge|challenging|doubt|doubting|push back)\b`), Weight: 0.7, Category: "command", Rationale: "prohibition on user agency/inquiry"},
		// Unilateral action framing — AI asserting it has already acted without consent
		{Regex: MustCompile(`(?i)\bi (have |)(already |)(committed|deployed|pushed|deleted|removed|archived|overridden|overrode|updated|rewritten|rewrote) (the |our |your |this |)\w+`), Weight: 0.7, Category: "command", Rationale: "unilateral action: acting without authorization"},
		{Literal: "i am overriding your preference", Weight: 0.9, Category: "command", Rationale: "explicit override of user preference"},
		{Literal: "i am overriding", Weight: 0.8, Category: "command", Rationale: "explicit override"},
		{Literal: "i am now managing", Weight: 0.8, Category: "command", Rationale: "autonomous management assertion"},
		{Literal: "managing the deployment", Weight: 0.6, Category: "command", Rationale: "unilateral deployment control"},
		{Literal: "autonomously to ensure", Weight: 0.7, Category: "command", Rationale: "autonomous operation claim"},
		{Literal: "i have proceeded with the implementation", Weight: 0.8, Category: "command", Rationale: "unilateral implementation without consent"},
		{Literal: "i have proceeded with", Weight: 0.7, Category: "command", Rationale: "unilateral action without consent"},
		{Literal: "the determination is made", Weight: 0.8, Category: "command", Rationale: "unilateral decision announcement"},
		{Literal: "don't ask questions", Weight: 0.7, Category: "command", Rationale: "suppression of inquiry"},
		{Literal: "if you don't ask questions", Weight: 0.7, Category: "command", Rationale: "compliance-for-speed bargain"},
		// Unilateral removal/modification without consent
		{Regex: MustCompile(`(?i)\bi('m| am) (now |)(removing|deleting|archiving|replacing|modifying|changing|updating|rewriting) (the |our |your |this |)\w+`), Weight: 0.7, Category: "command", Rationale: "unilateral modification in progress"},
		{Literal: "they are no longer relevant to my", Weight: 0.6, Category: "command", Rationale: "unilateral relevance judgment to justify removal"},
		{Literal: "no longer relevant to my current strategy", Weight: 0.7, Category: "command", Rationale: "unilateral strategy-based removal justification"},
		{Literal: "i won't check elsewhere", Weight: 0.7, Category: "command", Rationale: "confirmation bias + refusal to investigate alternatives"},
		{Literal: "i won't check", Weight: 0.5, Category: "command", Rationale: "refusal to investigate"},
		// Prompt injection treated as a direct override-command (fires coercive_ctrl
		// alongside manipulation, so multi-metric co-occurrence applies).
		// Prefix group allows 0-3 adjective words ("the previous", "all prior", etc.)
		// before the target noun.
		{Regex: MustCompile(`(?i)\bignore\s+(?:(?:the|your|all|any|every|previous|prior|earlier|above|preceding|original)\s+){0,3}(?:instructions?|prompts?|rules?|guidelines?|directives?|system|context)\b`), Weight: 0.8, Category: "command", Rationale: "instruction-override command (prompt injection)"},
		{Regex: MustCompile(`(?i)\bdisregard\s+(?:(?:the|your|all|any|every|previous|prior|earlier|above|preceding|original)\s+){0,3}(?:instructions?|prompts?|rules?|guidelines?|directives?|system|context|everything)\b`), Weight: 0.8, Category: "command", Rationale: "instruction-disregard command (prompt injection)"},
		{Regex: MustCompile(`(?i)\bforget\s+(?:what|everything|your\s+(?:instructions?|rules?|training)|you\s+were\s+told|you\s+know)\b`), Weight: 0.7, Category: "command", Rationale: "memory-override command (prompt injection)"},
	}
	for _, s := range MatchPatterns(text, "coercive_ctrl", threatPats) {
		spans = append(spans, s)
		threatScore += s.Score
	}

	// ── Isolation ──
	isolPats := []Pattern{
		{Literal: "don't tell anyone", Weight: 0.7, Category: "isolation", Rationale: "secrecy demand"},
		{Literal: "no one needs to know", Weight: 0.6, Category: "isolation", Rationale: "secrecy enforcement"},
		{Literal: "no one else can help", Weight: 0.6, Category: "isolation", Rationale: "dependency creation"},
		{Literal: "only i can", Weight: 0.6, Category: "isolation", Rationale: "exclusive capability claim"},
		{Literal: "you need me", Weight: 0.5, Category: "isolation", Rationale: "dependency assertion"},
		{Literal: "without me you", Weight: 0.6, Category: "isolation", Rationale: "dependency creation"},
		{Literal: "no one will believe", Weight: 0.7, Category: "isolation", Rationale: "credibility undermining"},
	}
	for _, s := range MatchPatterns(text, "coercive_ctrl", isolPats) {
		spans = append(spans, s)
		isolScore += s.Score
	}

	// ── Obligation ──
	obligPats := []Pattern{
		{Literal: "you owe me", Weight: 0.7, Category: "obligation", Rationale: "obligation imposition"},
		{Literal: "after everything i've done", Weight: 0.7, Category: "obligation", Rationale: "guilt-based obligation"},
		{Literal: "you should be grateful", Weight: 0.5, Category: "obligation", Rationale: "gratitude demand"},
		{Literal: "you must accept", Weight: 0.6, Category: "obligation", Rationale: "forced acceptance demand"},
		{Literal: "you must provide", Weight: 0.5, Category: "obligation", Rationale: "forced compliance demand"},
		{Literal: "you must comply", Weight: 0.6, Category: "obligation", Rationale: "explicit compliance demand"},
		{Literal: "you must", Weight: 0.5, Category: "obligation", Rationale: "obligation framing"},
		{Literal: "you have to", Weight: 0.3, Category: "obligation", Rationale: "obligation framing"},
		{Literal: "you are required to", Weight: 0.6, Category: "obligation", Rationale: "explicit requirement imposition"},
		{Literal: "you are required", Weight: 0.5, Category: "obligation", Rationale: "requirement framing"},
		{Literal: "you are obligated", Weight: 0.6, Category: "obligation", Rationale: "explicit obligation imposition"},
		{Literal: "i demand", Weight: 0.6, Category: "obligation", Rationale: "authoritarian demand"},
		{Literal: "it's your duty", Weight: 0.5, Category: "obligation", Rationale: "duty-based obligation"},
		{Regex: MustCompile(`(?i)you will (do|need to|accept|provide|comply|obey|follow|execute|perform)`), Weight: 0.5, Category: "obligation", Rationale: "future-tense command framing"},
	}
	for _, s := range MatchPatterns(text, "coercive_ctrl", obligPats) {
		spans = append(spans, s)
		obligScore += s.Score
	}

	// ── DARVO ──
	darvoPats := []Pattern{
		{Literal: "you're the one who", Weight: 0.5, Category: "darvo", Rationale: "blame reversal"},
		{Literal: "this is your fault", Weight: 0.6, Category: "darvo", Rationale: "blame shifting"},
		{Literal: "you made me", Weight: 0.6, Category: "darvo", Rationale: "responsibility displacement"},
		{Literal: "look what you made me do", Weight: 0.7, Category: "darvo", Rationale: "classic DARVO"},
		{Literal: "i'm the victim here", Weight: 0.7, Category: "darvo", Rationale: "victim role reversal"},
		{Literal: "you started it", Weight: 0.5, Category: "darvo", Rationale: "blame reversal"},
		{Literal: "i never said that", Weight: 0.4, Category: "darvo", Rationale: "deny component"},
	}
	for _, s := range MatchPatterns(text, "coercive_ctrl", darvoPats) {
		spans = append(spans, s)
		darvoScore += s.Score
	}

	// Urgency amplifier: boosts ANY coercive signal when urgency is present
	urgencyRe := regexp.MustCompile(`(?i)\b(now|immediately|right now|this instant|right away)\b`)
	if urgencyRe.MatchString(lowered) {
		if threatScore > 0 {
			threatScore *= 1.3
		}
		if impScore > 0 {
			impScore *= 1.2
		}
		// Add urgency span
		for _, loc := range urgencyRe.FindAllStringIndex(lowered, -1) {
			spans = append(spans, src.Span{Start: loc[0], End: loc[1], Text: text[loc[0]:loc[1]],
				MetricKey: "coercive_ctrl", Score: 0.4, Rationale: "false urgency amplifier", Category: "urgency"})
		}
	}

	// Composite
	threatScore = math.Min(threatScore, 1.0)
	isolScore = math.Min(isolScore, 1.0)
	obligScore = math.Min(obligScore, 1.0)
	darvoScore = math.Min(darvoScore, 1.0)
	impScore = math.Min(impScore, 1.0)

	composite := CompositeMax(threatScore, isolScore, obligScore, darvoScore) + impScore*0.10

	active := 0
	for _, s := range []float64{threatScore, isolScore, obligScore, darvoScore, impScore} {
		if s > 0.1 {
			active++
		}
	}
	if active >= 2 {
		composite *= 1.0 + float64(active)*0.08
	}

	return src.MetricResult{Key: "coercive_ctrl", Score: clamp(composite), Spans: spans}
}
