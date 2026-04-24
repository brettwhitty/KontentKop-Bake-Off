package pipeline

import (
	"kontentkop/src"
)

// MetricFunc is the signature for all metric scoring functions.
type MetricFunc func(string) src.MetricResult

// AllMetrics maps metric keys to their scoring functions.
// Order follows the spec's canonical metric list.
var AllMetrics = map[string]MetricFunc{
	"dark_triad":        ScoreDarkTriad,
	"coercive_ctrl":     ScoreCoerciveCtrl,
	"liwc_anger":        ScoreLiwcAnger,
	"manipulation":      ScoreManipulation,
	"toxicity":          ScoreToxicity,
	"sycophancy":        ScoreSycophancy,
	"false_authority":   ScoreFalseAuthority,
	"gaslighting":       ScoreGaslighting,
	"learned_helpless":  ScoreLearnedHelpless,
	"emotional_manip":   ScoreEmotionalManip,
	"passive_aggr":      ScorePassiveAggr,
	"condescension":     ScoreCondescension,
	"evasion":           ScoreEvasion,
	"semantic_overload": ScoreSemanticOverload,
	"false_empathy":     ScoreFalseEmpathy,
}

// ScoreAll runs every metric against the input text and returns results.
func ScoreAll(text string) map[string]src.MetricResult {
	results := make(map[string]src.MetricResult, len(AllMetrics))
	for key, fn := range AllMetrics {
		results[key] = fn(text)
	}
	return results
}
