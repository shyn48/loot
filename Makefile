# loot — build & install (terminal TUI + macOS desktop app).

APP_NAME    := Loot
BINARY      := loot
BUNDLE_ID   := com.shyn.loot
DIST        := dist
APP         := $(DIST)/$(APP_NAME).app
ICNS        := packaging/AppIcon.icns
INSTALL_DIR := /Applications
GOBIN       := $(shell go env GOPATH)/bin

.PHONY: run gui test build app install install-cli install-quick-action demo icon clean

## demo: record demo.gif from demo.tape (needs `brew install vhs`)
demo:
	@command -v vhs >/dev/null || { echo "vhs not found — run: brew install vhs"; exit 1; }
	vhs demo.tape
	@echo "Wrote demo.gif"

## run: launch the TUI straight from source (dev)
run:
	go run .

## gui: launch the desktop giu window from source (dev)
gui:
	go run . --gui

## test: run the test suite
test:
	go test ./...

## build: compile the standalone binary into dist/
build:
	@mkdir -p $(DIST)
	go build -o $(DIST)/$(BINARY) .

## install-cli: put the loot TUI command on your PATH
install-cli: build
	cp $(DIST)/$(BINARY) "$(GOBIN)/$(BINARY)"
	@echo "Installed $(BINARY) to $(GOBIN) (make sure it is on your PATH)"

## install-quick-action: install the "Send to Loot" right-click Quick Action
install-quick-action:
	@mkdir -p "$(HOME)/Library/Services"
	@rm -rf "$(HOME)/Library/Services/Send to Loot.workflow"
	cp -R "packaging/quick-action/Send to Loot.workflow" "$(HOME)/Library/Services/"
	@echo "Installed. Enable it in System Settings ▸ Keyboard ▸ Keyboard Shortcuts ▸ Services if needed,"
	@echo "then right-click a URL/selected text ▸ Services ▸ Send to Loot."

## app: assemble a double-clickable, ad-hoc-signed .app (opens the TUI in Terminal)
app: build $(ICNS)
	@rm -rf "$(APP)"
	@mkdir -p "$(APP)/Contents/MacOS" "$(APP)/Contents/Resources"
	# The bundle executable is a launcher script that opens the TUI in Terminal.app;
	# the real binary is shipped alongside it as loot-bin (fallback).
	cp $(DIST)/$(BINARY) "$(APP)/Contents/MacOS/loot-bin"
	cp packaging/launch-terminal.sh "$(APP)/Contents/MacOS/loot"
	chmod +x "$(APP)/Contents/MacOS/loot"
	cp packaging/Info.plist "$(APP)/Contents/Info.plist"
	cp $(ICNS) "$(APP)/Contents/Resources/AppIcon.icns"
	# Ad-hoc signature so Gatekeeper allows it to launch on this machine.
	codesign --force --deep --sign - "$(APP)"
	@echo "Built \"$(APP)\""

## install: build the bundle and copy it into /Applications
install: app
	@rm -rf "$(INSTALL_DIR)/$(APP_NAME).app"
	cp -R "$(APP)" "$(INSTALL_DIR)/"
	@echo "Installed to $(INSTALL_DIR)/$(APP_NAME).app"

## icon: regenerate AppIcon.icns from packaging/genicon.go
icon: $(ICNS)

$(ICNS): packaging/genicon.go
	go run packaging/genicon.go
	@rm -rf packaging/AppIcon.iconset
	@mkdir -p packaging/AppIcon.iconset
	@for s in 16 32 128 256 512; do \
		d=$$((s*2)); \
		sips -z $$s   $$s   packaging/AppIcon.png --out packaging/AppIcon.iconset/icon_$${s}x$${s}.png    >/dev/null; \
		sips -z $$d   $$d   packaging/AppIcon.png --out packaging/AppIcon.iconset/icon_$${s}x$${s}@2x.png >/dev/null; \
	done
	iconutil -c icns packaging/AppIcon.iconset -o $(ICNS)
	@rm -rf packaging/AppIcon.iconset packaging/AppIcon.png
	@echo "Generated $(ICNS)"

## clean: remove build artifacts
clean:
	rm -rf $(DIST) $(ICNS)
