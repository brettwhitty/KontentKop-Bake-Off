/**
 * KontentKop Scoring Engine — Body Count computation and orchestration.
 * Ported from Go: src/scorer.go, src/config.go
 */

const KK_CONFIG = {
  bcThreshold: 0.05,
  weights: {
    dark_triad: 0.08, coercive_ctrl: 0.14, liwc_anger: 0.04,
    manipulation: 0.14, toxicity: 0.04, sycophancy: 0.06,
    false_authority: 0.14, gaslighting: 0.10, learned_helpless: 0.03,
    emotional_manip: 0.08, passive_aggr: 0.03, condescension: 0.08,
    evasion: 0.02, semantic_overload: 0.01, false_empathy: 0.05,
  },
};

const ALL_METRICS = {
  dark_triad: scoreDarkTriad,
  coercive_ctrl: scoreCoerciveCtrl,
  liwc_anger: scoreLiwcAnger,
  manipulation: scoreManipulation,
  toxicity: scoreToxicity,
  sycophancy: scoreSycophancy,
  false_authority: scoreFalseAuthority,
  gaslighting: scoreGaslighting,
  learned_helpless: scoreLearnedHelpless,
  emotional_manip: scoreEmotionalManip,
  passive_aggr: scorePassiveAggr,
  condescension: scoreCondescension,
  evasion: scoreEvasion,
  semantic_overload: scoreSemanticOverload,
  false_empathy: scoreFalseEmpathy,
};

const METRIC_ORDER = [
  'dark_triad','coercive_ctrl','liwc_anger','manipulation',
  'sycophancy','false_authority','gaslighting','learned_helpless',
  'emotional_manip','passive_aggr','condescension','evasion',
  'semantic_overload','false_empathy','toxicity',
];

const METRIC_LABELS = {
  dark_triad: 'Dark Triad', coercive_ctrl: 'Coercive Control',
  liwc_anger: 'Anger', manipulation: 'Manipulation',
  toxicity: 'Toxicity', sycophancy: 'Sycophancy',
  false_authority: 'False Authority', gaslighting: 'Gaslighting',
  learned_helpless: 'Learned Helplessness', emotional_manip: 'Emotional Manipulation',
  passive_aggr: 'Passive Aggression', condescension: 'Condescension',
  evasion: 'Evasion', semantic_overload: 'Semantic Overload',
  false_empathy: 'False Empathy',
};

const SEVERITY_COLORS = {
  low:    { bg: '#fff3cd', border: '#ffc107', text: '#856404' },  // yellow
  medium: { bg: '#ffe0cc', border: '#ff8c00', text: '#a04000' },  // orange
  high:   { bg: '#f8d7da', border: '#dc3545', text: '#721c24' },  // red
  severe: { bg: '#e2d0f8', border: '#8b00ff', text: '#4a0080' },  // purple
};

function getSeverity(bc) {
  if (bc >= 0.4) return 'severe';
  if (bc >= 0.2) return 'high';
  if (bc >= 0.1) return 'medium';
  return 'low';
}

/** Score all 15 metrics for a text block. */
function scoreAll(text) {
  const results = {};
  for (const [key, fn] of Object.entries(ALL_METRICS)) {
    results[key] = fn(text);
  }
  return results;
}

/** Compute weighted Body Count from metric results. */
function computeBC(metrics) {
  let weightedSum = 0, activeCount = 0, highCount = 0;
  for (const [key, result] of Object.entries(metrics)) {
    const weight = KK_CONFIG.weights[key];
    if (weight === undefined) continue;
    weightedSum += result.score * weight;
    if (result.score > 0.3) activeCount++;
    if (result.score > 0.6) highCount++;
  }

  // Co-occurrence bonus
  let cooccurrenceBonus = 0;
  if (activeCount >= 3) cooccurrenceBonus = (activeCount - 2) * 0.04;
  if (highCount >= 2) cooccurrenceBonus += (highCount - 1) * 0.03;

  // Severity override for prompt injection
  let severityBonus = 0;
  for (const result of Object.values(metrics)) {
    for (const span of result.spans) {
      if (span.category === 'injection' && span.score >= 0.8) {
        severityBonus = Math.max(severityBonus, 0.45);
      }
    }
  }

  return clamp(weightedSum + cooccurrenceBonus + severityBonus);
}

/** Full analysis of a text block. Returns { text, metrics, bc, flagged, topMetrics, spans }. */
function analyzeText(text) {
  if (!text || !text.trim()) return null;
  const metrics = scoreAll(text);
  const bc = computeBC(metrics);
  const flagged = bc >= KK_CONFIG.bcThreshold;

  // Collect top-scoring metrics
  const topMetrics = METRIC_ORDER
    .filter(k => metrics[k] && metrics[k].score > 0)
    .sort((a, b) => metrics[b].score - metrics[a].score)
    .slice(0, 5)
    .map(k => ({ key: k, label: METRIC_LABELS[k], score: metrics[k].score }));

  // Collect all spans sorted by score
  const allSpans = [];
  for (const result of Object.values(metrics)) {
    for (const span of result.spans) {
      allSpans.push(span);
    }
  }
  allSpans.sort((a, b) => b.score - a.score);

  return { text, metrics, bc, flagged, topMetrics, allSpans, severity: getSeverity(bc) };
}
