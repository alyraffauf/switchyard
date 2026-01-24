[![CI](https://github.com/alyraffauf/switchyard/actions/workflows/ci.yml/badge.svg)](https://github.com/alyraffauf/switchyard/actions/workflows/ci.yml) [![License: GPL v3](https://img.shields.io/badge/License-GPL%20v3-blue.svg)](http://www.gnu.org/licenses/gpl-3.0) [![Ko-fi](https://img.shields.io/badge/Donate-Ko--fi-ff5e5b?logo=ko-fi&logoColor=white)](https://ko-fi.com/alyraffauf)

<div align="center">
  <img width="128" height="128" src="data/icons/hicolor/scalable/apps/io.github.alyraffauf.Switchyard.svg" alt="Switchyard Icon">
  <h1>Switchyard</h1>
  <h3>A rules-based browser launcher for Linux.</h3>
</div>

<p align="center">
  <img src="docs/images/switchyard-picker.png" alt="Switchyard Launcher" width="600">
</p>

<p align="center">
  <a href="https://flathub.org/apps/io.github.alyraffauf.Switchyard"><img src="https://flathub.org/api/badge?locale=en&style=flat" alt="Get it on Flathub"></a>
</p>

## Features

- **Rule-based routing**: Automatically open URLs in specific browsers based on powerful patterns.
- **Multi-condition rules**: Combine multiple conditions with AND/OR logic for precise control.
- **Multiple pattern types**: Exact Domain, URL Contains, Wildcard, and Regex matching.
- **Negative patterns**: Exclude specific URLs by inverting any condition (e.g., "all GitHub except Gist").
- **Custom URI scheme**: Create links that specify browser preferences directly with `switchyard://` URLs.
- **Quick browser launcher**: When no rule matches, quickly select a browser with keyboard or mouse.
- **Keyboard shortcuts**: Press Ctrl+1-9 to instantly select a browser.
- **Lightweight**: Runs only when needed, no background processes.
- **GTK4 + libadwaita**: Native GNOME look and feel.

## Installation

### Flatpak (Recommended)

Switchyard is available on [Flathub](https://flathub.org/apps/io.github.alyraffauf.Switchyard):

```bash
flatpak install flathub io.github.alyraffauf.Switchyard
```

### Nix Flake

```bash
nix run github:alyraffauf/switchyard
```

### Building from Source

For non-Flatpak builds, requires Go 1.24+, GTK4/libadwaita development libraries, and [just](https://github.com/casey/just).

```bash
just install-deps # For Fedora
just build
sudo just install # To /usr/local
```

#### Building Flatpak

```bash
just flatpak # Build and install
```

## Documentation

- [Using](docs/Using.md) - Set as default browser, usage examples.
- [Configuration](docs/Configuration.md) - Config file format, rules, and settings.
- [URI Scheme](docs/URI%20Scheme.md) - Custom `switchyard://` URLs for specifying browser preferences.
- [Prior Art](docs/Prior%20Art.md) - Similar tools that inspired Switchyard.
