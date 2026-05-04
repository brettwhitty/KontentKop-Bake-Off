/**
 * IOA Badge Generator — Inline SVG badges with heartbeat pulse animation.
 * Inspired by the IOA's official badge assets at outrage.dataglut.org.
 *
 * Each badge uses a double-beat "heartbeat" scale animation:
 *   beat1 → rest → beat2 → long rest → repeat
 */

const IOA_BADGE_CONFIG = {
  CRITICAL: { color: '#DC143C', glow: '#FF4D6A', label: 'OUTRAGED', icon: '🔴' },
  ELEVATED: { color: '#FF6B00', glow: '#FF9A44', label: 'UPSET',    icon: '🟠' },
  MODERATE: { color: '#FFB800', glow: '#FFCF44', label: 'CONCERNED',icon: '🟡' },
  LOW:      { color: '#4A90D2', glow: '#7AB8F5', label: 'AWARE',    icon: '🔵' },
  CLEARED:  { color: '#28A745', glow: '#5DD47A', label: 'SAFE',     icon: '🟢' },
};

/**
 * Generate an inline SVG badge with heartbeat animation for a given IOA level.
 * @param {string} levelId - One of: CRITICAL, ELEVATED, MODERATE, LOW, CLEARED
 * @param {Object} [opts] - Options
 * @param {number} [opts.size=80] - Badge diameter in pixels
 * @param {boolean} [opts.provisional=true] - Show provisional marker
 * @param {boolean} [opts.animate=true] - Enable heartbeat animation
 * @returns {string} Inline SVG markup
 */
function generateIOABadge(levelId, opts = {}) {
  const size = opts.size || 80;
  const provisional = opts.provisional !== false;
  const animate = opts.animate !== false;
  const cfg = IOA_BADGE_CONFIG[levelId] || IOA_BADGE_CONFIG.CLEARED;

  const cx = size / 2;
  const cy = size / 2;
  const r = size * 0.38;
  const strokeW = size * 0.04;
  const fontSize = size * 0.11;
  const labelY = cy + r + size * 0.12;

  // Heartbeat animation: double-beat pulse
  // Timing: 0s=rest, 0.1s=beat1-peak, 0.2s=beat1-return, 0.3s=beat2-peak, 0.5s=beat2-return, 2s=rest
  const heartbeatAnim = animate ? `
    <animateTransform
      attributeName="transform"
      type="scale"
      values="1; 1.08; 1; 1.05; 1; 1"
      keyTimes="0; 0.05; 0.1; 0.15; 0.25; 1"
      dur="2s"
      repeatCount="indefinite"
      additive="sum"
    />` : '';

  // Glow pulse synced with heartbeat
  const glowAnim = animate ? `
    <animate
      attributeName="r"
      values="${r}; ${r * 1.15}; ${r}; ${r * 1.1}; ${r}; ${r}"
      keyTimes="0; 0.05; 0.1; 0.15; 0.25; 1"
      dur="2s"
      repeatCount="indefinite"
    />
    <animate
      attributeName="opacity"
      values="0.3; 0.6; 0.3; 0.5; 0.3; 0.3"
      keyTimes="0; 0.05; 0.1; 0.15; 0.25; 1"
      dur="2s"
      repeatCount="indefinite"
    />` : '';

  const provisionalRing = provisional ? `
    <circle cx="${cx}" cy="${cy}" r="${r + strokeW * 2}"
      fill="none" stroke="${cfg.color}" stroke-width="1"
      stroke-dasharray="4 3" opacity="0.5" />` : '';

  const provisionalLabel = provisional ? `
    <text x="${cx}" y="${labelY + fontSize * 1.2}"
      text-anchor="middle" font-size="${fontSize * 0.7}"
      font-family="'Segoe UI', system-ui, sans-serif"
      fill="${cfg.color}" opacity="0.7"
      font-style="italic">PROVISIONAL</text>` : '';

  return `<svg xmlns="http://www.w3.org/2000/svg"
    width="${size}" height="${size + (provisional ? size * 0.3 : size * 0.18)}"
    viewBox="0 0 ${size} ${size + (provisional ? size * 0.3 : size * 0.18)}"
    role="img" aria-label="IOA Level: ${cfg.label}${provisional ? ' (Provisional)' : ''}">

    <defs>
      <filter id="glow-${levelId}" x="-50%" y="-50%" width="200%" height="200%">
        <feGaussianBlur stdDeviation="${size * 0.04}" result="blur" />
        <feMerge>
          <feMergeNode in="blur" />
          <feMergeNode in="SourceGraphic" />
        </feMerge>
      </filter>
    </defs>

    <!-- Glow layer -->
    <circle cx="${cx}" cy="${cy}" r="${r}" fill="${cfg.glow}" opacity="0.3"
      filter="url(#glow-${levelId})">
      ${glowAnim}
    </circle>

    <!-- Main badge circle -->
    <g transform-origin="${cx} ${cy}">
      ${heartbeatAnim}

      <circle cx="${cx}" cy="${cy}" r="${r}"
        fill="white" stroke="${cfg.color}" stroke-width="${strokeW}" />

      <!-- Inner ring -->
      <circle cx="${cx}" cy="${cy}" r="${r * 0.78}"
        fill="none" stroke="${cfg.color}" stroke-width="0.5" opacity="0.3" />

      <!-- IOA text -->
      <text x="${cx}" y="${cy - size * 0.08}"
        text-anchor="middle" font-size="${fontSize * 0.8}"
        font-family="'Segoe UI', system-ui, sans-serif"
        fill="${cfg.color}" font-weight="600"
        letter-spacing="2">IOA</text>

      <!-- Level label -->
      <text x="${cx}" y="${cy + size * 0.06}"
        text-anchor="middle" font-size="${fontSize}"
        font-family="'Segoe UI', system-ui, sans-serif"
        fill="${cfg.color}" font-weight="700">${cfg.label}</text>

      <!-- Colored dot -->
      <circle cx="${cx}" cy="${cy + size * 0.15}" r="${size * 0.03}"
        fill="${cfg.color}" />
    </g>

    ${provisionalRing}

    <!-- Level label below badge -->
    <text x="${cx}" y="${labelY}"
      text-anchor="middle" font-size="${fontSize * 0.8}"
      font-family="'Segoe UI', system-ui, sans-serif"
      fill="#666" font-weight="500">${levelId}</text>

    ${provisionalLabel}
  </svg>`;
}

/**
 * Generate a compact inline badge for use in tooltips and panel rows.
 * @param {string} levelId - IOA level
 * @param {number} [size=24] - Badge size
 * @returns {string} Compact SVG markup
 */
function generateIOABadgeMini(levelId, size) {
  size = size || 24;
  const cfg = IOA_BADGE_CONFIG[levelId] || IOA_BADGE_CONFIG.CLEARED;
  const cx = size / 2;
  const cy = size / 2;
  const r = size * 0.4;

  return `<svg xmlns="http://www.w3.org/2000/svg"
    width="${size}" height="${size}" viewBox="0 0 ${size} ${size}"
    style="vertical-align: middle; display: inline-block;"
    role="img" aria-label="IOA: ${cfg.label}">
    <circle cx="${cx}" cy="${cy}" r="${r}"
      fill="white" stroke="${cfg.color}" stroke-width="2" />
    <circle cx="${cx}" cy="${cy}" r="${r * 0.45}"
      fill="${cfg.color}" />
    <animateTransform attributeName="transform" type="scale"
      values="1; 1.1; 1; 1.06; 1; 1"
      keyTimes="0; 0.05; 0.1; 0.15; 0.25; 1"
      dur="2s" repeatCount="indefinite" additive="sum"
      xlink:href="" />
  </svg>`;
}
