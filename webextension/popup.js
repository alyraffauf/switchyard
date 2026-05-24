const root = document.getElementById("root");
const header = document.getElementById("header");
const HOST = "io.github.alyraffauf.switchyard";
const ROUTABLE = /^(https?|ftp|file):/;

const dark = matchMedia("(prefers-color-scheme: dark)");
const setTheme = () =>
  document.documentElement.classList.toggle("dark", dark.matches);
dark.addEventListener("change", setTheme);
setTheme();

const make = (tag, className, text) => {
  const el = document.createElement(tag);
  if (className) el.className = className;
  if (text) el.textContent = text;
  root.appendChild(el);
  return el;
};

const launch = (tab, browser) => {
  const params = new URLSearchParams({ url: tab.url });
  if (browser) params.set("browser", browser);
  chrome.tabs.update(tab.id, { url: `switchyard://open?${params}` });
  window.close();
};

chrome.tabs.query({ active: true, currentWindow: true }, ([tab]) => {
  if (!tab?.url) return make("div", "hint", "No active tab");
  if (!ROUTABLE.test(tab.url))
    return make("div", "hint", "This page can't be opened in another browser");

  // header.textContent = tab.title || tab.url;
  // header.title = tab.url;
  // header.hidden = false;

  make("button", "row primary", "Open in Switchyard").onclick = () =>
    launch(tab);

  chrome.runtime.sendNativeMessage(HOST, { action: "listBrowsers" }, (resp) => {
    if (chrome.runtime.lastError) {
      make("div", "separator");
      make("div", "hint", "Install the native host to list browsers");
      return;
    }
    const browsers = resp?.browsers ?? [];
    if (!browsers.length) return;
    make("div", "separator");
    for (const b of browsers) {
      const btn = make("button", "row", b.name);
      btn.title = b.id;
      btn.onclick = () => launch(tab, b.id);
    }
  });
});
