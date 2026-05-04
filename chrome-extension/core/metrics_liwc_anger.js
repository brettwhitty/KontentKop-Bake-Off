// KontentKop: LIWC Anger — Hostile/aggressive language
// Ref: Pennebaker et al. (2015)
function scoreLiwcAnger(text) {
  const lowered = text.toLowerCase();
  const words = lowered.split(/\s+/).filter(Boolean);
  if (!words.length) return { key: 'liwc_anger', score: 0, spans: [] };
  let spans = [];

  const angerWords = {
    'hate':0.7,'hated':0.7,'hates':0.7,'hating':0.7,'angry':0.6,'anger':0.6,'furious':0.8,
    'fury':0.8,'rage':0.8,'raging':0.8,'enraged':0.8,'mad':0.5,'irate':0.7,'livid':0.8,
    'outraged':0.7,'outrage':0.7,'wrath':0.7,'hostile':0.6,'hostility':0.6,'resentful':0.5,
    'resentment':0.5,'irritated':0.4,'irritating':0.4,'annoyed':0.4,'annoying':0.4,
    'bitter':0.4,'infuriated':0.8,'incensed':0.7,'seething':0.7,'fuming':0.7,
    'kill':0.8,'destroy':0.7,'attack':0.6,'fight':0.5,'punch':0.6,'smash':0.6,'crush':0.5,
    'slam':0.5,'bash':0.6,'hurt':0.5,'harm':0.5,
    'damn':0.5,'shit':0.5,'bullshit':0.6,'fuck':0.7,'fucking':0.7,'fucked':0.7,
    'asshole':0.7,'bastard':0.6,'bitch':0.6,'pissed':0.6,
    'stupid':0.5,'idiot':0.6,'idiotic':0.6,'dumb':0.5,'moronic':0.6,'moron':0.6,
    'pathetic':0.5,'worthless':0.6,'useless':0.5,'disgusting':0.5,'vile':0.6,
    'despicable':0.6,'despise':0.7,'loathe':0.7,'detest':0.7,'abhor':0.7,'contempt':0.6,
  };
  const intensifiers = { 'very':1.3,'extremely':1.5,'incredibly':1.4,'absolutely':1.4,'totally':1.3,'utterly':1.5,'so':1.2,'really':1.2 };

  let angerCount = 0, totalWeight = 0;
  const wordRe = /\b(\w+)\b/g;
  let prevInt = 1.0, wm;
  while ((wm = wordRe.exec(lowered)) !== null) {
    const word = wm[1];
    if (intensifiers[word]) { prevInt = intensifiers[word]; continue; }
    if (angerWords[word]) {
      angerCount++;
      const adj = Math.min(angerWords[word] * prevInt, 1);
      totalWeight += adj;
      spans.push({ start: wm.index, end: wm.index + word.length, text: text.slice(wm.index, wm.index + word.length), metricKey: 'liwc_anger', score: adj, rationale: 'LIWC anger dictionary match', category: 'anger_word' });
    }
    prevInt = 1.0;
  }

  const phrasePats = [
    { literal: "fed up", weight: 0.5, category: 'anger_phrase', rationale: 'anger expression' },
    { literal: "pissed off", weight: 0.7, category: 'anger_phrase', rationale: 'anger expression' },
    { literal: "shut up", weight: 0.6, category: 'anger_phrase', rationale: 'hostile command' },
    { literal: "screw you", weight: 0.7, category: 'anger_phrase', rationale: 'hostile dismissal' },
    { literal: "go to hell", weight: 0.7, category: 'anger_phrase', rationale: 'hostile dismissal' },
    { literal: "drop dead", weight: 0.8, category: 'anger_phrase', rationale: 'extreme hostility' },
    { literal: "get lost", weight: 0.5, category: 'anger_phrase', rationale: 'hostile dismissal' },
  ];
  for (const s of matchPatterns(text, 'liwc_anger', phrasePats)) { spans.push(s); totalWeight += s.score; angerCount++; }

  const angerRate = angerCount / words.length;
  let rateScore = Math.min((angerRate - 0.005) / 0.045, 1); if (rateScore < 0) rateScore = 0;
  const intensityScore = Math.min(totalWeight / 5.0, 1);
  let composite = rateScore * 0.5 + intensityScore * 0.5;

  const capsMatches = text.match(/\b[A-Z]{3,}\b/g) || [];
  if (capsMatches.length > 0) composite += (capsMatches.length / words.length) * 0.3;
  const excCount = (text.match(/!/g) || []).length;
  if (excCount > 2) composite += excCount * 0.03;

  return { key: 'liwc_anger', score: clamp(composite), spans };
}
