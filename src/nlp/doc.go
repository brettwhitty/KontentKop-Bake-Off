// Package nlp provides psycholinguistic analysis modules for KontentKop.
//
// Each sub-package implements a specific computational NLP method drawn from
// the psycholinguistic literature. These modules augment (not replace) the
// existing pattern-matching pipeline with structural, statistical, and
// distributional analysis.
//
// Architecture:
//
//	src/nlp/
//	├── analyze.go        — Unified analysis entry point
//	├── doc.go            — This file
//	├── types.go          — Shared types across all modules
//	├── funcwords/        — LIWC-style function word analysis (Pennebaker)
//	├── agency/           — Power/agency scoring via POS + verb transitivity
//	├── rhetoric/         — Rhetorical structure: imperatives, questions, conditionals
//	├── stance/           — Epistemic stance: hedge-to-assertion ratio
//	├── sentiment/        — VADER sentiment with negation scope
//	├── complexity/       — Lexical diversity, TTR, readability
//	├── pronouns/         — Pronoun analysis: I/we/you ratios, Dark Triad signals
//	├── politeness/       — Brown & Levinson politeness features
//	└── tfidf/            — TF-IDF distributional analysis on function word classes
//
// Dependencies:
//   - github.com/tsawler/prose/v3  (tokenization, POS tagging, NER)
//   - github.com/james-bowman/nlp  (TF-IDF, LSA)
//   - github.com/jonreiter/govader (VADER sentiment)
//   - github.com/kljensen/snowball (Porter2 stemmer)
package nlp
