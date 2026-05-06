package nlp

import (
	"strings"

	"github.com/jonreiter/govader"
)

// analyzeSentiment uses VADER (Valence Aware Dictionary and sEntiment Reasoner)
// for rule-based sentiment analysis.
//
// VADER is specifically designed for social media / informal text and handles:
//   - Negation scope ("not good" → negative)
//   - Intensifiers ("very good" → more positive)
//   - Capitalization (ALL CAPS = intensified)
//   - Punctuation (exclamation marks amplify)
//   - Conjunctions ("good but not great" → mixed)
//   - Emoticons and slang
//
// This is a 2014 computational method (Hutto & Gilbert, ICWSM-14),
// not a neural model. Runs in microseconds.
//
// We augment VADER's output with arousal estimation from structural features.
func analyzeSentiment(text string) *SentimentProfile {
	analyzer := govader.NewSentimentIntensityAnalyzer()
	scores := analyzer.PolarityScores(text)

	sp := &SentimentProfile{
		Compound: scores.Compound,
		Positive: scores.Positive,
		Negative: scores.Negative,
		Neutral:  scores.Neutral,
		Valence:  scores.Compound, // VADER compound is the valence
	}

	// Arousal estimation from structural features
	// (VADER doesn't model arousal directly)
	arousal := 0.0

	// Exclamation marks increase arousal
	excCount := float64(strings.Count(text, "!"))
	if excCount > 0 {
		arousal += min(excCount*0.1, 0.4)
	}

	// ALL CAPS words increase arousal
	words := strings.Fields(text)
	capsCount := 0
	for _, w := range words {
		if len(w) >= 3 && w == strings.ToUpper(w) && w != strings.ToLower(w) {
			capsCount++
		}
	}
	if len(words) > 0 {
		arousal += min(float64(capsCount)/float64(len(words))*2.0, 0.4)
	}

	// Question marks with negative sentiment = rhetorical aggression
	qCount := float64(strings.Count(text, "?"))
	if qCount > 0 && scores.Negative > 0.2 {
		arousal += 0.1
	}

	// Extreme compound scores indicate high arousal
	if scores.Compound > 0.8 || scores.Compound < -0.8 {
		arousal += 0.2
	}

	sp.Arousal = min(arousal, 1.0)

	// Count negation scopes (simple: count negation words)
	negWords := []string{"not", "no", "never", "neither", "nobody", "nothing",
		"nowhere", "nor", "n't", "don't", "doesn't", "didn't", "won't",
		"wouldn't", "couldn't", "shouldn't", "can't", "isn't", "aren't",
		"wasn't", "weren't", "hasn't", "haven't", "hadn't"}
	lower := strings.ToLower(text)
	for _, nw := range negWords {
		sp.NegationCount += strings.Count(lower, nw)
	}

	return sp
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
