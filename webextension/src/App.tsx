import { useEffect, useState } from "react";
import { createRoot } from "react-dom/client";

interface Browser {
  id: string;
  name: string;
}

type BrowserList = { browsers: Browser[] };

const HOST = "io.github.alyraffauf.switchyard";
const ROUTABLE = /^(https?|ftp|file):/;
const CACHE_KEY = "browsers_v2";

function getCachedBrowsers(): Browser[] {
  try {
    const raw = localStorage.getItem(CACHE_KEY);
    if (!raw) return [];
    const data = JSON.parse(raw);
    return Array.isArray(data) ? data : [];
  } catch {
    return [];
  }
}

function cacheBrowsers(browsers: Browser[]) {
  try {
    localStorage.setItem(CACHE_KEY, JSON.stringify(browsers));
  } catch {
    /* quota exceeded — ignore */
  }
}

function launch(url: string, tabId: number, browser?: string) {
  const params = new URLSearchParams({ url });
  if (browser) params.set("browser", browser);
  chrome.tabs.update(tabId, { url: `switchyard://open?${params}` });
  window.close();
}

function App() {
  const [tab, setTab] = useState<chrome.tabs.Tab | null>(null);
  const [browsers, setBrowsers] = useState<Browser[]>(getCachedBrowsers);
  const [nativeHostMissing, setNativeHostMissing] = useState(false);

  useEffect(() => {
    chrome.tabs.query({ active: true, currentWindow: true }, ([t]) =>
      setTab(t ?? null),
    );
  }, []);

  useEffect(() => {
    if (!tab?.url || !ROUTABLE.test(tab.url)) return;

    chrome.runtime.sendNativeMessage(
      HOST,
      { action: "listBrowsers" },
      (resp) => {
        if (chrome.runtime.lastError) {
          console.error(
            "Switchyard native host error:",
            chrome.runtime.lastError.message,
          );
          setNativeHostMissing(true);
          return;
        }
        setNativeHostMissing(false);
        const list = (resp as BrowserList)?.browsers ?? [];
        cacheBrowsers(list);
        setBrowsers(list);
      },
    );
  }, [tab]);

  if (!tab?.url) return <div className="hint">No active tab</div>;
  if (!ROUTABLE.test(tab.url))
    return (
      <div className="hint">This page can't be opened in another browser</div>
    );

  const { url } = tab;
  const tabId = tab.id!;

  return (
    <>
      <button className="row primary" onClick={() => launch(url, tabId)}>
        Open in Switchyard
      </button>
      {(nativeHostMissing || browsers.length > 0) && (
        <div className="separator" />
      )}
      {nativeHostMissing && (
        <div className="hint">Install the native host to list browsers</div>
      )}
      {browsers.map((b) => (
        <button
          key={b.id}
          className="row"
          title={b.id}
          onClick={() => launch(url, tabId, b.id)}
        >
          {b.name}
        </button>
      ))}
    </>
  );
}

createRoot(document.getElementById("root")!).render(<App />);
