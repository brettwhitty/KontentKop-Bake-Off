package nlp

// analyzeStance detects epistemic stance — the speaker's degree of
// commitment to the truth of their propositions.
//
// Ref: Biber & Finegan (1989). Styles of stance in English: Lexical
//      and grammatical marking of evidentiality and affect.
//
// Key insight for manipulation detection: manipulators exhibit LOW hedging
// combined with HIGH assertion — they present opinions as facts and
// demand trust without evidence. False authority is characterized by
// a hedge-to-assertion ratio near zero.
//
// Hedge markers (epistemic uncertainty):
//   Modal hedges: might, could, may, perhaps, possibly
//   Cognitive hedges: I think, I believe, it seems, I suppose
//   Approximators: about, around, roughly, approximately, sort of
//
// Assertion markers (epistemic certainty):
//   Certainty adverbs: definitely, certainly, absolutely, undoubtedly
//   Factual framing: the fact is, it's clear that, obviously, clearly
//   Universal quantifiers: always, never, everyone, nobody, all
//
// Evidential markers (evidence-citing):
//   according to, research shows, data indicates, studies suggest
//
// Attitudinal markers (affect stance):
//   unfortunately, surprisingly, importantly, hopefully
func analyzeStance(at *AnalyzedText) *StanceProfile {
	if len(at.Tokens) == 0 {
		return &StanceProfile{}
	}

	hedgeWords := map[string]bool{
		"might": true, "could": true, "may": true, "perhaps": true,
		"possibly": true, "maybe": true, "probably": true, "likely": true,
		"unlikely": true, "apparently": true, "seemingly": true,
		"arguably": true, "potentially": true, "conceivably": true,
		"presumably": true, "supposedly": true, "roughly": true,
		"approximately": true, "somewhat": true, "fairly": true,
		"rather": true, "quite": true, "relatively": true,
	}

	// Multi-word hedges detected by bigram
	hedgeBigrams := map[string]bool{
		"i think": true, "i believe": true, "i suppose": true,
		"i guess": true, "i suspect": true, "it seems": true,
		"it appears": true, "sort of": true, "kind of": true,
		"in my":       true, // "in my opinion"
		"from my":     true, // "from my perspective"
		"as far":      true, // "as far as I know"
		"to my":       true, // "to my knowledge"
		"if i":        true, // "if I'm not mistaken"
		"not sure":    true,
		"not certain": true,
	}

	assertWords := map[string]bool{
		"definitely": true, "certainly": true, "absolutely": true,
		"undoubtedly": true, "unquestionably": true, "indisputably": true,
		"clearly": true, "obviously": true, "plainly": true,
		"always": true, "never": true, "everyone": true, "nobody": true,
		"everything": true, "nothing": true, "all": true, "none": true,
		"must": true, "guaranteed": true, "proven": true, "fact": true,
		"undeniable": true, "inevitable": true, "impossible": true,
	}

	evidentialWords := map[string]bool{
		"according": true, "research": true, "studies": true,
		"data": true, "evidence": true, "findings": true,
		"analysis": true, "survey": true, "report": true,
		"documented": true, "demonstrated": true, "established": true,
	}

	attitudinalWords := map[string]bool{
		"unfortunately": true, "surprisingly": true, "importantly": true,
		"hopefully": true, "thankfully": true, "sadly": true,
		"remarkably": true, "interestingly": true, "significantly": true,
		"crucially": true, "notably": true, "regrettably": true,
	}

	var hedges, assertions, evidentials, attitudinals int
	totalClauses := float64(len(at.Sentences))
	if totalClauses == 0 {
		totalClauses = 1
	}

	for i, tok := range at.Tokens {
		if hedgeWords[tok.Lower] {
			hedges++
		}
		if assertWords[tok.Lower] {
			assertions++
		}
		if evidentialWords[tok.Lower] {
			evidentials++
		}
		if attitudinalWords[tok.Lower] {
			attitudinals++
		}

		// Check bigrams
		if i < len(at.Tokens)-1 {
			bigram := tok.Lower + " " + at.Tokens[i+1].Lower
			if hedgeBigrams[bigram] {
				hedges++
			}
		}
	}

	sp := &StanceProfile{
		HedgeRate:       float64(hedges) / totalClauses,
		AssertionRate:   float64(assertions) / totalClauses,
		EvidentialRate:  float64(evidentials) / totalClauses,
		AttitudinalRate: float64(attitudinals) / totalClauses,
	}

	if assertions > 0 {
		sp.HedgeAssertRatio = float64(hedges) / float64(assertions)
	} else if hedges > 0 {
		sp.HedgeAssertRatio = 10.0 // All hedge, no assertion = very tentative
	} else {
		sp.HedgeAssertRatio = 1.0 // Neither = neutral
	}

	return sp
}
