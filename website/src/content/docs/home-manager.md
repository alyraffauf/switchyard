---
title: home-manager
description: Install Switchyard and register it as the default browser via home-manager.
order: 22
---

Switchyard's flake exposes a [home-manager](https://nix-community.github.io/home-manager/) module that installs the package, optionally writes `~/.config/switchyard/config.toml`, and registers Switchyard as the default handler for `http(s)`/`text/html`.

## Setup

Add the flake to your inputs and import the module:

```nix
# flake.nix
{
  inputs.switchyard.url = "github:alyraffauf/switchyard";
  # ...
  outputs = { self, nixpkgs, home-manager, switchyard, ... }: {
    homeConfigurations.you = home-manager.lib.homeManagerConfiguration {
      pkgs = nixpkgs.legacyPackages.x86_64-linux;
      modules = [
        switchyard.homeManagerModules.switchyard
        ./home.nix
      ];
    };
  };
}
```

```nix
# home.nix
{ ... }: {
  programs.switchyard = {
    enable = true;
    setAsDefaultBrowser = true;
  };
}
```

## Options

- **`enable`** *(bool, default `false`)* — Install Switchyard and enable the module.
- **`package`** *(package, default `switchyard`)* — The Switchyard derivation to install.
- **`setAsDefaultBrowser`** *(bool, default `true`)* — Register Switchyard as the default handler for `x-scheme-handler/http`, `x-scheme-handler/https`, and `text/html` via `xdg.mimeApps`.
- **`settings`** *(TOML, default `{}`)* — Contents of `~/.config/switchyard/config.toml`. Freeform; any key from Switchyard's `Config` struct is accepted, omitted keys fall back to built-in defaults.

## Declarative Configuration

`settings` is a freeform TOML value — any key Switchyard understands is accepted, and new fields added to Switchyard's `Config` struct work without changes to the module:

```nix
programs.switchyard.settings = {
  favorite_browser = "firefox";
  remove_tracking_parameters = true;
  rules = [
    {
      name = "work";
      browser = "chromium";
      conditions = [ { type = "domain"; pattern = "corp.example.com"; } ];
    }
  ];
};
```

> **Warning:** Switchyard overwrites `config.toml` on every in-app config change. Edits made through the GUI will be lost on the next `home-manager switch`. Treat `settings` as declarative: pick one source of truth.

See the [configuration reference](/docs/configuration/) for the full schema.
