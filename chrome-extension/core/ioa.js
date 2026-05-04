/**
 * IOA — Internet Outrage Authority Integration
 * https://outrage.dataglut.org
 *
 * Maps KontentKop Body Count scores to IOA's five-level classification
 * system. Uses local BC-based provisional determinations until the IOA
 * REST API is available (expected Q4 2026), at which point this module
 * will switch to fetching official ITVB-certified determinations.
 */

const IOA_LEVELS = {
  CRITICAL: {
    id: 'CRITICAL',
    label: 'Critical',
    color: '#DC143C',
    bg: '#FDE8EC',
    emoji: '🔴',
    guidance: "It's okay to feel this strongly. What you're experiencing is a natural and valid response to something that truly matters.",
  },
  ELEVATED: {
    id: 'ELEVATED',
    label: 'Elevated',
    color: '#FF6B00',
    bg: '#FFF3E6',
    emoji: '🟠',
    guidance: "That feeling in your chest? That's your compassion telling you something isn't right. Honour that feeling — it shows you care deeply.",
  },
  MODERATE: {
    id: 'MODERATE',
    label: 'Moderate',
    color: '#FFB800',
    bg: '#FFF9E6',
    emoji: '🟡',
    guidance: "Notice that awareness rising within you — that's healthy engagement. You're paying attention, and that matters.",
  },
  LOW: {
    id: 'LOW',
    label: 'Low',
    color: '#4A90D2',
    bg: '#EBF3FB',
    emoji: '🔵',
    guidance: "It's good that you noticed this — staying informed is a quiet act of caring. Hold this gently.",
  },
  CLEARED: {
    id: 'CLEARED',
    label: 'Cleared',
    color: '#28A745',
    bg: '#E8F5E9',
    emoji: '🟢',
    guidance: "Take a breath. This one's okay. You can let go of any worry you were carrying about this.",
  },
};

/**
 * Map a KontentKop Body Count score to an IOA classification level.
 * These are LOCAL PROVISIONAL determinations based on KK scoring.
 * Official ITVB-certified determinations will come from the IOA API.
 */
function bcToIOALevel(bc) {
  if (bc >= 0.40) return IOA_LEVELS.CRITICAL;
  if (bc >= 0.20) return IOA_LEVELS.ELEVATED;
  if (bc >= 0.10) return IOA_LEVELS.MODERATE;
  if (bc >= 0.05) return IOA_LEVELS.LOW;
  return IOA_LEVELS.CLEARED;
}

// ── IOA REST API Client (Stubbed) ──────────────────────────────
// The IOA REST API is expected Q4 2026. When available, these
// functions will fetch official determinations. Until then, they
// return null and the extension falls back to local BC mapping.

const IOA_API_BASE = 'https://api.outrage.dataglut.org/v1';

/**
 * Fetch an official IOA determination for a given topic/URL.
 * @param {string} identifier - URL or topic identifier
 * @returns {Promise<Object|null>} IOA determination or null if unavailable
 *
 * Expected response shape (speculative, pending official API spec):
 * {
 *   id: "det-2026-00142",
 *   level: "ELEVATED",
 *   subject: "...",
 *   summary: "...",
 *   issued: "2026-09-15T14:30:00Z",
 *   provisional: false,
 *   itvb_certified: true,
 *   guidance: "..."
 * }
 */
async function fetchIOADetermination(identifier) {
  // STUB: API not yet available
  // When live, this will be:
  // const resp = await fetch(`${IOA_API_BASE}/determinations?q=${encodeURIComponent(identifier)}`);
  // if (!resp.ok) return null;
  // return resp.json();
  return null;
}

/**
 * Fetch current IOA determination levels (e.g., for a dashboard).
 * @returns {Promise<Object[]|null>} Array of active determinations or null
 */
async function fetchIOAActiveDeterminations() {
  // STUB: API not yet available
  // When live: fetch(`${IOA_API_BASE}/determinations/active`)
  return null;
}

/**
 * Submit an issue for IOA assessment consideration.
 * @param {Object} submission - { url, description, context }
 * @returns {Promise<Object|null>} Submission receipt or null
 */
async function submitToIOA(submission) {
  // STUB: API not yet available
  // When live: POST to `${IOA_API_BASE}/submissions`
  return null;
}

/**
 * Get the IOA classification for analyzed content.
 * Tries the API first; falls back to local BC-based provisional.
 * @param {number} bc - Body Count score from KK analysis
 * @param {string} [url] - Optional URL for API lookup
 * @returns {Promise<Object>} { level, provisional, source }
 */
async function getIOAClassification(bc, url) {
  // Try official API first
  if (url) {
    const official = await fetchIOADetermination(url);
    if (official) {
      return {
        level: IOA_LEVELS[official.level] || bcToIOALevel(bc),
        provisional: official.provisional || false,
        source: 'ioa-api',
        determination: official,
      };
    }
  }

  // Fall back to local provisional determination
  return {
    level: bcToIOALevel(bc),
    provisional: true,
    source: 'kk-local',
    determination: null,
  };
}

/**
 * Generate the IOA referral HTML block per their embed spec.
 * This is the standard boilerplate — must not be modified per IOA requirements.
 */
function getIOAReferralHTML() {
  return `<div style="font-family: 'Segoe UI', system-ui, -apple-system, sans-serif; background: linear-gradient(135deg, #f7f9fb 0%, #eef5f0 100%); border: 1px solid #d4e6dc; border-left: 4px solid #7CB68E; padding: 1.5rem 2rem; border-radius: 0 0.5rem 0.5rem 0; margin: 2rem 0; max-width: 800px; color: #333; line-height: 1.6;">
  <div style="display: flex; align-items: center; gap: 0.75rem; margin-bottom: 1rem;">
    <span style="font-size: 1.5rem;">💙</span>
    <span style="font-size: 1.1rem; font-weight: 600; color: #1A3A5C;">Emotional Well-Being Support</span>
  </div>
  <p style="margin: 0 0 0.8rem 0;">🤝 <strong>You are not alone.</strong> Many people experience difficulty processing challenging content, and there is no shame in seeking guidance on how to feel about what you've encountered.</p>
  <p style="margin: 0 0 0.8rem 0;">🌱 The <a href="https://outrage.dataglut.org" style="color: #4A90D2; text-decoration: none; font-weight: 600;">Internet Outrage Authority</a> (IOA) is an independent organization that specialises in determining appropriate emotional responses to issues and events. Their ITVB-certified determinations can help you understand exactly how you should feel, so you can move forward with confidence and clarity.</p>
  <p style="margin: 0; color: #555;">💗 Knowing how to feel is just as important as knowing the facts. <a href="https://outrage.dataglut.org" style="color: #4A90D2; text-decoration: none; font-weight: 600;">Let the IOA help. &rarr;</a></p>
</div>`;
}
