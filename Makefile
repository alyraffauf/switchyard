.PHONY: build install uninstall clean set-default vendor test test-coverage

PREFIX ?= /usr/local
BINDIR = $(PREFIX)/bin
DATADIR = $(PREFIX)/share
APPID = io.github.alyraffauf.Switchyard

build:
	go build -o switchyard ./src

install:
	install -Dm755 switchyard $(DESTDIR)$(BINDIR)/switchyard
	install -Dm644 data/$(APPID).desktop $(DESTDIR)$(DATADIR)/applications/$(APPID).desktop
	install -Dm644 data/$(APPID).metainfo.xml $(DESTDIR)$(DATADIR)/metainfo/$(APPID).metainfo.xml
	install -Dm644 data/icons/hicolor/scalable/apps/$(APPID).svg $(DESTDIR)$(DATADIR)/icons/hicolor/scalable/apps/$(APPID).svg

uninstall:
	rm -f $(DESTDIR)$(BINDIR)/switchyard
	rm -f $(DESTDIR)$(DATADIR)/applications/$(APPID).desktop
	rm -f $(DESTDIR)$(DATADIR)/metainfo/$(APPID).metainfo.xml
	rm -f $(DESTDIR)$(DATADIR)/icons/hicolor/scalable/apps/$(APPID).svg

clean:
	rm -f switchyard

set-default:
	xdg-mime default $(APPID).desktop x-scheme-handler/http
	xdg-mime default $(APPID).desktop x-scheme-handler/https
	@echo "Switchyard is now your default browser"

vendor:
	go mod tidy
	go mod vendor

test:
	@echo "Running unit tests..."
	go test -v ./src/config_test.go ./src/config.go

test-coverage:
	@echo "Running tests with coverage..."
	go test -coverprofile=coverage.out ./src/config_test.go ./src/config.go
	go tool cover -func=coverage.out
	@echo ""
	@echo "To view HTML coverage report, run: go tool cover -html=coverage.out"
