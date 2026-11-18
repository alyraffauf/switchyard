// SPDX-License-Identifier: GPL-3.0-or-later
import assert from "node:assert/strict";
import { test } from "node:test";
import { compile, openPopup } from "./local-browser.mjs";

const script = compile("App.tsx");

function open(cache) {
  const popup = openPopup(script, cache, () =>
    assert.fail("unexpected navigation"),
  );
  popup.resolveTab({ id: 2, url: "https://example.test/" });
  return popup;
}

test("an explicit native host error keeps the last successful browser list", () => {
  const cachedBrowser = {
    id: "org.example.Browser.desktop",
    name: "Example Browser",
  };
  const saved = JSON.stringify([cachedBrowser]);
  const cache = new Map([["browsers_v2", saved]]);
  const popup = open(cache);

  popup.reply({ error: "Unable to list browsers" });

  assert.deepEqual(
    popup.rows().map(({ id }) => id),
    [null, cachedBrowser.id],
  );
  assert.equal(cache.get("browsers_v2"), saved);
});

test("a host error and a successful empty list produce different popup messages", () => {
  const cache = new Map();
  const rejected = open(cache);
  rejected.reply({ error: "Unable to list browsers" });
  assert.match(rejected.text(), /Set up desktop integration/);
  assert.equal(cache.has("browsers_v2"), false);

  const recovered = open(cache);
  recovered.reply({ browsers: [] });
  assert.match(recovered.text(), /No visible browsers/);
  assert.equal(cache.get("browsers_v2"), "[]");
});
