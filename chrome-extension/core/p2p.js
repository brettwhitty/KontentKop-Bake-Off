/**
 * KontentKop P2P Community Voting Layer
 *
 * Stack:
 *   - Gun.js: Real-time p2p sync between active extension peers
 *   - OrbitDB on IPFS: Persistent content-addressed storage of votes
 *
 * Content is identified by SHA-256 hash of the canonical page URL.
 * Users can agree/disagree with KK analysis findings.
 * Community consensus is computed from aggregated votes.
 * When the IOA REST API goes live, this layer bridges to official determinations.
 */

// ── Configuration ──
const P2P_CONFIG = {
  gunPeers: [
    'https://gun-manhattan.herokuapp.com/gun',
    'https://gun-us.herokuapp.com/gun',
  ],
  gunNamespace: 'kontentkop-ioa-v1',
  orbitDbName: 'kontentkop-votes-v1',
  // IOA API endpoint — will be updated when their REST API launches
  ioaApiBase: 'https://api.outrage.dataglut.org/v1',
};

// ── State ──
let gunInstance = null;
let orbitDb = null;
let ipfsNode = null;
let votesDb = null;
let peerId = null;
let p2pReady = false;
let p2pInitPromise = null;

// ── Content ID Generation ──

/**
 * Generate a stable content ID from a URL using SHA-256.
 * Strips fragments and normalizes to canonical form.
 * @param {string} url - Page URL
 * @returns {Promise<string>} Hex-encoded SHA-256 hash
 */
async function generateContentId(url) {
  // Normalize: strip fragment, lowercase host, sort query params
  try {
    const u = new URL(url);
    u.hash = '';
    const params = [...u.searchParams.entries()].sort((a, b) => a[0].localeCompare(b[0]));
    u.search = '';
    for (const [k, v] of params) u.searchParams.set(k, v);
    url = u.toString();
  } catch (e) {
    // If URL parsing fails, use as-is
  }

  const encoder = new TextEncoder();
  const data = encoder.encode(url);
  const hashBuffer = await crypto.subtle.digest('SHA-256', data);
  const hashArray = Array.from(new Uint8Array(hashBuffer));
  return hashArray.map(b => b.toString(16).padStart(2, '0')).join('');
}

/**
 * Generate an anonymous peer ID for this browser instance.
 * Persisted in chrome.storage.local so it's stable across sessions.
 * @returns {Promise<string>}
 */
async function getOrCreatePeerId() {
  return new Promise((resolve) => {
    chrome.storage.local.get(['kk_peer_id'], (data) => {
      if (data.kk_peer_id) {
        resolve(data.kk_peer_id);
      } else {
        const id = 'kk-' + crypto.randomUUID();
        chrome.storage.local.set({ kk_peer_id: id });
        resolve(id);
      }
    });
  });
}

// ── Gun.js Layer (Real-time Sync) ──

/**
 * Initialize Gun.js for real-time peer synchronization.
 * Gun handles peer discovery and conflict resolution automatically.
 */
function initGun() {
  if (gunInstance) return gunInstance;

  // Gun.js is loaded via the manifest content_scripts
  if (typeof Gun === 'undefined') {
    console.warn('[KK-P2P] Gun.js not loaded — real-time sync unavailable');
    return null;
  }

  gunInstance = Gun({
    peers: P2P_CONFIG.gunPeers,
    localStorage: true,
    radisk: true,
  });

  console.log('[KK-P2P] Gun.js initialized with peers:', P2P_CONFIG.gunPeers);
  return gunInstance;
}

/**
 * Get the Gun.js node for a specific content ID.
 * @param {string} contentId - SHA-256 hash of the URL
 * @returns {Object} Gun node reference
 */
function getGunNode(contentId) {
  const gun = initGun();
  if (!gun) return null;
  return gun.get(P2P_CONFIG.gunNamespace).get(contentId);
}

/**
 * Submit a vote via Gun.js for real-time propagation.
 * @param {string} contentId
 * @param {string} peerId
 * @param {'agree'|'disagree'} vote
 * @param {Object} kkAnalysis - Local KK analysis results
 */
function gunSubmitVote(contentId, peerId, vote, kkAnalysis) {
  const node = getGunNode(contentId);
  if (!node) return;

  // Store the vote under the peer's ID to prevent double-voting
  node.get('votes').get(peerId).put({
    vote: vote,
    timestamp: new Date().toISOString(),
    bc: kkAnalysis.bc,
    ioaLevel: kkAnalysis.ioaLevel,
  });

  // Update the analysis snapshot
  node.get('analysis').put({
    bc: kkAnalysis.bc,
    ioaLevel: kkAnalysis.ioaLevel,
    topMetrics: JSON.stringify(kkAnalysis.topMetrics),
    lastUpdated: new Date().toISOString(),
  });
}

/**
 * Subscribe to real-time vote updates for a content ID.
 * @param {string} contentId
 * @param {function} callback - Called with { agree, disagree, total, communityLevel }
 */
function gunSubscribeVotes(contentId, callback) {
  const node = getGunNode(contentId);
  if (!node) return;

  node.get('votes').map().on((data, peerId) => {
    if (!data || !data.vote) return;
    // Recount all votes
    gunCountVotes(contentId, callback);
  });
}

/**
 * Count all votes for a content ID from Gun.js.
 * @param {string} contentId
 * @param {function} callback
 */
function gunCountVotes(contentId, callback) {
  const node = getGunNode(contentId);
  if (!node) return;

  let agree = 0, disagree = 0;
  const votes = {};

  node.get('votes').map().once((data, pid) => {
    if (!data || !data.vote) return;
    votes[pid] = data.vote;

    // Recount
    agree = 0; disagree = 0;
    for (const v of Object.values(votes)) {
      if (v === 'agree') agree++;
      else if (v === 'disagree') disagree++;
    }

    const total = agree + disagree;
    const agreeRatio = total > 0 ? agree / total : 0.5;

    // Community consensus: if >66% agree, adopt the KK level; if >66% disagree, lower it
    let communityLevel = null;
    if (total >= 3) {
      if (agreeRatio >= 0.66) communityLevel = 'confirmed';
      else if (agreeRatio <= 0.33) communityLevel = 'disputed';
      else communityLevel = 'mixed';
    }

    callback({ agree, disagree, total, communityLevel, agreeRatio });
  });
}

// ── OrbitDB/IPFS Layer (Persistent Storage) ──

/**
 * Initialize IPFS and OrbitDB for persistent content-addressed storage.
 * This is heavier than Gun.js and initializes asynchronously.
 * @returns {Promise<boolean>} Whether initialization succeeded
 */
async function initOrbitDB() {
  if (orbitDb) return true;

  // Helia (IPFS) and OrbitDB are loaded via manifest content_scripts
  if (typeof createHelia === 'undefined' || typeof OrbitDB === 'undefined') {
    console.warn('[KK-P2P] Helia/OrbitDB not loaded — persistent p2p storage unavailable');
    return false;
  }

  try {
    // Create an IPFS node using Helia (lightweight IPFS for browsers)
    ipfsNode = await createHelia();
    console.log('[KK-P2P] IPFS/Helia node started');

    // Create OrbitDB instance
    orbitDb = await OrbitDB.createInstance(ipfsNode);
    console.log('[KK-P2P] OrbitDB initialized');

    // Open (or create) the votes database
    // Using a key-value store keyed by contentId
    votesDb = await orbitDb.keyvalue(P2P_CONFIG.orbitDbName, {
      accessController: { write: ['*'] }, // Open write for all peers
    });

    await votesDb.load();
    console.log('[KK-P2P] Votes database loaded:', votesDb.address.toString());

    return true;
  } catch (err) {
    console.error('[KK-P2P] OrbitDB init failed:', err);
    return false;
  }
}

/**
 * Persist a vote record to OrbitDB.
 * @param {string} contentId
 * @param {Object} voteRecord
 */
async function orbitPersistVote(contentId, voteRecord) {
  if (!votesDb) return;

  try {
    // Get existing record or create new
    const existing = votesDb.get(contentId) || {
      contentId,
      url: voteRecord.url,
      analysis: voteRecord.analysis,
      votes: [],
      created: new Date().toISOString(),
    };

    // Append vote (deduplicate by peerId)
    existing.votes = existing.votes.filter(v => v.peerId !== voteRecord.peerId);
    existing.votes.push({
      peerId: voteRecord.peerId,
      vote: voteRecord.vote,
      timestamp: new Date().toISOString(),
    });

    existing.lastUpdated = new Date().toISOString();

    await votesDb.put(contentId, existing);
    console.log('[KK-P2P] Vote persisted to OrbitDB:', contentId);
  } catch (err) {
    console.error('[KK-P2P] OrbitDB persist failed:', err);
  }
}

/**
 * Retrieve a vote record from OrbitDB.
 * @param {string} contentId
 * @returns {Object|null}
 */
function orbitGetVotes(contentId) {
  if (!votesDb) return null;
  return votesDb.get(contentId) || null;
}

// ── Unified P2P Interface ──

/**
 * Initialize the full p2p stack.
 * Gun.js starts immediately; OrbitDB initializes in background.
 */
async function initP2P() {
  if (p2pInitPromise) return p2pInitPromise;

  p2pInitPromise = (async () => {
    peerId = await getOrCreatePeerId();
    console.log('[KK-P2P] Peer ID:', peerId);

    // Gun.js — immediate, synchronous
    initGun();

    // OrbitDB — async, may fail gracefully
    const orbitOk = await initOrbitDB();
    if (!orbitOk) {
      console.warn('[KK-P2P] Running in Gun.js-only mode (OrbitDB unavailable)');
    }

    p2pReady = true;
    console.log('[KK-P2P] P2P layer ready');
  })();

  return p2pInitPromise;
}

/**
 * Submit a community vote on KK analysis findings.
 * Writes to both Gun.js (real-time) and OrbitDB (persistent).
 * @param {string} url - Page URL
 * @param {'agree'|'disagree'} vote
 * @param {Object} analysis - KK analysis results { bc, ioaLevel, topMetrics }
 * @returns {Promise<Object>} Updated vote counts
 */
async function submitVote(url, vote, analysis) {
  if (!p2pReady) await initP2P();

  const contentId = await generateContentId(url);

  // Gun.js — real-time propagation
  gunSubmitVote(contentId, peerId, vote, analysis);

  // OrbitDB — persistent storage
  await orbitPersistVote(contentId, {
    url,
    peerId,
    vote,
    analysis: {
      bc: analysis.bc,
      ioaLevel: analysis.ioaLevel,
      topMetrics: analysis.topMetrics,
    },
  });

  // Also store locally for offline access
  const localKey = 'kk_vote_' + contentId;
  chrome.storage.local.set({
    [localKey]: { vote, timestamp: new Date().toISOString() },
  });

  return new Promise((resolve) => {
    gunCountVotes(contentId, resolve);
  });
}

/**
 * Get current vote state for a URL.
 * Checks Gun.js first (real-time), falls back to OrbitDB, then local storage.
 * @param {string} url
 * @returns {Promise<Object>} { agree, disagree, total, communityLevel, myVote }
 */
async function getVoteState(url) {
  if (!p2pReady) await initP2P();

  const contentId = await generateContentId(url);

  // Check local vote
  const localKey = 'kk_vote_' + contentId;
  const localData = await new Promise(resolve => {
    chrome.storage.local.get([localKey], data => resolve(data[localKey] || null));
  });

  // Try OrbitDB for historical data
  const orbitRecord = orbitGetVotes(contentId);

  // Get real-time counts from Gun.js
  return new Promise((resolve) => {
    let resolved = false;

    gunCountVotes(contentId, (counts) => {
      if (!resolved) {
        resolved = true;
        resolve({
          ...counts,
          myVote: localData ? localData.vote : null,
          orbitRecord,
          contentId,
        });
      }
    });

    // Timeout fallback — if Gun.js has no data yet
    setTimeout(() => {
      if (!resolved) {
        resolved = true;
        const fallback = { agree: 0, disagree: 0, total: 0, communityLevel: null, agreeRatio: 0.5 };
        if (orbitRecord) {
          for (const v of orbitRecord.votes || []) {
            if (v.vote === 'agree') fallback.agree++;
            else fallback.disagree++;
          }
          fallback.total = fallback.agree + fallback.disagree;
        }
        resolve({
          ...fallback,
          myVote: localData ? localData.vote : null,
          orbitRecord,
          contentId,
        });
      }
    }, 2000);
  });
}

/**
 * Subscribe to real-time vote updates for a URL.
 * @param {string} url
 * @param {function} callback
 */
async function subscribeToVotes(url, callback) {
  if (!p2pReady) await initP2P();
  const contentId = await generateContentId(url);
  gunSubscribeVotes(contentId, callback);
}
