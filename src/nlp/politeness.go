package nlp

import (
	"strings"
)

// analyzePoliteness implements Brown & Levinson's (1987) politeness theory.
//
// Ref: Brown, P. & Levinson, S.C. (1987). Politeness: Some universals in
//
//	language usage. Cambridge University Press.
//
// Core concept: Face-Threatening Acts (FTAs). Every communicative act
// potentially threatens the hearer's "face" (self-image). Speakers choose
// strategies to mitigate this threat:
//
//  1. Bald on-record: Direct, unmitigated ("Do X", "You're wrong")
//     → High coercion signal
//  2. Positive politeness: Solidarity, compliments, in-group markers
//     ("We should...", "Great idea, and...")
//     → Can be genuine or sycophantic
//  3. Negative politeness: Hedges, deference, apologies, indirectness
//     ("Would you mind...", "I'm sorry to bother you, but...")
//     → Appropriate respect for autonomy
//  4. Off-record: Hints, irony, rhetorical questions
//     ("It's cold in here" = close the window)
//     → Can be passive-aggressive
//
// For manipulation detection:
//   - High bald-on-record + low mitigation = coercive control
//   - High positive politeness + low substance = sycophancy / false empathy
//   - Zero negative politeness = disregard for hearer's autonomy
//   - High off-record = passive aggression / evasion
func analyzePoliteness(at *AnalyzedText) *PolitenessProfile {
	if len(at.Tokens) == 0 {
		return &PolitenessProfile{}
	}

	var posPol, negPol, baldOR, offRec int
	var ftaCount, mitigatedFTA int
	totalSent := float64(len(at.Sentences))
	if totalSent == 0 {
		totalSent = 1
	}

	for _, sent := range at.Sentences {
		text := strings.ToLower(sent.Text)
		isFTA := false
		isMitigated := false

		// ── Bald on-record markers ──
		// Direct commands, unmitigated assertions
		first := firstNonPunct(sent.Tokens)
		if first != nil && first.Tag == "VB" {
			baldOR++
			isFTA = true
		}
		if containsAny(text, []string{
			"you must", "you need to", "you have to", "you will",
			"do it", "stop it", "shut up", "listen to me",
			"i demand", "i insist", "i order",
		}) {
			baldOR++
			isFTA = true
		}

		// ── Positive politeness markers ──
		// Solidarity, compliments, in-group language
		if containsAny(text, []string{
			"we should", "we could", "let's", "we can",
			"great idea", "good point", "i agree", "you're right",
			"nice work", "well done", "i appreciate",
			"our team", "together", "us",
		}) {
			posPol++
		}

		// ── Negative politeness markers ──
		// Hedges, deference, indirectness, apologies
		if containsAny(text, []string{
			"would you mind", "could you please", "if you don't mind",
			"i was wondering", "i'm sorry to", "excuse me",
			"if it's not too much", "when you get a chance",
			"perhaps you could", "might i suggest",
			"i don't mean to", "with all due respect",
			"if possible", "at your convenience",
			"please", "thank you", "thanks",
		}) {
			negPol++
			if isFTA {
				isMitigated = true
			}
		}

		// ── Off-record markers ──
		// Hints, rhetorical questions, irony signals
		if strings.HasSuffix(strings.TrimSpace(sent.Text), "?") &&
			containsAny(text, []string{
				"don't you think", "wouldn't you say", "isn't it",
				"don't you", "can't you see", "wouldn't it be",
			}) {
			offRec++
		}
		if containsAny(text, []string{
			"just saying", "i'm just", "no offense",
			"not that i", "i mean", "whatever",
		}) {
			offRec++
		}

		if isFTA {
			ftaCount++
			if isMitigated {
				mitigatedFTA++
			}
		}
	}

	pp := &PolitenessProfile{
		PositivePoliteness: float64(posPol) / totalSent,
		NegativePoliteness: float64(negPol) / totalSent,
		BaldOnRecord:       float64(baldOR) / totalSent,
		OffRecord:          float64(offRec) / totalSent,
		FTADensity:         float64(ftaCount) / totalSent,
	}

	if ftaCount > 0 {
		pp.MitigationRatio = float64(mitigatedFTA) / float64(ftaCount)
	}

	return pp
}

func containsAny(text string, phrases []string) bool {
	for _, p := range phrases {
		if strings.Contains(text, p) {
			return true
		}
	}
	return false
}
