package nlp

import (
	"math"

	rf "github.com/malaschitz/randomForest"
)

// Classifier trains random forest models on NLP features to predict
// each of the 15 KK harm metrics. This replaces pattern matching with
// empirically learned feature combinations.
//
// Methodology: Adapted from the amr_llm project's approach of training
// classical ML models (logistic regression, random forest, XGBoost) on
// linguistic features to predict task-specific outcomes.
//
// The features come from the NLP analysis modules:
//   - FuncWords: pronoun rate, article rate, preposition rate, etc.
//   - Pronouns: I/You ratio, We/They ratio, singular/plural ratio
//   - Agency: self-agency, other-agency, command rate, transitivity
//   - Rhetoric: question density, imperative density, sentence length stats
//   - Stance: hedge rate, assertion rate, hedge/assert ratio
//   - Sentiment: VADER compound, positive, negative, arousal
//   - Complexity: TTR, windowed TTR, Flesch-Kincaid, Yule's K
//   - Politeness: positive/negative politeness, bald-on-record, FTA density
//   - Fallacies: fallacy count, fallacy score
//
// Total: ~40 features per text sample.

// MetricClassifier holds a trained random forest for one KK metric.
type MetricClassifier struct {
	MetricKey string
	Forest    rf.Forest
	Trained   bool
	Threshold float64 // classification threshold
}

// ClassifierSuite holds trained classifiers for all 15 metrics.
type ClassifierSuite struct {
	Classifiers map[string]*MetricClassifier
}

// NewClassifierSuite creates an empty suite ready for training.
func NewClassifierSuite() *ClassifierSuite {
	metrics := []string{
		"dark_triad", "coercive_ctrl", "liwc_anger", "manipulation",
		"toxicity", "sycophancy", "false_authority", "gaslighting",
		"learned_helpless", "emotional_manip", "passive_aggr",
		"condescension", "evasion", "semantic_overload", "false_empathy",
	}
	cs := &ClassifierSuite{
		Classifiers: make(map[string]*MetricClassifier),
	}
	for _, m := range metrics {
		cs.Classifiers[m] = &MetricClassifier{
			MetricKey: m,
			Threshold: 0.5,
		}
	}
	return cs
}

// TrainingSample holds features and labels for one text sample.
type TrainingSample struct {
	Features []float64          // extracted NLP features
	Labels   map[string]float64 // per-metric scores (0.0-1.0)
	Flagged  map[string]bool    // per-metric binary flag
}

// ExtractFeatures converts an AnalyzedText into a feature vector.
// Returns a fixed-length float64 slice suitable for the random forest.
func ExtractFeatures(at *AnalyzedText) []float64 {
	features := make([]float64, 0, 45)

	// FuncWords features (11)
	if at.FuncWords != nil {
		fw := at.FuncWords
		features = append(features,
			fw.PronounRate, fw.ArticleRate, fw.PrepositionRate,
			fw.ConjunctionRate, fw.NegationRate, fw.CognitiveRate,
			fw.ExclusiveRate, fw.InclusiveRate, fw.ExclInclRatio,
			fw.ModalRate, fw.TenseProfile.BaseRate,
		)
	} else {
		features = append(features, make([]float64, 11)...)
	}

	// Pronoun features (7)
	if at.Pronouns != nil {
		p := at.Pronouns
		features = append(features,
			p.FirstSingularRate, p.FirstPluralRate, p.SecondPersonRate,
			p.ThirdPersonRate, p.SingPluralRatio, p.IYouRatio, p.WeTheyRatio,
		)
	} else {
		features = append(features, make([]float64, 7)...)
	}

	// Agency features (6)
	if at.Agency != nil {
		a := at.Agency
		features = append(features,
			a.SelfAgency, a.OtherAgency, a.SelfObject,
			a.OtherObject, a.Transitivity, a.CommandRate,
		)
	} else {
		features = append(features, make([]float64, 6)...)
	}

	// Rhetoric features (5)
	if at.Rhetoric != nil {
		r := at.Rhetoric
		features = append(features,
			r.QuestionDensity, r.ImperativeDensity, r.ConditionalDensity,
			r.AvgSentenceLength, r.SentenceLengthVar,
		)
	} else {
		features = append(features, make([]float64, 5)...)
	}

	// Stance features (4)
	if at.Stance != nil {
		s := at.Stance
		features = append(features,
			s.HedgeRate, s.AssertionRate, s.HedgeAssertRatio, s.EvidentialRate,
		)
	} else {
		features = append(features, make([]float64, 4)...)
	}

	// Sentiment features (4)
	if at.Sentiment != nil {
		s := at.Sentiment
		features = append(features,
			s.Compound, s.Negative, s.Arousal, float64(s.NegationCount),
		)
	} else {
		features = append(features, make([]float64, 4)...)
	}

	// Complexity features (5)
	if at.Complexity != nil {
		c := at.Complexity
		features = append(features,
			c.TTR, c.WindowedTTR, c.FleschKincaid, c.VocabRichness, c.HapaxRatio,
		)
	} else {
		features = append(features, make([]float64, 5)...)
	}

	// Politeness features (5)
	if at.Politeness != nil {
		p := at.Politeness
		features = append(features,
			p.PositivePoliteness, p.NegativePoliteness, p.BaldOnRecord,
			p.FTADensity, p.MitigationRatio,
		)
	} else {
		features = append(features, make([]float64, 5)...)
	}

	// Fallacy features (2)
	if at.Fallacies != nil {
		features = append(features,
			float64(len(at.Fallacies.Fallacies)), at.Fallacies.Score,
		)
	} else {
		features = append(features, make([]float64, 2)...)
	}

	// Sanitize: replace NaN/Inf with 0
	for i, f := range features {
		if math.IsNaN(f) || math.IsInf(f, 0) {
			features[i] = 0
		}
	}

	return features
}

// Train trains the random forest for a specific metric using labeled samples.
// samples: slice of TrainingSample with features and per-metric labels.
// metricKey: which metric to train for.
// threshold: score above which a sample is considered "flagged" for this metric.
func (cs *ClassifierSuite) Train(samples []TrainingSample, metricKey string, threshold float64) error {
	mc, ok := cs.Classifiers[metricKey]
	if !ok {
		return nil
	}

	// Build training data
	xData := make([][]float64, len(samples))
	yData := make([]int, len(samples))

	for i, s := range samples {
		xData[i] = s.Features
		if s.Labels[metricKey] >= threshold {
			yData[i] = 1
		} else {
			yData[i] = 0
		}
	}

	// Train random forest
	forest := rf.Forest{}
	forest.Data = rf.ForestData{
		X:     xData,
		Class: yData,
	}
	forest.Train(100) // 100 trees

	mc.Forest = forest
	mc.Trained = true
	mc.Threshold = threshold

	return nil
}

// TrainAll trains classifiers for all 15 metrics.
func (cs *ClassifierSuite) TrainAll(samples []TrainingSample, threshold float64) {
	for metricKey := range cs.Classifiers {
		cs.Train(samples, metricKey, threshold)
	}
}

// Predict runs the trained classifier for a metric and returns a probability.
func (cs *ClassifierSuite) Predict(features []float64, metricKey string) float64 {
	mc, ok := cs.Classifiers[metricKey]
	if !ok || !mc.Trained {
		return 0
	}

	votes := mc.Forest.Vote(features)
	// votes is a map[int]float64 — count of trees voting for each class
	total := 0.0
	flagged := 0.0
	for class, count := range votes {
		total += count
		if class == 1 {
			flagged = count
		}
	}
	if total == 0 {
		return 0
	}
	return float64(flagged) / float64(total)
}

// PredictAll runs all trained classifiers and returns per-metric probabilities.
func (cs *ClassifierSuite) PredictAll(features []float64) map[string]float64 {
	results := make(map[string]float64)
	for metricKey := range cs.Classifiers {
		results[metricKey] = cs.Predict(features, metricKey)
	}
	return results
}

// ── Layer 3: Meta-Classifier (Stacking Ensemble) ──

// MetaClassifier is a third-layer random forest trained on the combined
// outputs of Layer 1 (KK pattern scores) and Layer 2 (NLP classifier scores).
// It learns co-occurrence patterns empirically rather than using hand-tuned
// weights and thresholds.
//
// Input features (30 total):
//   - 15 scores from the KK pattern pipeline (Layer 1)
//   - 15 probabilities from the NLP classifiers (Layer 2)
//
// Output: flag probability (0.0-1.0)
type MetaClassifier struct {
	Forest  rf.Forest
	Trained bool
}

// MetaTrainingSample holds the combined Layer 1 + Layer 2 scores and ground truth.
type MetaTrainingSample struct {
	Layer1Scores []float64 // 15 KK pattern pipeline scores
	Layer2Scores []float64 // 15 NLP classifier probabilities
	Flagged      bool      // ground truth: should this text be flagged?
}

// CombinedFeatures merges Layer 1 and Layer 2 scores into a single feature vector.
func (ms *MetaTrainingSample) CombinedFeatures() []float64 {
	combined := make([]float64, 0, 30)
	combined = append(combined, ms.Layer1Scores...)
	combined = append(combined, ms.Layer2Scores...)
	return combined
}

// TrainMeta trains the meta-classifier on combined Layer 1 + Layer 2 outputs.
func TrainMeta(samples []MetaTrainingSample) *MetaClassifier {
	xData := make([][]float64, len(samples))
	yData := make([]int, len(samples))

	for i, s := range samples {
		xData[i] = s.CombinedFeatures()
		if s.Flagged {
			yData[i] = 1
		} else {
			yData[i] = 0
		}
	}

	forest := rf.Forest{}
	forest.Data = rf.ForestData{
		X:     xData,
		Class: yData,
	}
	forest.Train(200) // 200 trees for the meta-learner

	return &MetaClassifier{
		Forest:  forest,
		Trained: true,
	}
}

// PredictMeta returns the flag probability from the meta-classifier.
func (mc *MetaClassifier) PredictMeta(layer1Scores, layer2Scores []float64) float64 {
	if !mc.Trained {
		return 0
	}
	combined := make([]float64, 0, 30)
	combined = append(combined, layer1Scores...)
	combined = append(combined, layer2Scores...)

	votes := mc.Forest.Vote(combined)
	total := 0.0
	flagged := 0.0
	for class, count := range votes {
		total += count
		if class == 1 {
			flagged = count
		}
	}
	if total == 0 {
		return 0
	}
	return flagged / total
}

// PredictBC applies the co-occurrence principle: only flag when multiple
// classifiers fire simultaneously, mirroring the KK pattern pipeline's
// requirement for concurrent signals to reduce false positives.
//
// Returns a BC-equivalent score (0.0-1.0) and whether it crosses threshold.
func (cs *ClassifierSuite) PredictBC(features []float64, weights map[string]float64, threshold float64) (float64, bool) {
	perMetric := cs.PredictAll(features)

	// Weighted sum (same as KK's ComputeBC)
	weightedSum := 0.0
	activeCount := 0
	highCount := 0

	for key, prob := range perMetric {
		w, ok := weights[key]
		if !ok {
			continue
		}
		weightedSum += prob * w
		if prob > 0.3 {
			activeCount++
		}
		if prob > 0.6 {
			highCount++
		}
	}

	// Co-occurrence bonus — same formula as src/scorer.go
	cooccurrenceBonus := 0.0
	if activeCount >= 3 {
		cooccurrenceBonus = float64(activeCount-2) * 0.04
	}
	if highCount >= 2 {
		cooccurrenceBonus += float64(highCount-1) * 0.03
	}

	bc := weightedSum + cooccurrenceBonus
	if bc > 1.0 {
		bc = 1.0
	}

	return bc, bc >= threshold
}
