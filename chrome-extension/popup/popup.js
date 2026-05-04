// KontentKop Popup Script
document.addEventListener('DOMContentLoaded', () => {
  const toggleEnabled = document.getElementById('toggle-enabled');
  const toggleAutoscan = document.getElementById('toggle-autoscan');
  const toggleBadge = document.getElementById('toggle-badge');
  const thresholdSlider = document.getElementById('threshold');
  const thresholdValue = document.getElementById('threshold-value');
  const btnScan = document.getElementById('btn-scan');
  const btnClear = document.getElementById('btn-clear');
  const status = document.getElementById('status');

  // Load saved settings
  chrome.storage.sync.get(['kk_settings'], (data) => {
    const s = data.kk_settings || {};
    if (s.enabled !== undefined) toggleEnabled.checked = s.enabled;
    if (s.autoScan !== undefined) toggleAutoscan.checked = s.autoScan;
    if (s.showBadge !== undefined) toggleBadge.checked = s.showBadge;
    if (s.threshold !== undefined) {
      thresholdSlider.value = s.threshold;
      thresholdValue.textContent = s.threshold.toFixed(2);
    }
  });

  // Check current page status
  sendToContent({ action: 'getResults' }, (response) => {
    if (response && response.active) {
      const flagged = (response.results || []).filter(r => r.flagged).length;
      const total = (response.results || []).length;
      if (flagged > 0) {
        status.textContent = `${flagged} of ${total} blocks flagged`;
        status.className = 'status status-flagged';
      } else {
        status.textContent = `${total} blocks analyzed — all clean`;
        status.className = 'status status-active';
      }
    }
  });

  // Settings changes
  function saveSettings() {
    const settings = {
      enabled: toggleEnabled.checked,
      autoScan: toggleAutoscan.checked,
      showBadge: toggleBadge.checked,
      threshold: parseFloat(thresholdSlider.value),
    };
    sendToContent({ action: 'updateSettings', settings });
  }

  toggleEnabled.addEventListener('change', saveSettings);
  toggleAutoscan.addEventListener('change', saveSettings);
  toggleBadge.addEventListener('change', saveSettings);
  thresholdSlider.addEventListener('input', () => {
    thresholdValue.textContent = parseFloat(thresholdSlider.value).toFixed(2);
  });
  thresholdSlider.addEventListener('change', saveSettings);

  // Scan button
  btnScan.addEventListener('click', () => {
    status.textContent = 'Scanning...';
    status.className = 'status';
    sendToContent({ action: 'scan' }, () => {
      setTimeout(() => {
        sendToContent({ action: 'getResults' }, (response) => {
          if (response && response.results) {
            const flagged = response.results.filter(r => r.flagged).length;
            const total = response.results.length;
            if (flagged > 0) {
              status.textContent = `${flagged} of ${total} blocks flagged`;
              status.className = 'status status-flagged';
            } else {
              status.textContent = `${total} blocks analyzed — all clean`;
              status.className = 'status status-active';
            }
          }
        });
      }, 500);
    });
  });

  // Clear button
  btnClear.addEventListener('click', () => {
    sendToContent({ action: 'clear' });
    status.textContent = 'Highlights cleared';
    status.className = 'status';
  });

  function sendToContent(msg, callback) {
    chrome.tabs.query({ active: true, currentWindow: true }, (tabs) => {
      if (tabs[0]) {
        chrome.tabs.sendMessage(tabs[0].id, msg, callback || (() => {}));
      }
    });
  }
});
