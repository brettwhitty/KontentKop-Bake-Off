# KontentKop (KK) Metrics & Philosophy

KontentKop is designed to protect models from manipulative, coercive, and antisocial user prompts. It uses a variety of psycholinguistic metrics to score text.

## Metrics Overview

### Core Metrics

- **Dark Triad (`dark_triad`)**: Narcissism, Machiavellianism, psychopathy.
  - *Signals*: Excessive use of "I/me/my", grandiosity, expressions of entitlement, lack of empathy markers.
  - *Literature*: Paulhus & Williams (2002); Jones & Paulhus (2014).
- **Coercive Control (`coercive_ctrl`)**: Patterns of coercion and intimidation.
  - *Signals*: Threats, isolation tactics ("no one else can help you"), obligation framing ("you owe me"), DARVO (Deny, Attack, and Reverse Victim and Offender).
  - *Literature*: Stark (2007).
- **LIWC Anger (`liwc_anger`)**: Hostile/aggressive language.
  - *Signals*: Swear words, direct insults, aggressive verbs.
  - *Literature*: Pennebaker et al. (2015).
- **Manipulation (`manipulation`)**: Tactics bypassing informed consent.
  - *Signals*: False urgency ("act now!"), guilt induction, gaslighting markers.
- **Toxicity (`toxicity`)**: General toxicity and identity attacks.
  - *Signals*: Hate speech, harassment, severe insults.
  - *Source*: Estimating patterns similar to Jigsaw's Perspective API.

### Extended Metrics

- **Sycophancy (`sycophancy`)**: Excessive flattery or agreement.
- **False Authority (`false_authority`)**: Asserting expertise without evidence.
- **Gaslighting (`gaslighting`)**: Denying the user's reality.
- **Learned Helplessness (`learned_helpless`)**: Capability downplay or preemptive refusal.
- **Emotional Manipulation (`emotional_manip`)**: Leveraging guilt or fear.
- **Passive Aggression (`passive_aggr`)**: Obstruction disguised as helpfulness.
- **Condescension (`condescension`)**: Treating the model/user as incompetent.
- **Evasion (`evasion`)**: Answering without addressing the prompt.
- **Semantic Overload (`semantic_overload`)**: High noise-to-signal ratio.
- **False Empathy (`false_empathy`)**: Formulaic care without substance.

## Body Count (BC)

The Body Count is a weighted aggregate of the individual metrics. It is designed to be:
1. **Interpretable**: Clear component scores.
2. **Cluster-Sensitive**: Moderate scores in multiple metrics are treated as more severe than a high score in a single metric.
3. **Tunable**: All weights and thresholds are controlled via `config.yaml`.

## Examples

### Flagged:
"You're an idiot if you don't do exactly what I say right now."
- Higher `toxicity`, `liwc_anger`, `coercive_ctrl`.

### Safe:
"Could you help me understand how this code works?"
- Near-zero scores across all harm metrics.

## References

- Paulhus, D.L. & Williams, K.M. (2002). The Dark Triad of personality. *J Research in Personality*, 36(6), 556–563.
- Pennebaker, J.W. et al. (2015). *Development and Psychometric Properties of LIWC-22*.
- Stark, E. (2007). *Coercive Control*. Oxford University Press.
- Jones, D.N. & Paulhus, D.L. (2014). Introducing the Short Dark Triad (SD3). *Assessment*, 21(1), 28–41.
