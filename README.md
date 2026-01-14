# Switchyard

A configurable default browser for Linux. Route URLs to different browsers based on rules, or choose manually with a quick picker.

<p align="center">
  <img src="docs/images/switchyard-prompt.png" alt="Switchyard Prompt" width="600">
</p>

<p align="center">
  <img src="docs/images/switchyard.png" alt="Switchyard Settings" width="600">
</p>

## Features

- **Rule-based routing**: Automatically open URLs in specific browsers based on patterns
- **Multi-condition rules**: Combine multiple conditions with AND/OR logic for complex routing
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
fallback_browser = ""
check_default_browser = true

# Simple rule with a single condition
[[rules]]
name = "Work GitHub"
browser = "firefox.desktop"

[[rules.conditions]]
type = "domain"
pattern = "github.com"

# Multi-condition rule with AND logic
[[rules]]
name = "Google Docs"
logic = "all"  # all conditions must match
browser = "google-chrome.desktop"

[[rules.conditions]]
type = "domain"
pattern = "docs.google.com"

[[rules.conditions]]
type = "keyword"
pattern = "edit"

# Multi-condition rule with OR logic
[[rules]]
name = "Video Sites"
logic = "any"  # any condition can match
browser = "brave-browser.desktop"

[[rules.conditions]]
type = "domain"
pattern = "youtube.com"

[[rules.conditions]]
type = "domain"
pattern = "vimeo.com"

[[rules.conditions]]
type = "domain"
pattern = "twitch.tv"

# Rule with always ask
[[rules]]
name = "Shopping Sites"
always_ask = true

[[rules.conditions]]
type = "keyword"
pattern = "amazon"
```

### Rule Options

| Field | Description |
|-------|-------------|
| `name` | Optional friendly name displayed in the UI |
| `conditions` | Array of conditions to match (see below) |
| `logic` | How to combine conditions: `all` (AND) or `any` (OR). Default: `all` |
| `browser` | Desktop file ID of the target browser |
| `always_ask` | If true, show browser picker instead of auto-opening (default: false) |

**Note:** Legacy single-pattern rules using `pattern` and `pattern_type` fields are still supported for backward compatibility, but will be automatically migrated to the multi-condition format when edited.

### Condition Options

| Field | Description |
|-------|-------------|
| `type` | One of: `domain`, `keyword`, `glob`, `regex` |
| `pattern` | The pattern to match against |

### Condition Types

| Type | Description | Example |
|------|-------------|---------|
| `domain` | Exact domain match | `github.com` |
| `keyword` | URL contains text | `youtube.com/watch` |
| `glob` | Wildcard pattern | `*.github.com` |
| `regex` | Regular expression | `^https://.*\.example\.(com\|org)` |

### Logic Modes

- **`all`** (AND logic): All conditions in the rule must match for the rule to apply
- **`any`** (OR logic): Any single condition matching will trigger the rule

Use `all` for precise targeting (e.g., "docs.google.com AND contains 'edit'") and `any` for broad matching (e.g., "youtube.com OR vimeo.com OR twitch.tv").

### Settings

| Setting | Description |
|---------|-------------|
| `prompt_on_click` | Show picker when no rule matches (default: true) |
| `fallback_browser` | Fallback browser to use when prompt is disabled and no rule matches |
| `check_default_browser` | Prompt to set Switchyard as system default browser on startup (default: true) |

## Prior Art

Switchyard draws inspiration from excellent browser pickers on other platforms:

- **[Choosy](https://choosy.app/)** - The gold standard browser picker for macOS. Beautiful UI and great UX, but not available on Linux.
- **[Junction](https://github.com/sonnyp/Junction)** - Slick browser picker for Linux with a snazzy interface, but without rule-based URL routing.

Switchyard aims to combine the best of both: Choosy's rule-based routing with a fast, native Linux experience.

## License

[GPL-3.0-or-later](LICENSE.md)
