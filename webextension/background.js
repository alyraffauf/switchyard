chrome.action.onClicked.addListener((tab) => {
    const encoded = encodeURIComponent(tab.url);
    const switchyardURL = `switchyard://open?url=${encoded}`;
    chrome.tabs.update(tab.id, { url: switchyardURL });
});
