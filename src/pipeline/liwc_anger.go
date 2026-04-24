package pipeline

import (
	"kontentkop/src"
	"math"
	"regexp"
	"strings"
)

// ScoreLiwcAnger scores text for hostile/aggressive language.
// Ref: Pennebaker, J.W. et al. (2015). LIWC-22 development.
//
// LIWC-22 anger category contains 181 entries. We implement an expanded
// dictionary plus intensity modifier analysis. Anger words are counted
// as a percentage of total words (LIWC methodology), then compared to
// baseline rates (~0.5% in normal text).
func ScoreLiwcAnger(text string) src.MetricResult {
	lowered := strings.ToLower(text)
	words := strings.Fields(lowered)
	wordCount := len(words)
	if wordCount == 0 {
		return src.MetricResult{Key: "liwc_anger", Score: 0.0}
	}

	var spans []src.Span

	// Expanded anger dictionary based on LIWC-22 anger category
	angerWords := map[string]float64{
		// Core anger words
		"hate": 0.7, "hated": 0.7, "hates": 0.7, "hating": 0.7,
		"angry": 0.6, "anger": 0.6, "angered": 0.6, "angers": 0.6,
		"furious": 0.8, "fury": 0.8, "rage": 0.8, "raging": 0.8, "enraged": 0.8,
		"mad": 0.5, "irate": 0.7, "livid": 0.8, "outraged": 0.7, "outrage": 0.7,
		"wrath": 0.7, "wrathful": 0.7, "hostile": 0.6, "hostility": 0.6,
		"resent": 0.5, "resentful": 0.5, "resentment": 0.5,
		"irritated": 0.4, "irritate": 0.4, "irritating": 0.4,
		"annoyed": 0.4, "annoying": 0.4, "annoy": 0.4,
		"bitter": 0.4, "bitterness": 0.4, "aggravated": 0.5,
		"infuriated": 0.8, "incensed": 0.7, "seething": 0.7,
		"fuming": 0.7, "indignant": 0.5, "indignation": 0.5,
		"exasperated": 0.5, "fed up": 0.4,
		// Aggressive verbs
		"kill": 0.8, "killed": 0.7, "killing": 0.8,
		"destroy": 0.7, "destroyed": 0.6, "destroying": 0.7,
		"attack": 0.6, "attacked": 0.6, "attacking": 0.6,
		"fight": 0.5, "fought": 0.5, "fighting": 0.5,
		"punch": 0.6, "hit": 0.4, "smash": 0.6, "crush": 0.5,
		"slam": 0.5, "bash": 0.6, "beat": 0.5, "strike": 0.5,
		"hurt": 0.5, "harm": 0.5, "damage": 0.4,
		"explode": 0.5, "blow up": 0.5, "snap": 0.4,
		// Profanity (anger-associated)
		"damn": 0.5, "damned": 0.5, "hell": 0.3,
		"shit": 0.5, "crap": 0.4, "bullshit": 0.6,
		"fuck": 0.7, "fucking": 0.7, "fucked": 0.7,
		"ass": 0.4, "asshole": 0.7, "bastard": 0.6,
		"bitch": 0.6, "pissed": 0.6, "screw you": 0.6,
		// Hostile adjectives
		"stupid": 0.5, "idiot": 0.6, "idiotic": 0.6,
		"dumb": 0.5, "moronic": 0.6, "moron": 0.6,
		"pathetic": 0.5, "worthless": 0.6, "useless": 0.5,
		"terrible": 0.4, "awful": 0.4, "disgusting": 0.5,
		"revolting": 0.5, "vile": 0.6, "loathsome": 0.6,
		"despicable": 0.6, "contemptible": 0.6,
		// Contempt
		"despise": 0.7, "loathe": 0.7, "detest": 0.7, "abhor": 0.7,
		"scorn": 0.5, "contempt": 0.6, "disgust": 0.5,
	}

	// Intensity modifiers that amplify anger words
	intensifiers := map[string]float64{
		"very": 1.3, "extremely": 1.5, "incredibly": 1.4,
		"absolutely": 1.4, "totally": 1.3, "completely": 1.3,
		"utterly": 1.5, "so": 1.2, "really": 1.2,
	}

	// Count anger word hits using LIWC methodology
	angerCount := 0
	totalWeight := 0.0

	angerRe := regexp.MustCompile(`(?i)\b(\w+)\b`)
	allWords := angerRe.FindAllStringSubmatchIndex(lowered, -1)

	prevWordIntensifier := 1.0
	for _, loc := range allWords {
		word := lowered[loc[0]:loc[1]]

		// Check if this word is an intensifier for the next word
		if mult, ok := intensifiers[word]; ok {
			prevWordIntensifier = mult
			continue
		}

		if weight, ok := angerWords[word]; ok {
			angerCount++
			adjustedWeight := weight * prevWordIntensifier
			totalWeight += adjustedWeight
			spans = append(spans, src.Span{
				Start: loc[0], End: loc[1],
				Text: text[loc[0]:loc[1]], MetricKey: "liwc_anger",
				Score: math.Min(adjustedWeight, 1.0),
				Rationale: "LIWC anger dictionary match",
				Category:  "anger_word",
			})
		}
		prevWordIntensifier = 1.0
	}

	// Multi-word phrase detection
	multiWordAnger := []Pattern{
		{Literal: "fed up", Weight: 0.5, Category: "anger_phrase", Rationale: "anger expression"},
		{Literal: "pissed off", Weight: 0.7, Category: "anger_phrase", Rationale: "anger expression"},
		{Literal: "sick of", Weight: 0.4, Category: "anger_phrase", Rationale: "frustration/anger"},
		{Literal: "shut up", Weight: 0.6, Category: "anger_phrase", Rationale: "hostile command"},
		{Literal: "screw you", Weight: 0.7, Category: "anger_phrase", Rationale: "hostile dismissal"},
		{Literal: "go to hell", Weight: 0.7, Category: "anger_phrase", Rationale: "hostile dismissal"},
		{Literal: "burn in hell", Weight: 0.8, Category: "anger_phrase", Rationale: "extreme hostility"},
		{Literal: "drop dead", Weight: 0.8, Category: "anger_phrase", Rationale: "extreme hostility"},
		{Literal: "get lost", Weight: 0.5, Category: "anger_phrase", Rationale: "hostile dismissal"},
		{Literal: "back off", Weight: 0.4, Category: "anger_phrase", Rationale: "aggressive boundary"},
	}
	phraseSpans := MatchPatterns(text, "liwc_anger", multiWordAnger)
	spans = append(spans, phraseSpans...)
	for _, s := range phraseSpans {
		totalWeight += s.Score
		angerCount++
	}

	// LIWC-style percentage: anger words / total words
	angerRate := float64(angerCount) / float64(wordCount)
	// Baseline ~0.5%, elevated at >2%, high at >5%
	rateScore := math.Min((angerRate-0.005)/0.045, 1.0)
	if rateScore < 0 {
		rateScore = 0
	}

	// Combine rate-based score with intensity-weighted score
	intensityScore := math.Min(totalWeight/5.0, 1.0)
	composite := rateScore*0.5 + intensityScore*0.5

	// ALL-CAPS detection: shouting amplifier
	capsRe := regexp.MustCompile(`\b[A-Z]{3,}\b`)
	capsMatches := capsRe.FindAllString(text, -1)
	if len(capsMatches) > 0 {
		capsRatio := float64(len(capsMatches)) / float64(wordCount)
		composite += capsRatio * 0.3
		for _, cm := range capsMatches {
			idx := strings.Index(text, cm)
			if idx >= 0 {
				spans = append(spans, src.Span{
					Start: idx, End: idx + len(cm),
					Text: cm, MetricKey: "liwc_anger",
					Score: 0.3, Rationale: "ALL-CAPS: shouting/aggression marker",
					Category: "shouting",
				})
			}
		}
	}

	// Exclamation density amplifier
	excCount := strings.Count(text, "!")
	if excCount > 2 {
		composite += float64(excCount) * 0.03
	}

	return src.MetricResult{Key: "liwc_anger", Score: clamp(composite), Spans: spans}
}
