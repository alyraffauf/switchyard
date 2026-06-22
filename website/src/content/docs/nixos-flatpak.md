---
title: NixOS & Flatpak
description: Allow the Flatpak build of Switchyard to find host browser desktop files on NixOS.
order: 25
---

NixOS users should consider using the Nix package and/or the [home-manager module](/docs/home-manager/). If you do use the Flatpak, Switchyard may not be able to accurately discover your installed browsers.

On NixOS, browser desktop files live in `/run/current-system/sw/share/applications`. That path is not visible inside Flatpak by default.

To expose them to Switchyard, add the path to the Flatpak sandbox and include it in the XDG application search path:

```bash
flatpak override --user io.github.alyraffauf.Switchyard \
  --filesystem=/run/current-system/sw/share:ro \
  --env=XDG_DATA_DIRS=/app/share:/usr/share:/run/current-system/sw/share
```

Restart Switchyard after applying the override.
