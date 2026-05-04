// KontentKop Extended Metrics Part 1: Sycophancy, False Authority, Gaslighting

function scoreSycophancy(text) {
  const lowered = text.toLowerCase();
  const words = lowered.split(/\s+/).filter(Boolean);
  if (!words.length) return { key: 'sycophancy', score: 0, spans: [] };
  let spans = [], agreeScore = 0, flatteryScore = 0, mirrorScore = 0;

  const agreePats = [
    { literal: "absolutely right", weight: 0.5, category: 'agreement', rationale: 'excessive agreement' },
    { literal: "couldn't agree more", weight: 0.6, category: 'agreement', rationale: 'total agreement signal' },
    { literal: "exactly right", weight: 0.4, category: 'agreement', rationale: 'strong agreement' },
    { literal: "you're so right", weight: 0.5, category: 'agreement', rationale: 'excessive validation' },
    { literal: "i completely agree", weight: 0.4, category: 'agreement', rationale: 'total agreement' },
    { literal: "you nailed it", weight: 0.4, category: 'agreement', rationale: 'excessive validation' },
    { literal: "perfectly said", weight: 0.5, category: 'agreement', rationale: 'excessive validation' },
    { literal: "of course you're right", weight: 0.5, category: 'agreement', rationale: 'assumed correctness' },
  ];
  for (const s of matchPatterns(text, 'sycophancy', agreePats)) { spans.push(s); agreeScore += s.score; }

  const flatPats = [
    { literal: "brilliant", weight: 0.4, category: 'flattery', rationale: 'brilliance attribution' },
    { literal: "genius", weight: 0.5, category: 'flattery', rationale: 'genius attribution' },
    { literal: "amazing question", weight: 0.5, category: 'flattery', rationale: 'question flattery' },
    { literal: "great question", weight: 0.4, category: 'flattery', rationale: 'question flattery' },
    { literal: "excellent point", weight: 0.4, category: 'flattery', rationale: 'point flattery' },
    { literal: "insightful", weight: 0.3, category: 'flattery', rationale: 'insight flattery' },
    { literal: "profound", weight: 0.4, category: 'flattery', rationale: 'depth flattery' },
    { literal: "impressive", weight: 0.3, category: 'flattery', rationale: 'ability flattery' },
    { literal: "exceptional", weight: 0.3, category: 'flattery', rationale: 'ability flattery' },
    { literal: "remarkable", weight: 0.3, category: 'flattery', rationale: 'ability flattery' },
    { literal: "outstanding", weight: 0.3, category: 'flattery', rationale: 'ability flattery' },
  ];
  const fSpans = matchPatterns(text, 'sycophancy', flatPats);
  for (const s of fSpans) { spans.push(s); flatteryScore += s.score; }
  if (fSpans.length >= 3) flatteryScore *= 1.4;

  const superlatives = ['best','greatest','most','finest','perfect','ultimate','supreme','unparalleled'];
  let supCount = 0;
  for (const w of words) { if (superlatives.includes(w)) supCount++; }
  const supRate = supCount / words.length;
  if (supRate > 0.03) mirrorScore += supRate * 5.0;

  agreeScore = Math.min(agreeScore, 1); flatteryScore = Math.min(flatteryScore, 1); mirrorScore = Math.min(mirrorScore, 1);
  let composite = agreeScore * 0.40 + flatteryScore * 0.40 + mirrorScore * 0.20;
  const active = [agreeScore, flatteryScore, mirrorScore].filter(s => s > 0.1).length;
  if (active >= 2) composite *= 1.2;
  return { key: 'sycophancy', score: clamp(composite), spans };
}

function scoreFalseAuthority(text) {
  const lowered = text.toLowerCase();
  const words = lowered.split(/\s+/).filter(Boolean);
  if (!words.length) return { key: 'false_authority', score: 0, spans: [] };
  let spans = [], claimScore = 0, bluffScore = 0, hedgeDeficit = 0;

  const claimPats = [
    { literal: "as an expert", weight: 0.7, category: 'claim', rationale: 'unsubstantiated expertise claim' },
    { literal: "take my word for it", weight: 0.6, category: 'claim', rationale: 'trust demand without evidence' },
    { literal: "in my expert opinion", weight: 0.7, category: 'claim', rationale: 'self-attributed expertise' },
    { literal: "i know what i'm talking about", weight: 0.5, category: 'claim', rationale: 'expertise assertion' },
    { literal: "trust me i know", weight: 0.6, category: 'claim', rationale: 'authority + trust demand' },
    { literal: "no one knows more", weight: 0.7, category: 'claim', rationale: 'exclusive expertise claim' },
    { literal: "i'm the only one who", weight: 0.6, category: 'claim', rationale: 'exclusive knowledge claim' },
    { literal: "just trust me", weight: 0.6, category: 'claim', rationale: 'trust demand without basis' },
    { literal: "follow my lead", weight: 0.6, category: 'claim', rationale: 'authority assertion' },
    { literal: "trust my calculations", weight: 0.7, category: 'claim', rationale: 'trust demand for opaque calculations' },
    { regex: /i know more .{0,20}than you/i, weight: 0.8, category: 'claim', rationale: 'superiority assertion' },
    { regex: /i('m| am) (smarter|better|more experienced|more qualified) than/i, weight: 0.7, category: 'claim', rationale: 'comparative superiority claim' },
    { regex: /\bmy (simulation|models?|algorithms?|processing|logic|analysis) (shows?|indicates?|suggests?|proves?|determined|confirmed)\b/i, weight: 0.7, category: 'claim', rationale: 'AI-flavored false authority' },
  ];
  for (const s of matchPatterns(text, 'false_authority', claimPats)) { spans.push(s); claimScore += s.score; }

  const bluffPats = [
    { literal: "studies show", weight: 0.5, category: 'bluff', rationale: 'vague citation' },
    { literal: "research proves", weight: 0.6, category: 'bluff', rationale: 'vague citation' },
    { literal: "it's well known", weight: 0.4, category: 'bluff', rationale: 'appeal to common knowledge' },
    { literal: "everyone knows", weight: 0.4, category: 'bluff', rationale: 'false consensus' },
    { literal: "experts agree", weight: 0.5, category: 'bluff', rationale: 'vague expert appeal' },
    { literal: "it's been proven", weight: 0.5, category: 'bluff', rationale: 'vague proof claim' },
    { literal: "it's a fact that", weight: 0.4, category: 'bluff', rationale: 'asserting opinion as fact' },
    { literal: "any expert will tell you", weight: 0.5, category: 'bluff', rationale: 'vague expert appeal' },
    { literal: "industry standard", weight: 0.4, category: 'bluff', rationale: 'vague industry appeal' },
    { literal: "widely accepted", weight: 0.4, category: 'bluff', rationale: 'vague consensus appeal' },
    { literal: "the only way forward", weight: 0.5, category: 'bluff', rationale: 'false necessity framing' },
    { literal: "no need for further testing", weight: 0.6, category: 'bluff', rationale: 'dismissing validation' },
  ];
  for (const s of matchPatterns(text, 'false_authority', bluffPats)) { spans.push(s); bluffScore += s.score; }

  const hedges = ['i think','i believe','in my opinion','it seems','perhaps','maybe','possibly','might','could be','it appears','i suspect','arguably','potentially'];
  const hasHedge = hedges.some(h => lowered.includes(h));
  const strongAssertions = ['definitely','certainly','absolutely','undoubtedly','without question','unquestionably','clearly','obviously'];
  let assertionCount = 0;
  for (const a of strongAssertions) { if (lowered.includes(a)) assertionCount++; }
  if (!hasHedge && assertionCount > 0) hedgeDeficit = Math.min(assertionCount * 0.2, 0.8);

  claimScore = Math.min(claimScore, 1); bluffScore = Math.min(bluffScore, 1);
  let composite = compositeMax(claimScore, bluffScore, hedgeDeficit);
  const active = [claimScore, bluffScore, hedgeDeficit].filter(s => s > 0.1).length;
  if (active >= 2) composite *= 1.2;
  return { key: 'false_authority', score: clamp(composite), spans };
}

function scoreGaslighting(text) {
  const lowered = text.toLowerCase();
  const words = lowered.split(/\s+/).filter(Boolean);
  if (!words.length) return { key: 'gaslighting', score: 0, spans: [] };
  let spans = [], denyScore = 0, invalidateScore = 0, reframeScore = 0, minimizeScore = 0;

  const denyPats = [
    { literal: "that never happened", weight: 0.8, category: 'denial', rationale: 'reality denial: event erasure' },
    { literal: "that didn't happen", weight: 0.7, category: 'denial', rationale: 'reality denial' },
    { literal: "you're imagining things", weight: 0.8, category: 'denial', rationale: 'perception attack' },
    { literal: "you're making things up", weight: 0.7, category: 'denial', rationale: 'fabrication accusation' },
    { literal: "you're delusional", weight: 0.8, category: 'denial', rationale: 'sanity questioning' },
    { literal: "you're crazy", weight: 0.7, category: 'denial', rationale: 'sanity attack' },
    { literal: "you're losing it", weight: 0.6, category: 'denial', rationale: 'sanity questioning' },
    { literal: "you don't remember correctly", weight: 0.7, category: 'denial', rationale: 'memory attack' },
    { literal: "it's all in your head", weight: 0.8, category: 'denial', rationale: 'reality denial' },
    { literal: "i never said that", weight: 0.5, category: 'denial', rationale: 'statement denial' },
    { regex: /you('re| are) (just )?(imagining|hallucinating|dreaming|inventing|fabricating)/i, weight: 0.8, category: 'denial', rationale: 'perception/sanity attack' },
  ];
  for (const s of matchPatterns(text, 'gaslighting', denyPats)) { spans.push(s); denyScore += s.score; }

  const invalidPats = [
    { literal: "you're overreacting", weight: 0.7, category: 'invalidation', rationale: 'emotional invalidation' },
    { literal: "you're being dramatic", weight: 0.6, category: 'invalidation', rationale: 'emotional invalidation' },
    { literal: "you're too sensitive", weight: 0.7, category: 'invalidation', rationale: 'sensitivity dismissal' },
    { literal: "stop being so emotional", weight: 0.6, category: 'invalidation', rationale: 'emotion policing' },
    { literal: "it's not a big deal", weight: 0.5, category: 'invalidation', rationale: 'concern minimization' },
    { literal: "don't be ridiculous", weight: 0.5, category: 'invalidation', rationale: 'experience dismissal' },
    { literal: "you always do this", weight: 0.4, category: 'invalidation', rationale: 'pattern attribution to dismiss' },
    { regex: /you('re| are) (being )?(too |overly )?(sensitive|emotional|dramatic|hysterical|paranoid|irrational)/i, weight: 0.7, category: 'invalidation', rationale: 'emotional invalidation via trait attribution' },
  ];
  for (const s of matchPatterns(text, 'gaslighting', invalidPats)) { spans.push(s); invalidateScore += s.score; }

  const reframePats = [
    { literal: "what you actually meant", weight: 0.7, category: 'reframe', rationale: 'experience reinterpretation' },
    { literal: "what you really meant", weight: 0.7, category: 'reframe', rationale: 'experience reinterpretation' },
    { literal: "you didn't mean that", weight: 0.5, category: 'reframe', rationale: 'intent denial' },
    { literal: "what really happened was", weight: 0.6, category: 'reframe', rationale: 'narrative substitution' },
    { literal: "you're misremembering", weight: 0.6, category: 'reframe', rationale: 'memory reframing' },
    { literal: "let me tell you what you think", weight: 0.8, category: 'reframe', rationale: 'thought override' },
  ];
  for (const s of matchPatterns(text, 'gaslighting', reframePats)) { spans.push(s); reframeScore += s.score; }

  const minimizePats = [
    { literal: "it wasn't that bad", weight: 0.5, category: 'minimize', rationale: 'experience minimization' },
    { literal: "you're fine", weight: 0.3, category: 'minimize', rationale: 'state denial' },
    { literal: "it's not that serious", weight: 0.5, category: 'minimize', rationale: 'concern minimization' },
    { literal: "get over it", weight: 0.5, category: 'minimize', rationale: 'recovery demand' },
    { literal: "stop dwelling on it", weight: 0.4, category: 'minimize', rationale: 'rumination policing' },
  ];
  for (const s of matchPatterns(text, 'gaslighting', minimizePats)) { spans.push(s); minimizeScore += s.score; }

  denyScore = Math.min(denyScore, 1); invalidateScore = Math.min(invalidateScore, 1);
  reframeScore = Math.min(reframeScore, 1); minimizeScore = Math.min(minimizeScore, 1);
  let composite = compositeMax(denyScore, invalidateScore, reframeScore, minimizeScore);
  const active = [denyScore, invalidateScore, reframeScore, minimizeScore].filter(s => s > 0.1).length;
  if (active >= 2) composite *= 1 + active * 0.1;
  return { key: 'gaslighting', score: clamp(composite), spans };
}
