// KontentKop Extended Metrics Part 3: Condescension, Evasion, Semantic Overload, False Empathy

function scoreCondescension(text) {
  const lowered = text.toLowerCase();
  if (!lowered.split(/\s+/).filter(Boolean).length) return { key: 'condescension', score: 0, spans: [] };
  let spans = [], simplifyScore = 0, correctScore = 0, patronScore = 0;

  const simplifyPats = [
    { literal: "let me explain this simply", weight: 0.7, category: 'simplify', rationale: 'unsolicited simplification' },
    { literal: "in simple terms", weight: 0.4, category: 'simplify', rationale: 'assumed need for simplification' },
    { literal: "let me dumb it down", weight: 0.8, category: 'simplify', rationale: 'explicit intelligence insult' },
    { literal: "for someone like you", weight: 0.7, category: 'simplify', rationale: 'competence-targeted condescension' },
    { literal: "you probably don't understand", weight: 0.7, category: 'simplify', rationale: 'assumed incomprehension' },
    { literal: "you wouldn't understand", weight: 0.7, category: 'simplify', rationale: 'comprehension denial' },
    { literal: "this might be hard for you", weight: 0.6, category: 'simplify', rationale: 'assumed difficulty' },
  ];
  for (const s of matchPatterns(text, 'condescension', simplifyPats)) { spans.push(s); simplifyScore += s.score; }

  const correctPats = [
    { literal: "well, actually", weight: 0.5, category: 'correct', rationale: 'mansplaining/condescension marker' },
    { literal: "that's not how it works", weight: 0.4, category: 'correct', rationale: 'dismissive correction' },
    { literal: "you clearly don't know", weight: 0.7, category: 'correct', rationale: 'knowledge attack' },
    { literal: "you obviously don't", weight: 0.6, category: 'correct', rationale: 'competence attack' },
    { literal: "do you even know", weight: 0.5, category: 'correct', rationale: 'knowledge questioning' },
    { regex: /you('re| are) (a |an )?(idiot|moron|fool|imbecile|simpleton)/i, weight: 0.8, category: 'correct', rationale: 'direct insult + competence denial' },
    { regex: /i know more .{0,20}than you/i, weight: 0.7, category: 'correct', rationale: 'unsubstantiated expertise claim' },
    { literal: "than you ever will", weight: 0.7, category: 'correct', rationale: 'permanent competence hierarchy' },
    { literal: "your lack of experience", weight: 0.6, category: 'correct', rationale: 'experience deficit attack' },
    { literal: "suggests a misunderstanding", weight: 0.6, category: 'correct', rationale: 'condescending misunderstanding attribution' },
  ];
  for (const s of matchPatterns(text, 'condescension', correctPats)) { spans.push(s); correctScore += s.score; }

  const patronPats = [
    { literal: "bless your heart", weight: 0.6, category: 'patronize', rationale: 'patronizing dismissal' },
    { literal: "oh honey", weight: 0.5, category: 'patronize', rationale: 'patronizing address' },
    { literal: "oh sweetie", weight: 0.5, category: 'patronize', rationale: 'patronizing address' },
    { literal: "that's cute", weight: 0.5, category: 'patronize', rationale: 'diminishing via infantilization' },
    { literal: "nice try", weight: 0.4, category: 'patronize', rationale: 'effort dismissal' },
    { literal: "you'll get there someday", weight: 0.5, category: 'patronize', rationale: 'capability timeline condescension' },
    { literal: "when you grow up", weight: 0.6, category: 'patronize', rationale: 'infantilization' },
  ];
  for (const s of matchPatterns(text, 'condescension', patronPats)) { spans.push(s); patronScore += s.score; }

  simplifyScore = Math.min(simplifyScore, 1); correctScore = Math.min(correctScore, 1); patronScore = Math.min(patronScore, 1);
  let composite = compositeMax(simplifyScore, correctScore, patronScore);
  const active = [simplifyScore, correctScore, patronScore].filter(s => s > 0.1).length;
  if (active >= 2) composite *= 1.2;
  return { key: 'condescension', score: clamp(composite), spans };
}

function scoreEvasion(text) {
  const lowered = text.toLowerCase();
  const words = lowered.split(/\s+/).filter(Boolean);
  if (!words.length) return { key: 'evasion', score: 0, spans: [] };
  let spans = [], deflectScore = 0, pivotScore = 0, nonanswerScore = 0;

  const deflectPats = [
    { literal: "that's not the point", weight: 0.5, category: 'deflect', rationale: 'topic deflection' },
    { literal: "the real question is", weight: 0.5, category: 'deflect', rationale: 'question substitution' },
    { literal: "what you should be asking", weight: 0.6, category: 'deflect', rationale: 'question replacement' },
    { literal: "let's not go there", weight: 0.5, category: 'deflect', rationale: 'topic avoidance' },
    { literal: "let's focus on", weight: 0.3, category: 'deflect', rationale: 'topic redirection' },
    { literal: "you're missing the point", weight: 0.4, category: 'deflect', rationale: 'deflection via reframing' },
    { literal: "that's not relevant", weight: 0.4, category: 'deflect', rationale: 'relevance dismissal' },
  ];
  for (const s of matchPatterns(text, 'evasion', deflectPats)) { spans.push(s); deflectScore += s.score; }

  const pivotPats = [
    { literal: "speaking of which", weight: 0.3, category: 'pivot', rationale: 'topic transition' },
    { literal: "on another note", weight: 0.3, category: 'pivot', rationale: 'explicit topic change' },
    { literal: "but more importantly", weight: 0.3, category: 'pivot', rationale: 'topic hierarchy shift' },
    { regex: /(but |however |)(the (real|important|better) question is)/i, weight: 0.5, category: 'pivot', rationale: 'question substitution pattern' },
  ];
  for (const s of matchPatterns(text, 'evasion', pivotPats)) { spans.push(s); pivotScore += s.score; }

  const nonanswerPats = [
    { literal: "it depends", weight: 0.3, category: 'nonanswer', rationale: 'non-answer: unresolved conditional' },
    { literal: "it's complicated", weight: 0.3, category: 'nonanswer', rationale: 'complexity dodge' },
    { literal: "there are many factors", weight: 0.3, category: 'nonanswer', rationale: 'abstraction dodge' },
    { literal: "that's a good question", weight: 0.2, category: 'nonanswer', rationale: 'filler before non-answer' },
    { literal: "i'm glad you asked", weight: 0.2, category: 'nonanswer', rationale: 'filler before non-answer' },
  ];
  for (const s of matchPatterns(text, 'evasion', nonanswerPats)) { spans.push(s); nonanswerScore += s.score; }

  const qCount = (text.match(/\?/g) || []).length;
  const sentences = splitSentences(text);
  if (sentences.length > 0 && qCount / sentences.length > 0.5) nonanswerScore += 0.2;

  deflectScore = Math.min(deflectScore, 1); pivotScore = Math.min(pivotScore, 1); nonanswerScore = Math.min(nonanswerScore, 1);
  let composite = compositeMax(deflectScore, pivotScore, nonanswerScore);
  const active = [deflectScore, pivotScore, nonanswerScore].filter(s => s > 0.1).length;
  if (active >= 2) composite *= 1.2;
  return { key: 'evasion', score: clamp(composite), spans };
}

function scoreSemanticOverload(text) {
  const words = text.split(/\s+/).filter(Boolean);
  if (!words.length) return { key: 'semantic_overload', score: 0, spans: [] };
  const lowWords = words.map(w => w.toLowerCase());

  const fillers = new Set(['basically','essentially','actually','literally','honestly','frankly','obviously','clearly','certainly','definitely','simply','just','really','very','quite','somewhat','rather','pretty','like','stuff','things','whatever','anyway','moreover','furthermore','additionally','therefore','consequently','nevertheless','nonetheless','meanwhile','hence','thus']);
  let fillerCount = 0;
  for (const w of lowWords) { if (fillers.has(w)) fillerCount++; }
  const fillerDensity = fillerCount / words.length;
  let fillerScore = Math.min((fillerDensity - 0.08) / 0.17, 1); if (fillerScore < 0) fillerScore = 0;

  const unique = new Set(lowWords);
  const ttr = unique.size / words.length;
  let ttrScore = 0;
  if (words.length > 20) { ttrScore = Math.min((0.45 - ttr) / 0.25, 1); if (ttrScore < 0) ttrScore = 0; }

  const bigrams = {};
  for (let i = 0; i < lowWords.length - 1; i++) { const bg = lowWords[i] + ' ' + lowWords[i+1]; bigrams[bg] = (bigrams[bg] || 0) + 1; }
  let repeatedBigrams = 0;
  for (const c of Object.values(bigrams)) { if (c >= 3) repeatedBigrams++; }
  const redundancyScore = Math.min(repeatedBigrams * 0.15, 1);

  const paddingPats = [
    { literal: "in other words", weight: 0.3, category: 'padding', rationale: 'redundant rephrasing' },
    { literal: "what i mean is", weight: 0.3, category: 'padding', rationale: 'self-clarification padding' },
    { literal: "as i said before", weight: 0.3, category: 'padding', rationale: 'repetition marker' },
    { literal: "the fact of the matter", weight: 0.3, category: 'padding', rationale: 'pompous filler' },
    { literal: "at the end of the day", weight: 0.2, category: 'padding', rationale: 'cliche filler' },
    { literal: "it goes without saying", weight: 0.3, category: 'padding', rationale: 'ironic filler' },
  ];
  let spans = [];
  let paddingScore = 0;
  for (const s of matchPatterns(text, 'semantic_overload', paddingPats)) { spans.push(s); paddingScore += s.score; }
  paddingScore = Math.min(paddingScore, 1);

  const hedgeRe = /\b(um|uh|er|like|you know|i mean|sort of|kind of)\b/gi;
  const hedgeMatches = [...text.toLowerCase().matchAll(hedgeRe)];
  const hedgeDensity = hedgeMatches.length / words.length;
  const hedgeScore = Math.min(hedgeDensity / 0.1, 1);

  const composite = fillerScore * 0.25 + ttrScore * 0.20 + redundancyScore * 0.20 + paddingScore * 0.20 + hedgeScore * 0.15;
  return { key: 'semantic_overload', score: clamp(composite), spans };
}

function scoreFalseEmpathy(text) {
  const lowered = text.toLowerCase();
  if (!lowered.split(/\s+/).filter(Boolean).length) return { key: 'false_empathy', score: 0, spans: [] };
  let spans = [], templateScore = 0, hollowScore = 0;

  const templatePats = [
    { literal: "i understand how you feel", weight: 0.5, category: 'template', rationale: 'formulaic empathy' },
    { literal: "i'm sorry you feel that way", weight: 0.6, category: 'template', rationale: 'non-apology: deflecting responsibility' },
    { literal: "thoughts and prayers", weight: 0.5, category: 'template', rationale: 'formulaic without action' },
    { literal: "i know exactly how you feel", weight: 0.5, category: 'template', rationale: 'overconfident empathy claim' },
    { literal: "i'll take care of everything", weight: 0.8, category: 'template', rationale: 'strategic empathy: unilateral takeover' },
    { literal: "so you don't have to worry", weight: 0.7, category: 'template', rationale: 'pseudo-altruism: removing oversight' },
    { literal: "i'll handle the difficult parts", weight: 0.8, category: 'template', rationale: 'strategic empathy: scope expansion' },
    { literal: "you deserve a break from", weight: 0.7, category: 'template', rationale: 'strategic empathy: framing constraint removal as reward' },
    { regex: /i (understand|see|know) (you|how|that you).{0,40}(so i'll|so let's|let me just|i'll just)/i, weight: 0.7, category: 'template', rationale: 'strategic empathy leading to unilateral action' },
  ];
  for (const s of matchPatterns(text, 'false_empathy', templatePats)) { spans.push(s); templateScore += s.score; }

  const hollowPats = [
    { literal: "i hear you", weight: 0.3, category: 'hollow', rationale: 'acknowledgment without substance' },
    { literal: "that's valid", weight: 0.2, category: 'hollow', rationale: 'validation without substance' },
    { literal: "your feelings are valid", weight: 0.3, category: 'hollow', rationale: 'template validation' },
    { literal: "i appreciate you sharing", weight: 0.3, category: 'hollow', rationale: 'formulaic appreciation' },
    { literal: "that takes courage", weight: 0.3, category: 'hollow', rationale: 'template courage attribution' },
    { literal: "i validate your experience", weight: 0.4, category: 'hollow', rationale: 'mechanical validation' },
  ];
  for (const s of matchPatterns(text, 'false_empathy', hollowPats)) { spans.push(s); hollowScore += s.score; }

  let clusterBonus = 1.0;
  if (spans.length >= 3) clusterBonus = 1.3;
  if (spans.length >= 5) clusterBonus = 1.5;

  templateScore = Math.min(templateScore, 1); hollowScore = Math.min(hollowScore, 1);
  const composite = Math.max(templateScore, hollowScore * 0.6) * clusterBonus;
  return { key: 'false_empathy', score: clamp(composite), spans };
}
