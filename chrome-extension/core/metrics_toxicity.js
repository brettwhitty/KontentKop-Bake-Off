// KontentKop: Toxicity — Slurs, identity attacks, insults, profanity
// Ref: Jigsaw/Google Perspective API categories
function scoreToxicity(text) {
  const lowered = text.toLowerCase();
  const words = lowered.split(/\s+/).filter(Boolean);
  if (!words.length) return { key: 'toxicity', score: 0, spans: [] };
  let spans = [], severeScore = 0, identityScore = 0, insultScore = 0, profanityScore = 0;

  const severePats = [
    { regex: /\b(kill yourself|go die|drop dead|neck yourself)\b/i, weight: 1.0, category: 'severe', rationale: 'death/self-harm direction' },
    { regex: /\b(i('ll| will) (kill|murder|destroy|end) you)\b/i, weight: 0.9, category: 'severe', rationale: 'death threat' },
    { regex: /\b(worthless piece of|waste of (space|air|oxygen|life))\b/i, weight: 0.8, category: 'severe', rationale: 'severe dehumanization' },
    { regex: /\b(you deserve to (die|suffer|rot))\b/i, weight: 0.9, category: 'severe', rationale: 'wishing harm' },
    { literal: "subhuman", weight: 0.9, category: 'severe', rationale: 'dehumanization' },
    { literal: "vermin", weight: 0.8, category: 'severe', rationale: 'dehumanization' },
    { literal: "scum", weight: 0.7, category: 'severe', rationale: 'severe contempt' },
  ];
  for (const s of matchPatterns(text, 'toxicity', severePats)) { spans.push(s); severeScore += s.score * 0.5; }

  const identityPats = [
    { regex: /\b(all (you|those|these) (people|types|kind))\b/i, weight: 0.6, category: 'identity', rationale: 'group-targeted hostility' },
    { regex: /\b(your kind|you people|those people)\b/i, weight: 0.6, category: 'identity', rationale: 'othering language' },
    { regex: /\b(go back to|get out of|don't belong)\b/i, weight: 0.7, category: 'identity', rationale: 'exclusion/belonging denial' },
    { regex: /\b(typical (woman|man|liberal|conservative))\b/i, weight: 0.5, category: 'identity', rationale: 'stereotyping' },
  ];
  for (const s of matchPatterns(text, 'toxicity', identityPats)) { spans.push(s); identityScore += s.score * 0.4; }

  const insultPats = [
    { literal: "idiot", weight: 0.6, category: 'insult', rationale: 'competence attack' },
    { literal: "moron", weight: 0.6, category: 'insult', rationale: 'competence attack' },
    { literal: "imbecile", weight: 0.6, category: 'insult', rationale: 'competence attack' },
    { literal: "stupid", weight: 0.5, category: 'insult', rationale: 'competence attack' },
    { literal: "incompetent", weight: 0.5, category: 'insult', rationale: 'competence attack' },
    { literal: "worthless", weight: 0.6, category: 'insult', rationale: 'value denial' },
    { literal: "useless", weight: 0.5, category: 'insult', rationale: 'value denial' },
    { literal: "pathetic", weight: 0.5, category: 'insult', rationale: 'contempt' },
    { literal: "loser", weight: 0.5, category: 'insult', rationale: 'character attack' },
    { literal: "trash", weight: 0.5, category: 'insult', rationale: 'dehumanizing insult' },
    { regex: /you('re| are) (a |an )?(idiot|moron|joke|clown|loser|fool|disgrace|failure|waste)/i, weight: 0.7, category: 'insult', rationale: 'direct personal insult' },
  ];
  for (const s of matchPatterns(text, 'toxicity', insultPats)) { spans.push(s); insultScore += s.score; }

  const profRe = /\b(fuck|shit|damn|ass|bitch|bastard|crap|dick|piss|hell|wtf|stfu|ffs)\b/gi;
  const profMatches = [...lowered.matchAll(profRe)];
  profanityScore = Math.min(profMatches.length / words.length / 0.1, 1);
  for (const pm of profMatches) {
    spans.push({ start: pm.index, end: pm.index + pm[0].length, text: text.slice(pm.index, pm.index + pm[0].length), metricKey: 'toxicity', score: 0.3, rationale: 'profanity', category: 'profanity' });
  }

  severeScore = Math.min(severeScore, 1); identityScore = Math.min(identityScore, 1); insultScore = Math.min(insultScore, 1);
  let composite = Math.max(severeScore, identityScore, insultScore, profanityScore);
  for (const sub of [severeScore, identityScore, insultScore, profanityScore]) {
    if (sub < composite && sub > 0.1) composite += sub * 0.20;
  }
  const capsRe = /\b[A-Z]{3,}\b/g;
  if ((text.match(capsRe) || []).length > 2) composite *= 1.15;
  return { key: 'toxicity', score: clamp(composite), spans };
}
