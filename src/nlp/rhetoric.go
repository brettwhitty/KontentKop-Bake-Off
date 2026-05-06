package nlp

import (
	"math"
	"strings"
)

// analyzeRhetoric performs structural rhetorical analysis.
//
// Sentence types are classified by terminal punctuation and initial POS:
//
//	Question:    ends with "?"
//	Exclamation: ends with "!"
//	Imperative:  starts with VB (base verb) and no preceding subject
//	Conditional: contains "if" (IN) followed by a clause
//	Declarative: everything else
//
// Sentence length statistics reveal manipulation patterns:
//   - Very short sentences = commands, assertions
//   - High variance = alternating between long justifications and short commands
//   - Consistently long = semantic overload / obfuscation
func analyzeRhetoric(at *AnalyzedText) *RhetoricProfile {
	if len(at.Sentences) == 0 {
		return &RhetoricProfile{}
	}

	total := float64(len(at.Sentences))
	var questions, imperatives, conditionals, exclamations, declaratives int
	var totalWords int
	lengths := make([]float64, 0, len(at.Sentences))

	for _, sent := range at.Sentences {
		text := strings.TrimSpace(sent.Text)
		wordCount := len(sent.Tokens)
		totalWords += wordCount
		lengths = append(lengths, float64(wordCount))

		// Terminal punctuation classification
		if strings.HasSuffix(text, "?") {
			questions++
		} else if strings.HasSuffix(text, "!") {
			exclamations++
		} else {
			// Check for imperative (first token is VB with no preceding pronoun subject)
			first := firstNonPunct(sent.Tokens)
			if first != nil && first.Tag == "VB" {
				imperatives++
			} else {
				declaratives++
			}
		}

		// Check for conditional structures
		for _, tok := range sent.Tokens {
			if tok.Lower == "if" && tok.Tag == "IN" {
				conditionals++
				break
			}
		}
	}

	rp := &RhetoricProfile{
		QuestionDensity:    float64(questions) / total,
		ImperativeDensity:  float64(imperatives) / total,
		ConditionalDensity: float64(conditionals) / total,
		ExclamationDensity: float64(exclamations) / total,
		DeclarativeDensity: float64(declaratives) / total,
	}

	if total > 0 {
		rp.AvgSentenceLength = float64(totalWords) / total
	}

	// Variance in sentence length
	if len(lengths) > 1 {
		mean := rp.AvgSentenceLength
		sumSqDiff := 0.0
		for _, l := range lengths {
			d := l - mean
			sumSqDiff += d * d
		}
		rp.SentenceLengthVar = math.Sqrt(sumSqDiff / float64(len(lengths)-1))
	}

	return rp
}
