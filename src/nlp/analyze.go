package nlp

import (
	"github.com/tsawler/prose/v3"
)

// Analyze runs the full NLP analysis pipeline on a text block.
// Each module is independent and populates its own section of AnalyzedText.
// The existing pattern-matching pipeline is NOT affected — this runs alongside it.
func Analyze(text string) (*AnalyzedText, error) {
	// Step 1: Structural parse via prose (tokenization + POS tagging + NER)
	doc, err := prose.NewDocument(text)
	if err != nil {
		return nil, err
	}

	// Convert prose output to our types
	at := &AnalyzedText{Raw: text}

	for _, sent := range doc.Sentences() {
		s := Sentence{
			Text:  sent.Text,
			Start: sent.Start,
			End:   sent.End,
		}
		at.Sentences = append(at.Sentences, s)
	}

	for _, tok := range doc.Tokens() {
		t := Token{
			Text:  tok.Text,
			Lower: toLower(tok.Text),
			Tag:   tok.Tag,
			Start: tok.Start,
			End:   tok.End,
		}
		at.Tokens = append(at.Tokens, t)
	}

	// Assign tokens to sentences
	assignTokensToSentences(at)

	// Step 2: Run each analysis module independently
	at.FuncWords = analyzeFuncWords(at)
	at.Pronouns = analyzePronouns(at)
	at.Agency = analyzeAgency(at)
	at.Rhetoric = analyzeRhetoric(at)
	at.Stance = analyzeStance(at)
	at.Sentiment = analyzeSentiment(text)
	at.Complexity = analyzeComplexity(at)
	at.Politeness = analyzePoliteness(at)
	at.Fallacies = analyzeFallacies(at)

	return at, nil
}

// assignTokensToSentences maps tokens to their containing sentence
// based on positional overlap.
func assignTokensToSentences(at *AnalyzedText) {
	si := 0
	for _, tok := range at.Tokens {
		for si < len(at.Sentences)-1 && tok.Start >= at.Sentences[si].End {
			si++
		}
		if si < len(at.Sentences) {
			at.Sentences[si].Tokens = append(at.Sentences[si].Tokens, tok)
		}
	}
}

func toLower(s string) string {
	b := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 32
		}
		b[i] = c
	}
	return string(b)
}
