# Switchyard URI Scheme

Switchyard registers a custom URI scheme that allows links to specify browser preferences directly. This is useful for situations where you want to create links that always open in a specific browser, but don't want a permanent rule. Example use cases include note-taking, to-do apps, etc.

## Format

```
switchyard://open?url=<encoded-url>&browser=<desktop-file-id>[,<desktop-file-id>,...]
```

## Parameters

- **url** (required): The URL to open, percent-encoded.
- **browser** (optional): A comma-separated list of [desktop file IDs](https://specifications.freedesktop.org/desktop-entry/latest/) in order of preference. Switchyard opens the URL in the first listed browser that is installed. If omitted, standard routing rules apply.

Desktop file IDs match the `.desktop` filename without the extension (e.g. `org.mozilla.firefox`, `com.google.Chrome`).

## Behavior

1. Switchyard looks for each browser in the preference list in order.
2. The URL opens in the first browser found on the system.
3. If no listed browser is installed, Switchyard displays the browser launcher, allowing the user to select from installed browsers.

## Examples

Open in Firefox:

```
switchyard://open?url=https://example.com&browser=org.mozilla.firefox
```

Open in Firefox if installed, otherwise Chrome:

```
switchyard://open?url=https://example.com&browser=org.mozilla.firefox,com.google.Chrome
```

Percent-encoded special characters:

```
switchyard://open?url=https%3A%2F%2Fexample.com%3Ffoo%3Dbar&browser=org.mozilla.firefox
```
