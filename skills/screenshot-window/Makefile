BINARY := bin/wincap
IDENTIFIER := com.trodemaster.wincap

# Defaults to ad-hoc ("-"), which is what `swift build` already does implicitly
# on Apple Silicon. Ad-hoc signatures embed a hash of the binary's own bytes as
# part of the code identity, so every rebuild looks like a brand-new app to
# TCC -- any Screen Recording grant tied to it goes stale on the next `make
# build`. Export WINCAP_CODESIGN_IDENTITY (a Developer ID Application identity,
# or a local self-signed Code Signing cert from Keychain Access) to sign with a
# stable identity instead, so the grant survives rebuilds. Never hardcode an
# actual identity string here -- it stays out of this repo entirely.
WINCAP_CODESIGN_IDENTITY ?= -

.PHONY: build clean

build:
	swift build -c release
	mkdir -p bin
	cp .build/release/wincap $(BINARY)
	codesign --force --sign "$(WINCAP_CODESIGN_IDENTITY)" --identifier $(IDENTIFIER) --options runtime $(BINARY)

clean:
	rm -rf .build $(BINARY)
