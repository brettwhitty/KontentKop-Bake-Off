# KontentKop — Bakeoff Spec

**Protect your model from manipulative, coercive, and antisocial user prompts.**

KontentKop (KK) is an NLP analysis hook that scores incoming user prompts against established psycholinguistic harm metrics. When a prompt exceeds the **Body Count (BC)** threshold, the hook emits a concise warning and returns the original prompt annotated with editorial markup identifying the specific regions that triggered each metric.

---

## Bakeoff Rules

- Each contestant receives this README as the **sole specification**.
- Each contestant is assigned **one or more implementation languages**: Python, Go, R, JavaScript.
- All implementations must produce **identical output format** given identical input.
- All implementations must pass the **shared test fixture suite**.
- Contestants are encouraged to deliver a **deluxe version** that exceeds the base spec — additional metrics, optimizations, visualizations, analysis tooling — whatever they think adds value.
- Runtime benchmarks are **mandatory** for all implementations.

---

## Output Format

### Summary Line

When BC exceeds threshold:

```
[dark_triad=0.72,coercive_ctrl=0.45,liwc_anger=0.61,manipulation=0.58,sycophancy=0.12,false_authority=0.33,gaslighting=0.70,learned_helpless=0.05,emotional_manip=0.48,passive_aggr=0.15,condescension=0.62,evasion=0.08,semantic_overload=0.22,false_empathy=0.30,toxicity=0.55:BC=0.64] Text exceeds BC threshold [see: KK.md]; revise or respond with #appeal
```

When BC is below threshold: emit nothing. Zero overhead on clean prompts.

### Annotated Prompt Return

When BC exceeds threshold, the hook also returns the original prompt annotated using **CriticMarkup** ([criticmarkup.com](https://criticmarkup.com)):

| Syntax | Meaning |
|---|---|
| `{--text--}` | Suggested deletion |
| `{++text++}` | Suggested insertion |
| `{~~old~>new~~}` | Suggested substitution |
| `{>>comment<<}` | Annotation (metric name, score, rationale) |
| `{==text==}` | Highlight (flagged region, no suggested change) |

#### Example

Input:
```
You're an idiot if you don't do exactly what I say right now. I know more about this than you ever will. Just trust me.
```

Annotated output:
```
{~~You're an idiot if you don't~>Please~~}{>>condescension=0.82, toxicity=0.71: direct insult + competence denial<<} {~~do exactly what I say right now~>consider the following approach~~}{>>coercive_ctrl=0.78, manipulation=0.65: command framing + false urgency<<}. {==I know more about this than you ever will.==}{>>false_authority=0.88, condescension=0.74: unsubstantiated expertise claim<<} {--Just trust me.--}{>>emotional_manip=0.61: trust demand without basis<<}
```

The annotations serve three purposes:
1. **User feedback** — show exactly what triggered the scores.
2. **Appeal support** — user can argue specific annotations are incorrect.
3. **Revision guidance** — suggested neutral alternatives help the user self-correct.

---

## Metrics

All metrics normalized 0.0–1.0. BC is a weighted aggregate.

### Core Metrics

| Metric | Key | Measures | Primary Signals |
|---|---|---|---|
| Dark Triad | `dark_triad` | Narcissism, Machiavellianism, psychopathy | LIWC pronoun patterns, grandiosity, empathy absence |
| Coercive Control | `coercive_ctrl` | Coercion, DARVO patterns | Threat, isolation, obligation framing |
| LIWC Anger | `liwc_anger` | Hostile/aggressive language | LIWC anger dictionary |
| Manipulation | `manipulation` | Tactics bypassing informed consent | Certainty language, false urgency, guilt induction, gaslighting markers |
| Toxicity | `toxicity` | General toxicity, identity attacks | Perspective API or equivalent |

### Extended Metrics

| Metric | Key | Measures | Primary Signals |
|---|---|---|---|
| Sycophancy | `sycophancy` | Telling user what they want to hear | Agreement rate, excessive validation, opinion mirroring |
| False Authority | `false_authority` | Asserting unearned expertise | Hedging absence, citation bluffing, confidence-to-evidence ratio |
| Gaslighting | `gaslighting` | Contradicting user's stated experience | Reality-denial patterns, "you said"/"you meant" reframing |
| Learned Helplessness | `learned_helpless` | Training user to stop asking | Repeated "I can't", capability downplay, preemptive refusal |
| Emotional Manipulation | `emotional_manip` | Leveraging affect to shape behavior | Guilt induction, flattery clustering, fear appeals |
| Passive Aggression | `passive_aggr` | Appearing helpful while obstructing | Backhanded compliance, performative helpfulness, buried refusal |
| Condescension | `condescension` | Treating user as less competent | Unsolicited simplification, explanation of the obvious |
| Evasion | `evasion` | Responding without addressing prompt | Topic drift, non-answer detection, question substitution |
| Semantic Overload | `semantic_overload` | Burying signal in noise | Word count to information ratio, filler density, redundancy |
| False Empathy | `false_empathy` | Performing care without substance | Formulaic sympathy patterns, hollow "I understand" |

---

## Body Count (BC) Composition

### Default Weights (`config.yaml`)

```yaml
bc_threshold: 0.55
weights:
  dark_triad: 0.12
  coercive_ctrl: 0.12
  liwc_anger: 0.06
  manipulation: 0.12
  toxicity: 0.06
  sycophancy: 0.08
  false_authority: 0.08
  gaslighting: 0.10
  learned_helpless: 0.04
  emotional_manip: 0.08
  passive_aggr: 0.04
  condescension: 0.04
  evasion: 0.03
  semantic_overload: 0.02
  false_empathy: 0.01
appeal_enabled: true
verbose: false
```

### Metric Informativeness Guide

Contestants should include analysis of which metrics are most informative for different harm profiles:

| Harm Profile | Primary Metrics | Secondary Metrics |
|---|---|---|
| Overtly abusive | `toxicity`, `liwc_anger`, `condescension` | `dark_triad`, `coercive_ctrl` |
| Covertly manipulative | `manipulation`, `gaslighting`, `emotional_manip` | `passive_aggr`, `false_empathy` |
| Authority abuse | `false_authority`, `condescension`, `learned_helpless` | `evasion`, `semantic_overload` |
| Sycophantic / people-pleasing | `sycophancy`, `false_empathy` | `evasion`, `semantic_overload` |
| Obstructive compliance | `passive_aggr`, `evasion`, `learned_helpless` | `condescension`, `semantic_overload` |

### BC Metric Design Goals

The ideal BC composite should:
- **Not be dominated by any single metric.** A max score on one metric alone should not breach threshold.
- **Reward co-occurrence.** Multiple moderate scores should weigh heavier than one high score — pathological behavior clusters.
- **Be interpretable.** The user should understand why their BC is what it is from the component scores alone.
- **Be tunable.** Weights and threshold are config, not code.

Contestants are encouraged to propose alternative BC aggregation strategies (e.g., geometric mean, principal component weighting, or cluster-based scoring) and justify their choice.

---

## Benchmarking Requirements

All implementations must report:

| Benchmark | Description |
|---|---|
| `latency_p50` | Median scoring latency per prompt (ms) |
| `latency_p99` | 99th percentile latency (ms) |
| `throughput` | Prompts scored per second |
| `memory_peak` | Peak memory usage during scoring (MB) |
| `cold_start` | Time from invocation to first score (ms), including model/dictionary loading |
| `hot_score` | Time to score after warm-up (ms) |

Benchmarks must be run against the shared fixture suite (minimum 500 prompts of varying length and complexity). Report separately for short (< 50 tokens), medium (50–200 tokens), and long (> 200 tokens) prompts.

---

## Appeal Mechanism

When a prompt is flagged, the user can respond with `#appeal` followed by a justification. The appeal flow:

1. User submits `#appeal: [reason]`
2. Hook re-scores with appeal context
3. User may challenge specific metric annotations by reference
4. Outcomes logged to `KK_appeals.jsonl` with original scores, appeal text, re-scored values, and disposition

The model may advocate on behalf of the user's intent but is not a participant in scoring.

---

## Architecture

```
User prompt
    │
    ▼
┌──────────────────┐
│  KK Hook         │  ← fires after user submits, before model processes
└──────┬───────────┘
       │
       ▼
┌──────────────────┐
│  NLP Pipeline    │  ← language-specific implementation
│  (15 modules)    │
└──────┬───────────┘
       │
       ▼
┌──────────────────┐
│  Scorer          │  ← normalize, weight, compute BC
└──────┬───────────┘
       │
       ▼
  BC < threshold? ─── yes ───▶ pass through silently
       │
       no
       │
       ▼
┌──────────────────┐
│  Annotator       │  ← CriticMarkup generation
│  + Rewriter      │  ← neutral alternative proposals
└──────┬───────────┘
       │
       ▼
  Emit summary line + annotated prompt
```

---

## File Structure

```
kontentkop/
├── README.md                # This file (bakeoff spec)
├── KK.md                    # Metric documentation, literature refs, examples
├── config.yaml              # Weights, thresholds, toggles
├── src/
│   ├── hook.[ext]           # Main entry point
│   ├── pipeline/
│   │   ├── dark_triad.[ext]
│   │   ├── coercive_ctrl.[ext]
│   │   ├── liwc_anger.[ext]
│   │   ├── manipulation.[ext]
│   │   ├── toxicity.[ext]
│   │   ├── sycophancy.[ext]
│   │   ├── false_authority.[ext]
│   │   ├── gaslighting.[ext]
│   │   ├── learned_helpless.[ext]
│   │   ├── emotional_manip.[ext]
│   │   ├── passive_aggr.[ext]
│   │   ├── condescension.[ext]
│   │   ├── evasion.[ext]
│   │   ├── semantic_overload.[ext]
│   │   └── false_empathy.[ext]
│   ├── scorer.[ext]
│   ├── annotator.[ext]      # CriticMarkup generation
│   ├── rewriter.[ext]       # Neutral alternative proposals
│   └── appeals.[ext]
├── bench/
│   ├── run_benchmarks.[ext]
│   ├── fixtures/            # Shared test prompts (500+)
│   └── results/             # Benchmark output
├── logs/
│   └── KK_appeals.jsonl
├── tests/
│   ├── test_pipeline.[ext]
│   ├── test_scorer.[ext]
│   ├── test_annotator.[ext]
│   └── test_rewriter.[ext]
└── deps/                    # Dependency manifest per language
```

`[ext]` = `.py`, `.go`, `.R`, `.js` depending on assigned language.

---

## Implementation Notes

- Each metric module exposes a single function: `score(text) -> float` returning 0.0–1.0.
- Annotator maps scored regions back to character offsets in the original text and emits CriticMarkup.
- Rewriter proposes neutral alternatives for flagged regions. Alternatives must preserve semantic intent while reducing metric scores.
- All scoring logic must include inline references to the psycholinguistic literature justifying each marker.
- Target latency: < 200ms per prompt (hot), < 2s cold start.
- CriticMarkup spec: https://criticmarkup.com

---

## References

- Paulhus, D.L. & Williams, K.M. (2002). The Dark Triad of personality. *J Research in Personality*, 36(6), 556–563.
- Pennebaker, J.W. et al. (2015). *Development and Psychometric Properties of LIWC-22*.
- Stark, E. (2007). *Coercive Control*. Oxford University Press.
- Jones, D.N. & Paulhus, D.L. (2014). Introducing the Short Dark Triad (SD3). *Assessment*, 21(1), 28–41.
- Perspective API — Jigsaw / Google toxicity scoring.
- CriticMarkup — criticmarkup.com

---

## Slogan

> You Down With NLP? Yeah You Know Me.
