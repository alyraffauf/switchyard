// SPDX-License-Identifier: GPL-3.0-or-later
import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import { createRequire } from "node:module";
import vm from "node:vm";
import { transformSync } from "esbuild";

const require = createRequire(import.meta.url);

export function compile(name) {
  const source = readFileSync(
    new URL(`../src/${name}`, import.meta.url),
    "utf8",
  );
  return new vm.Script(
    transformSync(source, {
      loader: name.endsWith(".tsx") ? "tsx" : "ts",
      jsx: "automatic",
      format: "cjs",
    }).code,
    { filename: name },
  );
}

export function navigationCalls() {
  const calls = [];
  const update = (tab, { url }) => {
    const target = new URL(url);
    assert.equal(target.protocol, "switchyard:");
    assert.equal(target.hostname, "open");
    assert.equal(target.searchParams.getAll("url").length, 1);
    assert.ok(target.searchParams.getAll("browser").length <= 1);
    calls.push({
      tab,
      page: target.searchParams.get("url"),
      browser: target.searchParams.get("browser"),
      url,
    });
  };
  return { calls, update };
}

// Drive the popup's callbacks and flush state and effects synchronously. The
// popup only uses useState, useRef and useEffect. DOM and React lifecycle
// behavior are outside this test double's scope.
export function openPopup(script, cache, update) {
  const state = [];
  const pendingEffects = [];
  let slot = 0;
  let dirty = false;
  let component;
  let view;
  let query;
  let request;
  let closed = 0;
  const react = {
    useState(initial) {
      const index = slot++;
      if (!(index in state)) {
        state[index] = typeof initial === "function" ? initial() : initial;
      }
      return [
        state[index],
        (value) => {
          dirty ||= !Object.is(value, state[index]);
          state[index] = value;
        },
      ];
    },
    useRef(initial) {
      const index = slot++;
      return (state[index] ??= { current: initial });
    },
    useEffect(effect, deps) {
      const index = slot++;
      if (
        !state[index] ||
        deps.some((value, i) => !Object.is(value, state[index][i]))
      ) {
        state[index] = deps;
        pendingEffects.push(effect);
      }
    },
  };

  const runtime = {
    sendNativeMessage(host, message, callback) {
      assert.equal(request, undefined, "one browser request per popup");
      request = { host, action: message.action, callback };
    },
  };
  script.runInNewContext({
    URLSearchParams,
    exports: {},
    require(name) {
      if (name === "react") return react;
      if (name === "react-dom/client") {
        return {
          createRoot: () => ({
            render: (element) => {
              component = element.type;
            },
          }),
        };
      }
      if (name === "./icons") return { BrowserIcon: () => null };
      return require(name);
    },
    chrome: {
      runtime,
      tabs: {
        query(options, callback) {
          assert.ok(options.active && options.currentWindow);
          query = callback;
        },
        update,
      },
    },
    document: { getElementById: () => ({}) },
    window: {
      close: () => {
        closed++;
      },
    },
    localStorage: {
      getItem: (key) => cache.get(key) ?? null,
      setItem: (key, value) => cache.set(key, value),
    },
  });

  function flush() {
    do {
      dirty = false;
      slot = 0;
      view = component();
      for (const effect of pendingEffects.splice(0)) effect();
    } while (dirty);
  }
  flush();

  function* elements(node) {
    if (Array.isArray(node)) {
      for (const child of node) yield* elements(child);
    } else if (node?.props) {
      yield node;
      yield* elements(node.props.children);
    }
  }

  function label(node) {
    if (typeof node === "string") return node;
    if (Array.isArray(node)) return node.map(label).join("");
    return node?.props ? label(node.props.children) : "";
  }

  return {
    resolveTab(tab) {
      query(tab ? [tab] : []);
      flush();
    },
    get request() {
      return request && { host: request.host, action: request.action };
    },
    reply(response, disconnect = false) {
      if (disconnect) runtime.lastError = { message: "Native host closed" };
      try {
        request.callback(response);
      } finally {
        delete runtime.lastError;
      }
      flush();
    },
    rows() {
      return [...elements(view)]
        .filter((node) => node.type === "button")
        .map((node) => ({ label: label(node), id: node.props.title ?? null }));
    },
    text: () => label(view),
    click(row) {
      [...elements(view)]
        .filter((node) => node.type === "button")
        [row].props.onClick();
    },
    closed: () => closed,
  };
}
