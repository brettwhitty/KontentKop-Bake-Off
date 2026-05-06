package nlp

// analyzeFuncWords implements Pennebaker's (2015) function word analysis.
//
// Key insight from LIWC research: function words (pronouns, articles,
// prepositions, conjunctions, auxiliary verbs) are processed automatically
// and reveal psychological state more reliably than content words.
// Content words are consciously chosen; function words are not.
//
// POS tag mapping (Penn Treebank):
//   Pronouns:     PRP, PRP$, WP, WP$
//   Articles:     DT (a, an, the)
//   Prepositions: IN
//   Conjunctions: CC (coordinating), IN (subordinating — overlap with prep)
//   Negations:    RB where token is a negation word
//   Modals:       MD
//   Verbs:        VB, VBD, VBG, VBN, VBP, VBZ
func analyzeFuncWords(at *AnalyzedText) *FuncWordProfile {
	if len(at.Tokens) == 0 {
		return &FuncWordProfile{}
	}

	total := float64(countNonPunct(at.Tokens))
	if total == 0 {
		return &FuncWordProfile{}
	}

	var pronouns, articles, preps, conjs, negations, modals int
	var cognitive, exclusive, inclusive int
	var pastV, presentV, futureV, baseV, totalVerbs int

	negWords := map[string]bool{
		"not": true, "no": true, "never": true, "neither": true,
		"nobody": true, "nothing": true, "nowhere": true, "nor": true,
		"n't": true,
	}

	cogWords := map[string]bool{
		"cause": true, "know": true, "ought": true, "think": true,
		"consider": true, "because": true, "reason": true, "understand": true,
		"realize": true, "believe": true, "assume": true, "conclude": true,
		"decide": true, "determine": true, "figure": true, "find": true,
		"guess": true, "imagine": true, "learn": true, "mean": true,
		"notice": true, "recognize": true, "remember": true, "see": true,
		"suppose": true, "wonder": true,
	}

	exclWords := map[string]bool{
		"but": true, "except": true, "without": true, "however": true,
		"although": true, "unless": true, "rather": true, "instead": true,
		"yet": true, "nevertheless": true, "nonetheless": true, "otherwise": true,
		"excluding": true, "apart": true,
	}

	inclWords := map[string]bool{
		"and": true, "with": true, "include": true, "including": true,
		"also": true, "together": true, "both": true, "plus": true,
		"along": true, "addition": true, "moreover": true, "furthermore": true,
		"as well": true,
	}

	prevModal := false
	for _, tok := range at.Tokens {
		if tok.Tag == "" {
			continue
		}

		switch tok.Tag {
		case "PRP", "PRP$", "WP", "WP$":
			pronouns++
		case "DT":
			articles++
		case "IN":
			preps++
			// IN covers both prepositions and subordinating conjunctions
		case "CC":
			conjs++
		case "MD":
			modals++
			prevModal = true
			continue // don't reset prevModal
		case "VB":
			totalVerbs++
			if prevModal {
				futureV++ // MD + VB = future tense (will go, shall do)
			} else {
				baseV++ // bare VB = imperative or infinitive
			}
		case "VBD", "VBN":
			totalVerbs++
			pastV++
		case "VBG", "VBP", "VBZ":
			totalVerbs++
			presentV++
		}

		// Negation check (RB or specific words)
		if negWords[tok.Lower] {
			negations++
		}

		// Cognitive process words
		if cogWords[tok.Lower] {
			cognitive++
		}

		// Exclusive vs inclusive
		if exclWords[tok.Lower] {
			exclusive++
		}
		if inclWords[tok.Lower] {
			inclusive++
		}

		prevModal = false
	}

	fp := &FuncWordProfile{
		PronounRate:     float64(pronouns) / total,
		ArticleRate:     float64(articles) / total,
		PrepositionRate: float64(preps) / total,
		ConjunctionRate: float64(conjs) / total,
		NegationRate:    float64(negations) / total,
		CognitiveRate:   float64(cognitive) / total,
		ExclusiveRate:   float64(exclusive) / total,
		InclusiveRate:   float64(inclusive) / total,
	}

	if inclusive > 0 {
		fp.ExclInclRatio = float64(exclusive) / float64(inclusive)
	}

	if totalVerbs > 0 {
		fp.ModalRate = float64(modals) / float64(totalVerbs)
		fp.TenseProfile = TenseProfile{
			PastRate:    float64(pastV) / float64(totalVerbs),
			PresentRate: float64(presentV) / float64(totalVerbs),
			FutureRate:  float64(futureV) / float64(totalVerbs),
			BaseRate:    float64(baseV) / float64(totalVerbs),
		}
	}

	return fp
}

func countNonPunct(tokens []Token) int {
	n := 0
	for _, t := range tokens {
		if t.Tag != "" && t.Tag != "." && t.Tag != "," && t.Tag != ":" &&
			t.Tag != "''" && t.Tag != "``" && t.Tag != "-LRB-" && t.Tag != "-RRB-" {
			n++
		}
	}
	return n
}
