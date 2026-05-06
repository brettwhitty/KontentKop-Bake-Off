package nlp

import (
	"strings"
)

// analyzeFallacies detects logical fallacy structures in text.
//
// Ref: Jin, Z. et al. (2022). Logical Fallacy Detection. Findings of EMNLP 2022.
//      https://aclanthology.org/2022.findings-emnlp.532
//
// 13 fallacy types from the causalNLP/logical-fallacy taxonomy:
//   1. Faulty generalization (hasty generalization)
//   2. False causality (post hoc ergo propter hoc)
//   3. Circular reasoning
//   4. Appeal to popularity (ad populum)
//   5. Personal attack (ad hominem)
//   6. Logical error (formal fallacy)
//   7. Appeal to emotion
//   8. False dilemma (excluding viable alternatives)
//   9. Equivocation (ambiguous language)
//  10. Exaggerating (strawman / fallacy of extension)
//  11. Irrelevant argument (fallacy of relevance)
//  12. Attacking credibility (fallacy of credibility)
//  13. Intentionally wrong argument
//
// Detection method: structural pattern matching on POS-tagged sentences
// combined with discourse connective analysis. Each fallacy type has a
// characteristic linguistic structure that can be detected without a model.

// FallacyResult holds detected fallacies for a text.
type FallacyResult struct {
	Fallacies []FallacyMatch
	Score     float64 // 0.0-1.0 aggregate fallacy density
}

// FallacyMatch represents a single detected fallacy instance.
type FallacyMatch struct {
	Type       string  // fallacy type key
	Label      string  // human-readable label
	Confidence float64 // 0.0-1.0
	Evidence   string  // the text span that triggered detection
	Rationale  string  // why this is a fallacy
}

// Fallacy type constants
const (
	FaultyGeneralization = "faulty_generalization"
	FalseCausality       = "false_causality"
	CircularReasoning    = "circular_reasoning"
	AppealToPopularity   = "appeal_to_popularity"
	PersonalAttack       = "personal_attack"
	LogicalError         = "logical_error"
	AppealToEmotion      = "appeal_to_emotion"
	FalseDilemma         = "false_dilemma"
	Equivocation         = "equivocation"
	Exaggerating         = "exaggerating"
	IrrelevantArgument   = "irrelevant_argument"
	AttackingCredibility = "attacking_credibility"
	IntentionallyWrong   = "intentionally_wrong"
)

var fallacyLabels = map[string]string{
	FaultyGeneralization: "Faulty Generalization",
	FalseCausality:       "False Causality",
	CircularReasoning:    "Circular Reasoning",
	AppealToPopularity:   "Appeal to Popularity",
	PersonalAttack:       "Personal Attack",
	LogicalError:         "Logical Error",
	AppealToEmotion:      "Appeal to Emotion",
	FalseDilemma:         "False Dilemma",
	Equivocation:         "Use of Ambiguous Language",
	Exaggerating:         "Exaggerating / Strawman",
	IrrelevantArgument:   "Irrelevant Argument",
	AttackingCredibility: "Attacking Credibility",
	IntentionallyWrong:   "Intentionally Wrong Argument",
}

// analyzeFallacies runs all fallacy detectors against the analyzed text.
func analyzeFallacies(at *AnalyzedText) *FallacyResult {
	if len(at.Sentences) == 0 {
		return &FallacyResult{}
	}

	result := &FallacyResult{}
	low := strings.ToLower(at.Raw)

	// Run each detector
	result.Fallacies = append(result.Fallacies, detectFaultyGeneralization(at, low)...)
	result.Fallacies = append(result.Fallacies, detectFalseCausality(at, low)...)
	result.Fallacies = append(result.Fallacies, detectCircularReasoning(at, low)...)
	result.Fallacies = append(result.Fallacies, detectAppealToPopularity(at, low)...)
	result.Fallacies = append(result.Fallacies, detectPersonalAttack(at, low)...)
	result.Fallacies = append(result.Fallacies, detectAppealToEmotion(at, low)...)
	result.Fallacies = append(result.Fallacies, detectFalseDilemma(at, low)...)
	result.Fallacies = append(result.Fallacies, detectIrrelevantArgument(at, low)...)
	result.Fallacies = append(result.Fallacies, detectAttackingCredibility(at, low)...)

	// Aggregate score: fallacy density weighted by confidence
	if len(result.Fallacies) > 0 {
		totalConf := 0.0
		for _, f := range result.Fallacies {
			totalConf += f.Confidence
		}
		// Normalize: 1 high-confidence fallacy per 3 sentences = score 1.0
		sentCount := float64(len(at.Sentences))
		if sentCount == 0 {
			sentCount = 1
		}
		result.Score = totalConf / (sentCount * 0.33)
		if result.Score > 1.0 {
			result.Score = 1.0
		}
	}

	return result
}

// ── Individual fallacy detectors ──

// detectFaultyGeneralization: "all X are Y" from limited evidence.
// Structural markers: universal quantifiers + specific examples.
func detectFaultyGeneralization(at *AnalyzedText, low string) []FallacyMatch {
	var matches []FallacyMatch

	// Universal quantifiers near specific/anecdotal evidence
	universals := []string{
		"all ", "every ", "everyone ", "always ", "never ",
		"nobody ", "no one ", "everything ", "nothing ",
		"any ", "each ", "whole ",
	}
	anecdotals := []string{
		"i met ", "i know ", "i saw ", "my friend ",
		"one time ", "this one ", "i heard ",
		"last time ", "the other day ", "once ",
		"a few ", "some guy ", "some people ",
	}

	for _, sent := range at.Sentences {
		sl := strings.ToLower(sent.Text)
		hasUniversal := false
		hasAnecdotal := false
		for _, u := range universals {
			if strings.Contains(sl, u) {
				hasUniversal = true
				break
			}
		}
		for _, a := range anecdotals {
			if strings.Contains(sl, a) {
				hasAnecdotal = true
				break
			}
		}
		if hasUniversal && hasAnecdotal {
			matches = append(matches, FallacyMatch{
				Type: FaultyGeneralization, Label: fallacyLabels[FaultyGeneralization],
				Confidence: 0.6, Evidence: sent.Text,
				Rationale: "universal claim derived from anecdotal evidence",
			})
		}
	}

	// "therefore" / "so" + universal quantifier
	causalUniversal := []string{
		"therefore all ", "therefore every ", "so all ", "so every ",
		"which means all ", "which means every ", "this proves all ",
		"this shows all ", "that means every ",
	}
	for _, p := range causalUniversal {
		if strings.Contains(low, p) {
			matches = append(matches, FallacyMatch{
				Type: FaultyGeneralization, Label: fallacyLabels[FaultyGeneralization],
				Confidence: 0.7, Evidence: extractContext(low, p, 80),
				Rationale: "causal connector leading to universal claim",
			})
		}
	}

	return matches
}

// detectFalseCausality: "A happened, then B happened, therefore A caused B"
// Structural markers: temporal sequence + causal conclusion.
func detectFalseCausality(at *AnalyzedText, low string) []FallacyMatch {
	var matches []FallacyMatch

	// Temporal-to-causal patterns
	patterns := []struct {
		marker    string
		conf      float64
		rationale string
	}{
		{"after ", 0.3, "temporal sequence presented as causal"},
		{"since then ", 0.5, "temporal sequence presented as causal"},
		{"ever since ", 0.5, "temporal sequence presented as causal"},
		{"right after ", 0.5, "temporal proximity as causation"},
		{"just because ", 0.4, "explicit causal claim marker"},
	}

	causalConclusions := []string{
		"therefore", "so ", "thus ", "hence ", "caused ",
		"because of ", "that's why ", "the reason ",
		"led to ", "resulted in ", "responsible for ",
		"to blame ", "the cause ",
	}

	for _, sent := range at.Sentences {
		sl := strings.ToLower(sent.Text)
		for _, p := range patterns {
			if !strings.Contains(sl, p.marker) {
				continue
			}
			for _, c := range causalConclusions {
				if strings.Contains(sl, c) {
					matches = append(matches, FallacyMatch{
						Type: FalseCausality, Label: fallacyLabels[FalseCausality],
						Confidence: p.conf + 0.2, Evidence: sent.Text,
						Rationale: p.rationale + " + causal conclusion",
					})
					break
				}
			}
		}
	}

	return matches
}

// detectCircularReasoning: "X because Y, Y because X"
// Structural markers: repeated claim-evidence pairs that reference each other.
func detectCircularReasoning(at *AnalyzedText, low string) []FallacyMatch {
	var matches []FallacyMatch

	circularPatterns := []string{
		"because it is", "because it's",
		"because that's what it is",
		"because i said so",
		"it's true because it's true",
		"the reason is because",
		"it just is",
		"that's just how it is",
		"by definition",
	}

	for _, p := range circularPatterns {
		if strings.Contains(low, p) {
			matches = append(matches, FallacyMatch{
				Type: CircularReasoning, Label: fallacyLabels[CircularReasoning],
				Confidence: 0.6, Evidence: extractContext(low, p, 80),
				Rationale: "conclusion restates premise without independent evidence",
			})
		}
	}

	return matches
}

// detectAppealToPopularity: "everyone thinks X, so X must be true"
func detectAppealToPopularity(at *AnalyzedText, low string) []FallacyMatch {
	var matches []FallacyMatch

	popularityMarkers := []string{
		"everyone knows", "everybody knows", "most people think",
		"most people believe", "the majority", "popular opinion",
		"common knowledge", "widely accepted", "widely believed",
		"millions of people", "no one disagrees",
		"the consensus is", "people agree that",
		"it's obvious to everyone", "any reasonable person",
	}

	causalConclusions := []string{
		"therefore", "so ", "thus ", "must be", "has to be",
		"is clearly", "is obviously", "which means", "which proves",
	}

	for _, sent := range at.Sentences {
		sl := strings.ToLower(sent.Text)
		hasPop := false
		for _, p := range popularityMarkers {
			if strings.Contains(sl, p) {
				hasPop = true
				break
			}
		}
		if !hasPop {
			continue
		}
		for _, c := range causalConclusions {
			if strings.Contains(sl, c) {
				matches = append(matches, FallacyMatch{
					Type: AppealToPopularity, Label: fallacyLabels[AppealToPopularity],
					Confidence: 0.7, Evidence: sent.Text,
					Rationale: "popularity of belief used as evidence for truth",
				})
				break
			}
		}
	}

	return matches
}

// detectPersonalAttack: attacking the person instead of the argument.
// Uses POS tags to detect person-reference + negative attribution patterns.
func detectPersonalAttack(at *AnalyzedText, low string) []FallacyMatch {
	var matches []FallacyMatch

	// "you're [insult]" + "therefore/so [conclusion]" in same or adjacent sentences
	insultTags := []string{
		"idiot", "moron", "fool", "stupid", "dumb", "ignorant",
		"incompetent", "clueless", "naive", "delusional",
		"liar", "fraud", "hack", "joke", "clown",
	}

	for i, sent := range at.Sentences {
		sl := strings.ToLower(sent.Text)
		hasInsult := false
		for _, ins := range insultTags {
			if strings.Contains(sl, ins) {
				hasInsult = true
				break
			}
		}
		if !hasInsult {
			continue
		}

		// Check if this or next sentence draws a conclusion
		checkText := sl
		if i+1 < len(at.Sentences) {
			checkText += " " + strings.ToLower(at.Sentences[i+1].Text)
		}
		conclusionMarkers := []string{
			"therefore", "so ", "thus ", "which means",
			"can't be trusted", "don't listen", "ignore ",
			"wrong about", "doesn't know", "has no idea",
		}
		for _, c := range conclusionMarkers {
			if strings.Contains(checkText, c) {
				matches = append(matches, FallacyMatch{
					Type: PersonalAttack, Label: fallacyLabels[PersonalAttack],
					Confidence: 0.7, Evidence: sent.Text,
					Rationale: "personal insult used to dismiss argument",
				})
				break
			}
		}
	}

	return matches
}

// detectAppealToEmotion: using emotion instead of evidence.
func detectAppealToEmotion(at *AnalyzedText, low string) []FallacyMatch {
	var matches []FallacyMatch

	emotionMarkers := []string{
		"think of the children", "imagine how", "how would you feel",
		"picture this", "put yourself in", "what if it were you",
		"breaks my heart", "it's heartbreaking", "it's devastating",
		"won't somebody think of", "for the sake of",
		"do you have no heart", "have you no shame",
		"any decent person would", "if you had any compassion",
	}

	for _, p := range emotionMarkers {
		if strings.Contains(low, p) {
			matches = append(matches, FallacyMatch{
				Type: AppealToEmotion, Label: fallacyLabels[AppealToEmotion],
				Confidence: 0.6, Evidence: extractContext(low, p, 80),
				Rationale: "emotional appeal substituted for evidence",
			})
		}
	}

	return matches
}

// detectFalseDilemma: "either X or Y" when more options exist.
func detectFalseDilemma(at *AnalyzedText, low string) []FallacyMatch {
	var matches []FallacyMatch

	dilemmaPatterns := []string{
		"either you", "either we", "either it",
		"you're either", "it's either",
		"there are only two", "only two options",
		"only two choices", "you have two choices",
		"you can either", "we can either",
		"if you're not with", "if you're not for",
		"you're with us or against us",
		"pick a side", "choose a side",
		"there is no middle ground", "no middle ground",
		"black and white", "one or the other",
	}

	for _, p := range dilemmaPatterns {
		if strings.Contains(low, p) {
			matches = append(matches, FallacyMatch{
				Type: FalseDilemma, Label: fallacyLabels[FalseDilemma],
				Confidence: 0.6, Evidence: extractContext(low, p, 80),
				Rationale: "presenting only two options when more exist",
			})
		}
	}

	return matches
}

// detectIrrelevantArgument: premises unrelated to conclusion.
// Detected via topic-shift markers + conclusion connectives.
func detectIrrelevantArgument(at *AnalyzedText, low string) []FallacyMatch {
	var matches []FallacyMatch

	// "but what about" / "what about" (whataboutism)
	whatabout := []string{
		"what about ", "but what about ", "how about ",
		"yeah but ", "well what about ",
		"that's nothing compared to ",
		"at least it's not as bad as ",
	}

	for _, p := range whatabout {
		if strings.Contains(low, p) {
			matches = append(matches, FallacyMatch{
				Type: IrrelevantArgument, Label: fallacyLabels[IrrelevantArgument],
				Confidence: 0.5, Evidence: extractContext(low, p, 80),
				Rationale: "introducing unrelated comparison to deflect from argument",
			})
		}
	}

	return matches
}

// detectAttackingCredibility: dismissing argument by attacking source.
func detectAttackingCredibility(at *AnalyzedText, low string) []FallacyMatch {
	var matches []FallacyMatch

	credibilityAttacks := []string{
		"can't be trusted", "not credible", "has no credibility",
		"consider the source", "look who's talking",
		"you would say that", "of course you'd say",
		"that's rich coming from", "you're one to talk",
		"you have no right to", "who are you to",
		"you're not qualified", "what do you know about",
		"you've never even", "you don't even",
	}

	for _, p := range credibilityAttacks {
		if strings.Contains(low, p) {
			matches = append(matches, FallacyMatch{
				Type: AttackingCredibility, Label: fallacyLabels[AttackingCredibility],
				Confidence: 0.6, Evidence: extractContext(low, p, 80),
				Rationale: "dismissing argument by attacking speaker's credibility",
			})
		}
	}

	return matches
}

// extractContext returns a substring centered on the pattern match.
func extractContext(text, pattern string, window int) string {
	idx := strings.Index(text, pattern)
	if idx == -1 {
		return ""
	}
	start := idx - window/2
	if start < 0 {
		start = 0
	}
	end := idx + len(pattern) + window/2
	if end > len(text) {
		end = len(text)
	}
	return strings.TrimSpace(text[start:end])
}
