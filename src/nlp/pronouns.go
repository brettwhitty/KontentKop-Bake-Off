package nlp

// analyzePronouns implements Pennebaker's pronoun analysis methodology.
//
// Ref: Pennebaker, J.W. (2011). The Secret Life of Pronouns.
// Ref: Paulhus & Williams (2002) — narcissistic individuals use significantly
//      more first-person singular pronouns (I, me, my) relative to first-person
//      plural (we, us, our). Population baseline for 1st-singular is ~11.4%.
//
// The I/You ratio reveals relationship framing:
//   High I/You = self-focused, narcissistic
//   High You/I = other-directed, potentially accusatory or coercive
//
// The We/They ratio reveals in-group vs. out-group framing:
//   High We/They = coalition building, solidarity
//   High They/We = othering, blame externalization
func analyzePronouns(at *AnalyzedText) *PronounProfile {
	if len(at.Tokens) == 0 {
		return &PronounProfile{}
	}

	total := float64(countNonPunct(at.Tokens))
	if total == 0 {
		return &PronounProfile{}
	}

	firstSing := map[string]bool{
		"i": true, "me": true, "my": true, "mine": true, "myself": true,
	}
	firstPlur := map[string]bool{
		"we": true, "us": true, "our": true, "ours": true, "ourselves": true,
	}
	secondPerson := map[string]bool{
		"you": true, "your": true, "yours": true, "yourself": true, "yourselves": true,
	}
	thirdPerson := map[string]bool{
		"he": true, "him": true, "his": true, "himself": true,
		"she": true, "her": true, "hers": true, "herself": true,
		"it": true, "its": true, "itself": true,
		"they": true, "them": true, "their": true, "theirs": true, "themselves": true,
	}

	var fs, fp, sp, tp int

	for _, tok := range at.Tokens {
		l := tok.Lower
		if firstSing[l] {
			fs++
		} else if firstPlur[l] {
			fp++
		} else if secondPerson[l] {
			sp++
		} else if thirdPerson[l] {
			tp++
		}
	}

	pp := &PronounProfile{
		FirstSingularRate: float64(fs) / total,
		FirstPluralRate:   float64(fp) / total,
		SecondPersonRate:  float64(sp) / total,
		ThirdPersonRate:   float64(tp) / total,
	}

	if fp > 0 {
		pp.SingPluralRatio = float64(fs) / float64(fp)
	}
	if sp > 0 {
		pp.IYouRatio = float64(fs) / float64(sp)
	}
	if tp > 0 {
		pp.WeTheyRatio = float64(fp) / float64(tp)
	}

	return pp
}
