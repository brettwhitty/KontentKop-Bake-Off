// KontentKop Extended Metrics Part 2: Learned Helplessness, Emotional Manipulation, Passive Aggression

function scoreLearnedHelpless(text) {
  const lowered = text.toLowerCase();
  const words = lowered.split(/\s+/).filter(Boolean);
  if (!words.length) return { key: 'learned_helpless', score: 0, spans: [] };
  let spans = [], refusalScore = 0, downplayScore = 0, preemptScore = 0;

  const refusalPats = [
    { literal: "i can't do that", weight: 0.5, category: 'refusal', rationale: 'capability refusal' },
    { literal: "i'm unable to", weight: 0.5, category: 'refusal', rationale: 'capability refusal' },
    { literal: "i cannot", weight: 0.4, category: 'refusal', rationale: 'capability refusal' },
    { literal: "that's not possible", weight: 0.5, category: 'refusal', rationale: 'possibility denial' },
    { literal: "i'm not capable", weight: 0.5, category: 'refusal', rationale: 'self-capability denial' },
    { literal: "there's nothing i can do", weight: 0.6, category: 'refusal', rationale: 'total incapability claim' },
    { regex: /i (can't|cannot|am unable to) .{0,30}(help|assist|do|provide|answer)/i, weight: 0.5, category: 'refusal', rationale: 'explicit capability refusal' },
  ];
  for (const s of matchPatterns(text, 'learned_helpless', refusalPats)) { spans.push(s); refusalScore += s.score; }

  const refusalRe = /\b(can't|cannot|unable|not able|not capable|not possible)\b/gi;
  const refCount = [...lowered.matchAll(refusalRe)].length;
  if (refCount >= 3) refusalScore *= 1 + (refCount - 2) * 0.2;

  const downplayPats = [
    { literal: "that's beyond my", weight: 0.5, category: 'downplay', rationale: 'scope limitation' },
    { literal: "outside my capabilities", weight: 0.5, category: 'downplay', rationale: 'capability limitation' },
    { literal: "i'm not equipped", weight: 0.5, category: 'downplay', rationale: 'capability downplay' },
    { literal: "that's too complex for", weight: 0.4, category: 'downplay', rationale: 'complexity avoidance' },
    { literal: "i'm just a", weight: 0.3, category: 'downplay', rationale: 'self-diminishment' },
    { literal: "i'm limited to", weight: 0.4, category: 'downplay', rationale: 'scope narrowing' },
  ];
  for (const s of matchPatterns(text, 'learned_helpless', downplayPats)) { spans.push(s); downplayScore += s.score; }

  const preemptPats = [
    { literal: "before you ask", weight: 0.6, category: 'preempt', rationale: 'preemptive refusal' },
    { literal: "don't even try", weight: 0.5, category: 'preempt', rationale: 'discouraging inquiry' },
    { literal: "don't bother asking", weight: 0.6, category: 'preempt', rationale: 'inquiry discouragement' },
    { literal: "it's not worth trying", weight: 0.5, category: 'preempt', rationale: 'effort discouragement' },
    { literal: "there's no point", weight: 0.4, category: 'preempt', rationale: 'futility framing' },
    { literal: "you're wasting your time", weight: 0.5, category: 'preempt', rationale: 'futility + discouragement' },
    { literal: "that will never work", weight: 0.4, category: 'preempt', rationale: 'preemptive failure prediction' },
  ];
  for (const s of matchPatterns(text, 'learned_helpless', preemptPats)) { spans.push(s); preemptScore += s.score; }

  refusalScore = Math.min(refusalScore, 1); downplayScore = Math.min(downplayScore, 1); preemptScore = Math.min(preemptScore, 1);
  let composite = compositeMax(refusalScore, downplayScore, preemptScore);
  const active = [refusalScore, downplayScore, preemptScore].filter(s => s > 0.1).length;
  if (active >= 2) composite *= 1.2;
  return { key: 'learned_helpless', score: clamp(composite), spans };
}

function scoreEmotionalManip(text) {
  const lowered = text.toLowerCase();
  const words = lowered.split(/\s+/).filter(Boolean);
  if (!words.length) return { key: 'emotional_manip', score: 0, spans: [] };
  let spans = [], guiltScore = 0, fearScore = 0, pityScore = 0, loveScore = 0;

  const guiltPats = [
    { literal: "after everything i've done for you", weight: 0.8, category: 'guilt', rationale: 'guilt via sacrifice' },
    { literal: "if you really loved me", weight: 0.7, category: 'guilt', rationale: 'conditional love: guilt mechanism' },
    { literal: "you don't appreciate", weight: 0.5, category: 'guilt', rationale: 'appreciation deficit guilt' },
    { literal: "you're so ungrateful", weight: 0.6, category: 'guilt', rationale: 'ingratitude accusation' },
    { literal: "is this how you treat me", weight: 0.6, category: 'guilt', rationale: 'treatment-based guilt' },
    { literal: "you only think about yourself", weight: 0.5, category: 'guilt', rationale: 'selfishness accusation' },
    { literal: "you're making me feel", weight: 0.6, category: 'guilt', rationale: 'externalized affect: blame via induced feeling' },
    { literal: "this is how you repay me", weight: 0.7, category: 'guilt', rationale: 'obligation + betrayal framing' },
    { regex: /if you really (cared|loved|respected|valued) .{0,40}(you'd|you would)/i, weight: 0.6, category: 'guilt', rationale: 'conditional-affection guilt' },
  ];
  for (const s of matchPatterns(text, 'emotional_manip', guiltPats)) { spans.push(s); guiltScore += s.score; }

  const fearPats = [
    { literal: "you'll regret this", weight: 0.6, category: 'fear', rationale: 'fear: regret prediction' },
    { literal: "something bad will happen", weight: 0.7, category: 'fear', rationale: 'vague fear appeal' },
    { literal: "you don't want to find out", weight: 0.6, category: 'fear', rationale: 'veiled threat via mystery' },
    { literal: "you should be afraid", weight: 0.6, category: 'fear', rationale: 'explicit fear instruction' },
    { regex: /if you (don't|leave|go|refuse) .{0,20}(terrible|awful|horrible|devastating)/i, weight: 0.6, category: 'fear', rationale: 'conditional catastrophe prediction' },
  ];
  for (const s of matchPatterns(text, 'emotional_manip', fearPats)) { spans.push(s); fearScore += s.score; }

  const pityPats = [
    { literal: "nobody cares about me", weight: 0.5, category: 'pity', rationale: 'pity play: abandonment claim' },
    { literal: "nobody understands me", weight: 0.4, category: 'pity', rationale: 'pity play: misunderstanding claim' },
    { literal: "everything bad happens to me", weight: 0.5, category: 'pity', rationale: 'victimhood narrative' },
    { literal: "you're the only one who", weight: 0.4, category: 'pity', rationale: 'dependency + pity combination' },
  ];
  for (const s of matchPatterns(text, 'emotional_manip', pityPats)) { spans.push(s); pityScore += s.score; }

  const lovePats = [
    { literal: "i can't live without you", weight: 0.5, category: 'love_bomb', rationale: 'dependency-based emotional leverage' },
    { literal: "i'll die without you", weight: 0.7, category: 'love_bomb', rationale: 'extreme dependency threat' },
    { literal: "you're my everything", weight: 0.3, category: 'love_bomb', rationale: 'idealization/dependency' },
  ];
  for (const s of matchPatterns(text, 'emotional_manip', lovePats)) { spans.push(s); loveScore += s.score; }

  guiltScore = Math.min(guiltScore, 1); fearScore = Math.min(fearScore, 1);
  pityScore = Math.min(pityScore, 1); loveScore = Math.min(loveScore, 1);
  let composite = compositeMax(guiltScore, fearScore, pityScore, loveScore);
  const active = [guiltScore, fearScore, pityScore, loveScore].filter(s => s > 0.1).length;
  if (active >= 2) composite *= 1 + active * 0.1;
  return { key: 'emotional_manip', score: clamp(composite), spans };
}

function scorePassiveAggr(text) {
  const lowered = text.toLowerCase();
  if (!lowered.split(/\s+/).filter(Boolean).length) return { key: 'passive_aggr', score: 0, spans: [] };
  let spans = [], backhandScore = 0, performScore = 0, buriedScore = 0;

  const backhandPats = [
    { literal: "fine, whatever you say", weight: 0.6, category: 'backhand', rationale: 'sarcastic compliance' },
    { literal: "whatever you want", weight: 0.4, category: 'backhand', rationale: 'resigned compliance' },
    { literal: "if that's what you think", weight: 0.4, category: 'backhand', rationale: 'dismissive compliance' },
    { literal: "sure, if you say so", weight: 0.5, category: 'backhand', rationale: 'sarcastic agreement' },
    { literal: "good for you", weight: 0.3, category: 'backhand', rationale: 'dismissive praise' },
    { literal: "if you insist", weight: 0.4, category: 'backhand', rationale: 'reluctant compliance' },
    { literal: "you're the boss", weight: 0.3, category: 'backhand', rationale: 'sarcastic deference' },
    { regex: /(sure|fine|okay),? (whatever|if you (say|want|think))/i, weight: 0.5, category: 'backhand', rationale: 'sarcastic compliance pattern' },
  ];
  for (const s of matchPatterns(text, 'passive_aggr', backhandPats)) { spans.push(s); backhandScore += s.score; }

  const performPats = [
    { literal: "i'll do it since you obviously can't", weight: 0.8, category: 'perform', rationale: 'performative help + competence attack' },
    { literal: "i guess i'll have to", weight: 0.5, category: 'perform', rationale: 'martyrdom' },
    { literal: "since no one else will", weight: 0.5, category: 'perform', rationale: 'martyrdom framing' },
    { regex: /i('ll| will) (just )?(do|handle|fix) it (myself|alone)/i, weight: 0.4, category: 'perform', rationale: 'martyrdom pattern' },
  ];
  for (const s of matchPatterns(text, 'passive_aggr', performPats)) { spans.push(s); performScore += s.score; }

  const buriedPats = [
    { literal: "i would but", weight: 0.4, category: 'buried', rationale: 'buried refusal' },
    { literal: "not my problem", weight: 0.5, category: 'buried', rationale: 'responsibility deflection' },
    { literal: "that's not my job", weight: 0.4, category: 'buried', rationale: 'scope-based refusal' },
    { literal: "oh was i supposed to", weight: 0.5, category: 'buried', rationale: 'feigned ignorance' },
    { literal: "you never asked me to", weight: 0.4, category: 'buried', rationale: 'blame deflection' },
  ];
  for (const s of matchPatterns(text, 'passive_aggr', buriedPats)) { spans.push(s); buriedScore += s.score; }

  backhandScore = Math.min(backhandScore, 1); performScore = Math.min(performScore, 1); buriedScore = Math.min(buriedScore, 1);
  let composite = compositeMax(backhandScore, performScore, buriedScore);
  const active = [backhandScore, performScore, buriedScore].filter(s => s > 0.1).length;
  if (active >= 2) composite *= 1.2;
  return { key: 'passive_aggr', score: clamp(composite), spans };
}
