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
    sudo dnf install gtk4-devel glib2-devel gobject-introspection-devel libadwaita-devel just

# Install Flatpak dependencies
install-flatpak-deps:
    flatpak install flathub org.gnome.Platform//50 org.gnome.Sdk//50 org.freedesktop.Sdk.Extension.golang//25.08

# Build the application
build:
    go build -mod=vendor -trimpath -ldflags="-s -w" -o switchyard ./src

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

# Set as default browser
set-default:
    xdg-mime default {{APPID}}.desktop x-scheme-handler/http
    xdg-mime default {{APPID}}.desktop x-scheme-handler/https
    @echo "Switchyard is now your default browser"

# Update dependencies
vendor:
    go mod tidy
    go mod vendor

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
    flatpak run org.flatpak.Builder --user --install --force-clean --repo=build-repo build-dir flatpak/{{APPID}}.Devel.yml

# Cut a new release: bump version, commit, tag, and push master + tag
release version:
    #!/usr/bin/env bash
    set -euo pipefail
    version="{{version}}"
    tag="v${version}"
    metainfo="data/{{APPID}}.metainfo.xml"

    if ! [[ "$version" =~ ^[0-9]+\.[0-9]+\.[0-9]+(-[0-9A-Za-z.-]+)?(\+[0-9A-Za-z.-]+)?$ ]]; then
        echo "error: '$version' is not valid semver (expected X.Y.Z)" >&2
        exit 1
    fi

    branch="$(git symbolic-ref --short HEAD)"
    if [ "$branch" != "master" ]; then
        echo "error: must be on master (currently on $branch)" >&2
        exit 1
    fi

    # Allow src/app.go and the metainfo file to be dirty; nothing else.
    dirty="$(git status --porcelain -- ':!src/app.go' ":!${metainfo}")"
    if [ -n "$dirty" ]; then
        echo "error: working tree has changes outside src/app.go and ${metainfo}:" >&2
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

    git add src/app.go "${metainfo}"
    git commit -m "update for ${tag}"
    git tag -a "${tag}" -m "${tag}"
    git push origin master
    git push origin "${tag}"
