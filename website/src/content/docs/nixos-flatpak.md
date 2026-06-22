---
title: NixOS & Flatpak
description: Allow the Flatpak build of Switchyard to find host browser desktop files on NixOS.
order: 25
---

On NixOS, browser desktop files live in `/run/current-system/sw/share/applications`. That path is not visible inside Flatpak by default, so Switchyard may not detect host browsers.

To expose them to Switchyard, add the path to the Flatpak sandbox and include it in the XDG application search path:

```bash
flatpak override --user io.github.alyraffauf.Switchyard \
  --filesystem=/run/current-system/sw/share:ro \
  --env=XDG_DATA_DIRS=/app/share:/usr/share:/run/current-system/sw/share
```

Restart Switchyard after applying the override.
