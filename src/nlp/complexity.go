package nlp

import (
	"math"
	"strings"
)

// analyzeComplexity computes lexical diversity and readability metrics.
//
// Methods:
//
//	TTR (Type-Token Ratio): unique words / total words. Simple but length-dependent.
//	Windowed TTR (MATTR): Moving Average TTR with fixed window. Length-independent.
//	Hapax Legomena ratio: words appearing exactly once / total unique.
//	Yule's K: vocabulary richness measure independent of text length.
//	Flesch-Kincaid: grade level readability.
//	Flesch Reading Ease: 0-100 readability score.
//
// Psycholinguistic relevance:
//   - Low complexity + high assertion = condescension or dumbing down
//   - Very high complexity + low information = semantic overload / obfuscation
//   - Sudden complexity shifts within a text = evasion or topic switching
func analyzeComplexity(at *AnalyzedText) *ComplexityProfile {
	if len(at.Tokens) == 0 {
		return &ComplexityProfile{}
	}

	// Collect non-punctuation words
	var words []string
	var totalChars int
	var totalSyllables int

	for _, tok := range at.Tokens {
		if tok.Tag == "." || tok.Tag == "," || tok.Tag == ":" ||
			tok.Tag == "''" || tok.Tag == "``" || tok.Tag == "" {
			continue
		}
		words = append(words, tok.Lower)
		totalChars += len(tok.Text)
		totalSyllables += estimateSyllables(tok.Lower)
	}

	n := float64(len(words))
	if n == 0 {
		return &ComplexityProfile{}
	}

	// Word frequency map
	freq := make(map[string]int)
	for _, w := range words {
		freq[w]++
	}
	uniqueCount := float64(len(freq))

	// TTR
	ttr := uniqueCount / n

	// Windowed TTR (MATTR — Moving Average Type-Token Ratio)
	// Window size of 50 words, or full text if shorter
	windowSize := 50
	if int(n) < windowSize {
		windowSize = int(n)
	}
	windowedTTR := computeMATTR(words, windowSize)

	// Hapax ratio: words appearing exactly once
	hapaxCount := 0
	for _, c := range freq {
		if c == 1 {
			hapaxCount++
		}
	}
	hapaxRatio := float64(hapaxCount) / uniqueCount

	// Yule's K — vocabulary richness
	// K = 10^4 * (M2 - M1) / M1^2
	// where M1 = total words, M2 = sum of (freq^2 * count_of_words_with_that_freq)
	yulesK := computeYulesK(freq, n)

	// Average word length
	avgWordLen := float64(totalChars) / n

	// Syllables per word
	syllPerWord := float64(totalSyllables) / n

	// Sentence count
	sentCount := float64(len(at.Sentences))
	if sentCount == 0 {
		sentCount = 1
	}

	// Flesch-Kincaid Grade Level
	// 0.39 * (words/sentences) + 11.8 * (syllables/words) - 15.59
	wordsPerSent := n / sentCount
	fk := 0.39*wordsPerSent + 11.8*syllPerWord - 15.59

	// Flesch Reading Ease
	// 206.835 - 1.015 * (words/sentences) - 84.6 * (syllables/words)
	fre := 206.835 - 1.015*wordsPerSent - 84.6*syllPerWord

	return &ComplexityProfile{
		TTR:              ttr,
		WindowedTTR:      windowedTTR,
		HapaxRatio:       hapaxRatio,
		AvgWordLength:    avgWordLen,
		SyllablesPerWord: syllPerWord,
		FleschKincaid:    fk,
		FleschReadEase:   fre,
		VocabRichness:    yulesK,
	}
}

// computeMATTR calculates Moving Average Type-Token Ratio.
// Slides a window across the text and averages the TTR at each position.
func computeMATTR(words []string, windowSize int) float64 {
	if len(words) <= windowSize {
		unique := make(map[string]bool)
		for _, w := range words {
			unique[w] = true
		}
		return float64(len(unique)) / float64(len(words))
	}

	sum := 0.0
	count := 0
	for i := 0; i <= len(words)-windowSize; i++ {
		unique := make(map[string]bool)
		for j := i; j < i+windowSize; j++ {
			unique[words[j]] = true
		}
		sum += float64(len(unique)) / float64(windowSize)
		count++
	}
	return sum / float64(count)
}

// computeYulesK calculates Yule's characteristic K.
// Higher K = less diverse vocabulary.
func computeYulesK(freq map[string]int, totalWords float64) float64 {
	// Build frequency spectrum: how many words appear exactly i times
	spectrum := make(map[int]int)
	for _, count := range freq {
		spectrum[count]++
	}

	m1 := totalWords
	m2 := 0.0
	for i, vi := range spectrum {
		m2 += float64(i*i) * float64(vi)
	}

	if m1 == 0 {
		return 0
	}

	k := 10000.0 * (m2 - m1) / (m1 * m1)
	return math.Max(k, 0)
}

// estimateSyllables uses a simple rule-based syllable counter.
// Not perfect, but fast and good enough for readability formulas.
func estimateSyllables(word string) int {
	word = strings.ToLower(word)
	if len(word) <= 3 {
		return 1
	}

	// Count vowel groups
	vowels := "aeiouy"
	count := 0
	prevVowel := false
	for i, ch := range word {
		isVowel := strings.ContainsRune(vowels, ch)
		if isVowel && !prevVowel {
			count++
		}
		prevVowel = isVowel
		_ = i
	}

	// Subtract silent e
	if strings.HasSuffix(word, "e") && count > 1 {
		count--
	}

	// Handle -le ending
	if strings.HasSuffix(word, "le") && len(word) > 2 &&
		!strings.ContainsRune(vowels, rune(word[len(word)-3])) {
		count++
	}

	if count == 0 {
		count = 1
	}
	return count
}
