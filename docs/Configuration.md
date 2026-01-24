# Configuration

Switchyard can be configured through its settings UI or by editing the config file directly.

## Config File Location

- Standard: `~/.config/switchyard/config.toml`
- Flatpak: `~/.var/app/io.github.alyraffauf.Switchyard/config/switchyard/config.toml`

## Example

```toml
prompt_on_click = true
favorite_browser = ""
check_default_browser = true

# URL rewrites (domain type is default)
[[rewrites]]
find = "x.com"
replace = "twitter.com"

[[rewrites]]
find = "reddit.com"
replace = "old.reddit.com"

# URL rewrite to remove query parameters
[[rewrites]]
type = "url"
find = "?utm_*"
replace = "" 

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

# Rule with negated condition
[[rules]]
name = "GitHub (not Gist)"
logic = "all"
browser = "firefox.desktop"

[[rules.conditions]]
type = "glob"
pattern = "*.github.com"

[[rules.conditions]]
type = "domain"
pattern = "gist.github.com"
negate = true  # URL must NOT match this pattern

# Rule with always ask
[[rules]]
name = "Shopping Sites"
always_ask = true

[[rules.conditions]]
type = "keyword"
pattern = "amazon"
```

## Settings

- **prompt_on_click**: Show launcher when no rule matches (default: true).
- **favorite_browser**: Favorite browser that always appears first in launcher and is used as fallback when launcher is disabled.
- **check_default_browser**: Prompt to set Switchyard as system default browser on startup (default: true).

## Rules

Rules define how URLs are routed to browsers. Each rule has conditions that determine when it matches.

- **name**: Optional friendly name displayed in the UI.
- **conditions**: Array of conditions to match (see below).
- **logic**: How to combine conditions: `all` (AND) or `any` (OR). Default: `all`.
- **browser**: [Desktop file ID](https://specifications.freedesktop.org/desktop-entry-spec/latest/) of the target browser (e.g. `firefox.desktop`, `com.google.Chrome.desktop`).
- **always_ask**: If true, show browser launcher instead of auto-opening (default: false).

## Conditions

Each condition specifies a pattern to match against the URL.

- **type**: Match type (see below).
- **pattern**: The pattern to match against.
- **negate**: If true, the condition is inverted (URL must NOT match). Default: false.

## Condition Types

- **domain**: Exact domain match (e.g. `github.com`).
- **keyword**: URL contains text (e.g. `youtube.com/watch`).
- **glob**: Wildcard pattern with `*` (e.g. `*.github.com`).
- **regex**: Regular expression (e.g. `^https://.*\.example\.(com|org)`).

## Logic Modes

- **all** (AND): All conditions must match for the rule to apply.
- **any** (OR): Any single condition matching triggers the rule.

Use `all` for precise targeting (e.g., "docs.google.com AND contains 'edit'") and `any` for broad matching (e.g., "youtube.com OR vimeo.com OR twitch.tv").

## URL Rewrites

Rewrites modify URLs before routing rules are evaluated. There are two types:

- **type**: `domain` (default) or `url`.
- **find**: Pattern to match.
- **replace**: Text to replace with. Leave empty to remove matches.

### Rewrite Types

**Domain** rewrites match the exact hostname. Use this to switch between sites:

```toml
[[rewrites]]
find = "reddit.com"
replace = "old.reddit.com"

[[rewrites]]
find = "x.com"
replace = "twitter.com"
```

**URL** rewrites match anywhere in the URL and support `*` wildcards. Use this to clean up links:

```toml
[[rewrites]]
type = "url"
find = "?utm_*"
replace = ""

[[rewrites]]
type = "url"
find = "&fbclid=*"
replace = ""
```

Rewrites are applied in order, and matching is case-insensitive.
