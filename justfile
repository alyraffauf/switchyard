PREFIX := env_var_or_default('PREFIX', '/usr/local')
DESTDIR := env_var_or_default('DESTDIR', '')
BINDIR := PREFIX / 'bin'
DATADIR := PREFIX / 'share'
APPID := 'io.github.alyraffauf.Switchyard'

# Show available recipes
default:
    @just --list

# Install development dependencies
install-deps:
    sudo dnf install golang gtk4-devel glib2-devel gobject-introspection-devel libadwaita-devel just

# Install Flatpak dependencies
install-flatpak-deps:
    flatpak install flathub org.gnome.Platform//50 org.gnome.Sdk//50 org.freedesktop.Sdk.Extension.golang//25.08

# Build the application
build:
    go build -mod=mod -trimpath -ldflags="-s -w" -o switchyard ./src

# Install to system
install: build
    install -Dm755 switchyard {{DESTDIR}}{{BINDIR}}/switchyard
    install -Dm644 data/{{APPID}}.desktop {{DESTDIR}}{{DATADIR}}/applications/{{APPID}}.desktop
    install -Dm644 data/{{APPID}}.metainfo.xml {{DESTDIR}}{{DATADIR}}/metainfo/{{APPID}}.metainfo.xml
    install -Dm644 data/icons/hicolor/scalable/apps/{{APPID}}.svg {{DESTDIR}}{{DATADIR}}/icons/hicolor/scalable/apps/{{APPID}}.svg

# Uninstall from system
uninstall:
    rm -f {{DESTDIR}}{{BINDIR}}/switchyard
    rm -f {{DESTDIR}}{{DATADIR}}/applications/{{APPID}}.desktop
    rm -f {{DESTDIR}}{{DATADIR}}/metainfo/{{APPID}}.metainfo.xml
    rm -f {{DESTDIR}}{{DATADIR}}/icons/hicolor/scalable/apps/{{APPID}}.svg

# Clean build artifacts
clean:
    rm -f switchyard

# Update the bundled AdGuard tracking-parameter filter list
update-adguard:
    curl -fsSL https://raw.githubusercontent.com/AdguardTeam/AdGuardFilters/master/TrackParamFilter/sections/general_url.txt -o src/embedded/adguard_filter.txt
    @echo "Updated src/embedded/adguard_filter.txt"

# Set as default browser
set-default:
    xdg-mime default {{APPID}}.desktop x-scheme-handler/http
    xdg-mime default {{APPID}}.desktop x-scheme-handler/https
    @echo "Switchyard is now your default browser"

# Update Go dependencies
update-go-deps:
    go mod tidy
    scripts/generate-flatpak-go-modules.py

# Run unit tests
test:
    @echo "Running unit tests..."
    go test -v ./src/config_test.go ./src/url_test.go ./src/pattern_test.go ./src/validation_test.go ./src/redirection_test.go ./src/app.go ./src/config.go ./src/url.go ./src/pattern.go ./src/validation.go ./src/redirection.go

# Run tests with coverage report
test-coverage:
    @echo "Running tests with coverage..."
    go test -coverprofile=coverage.out ./src/config_test.go ./src/url_test.go ./src/pattern_test.go ./src/validation_test.go ./src/redirection_test.go ./src/app.go ./src/config.go ./src/url.go ./src/pattern.go ./src/validation.go ./src/redirection.go
    go tool cover -func=coverage.out
    @echo ""
    @echo "To view HTML coverage report, run: go tool cover -html=coverage.out"

# Build and install Flatpak (development version)
flatpak:
    [ -f build-repo/config ] || rm -rf build-repo
    flatpak run org.flatpak.Builder --user --install --force-clean --repo=build-repo build-dir flatpak/{{APPID}}.Devel.yml

# Build Flatpak and export a .flatpak bundle for distribution
flatpak-bundle: flatpak
    flatpak build-bundle build-repo switchyard.flatpak {{APPID}}.Devel

# Build the browser extension (TypeScript + React)
build-extension:
    cd webextension && npm ci && npm run build

# Package the already-built extension as a Firefox-targeted directory
package-extension-firefox:
    #!/usr/bin/env bash
    set -euo pipefail
    root="{{justfile_directory()}}"
    extdir="${root}/webextension"
    outdir="${root}/firefox-extension"

    rm -rf "${outdir}"
    mkdir -p "${outdir}"
    cp -R "${extdir}/build" "${extdir}/icons" "${extdir}/popup.html" "${outdir}/"
    rm -f "${outdir}/build/manifest.firefox.json"
    cp "${extdir}/manifest.firefox.json" "${outdir}/manifest.json"
    echo "Extension packaged: firefox-extension/"

# Package the already-built extension as a Chrome-targeted zip archive
package-extension-chrome:
    #!/usr/bin/env bash
    set -euo pipefail
    root="{{justfile_directory()}}"
    extdir="${root}/webextension"
    outfile="${root}/switchyard-webextension-chrome.zip"
    staging="$(mktemp -d)"
    trap 'rm -rf "${staging}"' EXIT

    rm -f "${outfile}"
    cp -R "${extdir}/build" "${extdir}/icons" "${extdir}/popup.html" "${staging}/"
    jq 'del(.key)' "${extdir}/manifest.json" > "${staging}/manifest.json"
    cd "${staging}" && zip -r "${outfile}" .
    echo "Extension bundled: switchyard-webextension-chrome.zip"

# Build the extension once and package it for both Firefox and Chrome
bundle-extension: build-extension package-extension-firefox package-extension-chrome

# Cut a new release: bump version, commit, tag, and push master + tag
release version:
    #!/usr/bin/env bash
    set -euo pipefail
    version="{{version}}"
    tag="v${version}"
    metainfo="data/{{APPID}}.metainfo.xml"
    manifest="webextension/manifest.json"
    firefox_manifest="webextension/manifest.firefox.json"
    pkgjson="webextension/package.json"

    if ! [[ "$version" =~ ^[0-9]+\.[0-9]+\.[0-9]+(-[0-9A-Za-z.-]+)?(\+[0-9A-Za-z.-]+)?$ ]]; then
        echo "error: '$version' is not valid semver (expected X.Y.Z)" >&2
        exit 1
    fi

    branch="$(git symbolic-ref --short HEAD)"
    if [ "$branch" != "master" ]; then
        echo "error: must be on master (currently on $branch)" >&2
        exit 1
    fi

    # Allow src/app.go, metainfo, extension manifests, and package.json to be dirty; nothing else.
    dirty="$(git status --porcelain -- ':!src/app.go' ":!${metainfo}" ":!${manifest}" ":!${firefox_manifest}" ":!${pkgjson}")"
    if [ -n "$dirty" ]; then
        echo "error: working tree has changes outside src/app.go, ${metainfo}, ${manifest}, ${firefox_manifest}:" >&2
        echo "$dirty" >&2
        exit 1
    fi

    if git rev-parse --verify --quiet "refs/tags/${tag}" >/dev/null; then
        echo "error: tag ${tag} already exists" >&2
        exit 1
    fi

    if ! grep -q "<release version=\"${version}\"" "${metainfo}"; then
        echo "error: ${metainfo} has no <release version=\"${version}\"> entry" >&2
        echo "       add a release entry with notes before running 'just release'" >&2
        exit 1
    fi

    sed -i -E "s/^(\s*Version\s*=\s*)\"[^\"]+\"/\1\"${version}\"/" src/app.go
    if ! grep -qE "Version\s*=\s*\"${version}\"" src/app.go; then
        echo "error: failed to update Version in src/app.go" >&2
        exit 1
    fi

    sed -i -E 's/^(\s*"version"\s*:\s*)"[^"]+"/\1"'"${version}"'"/' "${manifest}"
    if ! grep -qE '"version"\s*:\s*"'"${version}"'"' "${manifest}"; then
        echo "error: failed to update version in ${manifest}" >&2
        exit 1
    fi

    sed -i -E 's/^(\s*"version"\s*:\s*)"[^"]+"/\1"'"${version}"'"/' "${firefox_manifest}"
    if ! grep -qE '"version"\s*:\s*"'"${version}"'"' "${firefox_manifest}"; then
        echo "error: failed to update version in ${firefox_manifest}" >&2
        exit 1
    fi

    sed -i -E 's/^(\s*"version"\s*:\s*)"[^"]+"/\1"'"${version}"'"/' "${pkgjson}"
    if ! grep -qE '"version"\s*:\s*"'"${version}"'"' "${pkgjson}"; then
        echo "error: failed to update version in ${pkgjson}" >&2
        exit 1
    fi

    git add src/app.go "${metainfo}" "${manifest}" "${firefox_manifest}" "${pkgjson}"
    git commit -m "update for ${tag}"
    git tag -a "${tag}" -m "${tag}"
    git push origin master
    git push origin "${tag}"
