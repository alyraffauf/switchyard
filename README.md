# Switchyard

A configurable default browser for Linux. Route URLs to different browsers based on rules, or choose manually with a quick picker.

Inspired by [Choosy](https://choosy.app/) for macOS.

<p align="center">
  <img src="docs/images/switchyard.png" alt="Switchyard Settings" width="400">
  <img src="docs/images/switchyard-prompt.png" alt="Switchyard Picker" width="280">
</p>

## Features

- **Rule-based routing**: Automatically open URLs in specific browsers based on patterns
- **Multiple pattern types**: Exact domain, URL contains, wildcard (glob), and regex
- **Quick picker**: When no rule matches, choose from detected browsers with keyboard or mouse
- **Keyboard-first**: Press 1-9 to instantly select a browser, arrow keys to navigate
- **Live config reload**: Edit the config file externally and changes apply immediately
- **GTK4 + libadwaita**: Native GNOME look and feel

## Installation

### Building from Source

Requires Go 1.21+ and GTK4/libadwaita development libraries.

```bash
# Fedora
sudo dnf install gtk4-devel glib2-devel gobject-introspection-devel libadwaita-devel

# Build
make build

# Install to /usr/local (requires build first)
sudo make install

# Or install to custom prefix
sudo make install PREFIX=/usr
```

### Building Flatpak

```bash
# Ensure dependencies are vendored
go mod tidy
go mod vendor

# Install GNOME runtime
flatpak install flathub org.gnome.Platform//49 org.gnome.Sdk//49 org.freedesktop.Sdk.Extension.golang

# Build and install
flatpak-builder --user --install --force-clean build-dir flatpak/io.github.alyraffauf.Switchyard.yml
```

### Set as Default Browser

```bash
xdg-settings set default-web-browser io.github.alyraffauf.Switchyard.desktop
```

Or use your desktop environment's settings to set Switchyard as the default browser.

## Usage

```bash
# Open settings
switchyard

# Open a URL (typically called automatically by the system)
switchyard "https://example.com"
```

### Keyboard Shortcuts

**In the picker:**
- `1-9` - Select browser by number
- `↑/↓` - Navigate list
- `Enter` - Open in selected browser
- `Escape` - Cancel

**In settings:**
- `Ctrl+Q` - Quit

## Configuration

Config file location: `~/.config/switchyard/config.toml`

```toml
prompt_on_click = true
default_browser = ""

[[rules]]
name = "Work GitHub"
pattern = "github.com"
pattern_type = "domain"
browser = "firefox.desktop"

[[rules]]
name = "Google Apps"
pattern = "*.google.com"
pattern_type = "glob"
browser = "google-chrome.desktop"

[[rules]]
pattern = "youtube.com/watch"
pattern_type = "keyword"
browser = "brave-browser.desktop"
```

### Rule Options

| Field | Description |
|-------|-------------|
| `name` | Optional friendly name displayed in the UI |
| `pattern` | The pattern to match against |
| `pattern_type` | One of: `domain`, `keyword`, `glob`, `regex` |
| `browser` | Desktop file ID of the target browser |

### Pattern Types

| Type | Description | Example |
|------|-------------|---------|
| `domain` | Exact domain match | `github.com` |
| `keyword` | URL contains text | `youtube.com/watch` |
| `glob` | Wildcard pattern | `*.github.com` |
| `regex` | Regular expression | `^https://.*\.example\.(com\|org)` |

### Settings

| Setting | Description |
|---------|-------------|
| `prompt_on_click` | Show picker when no rule matches (default: true) |
| `default_browser` | Browser to use when prompt is disabled and no rule matches |

## License

[GPL-3.0-or-later](LICENSE.md)
