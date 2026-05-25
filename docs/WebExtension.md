# Switchyard Browser Extension

A companion browser extension is available that allows you to open any page in Switchyard with one click. For now, the extension uses our URI scheme under the hood. It redirects the current tab to a `switchyard://open` URL, handing off to Switchyard for browser selection.

The extension lives in the `webextension/` directory and ships as a standard WebExtension compatible with both Firefox and Chromium browsers. It is included in GitHub releases as `switchyard-webextension.zip`.

The intent is to add deeper integration with the desktop app. Eventually, it will be submitted to the Chrome Web Store and Mozilla.

## Desktop Integration

The Switchyard extension can also show your installed browsers directly in the popup, letting you send the current tab to a specific browser in one click. This requires installing a native messaging host on your system.

Please read [`scripts/install-desktop-integration.sh`](../scripts/install-desktop-integration.sh) before running this command:

```bash
curl -fsSL https://raw.githubusercontent.com/alyraffauf/Switchyard/master/scripts/install-desktop-integration.sh | bash
```

After installation, restart your browser(s). The extension popup will list your installed browsers.

### Uninstall

```bash
bash install-desktop-integration.sh --uninstall
```

This removes all manifests and wrappers, but the `flatpak override` permissions granted during install are left in place. It's up to the user to decide what to do with them.

## How It Works

The browser extension stays simple by making use of our [URI Scheme](./URI%20Scheme.md). When you click a browser button or "Switchyard", it redirects the current tab to a `switchyard://open?url=...` URL. The operating system sees the registered `switchyard://` scheme and launches the Switchyard desktop app, which parses the URL and forwards your page to the right browser.

With [desktop integration](#desktop-integration) enabled, the extension can also ask Switchyard for your installed browsers and display them directly in the popup for one-click launching.

## Privacy Policy

Switchyard does not collect, store, or transmit any user data. The browser extension reads the current tab's URL solely to pass it to the local Switchyard desktop app. No data leaves your device.
