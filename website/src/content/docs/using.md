---
title: Using Switchyard
description: Learn what Switchyard does and how to start routing links.
order: 10
---

Switchyard is a rules-based browser launcher for Linux. Once it is set as your default browser, clicked links pass through Switchyard first. It can then open the link in a specific browser, ask you which browser to use, or rewrite the URL before opening it.

Use it when different parts of your life belong in different browsers: work links in Chrome, personal links in Firefox, video links in Brave, or privacy-friendly redirects before anything opens.

The basic flow is:

1. Set Switchyard as your default browser.
2. Add browser rules for domains, keywords, globs, or regular expressions.
3. Optionally add link redirections to clean up or rewrite URLs.
4. Let Switchyard route links automatically, or show the launcher when no rule matches.

## Set as Default Browser

After installation, set Switchyard as your default browser so it can route all clicked links:

```bash
xdg-settings set default-web-browser io.github.alyraffauf.Switchyard.desktop
```

Or use your desktop environment's graphical settings to set Switchyard as the default browser.
