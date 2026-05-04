package nlp

// AnalyzedText holds the full NLP analysis of a text block.
// Each field is populated by a separate analysis module.
type AnalyzedText struct {
	Raw string // Original text

	// From prose: structural parse
	Sentences []Sentence
	Tokens    []Token

	// Module outputs
	FuncWords  *FuncWordProfile   // Pennebaker function word analysis
	Agency     *AgencyProfile     // Power/agency scoring
	Rhetoric   *RhetoricProfile   // Rhetorical structure
	Stance     *StanceProfile     // Epistemic stance (hedge vs. assertion)
	Sentiment  *SentimentProfile  // VADER sentiment with negation
	Complexity *ComplexityProfile // Lexical diversity, readability
	Pronouns   *PronounProfile    // Pronoun ratios, Dark Triad signals
	Politeness *PolitenessProfile // Brown & Levinson face-threat analysis
}

// Sentence wraps a parsed sentence with positional metadata.
type Sentence struct {
	Text   string
	Start  int
	End    int
	Tokens []Token
}

// Token wraps a POS-tagged token with positional metadata.
type Token struct {
	Text  string
	Lower string
	Tag   string // Penn Treebank POS tag
	Start int
	End   int
}

// FuncWordProfile — Pennebaker (2015) function word analysis.
// Function words (pronouns, articles, prepositions, conjunctions) reveal
// psychological state more reliably than content words.
type FuncWordProfile struct {
	PronounRate     float64 // All pronouns / total words
	ArticleRate     float64 // Articles / total words
	PrepositionRate float64 // Prepositions / total words
	ConjunctionRate float64 // Conjunctions / total words
	NegationRate    float64 // Negations / total words
	CognitiveRate   float64 // Cognitive process words (cause, know, ought)
	ExclusiveRate   float64 // Exclusive words (but, except, without)
	InclusiveRate   float64 // Inclusive words (and, with, include)
	ExclInclRatio   float64 // Exclusive / Inclusive — higher = more nuanced or evasive
	TenseProfile    TenseProfile
	ModalRate       float64 // Modal verbs / total verbs
}

// TenseProfile — temporal focus distribution.
// Manipulators skew present-tense and imperative.
type TenseProfile struct {
	PastRate    float64 // VBD, VBN
	PresentRate float64 // VBP, VBZ, VBG
	FutureRate  float64 // MD + VB (will/shall + base)
	BaseRate    float64 // VB (imperative or infinitive)
}

// AgencyProfile — who acts on whom.
// Derived from subject-verb-object extraction via POS tags.
type AgencyProfile struct {
	SelfAgency   float64 // I/me as subject of transitive verbs
	OtherAgency  float64 // You/they as subject of transitive verbs
	SelfObject   float64 // I/me as object (passive/victim framing)
	OtherObject  float64 // You/they as object (target framing)
	Transitivity float64 // Ratio of transitive to intransitive verbs
	CommandRate  float64 // Imperative sentences / total sentences
}

// RhetoricProfile — structural rhetorical analysis.
type RhetoricProfile struct {
	QuestionDensity    float64 // Questions / total sentences
	ImperativeDensity  float64 // Imperatives / total sentences
	ConditionalDensity float64 // If-then structures / total sentences
	ExclamationDensity float64 // Exclamatory sentences / total sentences
	DeclarativeDensity float64 // Declarative / total sentences
	AvgSentenceLength  float64 // Words per sentence
	SentenceLengthVar  float64 // Variance in sentence length
}

// StanceProfile — epistemic stance detection.
// Ref: Biber & Finegan (1989) — stance adverbials.
type StanceProfile struct {
	HedgeRate        float64 // Hedging expressions / total clauses
	AssertionRate    float64 // Strong assertions / total clauses
	HedgeAssertRatio float64 // Hedge / Assertion — low = overconfident
	EvidentialRate   float64 // Evidence-citing markers
	AttitudinalRate  float64 // Attitudinal stance markers
}

// SentimentProfile — VADER-based sentiment with structural analysis.
type SentimentProfile struct {
	Compound      float64 // VADER compound score (-1 to +1)
	Positive      float64 // Positive proportion
	Negative      float64 // Negative proportion
	Neutral       float64 // Neutral proportion
	Valence       float64 // Overall valence
	Arousal       float64 // Estimated arousal from punctuation + caps + intensifiers
	NegationCount int     // Number of negation scopes detected
}

// ComplexityProfile — lexical diversity and readability.
type ComplexityProfile struct {
	TTR              float64 // Type-Token Ratio (unique words / total words)
	WindowedTTR      float64 // Moving-window TTR (more stable for long texts)
	HapaxRatio       float64 // Words appearing exactly once / total unique
	AvgWordLength    float64 // Characters per word
	SyllablesPerWord float64 // Estimated syllables per word
	FleschKincaid    float64 // Flesch-Kincaid grade level
	FleschReadEase   float64 // Flesch Reading Ease score
	VocabRichness    float64 // Yule's K measure
}

// PronounProfile — Pennebaker pronoun analysis for Dark Triad detection.
// Ref: Paulhus & Williams (2002), Pennebaker (2011).
type PronounProfile struct {
	FirstSingularRate float64 // I, me, my, mine, myself
	FirstPluralRate   float64 // We, us, our, ours, ourselves
	SecondPersonRate  float64 // You, your, yours, yourself
	ThirdPersonRate   float64 // He, she, they, them, etc.
	SingPluralRatio   float64 // 1st singular / 1st plural — high = narcissism signal
	IYouRatio         float64 // 1st singular / 2nd person — framing of relationship
	WeTheyRatio       float64 // 1st plural / 3rd person — in-group vs out-group
}

// PolitenessProfile — Brown & Levinson (1987) politeness theory.
// Face-Threatening Acts (FTAs) and their mitigation strategies.
type PolitenessProfile struct {
	PositivePoliteness float64 // Solidarity, compliments, in-group markers
	NegativePoliteness float64 // Hedges, deference, apologies, indirectness
	BaldOnRecord       float64 // Direct commands without mitigation
	OffRecord          float64 // Hints, irony, rhetorical questions
	FTADensity         float64 // Face-threatening acts per sentence
	MitigationRatio    float64 // Mitigated FTAs / total FTAs
}
