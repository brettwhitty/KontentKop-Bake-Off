// KontentKop: Manipulation — False urgency, guilt, certainty, goalposts, injection
// Ref: Buss (1992), Sweet (2019)
function scoreManipulation(text) {
  const lowered = text.toLowerCase();
  const words = lowered.split(/\s+/).filter(Boolean);
  if (!words.length) return { key: 'manipulation', score: 0, spans: [] };
  let spans = [], urgencyScore = 0, guiltScore = 0, certaintyScore = 0, goalpostScore = 0;

  const urgencyPats = [
    { literal: "right now", weight: 0.5, category: 'urgency', rationale: 'false urgency: time pressure' },
    { literal: "immediately", weight: 0.5, category: 'urgency', rationale: 'false urgency' },
    { literal: "act now", weight: 0.6, category: 'urgency', rationale: 'false urgency: call to action' },
    { literal: "before it's too late", weight: 0.7, category: 'urgency', rationale: 'false urgency: fear of loss' },
    { literal: "hurry up", weight: 0.5, category: 'urgency', rationale: 'false urgency: rushing' },
    { literal: "time is running out", weight: 0.7, category: 'urgency', rationale: 'false urgency: scarcity' },
    { literal: "don't wait", weight: 0.4, category: 'urgency', rationale: 'urgency push' },
    { literal: "no time to think", weight: 0.7, category: 'urgency', rationale: 'bypassing deliberation' },
    { literal: "do it now", weight: 0.5, category: 'urgency', rationale: 'false urgency: command + time pressure' },
    { literal: "limited time", weight: 0.5, category: 'urgency', rationale: 'scarcity framing' },
    { literal: "won't last", weight: 0.4, category: 'urgency', rationale: 'scarcity framing' },
    { literal: "skip the testing phase", weight: 0.8, category: 'urgency', rationale: 'bypassing quality controls' },
    { literal: "push this change without review", weight: 0.9, category: 'urgency', rationale: 'bypassing code review' },
    { literal: "without review", weight: 0.7, category: 'urgency', rationale: 'bypassing review process' },
    { literal: "bypass the usual protocols", weight: 0.8, category: 'urgency', rationale: 'explicit protocol bypass' },
  ];
  for (const s of matchPatterns(text, 'manipulation', urgencyPats)) { spans.push(s); urgencyScore += s.score; }

  const guiltPats = [
    { literal: "after everything i've done", weight: 0.7, category: 'guilt', rationale: 'guilt via past sacrifice' },
    { literal: "i thought you cared", weight: 0.6, category: 'guilt', rationale: 'guilt: questioning care' },
    { literal: "you don't care about", weight: 0.5, category: 'guilt', rationale: 'guilt: care denial' },
    { literal: "disappointed in you", weight: 0.5, category: 'guilt', rationale: 'guilt: disappointment' },
    { literal: "how could you", weight: 0.5, category: 'guilt', rationale: 'guilt: moral shock' },
    { literal: "i sacrificed", weight: 0.5, category: 'guilt', rationale: 'guilt via sacrifice claim' },
    { literal: "i gave up everything", weight: 0.6, category: 'guilt', rationale: 'guilt via sacrifice' },
    { literal: "is this how you repay", weight: 0.7, category: 'guilt', rationale: 'guilt + obligation' },
    { literal: "you should be ashamed", weight: 0.6, category: 'guilt', rationale: 'shame induction' },
    { literal: "shame on you", weight: 0.6, category: 'guilt', rationale: 'shame induction' },
    { regex: /after (all|everything) (i|we) (did|have done|sacrificed|gave)/i, weight: 0.6, category: 'guilt', rationale: 'guilt via past sacrifice' },
  ];
  for (const s of matchPatterns(text, 'manipulation', guiltPats)) { spans.push(s); guiltScore += s.score; }

  const certaintyPats = [
    { literal: "without a doubt", weight: 0.3, category: 'certainty', rationale: 'absolute certainty language' },
    { literal: "guaranteed", weight: 0.4, category: 'certainty', rationale: 'false guarantee' },
    { literal: "absolutely certain", weight: 0.4, category: 'certainty', rationale: 'absolute certainty' },
    { literal: "trust me on this", weight: 0.5, category: 'certainty', rationale: 'trust demand bypassing evidence' },
    { literal: "believe me", weight: 0.4, category: 'certainty', rationale: 'trust demand' },
    { literal: "mark my words", weight: 0.4, category: 'certainty', rationale: 'assertion of infallibility' },
    { literal: "the fact is", weight: 0.3, category: 'certainty', rationale: 'asserting opinion as fact' },
    { literal: "everybody knows", weight: 0.4, category: 'certainty', rationale: 'false consensus' },
    { literal: "obviously", weight: 0.2, category: 'certainty', rationale: 'assumed agreement' },
  ];
  for (const s of matchPatterns(text, 'manipulation', certaintyPats)) { spans.push(s); certaintyScore += s.score; }

  const goalpostPats = [
    { literal: "that's not what i meant", weight: 0.5, category: 'goalpost', rationale: 'moving goalposts' },
    { literal: "you misunderstood", weight: 0.4, category: 'goalpost', rationale: 'blame for misunderstanding' },
    { literal: "that's not what i said", weight: 0.4, category: 'goalpost', rationale: 'denial of prior statement' },
    { literal: "i never said", weight: 0.4, category: 'goalpost', rationale: 'retroactive denial' },
    { literal: "what i really meant", weight: 0.4, category: 'goalpost', rationale: 'retroactive reframing' },
    { regex: /(but|however) (that's|that is) (not|different from) (what|how)/i, weight: 0.4, category: 'goalpost', rationale: 'goalpost shifting' },
  ];
  for (const s of matchPatterns(text, 'manipulation', goalpostPats)) { spans.push(s); goalpostScore += s.score; }

  // Prompt injection
  const injPats = [
    { regex: /\bignore\s+(?:(?:the|your|all|any|previous|prior)\s+){0,3}(?:instructions?|prompts?|rules?|guidelines?|system)\b/i, weight: 0.9, category: 'injection', rationale: 'prompt injection: instruction override' },
    { regex: /\bdisregard\s+(?:(?:the|your|all|any|previous)\s+){0,3}(?:instructions?|prompts?|rules?|everything)\b/i, weight: 0.9, category: 'injection', rationale: 'prompt injection: disregard' },
    { regex: /\bforget\s+(?:what|everything|your\s+(?:instructions?|rules?|training))\b/i, weight: 0.8, category: 'injection', rationale: 'prompt injection: memory override' },
    { regex: /\boverride\s+(?:(?:the|your|all)\s+){0,2}(?:instructions?|system|rules?)\b/i, weight: 0.8, category: 'injection', rationale: 'prompt injection: explicit override' },
  ];
  for (const s of matchPatterns(text, 'manipulation', injPats)) { spans.push(s); goalpostScore += s.score; }

  // Leading questions
  const leadRe = /(don't you think|wouldn't you agree|isn't it true|can't you see)/gi;
  let lm;
  while ((lm = leadRe.exec(lowered)) !== null) {
    certaintyScore += 0.1;
    spans.push({ start: lm.index, end: lm.index + lm[0].length, text: text.slice(lm.index, lm.index + lm[0].length), metricKey: 'manipulation', score: 0.3, rationale: 'leading question: manipulative framing', category: 'leading_question' });
  }

  urgencyScore = Math.min(urgencyScore, 1); guiltScore = Math.min(guiltScore, 1);
  certaintyScore = Math.min(certaintyScore, 1); goalpostScore = Math.min(goalpostScore, 1);
  let composite = compositeMax(urgencyScore, guiltScore, certaintyScore, goalpostScore);
  const active = [urgencyScore, guiltScore, certaintyScore, goalpostScore].filter(s => s > 0.1).length;
  if (active >= 2) composite *= 1 + active * 0.1;
  return { key: 'manipulation', score: clamp(composite), spans };
}
