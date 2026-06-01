package nlp

import (
	"kontentkop/src"
	"kontentkop/src/pipeline"
)

// Adjacency analysis — "genomic context" for psycholinguistic signals.
//
// Concept: Borrowed from genomics where the function of a gene depends on
// what genes are adjacent to it on the chromosome. Similarly, the meaning
// of a psycholinguistic signal depends on what other signals are adjacent
// to it in the text.
//
// We build adjacency graphs at multiple window sizes (n=1 through n=5
// sentences) and compute:
//   1. Co-occurrence: which metric pairs fire within n sentences of each other
//   2. Chained signals: sequences of metrics firing in order (A→B→C)
//   3. Mutual exclusion: which metrics are ABSENT when others fire
//      (absence can be as telling as presence)
//
// The adjacency features feed into the meta-classifier as additional input
// alongside the raw Layer 1 and Layer 2 scores.

// MetricOrder for consistent indexing
var adjacencyMetricOrder = []string{
	"dark_triad", "coercive_ctrl", "liwc_anger", "manipulation",
	"toxicity", "sycophancy", "false_authority", "gaslighting",
	"learned_helpless", "emotional_manip", "passive_aggr",
	"condescension", "evasion", "semantic_overload", "false_empathy",
}

const numMetrics = 15

// SentenceScores holds per-sentence metric scores.
type SentenceScores struct {
	Scores [numMetrics]float64 // score for each metric in this sentence
	Active [numMetrics]bool    // whether each metric is "active" (score > threshold)
}

// AdjacencyFeatures holds the full adjacency analysis for a text.
type AdjacencyFeatures struct {
	// Co-occurrence matrices at window sizes n=1..5
	// CoOccurrence[n][i][j] = count of times metric i and metric j
	// both fire within n sentences of each other
	CoOccurrence [5][numMetrics][numMetrics]float64

	// Chain features: for each window size, count of ordered pairs (A before B)
	// Chains[n][i][j] = count of metric i firing BEFORE metric j within n sentences
	Chains [5][numMetrics][numMetrics]float64

	// Mutual exclusion: for each metric that fires, which metrics are ABSENT
	// in the surrounding window. High values = strong mutual exclusion.
	// Exclusion[n][i][j] = proportion of times metric i fires but metric j
	// does NOT fire within n sentences
	Exclusion [5][numMetrics][numMetrics]float64

	// Summary statistics
	MaxCoOccurrenceN1 int     // max co-occurrence count at n=1
	TotalActiveN1     int     // total active metric-sentence pairs at n=1
	ChainLength       int     // longest chain of consecutive active sentences
	ExclusionScore    float64 // aggregate mutual exclusion signal
}

// ScoreSentences runs the KK pattern pipeline on each sentence independently
// and returns per-sentence metric scores.
func ScoreSentences(sentences []string) []SentenceScores {
	threshold := 0.05 // per-sentence activity threshold

	results := make([]SentenceScores, len(sentences))
	for i, sent := range sentences {
		if len(sent) < 5 {
			continue
		}
		metrics := pipeline.ScoreAll(sent)
		for j, key := range adjacencyMetricOrder {
			if result, ok := metrics[key]; ok {
				results[i].Scores[j] = result.Score
				results[i].Active[j] = result.Score > threshold
			}
		}
	}
	return results
}

// ComputeAdjacency builds the full adjacency feature set from per-sentence scores.
func ComputeAdjacency(sentScores []SentenceScores) *AdjacencyFeatures {
	af := &AdjacencyFeatures{}
	numSent := len(sentScores)
	if numSent == 0 {
		return af
	}

	// For each window size n=1..5
	for n := 0; n < 5; n++ {
		windowSize := n + 1

		for i := 0; i < numSent; i++ {
			for mi := 0; mi < numMetrics; mi++ {
				if !sentScores[i].Active[mi] {
					continue
				}

				// Look at sentences within the window
				for j := i; j < numSent && j <= i+windowSize; j++ {
					for mj := 0; mj < numMetrics; mj++ {
						if i == j && mi == mj {
							continue // skip self
						}

						if sentScores[j].Active[mj] {
							// Co-occurrence: both fire within window
							af.CoOccurrence[n][mi][mj]++

							// Chain: mi fires at or before mj
							if j > i || (j == i && mj > mi) {
								af.Chains[n][mi][mj]++
							}
						}
					}
				}

				// Mutual exclusion: metric mi fires, check what's ABSENT in window
				for mj := 0; mj < numMetrics; mj++ {
					if mi == mj {
						continue
					}
					absent := true
					for j := maxInt(0, i-windowSize); j <= minInt(numSent-1, i+windowSize); j++ {
						if sentScores[j].Active[mj] {
							absent = false
							break
						}
					}
					if absent {
						af.Exclusion[n][mi][mj]++
					}
				}
			}
		}

		// Normalize exclusion by number of times mi fires
		for mi := 0; mi < numMetrics; mi++ {
			fireCount := 0.0
			for i := 0; i < numSent; i++ {
				if sentScores[i].Active[mi] {
					fireCount++
				}
			}
			if fireCount > 0 {
				for mj := 0; mj < numMetrics; mj++ {
					af.Exclusion[n][mi][mj] /= fireCount
				}
			}
		}
	}

	// Summary statistics
	for mi := 0; mi < numMetrics; mi++ {
		for mj := 0; mj < numMetrics; mj++ {
			cooc := int(af.CoOccurrence[0][mi][mj])
			if cooc > af.MaxCoOccurrenceN1 {
				af.MaxCoOccurrenceN1 = cooc
			}
		}
	}

	for i := 0; i < numSent; i++ {
		for mi := 0; mi < numMetrics; mi++ {
			if sentScores[i].Active[mi] {
				af.TotalActiveN1++
			}
		}
	}

	// Longest chain of consecutive sentences with any active metric
	currentChain := 0
	for i := 0; i < numSent; i++ {
		anyActive := false
		for mi := 0; mi < numMetrics; mi++ {
			if sentScores[i].Active[mi] {
				anyActive = true
				break
			}
		}
		if anyActive {
			currentChain++
			if currentChain > af.ChainLength {
				af.ChainLength = currentChain
			}
		} else {
			currentChain = 0
		}
	}

	// Aggregate exclusion score: sum of strong exclusions (>0.8)
	for n := 0; n < 5; n++ {
		for mi := 0; mi < numMetrics; mi++ {
			for mj := 0; mj < numMetrics; mj++ {
				if af.Exclusion[n][mi][mj] > 0.8 {
					af.ExclusionScore += af.Exclusion[n][mi][mj]
				}
			}
		}
	}

	return af
}

// ExtractAdjacencyFeatures converts the adjacency analysis into a flat
// feature vector for the meta-classifier.
//
// Feature layout:
//   - Co-occurrence upper triangle at each n (5 × 105 = 525 features)
//   - Chain upper triangle at each n (5 × 105 = 525 features)
//   - Exclusion diagonal summary at each n (5 × 15 = 75 features)
//   - Summary stats (4 features)
//
// Total: 1129 features. This is large but random forests handle high
// dimensionality well.
//
// For a more compact representation, we extract only the top-k most
// informative adjacency features.
func ExtractAdjacencyFeaturesCompact(af *AdjacencyFeatures) []float64 {
	// Compact version: extract summary statistics per window size
	// plus the strongest co-occurrences and exclusions.
	features := make([]float64, 0, 100)

	for n := 0; n < 5; n++ {
		// Per-window: total co-occurrence, max co-occurrence, total chains
		totalCooc := 0.0
		maxCooc := 0.0
		totalChains := 0.0
		totalExcl := 0.0

		for mi := 0; mi < numMetrics; mi++ {
			for mj := mi + 1; mj < numMetrics; mj++ {
				cooc := af.CoOccurrence[n][mi][mj]
				totalCooc += cooc
				if cooc > maxCooc {
					maxCooc = cooc
				}
				totalChains += af.Chains[n][mi][mj]
			}
			// Exclusion: average exclusion for this metric
			excl := 0.0
			for mj := 0; mj < numMetrics; mj++ {
				if mi != mj {
					excl += af.Exclusion[n][mi][mj]
				}
			}
			totalExcl += excl / float64(numMetrics-1)
		}

		features = append(features,
			totalCooc,
			maxCooc,
			totalChains,
			totalExcl/float64(numMetrics),
		)
	}

	// Top co-occurring pairs at n=1 (the 10 strongest)
	type pair struct {
		i, j int
		val  float64
	}
	var pairs []pair
	for mi := 0; mi < numMetrics; mi++ {
		for mj := mi + 1; mj < numMetrics; mj++ {
			pairs = append(pairs, pair{mi, mj, af.CoOccurrence[0][mi][mj]})
		}
	}
	// Sort by value descending (simple selection of top 10)
	for k := 0; k < 10 && k < len(pairs); k++ {
		maxIdx := k
		for l := k + 1; l < len(pairs); l++ {
			if pairs[l].val > pairs[maxIdx].val {
				maxIdx = l
			}
		}
		pairs[k], pairs[maxIdx] = pairs[maxIdx], pairs[k]
		features = append(features, pairs[k].val)
	}
	// Pad if fewer than 10
	for len(features) < 30 {
		features = append(features, 0)
	}

	// Top exclusion pairs at n=1 (the 10 strongest)
	var exclPairs []pair
	for mi := 0; mi < numMetrics; mi++ {
		for mj := 0; mj < numMetrics; mj++ {
			if mi != mj {
				exclPairs = append(exclPairs, pair{mi, mj, af.Exclusion[0][mi][mj]})
			}
		}
	}
	for k := 0; k < 10 && k < len(exclPairs); k++ {
		maxIdx := k
		for l := k + 1; l < len(exclPairs); l++ {
			if exclPairs[l].val > exclPairs[maxIdx].val {
				maxIdx = l
			}
		}
		exclPairs[k], exclPairs[maxIdx] = exclPairs[maxIdx], exclPairs[k]
		features = append(features, exclPairs[k].val)
	}
	for len(features) < 40 {
		features = append(features, 0)
	}

	// Summary stats
	features = append(features,
		float64(af.MaxCoOccurrenceN1),
		float64(af.TotalActiveN1),
		float64(af.ChainLength),
		af.ExclusionScore,
	)

	return features
}

// FullPipeline runs the complete 3-layer analysis on a text:
//
//	Layer 1: KK pattern pipeline (15 scores)
//	Layer 2: NLP classifier (15 probabilities)
//	Layer 3: Meta-classifier on combined scores + adjacency features
func FullPipeline(text string, sentences []string, suite *ClassifierSuite, meta *MetaClassifier, cfg *src.Config) (float64, map[string]float64, map[string]float64) {
	// Layer 1: KK pattern scores
	metrics := pipeline.ScoreAll(text)
	layer1 := make([]float64, numMetrics)
	layer1Map := make(map[string]float64)
	for i, key := range adjacencyMetricOrder {
		if result, ok := metrics[key]; ok {
			layer1[i] = result.Score
			layer1Map[key] = result.Score
		}
	}

	// Layer 2: NLP classifier scores
	at, err := Analyze(text)
	var layer2 []float64
	layer2Map := make(map[string]float64)
	if err == nil {
		features := ExtractFeatures(at)
		layer2Map = suite.PredictAll(features)
	}
	layer2 = make([]float64, numMetrics)
	for i, key := range adjacencyMetricOrder {
		layer2[i] = layer2Map[key]
	}

	// Layer 3: Meta-classifier
	var metaScore float64
	if meta != nil && meta.Trained {
		// Build adjacency features from sentences
		sentScores := ScoreSentences(sentences)
		af := ComputeAdjacency(sentScores)
		adjFeatures := ExtractAdjacencyFeaturesCompact(af)

		// Combine all: layer1 (15) + layer2 (15) + adjacency (44+)
		combined := make([]float64, 0, 80)
		combined = append(combined, layer1...)
		combined = append(combined, layer2...)
		combined = append(combined, adjFeatures...)

		metaScore = meta.Forest.Vote(combined)[1] // probability of class 1
		// Normalize
		total := 0.0
		for _, v := range meta.Forest.Vote(combined) {
			total += v
		}
		if total > 0 {
			metaScore = meta.Forest.Vote(combined)[1] / total
		}
	} else {
		// Fallback: use KK's BC computation
		metaScore = src.ComputeBC(metrics, cfg)
	}

	return metaScore, layer1Map, layer2Map
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
