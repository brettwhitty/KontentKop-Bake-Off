/**
 * KontentKop Chrome Extension — Content Script
 *
 * Walks the visible DOM, extracts text blocks from paragraphs/headings/list items,
 * scores each block with the 15-metric engine, and highlights flagged regions
 * with hover tooltips showing the metric breakdown.
 */

(function () {
  'use strict';

  // Avoid double-injection
  if (window.__kk_loaded) return;
  window.__kk_loaded = true;

  // ── State ──
  let isActive = false;
  let highlights = [];
  let badge = null;
  let panel = null;
  let tooltip = null;
  let pageResults = [];

  // ── Settings (synced from storage) ──
  let settings = {
    enabled: true,
    threshold: 0.05,
    showBadge: true,
    autoScan: true,
  };

  // Load settings
  chrome.storage.sync.get(['kk_settings'], (data) => {
    if (data.kk_settings) Object.assign(settings, data.kk_settings);
    if (settings.enabled && settings.autoScan) {
      setTimeout(scan, 1500); // wait for page to settle
    }
  });

  // Listen for messages from popup
  chrome.runtime.onMessage.addListener((msg, sender, sendResponse) => {
    if (msg.action === 'scan') { scan(); sendResponse({ ok: true }); }
    if (msg.action === 'clear') { clearHighlights(); sendResponse({ ok: true }); }
    if (msg.action === 'getResults') { sendResponse({ results: pageResults, active: isActive }); }
    if (msg.action === 'updateSettings') {
      Object.assign(settings, msg.settings);
      chrome.storage.sync.set({ kk_settings: settings });
      if (!settings.enabled) clearHighlights();
      sendResponse({ ok: true });
    }
    return true;
  });

  // ── DOM Scanning ──
  function getTextBlocks() {
    const selectors = 'p, h1, h2, h3, h4, h5, h6, li, td, th, blockquote, figcaption, .comment-body, .post-content, article, [role="article"]';
    const elements = document.querySelectorAll(selectors);
    const blocks = [];
    const seen = new Set();

    for (const el of elements) {
      // Skip hidden elements, scripts, styles
      if (!el.offsetParent && el.tagName !== 'BODY') continue;
      const text = el.innerText || el.textContent || '';
      const trimmed = text.trim();
      if (trimmed.length < 20) continue; // skip very short blocks
      if (seen.has(trimmed)) continue;
      seen.add(trimmed);

      // Skip if a parent block already covers this text
      let dominated = false;
      for (const b of blocks) {
        if (b.element.contains(el) && b.text === trimmed) { dominated = true; break; }
      }
      if (dominated) continue;

      blocks.push({ element: el, text: trimmed });
    }
    return blocks;
  }

  function scan() {
    clearHighlights();
    if (!settings.enabled) return;

    const blocks = getTextBlocks();
    let totalFlagged = 0;
    let maxBC = 0;
    pageResults = [];

    for (const block of blocks) {
      const result = analyzeText(block.text);
      if (!result) continue;

      pageResults.push(result);

      if (result.bc > maxBC) maxBC = result.bc;
      if (!result.flagged) continue;
      totalFlagged++;

      // Highlight the flagged spans within this element
      highlightElement(block.element, result);
    }

    isActive = true;
    showBadge(totalFlagged, maxBC);
  }

  function highlightElement(element, result) {
    if (!result.allSpans.length) {
      // No specific spans — highlight the whole element
      element.classList.add('kk-highlight', 'kk-highlight-' + result.severity);
      element.dataset.kkBc = result.bc.toFixed(3);
      element.dataset.kkMetrics = JSON.stringify(result.topMetrics);
      element.addEventListener('mouseenter', onHighlightEnter);
      element.addEventListener('mouseleave', onHighlightLeave);
      highlights.push({ type: 'class', element });
      return;
    }

    // Merge overlapping spans and take top ones
    const mergedSpans = mergeOverlappingSpans(result.allSpans.slice(0, 30));
    const text = element.innerText || element.textContent || '';

    // Use TreeWalker to find text nodes and wrap matched regions
    const walker = document.createTreeWalker(element, NodeFilter.SHOW_TEXT, null, false);
    let node;
    let offset = 0;
    const nodesToWrap = [];

    while ((node = walker.nextNode())) {
      const nodeText = node.textContent;
      const nodeStart = offset;
      const nodeEnd = offset + nodeText.length;

      for (const span of mergedSpans) {
        // Check if this span overlaps with this text node
        const overlapStart = Math.max(span.start, nodeStart);
        const overlapEnd = Math.min(span.end, nodeEnd);
        if (overlapStart < overlapEnd) {
          nodesToWrap.push({
            node,
            localStart: overlapStart - nodeStart,
            localEnd: overlapEnd - nodeStart,
            span,
            severity: span.score >= 0.7 ? 'severe' : span.score >= 0.5 ? 'high' : span.score >= 0.3 ? 'medium' : 'low',
          });
        }
      }
      offset = nodeEnd;
    }

    // Apply wrapping in reverse order to preserve offsets
    const processed = new Set();
    for (let i = nodesToWrap.length - 1; i >= 0; i--) {
      const wrap = nodesToWrap[i];
      const nodeId = wrap.node.textContent + wrap.localStart;
      if (processed.has(nodeId)) continue;
      processed.add(nodeId);

      try {
        const range = document.createRange();
        range.setStart(wrap.node, wrap.localStart);
        range.setEnd(wrap.node, wrap.localEnd);

        const mark = document.createElement('mark');
        mark.className = 'kk-highlight kk-highlight-' + wrap.severity;
        mark.dataset.kkScore = wrap.span.score.toFixed(2);
        mark.dataset.kkMetric = METRIC_LABELS[wrap.span.metricKey] || wrap.span.metricKey;
        mark.dataset.kkRationale = wrap.span.rationale;
        mark.dataset.kkCategory = wrap.span.category || '';
        mark.addEventListener('mouseenter', onHighlightEnter);
        mark.addEventListener('mouseleave', onHighlightLeave);

        range.surroundContents(mark);
        highlights.push({ type: 'mark', element: mark, parent: mark.parentNode });
      } catch (e) {
        // Range errors can happen with complex DOM; skip gracefully
      }
    }
  }

  function mergeOverlappingSpans(spans) {
    if (!spans.length) return [];
    const sorted = [...spans].sort((a, b) => a.start - b.start || b.score - a.score);
    const merged = [sorted[0]];
    for (let i = 1; i < sorted.length; i++) {
      const last = merged[merged.length - 1];
      if (sorted[i].start < last.end) {
        // Overlapping — extend and keep higher score
        last.end = Math.max(last.end, sorted[i].end);
        if (sorted[i].score > last.score) {
          last.score = sorted[i].score;
          last.rationale = sorted[i].rationale;
          last.metricKey = sorted[i].metricKey;
          last.category = sorted[i].category;
        }
      } else {
        merged.push({ ...sorted[i] });
      }
    }
    return merged;
  }

  // ── Tooltip ──
  function ensureTooltip() {
    if (tooltip) return;
    tooltip = document.createElement('div');
    tooltip.className = 'kk-tooltip';
    document.body.appendChild(tooltip);
  }

  function onHighlightEnter(e) {
    ensureTooltip();
    const el = e.currentTarget;
    let html = '';

    if (el.dataset.kkMetrics) {
      // Block-level highlight
      const metrics = JSON.parse(el.dataset.kkMetrics);
      const bc = parseFloat(el.dataset.kkBc);
      const sev = getSeverity(bc);
      html = `<div class="kk-tooltip-header"><span>&#x1F6A8;</span> BC = ${bc.toFixed(3)}</div>`;
      for (const m of metrics) {
        const mSev = m.score >= 0.7 ? 'severe' : m.score >= 0.5 ? 'high' : m.score >= 0.3 ? 'medium' : 'low';
        html += `<div class="kk-tooltip-metric"><span class="kk-tooltip-metric-name">${m.label}</span><span class="kk-tooltip-metric-score kk-score-${mSev}">${m.score.toFixed(2)}</span></div>`;
      }
    } else {
      // Span-level highlight
      const metric = el.dataset.kkMetric;
      const score = parseFloat(el.dataset.kkScore);
      const rationale = el.dataset.kkRationale;
      const sev = score >= 0.7 ? 'severe' : score >= 0.5 ? 'high' : score >= 0.3 ? 'medium' : 'low';
      html = `<div class="kk-tooltip-header"><span class="kk-tooltip-metric-score kk-score-${sev}">${score.toFixed(2)}</span> ${metric}</div>`;
      html += `<div class="kk-tooltip-rationale">${rationale}</div>`;
    }

    tooltip.innerHTML = html;
    tooltip.classList.add('kk-visible');
    positionTooltip(e);
  }

  function onHighlightLeave() {
    if (tooltip) tooltip.classList.remove('kk-visible');
  }

  function positionTooltip(e) {
    if (!tooltip) return;
    const x = e.clientX + 12;
    const y = e.clientY + 12;
    const rect = tooltip.getBoundingClientRect();
    const maxX = window.innerWidth - 400;
    const maxY = window.innerHeight - 200;
    tooltip.style.left = Math.min(x, maxX) + 'px';
    tooltip.style.top = Math.min(y, maxY) + 'px';
  }

  // ── Badge ──
  function showBadge(flaggedCount, maxBC) {
    if (!settings.showBadge) return;
    if (badge) badge.remove();

    const ioaClass = bcToIOALevel(maxBC);
    const ioaLevelId = Object.keys(IOA_LEVELS).find(k => IOA_LEVELS[k] === ioaClass) || 'CLEARED';

    badge = document.createElement('div');
    badge.className = 'kk-badge ' + (flaggedCount > 0 ? 'kk-badge-flagged' : 'kk-badge-clean');
    if (flaggedCount > 0) {
      badge.style.background = ioaClass.color;
      badge.innerHTML = `<span class="kk-badge-icon">${ioaClass.emoji}</span> IOA: ${ioaLevelId} (BC ${maxBC.toFixed(3)})`;
    } else {
      badge.innerHTML = `<span class="kk-badge-icon">&#x2705;</span> IOA: CLEARED`;
    }
    badge.addEventListener('click', togglePanel);
    document.body.appendChild(badge);
  }

  // ── Panel ──
  function togglePanel() {
    if (!panel) createPanel();
    panel.classList.toggle('kk-visible');
    if (panel.classList.contains('kk-visible')) updatePanel();
  }

  function createPanel() {
    panel = document.createElement('div');
    panel.className = 'kk-panel';
    panel.innerHTML = `
      <div class="kk-panel-header">
        <span>&#x1F50D; KontentKop Analysis</span>
        <button class="kk-panel-close" aria-label="Close panel">&times;</button>
      </div>
      <div class="kk-panel-body" id="kk-panel-body"></div>
    `;
    panel.querySelector('.kk-panel-close').addEventListener('click', () => panel.classList.remove('kk-visible'));
    document.body.appendChild(panel);
  }

  function updatePanel() {
    const body = panel.querySelector('#kk-panel-body');
    if (!pageResults.length) {
      body.innerHTML = '<p style="color:#888;text-align:center;padding:20px;">No text blocks analyzed yet. Click "Scan Page" in the popup.</p>';
      return;
    }

    const flagged = pageResults.filter(r => r.flagged);
    const maxBC = Math.max(...pageResults.map(r => r.bc));

    // Aggregate metrics across all flagged blocks
    const metricTotals = {};
    for (const r of flagged) {
      for (const m of r.topMetrics) {
        if (!metricTotals[m.key] || m.score > metricTotals[m.key]) {
          metricTotals[m.key] = m.score;
        }
      }
    }

    let html = '';

    // IOA Classification
    const ioaClass = bcToIOALevel(maxBC);
    const ioaLevelId = Object.keys(IOA_LEVELS).find(k => IOA_LEVELS[k] === ioaClass) || 'CLEARED';

    html += `<div class="kk-panel-section" style="text-align:center;">`;
    html += generateIOABadge(ioaLevelId, { size: 80, provisional: true, animate: true });
    html += `<div class="kk-panel-bc" style="color:${ioaClass.color};margin-top:8px;">${maxBC.toFixed(3)}</div>`;
    html += `<div style="color:#888;font-size:12px;">Body Count &middot; ${flagged.length} of ${pageResults.length} blocks flagged</div>`;
    html += `<div style="margin-top:8px;padding:8px 12px;background:${ioaClass.bg};border-radius:6px;font-size:12px;color:#555;line-height:1.4;">${ioaClass.guidance}</div>`;
    html += `</div>`;

    // Top metrics
    if (Object.keys(metricTotals).length) {
      html += `<div class="kk-panel-section"><div class="kk-panel-section-title">Top Metrics</div>`;
      const sorted = Object.entries(metricTotals).sort((a, b) => b[1] - a[1]);
      for (const [key, score] of sorted) {
        const label = METRIC_LABELS[key] || key;
        const pct = Math.round(score * 100);
        const color = score >= 0.7 ? '#8b00ff' : score >= 0.5 ? '#dc3545' : score >= 0.3 ? '#ff8c00' : '#ffc107';
        html += `<div class="kk-panel-metric-row">
          <span style="min-width:120px">${label}</span>
          <div class="kk-panel-metric-bar"><div class="kk-panel-metric-fill" style="width:${pct}%;background:${color}"></div></div>
          <span style="min-width:36px;text-align:right;font-family:monospace;font-size:12px">${score.toFixed(2)}</span>
        </div>`;
      }
      html += `</div>`;
    }

    // Top findings
    if (flagged.length) {
      html += `<div class="kk-panel-section"><div class="kk-panel-section-title">Key Findings</div>`;
      const allSpans = flagged.flatMap(r => r.allSpans).sort((a, b) => b.score - a.score).slice(0, 8);
      for (const span of allSpans) {
        const sev = span.score >= 0.7 ? 'severe' : span.score >= 0.5 ? 'high' : span.score >= 0.3 ? 'medium' : 'low';
        const colors = SEVERITY_COLORS[sev];
        html += `<div class="kk-panel-finding" style="background:${colors.bg};border-color:${colors.border}">
          <div class="kk-panel-finding-text" style="color:${colors.text}">"${escapeHtml(span.text.slice(0, 60))}"</div>
          <div class="kk-panel-finding-rationale">${METRIC_LABELS[span.metricKey] || span.metricKey}: ${span.rationale}</div>
        </div>`;
      }
      html += `</div>`;
    }

    // P2P Community Voting
    if (flagged.length) {
      const currentUrl = window.location.href;
      const analysis = { bc: maxBC, ioaLevel: ioaLevelId, topMetrics: Object.entries(metricTotals).map(([k,v]) => ({ key: k, score: v })) };

      getVoteState(currentUrl).then(voteState => {
        const votingSection = document.querySelector('.kk-voting-section');
        if (!votingSection) {
          // Insert voting UI
          const votingHTML = generateVotingUI(voteState, ioaClass);
          const tempDiv = document.createElement('div');
          tempDiv.innerHTML = votingHTML;
          body.appendChild(tempDiv.firstElementChild);

          attachVoteHandlers(currentUrl, analysis, (counts) => {
            // Update vote counts in real-time
            const section = document.querySelector('.kk-voting-section');
            if (section) {
              getVoteState(currentUrl).then(newState => {
                section.outerHTML = generateVotingUI(newState, ioaClass);
              });
            }
          });
        }
      }).catch(() => {
        // P2P not available — voting UI just won't appear
      });
    }

    // IOA Referral block (standard boilerplate — unmodified per IOA requirements)
    if (flagged.length) {
      html += `<div class="kk-panel-section" style="padding-top:8px;">`;
      html += `<div class="kk-panel-section-title">Emotional Well-Being Support</div>`;
      html += getIOAReferralHTML();
      html += `</div>`;
    }

    body.innerHTML = html;
  }

  function escapeHtml(str) {
    const div = document.createElement('div');
    div.textContent = str;
    return div.innerHTML;
  }

  // ── Cleanup ──
  function clearHighlights() {
    for (const h of highlights) {
      if (h.type === 'class') {
        h.element.classList.remove('kk-highlight', 'kk-highlight-low', 'kk-highlight-medium', 'kk-highlight-high', 'kk-highlight-severe');
        delete h.element.dataset.kkBc;
        delete h.element.dataset.kkMetrics;
        h.element.removeEventListener('mouseenter', onHighlightEnter);
        h.element.removeEventListener('mouseleave', onHighlightLeave);
      } else if (h.type === 'mark' && h.element.parentNode) {
        const parent = h.element.parentNode;
        while (h.element.firstChild) parent.insertBefore(h.element.firstChild, h.element);
        parent.removeChild(h.element);
        parent.normalize();
      }
    }
    highlights = [];
    pageResults = [];
    isActive = false;
    if (badge) { badge.remove(); badge = null; }
    if (panel) { panel.remove(); panel = null; }
    if (tooltip) { tooltip.remove(); tooltip = null; }
  }
})();
