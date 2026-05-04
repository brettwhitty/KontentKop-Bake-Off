/**
 * KontentKop P2P Voting UI
 *
 * Injects agree/disagree voting controls into the KK analysis panel.
 * Displays real-time community consensus from Gun.js + OrbitDB.
 */

/**
 * Generate the voting UI HTML for the panel.
 * @param {Object} voteState - From getVoteState()
 * @param {Object} ioaLevel - Current IOA level object
 * @returns {string} HTML markup
 */
function generateVotingUI(voteState, ioaLevel) {
  const myVote = voteState ? voteState.myVote : null;
  const agree = voteState ? voteState.agree : 0;
  const disagree = voteState ? voteState.disagree : 0;
  const total = agree + disagree;
  const communityLevel = voteState ? voteState.communityLevel : null;

  const agreeActive = myVote === 'agree' ? 'kk-vote-active' : '';
  const disagreeActive = myVote === 'disagree' ? 'kk-vote-active' : '';

  let consensusHTML = '';
  if (communityLevel === 'confirmed') {
    consensusHTML = '<span style="color:#28A745;font-weight:600;">✓ Community Confirmed</span>';
  } else if (communityLevel === 'disputed') {
    consensusHTML = '<span style="color:#DC3545;font-weight:600;">✗ Community Disputed</span>';
  } else if (communityLevel === 'mixed') {
    consensusHTML = '<span style="color:#FF8C00;font-weight:600;">◐ Mixed Response</span>';
  } else if (total > 0) {
    consensusHTML = '<span style="color:#888;">Gathering consensus...</span>';
  } else {
    consensusHTML = '<span style="color:#888;">Be the first to vote</span>';
  }

  // Agree/disagree bar visualization
  const agreeWidth = total > 0 ? Math.round((agree / total) * 100) : 50;
  const disagreeWidth = total > 0 ? 100 - agreeWidth : 50;

  return `
    <div class="kk-panel-section kk-voting-section">
      <div class="kk-panel-section-title">Community Voice</div>

      <div style="text-align:center;margin-bottom:10px;">
        <div style="font-size:12px;color:#888;margin-bottom:6px;">
          Do you agree with this IOA ${ioaLevel.id} classification?
        </div>

        <div style="display:flex;gap:8px;justify-content:center;margin-bottom:10px;">
          <button class="kk-vote-btn kk-vote-agree ${agreeActive}"
                  id="kk-vote-agree"
                  ${myVote ? 'disabled' : ''}>
            👍 Agree ${agree > 0 ? '(' + agree + ')' : ''}
          </button>
          <button class="kk-vote-btn kk-vote-disagree ${disagreeActive}"
                  id="kk-vote-disagree"
                  ${myVote ? 'disabled' : ''}>
            👎 Disagree ${disagree > 0 ? '(' + disagree + ')' : ''}
          </button>
        </div>

        ${total > 0 ? `
        <div style="display:flex;height:6px;border-radius:3px;overflow:hidden;background:#f0f0f0;margin-bottom:6px;">
          <div style="width:${agreeWidth}%;background:#28A745;transition:width 0.3s;"></div>
          <div style="width:${disagreeWidth}%;background:#DC3545;transition:width 0.3s;"></div>
        </div>
        <div style="font-size:11px;color:#888;">${total} vote${total !== 1 ? 's' : ''} from the community</div>
        ` : ''}

        <div style="margin-top:6px;font-size:12px;">${consensusHTML}</div>
      </div>

      <div style="font-size:10px;color:#aaa;text-align:center;border-top:1px solid #f0f0f0;padding-top:8px;">
        Votes synced via p2p network (Gun.js + OrbitDB/IPFS)
        <br>Provisional classification pending
        <a href="https://outrage.dataglut.org" target="_blank" style="color:#4A90D2;">IOA</a>
        ITVB certification
      </div>
    </div>
  `;
}

/**
 * Attach vote button event handlers.
 * Call this after injecting the voting UI HTML into the DOM.
 * @param {string} url - Current page URL
 * @param {Object} analysis - KK analysis { bc, ioaLevel, topMetrics }
 * @param {function} onVoteUpdate - Callback when votes change
 */
function attachVoteHandlers(url, analysis, onVoteUpdate) {
  const agreeBtn = document.getElementById('kk-vote-agree');
  const disagreeBtn = document.getElementById('kk-vote-disagree');

  if (agreeBtn) {
    agreeBtn.addEventListener('click', async () => {
      agreeBtn.disabled = true;
      disagreeBtn.disabled = true;
      agreeBtn.classList.add('kk-vote-active');
      const counts = await submitVote(url, 'agree', analysis);
      if (onVoteUpdate) onVoteUpdate(counts);
    });
  }

  if (disagreeBtn) {
    disagreeBtn.addEventListener('click', async () => {
      agreeBtn.disabled = true;
      disagreeBtn.disabled = true;
      disagreeBtn.classList.add('kk-vote-active');
      const counts = await submitVote(url, 'disagree', analysis);
      if (onVoteUpdate) onVoteUpdate(counts);
    });
  }

  // Subscribe to real-time updates
  subscribeToVotes(url, (counts) => {
    if (onVoteUpdate) onVoteUpdate(counts);
  });
}
