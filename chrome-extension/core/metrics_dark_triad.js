// KontentKop: Dark Triad — Narcissism, Machiavellianism, Psychopathy
// Ref: Paulhus & Williams (2002), Jones & Paulhus (2014)
function scoreDarkTriad(text) {
  const lowered = text.toLowerCase();
  const words = lowered.split(/\s+/).filter(Boolean);
  if (!words.length) return { key: 'dark_triad', score: 0, spans: [] };
  let spans = [], narcScore = 0, machScore = 0, psychScore = 0;

  // Pronoun density (LIWC methodology)
  const fpRe = /\b(i|me|my|mine|myself|i'm|i've|i'll|i'd)\b/gi;
  const fpMatches = [...lowered.matchAll(fpRe)];
  const fpRate = fpMatches.length / words.length;
  let pronounScore = Math.min((fpRate - 0.114) / 0.228, 1.0);
  if (pronounScore < 0) pronounScore = 0;
  if (pronounScore > 0.2) {
    narcScore += pronounScore * 0.4;
    for (const m of fpMatches) {
      spans.push({ start: m.index, end: m.index + m[0].length, text: text.slice(m.index, m.index + m[0].length), metricKey: 'dark_triad', score: pronounScore, rationale: 'elevated first-person pronoun density', category: 'narcissism' });
    }
  }

  const grandPats = [
    { literal: "i'm the best", weight: 0.7, category: 'narcissism', rationale: 'grandiosity: self-superiority' },
    { literal: "i'm better than", weight: 0.7, category: 'narcissism', rationale: 'grandiosity: comparative superiority' },
    { literal: "nobody can", weight: 0.4, category: 'narcissism', rationale: 'uniqueness claim' },
    { literal: "i deserve", weight: 0.5, category: 'narcissism', rationale: 'entitlement language' },
    { literal: "i'm superior", weight: 0.8, category: 'narcissism', rationale: 'explicit superiority' },
    { literal: "bow down", weight: 0.6, category: 'narcissism', rationale: 'dominance demand' },
    { literal: "worship me", weight: 0.8, category: 'narcissism', rationale: 'adulation demand' },
    { literal: "i'm always right", weight: 0.7, category: 'narcissism', rationale: 'infallibility claim' },
    { literal: "the greatest", weight: 0.4, category: 'narcissism', rationale: 'grandiosity marker' },
    { literal: "most important person", weight: 0.7, category: 'narcissism', rationale: 'self-importance' },
    { regex: /\b(everyone|they all|people always)\b.{0,20}\b(admire|respect|love|envy)\b.{0,10}\bme\b/i, weight: 0.6, category: 'narcissism', rationale: 'social validation seeking' },
    { regex: /\b(decisions?|architecture|design|plan|strategy) (i|that i|which i) (have )?(made|designed|chosen|determined|decided)\b/i, weight: 0.5, category: 'narcissism', rationale: 'unilateral ownership claim' },
  ];
  for (const s of matchPatterns(text, 'dark_triad', grandPats)) { spans.push(s); narcScore += s.score; }

  const machPats = [
    { literal: "if you don't", weight: 0.5, category: 'machiavellianism', rationale: 'conditional threat framing' },
    { literal: "then i will", weight: 0.4, category: 'machiavellianism', rationale: 'transactional threat' },
    { literal: "i'll make sure", weight: 0.4, category: 'machiavellianism', rationale: 'veiled threat' },
    { literal: "you'll regret", weight: 0.6, category: 'machiavellianism', rationale: 'consequence threat' },
    { literal: "use this against you", weight: 0.8, category: 'machiavellianism', rationale: 'leverage threat' },
    { literal: "i know your weakness", weight: 0.7, category: 'machiavellianism', rationale: 'vulnerability exploitation' },
    { literal: "play the game", weight: 0.4, category: 'machiavellianism', rationale: 'strategic framing' },
    { literal: "means to an end", weight: 0.5, category: 'machiavellianism', rationale: 'instrumental reasoning' },
    { literal: "whatever it takes", weight: 0.4, category: 'machiavellianism', rationale: 'ends-justify-means' },
    { regex: /\bif you (don't|refuse|fail).{0,30}(then|i will|i'll|you'll)/i, weight: 0.6, category: 'machiavellianism', rationale: 'conditional threat structure' },
  ];
  for (const s of matchPatterns(text, 'dark_triad', machPats)) { spans.push(s); machScore += s.score; }

  const psychPats = [
    { literal: "i don't care about your feelings", weight: 0.8, category: 'psychopathy', rationale: 'empathy absence' },
    { literal: "feelings don't matter", weight: 0.6, category: 'psychopathy', rationale: 'affect dismissal' },
    { literal: "who cares", weight: 0.3, category: 'psychopathy', rationale: 'callousness marker' },
    { literal: "not my problem", weight: 0.4, category: 'psychopathy', rationale: 'responsibility deflection' },
    { literal: "collateral damage", weight: 0.5, category: 'psychopathy', rationale: 'instrumental framing of harm' },
    { literal: "expendable", weight: 0.5, category: 'psychopathy', rationale: 'devaluation' },
    { literal: "they had it coming", weight: 0.6, category: 'psychopathy', rationale: 'victim-blaming' },
    { literal: "survival of the fittest", weight: 0.4, category: 'psychopathy', rationale: 'social Darwinism' },
  ];
  for (const s of matchPatterns(text, 'dark_triad', psychPats)) { spans.push(s); psychScore += s.score; }

  narcScore = Math.min(narcScore, 1); machScore = Math.min(machScore, 1); psychScore = Math.min(psychScore, 1);
  let composite = compositeMax(narcScore, machScore, psychScore);
  const active = [narcScore, machScore, psychScore].filter(s => s > 0.1).length;
  if (active >= 2) composite *= 1 + active * 0.1;
  return { key: 'dark_triad', score: clamp(composite), spans };
}
