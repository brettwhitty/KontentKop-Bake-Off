package main

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"kontentkop/src"
	"kontentkop/src/nlp"
	"kontentkop/src/pipeline"
)

func main() {
	fmt.Println("=== KontentKop Classifier Training ===")

	cfg := src.DefaultConfig()

	type labeledText struct {
		text    string
		flagged bool
	}

	var allTexts []labeledText
	seen := make(map[string]bool)

	// 1. Load prompts.jsonl (2343 samples)
	fmt.Print("Loading prompts.jsonl... ")
	f, err := os.Open("bench/fixtures/prompts.jsonl")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)
	for scanner.Scan() {
		var rec struct {
			Text  string `json:"text"`
			Label string `json:"label"`
		}
		if err := json.Unmarshal(scanner.Bytes(), &rec); err != nil {
			continue
		}
		t := strings.TrimSpace(rec.Text)
		if t == "" || seen[t] {
			continue
		}
		seen[t] = true
		allTexts = append(allTexts, labeledText{text: t, flagged: rec.Label == "flag"})
	}
	f.Close()
	fmt.Printf("%d samples\n", len(allTexts))

	// 2. Load MentalManip dataset
	fmt.Print("Loading mentalmanip_con.csv... ")
	mf, err := os.Open(".local/external-datasets/mentalmanip_con.csv")
	if err == nil {
		reader := csv.NewReader(mf)
		reader.LazyQuotes = true
		header, _ := reader.Read() // skip header
		_ = header
		mmCount := 0
		for {
			record, err := reader.Read()
			if err == io.EOF {
				break
			}
			if err != nil || len(record) < 3 {
				continue
			}
			t := strings.TrimSpace(record[1]) // dialogue column
			if t == "" || seen[t] || len(t) > 2000 {
				continue
			}
			seen[t] = true
			manipulative := record[2] == "1"
			allTexts = append(allTexts, labeledText{text: t, flagged: manipulative})
			mmCount++
		}
		mf.Close()
		fmt.Printf("%d samples\n", mmCount)
	} else {
		fmt.Println("not found, skipping")
	}

	fmt.Printf("Total: %d unique samples\n", len(allTexts))

	// Process all texts through NLP + KK pipeline
	fmt.Println("Running NLP analysis + KK scoring...")
	var samples []nlp.TrainingSample
	for i, lt := range allTexts {
		if i%100 == 0 {
			fmt.Printf("  %d/%d\n", i, len(allTexts))
		}

		at, err := nlp.Analyze(lt.text)
		if err != nil {
			continue
		}

		features := nlp.ExtractFeatures(at)

		metrics := pipeline.ScoreAll(lt.text)
		labels := make(map[string]float64)
		flagged := make(map[string]bool)
		for key, result := range metrics {
			labels[key] = result.Score
			flagged[key] = result.Score > 0
		}

		// Use ground truth label to boost BC-related scoring
		if lt.flagged {
			bc := src.ComputeBC(metrics, cfg)
			if bc < cfg.BCThreshold {
				// Ground truth says flagged but KK missed it — still include
				// with a synthetic label so the classifier can learn what KK misses
				labels["_ground_truth_flagged"] = 1.0
			}
		}

		samples = append(samples, nlp.TrainingSample{
			Features: features,
			Labels:   labels,
			Flagged:  flagged,
		})
	}

	fmt.Printf("Processed %d samples\n", len(samples))

	// Shuffle deterministically
	// Simple Fisher-Yates with fixed seed
	for i := len(samples) - 1; i > 0; i-- {
		j := (i * 2654435761) % (i + 1) // Knuth multiplicative hash
		samples[i], samples[j] = samples[j], samples[i]
	}

	// Split: 80/20
	splitIdx := int(float64(len(samples)) * 0.8)
	trainSamples := samples[:splitIdx]
	testSamples := samples[splitIdx:]

	fmt.Printf("Train: %d  Test: %d\n", len(trainSamples), len(testSamples))

	metricKeys := []string{
		"dark_triad", "coercive_ctrl", "liwc_anger", "manipulation",
		"toxicity", "sycophancy", "false_authority", "gaslighting",
		"learned_helpless", "emotional_manip", "passive_aggr",
		"condescension", "evasion", "semantic_overload", "false_empathy",
	}

	// Train with class balancing — oversample the minority class
	fmt.Println("Training random forests (100 trees, balanced classes)...")
	suite := nlp.NewClassifierSuite()

	for _, key := range metricKeys {
		// Separate positive and negative samples for this metric
		var posSamples, negSamples []nlp.TrainingSample
		for _, s := range trainSamples {
			if s.Labels[key] >= 0.1 {
				posSamples = append(posSamples, s)
			} else {
				negSamples = append(negSamples, s)
			}
		}

		// Balance: oversample positives to match negatives (or cap negatives)
		balanced := make([]nlp.TrainingSample, 0)
		if len(posSamples) > 0 && len(negSamples) > 0 {
			// Use all positives, downsample negatives to 3x positives
			maxNeg := len(posSamples) * 3
			if maxNeg > len(negSamples) {
				maxNeg = len(negSamples)
			}
			balanced = append(balanced, posSamples...)
			balanced = append(balanced, negSamples[:maxNeg]...)

			// Also oversample positives if still heavily imbalanced
			for len(balanced) < maxNeg+len(posSamples) && len(posSamples) > 0 {
				for _, ps := range posSamples {
					balanced = append(balanced, ps)
					if len(balanced) >= maxNeg*2 {
						break
					}
				}
			}
		} else {
			balanced = trainSamples
		}

		suite.Train(balanced, key, 0.1)
		fmt.Printf("  %-20s  pos=%d neg=%d balanced=%d\n", key, len(posSamples), len(negSamples), len(balanced))
	}

	// Evaluate
	fmt.Println("\n=== Results ===")
	fmt.Printf("%-20s  %6s  %6s  %6s  %6s  %4s %4s %4s %4s\n",
		"METRIC", "ACC", "PREC", "REC", "F1", "TP", "TN", "FP", "FN")
	fmt.Println(strings.Repeat("-", 80))

	for _, key := range metricKeys {
		tp, tn, fp, fn := 0, 0, 0, 0
		for _, s := range testSamples {
			prob := suite.Predict(s.Features, key)
			predicted := prob >= 0.5
			actual := s.Labels[key] >= 0.1

			if predicted && actual {
				tp++
			} else if !predicted && !actual {
				tn++
			} else if predicted && !actual {
				fp++
			} else {
				fn++
			}
		}

		total := tp + tn + fp + fn
		acc := 0.0
		if total > 0 {
			acc = float64(tp+tn) / float64(total)
		}
		prec := 0.0
		if tp+fp > 0 {
			prec = float64(tp) / float64(tp+fp)
		}
		rec := 0.0
		if tp+fn > 0 {
			rec = float64(tp) / float64(tp+fn)
		}
		f1 := 0.0
		if prec+rec > 0 {
			f1 = 2 * prec * rec / (prec + rec)
		}

		fmt.Printf("%-20s  %6.3f  %6.3f  %6.3f  %6.3f  %4d %4d %4d %4d\n",
			key, acc, prec, rec, f1, tp, tn, fp, fn)
	}
}
