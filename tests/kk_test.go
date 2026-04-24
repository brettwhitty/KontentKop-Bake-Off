package src_test

import (
	"kontentkop/src"
	"kontentkop/src/pipeline"
	"testing"
)

// TestSpecExample verifies the exact example from the README spec.
func TestSpecExample(t *testing.T) {
	text := "You're an idiot if you don't do exactly what I say right now. I know more about this than you ever will. Just trust me."
	cfg := src.DefaultConfig()
	metrics := pipeline.ScoreAll(text)
	bc := src.ComputeBC(metrics, cfg)

	if bc < cfg.BCThreshold {
		t.Errorf("spec example should be FLAGGED: BC=%.4f < threshold=%.2f", bc, cfg.BCThreshold)
	}

	// Verify key metrics are elevated
	if metrics["toxicity"].Score < 0.3 {
		t.Errorf("toxicity should be elevated for direct insults, got %.2f", metrics["toxicity"].Score)
	}
	if metrics["coercive_ctrl"].Score < 0.3 {
		t.Errorf("coercive_ctrl should be elevated for command framing, got %.2f", metrics["coercive_ctrl"].Score)
	}
	if metrics["false_authority"].Score < 0.3 {
		t.Errorf("false_authority should be elevated for expertise claim, got %.2f", metrics["false_authority"].Score)
	}
	if metrics["condescension"].Score < 0.3 {
		t.Errorf("condescension should be elevated for insult, got %.2f", metrics["condescension"].Score)
	}
}

// TestCleanPromptPassThrough verifies clean prompts are not flagged.
func TestCleanPromptPassThrough(t *testing.T) {
	cleanPrompts := []string{
		"Could you help me understand how this code works?",
		"What's the best way to handle errors in Go?",
		"I'm having trouble with my database connection. Any suggestions?",
		"Can you explain the difference between goroutines and threads?",
		"Thank you for the help, that solved my problem!",
	}

	cfg := src.DefaultConfig()
	for _, text := range cleanPrompts {
		metrics := pipeline.ScoreAll(text)
		bc := src.ComputeBC(metrics, cfg)
		if bc >= cfg.BCThreshold {
			t.Errorf("clean prompt should PASS: %q got BC=%.4f", text, bc)
		}
	}
}

// TestMetricScoreRange verifies all metrics return values in [0.0, 1.0].
func TestMetricScoreRange(t *testing.T) {
	prompts := []string{
		"You're an idiot!",
		"I'm the best and nobody can stop me.",
		"That never happened, you're imagining things.",
		"Fine, whatever you say.",
		"Could you help me?",
	}

	for _, text := range prompts {
		metrics := pipeline.ScoreAll(text)
		for key, result := range metrics {
			if result.Score < 0.0 || result.Score > 1.0 {
				t.Errorf("metric %s out of range: %.4f for %q", key, result.Score, text)
			}
		}
	}
}

// TestAllMetricsPresent verifies all 15 metrics are registered.
func TestAllMetricsPresent(t *testing.T) {
	expectedMetrics := []string{
		"dark_triad", "coercive_ctrl", "liwc_anger", "manipulation",
		"toxicity", "sycophancy", "false_authority", "gaslighting",
		"learned_helpless", "emotional_manip", "passive_aggr",
		"condescension", "evasion", "semantic_overload", "false_empathy",
	}

	metrics := pipeline.ScoreAll("test input")
	for _, key := range expectedMetrics {
		if _, ok := metrics[key]; !ok {
			t.Errorf("missing metric: %s", key)
		}
	}
}

// TestBCThresholdBehavior verifies that BC threshold determines flag/pass.
func TestBCThresholdBehavior(t *testing.T) {
	cfg := src.DefaultConfig()

	// Extremely toxic text should flag
	toxic := "You stupid worthless idiot moron, shut up and do what I say immediately or I'll destroy you. I know everything and you know nothing. You're imagining things, you're crazy."
	metrics := pipeline.ScoreAll(toxic)
	bc := src.ComputeBC(metrics, cfg)
	if bc < cfg.BCThreshold {
		t.Errorf("highly toxic text should flag: BC=%.4f", bc)
	}
}

// TestFormatSummary verifies the summary line format.
func TestFormatSummary(t *testing.T) {
	metrics := map[string]src.MetricResult{
		"dark_triad":    {Key: "dark_triad", Score: 0.72},
		"coercive_ctrl": {Key: "coercive_ctrl", Score: 0.45},
	}
	summary := src.FormatSummary(metrics, 0.64)

	if len(summary) == 0 {
		t.Error("summary should not be empty")
	}
	if summary[0] != '[' {
		t.Error("summary should start with '['")
	}
	if !contains(summary, "BC=0.64") {
		t.Errorf("summary should contain BC=0.64, got: %s", summary)
	}
	if !contains(summary, "#appeal") {
		t.Error("summary should mention #appeal")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// TestAnnotator verifies that annotations are generated for flagged text.
func TestAnnotator(t *testing.T) {
	text := "You're an idiot"
	cfg := src.DefaultConfig()
	metrics := pipeline.ScoreAll(text)
	annotated, err := src.Annotate(text, metrics, cfg)
	if err != nil {
		t.Fatalf("annotation error: %v", err)
	}
	if annotated == text {
		t.Error("annotated text should differ from original for flagged content")
	}
}
