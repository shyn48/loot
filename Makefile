# Shyn Download Manager — build & macOS packaging.

APP_NAME    := Shyn Download Manager
BINARY      := gownloader
BUNDLE_ID   := com.shyn.gownloader
DIST        := dist
APP         := $(DIST)/$(APP_NAME).app
ICNS        := packaging/AppIcon.icns
INSTALL_DIR := /Applications

.PHONY: run build app install icon clean

## run: launch the app straight from source (dev)
run:
	go run .

## build: compile the standalone binary into dist/
build:
	@mkdir -p $(DIST)
	go build -o $(DIST)/$(BINARY) .

## app: assemble a double-clickable, ad-hoc-signed .app bundle
app: build $(ICNS)
	@rm -rf "$(APP)"
	@mkdir -p "$(APP)/Contents/MacOS" "$(APP)/Contents/Resources"
	cp $(DIST)/$(BINARY) "$(APP)/Contents/MacOS/$(BINARY)"
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
