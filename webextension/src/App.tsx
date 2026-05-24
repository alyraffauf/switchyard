import { useEffect, useRef, useState } from "react";
import { createRoot } from "react-dom/client";
import { BrowserIcon } from "./icons";

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

function cacheLocal(browsers: Browser[]) {
  try {
    localStorage.setItem(CACHE_KEY, JSON.stringify(browsers));
  } catch {}
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
  const [loaded, setLoaded] = useState(false);

  const fetching = useRef(false);

  useEffect(() => {
    chrome.tabs.query({ active: true, currentWindow: true }, ([t]) =>
      setTab(t ?? null),
    );
  }, []);

  useEffect(() => {
    if (!tab?.url || !ROUTABLE.test(tab.url) || fetching.current) return;
    fetching.current = true;

    chrome.runtime.sendNativeMessage(
      HOST,
      { action: "listBrowsers" },
      (resp) => {
        fetching.current = false;
        if (chrome.runtime.lastError) {
          setLoaded(true);
          return;
        }
        const list = (resp as BrowserList)?.browsers ?? [];
        cacheLocal(list);
        setBrowsers(list);
        setLoaded(true);
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
        <img
          src="icons/switchyard.svg"
          width="16"
          height="16"
          style={{ flexShrink: 0, marginRight: 8 }}
          alt=""
        />
        Open in Switchyard
      </button>
      {((loaded && browsers.length === 0) || browsers.length > 0) && (
        <div className="separator" />
      )}
      {loaded && browsers.length === 0 && (
        <div className="hint">Install the native host to list browsers</div>
      )}
      {browsers.map((b) => (
        <button
          key={b.id}
          className="row"
          title={b.id}
          onClick={() => launch(url, tabId, b.id)}
        >
          <BrowserIcon browserId={b.id} />
          {b.name}
        </button>
      ))}
    </>
  );
}

createRoot(document.getElementById("root")!).render(<App />);
