package nlp

// analyzeAgency scores power dynamics via POS-based verb transitivity analysis.
//
// Core method: For each clause, identify the grammatical subject and determine
// whether it's acting (agent) or being acted upon (patient). This reveals:
//   - Self-agency: "I decided", "I will make you" — speaker as agent
//   - Other-agency: "You need to", "You must" — listener as agent (obligation)
//   - Self-as-object: "I was told", "They hurt me" — speaker as patient (victim framing)
//   - Other-as-object: "I'll destroy you" — listener as patient (threat)
//
// POS-based heuristic (no dependency parser needed):
//   Subject = pronoun (PRP) immediately preceding a verb
//   Transitive = verb followed by a noun/pronoun object within 3 tokens
//   Imperative = sentence starting with base verb (VB) with no preceding subject
func analyzeAgency(at *AnalyzedText) *AgencyProfile {
	if len(at.Sentences) == 0 {
		return &AgencyProfile{}
	}

	selfSubj := map[string]bool{"i": true}
	selfObj := map[string]bool{"me": true, "myself": true}
	otherSubj := map[string]bool{"you": true}
	otherObj := map[string]bool{"you": true} // "you" is both subject and object in English

	var selfAgent, otherAgent, selfObject, otherObject int
	var transitive, intransitive, imperatives int
	totalSent := float64(len(at.Sentences))

	for _, sent := range at.Sentences {
		toks := sent.Tokens
		if len(toks) == 0 {
			continue
		}

		// Check for imperative: first non-punct token is VB
		firstWord := firstNonPunct(toks)
		if firstWord != nil && firstWord.Tag == "VB" {
			imperatives++
		}

		// Walk tokens looking for subject-verb-object patterns
		for i, tok := range toks {
			if !isVerb(tok.Tag) {
				continue
			}

			// Look backward for subject (pronoun before verb)
			subj := findSubjectBefore(toks, i)
			// Look forward for object (noun/pronoun after verb)
			obj := findObjectAfter(toks, i)

			if obj != nil {
				transitive++
			} else {
				intransitive++
			}

			if subj != nil {
				sl := subj.Lower
				if selfSubj[sl] {
					selfAgent++
					if obj != nil && otherObj[obj.Lower] {
						otherObject++ // "I [verb] you"
					}
				} else if otherSubj[sl] {
					otherAgent++
					if obj != nil && selfObj[obj.Lower] {
						selfObject++ // "You [verb] me"
					}
				}
			}
		}
	}

	totalVerbs := float64(transitive + intransitive)
	ap := &AgencyProfile{
		CommandRate: float64(imperatives) / totalSent,
	}

	if totalVerbs > 0 {
		ap.Transitivity = float64(transitive) / totalVerbs
	}

	totalAgency := float64(selfAgent + otherAgent)
	if totalAgency > 0 {
		ap.SelfAgency = float64(selfAgent) / totalAgency
		ap.OtherAgency = float64(otherAgent) / totalAgency
	}

	totalPatient := float64(selfObject + otherObject)
	if totalPatient > 0 {
		ap.SelfObject = float64(selfObject) / totalPatient
		ap.OtherObject = float64(otherObject) / totalPatient
	}

	return ap
}

func isVerb(tag string) bool {
	switch tag {
	case "VB", "VBD", "VBG", "VBN", "VBP", "VBZ":
		return true
	}
	return false
}

func isPronoun(tag string) bool {
	return tag == "PRP" || tag == "PRP$"
}

func isNounOrPronoun(tag string) bool {
	switch tag {
	case "NN", "NNS", "NNP", "NNPS", "PRP", "PRP$":
		return true
	}
	return false
}

func firstNonPunct(toks []Token) *Token {
	for i := range toks {
		if toks[i].Tag != "" && toks[i].Tag != "." && toks[i].Tag != "," {
			return &toks[i]
		}
	}
	return nil
}

func findSubjectBefore(toks []Token, verbIdx int) *Token {
	// Look up to 3 tokens back for a pronoun
	for i := verbIdx - 1; i >= 0 && i >= verbIdx-3; i-- {
		if isPronoun(toks[i].Tag) {
			return &toks[i]
		}
		// Stop at sentence-internal boundaries
		if toks[i].Tag == "." || toks[i].Tag == "," || toks[i].Tag == "CC" {
			break
		}
	}
	return nil
}

func findObjectAfter(toks []Token, verbIdx int) *Token {
	// Look up to 3 tokens forward for a noun or pronoun
	for i := verbIdx + 1; i < len(toks) && i <= verbIdx+3; i++ {
		if isNounOrPronoun(toks[i].Tag) {
			return &toks[i]
		}
		// Stop at clause boundaries
		if toks[i].Tag == "." || toks[i].Tag == "CC" || toks[i].Tag == "IN" {
			break
		}
	}
	return nil
}
