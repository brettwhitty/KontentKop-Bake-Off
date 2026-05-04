/**
 * KontentKop Chrome Extension — Background Service Worker
 *
 * Handles extension lifecycle, badge updates, and coordinates
 * between popup and content scripts.
 */

chrome.runtime.onInstalled.addListener(() => {
  // Set default settings on install
  chrome.storage.sync.get(['kk_settings'], (data) => {
    if (!data.kk_settings) {
      chrome.storage.sync.set({
        kk_settings: {
          enabled: true,
          threshold: 0.05,
          showBadge: true,
          autoScan: true,
        },
      });
    }
  });
});

// Update extension badge based on content script results
chrome.runtime.onMessage.addListener((msg, sender, sendResponse) => {
  if (msg.action === 'updateBadge' && sender.tab) {
    const { flaggedCount, maxBC } = msg;
    if (flaggedCount > 0) {
      chrome.action.setBadgeText({ text: String(flaggedCount), tabId: sender.tab.id });
      chrome.action.setBadgeBackgroundColor({ color: '#DC3545', tabId: sender.tab.id });
    } else {
      chrome.action.setBadgeText({ text: '', tabId: sender.tab.id });
    }
    sendResponse({ ok: true });
  }
  return true;
});
