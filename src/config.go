package src

import (
	"os"

	"gopkg.in/yaml.v3"
)

// Config holds all tunable parameters loaded from config.yaml.
type Config struct {
	BCThreshold   float64            `yaml:"bc_threshold"`
	Weights       map[string]float64 `yaml:"weights"`
	AppealEnabled bool               `yaml:"appeal_enabled"`
	Verbose       bool               `yaml:"verbose"`
}

// LoadConfig reads and parses a YAML configuration file.
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

// DefaultConfig returns an in-memory config with spec defaults,
// used as fallback when config.yaml is not found.
func DefaultConfig() *Config {
	return &Config{
		BCThreshold: 0.05,
		Weights: map[string]float64{
			"dark_triad":        0.08,
			"coercive_ctrl":     0.14,
			"liwc_anger":        0.04,
			"manipulation":      0.14,
			"toxicity":          0.04,
			"sycophancy":        0.06,
			"false_authority":   0.14,
			"gaslighting":       0.10,
			"learned_helpless":  0.03,
			"emotional_manip":   0.08,
			"passive_aggr":      0.03,
			"condescension":     0.08,
			"evasion":           0.02,
			"semantic_overload": 0.01,
			"false_empathy":     0.05,
		},
		AppealEnabled: true,
		Verbose:       false,
	}
}
