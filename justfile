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
    flatpak install flathub org.gnome.Platform//49 org.gnome.Sdk//49 org.freedesktop.Sdk.Extension.golang//25.08

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
    go test -v ./src/config_test.go ./src/validation_test.go ./src/app.go ./src/config.go ./src/validation.go

# Run tests with coverage report
test-coverage:
    @echo "Running tests with coverage..."
    go test -coverprofile=coverage.out ./src/config_test.go ./src/validation_test.go ./src/app.go ./src/config.go ./src/validation.go
    go tool cover -func=coverage.out
    @echo ""
    @echo "To view HTML coverage report, run: go tool cover -html=coverage.out"

# Build and install Flatpak (development version)
flatpak:
    flatpak-builder --user --install --force-clean build-dir flatpak/{{APPID}}.Devel.yml
