const ROUTABLE = /^https?:/;
const ROUTABLE_PATTERNS = ["http://*/*", "https://*/*"];

function openInSwitchyard(url: string, tabId: number) {
  const params = new URLSearchParams({ url });
  chrome.tabs.update(tabId, { url: `switchyard://open?${params}` });
}

chrome.runtime.onInstalled.addListener(() => {
  chrome.contextMenus.removeAll(() => {
    chrome.contextMenus.create({
      id: "open-link",
      title: "Open in Switchyard",
      contexts: ["link"],
      targetUrlPatterns: ROUTABLE_PATTERNS,
    });
    chrome.contextMenus.create({
      id: "open-page",
      title: "Open in Switchyard",
      contexts: ["page"],
      documentUrlPatterns: ROUTABLE_PATTERNS,
    });
  });
});

chrome.contextMenus.onClicked.addListener((info, tab) => {
  const tabId = tab?.id;
  if (tabId == null) return;
  if (info.menuItemId === "open-link" && info.linkUrl && ROUTABLE.test(info.linkUrl)) {
    openInSwitchyard(info.linkUrl, tabId);
  } else if (info.menuItemId === "open-page" && tab?.url && ROUTABLE.test(tab.url)) {
    openInSwitchyard(tab.url, tabId);
  }
});
