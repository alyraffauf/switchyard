# Switchyard Browser Extension

A companion browser extension is available that allows you to open any page in Switchyard with one click. For now, the extension uses our URI scheme under the hood. It redirects the current tab to a `switchyard://open` URL, handing off to Switchyard for browser selection.

The extension lives in the `webextension/` directory and ships as a standard WebExtension compatible with both Firefox and Chromium browsers. It is included in GitHub releases as `switchyard-webextension.zip`. 

The intent is to add deeper integration with the desktop app. Eventually, it will be submitted to the Chrome Web Store and Mozilla.