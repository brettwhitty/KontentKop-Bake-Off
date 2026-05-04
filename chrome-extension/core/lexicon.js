/**
 * KontentKop Lexicon — Core pattern matching engine.
 * Ported from Go: src/pipeline/lexicon.go
 */

/**
 * @typedef {Object} Pattern
 * @property {RegExp|null} regex - Compiled regex (null if using literal match)
 * @property {string} literal - Literal phrase match (faster; used when regex is null)
 * @property {number} weight - 0.0–1.0: how strongly this signals the metric
 * @property {string} category - Sub-category (e.g., "narcissism" within dark_triad)
 * @property {string} rationale - Human-readable explanation for annotations
 */

/**
 * @typedef {Object} Span
 * @property {number} start - Character offset into the original text
 * @property {number} end - Character offset (exclusive)
 * @property {string} text - The matched substring
 * @property {string} metricKey - Which metric flagged this span
 * @property {number} score - Local score for this span (0.0–1.0)
 * @property {string} rationale - Human-readable explanation
 * @property {string} category - Sub-category
 */

/**
 * @typedef {Object} MetricResult
 * @property {string} key - Metric key (e.g., "dark_triad")
 * @property {number} score - Overall metric score (0.0–1.0)
 * @property {Span[]} spans - Specific flagged regions in the text
 */

/**
 * Check if position is at a word boundary.
 * Allows apostrophes and hyphens within words (for contractions).
 */
function isWordBound(text, pos) {
  if (pos <= 0 || pos >= text.length) return true;
  const left = text[pos - 1];
  const right = text[pos];
  const isWordChar = (ch) => /[\p{L}'\u2019]/u.test(ch);
  return !isWordChar(left) || !isWordChar(right);
}

/**
 * Scan text for all pattern matches and return Spans.
 * @param {string} text - Original text
 * @param {string} metricKey - Metric identifier
 * @param {Pattern[]} patterns - Array of patterns to match
 * @returns {Span[]}
 */
function matchPatterns(text, metricKey, patterns) {
  const lowered = text.toLowerCase();
  const spans = [];

  for (const p of patterns) {
    if (p.regex) {
      // Reset regex state for global matching
      const re = new RegExp(p.regex.source, p.regex.flags.includes('g') ? p.regex.flags : p.regex.flags + 'g');
      let match;
      while ((match = re.exec(lowered)) !== null) {
        spans.push({
          start: match.index,
          end: match.index + match[0].length,
          text: text.slice(match.index, match.index + match[0].length),
          metricKey,
          score: p.weight,
          rationale: p.rationale,
          category: p.category,
        });
      }
    } else if (p.literal) {
      const litLower = p.literal.toLowerCase();
      let idx = 0;
      while (idx < lowered.length) {
        const pos = lowered.indexOf(litLower, idx);
        if (pos === -1) break;
        const end = pos + litLower.length;
        if (end > text.length) break;
        if (isWordBound(lowered, pos) && isWordBound(lowered, end)) {
          spans.push({
            start: pos,
            end,
            text: text.slice(pos, end),
            metricKey,
            score: p.weight,
            rationale: p.rationale,
            category: p.category,
          });
        }
        idx = pos + 1;
      }
    }
  }
  return spans;
}

/** Clamp value to [0.0, 1.0]. */
function clamp(v) {
  return Math.max(0, Math.min(1, v));
}

/**
 * Composite score from multiple sub-dimension scores.
 * Highest sub-dimension sets the floor; other active dimensions add a bonus.
 */
function compositeMax(...scores) {
  if (scores.length === 0) return 0;
  let maxVal = Math.max(...scores);
  for (const s of scores) {
    if (s < maxVal && s > 0.1) {
      maxVal += s * 0.20;
    }
  }
  return clamp(maxVal);
}

/** Split text into sentences using simple punctuation heuristics. */
function splitSentences(text) {
  const parts = text.split(/[.!?]+\s+/).map(p => p.trim()).filter(Boolean);
  return parts.length > 0 ? parts : [text];
}

// Export for use by metric modules
if (typeof module !== 'undefined') {
  module.exports = { matchPatterns, isWordBound, clamp, compositeMax, splitSentences };
}
