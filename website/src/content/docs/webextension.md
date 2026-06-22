---
title: Browser Extension
description: Use the Switchyard browser extension and native messaging integration.
order: 40
---

A companion browser extension that lets you open any page in Switchyard with one click. It lives in the `webextension/` directory, ships as a standard WebExtension compatible with Firefox and Chromium, and is available on the [Firefox Add-ons](https://addons.mozilla.org/firefox/addon/switchyard/) and [Chrome Web Store](https://chromewebstore.google.com/detail/switchyard/ncehhpikkabfdcceimdhjjjodogflokc).

- [Get for Firefox](https://addons.mozilla.org/firefox/addon/switchyard/)
- [Get for Chrome](https://chromewebstore.google.com/detail/switchyard/ncehhpikkabfdcceimdhjjjodogflokc)

The source is also bundled in GitHub releases as `switchyard-webextension.zip` if you prefer to load it manually.

## How It Works

The browser extension stays simple by making use of our [URI Scheme](/docs/uri-scheme/). When you click a browser button or "Switchyard", it redirects the current tab to a `switchyard://open?url=...` URL. The operating system sees the registered `switchyard://` scheme and launches the Switchyard desktop app, which parses the URL and forwards your page to the right browser.

With [desktop integration](#desktop-integration) enabled, the extension can also ask Switchyard for your installed browsers and display them directly in the popup for one-click launching.

## Desktop Integration

The Switchyard extension can also show your installed browsers directly in the popup, letting you send the current tab to a specific browser in one click. This requires installing a native messaging host on your system.

Please read [`browser-setup/main.py`](https://github.com/alyraffauf/switchyard/blob/master/browser-setup/main.py) before running this command:

```bash
curl -fsSL https://raw.githubusercontent.com/alyraffauf/switchyard/master/scripts/install-desktop-integration.sh | bash -s -- --install --yes
```

After installation, restart your browser(s). The extension popup will list your installed browsers.

### Uninstall

```bash
curl -fsSL https://raw.githubusercontent.com/alyraffauf/switchyard/master/scripts/install-desktop-integration.sh | bash -s -- --uninstall
```

This removes all manifests and wrappers, but the `flatpak override` permissions granted during install are left in place. It's up to the user to decide what to do with them.

## Privacy Policy

Switchyard does not collect, store, or transmit any user data. The browser extension reads the current tab's URL solely to pass it to the local Switchyard desktop app. No data leaves your device.
