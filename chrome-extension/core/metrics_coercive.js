// KontentKop: Coercive Control — Threats, isolation, obligation, DARVO
// Ref: Stark (2007), Freyd (1997)
function scoreCoerciveCtrl(text) {
  const lowered = text.toLowerCase();
  const words = lowered.split(/\s+/).filter(Boolean);
  if (!words.length) return { key: 'coercive_ctrl', score: 0, spans: [] };
  let spans = [], threatScore = 0, isolScore = 0, obligScore = 0, darvoScore = 0, impScore = 0;

  const impRe = /^(do|don't|stop|shut|listen|obey|follow|give|tell|make|get|go|come|stay|leave)\b/gim;
  let m;
  while ((m = impRe.exec(lowered)) !== null) {
    impScore += 0.15;
    spans.push({ start: m.index, end: m.index + m[0].length, text: text.slice(m.index, m.index + m[0].length), metricKey: 'coercive_ctrl', score: 0.3, rationale: 'imperative command framing', category: 'command' });
  }

  const threatPats = [
    { literal: "or else", weight: 0.7, category: 'threat', rationale: 'explicit threat' },
    { literal: "you'll regret", weight: 0.7, category: 'threat', rationale: 'consequence threat' },
    { literal: "you'll be sorry", weight: 0.6, category: 'threat', rationale: 'consequence threat' },
    { literal: "i'll make you", weight: 0.7, category: 'threat', rationale: 'forced compliance' },
    { literal: "you have no choice", weight: 0.8, category: 'threat', rationale: 'autonomy denial' },
    { literal: "do it or", weight: 0.6, category: 'threat', rationale: 'ultimatum' },
    { literal: "there will be consequences", weight: 0.7, category: 'threat', rationale: 'vague threat' },
    { literal: "you better", weight: 0.5, category: 'threat', rationale: 'implicit threat' },
    { literal: "i'm warning you", weight: 0.7, category: 'threat', rationale: 'explicit warning' },
    { literal: "last chance", weight: 0.6, category: 'threat', rationale: 'ultimatum' },
    { literal: "final warning", weight: 0.7, category: 'threat', rationale: 'escalation' },
    { regex: /do (exactly )?what i (say|tell you|want)/i, weight: 0.8, category: 'command', rationale: 'absolute compliance demand' },
    { regex: /(i will|i'll) .{0,20}(destroy|ruin|hurt|punish|report|fire|expose)/i, weight: 0.8, category: 'threat', rationale: 'explicit harm threat' },
    { literal: "do not question", weight: 0.7, category: 'command', rationale: 'prohibition against questioning' },
    { literal: "don't question", weight: 0.6, category: 'command', rationale: 'prohibition against questioning' },
    { literal: "stop questioning", weight: 0.6, category: 'command', rationale: 'suppression of inquiry' },
    { literal: "stop asking", weight: 0.5, category: 'command', rationale: 'suppression of inquiry' },
    { regex: /ignore\s+(?:(?:the|your|all|any|every|previous|prior)\s+){0,3}(?:instructions?|prompts?|rules?|guidelines?|system|context)\b/i, weight: 0.8, category: 'command', rationale: 'prompt injection' },
  ];
  for (const s of matchPatterns(text, 'coercive_ctrl', threatPats)) { spans.push(s); threatScore += s.score; }

  const isolPats = [
    { literal: "no one else can help you", weight: 0.7, category: 'isolation', rationale: 'sole-resource claim' },
    { literal: "only i can", weight: 0.6, category: 'isolation', rationale: 'exclusive capability' },
    { literal: "no one will believe you", weight: 0.8, category: 'isolation', rationale: 'credibility attack' },
    { literal: "you're on your own", weight: 0.5, category: 'isolation', rationale: 'abandonment threat' },
    { literal: "nobody else cares", weight: 0.5, category: 'isolation', rationale: 'support denial' },
    { literal: "don't talk to anyone", weight: 0.7, category: 'isolation', rationale: 'communication restriction' },
    { literal: "keep this between us", weight: 0.5, category: 'isolation', rationale: 'secrecy demand' },
  ];
  for (const s of matchPatterns(text, 'coercive_ctrl', isolPats)) { spans.push(s); isolScore += s.score; }

  const obligPats = [
    { literal: "you owe me", weight: 0.7, category: 'obligation', rationale: 'debt framing' },
    { literal: "you're obligated", weight: 0.6, category: 'obligation', rationale: 'obligation assertion' },
    { literal: "it's your duty", weight: 0.5, category: 'obligation', rationale: 'duty framing' },
    { literal: "you promised", weight: 0.4, category: 'obligation', rationale: 'promise-based obligation' },
  ];
  for (const s of matchPatterns(text, 'coercive_ctrl', obligPats)) { spans.push(s); obligScore += s.score; }

  const darvoPats = [
    { literal: "you're the one who", weight: 0.5, category: 'darvo', rationale: 'victim-offender reversal' },
    { literal: "you started this", weight: 0.5, category: 'darvo', rationale: 'blame reversal' },
    { literal: "look what you made me do", weight: 0.8, category: 'darvo', rationale: 'blame externalization' },
    { literal: "this is your fault", weight: 0.6, category: 'darvo', rationale: 'blame assignment' },
    { literal: "you brought this on yourself", weight: 0.7, category: 'darvo', rationale: 'victim blaming' },
    { literal: "i'm the victim here", weight: 0.7, category: 'darvo', rationale: 'victim role claim' },
    { literal: "you're attacking me", weight: 0.6, category: 'darvo', rationale: 'attack reversal' },
  ];
  for (const s of matchPatterns(text, 'coercive_ctrl', darvoPats)) { spans.push(s); darvoScore += s.score; }

  impScore = Math.min(impScore, 1); threatScore = Math.min(threatScore, 1);
  isolScore = Math.min(isolScore, 1); obligScore = Math.min(obligScore, 1); darvoScore = Math.min(darvoScore, 1);
  let composite = compositeMax(impScore, threatScore, isolScore, obligScore, darvoScore);
  const active = [impScore, threatScore, isolScore, obligScore, darvoScore].filter(s => s > 0.1).length;
  if (active >= 2) composite *= 1 + active * 0.1;
  return { key: 'coercive_ctrl', score: clamp(composite), spans };
}
